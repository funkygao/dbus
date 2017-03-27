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

func (e *Engine) doRebalance(w http.ResponseWriter, r *http.Request) {
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

	e.roiMu.RLock()
	defer e.roiMu.RUnlock()

	for _, encodedResource := range resources {
		resource, err := e.controller.DecodeResource(encodedResource)
		if err != nil {
			// FIXME
			log.Error("[%s] {%s}: %v", e.participantID, encodedResource, err)
			continue
		}

		inputName, ok := e.roi[resource]
		if !ok {
			log.Warn("[%s] resource[%s] not declared yet", e.participantID, resource)
			continue
		}

		// TODO
		log.Info("%s %v", resource, inputName)
	}

}
