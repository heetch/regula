package sexpr

import (
	"fmt"
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
	s         *Scanner
	buf       *lexicalElement
	buffered  bool
	opCodeMap *opCodeMap
}

func NewParser(r io.Reader) *Parser {
	return &Parser{
		s:         NewScanner(r),
		opCodeMap: makeSymbolMap(),
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
	t := rule.Term{
		Type:        rule.BOOLEAN,
		Cardinality: rule.ONE,
	}
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}
	if !t.IsFulfilledBy(expr) {
		return nil, fmt.Errorf("The root expression in a rule must return a BOOLEAN, but it returns %q", expr.Contract().ReturnType)
	}
	return expr, nil
}

//
func (p *Parser) NewError(err error) error {
	// TODO, furnish this with more information
	return err
}

//
func (p *Parser) parseExpression() (rule.Expr, error) {
	var expr rule.Expr
	var inOperator bool
	var opExpr rule.Expr

Loop:
	for {
		le, err := p.scan()
		if err != nil {
			return nil, err
		}
		switch le.Token {
		case EOF:
			break Loop
		case WHITESPACE:
			// We just ignore white space
			continue
		case LPAREN:
			if inOperator {
				// This is a subexpression
				p.unscan()
				expr, err = p.parseExpression()
				if err != nil {
					return nil, err
				}
				if err := opExpr.(rule.Operator).PushExpr(expr); err != nil {
					// TODO: drastically improve this error message
					return nil, p.NewError(fmt.Errorf(
						"Type mismatch in subexpression",
					))
				}
				continue
			}
			// Left parenthesis must be followed by an operator
			opExpr, err = p.parseOperator()
			if err != nil {
				return nil, err
			}
			inOperator = true

		case BOOL:
			expr, err = p.parseBool(le)
			if !inOperator {
				break Loop
			}
			if err := opExpr.(rule.Operator).PushExpr(expr); err != nil {
				return nil, p.NewError(err)
			}

		case RPAREN:
			if !inOperator {
				return nil, p.NewError(fmt.Errorf("Unexpected closing parenthesis"))
			}
			inOperator = false
			if err := opExpr.(rule.Operator).Finalise(); err != nil {
				return nil, p.NewError(err)
			}
			expr = opExpr
			break Loop
		}

	}
	return expr, nil
}

// 	return p.parseExpression()
// }

// func (p *Parser) parseSubExpression(term rule.Term, parent string, pos int) (rule.Expr, error) {
// 	subExpr, err := p.parseExpression()
// 	if err != nil {
// 		return nil, err
// 	}

// 	if !term.IsFulfilledBy(subExpr) {
// 		return nil, fmt.Errorf(
// 			"Expression %q expected a %q in position %d, but got a %q",
// 			parent, term, pos, subExpr.Contract().ReturnType)
// 	}
// 	return subExpr, nil
// }

// //
// func (p *Parser) parseSubExpressions(expr rule.Expr) error {
// 	ce := expr.(rule.ComparableExpression)
// 	contract := expr.Contract()
// 	symbol, err := p.opCodeMap.getSymbolForOpCode(ce.GetKind())
// 	if err != nil {

// 	}

// 	for pos, term := range contract.Terms {
// 		switch term.Cardinality {
// 		case rule.ONE:
// 			nextLE, err := p.scan()
// 			if err != nil {
// 				return err
// 			}
// 			p.unscan()

// 			if nextLE.Token == RPAREN || nextLE.Token == EOF {
// 				return fmt.Errorf(
// 					"Operation %q expected a %q in position %d, but instead the expression was terminated", symbol, term.Type, pos+1)
// 			}

// 			subExpr, err := p.parseSubExpression(term, contract.OpCode, pos)
// 			if err != nil {
// 				return err
// 			}
// 			expr.(rule.Operator).PushExpr(subExpr)

// 		case rule.MANY:
// 			offset := 0
// 		ManyLoop:
// 			for {
// 				nextLE, err := p.scan()
// 				if err != nil {
// 					return err
// 				}
// 				p.unscan()

// 				if nextLE.Token == RPAREN || nextLE.Token == EOF {
// 					break ManyLoop
// 				}

// 				subExpr, err := p.parseSubExpression(term, contract.OpCode, pos+offset)
// 				if err != nil {
// 					return err
// 				}
// 				expr.(rule.Operator).PushExpr(subExpr)
// 				offset++
// 			}
// 		}
// 	}
// 	return nil
// }

// func (p *Parser) parseExpression() (rule.Expr, error) {
// 	var expr rule.Expr

// Loop:
// 	for {
// 		le, err := p.scan()
// 		if err != nil {
// 			return nil, err
// 		}
// 		switch le.Token {
// 		case EOF:
// 			break Loop
// 		case WHITESPACE:
// 			// We just ignore white space
// 			continue
// 		case LPAREN:
// 			// Left parenthesis must be followed by an operator
// 			expr, err = p.parseOperator()
// 			if err != nil {
// 				return nil, err
// 			}
// 			err = p.parseSubExpressions(expr)
// 			if err != nil {
// 				return nil, err
// 			}
// 		case BOOL:
// 			expr, err = p.parseBool(le)
// 			break Loop
// 		case RPAREN:
// 			break Loop
// 		}

// 	}
// 	return expr, nil

// }

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
		return nil, err
	}
	if le.Token != SYMBOL {
		return nil, fmt.Errorf("Expected an operator, but got the %s %q", le.Token, le.Literal)
	}

	op, err := p.opCodeMap.getOperatorForSymbol(le.Literal)
	if err != nil {
		return nil, err
	}

	return op, nil
}
