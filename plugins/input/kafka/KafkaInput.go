package kafka

import (
	"time"

	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/pkg/cluster"
	"github.com/funkygao/dbus/pkg/kafka"
	"github.com/funkygao/dbus/pkg/model"
	conf "github.com/funkygao/jsconf"
	log "github.com/funkygao/log4go"
)

// KafkaInput is an input plugin that consumes data stream from a single specified kafka topic.
type KafkaInput struct {
	c *kafka.Consumer
}

func (this *KafkaInput) Init(config *conf.Conf) {
}

func (*KafkaInput) SampleConfig() string {
	return ``
}

func (this *KafkaInput) Ack(pack *engine.Packet) error {
	// TODO checkpoint
	return nil
}

func (this *KafkaInput) End(r engine.InputRunner) {}

func (this *KafkaInput) Run(r engine.InputRunner, h engine.PluginHelper) error {
	name := r.Name()
	backoff := time.Second * 5
	ex := r.Exchange()
	stopper := r.Stopper()

	defer func() {
		if this.c != nil {
			this.c.Stop()
		}
	}()

	var myResources []cluster.Resource
	resourcesCh := r.Resources()
	cf := kafka.DefaultConfig()

	for {
	RESTART_CONSUME:

		// wait till got some resource
		for {
			if len(myResources) != 0 {
				log.Trace("[%s] bingo! %d: %+v", name, len(myResources), myResources)
				break
			}

			log.Trace("[%s] awaiting resources", name)

			select {
			case <-stopper:
				log.Debug("[%s] yes sir!", name)
				return nil
			case myResources = <-resourcesCh:
			}
		}

		dsns := make([]string, len(myResources))
		for i, res := range myResources {
			dsns[i] = res.DSN()
		}

		log.Trace("[%s] starting consumer from %+v...", name, dsns)

		this.c = kafka.NewConsumer(dsns, cf)
		if err := this.c.Start(); err != nil {
			panic(err)
		}

		msgs := this.c.Messages()
		kafkaErrors := this.c.Errors()
		for {
			select {
			case <-stopper:
				log.Debug("[%s] yes sir!", name)
				return nil

			case err, ok := <-kafkaErrors:
				if !ok {
					log.Debug("[%s] consumer stopped", name)
					return nil
				}

				log.Error("[%s] backoff %s: %v, stop from %+v", name, backoff, err, dsns)
				this.c.Stop()

				select {
				case <-time.After(backoff):
				case <-stopper:
					return nil
				}
				goto RESTART_CONSUME

			case pack := <-ex.InChan():
				select {
				case <-stopper:
					log.Debug("[%s] yes sir!", name)
					return nil

				case myResources = <-resourcesCh:
					log.Trace("[%s] cluster rebalanced, stop from %+v", name, dsns)
					this.c.Stop()
					goto RESTART_CONSUME

				case msg, ok := <-msgs:
					if !ok {
						log.Debug("[%s] consumer stopped", name)
						return nil
					}

					pack.Payload = model.ConsumerMessage{ConsumerMessage: msg}
					ex.Emit(pack)
				}
			}
		}

	}

	return nil
}
