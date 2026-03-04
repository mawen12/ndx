package proto

type ReadyForQuery struct {
}

func (dst *ReadyForQuery) Decode(src []byte) error {
	*dst = ReadyForQuery{}

	if len(src) < 2 {
		return &invalidMessageLenErr{messageType: "ReadyForQuery", expectedLen: 2, actualLen: len(src)}
	}
	if src[0] != ':' {
		return &invalidMessageFormatErr{messageType: "ReadyForQuery"}
	}
	return nil
}
