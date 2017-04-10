package mock

import (
	"runtime"
	"time"

	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/pkg/model"
	conf "github.com/funkygao/jsconf"
	log "github.com/funkygao/log4go"
)

type MockInput struct {
	stopChan chan struct{}
	inChan   <-chan *engine.Packet

	payload engine.Payloader
	sleep   time.Duration
}

func (this *MockInput) Init(config *conf.Conf) {
	this.sleep = config.Duration("sleep", 0)
	this.stopChan = make(chan struct{})
	switch config.String("payload", "Bytes") {
	case "RowsEvent":
		this.payload = &model.RowsEvent{
			Log:       "mysql-bin.0001",
			Position:  498876,
			Schema:    "mydabase",
			Table:     "user_account",
			Action:    "I",
			Timestamp: 1486554654,
			Rows:      [][]interface{}{{"user", 15, "hello world"}},
		}

	default:
		this.payload = model.Bytes(`{"log":"6633343-bin.006419","pos":795931083,"db":"owl_t_prod_cd","tbl":"owl_mi","dml":"U","ts":1488934500,"rows":[[132332,"expired_keys","过期的key的个数","mock-monitor|member-mock|10.1.1.1|10489|expired_keys","mock-monitor",244526,"10489",null,"2015-12-25 17:12:00",null,null,"28571284","Ok","TypeLong","2017-03-08 08:54:01",null,null,null,null,null,null],[132332,"expired_keys","过期的key的个数","mock-monitor|member-mock|10.1.1.6|10489|expired_keys","mock-monitor",244526,"10489",null,"2015-12-25 17:12:00",null,null,"28571320","Ok","TypeLong","2017-03-08 08:55:00",null,null,null,null,null,null]]}`)
	}
}

func (this *MockInput) OnAck(pack *engine.Packet) error {
	return nil
}

func (this *MockInput) Stop(r engine.InputRunner) {
	log.Debug("[%s] stopping...", r.Name())
	close(this.stopChan)
}

func (this *MockInput) CleanupForRestart() bool {
	return true
}

func (this *MockInput) Pause(r engine.InputRunner) error {
	log.Info("[%s] paused", r.Name())
	this.inChan = nil
	return nil
}

func (this *MockInput) Resume(r engine.InputRunner) error {
	log.Info("[%s] resumed", r.Name())
	this.inChan = r.InChan()
	return nil
}

func (this *MockInput) Run(r engine.InputRunner, h engine.PluginHelper) error {
	this.inChan = r.InChan()
	for {
		select {
		case <-this.stopChan:
			return nil

		case pack, ok := <-this.inChan:
			if !ok {
				log.Debug("[%s] yes sir!", r.Name())
				break
			}

			pack.Payload = this.payload
			r.Inject(pack)

			if this.sleep > 0 {
				time.Sleep(this.sleep)
			}

		default:
			runtime.Gosched()
		}
	}

	return nil
}
