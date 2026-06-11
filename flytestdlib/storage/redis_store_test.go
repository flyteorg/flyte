package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
)

func newTestRedisStore(t *testing.T) (RawStore, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	store, err := NewRedisRawStore(context.TODO(), &Config{
		Type:  TypeRedis,
		Redis: RedisConfig{Addr: mr.Addr()},
	}, metrics)
	require.NoError(t, err)
	return store, mr
}

func redisRef(mr *miniredis.Miniredis, key string) DataReference {
	return DataReference(fmt.Sprintf("redis://%s/%s", mr.Addr(), key))
}

func TestNewRedisRawStore_RequiresAddr(t *testing.T) {
	_, err := NewRedisRawStore(context.TODO(), &Config{Type: TypeRedis}, metrics)
	assert.Error(t, err)
}

func TestRedisStore_WriteReadRoundTrip(t *testing.T) {
	store, mr := newTestRedisStore(t)
	ref := redisRef(mr, "flyte/runs/r1/a0/0/inputs.pb")
	payload := []byte("serialized-literalmap")

	err := store.WriteRaw(context.TODO(), ref, int64(len(payload)), Options{}, bytes.NewReader(payload))
	require.NoError(t, err)

	reader, err := store.ReadRaw(context.TODO(), ref)
	require.NoError(t, err)
	defer func() { _ = reader.Close() }()
	data, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, payload, data)
}

func TestRedisStore_ReadRawMissing(t *testing.T) {
	store, mr := newTestRedisStore(t)
	_, err := store.ReadRaw(context.TODO(), redisRef(mr, "missing"))
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestRedisStore_ReadRawLimitExceeded(t *testing.T) {
	store, mr := newTestRedisStore(t)
	ref := redisRef(mr, "big")
	payload := bytes.Repeat([]byte("x"), int(MiB)+1)
	require.NoError(t, store.WriteRaw(context.TODO(), ref, int64(len(payload)), Options{}, bytes.NewReader(payload)))

	prevLimit := GetConfig().Limits.GetLimitMegabytes
	GetConfig().Limits.GetLimitMegabytes = 1
	defer func() { GetConfig().Limits.GetLimitMegabytes = prevLimit }()

	_, err := store.ReadRaw(context.TODO(), ref)
	require.Error(t, err)
	assert.True(t, IsExceedsLimit(err))
}

func TestRedisStore_Head(t *testing.T) {
	store, mr := newTestRedisStore(t)

	md, err := store.Head(context.TODO(), redisRef(mr, "missing"))
	require.NoError(t, err)
	assert.False(t, md.Exists())

	ref := redisRef(mr, "present")
	require.NoError(t, store.WriteRaw(context.TODO(), ref, 3, Options{}, bytes.NewReader([]byte("abc"))))
	md, err = store.Head(context.TODO(), ref)
	require.NoError(t, err)
	assert.True(t, md.Exists())
	assert.Equal(t, int64(3), md.Size())
}

func TestRedisStore_HeadEmptyValue(t *testing.T) {
	// A task with no inputs serializes to zero bytes; the key must still exist.
	store, mr := newTestRedisStore(t)
	ref := redisRef(mr, "empty")
	require.NoError(t, store.WriteRaw(context.TODO(), ref, 0, Options{}, bytes.NewReader(nil)))

	md, err := store.Head(context.TODO(), ref)
	require.NoError(t, err)
	assert.True(t, md.Exists())
	assert.Equal(t, int64(0), md.Size())
}

func TestRedisStore_Delete(t *testing.T) {
	store, mr := newTestRedisStore(t)
	ref := redisRef(mr, "doomed")
	require.NoError(t, store.WriteRaw(context.TODO(), ref, 1, Options{}, bytes.NewReader([]byte("x"))))

	require.NoError(t, store.Delete(context.TODO(), ref))
	md, err := store.Head(context.TODO(), ref)
	require.NoError(t, err)
	assert.False(t, md.Exists())

	assert.ErrorIs(t, store.Delete(context.TODO(), ref), os.ErrNotExist)
}

func TestRedisStore_List(t *testing.T) {
	store, mr := newTestRedisStore(t)
	keys := []string{"runs/r1/a.pb", "runs/r1/b.pb", "runs/r1/sub/c.pb"}
	for _, k := range keys {
		require.NoError(t, store.WriteRaw(context.TODO(), redisRef(mr, k), 1, Options{}, bytes.NewReader([]byte("x"))))
	}
	otherRef := redisRef(mr, "runs/r2/d.pb")
	require.NoError(t, store.WriteRaw(context.TODO(), otherRef, 1, Options{}, bytes.NewReader([]byte("x"))))

	seen := map[DataReference]bool{}
	cursor := NewCursorAtStart()
	for {
		items, nextCursor, err := store.List(context.TODO(), redisRef(mr, "runs/r1"), 2, cursor)
		require.NoError(t, err)
		for _, item := range items {
			seen[item] = true
		}
		if IsCursorEnd(nextCursor) {
			break
		}
		cursor = nextCursor
	}

	assert.Len(t, seen, len(keys))
	for _, k := range keys {
		assert.True(t, seen[redisRef(mr, k)], "missing %v", k)
	}
}

func TestRedisStore_RejectsNonRedisReference(t *testing.T) {
	store, _ := newTestRedisStore(t)
	_, err := store.ReadRaw(context.TODO(), DataReference("s3://bucket/key"))
	assert.Error(t, err)
}

func TestRedisStore_GetBaseContainerFQN(t *testing.T) {
	store, mr := newTestRedisStore(t)
	assert.Equal(t, DataReference("redis://"+mr.Addr()), store.GetBaseContainerFQN(context.TODO()))
}

func TestRedisStore_CreateSignedURLUnsupported(t *testing.T) {
	store, mr := newTestRedisStore(t)
	_, err := store.CreateSignedURL(context.TODO(), redisRef(mr, "x"), SignedURLProperties{})
	assert.Error(t, err)
}

func TestRedisStore_ThroughDataStore(t *testing.T) {
	// End to end through NewDataStore: type selection, reference construction, copy and raw IO.
	mr := miniredis.RunT(t)
	testScope := promutils.NewTestScope()
	ds, err := NewDataStore(&Config{
		Type:  TypeRedis,
		Redis: RedisConfig{Addr: mr.Addr()},
	}, testScope)
	require.NoError(t, err)

	base := ds.GetBaseContainerFQN(context.TODO())
	ref, err := ds.ConstructReference(context.TODO(), base, "metadata", "runs", "r1", "inputs.pb")
	require.NoError(t, err)
	assert.Equal(t, redisRef(mr, "metadata/runs/r1/inputs.pb"), ref)

	payload := []byte("through the datastore")
	require.NoError(t, ds.WriteRaw(context.TODO(), ref, int64(len(payload)), Options{}, bytes.NewReader(payload)))

	dest := redisRef(mr, "metadata/runs/r1/inputs-copy.pb")
	require.NoError(t, ds.CopyRaw(context.TODO(), ref, dest, Options{}))

	reader, err := ds.ReadRaw(context.TODO(), dest)
	require.NoError(t, err)
	defer func() { _ = reader.Close() }()
	data, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, payload, data)
}
