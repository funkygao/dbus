package output

import (
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/pkg/model"
	"github.com/funkygao/dbus/pkg/myslave"
	"github.com/funkygao/gafka/telemetry"
	"github.com/funkygao/gafka/zk"
	"github.com/funkygao/go-metrics"
	conf "github.com/funkygao/jsconf"
	log "github.com/funkygao/log4go"
)

type sentPos struct {
	Log string
	Pos uint32
}

// FIXME tightly coupled with binlog
type KafkaOutput struct {
	zone, cluster, topic string

	zkzone   *zk.ZkZone
	ack      sarama.RequiredAcks
	async    bool
	compress bool

	p           sarama.SyncProducer
	ap          sarama.AsyncProducer
	sendMessage func(row *model.RowsEvent)

	pos            *sentPos
	loadPosOnStart bool
	err            metrics.Counter

	// FIXME should be shared with MysqlbinlogInput
	// currently, KafkaOutput MUST setup master_host/master_port to correctly checkpoint position
	myslave *myslave.MySlave
}

func (this *KafkaOutput) Init(config *conf.Conf) {
	this.zone = config.String("zone", "")
	this.cluster = config.String("cluster", "")
	this.topic = config.String("topic", "")
	if this.cluster == "" || this.zone == "" || this.topic == "" {
		panic("invalid configuration")
	}

	// ack is ignored in async mode
	this.ack = sarama.RequiredAcks(config.Int("ack", int(sarama.WaitForAll)))
	this.async = config.Bool("async", false)
	if this.async {
		this.sendMessage = this.asyncSendMessage
	} else {
		this.sendMessage = this.syncSendMessage
	}
	this.compress = config.Bool("compress", false)
	this.zkzone = engine.Globals().GetOrRegisterZkzone(this.zone)

	this.pos = &sentPos{}
	this.loadPosOnStart = config.Bool("load_position", true)

	key := fmt.Sprintf("myslave.%s", config.String("myslave_key", ""))
	this.myslave = engine.Globals().Registered(key).(*myslave.MySlave)
}

func (this *KafkaOutput) Run(r engine.OutputRunner, h engine.PluginHelper) error {
	if this.loadPosOnStart {
		this.loadPosition()
	}

	if err := this.prepareProducer(); err != nil {
		return err
	}

	defer func() {
		log.Trace("[%s.%s.%s] start draining...", this.zone, this.cluster, this.topic)

		var err error
		if this.async {
			err = this.ap.Close()
		} else {
			err = this.p.Close()
		}

		if err != nil {
			log.Error("[%s.%s.%s] drain: %s", this.zone, this.cluster, this.topic, err)
		} else {
			log.Trace("[%s.%s.%s] drained ok", this.zone, this.cluster, this.topic)
		}
	}()

	for {
		select {
		case pack, ok := <-r.InChan():
			if !ok {
				log.Trace("[%s.%s.%s] yes sir!", this.zone, this.cluster, this.topic)
				return nil
			}

			row, ok := pack.Payload.(*model.RowsEvent)
			if !ok {
				log.Error("[%s.%s.%s] bad payload: %+v", this.zone, this.cluster, this.topic, pack.Payload)
				continue
			}

			// best effort to reduce dup message to kafka
			if row.Log > this.pos.Log ||
				(row.Log == this.pos.Log && row.Position > this.pos.Pos) {
				this.sendMessage(row)
			} else {
				this.markAsSent(row)
				log.Debug("[%s.%s.%s] skipped {%s}", this.zone, this.cluster, this.topic, row)
			}

			pack.Recycle()
		}
	}

	return nil
}

func (this *KafkaOutput) prepareProducer() error {
	cf := sarama.NewConfig()
	cf.ChannelBufferSize = 256 * 2 // default was 256
	cf.Producer.RequiredAcks = this.ack
	cf.Producer.Retry.Backoff = time.Millisecond * 300
	cf.Producer.Retry.Max = 5
	if this.compress {
		cf.Producer.Compression = sarama.CompressionSnappy
	}

	this.err = metrics.NewRegisteredCounter(telemetry.Tag(this.zone, this.cluster, this.topic)+"dbus.kafka.err", metrics.DefaultRegistry)
	zkcluster := this.zkzone.NewCluster(this.cluster)

	if !this.async {
		p, err := sarama.NewSyncProducer(zkcluster.BrokerList(), cf)
		if err != nil {
			return err
		}

		this.p = p
		return nil
	}

	// async producer
	cf.Producer.Return.Errors = true
	cf.Producer.Return.Successes = true
	cf.Producer.Flush.Frequency = time.Second
	cf.Producer.Flush.Messages = 2000 // TODO
	//cf.Producer.Flush.Bytes = 64 << 10
	cf.Producer.Flush.MaxMessages = 0 // unlimited
	ap, err := sarama.NewAsyncProducer(zkcluster.BrokerList(), cf)
	if err != nil {
		return err
	}

	this.ap = ap
	go func() {
		for {
			select {
			case msg, ok := <-this.ap.Successes():
				if !ok {
					// producer closed
					return
				}

				row := msg.Value.(*model.RowsEvent)
				this.markAsSent(row)
				if err := this.myslave.MarkAsProcessed(row); err != nil {
					log.Error("[%s.%s.%s] {%s} %v", this.zone, this.cluster, this.topic, row, err)
				}

			case err, ok := <-this.ap.Errors():
				if !ok {
					// producer closed
					return
				}

				this.err.Inc(1)

				// e,g.
				// kafka: Failed to produce message to topic dbustest: kafka server: Message was too large, server rejected it to avoid allocation error.
				// kafka server: Unexpected (unknown?) server error. TODO
				// java.lang.OutOfMemoryError: Direct buffer memory
				row := err.Msg.Value.(*model.RowsEvent)
				log.Error("[%s.%s.%s] %s %s", this.zone, this.cluster, this.topic, err, row.MetaInfo())
			}
		}
	}()

	return nil
}

func (this *KafkaOutput) syncSendMessage(row *model.RowsEvent) {
	var (
		msg = &sarama.ProducerMessage{
			Topic: this.topic,
			Value: row,
		}

		partition int32
		offset    int64
		err       error
	)
	for {
		if partition, offset, err = this.p.SendMessage(msg); err == nil {
			break
		}

		log.Error("[%s.%s.%s] {%s} %v", this.zone, this.cluster, this.topic, row, err)
		time.Sleep(time.Second)
	}

	this.markAsSent(row)
	if err = this.myslave.MarkAsProcessed(row); err != nil {
		log.Error("[%s.%s.%s] {%s} %v", this.zone, this.cluster, this.topic, row, err)
	}

	log.Debug("[%s.%s.%s] sync sent [%d/%d] %s", this.zone, this.cluster, this.topic, partition, offset, row)
}

func (this *KafkaOutput) asyncSendMessage(row *model.RowsEvent) {
	log.Debug("[%s.%s.%s] async sending: %s", this.zone, this.cluster, this.topic, row)

	this.ap.Input() <- &sarama.ProducerMessage{
		Topic: this.topic,
		Value: row,
	}
}

func (this *KafkaOutput) loadPosition() {
	b, err := this.zkzone.NewCluster(this.cluster).TailMessage(this.topic, 0, 1) // FIXME 1 partition allowed only
	if err != nil {
		panic(err)
	}

	if len(b) == 1 {
		// has checkpoint in kafka
		row := &model.RowsEvent{}
		if err := row.Decode(b[0]); err != nil {
			panic(err)
		}

		this.pos.Log = row.Log
		this.pos.Pos = row.Position
	}

	log.Trace("[%s.%s.%s/0] resumed at %+v", this.zone, this.cluster, this.topic, *this.pos)
}

func (this *KafkaOutput) markAsSent(row *model.RowsEvent) {
	this.pos.Log = row.Log
	this.pos.Pos = row.Position
}

func init() {
	engine.RegisterPlugin("KafkaOutput", func() engine.Plugin {
		return new(KafkaOutput)
	})
}
