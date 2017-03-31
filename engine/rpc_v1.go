package engine

import (
	"fmt"
	"io"
	"net/http"

	"github.com/funkygao/dbus/pkg/cluster"
	log "github.com/funkygao/log4go"
)

const (
	maxRPCBodyLen = 128 << 10
)

func (e *Engine) doLocalRebalance(w http.ResponseWriter, r *http.Request) {
	if r.ContentLength > maxRPCBodyLen {
		log.Warn("too large RPC request body: %d", r.ContentLength)
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		return
	}

	body := make([]byte, r.ContentLength)
	_, err := io.ReadFull(r.Body, body)
	if err != nil {
		log.Error("%v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	resources := cluster.UnmarshalRPCResources(body)
	log.Trace("got %d resources: %v", len(resources), resources)

	// merge resources by input plugin name
	inputResourcesMap := make(map[string][]cluster.Resource) // inputName:resources
	for _, res := range resources {
		if _, present := inputResourcesMap[res.InputPlugin]; !present {
			inputResourcesMap[res.InputPlugin] = []cluster.Resource{res}
		} else {
			inputResourcesMap[res.InputPlugin] = append(inputResourcesMap[res.InputPlugin], res)
		}
	}

	e.irmMu.Lock()
	defer e.irmMu.Unlock()

	// find the 'to be idle' input plugins
	for inputName := range e.irm {
		if _, present := inputResourcesMap[inputName]; !present {
			// will close this input
			if ir, ok := e.InputRunners[inputName]; ok {
				log.Trace("closing Input[%s]", inputName)
				ir.feedResources(nil)
			}
		}
	}

	e.irm = inputResourcesMap

	// dispatch decision to input plugins
	for inputName, rs := range inputResourcesMap {
		ir, ok := e.InputRunners[inputName]
		if !ok {
			// should never happen
			// if it happens, must be human operation fault
			panic(fmt.Sprintf("Input[%s] not found", inputName))
		}

		ir.feedResources(rs)
	}

}
