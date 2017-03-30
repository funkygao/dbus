package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/funkygao/dbus/engine"
)

// Consumer is a kafka low level consumer.
type Consumer struct {
	c sarama.Consumer
	p sarama.PartitionConsumer

	topic       string
	partitionID int32
}

func NewConsumer(dsn string, cf *Config) (*Consumer, error) {
	zone, cluster, topic, partitionID, err := ParseDSN(dsn)
	if err != nil {
		return nil, err
	}

	c, err := sarama.NewConsumer(engine.Globals().GetOrRegisterZkzone(zone).NewCluster(cluster).BrokerList(), cf.Sarama)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		c:           c,
		topic:       topic,
		partitionID: partitionID,
	}, nil
}

func (c *Consumer) Start() error {
	p, err := c.c.ConsumePartition(c.topic, c.partitionID, sarama.OffsetNewest) // TODO offset
	if err != nil {
		return err
	}

	c.p = p
	return nil
}

func (c *Consumer) Stop() error {
	c.p.Close()
	return c.c.Close()
}

func (c *Consumer) Messages() <-chan *sarama.ConsumerMessage {
	return c.p.Messages()
}
