package myslave

import (
	"testing"

	"github.com/funkygao/assert"
)

func TestParseMasterAddr(t *testing.T) {
	masterAddr := "1.2.3.4:3388"
	a, h, p, s := configMasterAddr(masterAddr)
	assert.Equal(t, "1.2.3.4", h)
	assert.Equal(t, uint16(3388), p)
	assert.Equal(t, masterAddr, a)
	assert.Equal(t, "", s)

	masterAddr = "1.2.3.4:3388/mydb"
	a, h, p, s = configMasterAddr(masterAddr)
	assert.Equal(t, "1.2.3.4", h)
	assert.Equal(t, uint16(3388), p)
	assert.Equal(t, "1.2.3.4:3388", a)
	assert.Equal(t, "mydb", s)

	masterAddr = "1.2.3.4:3388/"
	_, _, _, s = configMasterAddr(masterAddr)
	assert.Equal(t, "", s)
}
