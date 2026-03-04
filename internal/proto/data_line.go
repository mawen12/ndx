package proto

import (
	"bytes"
	"strconv"
)

type DataLine struct {
	CurNR int
	Line  string
}

func (dst *DataLine) Decode(src []byte) error {
	*dst = DataLine{}

	if len(src) < 4 {
		return &invalidMessageLenErr{messageType: "DataLine", expectedLen: 4, actualLen: len(src)}
	}
	if src[0] != ':' {
		return &invalidMessageFormatErr{messageType: "DataLine"}
	}

	src = src[1:]
	idx := bytes.IndexByte(src, ':')
	if idx == -1 {
		return &invalidMessageFormatErr{messageType: "DataLine"}
	}

	curNR, err := strconv.Atoi(string(src[:idx]))
	if err != nil {
		return &invalidMessageFormatErr{messageType: "DataLine"}
	}
	dst.CurNR = curNR

	dst.Line = string(src[idx+1:])
	return nil
}
