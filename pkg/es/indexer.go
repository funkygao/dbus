package es

import (
	"bytes"
	"strconv"

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

func indexDoc(ev *buffer.Event) *map[string]interface{} {
	f := *ev.Fields
	host := f["host"]
	file := f["file"]
	timestamp := f["timestamp"]
	message := strconv.Quote(*ev.Text)

	delete(f, "timestamp")
	delete(f, "line")
	delete(f, "host")
	delete(f, "file")

	return &map[string]interface{}{
		"@type":        f["type"],
		"@message":     &message,
		"@source_path": file,
		"@source_host": host,
		"@timestamp":   timestamp,
		"@fields":      &f,
		"@source":      ev.Source,
	}
}

func (i *Indexer) writeBulk(index string, _type string, data interface{}) error {
	w := `{"index":{"_index":"%s","_type":"%s"}}`

	i.buffer.WriteString(fmt.Sprintf(w, index, _type))
	i.buffer.WriteByte('\n')

	switch v := data.(type) {
	case *bytes.Buffer:
		io.Copy(i.buffer, v)
	case []byte:
		i.buffer.Write(v)
	case string:
		i.buffer.WriteString(v)
	default:
		body, err := json.Marshal(data)
		if err != nil {
			log.Printf("Error writing bulk data: %v", err)
			return err
		}
		i.buffer.Write(body)
	}
	i.buffer.WriteByte('\n')
	return nil
}

func (i *Indexer) flush() {
	if i.events == 0 {
		return
	}

	log.Printf("Flushing %d event(s) to elasticsearch", i.events)
	for j := 0; j < 3; j++ {
		if err := bulkSend(i.buffer); err != nil {
			log.Printf("Failed to index event (will retry): %v", err)
			time.Sleep(time.Duration(50) * time.Millisecond)
			continue
		}
		break
	}

	i.buffer.Reset()
	i.events = 0
}

func (i *Indexer) index(ev *buffer.Event) {
	doc := indexDoc(ev)
	idx := indexName("")
	typ := (*ev.Fields)["type"]

	i.events++
	i.writeBulk(idx, typ, doc)

	if i.events < esSendBuffer {
		return
	}

	log.Printf("Flushing %d event(s) to elasticsearch", i.events)
	for j := 0; j < 3; j++ {
		if err := bulkSend(i.buffer); err != nil {
			log.Printf("Failed to index event (will retry): %v", err)
			time.Sleep(time.Duration(50) * time.Millisecond)
			continue
		}
		break
	}

	i.buffer.Reset()
	i.events = 0
}
