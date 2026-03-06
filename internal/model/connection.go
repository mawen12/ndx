package model

import "io"

type Connection interface {
	io.Writer
	io.Closer
	Readout() (string, error)
}
