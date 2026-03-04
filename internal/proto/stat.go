package proto

import (
	"bytes"
	"fmt"
	"strconv"
)

type Stat struct {
	MinuteKey string
	LineCount int
}

func (dst *Stat) Decode(src []byte) error {
	*dst = Stat{}

	if len(src) < 4 {
		return &invalidMessageLenErr{messageType: "Stat", expectedLen: 4, actualLen: len(src)}
	}
	if src[0] != ':' {
		return &invalidMessageFormatErr{messageType: "Stat", Source: src, Message: "the first byte must be :"}
	}

	src = src[1:]
	idx := bytes.LastIndexByte(src, ':')
	if idx == -1 {
		return &invalidMessageFormatErr{messageType: "Stat", Source: src, Message: "must be exist :"}
	}

	dst.MinuteKey = string(src[:idx])
	count, err := strconv.Atoi(string(src[idx+1:]))
	if err != nil {
		return &invalidMessageFormatErr{messageType: "Stat", Source: src, Message: fmt.Sprintf("Stat value must be integer, but is %s", string(src[idx+1:]))}
	}
	dst.LineCount = count
	return nil
}
