package kafka

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"os"
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
	cf.ChannelBufferSize = 256 * 4 // default was 256
	cf.Producer.Timeout = time.Second * 30
	cf.Producer.Compression = sarama.CompressionNone
	cf.Producer.Retry.Max = 5
	cf.Producer.Retry.Backoff = time.Millisecond * 350
	cf.Producer.RequiredAcks = sarama.WaitForLocal // default ack
	cf.Metadata.RefreshFrequency = time.Minute * 10
	cf.Metadata.Retry.Max = 5
	cf.Metadata.Retry.Backoff = time.Millisecond * 500

	// async specific
	cf.Producer.Return.Errors = true
	cf.Producer.Return.Successes = true
	cf.Producer.Flush.Frequency = time.Second / 2
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

	// explicitly zero batch
	c.Sarama.Producer.Flush.Frequency = 0
	c.Sarama.Producer.Flush.Bytes = 0
	c.Sarama.Producer.Flush.Messages = 0
	return c
}

func (c *Config) AsyncMode() *Config {
	c.async = true
	return c
}

func generateClientID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}

	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40

	host, err := os.Hostname()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s:%s", host, fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])), nil
}

func init() {
	ctx.LoadFromHome()
	log.SetFlags(log.LstdFlags | log.Lshortfile) // for sarama
}
