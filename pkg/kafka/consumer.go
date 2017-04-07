package kafka

import (
	"sync"

	"github.com/Shopify/sarama"
	"github.com/funkygao/dbus/engine"
)

// Consumer is a kafka low level consumer that can consumer multiple zone/cluster kafka clusters.
type Consumer struct {
	dsns    []string
	cf      *Config
	stopper chan struct{}

	consumersMu sync.Mutex
	consumers   map[zoneCluster]sarama.Consumer

	wg   sync.WaitGroup
	once sync.Once

	errors   chan error
	messages chan *sarama.ConsumerMessage
}

// NewConsumer returns a kafka consumer.
func NewConsumer(dsns []string, cf *Config) *Consumer {
	return &Consumer{
		dsns:      dsns,
		cf:        cf,
		messages:  make(chan *sarama.ConsumerMessage, cf.consumeChanBufSize),
		errors:    make(chan error, cf.consumeChanBufSize),
		stopper:   make(chan struct{}),
		consumers: make(map[zoneCluster]sarama.Consumer),
	}
}

func (c *Consumer) Messages() <-chan *sarama.ConsumerMessage {
	return c.messages
}

func (c *Consumer) Errors() <-chan error {
	return c.errors
}

func (c *Consumer) Start() error {
	tps, err := parseDSNs(c.dsns)
	if err != nil {
		return err
	}

	for _, tp := range tps {
		consumer, err := sarama.NewConsumer(engine.Globals().GetOrRegisterZkzone(tp.zc.zone).
			NewCluster(tp.zc.cluster).BrokerList(), c.cf.Sarama)
		if err != nil {
			return err
		}
		c.consumersMu.Lock()
		c.consumers[tp.zc] = consumer
		c.consumersMu.Unlock()

		c.wg.Add(1)
		go c.consumerCluster(tp)
	}

	return nil
}

func (c *Consumer) Stop() error {
	err := ErrAlreadyClosed
	c.once.Do(func() {
		close(c.stopper)

		for _, c := range c.consumers {
			if e := c.Close(); e != nil {
				err = e
			}
		}

		c.wg.Wait()

		// safe to close
		close(c.errors)
		close(c.messages)
	})

	return err
}

func (c *Consumer) consumerCluster(tp topicPartitions) {
	defer c.wg.Done()

	c.consumersMu.Lock()
	consumer := c.consumers[tp.zc]
	c.consumersMu.Unlock()

	var wg sync.WaitGroup
	for topic, partitions := range tp.tps {
		for _, partitionID := range partitions {
			// FIXME initial offset, integration with checkpoint pkg
		RETRY:
			pc, err := consumer.ConsumePartition(topic, partitionID, sarama.OffsetNewest)
			if err != nil {
				if err == sarama.ErrOffsetOutOfRange {
					// TODO reset offset and retry
					goto RETRY
				}

				c.errors <- err
				return
			}

			wg.Add(1)
			go c.consumePartition(pc, &wg)
		}
	}

	wg.Wait()
}

func (c *Consumer) consumePartition(pc sarama.PartitionConsumer, wg *sync.WaitGroup) {
	defer func() {
		pc.Close()
		wg.Done()
	}()

	for {
		select {
		case <-c.stopper:
			return

		case err := <-pc.Errors():
			c.errors <- err

		case msg, ok := <-pc.Messages():
			if !ok {
				c.errors <- ErrConsumerBroken
				return
			}

			// TODO if blocked, consumer can't be stopped
			c.messages <- msg
		}
	}
}
