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
	contract := te.Contract()

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
					"Operation %q expected a %q in position %d, but intstead the expression was terminated", contract.Name, term.Type, pos+1)
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

	op, err := getOperatorExprForSymbol(le.Literal)
	if err != nil {
		return errVal, err
	}

	return op, nil
}

// Return the operator Expr representing the operator provided.
// Note that our symbolic expressions notation for operators is not
// an exact match for the naming of the expressions internally, so this
// is the definitive mapping.
func getOperatorExprForSymbol(symbolName string) (rule.Expr, error) {
	switch symbolName {
	case "=":
		return rule.GetOperatorExpr("eq")
	case "+":
		return rule.GetOperatorExpr("add")
	case "-":
		return rule.GetOperatorExpr("sub")
	case "*":
		return rule.GetOperatorExpr("mult")
	case "/":
		return rule.GetOperatorExpr("div")
	case "%":
		return rule.GetOperatorExpr("mod")
	case "in":
		return rule.GetOperatorExpr("in")
	case "and":
		return rule.GetOperatorExpr("and")
	case "or":
		return rule.GetOperatorExpr("or")
	case "not":
		return rule.GetOperatorExpr("not")
	case ">":
		return rule.GetOperatorExpr("gt")
	case ">=":
		return rule.GetOperatorExpr("gte")
	case "<":
		return rule.GetOperatorExpr("lt")
	case "<=":
		return rule.GetOperatorExpr("lte")
	case "hash":
		return rule.GetOperatorExpr("hash")
	case "percentile":
		return rule.GetOperatorExpr("percentile")
	}
	return nil, fmt.Errorf("Expected an operator, but got the Symbol %q", symbolName)
}
