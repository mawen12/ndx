package proto_test

import (
	"testing"

	"github.com/mawen12/ndx/internal/proto"
	"github.com/stretchr/testify/require"
)

func Test_CloseEncode(t *testing.T) {
	var close proto.Close
	dst, err := close.Encode(nil)
	require.NoError(t, err)

	expected := "rm -rf \n"
	require.Equal(t, expected, string(dst))
}
