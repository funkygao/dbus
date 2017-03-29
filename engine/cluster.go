package engine

import (
	"fmt"
	"strings"
)

const resourceEncodeSep = " "

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
