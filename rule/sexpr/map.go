package sexpr

import "fmt"

// opMap is a structure used to hold the core mapping between the
// symbollic expression language operators and the actual Expr
// implementations in the rule package.  This allows us a declarative
// style for the mapping, which saves a lot of boiler plate and
// reduces the chances of errors in maintenance.
type opMap struct {
	symToOp map[string]string
	opToSym map[string]string
}

// newOpMap returns a pointer to a freshly initialised opMap.
func newOpMap() *opMap {
	return &opMap{
		symToOp: make(map[string]string),
		opToSym: make(map[string]string),
	}
}

// mapSymbol takes one symbol string, and one operator name and
// creates a bidirectional mapping between them.
func (om *opMap) mapSymbol(symbol, opCode string) {
	om.symToOp[symbol] = opCode
	om.opToSym[opCode] = symbol
}

// getSymbolForOp returns the symbolic expression language symbol
// that is used for the provided regula rule.Expr type name.
func (om *opMap) getSymbolForOp(opCode string) (string, error) {
	sym, ok := om.opToSym[opCode]
	if !ok {
		return "invalid operator name", fmt.Errorf("%q is not a valid operator name", opCode)
	}
	return sym, nil
}

// getOpForSymbol returns the name of the regula rule.Expr
// implementation that is mapped to the symbolic expression language
// symbol provided.
func (om *opMap) getOpForSymbol(symbol string) (string, error) {
	op, ok := om.symToOp[symbol]
	if !ok {
		return "invalid symbol", fmt.Errorf("%q is not a valid symbol", symbol)
	}
	return op, nil
}
