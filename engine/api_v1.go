package engine

import (
	"net/http"

	log "github.com/funkygao/log4go"
	"github.com/gorilla/mux"
)

func (e *Engine) handleAPIPauseV1(w http.ResponseWriter, r *http.Request, params map[string]interface{}) (interface{}, error) {
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

func (e *Engine) handleAPIResumeV1(w http.ResponseWriter, r *http.Request, params map[string]interface{}) (interface{}, error) {
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

func (e *Engine) handleAPIDecisionV1(w http.ResponseWriter, r *http.Request, params map[string]interface{}) (interface{}, error) {
	m := e.ClusterManager()
	if m == nil {
		return nil, ErrInvalidParam
	}

	return m.CurrentDecision(), nil
}

func (e *Engine) handleQueuesV1(w http.ResponseWriter, r *http.Request, params map[string]interface{}) (interface{}, error) {
	rs := make(map[string]int)

	globals := Globals()
	rs["hub"] = len(e.router.hub)
	rs["hub.free"] = globals.HubChanSize - rs["hub"]

	for name, ch := range e.inputRecycleChans {
		rs["input."+name] = len(ch)
		rs["input."+name+".free"] = globals.InputRecyclePoolSize - rs["input."+name]
	}

	rs["filter"] = len(e.filterRecycleChan)
	rs["filter.free"] = globals.FilterRecyclePoolSize - rs["filter"]

	for _, om := range e.router.outputMatchers {
		rs["output."+om.runner.Name()] = len(om.InChan())
		rs["output."+om.runner.Name()+".free"] = globals.PluginChanSize - rs["output."+om.runner.Name()]
	}

	return rs, nil
}
