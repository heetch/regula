package rule

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// GetOperator allows us to get an instance of a private expression by its name.
func TestGetOperator(t *testing.T) {
	ex, err := GetOperator("eq")
	require.NoError(t, err)
	_, ok := ex.(*exprEq)
	require.True(t, ok)

	ex, err = GetOperator("not")
	require.NoError(t, err)
	_, ok = ex.(*exprNot)
	require.True(t, ok)

	ex, err = GetOperator("and")
	require.NoError(t, err)
	_, ok = ex.(*exprAnd)
	require.True(t, ok)

	ex, err = GetOperator("or")
	require.NoError(t, err)
	_, ok = ex.(*exprOr)
	require.True(t, ok)

	ex, err = GetOperator("in")
	require.NoError(t, err)
	_, ok = ex.(*exprIn)
	require.True(t, ok)

}
