package cache

import (
	"context"
	"testing"

	"github.com/eko/gocache/lib/v4/marshaler"
	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/flyteorg/flyte/flytestdlib/cache/mocks"
)

type DummyType struct {
	Something string
}

func TestMarshaler(t *testing.T) {
	ctx := context.Background()

	mockCache := mocks.NewMockCache[any](false)
	obj := DummyType{Something: "value"}
	raw, err := msgpack.Marshal(obj)
	require.NoError(t, err)
	require.NoError(t, mockCache.Set(ctx, "key", raw))

	marshalerCache := marshaler.New(mockCache)
	require.NoError(t, marshalerCache.Set(ctx, "key", obj))
	retObj, err := marshalerCache.Get(ctx, "key", new(DummyType))
	require.NoError(t, err)
	require.Equal(t, &obj, retObj)
}

func TestNamespaced(t *testing.T) {
	ctx := context.Background()

	mockCache := mocks.NewMockCache[any](false)
	obj := DummyType{Something: "value"}
	raw, err := msgpack.Marshal(obj)
	require.NoError(t, err)
	require.NoError(t, mockCache.Set(ctx, "namespace:key", raw))

	typedMarshaler := NewMsgPackMarshaler(mockCache)
	namespacedCache := NewNamespacedCache[DummyType]("namespace", NewTypedMarshaler[DummyType](typedMarshaler))
	require.NoError(t, namespacedCache.Set(ctx, "key", obj))
	retObj, err := namespacedCache.Get(ctx, "key")
	require.NoError(t, err)
	require.Equal(t, obj, retObj)
}
