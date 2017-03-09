package myslave

import (
	"testing"

	"github.com/funkygao/assert"
)

func TestParseDSN(t *testing.T) {
	dsn := "mysql://user1:pass1@1.1.1.1:3306"
	host, port, user, pass, err := parseDSN(dsn)
	assert.Equal(t, nil, err)
	assert.Equal(t, "user1", user)
	assert.Equal(t, "pass1", pass)
	assert.Equal(t, uint16(3306), port)
	assert.Equal(t, "1.1.1.1", host)

	// parseDSN auto fill the schema
	dsn = "user1:pass1@1.1.1.1:3306"
	host, port, user, pass, err = parseDSN(dsn)
	assert.Equal(t, nil, err)
	assert.Equal(t, "user1", user)
	assert.Equal(t, "pass1", pass)
	assert.Equal(t, uint16(3306), port)
	assert.Equal(t, "1.1.1.1", host)

	// missing port: error
	dsn = "user1:pass1@1.1.1.1"
	host, port, user, pass, err = parseDSN(dsn)
	assert.Equal(t, false, err == nil)

	// missing password: ok
	dsn = "user1:@1.1.1.1:3306"
	host, port, user, pass, err = parseDSN(dsn)
	assert.Equal(t, nil, err)
	assert.Equal(t, "user1", user)
	assert.Equal(t, "", pass)
	assert.Equal(t, uint16(3306), port)
	assert.Equal(t, "1.1.1.1", host)

}
