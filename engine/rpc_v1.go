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

	for _, res := range resources {
		resource, err := e.controller.DecodeResource(res)
		// TODO
		log.Info("%s %v", resource, err)
	}

}
