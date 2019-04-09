package storage

import (
	"context"
	"testing"

	"github.com/lyft/flytestdlib/promutils"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

type mockProtoMessage struct {
	X int64 `protobuf:"varint,2,opt,name=x,json=x,proto3" json:"x,omitempty"`
}

func (mockProtoMessage) Reset() {
}

func (m mockProtoMessage) String() string {
	return proto.CompactTextString(m)
}

func (mockProtoMessage) ProtoMessage() {
}

func TestDefaultProtobufStore_ReadProtobuf(t *testing.T) {
	t.Run("Read after Write", func(t *testing.T) {
		testScope := promutils.NewTestScope()
		s, err := NewDataStore(&Config{Type: TypeMemory}, testScope)
		assert.NoError(t, err)

		err = s.WriteProtobuf(context.TODO(), DataReference("hello"), Options{}, &mockProtoMessage{X: 5})
		assert.NoError(t, err)

		m := &mockProtoMessage{}
		err = s.ReadProtobuf(context.TODO(), DataReference("hello"), m)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), m.X)
	})
}
