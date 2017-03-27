package zk

import (
	"net"
	"strings"

	"github.com/funkygao/dbus/pkg/cluster"
)

func validateParticipantID(participantID string) error {
	host, port, err := net.SplitHostPort(participantID)
	if err != nil {
		return err
	}

	if strings.Contains(participantID, "/") {
		// participantID will part of the znode path, '/' not allowed
		return cluster.ErrInvalidParticipantID
	}

	if len(host) == 0 || len(port) == 0 {
		return cluster.ErrInvalidParticipantID
	}

	return nil
}
