package engine

import (
	conf "github.com/funkygao/jsconf"
)

/*
DAGBuilder is the builder that helps to build a pipeline DAG.

Example:

	package main

	import (
		"time"

		"github.com/funkygao/dbus/engine"
	)

	func main() {
		builder := engine.NewDAGBuilder()
		builder.Input("in.mock", "MockInput").
		    Filter("filter.mock", "MockFilter", "in.mock").
			Output("out.mock", "MockOutput", "filter.mock").
			    SetDuration("sleep", time.Second)

		engine.New(nil).SubmitDAG(builder.CreateDAG()).ServeForever()
	}
*/
type DAGBuilder struct {
	*conf.Conf
}

// NewDAGBuilder creates a DAGBuilder.
func NewDAGBuilder() *DAGBuilder {
	return &DAGBuilder{}
}

func (b *DAGBuilder) Input(name, class string) *DAGBuilder {
	return b
}

func (b *DAGBuilder) Filter(name, class string, matches ...string) *DAGBuilder {
	return b
}

func (b *DAGBuilder) Output(name, class string, matches ...string) *DAGBuilder {
	return b
}

func (b *DAGBuilder) Validate() error {
	return nil
}

func (b *DAGBuilder) BuildDAG() *conf.Conf {
	if err := b.Validate(); err != nil {
		panic(err)
	}

	return nil
}
