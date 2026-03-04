package proto

type NoticeResponse struct {
	Message string
}

func (dst *NoticeResponse) Decode(src []byte) error {
	*dst = NoticeResponse{}

	if len(src) < 2 {
		return &invalidMessageLenErr{messageType: "NoticeResponse", expectedLen: 2, actualLen: len(src)}
	}

	if src[0] != ':' {
		return &invalidMessageFormatErr{messageType: "NoticeResponse"}
	}

	dst.Message = string(src[1:])
	return nil
}
