package engine

import (
	"fmt"
)

func (e *Engine) callSOS(format string, reason ...interface{}) {
	globals := Globals()
	zkzone := globals.GetOrRegisterZkzone(globals.Zone)
	if zkzone == nil {
		return
	}

	if len(reason) == 0 {
		zkzone.CallSOS(e.participant.String(), format)
	} else {
		zkzone.CallSOS(e.participant.String(), fmt.Sprintf(format, reason...))
	}
}
