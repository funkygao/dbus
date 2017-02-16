package kafka

import (
	"github.com/Shopify/sarama"
	log "github.com/funkygao/log4go"
)

// Producer is a kafka producer that is transparent for sync/async mode.
type Producer struct {
	cf   *Config
	name string

	p  sarama.SyncProducer
	ap sarama.AsyncProducer

	sendMessage func(*sarama.ProducerMessage) error

	onError   func(*sarama.ProducerError)
	onSuccess func(*sarama.ProducerMessage)
}

func NewProducer(name string, brokers []string, cf *Config) (*Producer, error) {
	var err error
	p := &Producer{
		name: name,
		cf:   cf,
	}
	if cf.async {
		p.ap, err = sarama.NewAsyncProducer(brokers, cf.Sarama)
		p.sendMessage = p.asyncSend
	} else {
		p.p, err = sarama.NewSyncProducer(brokers, cf.Sarama)
		p.sendMessage = p.syncSend
	}

	return p, err
}

func (p *Producer) Start() error {
	if !p.cf.async {
		return nil
	}

	if p.onError == nil || p.onSuccess == nil {
		return ErrNotReady
	}

	go func() {
		// loop till Producer closed
		for {
			select {
			case msg, ok := <-p.ap.Successes():
				if !ok {
					log.Trace("[%s] closed", p.name)
					return
				}

				p.onSuccess(msg)

			case err, ok := <-p.ap.Errors():
				if !ok {
					log.Trace("[%s] closed", p.name)
					return
				}

				p.onError(err)
			}
		}
	}()

	return nil
}

// Close will drain and close the Producer.
func (p *Producer) Stop() error {
	if p.cf.async {
		return p.ap.Close()
	}

	return p.p.Close()
}

func (p *Producer) SetErrorHandler(f func(err *sarama.ProducerError)) error {
	if !p.cf.async {
		return ErrNotAllowed
	}

	p.onError = f
	return nil
}

func (p *Producer) SetSuccessHandler(f func(err *sarama.ProducerMessage)) error {
	if !p.cf.async {
		return ErrNotAllowed
	}

	p.onSuccess = f
	return nil
}

// Send will send a kafka message.
// For async mode, will never return error.
func (p *Producer) Send(m *sarama.ProducerMessage) error {
	return p.sendMessage(m)
}

func (p *Producer) asyncSend(m *sarama.ProducerMessage) error {
	log.Debug("[%s] async sending: %+v", p.name, m)

	p.ap.Input() <- m
	return nil
}

func (p *Producer) syncSend(m *sarama.ProducerMessage) error {
	log.Debug("[%s] sync sending: %+v", p.name, m)

	_, _, err := p.p.SendMessage(m)
	return err
}
