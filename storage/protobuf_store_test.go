package storage

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"testing"

	"github.com/lyft/flytestdlib/promutils"

	"github.com/golang/protobuf/proto"
	errs "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type mockProtoMessage struct {
	X int64 `protobuf:"varint,2,opt,name=x,json=x,proto3" json:"x,omitempty"`
}

type mockBigDataProtoMessage struct {
	X []byte `protobuf:"bytes,1,opt,name=X,proto3" json:"X,omitempty"`
}

func (mockProtoMessage) Reset() {
}

func (m mockProtoMessage) String() string {
	return proto.CompactTextString(m)
}

func (mockProtoMessage) ProtoMessage() {
}

func (mockBigDataProtoMessage) Reset() {
}

func (m mockBigDataProtoMessage) String() string {
	return proto.CompactTextString(m)
}

func (mockBigDataProtoMessage) ProtoMessage() {
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

func TestDefaultProtobufStore_BigDataReadAfterWrite(t *testing.T) {
	t.Run("Read after Write with Big Data", func(t *testing.T) {
		testScope := promutils.NewTestScope()

		s, err := NewDataStore(
			&Config{
				Type: TypeMemory,
				Cache: CachingConfig{
					MaxSizeMegabytes: 1,
					TargetGCPercent:  20,
				},
			}, testScope)
		assert.NoError(t, err)

		bigD := make([]byte, 1.5*1024*1024)
		// #nosec G404
		_, err = rand.Read(bigD)
		assert.NoError(t, err)

		mockMessage := &mockBigDataProtoMessage{X: bigD}

		err = s.WriteProtobuf(context.TODO(), DataReference("bigK"), Options{}, mockMessage)
		assert.NoError(t, err)

		m := &mockBigDataProtoMessage{}
		err = s.ReadProtobuf(context.TODO(), DataReference("bigK"), m)
		assert.NoError(t, err)
		assert.Equal(t, bigD, m.X)

	})
}

func TestDefaultProtobufStore_HardErrors(t *testing.T) {
	ctx := context.TODO()
	k1 := DataReference("k1")
	dummyHeadErrorMsg := "Dummy head error"
	dummyWriteErrorMsg := "Dummy write error"
	dummyReadErrorMsg := "Dummy read error"
	store := &dummyStore{
		HeadCb: func(ctx context.Context, reference DataReference) (Metadata, error) {
			return MemoryMetadata{}, fmt.Errorf(dummyHeadErrorMsg)
		},
		WriteRawCb: func(ctx context.Context, reference DataReference, size int64, opts Options, raw io.Reader) error {
			return fmt.Errorf(dummyWriteErrorMsg)
		},
		ReadRawCb: func(ctx context.Context, reference DataReference) (io.ReadCloser, error) {
			return nil, fmt.Errorf(dummyReadErrorMsg)
		},
	}
	testScope := promutils.NewTestScope()
	pbErroneousStore := NewDefaultProtobufStore(store, testScope)
	t.Run("Test if hard write errors are handled correctly", func(t *testing.T) {
		err := pbErroneousStore.WriteProtobuf(ctx, k1, Options{}, &mockProtoMessage{X: 5})
		assert.False(t, IsFailedWriteToCache(err))
		assert.Equal(t, dummyWriteErrorMsg, errs.Cause(err).Error())
	})

	t.Run("Test if hard read errors are handled correctly", func(t *testing.T) {
		m := &mockProtoMessage{}
		err := pbErroneousStore.ReadProtobuf(ctx, k1, m)
		assert.False(t, IsFailedWriteToCache(err))
		assert.Equal(t, dummyReadErrorMsg, errs.Cause(err).Error())
	})
}
