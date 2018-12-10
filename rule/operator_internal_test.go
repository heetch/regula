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

// When we pass both Int64 and Float64 as "NUMBER" types to fulfil one
// repeated Term of an operators Contract, then the Int64 will be
// converted to a Float64
func TestHomogeniseOperator(t *testing.T) {
	op := operator{
		contract: Contract{
			OpCode:     "goo",
			ReturnType: BOOLEAN,
			Terms: []Term{
				{
					Type:        NUMBER,
					Cardinality: MANY,
					Min:         2,
				},
			},
		},
	}
	err := op.PushExpr(Int64Value(1))
	require.NoError(t, err)
	err = op.PushExpr(Float64Value(1.1))
	require.NoError(t, err)
	op.homogenise()
	expected := IntToFloat(Int64Value(1.0)).(ComparableExpression)
	require.True(t, op.operands[0].(ComparableExpression).Same(expected))
	exp := Float64Value(1.1)
	require.True(t, op.operands[1].(ComparableExpression).Same(exp))
}

// operator.homogenise will promote the ReturnType of the operator
// from NUMBER to the type of its parameters if they defined in the
// Contract as a Term with Cardinality == MANY and Type == NUMBER.
// This is intended to allow us to define mathematical operators that
// can accept mixed type arguments.
func TestHomogenisedOperatorPromotesReturnTypeWhenNumber(t *testing.T) {
	op := operator{
		contract: Contract{
			OpCode:     "plus",
			ReturnType: NUMBER,
			Terms: []Term{
				{
					Type:        NUMBER,
					Cardinality: MANY,
					Min:         2,
				},
			},
		},
	}
	err := op.PushExpr(Int64Value(1))
	require.NoError(t, err)
	err = op.PushExpr(Float64Value(1.1))
	require.NoError(t, err)
	op.homogenise()
	require.Equal(t, FLOAT, op.Contract().ReturnType)
}
