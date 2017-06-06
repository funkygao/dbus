package discard

import (
	"github.com/funkygao/dbus/pkg/checkpoint"
)

var (
	_ checkpoint.Checkpoint = &checkpointDiscard{}
)

type checkpointDiscard struct {
}

func New() checkpoint.Checkpoint {
	return &checkpointDiscard{}
}

func (z *checkpointDiscard) Shutdown() error {
	return nil
}

func (z *checkpointDiscard) Commit(state checkpoint.State) error {
	return nil
}

func (z *checkpointDiscard) LastPersistedState(state checkpoint.State) (err error) {
	return
}
