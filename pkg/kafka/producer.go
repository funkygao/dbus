package kafka

import (
	"runtime/debug"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/pkg/batcher"
	log "github.com/funkygao/log4go"
)

// Producer is a uniform kafka producer that is transparent for sync/async mode.
type Producer struct {
	cf      *Config
	name    string
	brokers []string
	stopper chan struct{}
	wg      sync.WaitGroup
	b       *batcher.Batcher
	m       *producerMetrics

	p  sarama.SyncProducer
	ap sarama.AsyncProducer

	// Send will send a kafka message.
	Send func(*sarama.ProducerMessage) error

	onError   func(*sarama.ProducerError)
	onSuccess func(*sarama.ProducerMessage)
}

// NewProducer creates a uniform kafka producer.
func NewProducer(name string, brokers []string, cf *Config) *Producer {
	if !cf.dryrun && len(brokers) == 0 {
		panic(name + " empty brokers")
	}

	p := &Producer{
		name:    name,
		brokers: brokers,
		cf:      cf,
		stopper: make(chan struct{}),
	}

	return p
}

// Start is REQUIRED before the producer is able to produce.
func (p *Producer) Start() error {
	p.m = newMetrics(p.name)

	var err error
	if p.cf.dryrun {
		// dryrun mode
		p.b = batcher.New(p.cf.Sarama.Producer.Flush.Messages)
		p.Send = p.dryrunSend
		return nil
	}

	if !p.cf.async {
		// sync mode
		p.p, err = sarama.NewSyncProducer(p.brokers, p.cf.Sarama)
		p.Send = p.syncSend
		return err
	}

	// async mode
	if p.onError == nil || p.onSuccess == nil {
		return ErrNotReady
	}

	p.b = batcher.New(p.cf.Sarama.Producer.Flush.Messages)
	if p.ap, err = sarama.NewAsyncProducer(p.brokers, p.cf.Sarama); err != nil {
		return err
	}

	p.Send = p.asyncSend

	p.wg.Add(1)
	go p.dispatchCallbacks()
	p.wg.Add(1)
	go p.asyncSendWorker()

	return nil
}

// Close will drain and close the Producer.
func (p *Producer) Close() error {
	close(p.stopper)

	if p.cf.dryrun {
		return nil
	}

	if p.cf.async {
		p.ap.AsyncClose()
		p.b.Close()
		p.wg.Wait()
		return nil
	}

	return p.p.Close()
}

// ClientID returns the client id for the kafka connection.
func (p *Producer) ClientID() string {
	return p.cf.Sarama.ClientID
}

// SetErrorHandler setup the async producer unretriable errors, e.g:
// ErrInvalidPartition, ErrMessageSizeTooLarge, ErrIncompleteResponse
// ErrBreakerOpen(e,g. update leader fails).
// And it is *REQUIRED* for async producer.
// For sync producer it is not allowed.
func (p *Producer) SetErrorHandler(f func(err *sarama.ProducerError)) error {
	if !p.cf.async {
		return ErrNotAllowed
	}

	if f == nil {
		p.cf.Sarama.Producer.Return.Errors = false
	}
	if p.onError != nil {
		return ErrNotAllowed
	}
	p.onError = f
	return nil
}

// SetSuccessHandler sets the success produced message callback for async producer.
// And it is *REQUIRED* for async producer.
// For sync producer it is not allowed.
func (p *Producer) SetSuccessHandler(f func(err *sarama.ProducerMessage)) error {
	if !p.cf.async {
		return ErrNotAllowed
	}

	if f == nil {
		p.cf.Sarama.Producer.Return.Successes = false
	}
	if p.onSuccess != nil {
		return ErrNotAllowed
	}
	p.onSuccess = f
	return nil
}

func (p *Producer) syncSend(m *sarama.ProducerMessage) error {
	_, _, err := p.p.SendMessage(m)
	if err != nil {
		p.m.syncFail.Mark(1)
	} else {
		p.m.syncOk.Mark(1)
	}
	return err
}

func (p *Producer) dryrunSend(m *sarama.ProducerMessage) error {
	p.b.Put(m)
	p.b.Succeed() // i,e. onSuccess called silently
	pack := m.Metadata.(*engine.Packet)
	pack.Recycle()
	return nil
}

func (p *Producer) asyncSend(m *sarama.ProducerMessage) error {
	p.b.Put(m)
	return nil
}

func (p *Producer) asyncSendWorker() {
	defer p.wg.Done()

	for {
		select {
		case <-p.stopper:
			return

		default:
			if msg, err := p.b.Get(); err == nil {
				// FIXME what if msg is nil
				// FIXME a batch of 10, msg7 always fail, lead to dead loop
				pm := msg.(*sarama.ProducerMessage)
				p.ap.Input() <- pm
				p.m.asyncSend.Mark(1)
			} else {
				log.Trace("[%s] batcher closed", p.name)
				return
			}
		}
	}
}

func (p *Producer) dispatchCallbacks() {
	defer func() {
		if err := recover(); err != nil {
			log.Critical("[%s] %v\n%s", p.name, err, string(debug.Stack()))
		}

		p.wg.Done()
	}()

	errChan := p.ap.Errors()
	okChan := p.ap.Successes()
	for {
		select {
		case msg, ok := <-okChan:
			if !ok {
				okChan = nil
			} else {
				p.b.Succeed()
				p.m.asyncOk.Mark(1)
				p.onSuccess(msg)
			}

		case err, ok := <-errChan:
			if !ok {
				errChan = nil
			} else {
				if rewind := p.b.Fail(); rewind {
					// FIXME if kafka got wrong(e,g. Messages are rejected since there are fewer in-sync replicas than required)
					// there will be many duplicated messages
					//
					// we do sleep between the same batch reties to solve it
					time.Sleep(time.Second)
				}
				p.m.asyncFail.Mark(1)
				p.onError(err)
			}
		}

		if okChan == nil && errChan == nil {
			log.Trace("[%s] success & err chan both closed", p.name)
			return
		}
	}
}
