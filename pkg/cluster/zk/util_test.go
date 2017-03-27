package zk

import (
	"testing"

	"github.com/funkygao/assert"
)

func TestValidateController(t *testing.T) {
	assert.Equal(t, false, validateParticipantID("") == nil)
	assert.Equal(t, false, validateParticipantID("http://a.com") == nil)
	assert.Equal(t, false, validateParticipantID("12.1.1.1") == nil)
	assert.Equal(t, true, validateParticipantID("1.1.1.1:1222") == nil)
}
