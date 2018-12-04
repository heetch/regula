package sexpr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpMap(t *testing.T) {
	om := newOpCodeMap()
	om.mapSymbol("👨", "man")
	om.mapSymbol("👩", "woman")

	result, err := om.getOpCodeForSymbol("👨")
	require.NoError(t, err)
	require.Equal(t, "man", result)
	result, err = om.getOpCodeForSymbol("👩")
	require.NoError(t, err)
	require.Equal(t, "woman", result)
	result, err = om.getSymbolForOpCode("man")
	require.NoError(t, err)
	require.Equal(t, "👨", result)
	result, err = om.getSymbolForOpCode("woman")
	require.NoError(t, err)
	require.Equal(t, "👩", result)

	_, err = om.getOpCodeForSymbol("🐈")
	require.EqualError(t, err, `"🐈" is not a valid symbol`)
	_, err = om.getSymbolForOpCode("cat")
	require.EqualError(t, err, `"cat" is not a valid operator name`)
}
