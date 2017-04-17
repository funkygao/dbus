// Package dsn provides unified DSN scheme.
package dsn

import (
	"errors"
	"strings"
)

var ErrIllegalDSN = errors.New("illegal DSN")

// Parse extracts a unified DSN string and returns the scheme of the dsn and
// the scheme specific uri.
//
// Samples dsn:
//    kafka:local://me/foobar
//    mysql:prod://root@localhost:3306/mysql
func Parse(dsn string) (scheme string, uri string, err error) {
	tuples := strings.SplitN(dsn, ":", 2)
	if len(tuples) != 2 {
		err = ErrIllegalDSN
		return
	}

	scheme, uri = strings.TrimSpace(tuples[0]), strings.TrimSpace(tuples[1])
	return
}
