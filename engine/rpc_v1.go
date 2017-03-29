package engine

import (
	"encoding/json"
	"io"
	"net/http"

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

	buf := make([]byte, r.ContentLength)
	_, err := io.ReadFull(r.Body, buf)
	if err != nil {
		log.Error("%v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var resources []string
	if err = json.Unmarshal(buf, &resources); err != nil {
		log.Error("%v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	/*
		resourceMap := make(map[string][]string) // inputName:resources
		for _, encodedResource := range resources {
			inputName, resource := e.decodeResources(encodedResource)
			inputName, ok := e.roi[resource]
			if !ok {
				// i,e. zk might found a resource that no input is interested in
				log.Warn("[%s] resource[%s] not declared yet", e.participant, resource)
				continue
			}

			if _, present := resourceMap[inputName]; !present {
				resourceMap[inputName] = []string{resource}
			} else {
				resourceMap[inputName] = append(resourceMap[inputName], resource)
			}
		}

		// now got the new resource<->input mapping
		for inputName, resources := range resourceMap {
			ir, ok := e.InputRunners[inputName]
			if !ok {
				// should never happen
				log.Critical("%s not found", inputName)
				continue
			}

			ir.feedResources(resources)
			ir.rebalance()
		}
	*/

}
