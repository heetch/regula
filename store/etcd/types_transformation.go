package etcd

import (
	"fmt"

	"github.com/heetch/regula"
	"github.com/heetch/regula/rule"
	pb "github.com/heetch/regula/store/etcd/proto"
)

// rulesetToProtobuf creates a protobuf Ruleset from a regula.Ruleset.
func rulesetToProtobuf(rs *regula.Ruleset) *pb.Ruleset {
	pbrs := &pb.Ruleset{
		Type: rs.Type,
	}

	for _, r := range rs.Rules {
		pbr := &pb.Rule{
			Expr:   exprToProtobuf(r.Expr),
			Result: valueToProtobuf(r.Result),
		}
		pbrs.Rules = append(pbrs.Rules, pbr)
	}

	return pbrs
}

// exprToProtobuf creates a protobuf Expr from a rule.Expr.
func exprToProtobuf(expr rule.Expr) *pb.Expr {
	switch e := expr.(type) {
	case *rule.Value:
		v := &pb.Value{
			Type: e.Type,
			Kind: e.Kind,
			Data: e.Data,
		}

		return &pb.Expr{
			Expr: &pb.Expr_Value{Value: v},
		}
	case *rule.Param:
		p := &pb.Param{
			Kind: e.Kind,
			Type: e.Type,
			Name: e.Name,
		}

		return &pb.Expr{
			Expr: &pb.Expr_Param{Param: p},
		}
	}

	var (
		ope rule.Operander
		ok  bool
	)
	if ope, ok = expr.(rule.Operander); !ok {
		// there is something very weird, a rule.Expr which is not a rule.Value nor a rule.Param nor a rule.Operander
		// let's panic...
		panic(fmt.Sprintf("cannot create a pb.Expr - unexpected type: %T", expr))
	}

	o := &pb.Operator{
		Kind: expr.Contract().OpCode,
	}
	for _, op := range ope.Operands() {
		o.Operands = append(o.Operands, exprToProtobuf(op))
	}

	return &pb.Expr{
		Expr: &pb.Expr_Operator{Operator: o},
	}
}

// valueToProtobuf creates a protobuf Value from a rule.Value.
func valueToProtobuf(val *rule.Value) *pb.Value {
	return &pb.Value{
		Kind: val.Kind,
		Type: val.Type,
		Data: val.Data,
	}
}

// signatureToProtobuf creates a protobuf Signature from a regula.Signature.
func signatureToProtobuf(sig *regula.Signature) *pb.Signature {
	pbsig := &pb.Signature{
		ReturnType: sig.ReturnType,
		ParamTypes: make(map[string]string),
	}

	for k, v := range sig.ParamTypes {
		pbsig.ParamTypes[k] = v
	}

	return pbsig
}

// rulesetFromProtobuf creates a regula.Ruleset from a protobuf Ruleset.
func rulesetFromProtobuf(pbrs *pb.Ruleset) *regula.Ruleset {
	rs := &regula.Ruleset{
		Type: pbrs.Type,
	}

	for _, r := range pbrs.Rules {
		rr := &rule.Rule{
			Expr:   exprFromProtobuf(r.Expr),
			Result: valueFromProtobuf(r.Result),
		}
		rs.Rules = append(rs.Rules, rr)
	}

	return rs
}

// exprFromProtobuf creates a rule.Expr from a protobuf Expr.
func exprFromProtobuf(expr *pb.Expr) rule.Expr {
	switch e := expr.Expr.(type) {
	case *pb.Expr_Value:
		return &rule.Value{
			Kind: e.Value.Kind,
			Type: e.Value.Type,
			Data: e.Value.Data,
		}
	case *pb.Expr_Param:
		return &rule.Param{
			Kind: e.Param.Kind,
			Type: e.Param.Type,
			Name: e.Param.Name,
		}
	}

	var (
		pbop *pb.Expr_Operator
		ok   bool
	)
	if pbop, ok = expr.Expr.(*pb.Expr_Operator); !ok {
		// there is something very weird, a pb.Expr which is not a pb.Expr_Value nor a pb.Expr_Param nor a pb.Expr_Operator
		// let's panic...
		panic(fmt.Sprintf("cannot create a rule.Expr - unexpected type: %T", expr))
	}

	ope, err := rule.GetOperator(pbop.Operator.Kind)
	if err != nil {
		// every operator should be known at this place otherwise it's not a good sign
		// let's panic...
		panic(err.Error())
	}

	for _, o := range pbop.Operator.Operands {
		err := ope.PushExpr(exprFromProtobuf(o))
		if err != nil {
			// each operands should fulfil the appropriate Term of the Contract at this place otherwise it's not a good sign
			// let's panic
			panic(err.Error())
		}
	}

	return ope
}

// valueFromProtobuf creates a rule.Value from a protobuf Value.
func valueFromProtobuf(v *pb.Value) *rule.Value {
	return &rule.Value{
		Kind: v.Kind,
		Type: v.Type,
		Data: v.Data,
	}
}

// signatureFromProtobuf creates a regula.Signature from a protobuf Signature.
func signatureFromProtobuf(s *pb.Signature) *regula.Signature {
	sig := &regula.Signature{
		ReturnType: s.ReturnType,
		ParamTypes: make(map[string]string),
	}

	for k, v := range s.ParamTypes {
		sig.ParamTypes[k] = v
	}

	return sig
}
