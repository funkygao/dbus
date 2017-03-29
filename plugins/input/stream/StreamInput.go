package mock

import (
	"bufio"

	"github.com/funkygao/dbus/engine"
	"github.com/funkygao/dbus/pkg/model"
	"github.com/funkygao/golib/pipestream"
	conf "github.com/funkygao/jsconf"
	log "github.com/funkygao/log4go"
)

type StreamInput struct {
	cmdAndArgs []string
}

func (this *StreamInput) Init(config *conf.Conf) {
	this.cmdAndArgs = config.StringList("cmd", nil)
	if len(this.cmdAndArgs) == 0 {
		panic("empty cmd in config")
	}
}

func (this *StreamInput) OnAck(pack *engine.Packet) error {
	return nil
}

func (this *StreamInput) Stop(r engine.InputRunner) {}

func (this *StreamInput) Run(r engine.InputRunner, h engine.PluginHelper) error {
	cmd := pipestream.New(this.cmdAndArgs[0], this.cmdAndArgs[1:]...)
	if err := cmd.Open(); err != nil {
		return err
	}
	defer cmd.Close()

	scanner := bufio.NewScanner(cmd.Reader())
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Bytes()

		pack, ok := <-r.InChan()
		if !ok {
			log.Trace("yes sir!")
			break
		}

		pack.Payload = model.Bytes(line)
		r.Inject(pack)
	}

	return nil
}
