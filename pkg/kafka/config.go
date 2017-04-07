package kafka

import (
	"fmt"
	"log"
	"os"
	"sync/atomic"
	"time"

	"github.com/Shopify/sarama"
	"github.com/funkygao/gafka/ctx"
)

var clientID uint32

// Config is the configuration of kafka pkg.
type Config struct {
	async  bool
	dryrun bool

	consumeChanBufSize int

	Sarama *sarama.Config
}

// DefaultConfig creates a default Config as async producer.
func DefaultConfig() *Config {
	cf := sarama.NewConfig()
	clientID, err := generateClientID()
	if err != nil {
		panic(err)
	}
	cf.ClientID = clientID

	// network
	cf.Net.DialTimeout = time.Second * 30
	cf.Net.ReadTimeout = time.Second * 30
	cf.Net.WriteTimeout = time.Second * 30
	cf.Net.MaxOpenRequests = 5

	// common for sync and async
	cf.ChannelBufferSize = 256 * 4        // default was 256
	cf.Producer.MaxMessageBytes = 1000000 // 1MB
	cf.Producer.Timeout = time.Second * 30
	cf.Producer.Compression = sarama.CompressionNone
	cf.Producer.Retry.Max = 3
	cf.Producer.Retry.Backoff = time.Millisecond * 350
	cf.Producer.RequiredAcks = sarama.WaitForLocal // default ack
	cf.Metadata.RefreshFrequency = time.Minute * 10
	cf.Metadata.Retry.Max = 5
	cf.Metadata.Retry.Backoff = time.Second

	// async specific
	cf.Producer.Return.Errors = true
	cf.Producer.Return.Successes = true
	cf.Producer.Flush.Frequency = time.Second / 2
	cf.Producer.Flush.Messages = 1024 // TODO
	cf.Producer.Flush.MaxMessages = 0 // unlimited
	//cf.Producer.Flush.Bytes = 64 << 10

	// consumer related
	cf.Consumer.Return.Errors = true
	cf.Consumer.MaxProcessingTime = time.Second * 2

	return &Config{
		Sarama:             cf,
		async:              true,
		dryrun:             false,
		consumeChanBufSize: 1 << 8,
	}
}

// Ack sets the kafka producer required ack parameter.
func (c *Config) Ack(ack sarama.RequiredAcks) *Config {
	c.Sarama.Producer.RequiredAcks = ack
	return c
}

// SyncMode will switch the kafka producer to sync mode.
func (c *Config) SyncMode() *Config {
	c.async = false

	// explicitly zero batch
	c.Sarama.Producer.Flush.Frequency = 0
	c.Sarama.Producer.Flush.Bytes = 0
	c.Sarama.Producer.Flush.Messages = 0
	return c
}

// SyncMode will switch the kafka producer to async mode.
func (c *Config) AsyncMode() *Config {
	c.async = true
	return c
}

func (c *Config) DryrunMode() *Config {
	c.dryrun = true
	return c
}

func (c *Config) SetConsumerChanBuffer(size int) *Config {
	c.consumeChanBufSize = size
	return c
}

func generateClientID() (string, error) {
	host, err := os.Hostname()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("dbus-%s-%s", host, fmt.Sprintf("%d", atomic.AddUint32(&clientID, 1))), nil
}

func init() {
	ctx.LoadFromHome()
	log.SetFlags(log.LstdFlags | log.Lshortfile) // for sarama
}
