package engine

import (
	"net/http"

	log "github.com/funkygao/log4go"
	"github.com/gorilla/mux"
)

func (e *Engine) handleAPIPause(w http.ResponseWriter, r *http.Request, params map[string]interface{}) (interface{}, error) {
	vars := mux.Vars(r)
	inputPlugin := vars["input"]
	if _, present := e.InputRunners[inputPlugin]; !present {
		return nil, ErrInvalidParam
	}

	if p, ok := e.InputRunners[inputPlugin].Plugin().(Pauser); ok {
		return nil, p.Pause(e.InputRunners[inputPlugin])
	}

	log.Warn("plugin[%s] is not able to pause", inputPlugin)
	return nil, ErrInvalidParam
}

func (e *Engine) handleAPIResume(w http.ResponseWriter, r *http.Request, params map[string]interface{}) (interface{}, error) {
	vars := mux.Vars(r)
	inputPlugin := vars["input"]
	if _, present := e.InputRunners[inputPlugin]; !present {
		return nil, ErrInvalidParam
	}

	if p, ok := e.InputRunners[inputPlugin].Plugin().(Pauser); ok {
		return nil, p.Resume(e.InputRunners[inputPlugin])
	}

	log.Warn("plugin[%s] is not able to resume", inputPlugin)
	return nil, ErrInvalidParam
}
