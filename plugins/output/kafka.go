package output

import (
	"time"

	"github.com/Shopify/sarama"
	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/plugins/input/myslave"
	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/gafka/zk"
	conf "github.com/funkygao/jsconf"
	log "github.com/funkygao/log4go"
)

type KafkaOutput struct {
	zone, cluster, topic string

	zkzone   *zk.ZkZone
	ack      sarama.RequiredAcks
	async    bool
	compress bool

	p  sarama.SyncProducer
	ap sarama.AsyncProducer

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
	ack := sarama.RequiredAcks(config.Int("ack", int(sarama.WaitForLocal)))
	if ack != sarama.WaitForAll && ack != sarama.WaitForLocal && ack != sarama.NoResponse {
		// -1, 1, 0
		panic("invalid ack")
	}
	this.ack = ack
	this.async = config.Bool("async", false)
	this.compress = config.Bool("compress", false)
	this.zkzone = zk.NewZkZone(zk.DefaultConfig(this.zone, ctx.ZoneZkAddrs(this.zone)))
	this.myslave = myslave.New().LoadConfig(config)
}

func (this *KafkaOutput) Run(r engine.OutputRunner, h engine.PluginHelper) error {
	if err := this.prepareProducer(); err != nil {
		return err
	}

	defer func() {
		if this.async {
			this.ap.Close()
		} else {
			this.p.Close()
		}
	}()

	for {
		select {
		case pack, ok := <-r.InChan():
			if !ok {
				return nil
			}

			row, ok := pack.Payload.(*myslave.RowsEvent)
			if !ok {
				log.Error("wrong payload: %+v", pack.Payload)
				continue
			}

			this.sendMessage(row)
			pack.Recycle()
		}
	}

	return nil
}

func (this *KafkaOutput) prepareProducer() error {
	cf := sarama.NewConfig()
	if this.compress {
		cf.Producer.Compression = sarama.CompressionSnappy
	}

	zkcluster := this.zkzone.NewCluster(this.cluster)

	if !this.async {
		cf.ChannelBufferSize = 256 // TODO
		cf.Producer.RequiredAcks = this.ack
		p, err := sarama.NewSyncProducer(zkcluster.BrokerList(), cf)
		if err != nil {
			return err
		}

		this.p = p
		return nil
	}

	// async producer
	cf.Producer.RequiredAcks = sarama.NoResponse
	cf.Producer.Flush.Frequency = time.Second * 2
	cf.Producer.Flush.Messages = 1000 // TODO
	cf.Producer.Flush.MaxMessages = 0 // unlimited
	ap, err := sarama.NewAsyncProducer(zkcluster.BrokerList(), cf)
	if err != nil {
		return err
	}

	this.ap = ap
	go func() {
		log.Trace("reaping async kafka[%s] errors", this.cluster)

		for err := range this.ap.Errors() {
			log.Error("%s %s", this.cluster, err)
		}
	}()

	return nil
}

func (this *KafkaOutput) sendMessage(row *myslave.RowsEvent) {
	msg := &sarama.ProducerMessage{
		Topic: this.topic,
		Value: sarama.ByteEncoder(row.Bytes()),
	}

	var (
		partition int32
		offset    int64
		err       error
	)
	if !this.async {
		for {
			if partition, offset, err = this.p.SendMessage(msg); err == nil {
				break
			}

			log.Error("%s.%s.%s {%s} %v", this.zone, this.cluster, this.topic, row, err)
			time.Sleep(time.Second)
		}

		if err = this.myslave.MarkAsProcessed(row); err != nil {
			log.Warn("%s.%s.%s {%s} %v", this.zone, this.cluster, this.topic, row, err)
		}

		log.Debug("%d/%d %s", partition, offset, row)
		return
	}

	// async
	this.ap.Input() <- msg

}

func init() {
	engine.RegisterPlugin("KafkaOutput", func() engine.Plugin {
		return new(KafkaOutput)
	})
}
