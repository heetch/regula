package sexpr

import (
	"fmt"
	"io"

	"github.com/heetch/regula/rule"
)

var errVal = rule.BoolValue(false)

// makeSymbolMap returns a mapOp with the full map of the built-in
// symbols of our symbolic expression language to their implementation
// as regula rule.Exprs.
func makeSymbolMap() *opMap {
	sm := newOpMap()
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
	buf      *lexicalElement
	buffered bool
	opMap    *opMap
}

func NewParser(r io.Reader) *Parser {
	return &Parser{
		s:     NewScanner(r),
		opMap: makeSymbolMap(),
	}
}

// scan returns the next lexicalElement from the text to be parsed, or
// the buffered element if unscan was called prior to scan.
func (p *Parser) scan() (*lexicalElement, error) {
	var err error
	if !p.buffered {
		var next *lexicalElement
		for {
			next, err = p.s.Scan()
			if err != nil {
				break
			}
			// Ignore white-space
			if next.Token != WHITESPACE {
				p.buf = next
				break
			}
		}

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
	return p.parseExpression()
}

func (p *Parser) parseSubExpression(term rule.Term, parent string, pos int) (rule.Expr, error) {
	subExpr, err := p.parseExpression()
	if err != nil {
		return errVal, err
	}
	subTE := subExpr.(rule.TypedExpression)
	subC := subTE.Contract()
	if !term.IsFulfilledBy(subTE) {
		return errVal, fmt.Errorf(
			"Expression %q expected a %q in position %d, but got a %q",
			parent, term, pos, subC.ReturnType)
	}
	return subExpr, nil
}

//
func (p *Parser) parseSubExpressions(expr rule.Expr) error {
	te := expr.(rule.TypedExpression)
	ce := expr.(rule.ComparableExpression)
	contract := te.Contract()
	symbol, err := p.opMap.getSymbolForOp(ce.GetKind())
	if err != nil {
		
	}

	for pos, term := range contract.Terms {
		switch term.Cardinality {
		case rule.ONE:
			nextLE, err := p.scan()
			if err != nil {
				return err
			}
			p.unscan()

			if nextLE.Token == RPAREN || nextLE.Token == EOF {
				return fmt.Errorf(
					"Operation %q expected a %q in position %d, but intstead the expression was terminated", , term.Type, pos+1)
			}

			subExpr, err := p.parseSubExpression(term, contract.Name, pos)
			if err != nil {
				return err
			}
			expr.PushExpr(subExpr)

		case rule.MANY:
			offset := 0
		ManyLoop:
			for {
				nextLE, err := p.scan()
				if err != nil {
					return err
				}
				p.unscan()

				if nextLE.Token == RPAREN || nextLE.Token == EOF {
					break ManyLoop
				}

				subExpr, err := p.parseSubExpression(term, contract.Name, pos+offset)
				if err != nil {
					return err
				}
				expr.PushExpr(subExpr)
				offset++
			}
		}
	}
	return nil
}

func (p *Parser) parseExpression() (rule.Expr, error) {
	var expr rule.Expr

Loop:
	for {
		le, err := p.scan()
		if err != nil {
			return errVal, err
		}
		switch le.Token {
		case EOF:
			break Loop
		case WHITESPACE:
			// We just ignore white space
			continue
		case LPAREN:
			// Left parenthesis must be followed by an operator
			expr, err = p.parseOperator()
			if err != nil {
				return errVal, err
			}
			err = p.parseSubExpressions(expr)
			if err != nil {
				return errVal, err
			}
		case BOOL:
			expr, err = p.parseBool(le)
			break Loop
		case RPAREN:
			break Loop
		}

	}
	return expr, nil

}

//
func (p *Parser) parseBool(le *lexicalElement) (rule.Expr, error) {
	if le.Literal == "true" {
		return rule.BoolValue(true), nil
	}
	return rule.BoolValue(false), nil
}

//
func (p *Parser) parseOperator() (rule.Expr, error) {
	le, err := p.scan()
	if err != nil {
		return errVal, err
	}
	if le.Token != SYMBOL {
		return errVal, fmt.Errorf("Expected an operator, but got the %s %q", le.Token, le.Literal)
	}

	op, err := p.opMap.getExprForSymbol(le.Literal)
	if err != nil {
		return errVal, err
	}

	return op, nil
}
