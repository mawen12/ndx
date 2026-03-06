package model

type Message interface{}

type FrontendMessage interface {
	Message
	Encode(dst []byte) ([]byte, error)
}

type BackendMessage interface {
	Message
	Decode(data []byte) error
}
