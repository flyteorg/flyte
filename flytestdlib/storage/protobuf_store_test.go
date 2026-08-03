package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/protobuf/proto"
	errs "github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/flytestdlib/ioutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/stow/s3"
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

func TestDefaultProtobufStore(t *testing.T) {
	t.Run("Read after Write", func(t *testing.T) {
		testScope := promutils.NewTestScope()
		s, err := NewDataStore(&Config{Type: TypeMemory}, testScope)
		assert.NoError(t, err)

		err = s.WriteProtobuf(context.TODO(), "hello", Options{}, &mockProtoMessage{X: 5})
		assert.NoError(t, err)

		m := &mockProtoMessage{}
		err = s.ReadProtobuf(context.TODO(), "hello", m)
		assert.NoError(t, err)
		assert.Equal(t, int64(5), m.X)
	})

	t.Run("RefreshConfig", func(t *testing.T) {
		testScope := promutils.NewTestScope()
		s, err := NewDataStore(&Config{Type: TypeMemory}, testScope)
		require.NoError(t, err)
		require.IsType(t, DefaultProtobufStore{}, s.ComposedProtobufStore)
		// The raw store is always the multi-scheme routing store; its primary backend here is the
		// in-memory store selected by Type: mem.
		require.IsType(t, &routingStore{}, s.ComposedProtobufStore.(DefaultProtobufStore).RawStore)
		require.IsType(t, &InMemoryStore{}, s.ComposedProtobufStore.(DefaultProtobufStore).RawStore.(*routingStore).primaryStore)

		oldMetrics := s.metrics
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		err = s.RefreshConfig(context.TODO(), &Config{
			Type: TypeMinio,
			Stow: StowConfig{
				Kind: TypeS3,
				Config: map[string]string{
					s3.ConfigAccessKeyID: "key",
					s3.ConfigSecretKey:   "sec",
					s3.ConfigEndpoint:    server.URL,
				}},
			InitContainer: "b"})

		assert.NoError(t, err)
		require.IsType(t, DefaultProtobufStore{}, s.ComposedProtobufStore)
		require.IsType(t, &routingStore{}, s.ComposedProtobufStore.(DefaultProtobufStore).RawStore)
		assert.IsType(t, &StowStore{}, s.ComposedProtobufStore.(DefaultProtobufStore).RawStore.(*routingStore).primaryStore)
		assert.Equal(t, oldMetrics, s.metrics)
	})

	t.Run("invalid type", func(t *testing.T) {
		testScope := promutils.NewTestScope()

		_, err := NewDataStore(&Config{Type: "invalid"}, testScope)

		assert.EqualError(t, err, "type is of an invalid value [invalid]")
	})

	t.Run("coudln't create store", func(t *testing.T) {
		testScope := promutils.NewTestScope()

		_, err := NewDataStore(&Config{Type: TypeS3}, testScope)

		assert.EqualError(t, err, "initContainer is required even with `enable-multicontainer`")
	})
}

func TestDefaultProtobufStoreByteMetrics(t *testing.T) {
	message := &mockProtoMessage{X: 5}
	raw, err := proto.Marshal(message)
	require.NoError(t, err)

	store := &dummyStore{
		WriteRawCb: func(_ context.Context, _ DataReference, size int64, _ Options, reader io.Reader) error {
			actual, err := io.ReadAll(reader)
			require.NoError(t, err)
			require.Equal(t, int64(len(raw)), size)
			require.True(t, bytes.Equal(raw, actual))
			return nil
		},
		ReadRawCb: func(_ context.Context, _ DataReference) (io.ReadCloser, error) {
			return ioutils.NewBytesReadCloser(raw), nil
		},
	}
	scope := promutils.NewTestScope()
	protoMetrics := newProtoMetrics(scope, scope)
	protobufStore := NewDefaultProtobufStoreWithMetrics(store, protoMetrics)

	require.NoError(t, protobufStore.WriteProtobuf(context.Background(), "key", Options{}, message))
	require.Equal(t, float64(len(raw)), testutil.ToFloat64(protoMetrics.WrittenBytes))

	readMessage := &mockProtoMessage{}
	require.NoError(t, protobufStore.ReadProtobuf(context.Background(), "key", readMessage))
	require.Equal(t, float64(len(raw)), testutil.ToFloat64(protoMetrics.ReadBytes))
	require.Equal(t, message.X, readMessage.X)
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
			return MemoryMetadata{}, fmt.Errorf("%s", dummyHeadErrorMsg) //nolint:govet,staticcheck
		},
		WriteRawCb: func(ctx context.Context, reference DataReference, size int64, opts Options, raw io.Reader) error {
			return fmt.Errorf("%s", dummyWriteErrorMsg) //nolint:govet,staticcheck
		},
		ReadRawCb: func(ctx context.Context, reference DataReference) (io.ReadCloser, error) {
			return nil, fmt.Errorf("%s", dummyReadErrorMsg) //nolint:govet,staticcheck
		},
	}
	pbErroneousStore := NewDefaultProtobufStoreWithMetrics(store, metrics.protoMetrics)
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
