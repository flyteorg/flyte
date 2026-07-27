package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"
	"testing/iotest"

	goerrors "errors"

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

	// The limit lives in global config; restore it via t.Cleanup so a failed assertion can't leak
	// the override into other tests. Tests in this package must not use t.Parallel for this reason.
	prevLimit := GetConfig().Limits.GetLimitMegabytes
	GetConfig().Limits.GetLimitMegabytes = 1
	t.Cleanup(func() { GetConfig().Limits.GetLimitMegabytes = prevLimit })

	_, err := store.ReadRaw(context.TODO(), ref)
	require.Error(t, err)
	assert.True(t, IsExceedsLimit(err))
}

func TestRedisStore_WriteRawStreamExceedsDeclaredSize(t *testing.T) {
	store, mr := newTestRedisStore(t)
	ref := redisRef(mr, "mismatched")

	err := store.WriteRaw(context.TODO(), ref, 3, Options{}, bytes.NewReader([]byte("longer-than-declared")))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds its bound")

	// The failed write must not leave a (truncated) value behind.
	md, err := store.Head(context.TODO(), ref)
	require.NoError(t, err)
	assert.False(t, md.Exists())
}

func TestRedisStore_WriteRawDeclaredSizeOverRedisCap(t *testing.T) {
	store, mr := newTestRedisStore(t)

	// Rejected up front from the declared size alone; the stream is never read.
	err := store.WriteRaw(context.TODO(), redisRef(mr, "huge"), redisMaxValueBytes+1, Options{},
		iotest.ErrReader(goerrors.New("stream must not be read")))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "redis string value cap")
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

	var pages [][]DataReference
	seen := map[DataReference]bool{}
	cursor := NewCursorAtStart()
	for {
		items, nextCursor, err := store.List(context.TODO(), redisRef(mr, "runs/r1"), 2, cursor)
		require.NoError(t, err)
		require.LessOrEqual(t, len(items), 2, "List must never return more than maxItems")
		pages = append(pages, items)
		for _, item := range items {
			require.False(t, seen[item], "duplicate item %v across pages", item)
			seen[item] = true
		}
		if IsCursorEnd(nextCursor) {
			break
		}
		cursor = nextCursor
	}

	assert.Equal(t, [][]DataReference{
		{redisRef(mr, "runs/r1/a.pb"), redisRef(mr, "runs/r1/b.pb")},
		{redisRef(mr, "runs/r1/sub/c.pb")},
	}, pages, "pages must be sorted, exactly maxItems-sized, and lossless")
}

func TestRedisStore_ListContainerRoot(t *testing.T) {
	store, mr := newTestRedisStore(t)
	keys := []string{"runs/r1/a.pb", "runs/r2/b.pb", "loose"}
	for _, k := range keys {
		require.NoError(t, store.WriteRaw(context.TODO(), redisRef(mr, k), 1, Options{}, bytes.NewReader([]byte("x"))))
	}

	// A reference with no key (the container root) lists everything.
	items, cursor, err := store.List(context.TODO(), DataReference("redis://"+mr.Addr()), 10, NewCursorAtStart())
	require.NoError(t, err)
	assert.True(t, IsCursorEnd(cursor))
	assert.Equal(t, []DataReference{
		redisRef(mr, "loose"), redisRef(mr, "runs/r1/a.pb"), redisRef(mr, "runs/r2/b.pb"),
	}, items)
}

func TestRedisStore_ReferenceHostIsAdvisory(t *testing.T) {
	// The same Redis is commonly reachable under different addresses from different vantage points
	// (in-cluster DNS vs host port mapping), so references whose host differs from the configured
	// redis.addr still resolve against the configured server.
	store, mr := newTestRedisStore(t)
	require.NoError(t, store.WriteRaw(context.TODO(), redisRef(mr, "vantage"), 1, Options{}, bytes.NewReader([]byte("x"))))

	reader, err := store.ReadRaw(context.TODO(), DataReference("redis://elsewhere.example:6379/vantage"))
	require.NoError(t, err)
	defer func() { _ = reader.Close() }()
	data, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, []byte("x"), data)
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
