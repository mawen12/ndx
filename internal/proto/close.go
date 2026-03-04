package proto

import "fmt"

type Close struct {
	PathPrefix string
}

func (src *Close) Encode(dst []byte) ([]byte, error) {
	buf := []byte(fmt.Sprintf("rm -rf %s\n", src.PathPrefix))
	dst = append(dst, buf...)
	return dst, nil
}
