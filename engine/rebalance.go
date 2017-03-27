package engine

import (
	log "github.com/funkygao/log4go"
)

func (e *Engine) onControllerRebalance(decision map[string][]string) {
	for participant, resources := range decision {
		if e.participantID == participant {
			// TODO
			log.Trace("%+v", resources)
		} else {
			// TODO RPC remote dbus
		}
	}
}
