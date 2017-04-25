package binlog

import (
	"testing"

	"github.com/funkygao/assert"
)

func TestBinlogState(t *testing.T) {
	s := New("", "")
	s.File = "f1"
	s.Offset = 5
	assert.Equal(t, `{"file":"f1","offset":5}`, string(s.Marshal()))

	s.reset()
	assert.Equal(t, "", s.File)
	assert.Equal(t, uint32(0), s.Offset)

	s.Unmarshal([]byte(`{"file":"f1","offset":5}`))
	assert.Equal(t, "f1", s.File)
	assert.Equal(t, uint32(5), s.Offset)

	s1 := New("", "")
	s1.File = "f1"
	s1.Offset = 15
	assert.Equal(t, "10", s1.Delta(s))
	s1.File = "f2"
	assert.Equal(t, "", s1.Delta(s))
}
