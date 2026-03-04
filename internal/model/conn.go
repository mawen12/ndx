package model

import "io"

type Conn interface {
	Readout() (string, error)
	io.Writer
	io.Closer
}
