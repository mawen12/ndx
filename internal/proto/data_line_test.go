package proto_test

import (
	"testing"

	"github.com/mawen12/ndx/internal/proto"
	"github.com/stretchr/testify/require"
)

func Test_DataLineDecode(t *testing.T) {
	bytes := []byte(":123:Hello, World!")

	var dataLine proto.DataLine
	require.NoError(t, dataLine.Decode(bytes))

	require.Equal(t, 123, dataLine.CurNR)
	require.Equal(t, "Hello, World!", dataLine.Line)
}
