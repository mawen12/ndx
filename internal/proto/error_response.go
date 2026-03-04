package proto

type ErrorResponse struct {
	Code    byte
	Message string
}

func (dst *ErrorResponse) Decode(src []byte) error {
	*dst = ErrorResponse{}

	if len(src) < 4 {
		return &invalidMessageLenErr{messageType: "ErrorResponse", expectedLen: 4, actualLen: len(src)}
	}
	if src[0] != ':' {
		return &invalidMessageFormatErr{messageType: "ErrorResponse"}
	}
	if src[2] != ':' {
		return &invalidMessageFormatErr{messageType: "ErrorResponse"}
	}

	dst.Code = src[1]
	dst.Message = string(src[3:])
	return nil
}
