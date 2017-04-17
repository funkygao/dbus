package dsn

import (
	"testing"

	"github.com/funkygao/assert"
)

func TestParse(t *testing.T) {
	scheme, uri, err := Parse("illegal")
	assert.Equal(t, false, err == nil)

	scheme, uri, err = Parse("kafka:local://me/foobar")
	assert.Equal(t, nil, err)
	assert.Equal(t, "kafka", scheme)
	assert.Equal(t, "local://me/foobar", uri)

	// test trim spaces
	scheme, uri, err = Parse("kafka : local://me/foobar")
	assert.Equal(t, nil, err)
	assert.Equal(t, "kafka", scheme)
	assert.Equal(t, "local://me/foobar", uri)

	scheme, uri, err = Parse("mysql:prod://root@localhost:3306/mysql")
	assert.Equal(t, nil, err)
	assert.Equal(t, "mysql", scheme)
	assert.Equal(t, "prod://root@localhost:3306/mysql", uri)
}
