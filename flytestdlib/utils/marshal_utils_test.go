package utils

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/golang/protobuf/proto"
	structpb "github.com/golang/protobuf/ptypes/struct"
)

type SimpleType struct {
	StringValue string `json:"string_value,omitempty"`
}

// Simple proto
type TestProto struct {
	StringValue string `protobuf:"bytes,3,opt,name=string_value,json=stringValue,proto3" json:"string_value,omitempty"`
}

func (m *TestProto) Reset()         { *m = TestProto{} }
func (m *TestProto) String() string { return proto.CompactTextString(m) }
func (*TestProto) ProtoMessage()    {}
func (*TestProto) Descriptor() ([]byte, []int) {
	return []byte{}, []int{0}
}
func (m *TestProto) GetWorkflowID() string {
	if m != nil {
		return m.StringValue
	}
	return ""
}

func init() {
	proto.RegisterType((*TestProto)(nil), "test.package.TestProto")
}

func TestMarshalPbToString(t *testing.T) {
	type args struct {
		msg proto.Message
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"empty", args{msg: &TestProto{}}, "{}", false},
		{"has value", args{msg: &TestProto{StringValue: "hello"}}, `{"stringValue":"hello"}`, false},
		{"nil input", args{msg: nil}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MarshalPbToString(tt.args.msg)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalToString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MarshalToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMarshalObjToStruct(t *testing.T) {
	type args struct {
		input interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    *structpb.Struct
		wantErr bool
	}{
		{"has proto value", args{input: &TestProto{StringValue: "hello"}}, &structpb.Struct{Fields: map[string]*structpb.Value{
			"stringValue": {Kind: &structpb.Value_StringValue{StringValue: "hello"}},
		}}, false},
		{"has struct value", args{input: SimpleType{StringValue: "hello"}}, &structpb.Struct{Fields: map[string]*structpb.Value{
			"string_value": {Kind: &structpb.Value_StringValue{StringValue: "hello"}},
		}}, false},
		{"has string value", args{input: "hello"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MarshalObjToStruct(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalObjToStruct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalObjToStruct() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnmarshalStructToPb(t *testing.T) {
	type args struct {
		structObj *structpb.Struct
		msg       proto.Message
	}
	tests := []struct {
		name     string
		args     args
		expected proto.Message
		wantErr  bool
	}{
		{"empty", args{structObj: &structpb.Struct{Fields: map[string]*structpb.Value{}}, msg: &TestProto{}}, &TestProto{}, false},
		{"has value", args{structObj: &structpb.Struct{Fields: map[string]*structpb.Value{
			"stringValue": {Kind: &structpb.Value_StringValue{StringValue: "hello"}},
		}}, msg: &TestProto{}}, &TestProto{StringValue: "hello"}, false},
		{"nil input", args{structObj: nil}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := UnmarshalStructToPb(tt.args.structObj, tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalStructToPb() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				assert.Equal(t, tt.expected, tt.args.msg)
			}
		})
	}
}

func TestMarshalPbToStruct(t *testing.T) {
	type args struct {
		in proto.Message
	}
	tests := []struct {
		name     string
		args     args
		expected *structpb.Struct
		wantErr  bool
	}{
		{"empty", args{in: &TestProto{}}, &structpb.Struct{Fields: map[string]*structpb.Value{}}, false},
		{"has value", args{in: &TestProto{StringValue: "hello"}}, &structpb.Struct{Fields: map[string]*structpb.Value{
			"stringValue": {Kind: &structpb.Value_StringValue{StringValue: "hello"}},
		}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := MarshalPbToStruct(tt.args.in); (err != nil) != tt.wantErr {
				t.Errorf("MarshalPbToStruct() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				assert.Equal(t, tt.expected.Fields, got.Fields)
			}
		})
	}
}
