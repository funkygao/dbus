package output

import (
	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/gafka/cmd/kateway/hh"
	"github.com/funkygao/gafka/cmd/kateway/hh/disk"
	"github.com/funkygao/gafka/cmd/kateway/meta"
	"github.com/funkygao/gafka/cmd/kateway/meta/zkmeta"
	"github.com/funkygao/gafka/cmd/kateway/store"
	"github.com/funkygao/gafka/cmd/kateway/store/kafka"
	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/gafka/zk"
	conf "github.com/funkygao/jsconf"
)

type KafkaOutput struct {
	zone, cluster, topic string
	hhdirs               []string
	zkzone               *zk.ZkZone
}

func (this *KafkaOutput) Init(config *conf.Conf) {
	this.zone = config.String("zone", "")
	this.cluster = config.String("cluster", "")
	this.topic = config.String("topic", "")
	this.hhdirs = config.StringList("hhdirs", nil)
	if this.cluster == "" || this.zone == "" || this.topic == "" || len(this.hhdirs) == 0 {
		panic("invalid configuration")
	}

	this.zkzone = zk.NewZkZone(zk.DefaultConfig(this.zone, ctx.ZoneZkAddrs(this.zone)))

	meta.Default = zkmeta.New(zkmeta.DefaultConfig(), this.zkzone)
	meta.Default.Start()

	cfg := disk.DefaultConfig()
	cfg.Dirs = this.hhdirs
	if err := cfg.Validate(); err != nil {
		panic(err)
	}
	hh.Default = disk.New(cfg)
	if err := hh.Default.Start(); err != nil {
		panic(err)
	}

	store.DefaultPubStore = kafka.NewPubStore(100, 0, false, false, false)
}

func (this *KafkaOutput) Run(r engine.OutputRunner, h engine.PluginHelper) error {
	globals := engine.Globals()
	for {
		select {
		case pack, ok := <-r.InChan():
			if !ok {
				return nil
			}

			partition, offset, err := store.DefaultPubStore.SyncPub(this.cluster, this.topic, nil, pack.Payload.Bytes())
			if err != nil {
				hh.Default.Append(this.cluster, this.topic, nil, pack.Payload.Bytes())
			}

			globals.Printf("%d %d", partition, offset)

			pack.Recycle()

		}
	}

	return nil
}

func init() {
	engine.RegisterPlugin("KafkaOutput", func() engine.Plugin {
		return new(KafkaOutput)
	})
}
