package rule

func init() {
	Operators["percentile"] = func() Operator { return newExprPercentile() }
}

/////////////////////////
// Percentile Operator //
/////////////////////////

type exprPercentile struct {
	operator
}

func newExprPercentile() *exprPercentile {
	return &exprPercentile{
		operator: operator{
			contract: Contract{
				OpCode:     "percentile",
				ReturnType: BOOLEAN,
				Terms: []Term{
					{
						Type:        ANY,
						Cardinality: ONE,
					},
					{
						Type:        INTEGER,
						Cardinality: ONE,
					},
				},
			},
		},
	}
}

// Percentile indicates whether the provided value is within a given
// percentile of the group of all such values.  It is intended to be
// used to assign values to groups for experimentation.
func Percentile(v Expr, p Expr) Expr {
	e := newExprPercentile()
	e.consumeOperands(v, p)
	return e
}

// Eval make exprPercentile comply with the Expr interface.
func (n *exprPercentile) Eval(params Params) (*Value, error) {
	hash := FNV(n.operands[0])
	v, err := exprToInt64(hash, params)
	if err != nil {
		return nil, err
	}
	p, err := exprToInt64(n.operands[1], params)
	if err != nil {
		return nil, err
	}
	if (v % 100) <= p {
		return BoolValue(true), nil
	}
	return BoolValue(false), nil
}
