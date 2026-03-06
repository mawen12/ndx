package logclient

import (
	"fmt"
	"strings"

	config2 "github.com/mawen12/ndx/internal/config"
)

type connectError struct {
	config *config2.Config
	msg    string
	err    error
}

func (e *connectError) Error() string {
	sb := &strings.Builder{}
	fmt.Fprintf(sb, "failed to connect to `host=%s user=%s` : %s", e.config.Host, e.config.User, e.msg)
	if e.err != nil {
		fmt.Fprintf(sb, " (%s)", e.err.Error())
	}
	return sb.String()
}

func (e *connectError) Unwrap() error {
	return e.err
}

type connLockError struct {
	status string
}

func (e *connLockError) SafeToRetry() bool {
	return true
}

func (e *connLockError) Error() string {
	return e.status
}

type writeError struct {
	err         error
	safeToRetry bool
}

func (e *writeError) SafeToRetry() bool {
	return e.safeToRetry
}

func (e *writeError) Error() string {
	return fmt.Sprintf("write failed: %s", e.err.Error())
}
