package engine

import (
	conf "github.com/funkygao/jsconf"
)

// DAGBuilder is the builder that helps to build a pipeline DAG.
type DAGBuilder struct {
}

// NewDAGBuilder creates a DAGBuilder.
func NewDAGBuilder() *DAGBuilder {
	return &DAGBuilder{}
}

func (b *DAGBuilder) Input(name, class string) *DAGBuilder {
	return b
}

func (b *DAGBuilder) Filter(name, class string) *DAGBuilder {
	return b
}

func (b *DAGBuilder) Output(name, class string) *DAGBuilder {
	return b
}

func (b *DAGBuilder) Validate() error {
	return nil
}

func (b *DAGBuilder) CreateDAG() *conf.Conf {
	return nil
}
