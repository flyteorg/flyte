package pbhash

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var sampleTime = timestamppb.New(time.Date(2019, 03, 29, 12, 0, 0, 0, time.UTC))

func makeStruct(t *testing.T, fields map[string]interface{}) *structpb.Struct {
	t.Helper()

	s, err := structpb.NewStruct(fields)
	require.NoError(t, err)
	return s
}

func TestProtoHash(t *testing.T) {
	mockProto := makeStruct(t, map[string]interface{}{
		"integer":     18,
		"floatValue":  1.3,
		"stringValue": "lets test this",
		"boolean":     true,
		"datetime":    sampleTime.AsTime().Format(time.RFC3339Nano),
		"duration":    durationpb.New(time.Millisecond).AsDuration().String(),
		"mapValue": map[string]interface{}{
			"z": "last",
			"a": "first",
		},
		"collections": []interface{}{"1", "2", "3"},
	})

	expectedHashedMockProto := []byte{0x45, 0xd1, 0xe, 0x9, 0x5e, 0xe3, 0xf7, 0x3e, 0xe9, 0x9, 0xe9, 0xc9, 0x27, 0xd6,
		0xf5, 0x79, 0x81, 0xf6, 0x52, 0x48, 0x3f, 0x71, 0x8c, 0x2, 0x87, 0x1, 0x98, 0x58, 0x5b, 0x7e, 0xf, 0xda}
	expectedHashString := "RdEOCV7j9z7pCenJJ9b1eYH2Ukg/cYwChwGYWFt+D9o="

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
		mockProto.Fields["mapValue"] = structpb.NewStructValue(makeStruct(t, map[string]interface{}{
			"a": "first",
			"z": "last",
		}))
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

	mockProtoOmitEmpty := makeStruct(t, map[string]interface{}{
		"integer":     18,
		"floatValue":  1.3,
		"stringValue": "lets test this",
		"boolean":     true,
	})

	expectedHashedMockProtoOmitEmpty := []byte{0x6d, 0xfa, 0xc1, 0xc2, 0xe0, 0xee, 0xad, 0xe2, 0xa5, 0xad, 0x7d, 0x9e,
		0xad, 0x1c, 0x94, 0x11, 0x6a, 0x21, 0x23, 0xe1, 0xfb, 0xe2, 0x35, 0xd5, 0x37, 0x89, 0xf3, 0xfc, 0xa, 0xfb,
		0x3d, 0xe9}

	expectedHashStringOmitEmpty := "bfrBwuDureKlrX2erRyUEWohI+H74jXVN4nz/Ar7Pek="

	t.Run("TestPartial", func(t *testing.T) {
		hashedBytes, err := ComputeHash(context.Background(), mockProtoOmitEmpty)
		assert.Nil(t, err)
		assert.Equal(t, expectedHashedMockProtoOmitEmpty, hashedBytes)
		assert.Len(t, hashedBytes, 32)

		hashedString, err := ComputeHashString(context.Background(), mockProtoOmitEmpty)
		assert.Nil(t, err)
		assert.Equal(t, hashedString, expectedHashStringOmitEmpty)
	})

	mockOldProtoMessage := makeStruct(t, map[string]interface{}{
		"integer":     18,
		"floatValue":  1.3,
		"stringValue": "lets test this",
		"boolean":     true,
	})

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
