package kafka

import (
	"log"
	"time"

	"github.com/Shopify/sarama"
	"github.com/funkygao/gafka/ctx"
)

// Config is the configuration of kafka pkg.
type Config struct {
	async bool

	Sarama *sarama.Config
}

// DefaultConfig creates a default Config as async producer.
func DefaultConfig() *Config {
	cf := sarama.NewConfig()
	// common for sync and async
	cf.ChannelBufferSize = 256 * 4 // default was 256
	cf.Producer.Retry.Backoff = time.Millisecond * 300
	cf.Producer.Retry.Max = 5
	cf.Producer.RequiredAcks = sarama.WaitForLocal

	// async
	cf.Producer.Return.Errors = true
	cf.Producer.Return.Successes = true
	cf.Producer.Flush.Frequency = time.Second
	cf.Producer.Flush.Messages = 2000 // TODO
	cf.Producer.Flush.MaxMessages = 0 // unlimited
	//cf.Producer.Flush.Bytes = 64 << 10

	return &Config{
		Sarama: cf,
		async:  true,
	}
}

func (c *Config) Ack(ack sarama.RequiredAcks) *Config {
	c.Sarama.Producer.RequiredAcks = ack
	return c
}

func (c *Config) SyncMode() *Config {
	c.async = false
	return c
}

func (c *Config) AsyncMode() *Config {
	c.async = true
	return c
}

func (c *Config) Validate() error {
	return nil
}

func init() {
	ctx.LoadFromHome()
	log.SetFlags(log.LstdFlags | log.Lshortfile) // for sarama
}
