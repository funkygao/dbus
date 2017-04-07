package model

import (
	"bytes"

	"github.com/linkedin/goavro"
)

const rowsEventAvroSchema = `
{
	"namespace": "dbus",
	"type": "record",
	"name": "RowsEvent",
	"doc": "rows event record",
	"fields": [
	    {
	    	"name": "log",
	    	"type": "string",
	    	"doc": "mysql binlog event file, never null",
	    },
	    {
	    	"name": "pos",
	    	"type": "int",
	    	"doc": "mysql binlog event offset, never null",
	    },
	    {
	    	"name": "db",
	    	"type": "string",
	    	"doc": "database name, never null",
	    },
	    {
	    	"name": "tbl",
	    	"type": "string",
	    	"doc": "table name, never null",
	    },
	    {
	    	"name": "dml",
	    	"type": "string",
	    	"doc": "event type: <I|U|D> I=insert, U=update, D=delete",
	    },
	    {
	    	"name": "ts",
	    	"type": "int",
	    	"doc": "the binlog event timestamp, never null",
	    },
	    {
	    	"name": "rows",
	    	"type": {
	    		"type": "array",
	    		"items": "bytes",
	    	}
	    	"doc": "list of rows",
	    },
	]
}
`

func avroMarshaller(v interface{}) ([]byte, error) {
	codec, err := goavro.NewCodec(rowsEventAvroSchema)
	if err != nil {
		return nil, err
	}

	w := bytes.NewBuffer(nil) // TODO reuse memory

	if err = codec.Encode(w, v); err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}
