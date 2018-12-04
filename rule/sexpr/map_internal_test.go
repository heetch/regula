package sexpr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpMap(t *testing.T) {
	om := newOpMap()
	om.mapSymbol("👨", "man")
	om.mapSymbol("👩", "woman")

	result, err := om.getOpForSymbol("👨")
	require.NoError(t, err)
	require.Equal(t, "man", result)
	result, err = om.getOpForSymbol("👩")
	require.NoError(t, err)
	require.Equal(t, "woman", result)
	result, err = om.getSymbolForOp("man")
	require.NoError(t, err)
	require.Equal(t, "👨", result)
	result, err = om.getSymbolForOp("woman")
	require.NoError(t, err)
	require.Equal(t, "👩", result)

	_, err = om.getOpForSymbol("🐈")
	require.EqualError(t, err, `"🐈" is not a valid symbol`)
	_, err = om.getSymbolForOp("cat")
	require.EqualError(t, err, `"cat" is not a valid operator name`)
}
