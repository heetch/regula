package rule

import "encoding/json"

// The operator is the representation of an operation to be performed
// on some given set of operands.  Some Exprs are operators, but Value
// and Param types are not.
//
// Operators have a "kind" (the name of the operation), and a slice of
// operands (which are themselves Exprs).  Operands are added to an
// operator by means of the PushExpr call.
type operator struct {
	kind     string
	contract Contract
	operands []Expr
}

func (o *operator) PushExpr(e Expr) {
	o.operands = append(o.operands, e)
}

//
func (o *operator) Contract() Contract {
	return o.contract
}

func (o *operator) Same(c ComparableExpression) bool {
	if o.kind == c.GetKind() {
		o2, ok := c.(operander)
		if ok {
			ops := o2.Operands()
			len1 := len(o.operands)
			len2 := len(ops)
			if len1 != len2 {
				return false
			}
			for i := 0; i < len1; i++ {
				ce1 := o.operands[i].(ComparableExpression)
				ce2 := ops[i].(ComparableExpression)
				if !ce1.Same(ce2) {
					return false
				}
			}
			return true
		}
		return false
	}
	return false
}

//
func (o *operator) GetKind() string {
	return o.kind
}

func (o *operator) UnmarshalJSON(data []byte) error {
	var node struct {
		Kind     string
		Operands operands
	}

	err := json.Unmarshal(data, &node)
	if err != nil {
		return err
	}

	o.operands = node.Operands.Exprs
	o.kind = node.Kind

	return nil
}

func (o *operator) MarshalJSON() ([]byte, error) {
	var op struct {
		Kind     string `json:"kind"`
		Operands []Expr `json:"operands"`
	}

	op.Kind = o.kind
	op.Operands = o.operands
	return json.Marshal(&op)
}

func (o *operator) Eval(params Params) (*Value, error) {
	return nil, nil
}

func (o *operator) Operands() []Expr {
	return o.operands
}
