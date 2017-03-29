package engine

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/funkygao/dbus/pkg/cluster"
	"github.com/funkygao/gorequest"
)

func (e *Engine) callRPC(endpoint string, resources []cluster.Resource) int {
	decision, _ := json.Marshal(resources)

	resp, _, err := gorequest.New().
		Post(fmt.Sprintf("http://%s/v1/rebalance", endpoint)).
		Set("User-Agent", "dbus").
		SendString(string(decision)).
		End()
	if err != nil {
		return http.StatusBadRequest
	}
	return resp.StatusCode
}
