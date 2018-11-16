package sexpr

import (
	"io"

	"github.com/heetch/regula/rule"
)

// makeSymbolMap returns a opCodeMap with the full map of the built-in
// symbols of our symbolic expression language to their implementation
// as regula rule.Operators.
func makeSymbolMap() *opCodeMap {
	sm := newOpCodeMap()
	sm.mapSymbol("=", "eq")
	sm.mapSymbol("+", "add")
	sm.mapSymbol("-", "sub")
	sm.mapSymbol("*", "mult")
	sm.mapSymbol("/", "div")
	sm.mapSymbol("%", "mod")
	sm.mapSymbol("in", "in")
	sm.mapSymbol("and", "and")
	sm.mapSymbol("or", "or")
	sm.mapSymbol("not", "not")
	sm.mapSymbol(">", "gt")
	sm.mapSymbol(">=", "gte")
	sm.mapSymbol("<", "lt")
	sm.mapSymbol("<=", "lte")
	sm.mapSymbol("hash", "hash")
	sm.mapSymbol("percentile", "percentile")
	return sm
}

// Parser
type Parser struct {
	s        *Scanner
	buf      lexicalElement
	buffered bool
}

func NewParser(r io.Reader) *Parser {
	return &Parser{s: NewScanner(r)}
}

// scan returns the next lexicalElement from the text to be parsed, or
// the buffered element if unscan was called prior to scan.
func (p *Parser) scan() (lexicalElement, error) {
	var err error
	if !p.buffered {
		p.buf, err = p.s.Scan()
	}
	p.buffered = false
	return p.buf, err
}

// unscan instructs the Parser to use the buffered value for the next call to scan
func (p *Parser) unscan() {
	p.buffered = true
}

//
func (p *Parser) Parse() (rule.Expr, error) {

}
