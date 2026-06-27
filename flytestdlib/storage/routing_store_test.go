package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
)

// newRoutingDataStore builds a DataStore whose primary backend is the in-memory store and whose
// secondary redis:// scheme is dialed lazily against a miniredis. cfg.Redis.Addr is set so the
// redis backend resolves to the test server rather than the reference host.
func newRoutingDataStore(t *testing.T) (*DataStore, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	ds, err := NewDataStore(&Config{
		Type:  TypeMemory,
		Redis: RedisConfig{Addr: mr.Addr()},
	}, promutils.NewTestScope())
	require.NoError(t, err)
	return ds, mr
}

func TestRoutingStore_DispatchByScheme(t *testing.T) {
	ds, mr := newRoutingDataStore(t)
	redisRef := DataReference(fmt.Sprintf("redis://%s/meta/inputs.pb", mr.Addr()))
	memRef := DataReference("mem://container/raw/data.bin")

	require.NoError(t, ds.WriteRaw(context.TODO(), redisRef, 4, Options{}, bytes.NewReader([]byte("meta"))))
	require.NoError(t, ds.WriteRaw(context.TODO(), memRef, 3, Options{}, bytes.NewReader([]byte("raw"))))

	// The redis write landed in redis, keyed by the path portion.
	got, err := mr.Get("meta/inputs.pb")
	require.NoError(t, err)
	assert.Equal(t, "meta", got)

	// Each reference reads back through its own backend.
	rc, err := ds.ReadRaw(context.TODO(), redisRef)
	require.NoError(t, err)
	data, err := io.ReadAll(rc)
	require.NoError(t, err)
	assert.Equal(t, []byte("meta"), data)

	rc, err = ds.ReadRaw(context.TODO(), memRef)
	require.NoError(t, err)
	data, err = io.ReadAll(rc)
	require.NoError(t, err)
	assert.Equal(t, []byte("raw"), data)

	md, err := ds.Head(context.TODO(), redisRef)
	require.NoError(t, err)
	assert.True(t, md.Exists())
	assert.Equal(t, int64(4), md.Size())
}

func TestRoutingStore_CrossBackendCopy(t *testing.T) {
	ds, mr := newRoutingDataStore(t)
	src := DataReference("mem://container/raw/src.bin")
	dst := DataReference(fmt.Sprintf("redis://%s/meta/dst.bin", mr.Addr()))

	require.NoError(t, ds.WriteRaw(context.TODO(), src, 5, Options{}, bytes.NewReader([]byte("hello"))))
	require.NoError(t, ds.CopyRaw(context.TODO(), src, dst, Options{}))

	got, err := mr.Get("meta/dst.bin")
	require.NoError(t, err)
	assert.Equal(t, "hello", got)
}

func TestRoutingStore_BaseContainerIsPrimaryStore(t *testing.T) {
	ds, _ := newRoutingDataStore(t)
	// The base FQN is the primary (in-memory) store's, which is empty — not the redis base.
	assert.Equal(t, DataReference(""), ds.GetBaseContainerFQN(context.TODO()))
}

// TestRoutingStore_RedisDialsReferenceHostWithoutAddr exercises the lazy-dial fallback: with no
// redis.addr (and no per-scheme override) configured, a redis:// reference resolves its address
// from the host portion of the reference itself and "just works".
func TestRoutingStore_RedisDialsReferenceHostWithoutAddr(t *testing.T) {
	mr := miniredis.RunT(t)
	ds, err := NewDataStore(&Config{Type: TypeMemory}, promutils.NewTestScope())
	require.NoError(t, err)

	ref := DataReference(fmt.Sprintf("redis://%s/meta/inputs.pb", mr.Addr()))
	require.NoError(t, ds.WriteRaw(context.TODO(), ref, 4, Options{}, bytes.NewReader([]byte("meta"))))

	got, err := mr.Get("meta/inputs.pb")
	require.NoError(t, err)
	assert.Equal(t, "meta", got)
}

func TestRoutingStore_DeleteRoutes(t *testing.T) {
	ds, mr := newRoutingDataStore(t)
	ref := DataReference(fmt.Sprintf("redis://%s/meta/doomed.pb", mr.Addr()))
	require.NoError(t, ds.WriteRaw(context.TODO(), ref, 1, Options{}, bytes.NewReader([]byte("x"))))
	require.NoError(t, ds.Delete(context.TODO(), ref))
	md, err := ds.Head(context.TODO(), ref)
	require.NoError(t, err)
	assert.False(t, md.Exists())
}
