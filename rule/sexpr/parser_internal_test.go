package sexpr

import (
	"bytes"
	"testing"

	"github.com/heetch/regula/rule"
	"github.com/stretchr/testify/require"
)

func TestNewParser(t *testing.T) {
	var b bytes.Buffer
	p := NewParser(&b)
	require.False(t, p.buffered)
}

// Test that we can scan a single lexicalElement from the underlying Scanner
func TestScanOneLexicalElement(t *testing.T) {
	b := bytes.NewBufferString(`hello`)
	p := NewParser(b)
	le, err := p.scan()
	require.NoError(t, err)
	require.Equal(t, SYMBOL, le.Token)
	require.Equal(t, "hello", le.Literal)
}

func TestUnscanLexicalElement(t *testing.T) {
	b := bytes.NewBufferString(`(hello)`)
	p := NewParser(b)
	le, err := p.scan()
	require.NoError(t, err)
	require.Equal(t, LPAREN, le.Token)
	require.Equal(t, "(", le.Literal)
	require.False(t, p.buffered)
	// Note, we can't assert it's not in the buffer, because it
	// is, but this is an implementation detail, what we're
	// concerned about is the behaviour we see when parsing.
	p.unscan()
	require.True(t, p.buffered)
	require.Equal(t, LPAREN, p.buf.Token)
	require.Equal(t, "(", p.buf.Literal)
	require.True(t, p.buffered)
	// Now we should see the same thing again!
	le, err = p.scan()
	require.NoError(t, err)
	require.Equal(t, LPAREN, le.Token)
	require.Equal(t, "(", le.Literal)
	require.False(t, p.buffered)
	// finally if we scan again we'll get the next symbol
	le, err = p.scan()
	require.NoError(t, err)
	require.Equal(t, SYMBOL, le.Token)
	require.Equal(t, "hello", le.Literal)
}

func TestParse(t *testing.T) {
	b := bytes.NewBufferString(`(not #true)`)
	p := NewParser(b)
	var expr rule.Expr
	var err error
	expr, err = p.Parse()
	require.NoError(t, err)
	ce1 := expr.(rule.ComparableExpression)
	expected := rule.Not(rule.BoolValue(true)).(rule.ComparableExpression)
	require.True(t, ce1.Same(expected))
}

func TestParseOperatorNonSymbolInOperatorPosition(t *testing.T) {
	b := bytes.NewBufferString(`(#false)`)
	p := NewParser(b)
	_, err := p.Parse()
	require.EqualError(t, err, `Expected an operator, but got the Boolean "false"`)
}

func TestParseOperatorNonOperatorSymbolInOperatorPosition(t *testing.T) {
	b := bytes.NewBufferString(`(wobbly)`)
	p := NewParser(b)
	_, err := p.Parse()
	require.EqualError(t, err, `Expected an operator, but got the Symbol "wobbly"`)
}

func TestParseOperatorExpectationsMatch(t *testing.T) {
	b := bytes.NewBufferString(`(= #true #true)`)
	p := NewParser(b)
	expr, err := p.Parse()
	require.NoError(t, err)
	ce := expr.(rule.ComparableExpression)
	ee := rule.Eq(rule.BoolValue(true), rule.BoolValue(true)).(rule.ComparableExpression)
	require.True(t, ee.Same(ce))
}

// GetOperatorExpr returns a TypedExpression representing the named operator.
func TestGetOperatorExprForSymbol(t *testing.T) {
	o, err := getOperatorExprForSymbol("=")
	require.NoError(t, err)
	te := o.(rule.TypedExpression)
	c := rule.Contract{
		Name:       "eq",
		ReturnType: rule.BOOLEAN,
		Terms: []rule.Term{
			{
				Type:        rule.ANY,
				Cardinality: rule.MANY,
			},
		},
	}
	require.True(t, te.Contract().Equal(c))
}
