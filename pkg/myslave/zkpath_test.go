package myslave

import (
	"strings"
	"testing"

	"github.com/funkygao/assert"
)

func TestZkPath(t *testing.T) {
	mysqlMasterAddr := "localhost:3306"
	assert.Equal(t, "/dbus/pkg/myslave/localhost:3306", posPath(mysqlMasterAddr))
	assert.Equal(t, "/dbus/pkg/myslave/localhost:3306/owner", masterPath(mysqlMasterAddr))
	assert.Equal(t, true, strings.HasPrefix(myNodePath(mysqlMasterAddr), "/dbus/pkg/myslave/localhost:3306/ids"))
}
