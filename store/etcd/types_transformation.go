package etcd

import (
	"fmt"
	"log"

	"github.com/heetch/regula"
	"github.com/heetch/regula/rule"
	pb "github.com/heetch/regula/store/etcd/proto"
)

// toProtobufExpr creates a protobuf Expr from a rule.Expr.
func toProtobufExpr(expr rule.Expr) *pb.Expr {
	switch v := expr.(type) {
	case *rule.Value:
		fmt.Printf("op it's value of type %s and kind %s and value %s\n", v.Type, v.Kind, v.Data)
		pv := &pb.Value{}
		pv.Type = v.Type
		pv.Kind = v.Kind
		pv.Data = v.Data

		pe := &pb.Expr{}
		pe.Expr = &pb.Expr_Value{Value: pv}
		return pe

	case *rule.Param:
		fmt.Printf("op it's param of type %s and kind %s and name %s\n", v.Type, v.Kind, v.Name)
		pp := &pb.Param{}
		pp.Kind = v.Kind
		pp.Type = v.Type
		pp.Name = v.Name

		pe := &pb.Expr{}
		pe.Expr = &pb.Expr_Param{Param: pp}
		return pe
	}

	var (
		ope rule.Operander
		ok  bool
	)
	if ope, ok = expr.(rule.Operander); !ok {
		log.Fatal("wesh ?????")
	}

	po := &pb.Operator{}
	po.Kind = expr.Contract().OpCode

	for _, op := range ope.Operands() {
		po.Operands = append(po.Operands, toProtobufExpr(op))
	}

	peo := &pb.Expr_Operator{Operator: po}
	pee := &pb.Expr{Expr: peo}
	return pee
}

// toProtobufValue creates a protobuf Value from a rule.Value.
func toProtobufValue(val *rule.Value) *pb.Value {
	return &pb.Value{
		Kind: val.Kind,
		Type: val.Type,
		Data: val.Data,
	}
}

// toProtobufRuleset creates a protobuf Ruleset from a regula.Ruleset.
func toProtobufRuleset(rs *regula.Ruleset) *pb.Ruleset {
	prs := &pb.Ruleset{
		Type: rs.Type,
	}

	for _, r := range rs.Rules {
		pr := &pb.Rule{}

		pr.Expr = toProtobufExpr(r.Expr)
		pr.Result = toProtobufValue(r.Result)

		prs.Rules = append(prs.Rules, pr)
	}
	return prs
}

// fromProtobufValue creates a rule.Value from a protobuf Value.
func fromProtobufValue(v *pb.Value) *rule.Value {
	return &rule.Value{
		Kind: v.Kind,
		Type: v.Type,
		Data: v.Data,
	}
}

// fromProtobufExpr creates a rule.Expr from a protobuf Expr.
func fromProtobufExpr(e *pb.Expr) rule.Expr {
	fmt.Println(e)
	switch ee := e.Expr.(type) {
	case *pb.Expr_Value:
		fmt.Println("c'est une Value")
		v := &rule.Value{
			Kind: ee.Value.Kind,
			Type: ee.Value.Type,
			Data: ee.Value.Data,
		}
		// operands = append(operands, v)
		return v
	case *pb.Expr_Param:
		fmt.Println("c'est un param")
		p := &rule.Param{
			Kind: ee.Param.Kind,
			Type: ee.Param.Type,
			Name: ee.Param.Name,
		}
		return p
	}

	// fmt.Printf("C'est le type %T, donc on recurse poto!", ee)
	var (
		pbop *pb.Expr_Operator
		ok   bool
	)
	if pbop, ok = e.Expr.(*pb.Expr_Operator); !ok {
		log.Fatal("keskispass?")
	}

	// var operands []rule.Expr
	ope, _ := rule.GetOperator(pbop.Operator.Kind)

	for _, o := range pbop.Operator.Operands {
		_ = ope.PushExpr(fromProtobufExpr(o))
	}

	return ope
}

// fromProtobufRuleset creates a regula.Ruleset from a protobuf Ruleset.
func fromProtobufRuleset(pbrs *pb.Ruleset) *regula.Ruleset {
	fmt.Println("#$######$#$#$#$#$#$#$#$#")

	rs := &regula.Ruleset{}
	rs.Type = pbrs.Type

	for _, r := range pbrs.Rules {
		rr := &rule.Rule{}
		rr.Result = fromProtobufValue(r.Result)
		rr.Expr = fromProtobufExpr(r.Expr)
		rs.Rules = append(rs.Rules, rr)
	}
	return rs
}
