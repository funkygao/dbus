package output

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/funkygao/assert"
)

func TestSentPos(t *testing.T) {
	p := &sentPos{}
	fn := "test"
	defer os.Remove(fn)

	p.load(fn)
	p.Log = "a"
	p.Pos = 55
	p.dump()
	b, err := ioutil.ReadFile(fn)
	assert.Equal(t, nil, err)
	assert.Equal(t, `{"log":"a","pos":55}`, string(strings.TrimSpace(string(b))))

	p.Pos = 88
	p.close() // will dump before close
	b, err = ioutil.ReadFile(fn)
	assert.Equal(t, nil, err)
	assert.Equal(t, `{"log":"a","pos":88}`, string(strings.TrimSpace(string(b))))
}

func TestSendPosLoad(t *testing.T) {
	fn := "test"
	defer os.Remove(fn)

	err := ioutil.WriteFile(fn, []byte(`{"log":"foo","pos":32}`), 0600)
	assert.Equal(t, nil, err)

	p := &sentPos{}
	p.load(fn)
	assert.Equal(t, uint32(32), p.Pos)
	assert.Equal(t, "foo", p.Log)
}
