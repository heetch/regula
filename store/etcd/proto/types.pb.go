// Code generated by protoc-gen-go.
// source: types.proto
// DO NOT EDIT!

/*
Package proto is a generated protocol buffer package.

It is generated from these files:
	types.proto

It has these top-level messages:
	Rules
	Rule
	Expr
	Value
	Param
	Operator
	Signature
	Versions
*/
package proto

import proto1 "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto1.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto1.ProtoPackageIsVersion2 // please upgrade the proto package

type Rules struct {
	Rules []*Rule `protobuf:"bytes,1,rep,name=rules" json:"rules,omitempty"`
}

func (m *Rules) Reset()                    { *m = Rules{} }
func (m *Rules) String() string            { return proto1.CompactTextString(m) }
func (*Rules) ProtoMessage()               {}
func (*Rules) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Rules) GetRules() []*Rule {
	if m != nil {
		return m.Rules
	}
	return nil
}

type Rule struct {
	Expr   *Expr  `protobuf:"bytes,1,opt,name=expr" json:"expr,omitempty"`
	Result *Value `protobuf:"bytes,2,opt,name=result" json:"result,omitempty"`
}

func (m *Rule) Reset()                    { *m = Rule{} }
func (m *Rule) String() string            { return proto1.CompactTextString(m) }
func (*Rule) ProtoMessage()               {}
func (*Rule) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Rule) GetExpr() *Expr {
	if m != nil {
		return m.Expr
	}
	return nil
}

func (m *Rule) GetResult() *Value {
	if m != nil {
		return m.Result
	}
	return nil
}

type Expr struct {
	// Types that are valid to be assigned to Expr:
	//	*Expr_Value
	//	*Expr_Param
	//	*Expr_Operator
	Expr isExpr_Expr `protobuf_oneof:"expr"`
}

func (m *Expr) Reset()                    { *m = Expr{} }
func (m *Expr) String() string            { return proto1.CompactTextString(m) }
func (*Expr) ProtoMessage()               {}
func (*Expr) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

type isExpr_Expr interface {
	isExpr_Expr()
}

type Expr_Value struct {
	Value *Value `protobuf:"bytes,1,opt,name=value,oneof"`
}
type Expr_Param struct {
	Param *Param `protobuf:"bytes,2,opt,name=param,oneof"`
}
type Expr_Operator struct {
	Operator *Operator `protobuf:"bytes,3,opt,name=operator,oneof"`
}

func (*Expr_Value) isExpr_Expr()    {}
func (*Expr_Param) isExpr_Expr()    {}
func (*Expr_Operator) isExpr_Expr() {}

func (m *Expr) GetExpr() isExpr_Expr {
	if m != nil {
		return m.Expr
	}
	return nil
}

func (m *Expr) GetValue() *Value {
	if x, ok := m.GetExpr().(*Expr_Value); ok {
		return x.Value
	}
	return nil
}

func (m *Expr) GetParam() *Param {
	if x, ok := m.GetExpr().(*Expr_Param); ok {
		return x.Param
	}
	return nil
}

func (m *Expr) GetOperator() *Operator {
	if x, ok := m.GetExpr().(*Expr_Operator); ok {
		return x.Operator
	}
	return nil
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*Expr) XXX_OneofFuncs() (func(msg proto1.Message, b *proto1.Buffer) error, func(msg proto1.Message, tag, wire int, b *proto1.Buffer) (bool, error), func(msg proto1.Message) (n int), []interface{}) {
	return _Expr_OneofMarshaler, _Expr_OneofUnmarshaler, _Expr_OneofSizer, []interface{}{
		(*Expr_Value)(nil),
		(*Expr_Param)(nil),
		(*Expr_Operator)(nil),
	}
}

func _Expr_OneofMarshaler(msg proto1.Message, b *proto1.Buffer) error {
	m := msg.(*Expr)
	// expr
	switch x := m.Expr.(type) {
	case *Expr_Value:
		b.EncodeVarint(1<<3 | proto1.WireBytes)
		if err := b.EncodeMessage(x.Value); err != nil {
			return err
		}
	case *Expr_Param:
		b.EncodeVarint(2<<3 | proto1.WireBytes)
		if err := b.EncodeMessage(x.Param); err != nil {
			return err
		}
	case *Expr_Operator:
		b.EncodeVarint(3<<3 | proto1.WireBytes)
		if err := b.EncodeMessage(x.Operator); err != nil {
			return err
		}
	case nil:
	default:
		return fmt.Errorf("Expr.Expr has unexpected type %T", x)
	}
	return nil
}

func _Expr_OneofUnmarshaler(msg proto1.Message, tag, wire int, b *proto1.Buffer) (bool, error) {
	m := msg.(*Expr)
	switch tag {
	case 1: // expr.value
		if wire != proto1.WireBytes {
			return true, proto1.ErrInternalBadWireType
		}
		msg := new(Value)
		err := b.DecodeMessage(msg)
		m.Expr = &Expr_Value{msg}
		return true, err
	case 2: // expr.param
		if wire != proto1.WireBytes {
			return true, proto1.ErrInternalBadWireType
		}
		msg := new(Param)
		err := b.DecodeMessage(msg)
		m.Expr = &Expr_Param{msg}
		return true, err
	case 3: // expr.operator
		if wire != proto1.WireBytes {
			return true, proto1.ErrInternalBadWireType
		}
		msg := new(Operator)
		err := b.DecodeMessage(msg)
		m.Expr = &Expr_Operator{msg}
		return true, err
	default:
		return false, nil
	}
}

func _Expr_OneofSizer(msg proto1.Message) (n int) {
	m := msg.(*Expr)
	// expr
	switch x := m.Expr.(type) {
	case *Expr_Value:
		s := proto1.Size(x.Value)
		n += proto1.SizeVarint(1<<3 | proto1.WireBytes)
		n += proto1.SizeVarint(uint64(s))
		n += s
	case *Expr_Param:
		s := proto1.Size(x.Param)
		n += proto1.SizeVarint(2<<3 | proto1.WireBytes)
		n += proto1.SizeVarint(uint64(s))
		n += s
	case *Expr_Operator:
		s := proto1.Size(x.Operator)
		n += proto1.SizeVarint(3<<3 | proto1.WireBytes)
		n += proto1.SizeVarint(uint64(s))
		n += s
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

type Value struct {
	Kind string `protobuf:"bytes,1,opt,name=kind" json:"kind,omitempty"`
	Type string `protobuf:"bytes,2,opt,name=type" json:"type,omitempty"`
	Data string `protobuf:"bytes,3,opt,name=data" json:"data,omitempty"`
}

func (m *Value) Reset()                    { *m = Value{} }
func (m *Value) String() string            { return proto1.CompactTextString(m) }
func (*Value) ProtoMessage()               {}
func (*Value) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

type Param struct {
	Kind string `protobuf:"bytes,1,opt,name=kind" json:"kind,omitempty"`
	Type string `protobuf:"bytes,2,opt,name=type" json:"type,omitempty"`
	Name string `protobuf:"bytes,3,opt,name=name" json:"name,omitempty"`
}

func (m *Param) Reset()                    { *m = Param{} }
func (m *Param) String() string            { return proto1.CompactTextString(m) }
func (*Param) ProtoMessage()               {}
func (*Param) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

type Operator struct {
	Kind     string  `protobuf:"bytes,1,opt,name=kind" json:"kind,omitempty"`
	Operands []*Expr `protobuf:"bytes,2,rep,name=operands" json:"operands,omitempty"`
}

func (m *Operator) Reset()                    { *m = Operator{} }
func (m *Operator) String() string            { return proto1.CompactTextString(m) }
func (*Operator) ProtoMessage()               {}
func (*Operator) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *Operator) GetOperands() []*Expr {
	if m != nil {
		return m.Operands
	}
	return nil
}

type Signature struct {
	ReturnType string            `protobuf:"bytes,1,opt,name=returnType" json:"returnType,omitempty"`
	ParamTypes map[string]string `protobuf:"bytes,2,rep,name=paramTypes" json:"paramTypes,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *Signature) Reset()                    { *m = Signature{} }
func (m *Signature) String() string            { return proto1.CompactTextString(m) }
func (*Signature) ProtoMessage()               {}
func (*Signature) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *Signature) GetParamTypes() map[string]string {
	if m != nil {
		return m.ParamTypes
	}
	return nil
}

type Versions struct {
	Versions []string `protobuf:"bytes,1,rep,name=versions" json:"versions,omitempty"`
}

func (m *Versions) Reset()                    { *m = Versions{} }
func (m *Versions) String() string            { return proto1.CompactTextString(m) }
func (*Versions) ProtoMessage()               {}
func (*Versions) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func init() {
	proto1.RegisterType((*Rules)(nil), "proto.Rules")
	proto1.RegisterType((*Rule)(nil), "proto.Rule")
	proto1.RegisterType((*Expr)(nil), "proto.Expr")
	proto1.RegisterType((*Value)(nil), "proto.Value")
	proto1.RegisterType((*Param)(nil), "proto.Param")
	proto1.RegisterType((*Operator)(nil), "proto.Operator")
	proto1.RegisterType((*Signature)(nil), "proto.Signature")
	proto1.RegisterType((*Versions)(nil), "proto.Versions")
}

func init() { proto1.RegisterFile("types.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 367 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x92, 0xc1, 0x4a, 0xfb, 0x40,
	0x10, 0xc6, 0x9b, 0x36, 0x29, 0xc9, 0xf4, 0x0f, 0xfd, 0xb3, 0x78, 0x08, 0x3d, 0x68, 0x5d, 0x44,
	0x45, 0xb0, 0x07, 0xbd, 0x88, 0x20, 0x88, 0x52, 0xec, 0x45, 0x94, 0x55, 0x7a, 0x5f, 0xe9, 0x22,
	0xa5, 0xed, 0x26, 0x6c, 0x36, 0xa5, 0x7d, 0x04, 0xdf, 0xc5, 0x87, 0x94, 0x99, 0xdd, 0x84, 0xb4,
	0x78, 0xf1, 0xd4, 0x6f, 0xe7, 0xfb, 0xf5, 0x9b, 0xd9, 0xd9, 0x40, 0xcf, 0x6e, 0x73, 0x55, 0x8c,
	0x72, 0x93, 0xd9, 0x8c, 0x45, 0xf4, 0xc3, 0x2f, 0x20, 0x12, 0xe5, 0x52, 0x15, 0xec, 0x18, 0x22,
	0x83, 0x22, 0x0d, 0x86, 0x9d, 0xf3, 0xde, 0x55, 0xcf, 0x61, 0x23, 0x34, 0x85, 0x73, 0xf8, 0x33,
	0x84, 0x78, 0x64, 0x47, 0x10, 0xaa, 0x4d, 0x6e, 0xd2, 0x60, 0x18, 0x34, 0xc8, 0xf1, 0x26, 0x37,
	0x82, 0x0c, 0x76, 0x02, 0x5d, 0xa3, 0x8a, 0x72, 0x69, 0xd3, 0x36, 0x21, 0xff, 0x3c, 0x32, 0x95,
	0xcb, 0x52, 0x09, 0xef, 0xf1, 0xaf, 0x00, 0xc2, 0xb1, 0xc3, 0xa3, 0x35, 0x3a, 0x3e, 0x70, 0x87,
	0x9e, 0xb4, 0x84, 0x33, 0x91, 0xca, 0xa5, 0x91, 0xab, 0xbd, 0xcc, 0x57, 0xac, 0x21, 0x45, 0x26,
	0xbb, 0x84, 0x38, 0xcb, 0x95, 0x91, 0x36, 0x33, 0x69, 0x87, 0xc0, 0xbe, 0x07, 0x5f, 0x7c, 0x79,
	0xd2, 0x12, 0x35, 0xf2, 0xd0, 0x75, 0x57, 0xe1, 0x8f, 0x10, 0x51, 0x3b, 0xc6, 0x20, 0x5c, 0xcc,
	0xf5, 0x8c, 0x46, 0x49, 0x04, 0x69, 0xac, 0xe1, 0xe6, 0xa8, 0x71, 0x22, 0x48, 0x63, 0x6d, 0x26,
	0xad, 0xa4, 0x1e, 0x89, 0x20, 0x8d, 0x21, 0x34, 0xcd, 0x5f, 0x42, 0xb4, 0x5c, 0xa9, 0x2a, 0x04,
	0x35, 0x7f, 0x82, 0xb8, 0x9a, 0xf4, 0xd7, 0x9c, 0x33, 0x7f, 0x41, 0x3d, 0x2b, 0xd2, 0xf6, 0xce,
	0x53, 0xd1, 0x03, 0xd4, 0x26, 0xff, 0x0e, 0x20, 0x79, 0x9b, 0x7f, 0x6a, 0x69, 0x4b, 0xa3, 0xd8,
	0x21, 0x80, 0x51, 0xb6, 0x34, 0xfa, 0x1d, 0x87, 0x70, 0x81, 0x8d, 0x0a, 0xbb, 0x07, 0xa0, 0x05,
	0xe2, 0xa1, 0x0a, 0x1e, 0xfa, 0xe0, 0x3a, 0xc5, 0x2d, 0x9b, 0x90, 0xb1, 0xb6, 0x66, 0x2b, 0x1a,
	0xff, 0x19, 0xdc, 0x41, 0x7f, 0xcf, 0x66, 0xff, 0xa1, 0xb3, 0x50, 0x5b, 0xdf, 0x0d, 0x25, 0x3b,
	0xa8, 0x9e, 0xda, 0xad, 0xc1, 0x1d, 0x6e, 0xdb, 0x37, 0x01, 0x3f, 0x85, 0x78, 0xaa, 0x4c, 0x31,
	0xcf, 0x74, 0xc1, 0x06, 0x10, 0xaf, 0xbd, 0xa6, 0xcf, 0x31, 0x11, 0xf5, 0xf9, 0xa3, 0x4b, 0x33,
	0x5d, 0xff, 0x04, 0x00, 0x00, 0xff, 0xff, 0xac, 0xa1, 0x69, 0xec, 0xcd, 0x02, 0x00, 0x00,
}
