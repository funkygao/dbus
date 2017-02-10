package engine

import (
	"fmt"

	conf "github.com/funkygao/jsconf"
)

type topology struct {
	routes map[string][]string
}

func newTopology() *topology {
	return &topology{
		routes: make(map[string][]string),
	}
}

func (t *topology) load(cf *conf.Conf) {
	for i := 0; i < len(cf.List("topology", nil)); i++ {
		section, err := cf.Section(fmt.Sprintf("topology[%d]", i))
		if err != nil {
			panic(err)
		}

		from := section.String("from", "")
		if from == "" {
			panic("empty from")
		}
		if _, present := t.routes[from]; present {
			panic("dup from: " + from)
		}

		t.routes[from] = section.StringList("to", nil)
		if len(t.routes[from]) == 0 {
			panic("empty to: " + from)
		}
	}
}
