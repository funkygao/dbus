package model

import (
	"bytes"

	"github.com/linkedin/goavro"
)

func avroMarshaller(v interface{}) ([]byte, error) {
	someRecordSchemaJson := `{"type":"record","name":"Foo","fields":[{"name":"field1","type":"int"},{"name":"field2","type":"string","default":"happy"}]}`
	codec, err := goavro.NewCodec(someRecordSchemaJson)
	if err != nil {
		return nil, err
	}

	w := bytes.NewBuffer(nil) // TODO reuse memory

	if err = codec.Encode(w, v); err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}
