package rule

import "hash/fnv"

func init() {
	Operators["fnv"] = func() Operator { return newExprFNV() }
}

//////////////////
// FNV operator //
//////////////////

type exprFNV struct {
	operator
}

func newExprFNV() *exprFNV {
	return &exprFNV{
		operator: operator{
			contract: Contract{
				OpCode:     "fnv",
				ReturnType: INTEGER,
				Terms: []Term{
					{
						Type:        ANY,
						Cardinality: ONE,
					},
				},
			},
		},
	}
}

// FNV returns an Integer hash of any value it is provided.  It uses
// the Fowler-Noll-Vo non-cryptographic hash function.
func FNV(v0 Expr) Expr {
	e := newExprFNV()
	e.consumeOperands(v0)
	return e
}

func (n *exprFNV) Eval(params Params) (*Value, error) {
	h32 := fnv.New32()
	op := n.operands[0]
	v, err := op.Eval(params)
	if err != nil {
		return nil, err
	}
	_, err = h32.Write([]byte(v.Data))
	if err != nil {
		return nil, err
	}
	return Int64Value(int64(h32.Sum32())), nil
}
