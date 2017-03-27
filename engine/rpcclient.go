package engine

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/funkygao/gorequest"
)

func (e *Engine) callRPC(participentID string, resources []string) int {
	decision, _ := json.Marshal(resources)

	resp, _, err := gorequest.New().
		Post(fmt.Sprintf("http://%s/v1/rebalance", participentID)).
		Set("User-Agent", "dbus").
		SendString(string(decision)).
		End()
	if err != nil {
		return http.StatusBadRequest
	}
	return resp.StatusCode
}
