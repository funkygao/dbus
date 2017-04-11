package cluster

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
)

// State is state of a participant in a cluster.
type State uint8

const (
	StateUnknown State = 0
	StateOnline  State = 1
	StateOffline State = 2
)

var stateText = map[State]string{
	StateUnknown: "unknown",
	StateOnline:  "online",
	StateOffline: "offline",
}

// Participant is a live node that can get Resource assignment from controller leader.
type Participant struct {
	Endpoint string `json:"endpoint,omitempty"`
	Weight   int    `json:"weight"`
	State    State  `json:"state,omitempty"`
	Revision string `json:"revision,omitempty"`
	APIPort  int    `json:"api_port,omitempty"`
}

func (p *Participant) Marshal() []byte {
	b, _ := json.Marshal(p)
	return b
}

func (p *Participant) From(data []byte) {
	json.Unmarshal(data, p)
}

// AccceptResources returns whether the participant is able to accespt assigned resources from leader.
func (p *Participant) AccceptResources() bool {
	return p.State == StateOnline
}

func (p *Participant) RPCEndpoint() string {
	return fmt.Sprintf("http://%s", p.Endpoint)
}

func (p *Participant) APIEndpoint() string {
	host, _, _ := net.SplitHostPort(p.Endpoint)
	return fmt.Sprintf("http://%s:%d", host, p.APIPort)
}

func (p *Participant) StateText() string {
	return stateText[p.State]
}

func (p Participant) String() string {
	return p.Endpoint
}

func (p Participant) Equals(that Participant) bool {
	return p.Endpoint == that.Endpoint
}

func (p *Participant) Valid() bool {
	if strings.Contains(p.Endpoint, "/") {
		return false
	}

	host, port, err := net.SplitHostPort(p.Endpoint)
	if err != nil {
		return false
	}

	if len(host) == 0 || len(port) == 0 /*|| p.APIPort == 0 */ {
		return false
	}

	return true
}

type Participants []Participant

func (ps Participants) Len() int {
	return len(ps)
}

func (ps Participants) Less(i, j int) bool {
	return ps[i].Endpoint < ps[j].Endpoint
}

func (ps Participants) Swap(i, j int) {
	ps[i], ps[j] = ps[j], ps[i]
}
