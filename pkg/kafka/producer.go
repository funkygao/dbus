package kafka

import (
	"sync"

	"github.com/Shopify/sarama"
	log "github.com/funkygao/log4go"
)

// Producer is a uniform kafka producer that is transparent for sync/async mode.
type Producer struct {
	cf      *Config
	name    string
	brokers []string
	stopper chan struct{}
	wg      sync.WaitGroup

	p  sarama.SyncProducer
	ap sarama.AsyncProducer

	sendMessage func(*sarama.ProducerMessage) error

	onError   func(*sarama.ProducerError)
	onSuccess func(*sarama.ProducerMessage)
}

// NewProducer creates a uniform kafka producer.
func NewProducer(name string, brokers []string, cf *Config) *Producer {
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
	var err error
	if !p.cf.async {
		// sync mode
		p.p, err = sarama.NewSyncProducer(p.brokers, p.cf.Sarama)
		p.sendMessage = p.syncSend
		return err
	}

	// async mode
	if p.onError == nil || p.onSuccess == nil {
		return ErrNotReady
	}

	p.ap, err = sarama.NewAsyncProducer(p.brokers, p.cf.Sarama)
	if err != nil {
		return err
	}

	p.sendMessage = p.asyncSend

	p.wg.Add(1)
	go p.dispatchCallbacks()

	return nil
}

// Close will drain and close the Producer.
func (p *Producer) Close() error {
	close(p.stopper)

	if p.cf.async {
		p.ap.AsyncClose()
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

// Send will send a kafka message.
func (p *Producer) Send(m *sarama.ProducerMessage) error {
	return p.sendMessage(m)
}

func (p *Producer) asyncSend(m *sarama.ProducerMessage) error {
	log.Debug("[%s] async sending: %+v", p.name, m)

	select {
	case <-p.stopper:
		return ErrStopping

	case p.ap.Input() <- m:
	}
	return nil
}

func (p *Producer) syncSend(m *sarama.ProducerMessage) error {
	log.Debug("[%s] sync sending: %+v", p.name, m)

	_, _, err := p.p.SendMessage(m)
	return err
}

func (p *Producer) dispatchCallbacks() {
	defer func() {
		if err := recover(); err != nil {
			log.Critical("[%s] %v", p.name, err)
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
				p.onSuccess(msg)
			}

		case err, ok := <-errChan:
			if !ok {
				errChan = nil
			} else {
				p.onError(err)
			}
		}

		if okChan == nil && errChan == nil {
			log.Trace("[%s] success & err chan both closed", p.name)
			return
		}
	}
}
