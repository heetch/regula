package etcd

import (
	"fmt"

	"github.com/heetch/regula"
	"github.com/heetch/regula/rule"
	pb "github.com/heetch/regula/store/etcd/proto"
)

// rulesetToProtobuf transforms rules into a pb.Ruleset.
func rulesetToProtobuf(rs []*rule.Rule) *pb.Ruleset {
	list := pb.Ruleset{
		Rules: make([]*pb.Rule, len(rs)),
	}

	for i, r := range rs {
		list.Rules[i] = ruleToProtobuf(r)
	}

	return &list
}

// rulesetFromProtobuf creates a ruleset from a pb.Ruleset.
func rulesetFromProtobuf(pbrs *pb.Ruleset) *regula.Ruleset {
	rules := make([]*rule.Rule, len(pbrs.Rules))

	for i, r := range pbrs.Ruleset {
		rules[i] = &rule.Rule{
			Expr:   exprFromProtobuf(r.Expr),
			Result: valueFromProtobuf(r.Result),
		}
	}

	return regula.NewRuleset(rules...)
}

// ruleToProtobuf transforms a rule into a pb.Rule.
func ruleToProtobuf(r *rule.Rule) *pb.Rule {
	return &pb.Rule{
		Expr:   exprToProtobuf(r.Expr),
		Result: valueToProtobuf(r.Result),
	}
}

// ruleFromProtobuf transforms a rule into a pb.Rule.
func ruleFromProtobuf(r *pb.Rule) *rule.Rule {
	return &rule.Rule{
		Expr:   exprFromProtobuf(r.Expr),
		Result: valueFromProtobuf(r.Result),
	}
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
		Kind:     expr.Contract().OpCode,
		Operands: make([]*pb.Expr, len(ope.Operands())),
	}
	for i, op := range ope.Operands() {
		o.Operands[i] = exprToProtobuf(op)
	}

	return &pb.Expr{
		Expr: &pb.Expr_Operator{Operator: o},
	}
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

// valueToProtobuf creates a protobuf Value from a rule.Value.
func valueToProtobuf(val *rule.Value) *pb.Value {
	return &pb.Value{
		Kind: val.Kind,
		Type: val.Type,
		Data: val.Data,
	}
}

// valueFromProtobuf creates a rule.Value from a protobuf Value.
func valueFromProtobuf(v *pb.Value) *rule.Value {
	return &rule.Value{
		Kind: v.Kind,
		Type: v.Type,
		Data: v.Data,
	}
}

// signatureToProtobuf creates a protobuf Signature from a regula.Signature.
func signatureToProtobuf(sig *regula.Signature) *pb.Signature {
	return &pb.Signature{
		ReturnType: sig.ReturnType,
		ParamTypes: sig.ParamTypes,
	}
}

// signatureFromProtobuf creates a regula.Signature from a protobuf Signature.
func signatureFromProtobuf(s *pb.Signature) *regula.Signature {
	sig := &regula.Signature{
		ReturnType: s.ReturnType,
		ParamTypes: s.ParamTypes,
	}

	for k, v := range s.ParamTypes {
		sig.ParamTypes[k] = v
	}

	return sig
}
