package kafka

import (
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/pkg/kafka"
	"github.com/funkygao/golib/gofmt"
	conf "github.com/funkygao/jsconf"
	log "github.com/funkygao/log4go"
)

// KafkaOutput is an Output plugin that send pack to a single specified kafka topic.
type KafkaOutput struct {
	zone, cluster, topic string
	reporter             bool
}

// Init setup KafkaOutput state according to config section.
// Default kafka delivery: async WaitForAll.
func (this *KafkaOutput) Init(config *conf.Conf) {
	var err error
	this.zone, this.cluster, this.topic, _, err = kafka.ParseDSN(config.String("dsn", ""))
	if err != nil || this.cluster == "" || this.zone == "" || this.topic == "" {
		panic("invalid configuration: " + fmt.Sprintf("%s.%s.%s", this.zone, this.cluster, this.topic))
	}
	this.reporter = config.Bool("reporter", false)
}

func (*KafkaOutput) SampleConfig() string {
	return `
	batch_size: 1024
	ack: -1
	mode: "async" // async|sync|dryrun
	qos: "LossTolerant" // LossTolerant|ThroughputFirst
	dsn: "kafka:local://me/foobar"
	reporter: true
	`
}

func (this *KafkaOutput) CleanupForRestart() bool {
	return true // yes, restart allowed
}

func (this *KafkaOutput) Run(r engine.OutputRunner, h engine.PluginHelper) error {
	producer := this.setupProducer(r)
	if err := producer.Start(); err != nil {
		return err
	}

	defer func() {
		log.Trace("[%s.%s.%s] draining...", this.zone, this.cluster, this.topic)

		if err := producer.Close(); err != nil {
			log.Error("[%s.%s.%s] drain: %s", this.zone, this.cluster, this.topic, err)
		} else {
			log.Trace("[%s.%s.%s] drained ok", this.zone, this.cluster, this.topic)
		}
	}()

	tick := time.NewTicker(time.Second * 10)
	defer tick.Stop()

	var reportTimer <-chan time.Time
	if this.reporter {
		reportTimer = tick.C
	}

	var n, lastN int64
	for {
		select {
		case <-reportTimer:
			log.Trace("[%s] throughput %s/s", r.Name(), gofmt.Comma((n-lastN)/10))
			lastN = n

		case pack, ok := <-r.Exchange().InChan():
			if !ok {
				// engine stops
				log.Trace("[%s] %d packets received", r.Name(), n)
				return nil
			}

			n++

			// loop is for sync mode only: async send will never return error
			for {
				if err := producer.Send(&sarama.ProducerMessage{
					Topic:    this.topic,
					Value:    pack.Payload,
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
