package pbhash

import (
	"context"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/assert"
)

// Mock a Protobuf generated GO object
type mockProtoMessage struct {
	Integer     int64                `protobuf:"varint,1,opt,name=integer,proto3" json:"integer,omitempty"`
	FloatValue  float64              `protobuf:"fixed64,2,opt,name=float_value,json=floatValue,proto3" json:"float_value,omitempty"`
	StringValue string               `protobuf:"bytes,3,opt,name=string_value,json=stringValue,proto3" json:"string_value,omitempty"`
	Boolean     bool                 `protobuf:"varint,4,opt,name=boolean,proto3" json:"boolean,omitempty"`
	Datetime    *timestamp.Timestamp `protobuf:"bytes,5,opt,name=datetime,proto3" json:"datetime,omitempty"`
	Duration    *duration.Duration   `protobuf:"bytes,6,opt,name=duration,proto3" json:"duration,omitempty"`
	MapValue    map[string]string    `protobuf:"bytes,7,rep,name=map_value,json=mapValue,proto3" json:"map_value,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	Collections []string             `protobuf:"bytes,8,rep,name=collections,proto3" json:"collections,omitempty"`
}

func (mockProtoMessage) Reset() {
}

func (m mockProtoMessage) String() string {
	return proto.CompactTextString(m)
}

func (mockProtoMessage) ProtoMessage() {
}

// Mock an older version of the above pb object that doesn't have some fields
type mockOlderProto struct {
	Integer     int64   `protobuf:"varint,1,opt,name=integer,proto3" json:"integer,omitempty"`
	FloatValue  float64 `protobuf:"fixed64,2,opt,name=float_value,json=floatValue,proto3" json:"float_value,omitempty"`
	StringValue string  `protobuf:"bytes,3,opt,name=string_value,json=stringValue,proto3" json:"string_value,omitempty"`
	Boolean     bool    `protobuf:"varint,4,opt,name=boolean,proto3" json:"boolean,omitempty"`
}

func (mockOlderProto) Reset() {
}

func (m mockOlderProto) String() string {
	return proto.CompactTextString(m)
}

func (mockOlderProto) ProtoMessage() {
}

var sampleTime, _ = ptypes.TimestampProto(
	time.Date(2019, 03, 29, 12, 0, 0, 0, time.UTC))

func TestProtoHash(t *testing.T) {
	mockProto := &mockProtoMessage{
		Integer:     18,
		FloatValue:  1.3,
		StringValue: "lets test this",
		Boolean:     true,
		Datetime:    sampleTime,
		Duration:    ptypes.DurationProto(time.Millisecond),
		MapValue: map[string]string{
			"z": "last",
			"a": "first",
		},
		Collections: []string{"1", "2", "3"},
	}

	expectedHashedMockProto := []byte{0x62, 0x95, 0xb2, 0x2c, 0x23, 0xf5, 0x35, 0x6d, 0x3, 0x56, 0x4d, 0xc7, 0x8f, 0xae,
		0x2d, 0x2b, 0xbd, 0x7, 0xff, 0xdb, 0x7e, 0xe5, 0xf4, 0x25, 0x8f, 0xbc, 0xb2, 0xc, 0xad, 0xa5, 0x48, 0x44}
	expectedHashString := "YpWyLCP1NW0DVk3Hj64tK70H/9t+5fQlj7yyDK2lSEQ="

	t.Run("TestFullProtoHash", func(t *testing.T) {
		hashedBytes, err := ComputeHash(context.Background(), mockProto)
		assert.Nil(t, err)
		assert.Equal(t, expectedHashedMockProto, hashedBytes)
		assert.Len(t, hashedBytes, 32)

		hashedString, err := ComputeHashString(context.Background(), mockProto)
		assert.Nil(t, err)
		assert.Equal(t, hashedString, expectedHashString)
	})

	t.Run("TestFullProtoHashReorderKeys", func(t *testing.T) {
		mockProto.MapValue = map[string]string{"a": "first", "z": "last"}
		hashedBytes, err := ComputeHash(context.Background(), mockProto)
		assert.Nil(t, err)
		assert.Equal(t, expectedHashedMockProto, hashedBytes)
		assert.Len(t, hashedBytes, 32)

		hashedString, err := ComputeHashString(context.Background(), mockProto)
		assert.Nil(t, err)
		assert.Equal(t, hashedString, expectedHashString)
	})
}

func TestPartialFilledProtoHash(t *testing.T) {

	mockProtoOmitEmpty := &mockProtoMessage{
		Integer:     18,
		FloatValue:  1.3,
		StringValue: "lets test this",
		Boolean:     true,
	}

	expectedHashedMockProtoOmitEmpty := []byte{0x1a, 0x13, 0xcc, 0x4c, 0xab, 0xc9, 0x7d, 0x43, 0xc7, 0x2b, 0xc5, 0x37,
		0xbc, 0x49, 0xa8, 0x8b, 0xfc, 0x1d, 0x54, 0x1c, 0x7b, 0x21, 0x04, 0x8f, 0xab, 0x28, 0xc6, 0x5c, 0x06, 0x73,
		0xaa, 0xe2}

	expectedHashStringOmitEmpty := "GhPMTKvJfUPHK8U3vEmoi/wdVBx7IQSPqyjGXAZzquI="

	t.Run("TestPartial", func(t *testing.T) {
		hashedBytes, err := ComputeHash(context.Background(), mockProtoOmitEmpty)
		assert.Nil(t, err)
		assert.Equal(t, expectedHashedMockProtoOmitEmpty, hashedBytes)
		assert.Len(t, hashedBytes, 32)

		hashedString, err := ComputeHashString(context.Background(), mockProtoOmitEmpty)
		assert.Nil(t, err)
		assert.Equal(t, hashedString, expectedHashStringOmitEmpty)
	})

	mockOldProtoMessage := &mockOlderProto{
		Integer:     18,
		FloatValue:  1.3,
		StringValue: "lets test this",
		Boolean:     true,
	}

	t.Run("TestOlderProto", func(t *testing.T) {
		hashedBytes, err := ComputeHash(context.Background(), mockOldProtoMessage)
		assert.Nil(t, err)
		assert.Equal(t, expectedHashedMockProtoOmitEmpty, hashedBytes)
		assert.Len(t, hashedBytes, 32)

		hashedString, err := ComputeHashString(context.Background(), mockProtoOmitEmpty)
		assert.Nil(t, err)
		assert.Equal(t, hashedString, expectedHashStringOmitEmpty)
	})

}
