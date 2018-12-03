package rule

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// GetOperatorExpr allows us to get an instance of a private expression by its name.
func TestGetOperatorExpr(t *testing.T) {
	ex, err := GetOperatorExpr("eq")
	require.NoError(t, err)
	_, ok := ex.(*exprEq)
	require.True(t, ok)

	ex, err = GetOperatorExpr("not")
	require.NoError(t, err)
	_, ok = ex.(*exprNot)
	require.True(t, ok)

	ex, err = GetOperatorExpr("and")
	require.NoError(t, err)
	_, ok = ex.(*exprAnd)
	require.True(t, ok)

	ex, err = GetOperatorExpr("or")
	require.NoError(t, err)
	_, ok = ex.(*exprOr)
	require.True(t, ok)

	ex, err = GetOperatorExpr("in")
	require.NoError(t, err)
	_, ok = ex.(*exprIn)
	require.True(t, ok)

}
