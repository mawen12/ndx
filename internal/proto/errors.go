package proto

import "fmt"

type invalidMessageFormatErr struct {
	messageType string
	Source      []byte
	Message     string
}

func (e *invalidMessageFormatErr) Error() string {
	return fmt.Sprintf("%s body is invalid", e.messageType)
}

type invalidMessageLenErr struct {
	messageType string
	expectedLen int
	actualLen   int
}

func (e *invalidMessageLenErr) Error() string {
	return fmt.Sprintf("%s body must have length of %d, but it is %d", e.messageType, e.expectedLen, e.actualLen)
}
