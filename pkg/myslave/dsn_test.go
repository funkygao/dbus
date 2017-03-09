package myslave

import (
	"testing"

	"github.com/funkygao/assert"
)

func TestParseDSN(t *testing.T) {
	dsn := "mysql://user1:pass1@1.1.1.1:3306/"
	host, port, user, pass, dbs, err := parseDSN(dsn)
	assert.Equal(t, nil, err)
	assert.Equal(t, "user1", user)
	assert.Equal(t, "pass1", pass)
	assert.Equal(t, uint16(3306), port)
	assert.Equal(t, "1.1.1.1", host)
	assert.Equal(t, 0, len(dbs))

	// parseDSN auto fill the schema
	dsn = "user1:pass1@1.1.1.1:3306"
	host, port, user, pass, dbs, err = parseDSN(dsn)
	assert.Equal(t, nil, err)
	assert.Equal(t, "user1", user)
	assert.Equal(t, "pass1", pass)
	assert.Equal(t, uint16(3306), port)
	assert.Equal(t, "1.1.1.1", host)
	assert.Equal(t, 0, len(dbs))

	// missing port: error
	dsn = "user1:pass1@1.1.1.1"
	host, port, user, pass, dbs, err = parseDSN(dsn)
	assert.Equal(t, false, err == nil)

	// empty password: ok
	dsn = "user1:@1.1.1.1:3306"
	host, port, user, pass, dbs, err = parseDSN(dsn)
	assert.Equal(t, nil, err)
	assert.Equal(t, "user1", user)
	assert.Equal(t, "", pass)
	assert.Equal(t, uint16(3306), port)
	assert.Equal(t, "1.1.1.1", host)
	assert.Equal(t, 0, len(dbs))

	// 1 db in DSN
	dsn = "user1:@1.1.1.1:3306/db1"
	host, port, user, pass, dbs, err = parseDSN(dsn)
	assert.Equal(t, 1, len(dbs))
	assert.Equal(t, "db1", dbs[0])

	// 2 dbs in DSN
	dsn = "user1:@1.1.1.1:3306/db1,db2"
	host, port, user, pass, dbs, err = parseDSN(dsn)
	assert.Equal(t, 2, len(dbs))
	assert.Equal(t, "db1", dbs[0])
	assert.Equal(t, "db2", dbs[1])

	// 3 dbs in DSN, and auto trim space
	dsn = "user1:@1.1.1.1:3306/db1, db2,db3"
	host, port, user, pass, dbs, err = parseDSN(dsn)
	assert.Equal(t, 3, len(dbs))
	assert.Equal(t, "db1", dbs[0])
	assert.Equal(t, "db2", dbs[1])
	assert.Equal(t, "db3", dbs[2])

	// 4 dbs in DSN, and 1 empty slot
	dsn = "user1:@1.1.1.1:3306/db1, db2,db3,,db4"
	host, port, user, pass, dbs, err = parseDSN(dsn)
	assert.Equal(t, 4, len(dbs))
	assert.Equal(t, "db1", dbs[0])
	assert.Equal(t, "db2", dbs[1])
	assert.Equal(t, "db3", dbs[2])
	assert.Equal(t, "db4", dbs[3])
}
