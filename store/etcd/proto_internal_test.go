package etcd

import (
	"testing"

	"github.com/heetch/regula"
	"github.com/heetch/regula/rule"
	pb "github.com/heetch/regula/store/etcd/proto"
	"github.com/stretchr/testify/require"
)

// TestFromProtobufRuleset tests that the transformation from a protobuf ruleset to a regula ruleset works well.
func TestFromProtobufRuleset(t *testing.T) {
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
		{
			name: "Two rules",
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
					{
						Result: &pb.Value{
							Data: "21.210000",
							Kind: "value",
							Type: "float64",
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
				rs, err := regula.NewFloat64Ruleset(
					rule.New(rule.And(rule.Not(rule.BoolValue(false)), rule.BoolParam("param")), rule.Float64Value(42.42)),
					rule.New(rule.And(rule.BoolParam("1st-param"), rule.BoolParam("2nd-param")), rule.Float64Value(21.21)),
				)
				require.NoError(t, err)
				return rs
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var rs *regula.Ruleset
			require.NotPanics(t, func() { rs = rulesetFromProtobuf(c.pb) })
			require.Equal(t, c.exp(), rs)
		})
	}
}

// TestToProtobufRuleset tests that the transformation from a regula ruleset to a protobuf ruleset works well.
func TestToProtobufRuleset(t *testing.T) {
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
		{
			name: "Two rules",
			rs: func() *regula.Ruleset {
				rs, err := regula.NewFloat64Ruleset(
					rule.New(rule.And(rule.Not(rule.BoolValue(false)), rule.BoolParam("param")), rule.Float64Value(42.42)),
					rule.New(rule.And(rule.BoolParam("1st-param"), rule.BoolParam("2nd-param")), rule.Float64Value(21.21)),
				)
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
					{
						Result: &pb.Value{
							Data: "21.210000",
							Kind: "value",
							Type: "float64",
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
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var rs *pb.Ruleset
			require.NotPanics(t, func() { rs = rulesetToProtobuf(c.rs()) })
			require.Equal(t, c.exp, rs)
		})
	}
}

func TestToProtobufSignature(t *testing.T) {
	sig := &regula.Signature{
		ReturnType: "bool",
		ParamTypes: map[string]string{
			"1st": "bool",
			"2nd": "string",
			"3rd": "int64",
		},
	}

	pbsig := signatureToProtobuf(sig)

	exp := &pb.Signature{
		ReturnType: "bool",
		ParamTypes: map[string]string{
			"1st": "bool",
			"2nd": "string",
			"3rd": "int64",
		},
	}

	require.Equal(t, exp, pbsig)
}

func TestFromProtobufSignature(t *testing.T) {
	pbsig := &pb.Signature{
		ReturnType: "bool",
		ParamTypes: map[string]string{
			"1st": "bool",
			"2nd": "string",
			"3rd": "int64",
		},
	}

	sig := signatureFromProtobuf(pbsig)

	exp := &regula.Signature{
		ReturnType: "bool",
		ParamTypes: map[string]string{
			"1st": "bool",
			"2nd": "string",
			"3rd": "int64",
		},
	}

	require.Equal(t, exp, sig)
}
