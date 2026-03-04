package proto

import "bytes"

type ParameterStatus struct {
	Name  string
	Value string
}

func (dst *ParameterStatus) Decode(src []byte) error {
	*dst = ParameterStatus{}

	if len(src) < 4 {
		return &invalidMessageLenErr{messageType: "ParameterStatus", expectedLen: 4, actualLen: len(src)}
	}
	if src[0] != ':' {
		return &invalidMessageFormatErr{messageType: "ParameterStatus"}
	}
	src = src[1:]
	idx := bytes.IndexByte(src, ':')
	if idx == -1 {
		return &invalidMessageFormatErr{messageType: "ParameterStatus"}
	}

	dst.Name = string(src[:idx])
	dst.Value = string(src[idx+1:])
	return nil
}
