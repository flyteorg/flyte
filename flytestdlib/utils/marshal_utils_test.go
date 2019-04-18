package utils

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/golang/protobuf/proto"
	structpb "github.com/golang/protobuf/ptypes/struct"
)

// Simple proto
type TestProto struct {
	StringValue          string   `protobuf:"bytes,3,opt,name=string_value,json=stringValue,proto3" json:"string_value,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TestProto) Reset()         { *m = TestProto{} }
func (m *TestProto) String() string { return proto.CompactTextString(m) }
func (*TestProto) ProtoMessage()    {}
func (*TestProto) Descriptor() ([]byte, []int) {
	return []byte{}, []int{0}
}
func (m *TestProto) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TestProto.Unmarshal(m, b)
}
func (m *TestProto) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TestProto.Marshal(b, m, deterministic)
}
func (dst *TestProto) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TestProto.Merge(dst, src)
}
func (m *TestProto) XXX_Size() int {
	return xxx_messageInfo_TestProto.Size(m)
}
func (m *TestProto) XXX_DiscardUnknown() {
	xxx_messageInfo_TestProto.DiscardUnknown(m)
}

var xxx_messageInfo_TestProto proto.InternalMessageInfo

func (m *TestProto) GetWorkflowId() string {
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
		{"has value", args{input: &TestProto{StringValue: "hello"}}, &structpb.Struct{Fields: map[string]*structpb.Value{
			"string_value": {Kind: &structpb.Value_StringValue{StringValue: "hello"}},
		}}, false},
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
		in  proto.Message
		out *structpb.Struct
	}
	tests := []struct {
		name     string
		args     args
		expected *structpb.Struct
		wantErr  bool
	}{
		{"empty", args{in: &TestProto{}, out: &structpb.Struct{}}, &structpb.Struct{Fields: map[string]*structpb.Value{}}, false},
		{"has value", args{in: &TestProto{StringValue: "hello"}, out: &structpb.Struct{}}, &structpb.Struct{Fields: map[string]*structpb.Value{
			"stringValue": {Kind: &structpb.Value_StringValue{StringValue: "hello"}},
		}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := MarshalPbToStruct(tt.args.in, tt.args.out); (err != nil) != tt.wantErr {
				t.Errorf("MarshalPbToStruct() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				assert.Equal(t, tt.expected.Fields, tt.args.out.Fields)
			}
		})
	}
}
