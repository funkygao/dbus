package es

import (
	"bytes"

	"github.com/hailocab/elastigo/api"
)

type Indexer struct {
	events int
	buffer *bytes.Buffer
}

func NewIndexer() *Indexer {
	return &Indexer{}
}

func bulkSend(buf *bytes.Buffer) error {
	_, err := api.DoCommand("POST", "/_bulk", buf)
	if err != nil {
		return err
	}

	return nil
}
