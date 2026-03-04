package model

type StringReader interface {
	Next() (string, error)
}
