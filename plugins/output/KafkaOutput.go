package output

import (
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/pkg/kafka"
	"github.com/funkygao/dbus/pkg/model"
	"github.com/funkygao/dbus/pkg/myslave"
	conf "github.com/funkygao/jsconf"
	log "github.com/funkygao/log4go"
)

var (
	_ engine.Output = &KafkaOutput{}
)

// KafkaOutput is an Output plugin that send pack to a single specified kafka topic.
type KafkaOutput struct {
	zone, cluster, topic string

	myslave *myslave.MySlave // FIXME ugly design, coupled with other plugin
	p       *kafka.Producer
}

// Init setup KafkaOutput state according to config section.
// Default kafka delivery: async WaitForAll.
func (this *KafkaOutput) Init(config *conf.Conf) {
	this.zone = config.String("zone", "")
	this.cluster = config.String("cluster", "")
	this.topic = config.String("topic", "")
	if this.cluster == "" || this.zone == "" || this.topic == "" {
		panic("invalid configuration: " + fmt.Sprintf("%s.%s.%s", this.zone, this.cluster, this.topic))
	}

	cf := kafka.DefaultConfig()
	cf.Sarama.Producer.RequiredAcks = sarama.RequiredAcks(config.Int("ack", int(sarama.WaitForAll)))
	if !config.Bool("async", true) {
		cf.SyncMode()
	}

	// get the bootstrap broker list
	zkzone := engine.Globals().GetOrRegisterZkzone(this.zone)
	zkcluster := zkzone.NewCluster(this.cluster)
	this.p = kafka.NewProducer(fmt.Sprintf("%s.%s.%s", this.zone, this.cluster, this.topic), zkcluster.BrokerList(), cf)
	this.myslave = engine.Globals().Registered(fmt.Sprintf("myslave.%s", config.String("myslave_key", ""))).(*myslave.MySlave)
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
		row := msg.Value.(*model.RowsEvent)
		if err := this.myslave.MarkAsProcessed(row); err != nil {
			log.Error("[%s.%s.%s] {%s} %v", this.zone, this.cluster, this.topic, row, err)
		}
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

	for {
		select {
		case pack, ok := <-r.InChan():
			if !ok {
				log.Trace("[%s.%s.%s] yes sir!", this.zone, this.cluster, this.topic)
				return nil
			}

			row, ok := pack.Payload.(*model.RowsEvent)
			if !ok {
				log.Error("[%s.%s.%s] bad payload: %+v", this.zone, this.cluster, this.topic, pack.Payload)
				continue
			}

			// loop is for sync mode only: async send will never return error
			for {
				if err := this.p.Send(&sarama.ProducerMessage{
					Topic: this.topic,
					Value: row,
				}); err == nil {
					break
				} else {
					log.Error("[%s.%s.%s] %+v", this.zone, this.cluster, this.topic, err)

					time.Sleep(time.Millisecond * 500)
				}
			}

			pack.Recycle()
		}
	}

	return nil
}

func init() {
	engine.RegisterPlugin("KafkaOutput", func() engine.Plugin {
		return new(KafkaOutput)
	})
}
