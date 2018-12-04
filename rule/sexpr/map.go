package sexpr

import (
	"fmt"

	"github.com/heetch/regula/rule"
)

// opCodeMap is a structure used to hold the core mapping between the
// symbollic expression language operators and the actual Expr
// implementations in the rule package.  This allows us a declarative
// style for the mapping, which saves a lot of boiler plate and
// reduces the chances of errors in maintenance.
type opCodeMap struct {
	symToOpCode map[string]string
	opCodeToSym map[string]string
}

// newOpCodeMap returns a pointer to a freshly initialised opCodeMap.
func newOpCodeMap() *opCodeMap {
	return &opCodeMap{
		symToOpCode: make(map[string]string),
		opCodeToSym: make(map[string]string),
	}
}

// mapSymbol takes one symbol string, and one operator name and
// creates a bidirectional mapping between them.
func (om *opCodeMap) mapSymbol(symbol, opCode string) {
	om.symToOpCode[symbol] = opCode
	om.opCodeToSym[opCode] = symbol
}

// getSymbolForOpCode returns the symbolic expression language symbol
// that is used for the provided regula rule.Expr type name.
func (om *opCodeMap) getSymbolForOpCode(opCode string) (string, error) {
	sym, ok := om.opCodeToSym[opCode]
	if !ok {
		return "invalid operator name", fmt.Errorf("%q is not a valid operator name", opCode)
	}
	return sym, nil
}

// getOpCodeForSymbol returns the name of the regula rule.Operator
// implementation that is mapped to the symbolic expression language
// symbol provided.
func (om *opCodeMap) getOpCodeForSymbol(symbol string) (string, error) {
	op, ok := om.symToOpCode[symbol]
	if !ok {
		return "invalid symbol", fmt.Errorf("%q is not a valid symbol", symbol)
	}
	return op, nil
}

//getOperatorForSymbol returns the rule.Operator that is mapped to a symbol in
//our symbolic expression language.
func (om *opCodeMap) getOperatorForSymbol(symbol string) (rule.Expr, error) {
	op, err := om.getOpCodeForSymbol(symbol)
	if err != nil {
		return nil, err
	}
	return rule.GetOperator(op)
}
