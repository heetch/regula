package rule_test

import (
	"fmt"
	"testing"

	"github.com/heetch/regula/rule"
	"github.com/stretchr/testify/require"
)

// IsFulfilledBy indicates whether a given TypedExpression fulfils a Term.
func TestTermIsFulfilledBy(t *testing.T) {
	// We'll express a few fundamental TypedExpressions, to avoid repetition in the test cases.
	boolean := rule.BoolValue(true)
	integer := rule.Int64Value(1)
	float := rule.Float64Value(1.1)
	str := rule.StringValue("foo")
	boolParam := rule.BoolParam("foo")
	stringParam := rule.StringParam("foo")
	intParam := rule.Int64Param("foo")
	floatParam := rule.Float64Param("foo")
	not := rule.Not(boolean).(rule.TypedExpression)
	or := rule.Or(boolean, boolean).(rule.TypedExpression)
	and := rule.And(boolean, boolean).(rule.TypedExpression)
	eq := rule.Eq(boolean, boolean).(rule.TypedExpression)
	in := rule.In(boolean, boolean, boolean).(rule.TypedExpression)

	testCases := []struct {
		// Test cases define a list of positive expressions
		// (those that fulfil the Term), and a list of
		// negative expressions (those that do not fulfil the
		// Term).  We are attempting to be exhaustive here.
		name                string
		positiveExpressions []rule.TypedExpression
		negativeExpressions []rule.TypedExpression
		term                rule.Term
	}{
		{
			name: "Boolean",
			positiveExpressions: []rule.TypedExpression{
				boolean, boolParam, not, or, and, eq, in},
			negativeExpressions: []rule.TypedExpression{
				str, integer, float, stringParam, intParam, floatParam},
			term: rule.Term{Type: rule.BOOLEAN},
		},
		{
			name:                "String",
			positiveExpressions: []rule.TypedExpression{str, stringParam},
			negativeExpressions: []rule.TypedExpression{
				boolean, integer, float, boolParam, intParam, floatParam,
				or, and, eq, in, not},
			term: rule.Term{Type: rule.STRING},
		},
		{
			name:                "Integer",
			positiveExpressions: []rule.TypedExpression{integer, intParam},
			negativeExpressions: []rule.TypedExpression{
				boolean, str, float, boolParam, stringParam, floatParam,
				or, and, eq, in, not},
			term: rule.Term{Type: rule.INTEGER},
		},
		{
			name:                "Float",
			positiveExpressions: []rule.TypedExpression{float, floatParam},
			negativeExpressions: []rule.TypedExpression{
				boolean, str, integer, boolParam, stringParam, intParam,
				or, and, eq, in, not},
			term: rule.Term{Type: rule.FLOAT},
		},
		{
			name: "Number",
			positiveExpressions: []rule.TypedExpression{
				integer, intParam, float, floatParam},
			negativeExpressions: []rule.TypedExpression{
				boolean, str, boolParam, stringParam, or, and, eq, not,
			},
			term: rule.Term{Type: rule.NUMBER},
		},
		{
			name: "Any",
			positiveExpressions: []rule.TypedExpression{
				integer, intParam, float, floatParam,
				boolean, str, boolParam, stringParam,
				or, and, eq, not,
			},
			negativeExpressions: nil,
			term:                rule.Term{Type: rule.ANY},
		},
	}

	// Run "IsFullfilledBy" for each test case with each positive and negative expression.
	for i, tc := range testCases {
		for j, pc := range tc.positiveExpressions {
			t.Run(fmt.Sprintf("%s[%d] positive case %d", tc.name, i, j),
				func(t *testing.T) {
					require.True(t, tc.term.IsFulfilledBy(pc))
				})
		}
		for k, nc := range tc.negativeExpressions {
			t.Run(fmt.Sprintf("%s[%d] negative case %d", tc.name, i, k),
				func(t *testing.T) {
					require.False(t, tc.term.IsFulfilledBy(nc))
				})
		}
	}
}

func TestPushExprEnforcesArity(t *testing.T) {
	// Happy case (one expected, one given)
	not, err := rule.GetOperatorExpr("not")
	require.NoError(t, err)
	err = not.PushExpr(rule.BoolValue(true))
	require.NoError(t, err)
	err = not.Finalise()
	require.NoError(t, err)

	// Happy case (many expected)
	and, err := rule.GetOperatorExpr("and")
	require.NoError(t, err)
	err = and.PushExpr(rule.BoolValue(true))
	require.NoError(t, err)
	err = and.PushExpr(rule.BoolValue(true))
	require.NoError(t, err)
	err = and.PushExpr(rule.BoolValue(true))
	require.NoError(t, err)
	err = and.PushExpr(rule.BoolValue(true))
	require.NoError(t, err)
	//... we could go on until we run out of memory ☺
	err = and.Finalise()
	require.NoError(t, err)

	// Sad case (one two many operands - one expected, two given)
	not, err = rule.GetOperatorExpr("not")
	err = not.PushExpr(rule.BoolValue(true))
	require.NoError(t, err)
	// We already pushed a bool onto this 'not', and it only wants one operand.
	err = not.PushExpr(rule.BoolValue(true))
	require.EqualError(t, err, `attempted to call "not" with 2 arguments, but it requires 1 argument`)
	ae, ok := err.(rule.ArityError)
	require.True(t, ok)
	require.Equal(t, "not", ae.OpCode)
	require.Equal(t, 1, ae.MaxPos)
	require.Equal(t, 2, ae.ErrorPos)

	// Sad case (not enough operands, expected 2, but got 1)
	and, err = rule.GetOperatorExpr("and")
	require.NoError(t, err)
	err = and.PushExpr(rule.BoolValue(true))
	err = and.Finalise()
	require.EqualError(t, err, `attempted to call "and" with 1 argument, but it requires 2 arguments`)
	ae, ok = err.(rule.ArityError)
	require.True(t, ok)
	require.Equal(t, "and", ae.OpCode)
	require.Equal(t, 2, ae.MinPos)
	require.Equal(t, 1, ae.ErrorPos)

	// Sad case (not enough operands expected 1, got 0)
	not, err = rule.GetOperatorExpr("not")
	require.NoError(t, err)
	err = not.Finalise()
	require.EqualError(t, err, `attempted to call "not" with 0 arguments, but it requires 1 argument`)
	ae, ok = err.(rule.ArityError)
	require.True(t, ok)
	require.Equal(t, "not", ae.OpCode)
	require.Equal(t, 1, ae.MinPos)
	require.Equal(t, 0, ae.ErrorPos)

}
