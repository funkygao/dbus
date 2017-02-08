package output

import (
	"time"

	"github.com/Shopify/sarama"
	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/plugins/input/myslave"
	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/gafka/zk"
	conf "github.com/funkygao/jsconf"
	log "github.com/funkygao/log4go"
)

type KafkaOutput struct {
	zone, cluster, topic string

	zkzone  *zk.ZkZone
	myslave *myslave.MySlave
}

func (this *KafkaOutput) Init(config *conf.Conf) {
	this.zone = config.String("zone", "")
	this.cluster = config.String("cluster", "")
	this.topic = config.String("topic", "")
	if this.cluster == "" || this.zone == "" || this.topic == "" {
		panic("invalid configuration")
	}

	this.zkzone = zk.NewZkZone(zk.DefaultConfig(this.zone, ctx.ZoneZkAddrs(this.zone)))
	this.myslave = myslave.New().LoadConfig(config)
}

func (this *KafkaOutput) Run(r engine.OutputRunner, h engine.PluginHelper) error {
	// TODO async producer and SLA
	cf := sarama.NewConfig()
	cf.Producer.RequiredAcks = sarama.WaitForLocal
	p, err := sarama.NewSyncProducer(this.zkzone.NewCluster(this.cluster).BrokerList(), cf)
	if err != nil {
		return err
	}
	defer p.Close()

	var (
		partition int32
		offset    int64
	)
	for {
		select {
		case pack, ok := <-r.InChan():
			if !ok {
				return nil
			}

			row, ok := pack.Payload.(*myslave.RowsEvent)
			if !ok {
				log.Error("wrong payload: %+v", pack.Payload)
				continue
			}

			for {
				if partition, offset, err = p.SendMessage(&sarama.ProducerMessage{
					Topic: this.topic,
					Value: sarama.ByteEncoder(row.Bytes()),
				}); err == nil {
					break
				}

				log.Error("%s.%s.%s {%s} %v", this.zone, this.cluster, this.topic, row, err)
				time.Sleep(time.Second)
			}

			if err = this.myslave.MarkAsProcessed(row); err != nil {
				log.Warn("%s.%s.%s {%s} %v", this.zone, this.cluster, this.topic, row, err)
			}

			log.Debug("%d/%d %s", partition, offset, row)

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
