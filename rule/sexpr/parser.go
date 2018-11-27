package sexpr

import (
	"fmt"
	"io"

	"github.com/heetch/regula/rule"
)

var errVal = rule.BoolValue(false)

// Parser
type Parser struct {
	s        *Scanner
	buf      *lexicalElement
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
