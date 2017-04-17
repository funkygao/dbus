package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/pkg/kafka"
	"github.com/funkygao/dbus/pkg/model"
	log "github.com/funkygao/log4go"
)

func (this *KafkaOutput) setupProducer(r engine.OutputRunner) *kafka.Producer {
	cf := kafka.DefaultConfig()
	cf.Sarama.Producer.Flush.Messages = r.Conf().Int("batch_size", 1024)
	cf.Sarama.Producer.RequiredAcks = sarama.RequiredAcks(r.Conf().Int("ack", int(sarama.WaitForAll)))
	switch r.Conf().String("mode", "async") {
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
	producer := kafka.NewProducer(r.Conf().String("name", "undefined"), zkcluster.BrokerList(), cf)

	switch r.Conf().String("qos", "LossTolerant") {
	case "LossTolerant":
		cf.LossTolerant()

		producer.SetErrorHandler(func(err *sarama.ProducerError) {
			// e,g.
			// kafka: Failed to produce message to topic dbustest: kafka server: Message was too large, server rejected it to avoid allocation error.
			// kafka server: Unexpected (unknown?) server error.
			// java.lang.OutOfMemoryError: Direct buffer memory
			row := err.Msg.Value.(*model.RowsEvent)
			log.Error("[%s.%s.%s] %s %s", this.zone, this.cluster, this.topic, err, row.MetaInfo())
		})

		producer.SetSuccessHandler(func(msg *sarama.ProducerMessage) {
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
			// FIXME delayed recycle will block input channel, so currently let input chan bigger than batch size
			pack.Recycle()
		})

	case "ThroughputFirst":
		cf.ThroughputFirst()

	default:
		panic("invalid qos")
	}

	return producer
}
