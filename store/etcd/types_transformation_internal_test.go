package etcd

import (
	"encoding/json"
	"fmt"
	"testing"
	"unsafe"

	proto "github.com/golang/protobuf/proto"
	"github.com/heetch/regula"
	"github.com/heetch/regula/rule"
	"github.com/heetch/regula/store"
	pb "github.com/heetch/regula/store/etcd/proto"
	"github.com/stretchr/testify/require"
)

// TestFromProtobufTypes tests that the transformation from a protobuf ruleset to a regula ruleset works well.
func TestFromProtobufTypes(t *testing.T) {
	cases := []struct {
		name string
		pb   *pb.Ruleset
		exp  func() *regula.Ruleset
	}{
		{
			name: "Flat root",
			pb: &pb.Ruleset{
				Type: "int64",
				Rules: []*pb.Rule{
					{
						Result: &pb.Value{
							Data: "42",
							Kind: "value",
							Type: "int64",
						},
						Expr: &pb.Expr{
							Expr: &pb.Expr_Value{
								Value: &pb.Value{
									Data: "true",
									Kind: "value",
									Type: "bool",
								},
							},
						},
					},
				},
			},
			exp: func() *regula.Ruleset {
				rs, err := regula.NewInt64Ruleset(rule.New(rule.True(), rule.Int64Value(42)))
				require.NoError(t, err)
				return rs
			},
		},
		{
			name: "Eq returns a bool without parameters",
			pb: &pb.Ruleset{
				Type: "bool",
				Rules: []*pb.Rule{
					{
						Result: &pb.Value{
							Data: "true",
							Kind: "value",
							Type: "bool",
						},
						Expr: &pb.Expr{
							Expr: &pb.Expr_Operator{
								Operator: &pb.Operator{
									Kind: "eq",
									Operands: []*pb.Expr{
										{
											Expr: &pb.Expr_Value{
												Value: &pb.Value{
													Data: "true",
													Kind: "value",
													Type: "bool",
												},
											},
										},
										{
											Expr: &pb.Expr_Value{
												Value: &pb.Value{
													Data: "true",
													Kind: "value",
													Type: "bool",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			exp: func() *regula.Ruleset {
				rs, err := regula.NewBoolRuleset(rule.New(rule.Eq(rule.True(), rule.True()), rule.BoolValue(true)))
				require.NoError(t, err)
				return rs
			},
		},
		{
			name: "And returns a string with parameters",
			pb: &pb.Ruleset{
				Type: "string",
				Rules: []*pb.Rule{
					{
						Result: &pb.Value{
							Data: "foo",
							Kind: "value",
							Type: "string",
						},
						Expr: &pb.Expr{
							Expr: &pb.Expr_Operator{
								Operator: &pb.Operator{
									Kind: "and",
									Operands: []*pb.Expr{
										{
											Expr: &pb.Expr_Param{
												Param: &pb.Param{
													Name: "1st-param",
													Kind: "param",
													Type: "bool",
												},
											},
										},
										{
											Expr: &pb.Expr_Param{
												Param: &pb.Param{
													Name: "2nd-param",
													Kind: "param",
													Type: "bool",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			exp: func() *regula.Ruleset {
				rs, err := regula.NewStringRuleset(rule.New(rule.And(rule.BoolParam("1st-param"), rule.BoolParam("2nd-param")), rule.StringValue("foo")))
				require.NoError(t, err)
				return rs
			},
		},
		{
			name: "And with nested operator",
			pb: &pb.Ruleset{
				Type: "float64",
				Rules: []*pb.Rule{
					{
						Result: &pb.Value{
							Data: "42.420000",
							Kind: "value",
							Type: "float64",
						},
						Expr: &pb.Expr{
							Expr: &pb.Expr_Operator{
								Operator: &pb.Operator{
									Kind: "and",
									Operands: []*pb.Expr{
										{
											Expr: &pb.Expr_Operator{
												Operator: &pb.Operator{
													Kind: "not",
													Operands: []*pb.Expr{
														{
															Expr: &pb.Expr_Value{
																Value: &pb.Value{
																	Data: "false",
																	Kind: "value",
																	Type: "bool",
																},
															},
														},
													},
												},
											},
										},
										{
											Expr: &pb.Expr_Param{
												Param: &pb.Param{
													Name: "param",
													Kind: "param",
													Type: "bool",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			exp: func() *regula.Ruleset {
				rs, err := regula.NewFloat64Ruleset(rule.New(rule.And(rule.Not(rule.BoolValue(false)), rule.BoolParam("param")), rule.Float64Value(42.42)))
				require.NoError(t, err)
				return rs
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actRs := fromProtobufRuleset(c.pb)
			// require.NoError(t, err)
			require.Equal(t, c.exp(), actRs)
		})
	}
}

func TestToProtobufTypes(t *testing.T) {
	// rs, err := regula.NewBoolRuleset(rule.New(rule.And(rule.Eq(rule.StringParam("wesh-string-param"), rule.StringValue("wesh-string-value")), rule.Eq(rule.True(), rule.BoolParam("bool-param")), rule.BoolValue(true)), rule.BoolValue(true)))
	rs, err := regula.NewBoolRuleset(rule.New(rule.And(rule.True(), rule.Eq(rule.True(), rule.BoolParam("wesh-param"))), rule.BoolValue(true)))
	require.NoError(t, err)

	prs := toProtobufRuleset(rs)
	fmt.Println(prs)

	pre := pb.RulesetEntry{
		Path:    "a/b/c",
		Version: "abc123",
		Ruleset: prs,
	}
	b, err := proto.Marshal(&pre)
	require.NoError(t, err)

	fmt.Printf("proto: %T, %d\n", b, unsafe.Sizeof(string(b)))

	re := store.RulesetEntry{
		Path:    "a/b/c",
		Version: "abc123",
		Ruleset: rs,
	}
	raw, err := json.Marshal(&re)
	require.NoError(t, err)
	fmt.Printf("json: %T, %d\n", raw, unsafe.Sizeof(string(raw)))

	cases := []struct {
		name string
		rs   func() *regula.Ruleset
		exp  *pb.Ruleset
	}{
		{
			name: "Flat root",
			rs: func() *regula.Ruleset {
				rs, err := regula.NewInt64Ruleset(rule.New(rule.True(), rule.Int64Value(42)))
				require.NoError(t, err)
				return rs
			},
			exp: &pb.Ruleset{
				Type: "int64",
				Rules: []*pb.Rule{
					{
						Result: &pb.Value{
							Data: "42",
							Kind: "value",
							Type: "int64",
						},
						Expr: &pb.Expr{
							Expr: &pb.Expr_Value{
								Value: &pb.Value{
									Data: "true",
									Kind: "value",
									Type: "bool",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Eq returns a bool without parameters",
			rs: func() *regula.Ruleset {
				rs, err := regula.NewBoolRuleset(rule.New(rule.Eq(rule.True(), rule.True()), rule.BoolValue(true)))
				require.NoError(t, err)
				return rs
			},
			exp: &pb.Ruleset{
				Type: "bool",
				Rules: []*pb.Rule{
					{
						Result: &pb.Value{
							Data: "true",
							Kind: "value",
							Type: "bool",
						},
						Expr: &pb.Expr{
							Expr: &pb.Expr_Operator{
								Operator: &pb.Operator{
									Kind: "eq",
									Operands: []*pb.Expr{
										{
											Expr: &pb.Expr_Value{
												Value: &pb.Value{
													Data: "true",
													Kind: "value",
													Type: "bool",
												},
											},
										},
										{
											Expr: &pb.Expr_Value{
												Value: &pb.Value{
													Data: "true",
													Kind: "value",
													Type: "bool",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "And returns a string with parameters",
			rs: func() *regula.Ruleset {
				rs, err := regula.NewStringRuleset(rule.New(rule.And(rule.BoolParam("1st-param"), rule.BoolParam("2nd-param")), rule.StringValue("foo")))
				require.NoError(t, err)
				return rs
			},
			exp: &pb.Ruleset{
				Type: "string",
				Rules: []*pb.Rule{
					{
						Result: &pb.Value{
							Data: "foo",
							Kind: "value",
							Type: "string",
						},
						Expr: &pb.Expr{
							Expr: &pb.Expr_Operator{
								Operator: &pb.Operator{
									Kind: "and",
									Operands: []*pb.Expr{
										{
											Expr: &pb.Expr_Param{
												Param: &pb.Param{
													Name: "1st-param",
													Kind: "param",
													Type: "bool",
												},
											},
										},
										{
											Expr: &pb.Expr_Param{
												Param: &pb.Param{
													Name: "2nd-param",
													Kind: "param",
													Type: "bool",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "And with nested operator",
			rs: func() *regula.Ruleset {
				rs, err := regula.NewFloat64Ruleset(rule.New(rule.And(rule.Not(rule.BoolValue(false)), rule.BoolParam("param")), rule.Float64Value(42.42)))
				require.NoError(t, err)
				return rs
			},
			exp: &pb.Ruleset{
				Type: "float64",
				Rules: []*pb.Rule{
					{
						Result: &pb.Value{
							Data: "42.420000",
							Kind: "value",
							Type: "float64",
						},
						Expr: &pb.Expr{
							Expr: &pb.Expr_Operator{
								Operator: &pb.Operator{
									Kind: "and",
									Operands: []*pb.Expr{
										{
											Expr: &pb.Expr_Operator{
												Operator: &pb.Operator{
													Kind: "not",
													Operands: []*pb.Expr{
														{
															Expr: &pb.Expr_Value{
																Value: &pb.Value{
																	Data: "false",
																	Kind: "value",
																	Type: "bool",
																},
															},
														},
													},
												},
											},
										},
										{
											Expr: &pb.Expr_Param{
												Param: &pb.Param{
													Name: "param",
													Kind: "param",
													Type: "bool",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actRs := toProtobufRuleset(c.rs())
			// require.NoError(t, err)
			expRs := c.exp
			require.Equal(t, expRs, actRs)
		})
	}
}
