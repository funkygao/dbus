package engine

import (
	"fmt"
	"strings"

	log "github.com/funkygao/log4go"
)

const resourceEncodeSep = " "

func (e *Engine) DeclareResource(inputName string, resources []string) error {
	if e.controller == nil {
		return ErrClusterDisabled
	}
	if strings.Contains(inputName, resourceEncodeSep) {
		return ErrInvalidInputName
	}

	log.Trace("%s -> %+v", inputName, resources)

	e.roiMu.Lock()
	for _, res := range resources {
		if _, present := e.roi[res]; present {
			return ErrDupResource
		}

		e.roi[res] = inputName
	}
	e.roiMu.Unlock()

	return e.controller.RegisterResources(e.encodeResources(inputName, resources))
}

func (e *Engine) encodeResources(inputName string, resources []string) []string {
	r := make([]string, len(resources))
	for i, resource := range resources {
		r[i] = fmt.Sprintf("%s%s%s", inputName, resourceEncodeSep, resource)
	}
	return r
}

func (e *Engine) decodeResources(encodedResource string) (inputName, resource string) {
	tuples := strings.Split(encodedResource, resourceEncodeSep)
	return tuples[0], tuples[1]
}
