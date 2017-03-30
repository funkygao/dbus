package cluster

import (
	"encoding/json"
	"net"
	"strings"
)

// Participant is a live node that can get Resource assignment from controller leader.
type Participant struct {
	Endpoint string `json:"endpoint,omitempty"`
	Weight   int    `json:"weight"`
}

func (p *Participant) Marshal() []byte {
	b, _ := json.Marshal(p)
	return b
}

func (p *Participant) From(data []byte) {
	json.Unmarshal(data, p)
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

	if len(host) == 0 || len(port) == 0 {
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
