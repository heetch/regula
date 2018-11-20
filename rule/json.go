package rule

import (
	"encoding/json"
	"errors"

	"github.com/tidwall/gjson"
)

type operator struct {
	kind     string
	operands []Expr
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

type operands struct {
	Ops   []json.RawMessage `json:"operands"`
	Exprs []Expr
}

func (o *operands) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &o.Ops)
	if err != nil {
		return err
	}

	for _, op := range o.Ops {
		r := gjson.Get(string(op), "kind")
		n, err := unmarshalExpr(r.Str, []byte(op))
		if err != nil {
			return err
		}

		o.Exprs = append(o.Exprs, n)
	}

	return nil
}

func unmarshalExpr(kind string, data []byte) (Expr, error) {
	var e Expr
	var err error

	switch kind {
	case "value":
		var v Value
		e = &v
		err = json.Unmarshal(data, &v)
	case "param":
		var p Param
		e = &p
		err = json.Unmarshal(data, &p)
	case "eq":
		var eq exprEq
		e = &eq
		err = eq.UnmarshalJSON(data)
	case "in":
		var in exprIn
		e = &in
		err = in.UnmarshalJSON(data)
	case "not":
		var not exprNot
		e = &not
		err = not.UnmarshalJSON(data)
	case "and":
		var and exprAnd
		e = &and
		err = and.UnmarshalJSON(data)
	case "or":
		var or exprOr
		e = &or
		err = or.UnmarshalJSON(data)
	default:
		err = errors.New("unknown expression kind")
	}

	return e, err
}
