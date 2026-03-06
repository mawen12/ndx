package config

import "fmt"

type parseConfigsError struct {
	connStrings string
	msg         string
}

func (e *parseConfigsError) Error() string {
	return fmt.Sprintf("cannot parse `%s`: %s", e.connStrings, e.msg)
}

type parseConfigError struct {
	connString string
	msg        string
	err        error
}

func (e *parseConfigError) Error() string {
	return fmt.Sprintf("cannot parse `%s`: %s (%s)", e.connString, e.msg, e.err.Error())
}
