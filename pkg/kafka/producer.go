package kafka

import (
	"sync"

	"github.com/Shopify/sarama"
	log "github.com/funkygao/log4go"
)

// Producer is a kafka producer that is transparent for sync/async mode.
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

func NewProducer(name string, brokers []string, cf *Config) *Producer {
	p := &Producer{
		name:    name,
		brokers: brokers,
		cf:      cf,
		stopper: make(chan struct{}),
	}

	return p
}

func (p *Producer) Start() error {
	var err error
	if p.cf.async {
		p.ap, err = sarama.NewAsyncProducer(p.brokers, p.cf.Sarama)
		p.sendMessage = p.asyncSend
	} else {
		p.p, err = sarama.NewSyncProducer(p.brokers, p.cf.Sarama)
		p.sendMessage = p.syncSend
	}
	if err != nil {
		return err
	}

	if !p.cf.async {
		return nil
	}

	if p.onError == nil || p.onSuccess == nil {
		return ErrNotReady
	}

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()

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
	}()

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

func (p *Producer) ClientID() string {
	return p.cf.Sarama.ClientID
}

// SetErrorHandler setup the async producer unretriable errors, e.g:
// ErrInvalidPartition, ErrMessageSizeTooLarge, ErrIncompleteResponse
// ErrBreakerOpen(e,g. update leader fails)
func (p *Producer) SetErrorHandler(f func(err *sarama.ProducerError)) error {
	if !p.cf.async {
		return ErrNotAllowed
	}

	if f == nil {
		p.cf.Sarama.Producer.Return.Errors = false
	}
	p.onError = f
	return nil
}

func (p *Producer) SetSuccessHandler(f func(err *sarama.ProducerMessage)) error {
	if !p.cf.async {
		return ErrNotAllowed
	}

	if f == nil {
		p.cf.Sarama.Producer.Return.Successes = false
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
