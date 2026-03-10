package config

import "fmt"

type parseConfigsError struct {
	conns string
	msg   string
}

func (e *parseConfigsError) Error() string {
	return fmt.Sprintf("cannot parse `%s`: %s", e.conns, e.msg)
}

type parseConfigError struct {
	conns string
	msg   string
	err   error
}

func (e *parseConfigError) Error() string {
	return fmt.Sprintf("cannot parse `%s`: %s (%s)", e.conns, e.msg, e.err.Error())
}
