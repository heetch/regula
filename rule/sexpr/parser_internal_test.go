package sexpr

import (
	"bytes"
	"testing"

	"github.com/heetch/regula/rule"
	"github.com/stretchr/testify/require"
)

// The symbol map allows us to do bidirectional conversion of symbols
func TestMakeSymbolMap(t *testing.T) {
	sm := makeSymbolMap()
	o, err := sm.getOperatorForSymbol("=")
	require.NoError(t, err)
	require.Equal(t, "eq", o.Contract().OpCode)
}

// NewParser initialises the Parser with an io.Reader and no buffered lexicalElement
func TestNewParser(t *testing.T) {
	var b bytes.Buffer
	p := NewParser(&b)
	require.False(t, p.buffered)
}

// scan will scan a single lexicalElement from the underlying Scanner
func TestScanOneLexicalElement(t *testing.T) {
	b := bytes.NewBufferString(`hello`)
	p := NewParser(b)
	le, err := p.scan()
	require.NoError(t, err)
	require.Equal(t, SYMBOL, le.Token)
	require.Equal(t, "hello", le.Literal)
}

// unscan will allow the same lexicalElement to be rescanned from the buffer
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

// Parse can return a tree of Exprs representing a simple operator and its operand.
func TestParseSimpleOperator(t *testing.T) {
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

// Parse can return a tree of Exprs representing operators with other operators amongst their operands
func TestParseNestedOperator(t *testing.T) {
	b := bytes.NewBufferString(`
(not (= #true
        (or #false
            #true
            (and #true 
                 #true))))
`)
	p := NewParser(b)
	var expr rule.Expr
	var err error
	expr, err = p.Parse()
	require.NoError(t, err)
	ce1 := expr.(rule.ComparableExpression)
	expected := rule.Not(
		rule.Eq(
			rule.BoolValue(true),
			rule.Or(
				rule.BoolValue(false),
				rule.BoolValue(true),
				rule.And(
					rule.BoolValue(true),
					rule.BoolValue(true),
				),
			),
		),
	).(rule.ComparableExpression)
	require.True(t, ce1.Same(expected))
}

// Parse returns an error if it encounters EOF
func TestParserReturnsErrorIfItHitsEOF(t *testing.T) {
	b := bytes.NewBufferString(``)
	p := NewParser(b)
	var err error
	_, err = p.Parse()
	require.EqualError(t, err, `1:0: Error. unexpected end of file.`)

}

// Parse returns an error if the top level expression doesn't return a Boolean value
func TestParserReturnsErrorIfRootOfRuleIsNonBoolean(t *testing.T) {
	b := bytes.NewBufferString(`"eek"`)
	p := NewParser(b)
	var err error
	_, err = p.Parse()
	require.EqualError(t, err, `0:0: Type error. The root expression in a rule must return a Boolean, but it returns String.`)
}

// Parse returns an error if an operator doesn't follow the left parenthesis
func TestParseOperatorNonSymbolInOperatorPosition(t *testing.T) {
	b := bytes.NewBufferString(`(#false)`)
	p := NewParser(b)
	_, err := p.Parse()
	require.EqualError(t, err, `Expected an operator, but got the boolean "false"`)
}

// Parse returns an error if a symbol that is not an operator follows the left parenthesis
func TestParseOperatorNonOperatorSymbolInOperatorPosition(t *testing.T) {
	b := bytes.NewBufferString(`(wobbly)`)
	p := NewParser(b)
	_, err := p.Parse()
	require.EqualError(t, err, `"wobbly" is not a valid symbol`)
}

// makeBoolValue correctly constructs a BoolValue
func TestMakeBoolValue(t *testing.T) {
	b := bytes.NewBufferString(`#true #false`)
	p := NewParser(b)
	le, err := p.scan()
	require.NoError(t, err)
	bv := p.makeBoolValue(le)
	require.True(t, bv.(rule.ComparableExpression).Same(
		rule.BoolValue(true),
	))
	le, err = p.scan()
	require.NoError(t, err)
	bv = p.makeBoolValue(le)
	require.True(t, bv.(rule.ComparableExpression).Same(
		rule.BoolValue(false),
	))
}

//makeNumber constructs appropriate numeric value types with valid input
func TestMakeNumberHappyCases(t *testing.T) {
	cases := []struct {
		Name     string
		Input    string
		Expected rule.Expr
	}{
		{
			Name:     "positive integer",
			Input:    "123",
			Expected: rule.Int64Value(123),
		},
		{
			Name:     "negative integer",
			Input:    "-20",
			Expected: rule.Int64Value(-20),
		},
		{
			Name:     "positive float",
			Input:    "12.345",
			Expected: rule.Float64Value(12.345),
		},
		{
			Name:     "negative float",
			Input:    "-123.45",
			Expected: rule.Float64Value(-123.45),
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			b := bytes.NewBufferString(c.Input)
			p := NewParser(b)
			le, err := p.scan()
			require.NoError(t, err)
			expr, err := p.makeNumber(le)
			require.NoError(t, err)
			ce := expr.(rule.ComparableExpression)
			exp := c.Expected.(rule.ComparableExpression)
			require.True(t, ce.Same(exp))
		})

	}
}

//makeNumber returns an error on invalid input (where that input makes it past the lexer
func TestMakeNumberSadCases(t *testing.T) {
	cases := []struct {
		Name  string
		Input string
		Error string
	}{
		{
			Name:  "Missing whole part in negative",
			Input: "-.1",
			Error: "1:0: Error. strconv.ParseInt: parsing \"-\": invalid syntax.",
		},
		// This space reserved for any other case we can think of!
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			b := bytes.NewBufferString(c.Input)
			p := NewParser(b)
			le, err := p.scan()
			require.NoError(t, err)
			expr, err := p.makeNumber(le)
			require.EqualError(t, err, c.Error)
			require.Nil(t, expr)

		})
	}
}
