package rule

import (
	"encoding/json"
	"errors"

	"github.com/tidwall/gjson"
)

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
