package output

import (
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/pkg/kafka"
	"github.com/funkygao/dbus/pkg/model"
	"github.com/funkygao/dbus/pkg/myslave"
	"github.com/funkygao/gafka/telemetry"
	"github.com/funkygao/gafka/zk"
	"github.com/funkygao/go-metrics"
	conf "github.com/funkygao/jsconf"
	log "github.com/funkygao/log4go"
)

type sentPos struct {
	Log string
	Pos uint32
}

// FIXME tightly coupled with binlog
type KafkaOutput struct {
	zone, cluster, topic string

	zkzone *zk.ZkZone
	p      *kafka.Producer

	pos            *sentPos
	loadPosOnStart bool
	err            metrics.Counter

	// FIXME should be shared with MysqlbinlogInput
	// currently, KafkaOutput MUST setup master_host/master_port to correctly checkpoint position
	myslave *myslave.MySlave
}

func (this *KafkaOutput) Init(config *conf.Conf) {
	this.zone = config.String("zone", "")
	this.cluster = config.String("cluster", "")
	this.topic = config.String("topic", "")
	if this.cluster == "" || this.zone == "" || this.topic == "" {
		panic("invalid configuration")
	}

	this.zkzone = engine.Globals().GetOrRegisterZkzone(this.zone)

	// ack is ignored in async mode
	cf := kafka.DefaultConfig()
	cf.Sarama.Producer.RequiredAcks = sarama.RequiredAcks(config.Int("ack", int(sarama.WaitForAll)))
	if !config.Bool("async", true) {
		cf.SyncMode()
	}

	zkcluster := this.zkzone.NewCluster(this.cluster)
	this.p = kafka.NewProducer(fmt.Sprintf("%s.%s.%s", this.zone, this.cluster, this.topic), zkcluster.BrokerList(), cf)

	key := fmt.Sprintf("myslave.%s", config.String("myslave_key", ""))
	this.myslave = engine.Globals().Registered(key).(*myslave.MySlave)

	this.pos = &sentPos{}
	this.loadPosOnStart = config.Bool("load_position", true)
}

func (this *KafkaOutput) Run(r engine.OutputRunner, h engine.PluginHelper) error {
	this.err = metrics.NewRegisteredCounter(telemetry.Tag(this.zone, this.cluster, this.topic)+"dbus.kafka.err", metrics.DefaultRegistry)

	if this.loadPosOnStart {
		this.loadPosition()
	}

	this.p.SetErrorHandler(func(err *sarama.ProducerError) {
		this.err.Inc(1)

		// e,g.
		// kafka: Failed to produce message to topic dbustest: kafka server: Message was too large, server rejected it to avoid allocation error.
		// kafka server: Unexpected (unknown?) server error. TODO
		// java.lang.OutOfMemoryError: Direct buffer memory
		row := err.Msg.Value.(*model.RowsEvent)
		log.Error("[%s.%s.%s] %s %s", this.zone, this.cluster, this.topic, err, row.MetaInfo())
	})

	this.p.SetSuccessHandler(func(msg *sarama.ProducerMessage) {
		row := msg.Value.(*model.RowsEvent)
		this.markAsSent(row)
		if err := this.myslave.MarkAsProcessed(row); err != nil {
			log.Error("[%s.%s.%s] {%s} %v", this.zone, this.cluster, this.topic, row, err)
		}
	})
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

			// best effort to reduce dup message to kafka
			if row.Log > this.pos.Log ||
				(row.Log == this.pos.Log && row.Position > this.pos.Pos) {
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
			} else {
				this.markAsSent(row)
				log.Debug("[%s.%s.%s] skipped {%s}", this.zone, this.cluster, this.topic, row)
			}

			pack.Recycle()
		}
	}

	return nil
}

func (this *KafkaOutput) loadPosition() {
	b, err := this.zkzone.NewCluster(this.cluster).TailMessage(this.topic, 0, 1) // FIXME 1 partition allowed only
	if err != nil {
		panic(err)
	}

	if len(b) == 1 {
		// has checkpoint in kafka
		row := &model.RowsEvent{}
		if err := row.Decode(b[0]); err != nil {
			panic(err)
		}

		this.pos.Log = row.Log
		this.pos.Pos = row.Position
	}

	log.Trace("[%s.%s.%s/0] resumed at %+v", this.zone, this.cluster, this.topic, *this.pos)
}

func (this *KafkaOutput) markAsSent(row *model.RowsEvent) {
	this.pos.Log = row.Log
	this.pos.Pos = row.Position
}

func init() {
	engine.RegisterPlugin("KafkaOutput", func() engine.Plugin {
		return new(KafkaOutput)
	})
}
