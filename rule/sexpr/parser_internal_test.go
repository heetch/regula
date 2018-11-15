package sexpr

import (
	"bytes"
	"testing"

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
