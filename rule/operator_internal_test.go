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

// consumeOperands will attempt to push a slice of operands (in the
// form of Exprs) onto the operator.  If any operand breaches the
// operators Contract the code will panic.
func TestConsumeOperands(t *testing.T) {

	// Happy case, one operand
	not := newExprNot()
	require.NotPanics(t, func() {
		not.consumeOperands(
			BoolValue(false),
		)
	})

	// Happy case, multiple operands
	and := newExprAnd()
	require.NotPanics(t, func() {
		and.consumeOperands(
			BoolValue(true),
			BoolValue(true),
			BoolValue(true),
			BoolValue(false),
		)
	})

	// Sad case, type mismatch
	not = newExprNot()
	require.PanicsWithValue(t,
		`attempt to call "not" with a String in position 1, but it requires a Boolean`,
		func() {
			not.consumeOperands(
				StringValue("🐒"),
			)
		})

	// Sad case, arity error
	not = newExprNot()
	require.PanicsWithValue(t,
		`attempted to call "not" with 2 arguments, but it requires 1 argument`,
		func() {
			not.consumeOperands(
				BoolValue(true),
				BoolValue(true),
			)
		})
}
