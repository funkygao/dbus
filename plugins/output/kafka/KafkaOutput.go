package kafka

import (
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/pkg/kafka"
	"github.com/funkygao/dbus/pkg/model"
	"github.com/funkygao/golib/gofmt"
	conf "github.com/funkygao/jsconf"
	log "github.com/funkygao/log4go"
)

// KafkaOutput is an Output plugin that send pack to a single specified kafka topic.
type KafkaOutput struct {
	zone, cluster, topic string
	reporter             bool

	p *kafka.Producer
}

// Init setup KafkaOutput state according to config section.
// Default kafka delivery: async WaitForAll.
func (this *KafkaOutput) Init(config *conf.Conf) {
	var err error
	this.zone, this.cluster, this.topic, err = kafka.ParseDSN(config.String("dsn", ""))
	if err != nil || this.cluster == "" || this.zone == "" || this.topic == "" {
		panic("invalid configuration: " + fmt.Sprintf("%s.%s.%s", this.zone, this.cluster, this.topic))
	}
	this.reporter = config.Bool("reporter", false)

	cf := kafka.DefaultConfig()
	cf.Sarama.Producer.Flush.Messages = config.Int("batch_size", 1024)
	cf.Sarama.Producer.RequiredAcks = sarama.RequiredAcks(config.Int("ack", int(sarama.WaitForAll)))
	switch config.String("mode", "async") {
	case "sync":
		cf.SyncMode()
	case "async":
		cf.AsyncMode()
	case "dryrun":
		cf.DryrunMode()
	default:
		panic("invalid KafkaOut mode")
	}

	// get the bootstrap broker list
	zkzone := engine.Globals().GetOrRegisterZkzone(this.zone)
	zkcluster := zkzone.NewCluster(this.cluster)
	this.p = kafka.NewProducer(config.String("name", "undefined"), zkcluster.BrokerList(), cf)
}

func (this *KafkaOutput) CleanupForRestart() bool {
	return true // yes, restart allowed
}

func (this *KafkaOutput) Run(r engine.OutputRunner, h engine.PluginHelper) error {
	this.p.SetErrorHandler(func(err *sarama.ProducerError) {
		// e,g.
		// kafka: Failed to produce message to topic dbustest: kafka server: Message was too large, server rejected it to avoid allocation error.
		// kafka server: Unexpected (unknown?) server error.
		// java.lang.OutOfMemoryError: Direct buffer memory
		row := err.Msg.Value.(*model.RowsEvent)
		log.Error("[%s.%s.%s] %s %s", this.zone, this.cluster, this.topic, err, row.MetaInfo())
	})

	this.p.SetSuccessHandler(func(msg *sarama.ProducerMessage) {
		// FIXME what if:
		// [1, 2, 3, 4, 5] sent
		// [1, 2, 4, 5] ok, [3] fails
		// then shutdown dbusd? 3 might be lost
		pack := msg.Metadata.(*engine.Packet)
		if err := r.Ack(pack); err != nil {
			row := msg.Value.(*model.RowsEvent)
			log.Error("[%s.%s.%s] {%s} %v", this.zone, this.cluster, this.topic, row, err)
		}

		// safe to recycle
		// FIXME delayed recycle will block input channel, so currently let input chan
		// larger than batch size
		pack.Recycle()
	})

	// start the producer background routines
	if err := this.p.Start(); err != nil {
		return err
	}

	defer func() {
		log.Trace("[%s.%s.%s] start draining...", this.zone, this.cluster, this.topic)

		if err := this.p.Close(); err != nil {
			log.Error("[%s.%s.%s] drain: %s", this.zone, this.cluster, this.topic, err)
		} else {
			log.Trace("[%s.%s.%s] drained ok", this.zone, this.cluster, this.topic)
		}
	}()

	tick := time.NewTicker(time.Second * 10)
	defer tick.Stop()

	var ticker <-chan time.Time
	if this.reporter {
		ticker = tick.C
	}

	var n, lastN int64
	for {
		select {
		case <-ticker:
			log.Trace("[%s] throughput %s/s", r.Name(), gofmt.Comma((n-lastN)/10))
			lastN = n

		case pack, ok := <-r.InChan():
			if !ok {
				log.Trace("[%s.%s.%s] yes sir!", this.zone, this.cluster, this.topic)
				return nil
			}

			row, ok := pack.Payload.(*model.RowsEvent)
			if !ok {
				pack.Recycle()

				log.Error("[%s.%s.%s] bad payload: %+v", this.zone, this.cluster, this.topic, pack.Payload)
				continue
			}

			n++

			// loop is for sync mode only: async send will never return error
			for {
				if err := this.p.Send(&sarama.ProducerMessage{
					Topic:    this.topic,
					Value:    row,
					Metadata: pack,
				}); err == nil {
					break
				} else {
					log.Error("[%s.%s.%s] %+v", this.zone, this.cluster, this.topic, err)

					time.Sleep(time.Millisecond * 500)
				}
			}
		}
	}

	return nil
}
