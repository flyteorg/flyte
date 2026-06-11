package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"

	goerrors "errors"

	"github.com/redis/go-redis/v9"

	"github.com/flyteorg/flyte/v2/flytestdlib/errors"
)

// RedisStore implements RawStore on top of a Redis server. Each object is a single Redis string
// value whose key is the path portion of the DataReference, verbatim: redis://<addr>/<key>.
// Directories are emulated as key prefixes (a SCAN match on "<prefix>/*"). It is intended for
// metadata-sized objects (inputs/outputs protobufs); Redis caps a string value at 512MiB and keeps
// it in memory, so large raw data should stay in a blob store.
type RedisStore struct {
	copyImpl
	client  redis.UniversalClient
	baseRef DataReference
}

func (s *RedisStore) key(reference DataReference) (string, error) {
	scheme, _, key, err := reference.Split()
	if err != nil {
		return "", err
	}

	if scheme != TypeRedis {
		return "", fmt.Errorf("reference [%v] is not a redis reference", reference)
	}

	if len(key) == 0 {
		return "", fmt.Errorf("reference [%v] has an empty key", reference)
	}

	return key, nil
}

func (s *RedisStore) GetBaseContainerFQN(ctx context.Context) DataReference {
	return s.baseRef
}

// CreateSignedURL is not supported; Redis has no notion of pre-signed access.
func (s *RedisStore) CreateSignedURL(ctx context.Context, reference DataReference, properties SignedURLProperties) (
	SignedURLResponse, error) {
	return SignedURLResponse{}, fmt.Errorf("signed urls are unsupported by redis storage")
}

func (s *RedisStore) Head(ctx context.Context, reference DataReference) (Metadata, error) {
	key, err := s.key(reference)
	if err != nil {
		return MemoryMetadata{}, err
	}

	pipe := s.client.Pipeline()
	existsCmd := pipe.Exists(ctx, key)
	sizeCmd := pipe.StrLen(ctx, key)
	if _, err := pipe.Exec(ctx); err != nil {
		return MemoryMetadata{}, err
	}

	return MemoryMetadata{
		exists: existsCmd.Val() > 0,
		size:   sizeCmd.Val(),
	}, nil
}

func (s *RedisStore) List(ctx context.Context, reference DataReference, maxItems int, cursor Cursor) (
	[]DataReference, Cursor, error) {
	key, err := s.key(reference)
	if err != nil {
		return nil, NewCursorAtEnd(), err
	}

	var scanCursor uint64
	switch cursor.cursorState {
	case AtStartCursorState:
		scanCursor = 0
	case AtEndCursorState:
		return nil, NewCursorAtEnd(), fmt.Errorf("cursor cannot be at end for the List call")
	default:
		scanCursor, err = strconv.ParseUint(cursor.customPosition, 10, 64)
		if err != nil {
			return nil, NewCursorAtEnd(), fmt.Errorf("invalid redis list cursor [%v]: %w", cursor.customPosition, err)
		}
	}

	keys, nextCursor, err := s.client.Scan(ctx, scanCursor, key+"/*", int64(maxItems)).Result()
	if err != nil {
		return nil, NewCursorAtEnd(), err
	}

	_, container, _, err := s.baseRef.Split()
	if err != nil {
		return nil, NewCursorAtEnd(), err
	}

	results := make([]DataReference, len(keys))
	for i, k := range keys {
		results[i] = NewDataReference(TypeRedis, container, k)
	}

	if nextCursor == 0 {
		return results, NewCursorAtEnd(), nil
	}

	return results, NewCursorFromCustomPosition(strconv.FormatUint(nextCursor, 10)), nil
}

func (s *RedisStore) ReadRaw(ctx context.Context, reference DataReference) (io.ReadCloser, error) {
	key, err := s.key(reference)
	if err != nil {
		return nil, err
	}

	data, err := s.client.Get(ctx, key).Bytes()
	if goerrors.Is(err, redis.Nil) {
		return nil, os.ErrNotExist
	}

	if err != nil {
		return nil, err
	}

	if maxMB := GetConfig().Limits.GetLimitMegabytes; maxMB != 0 && int64(len(data)) > maxMB*MiB {
		return nil, errors.Errorf(ErrExceedsLimit,
			"limit exceeded. %.6fmb > %vmb. You can increase the limit by setting maxDownloadMBs.",
			float64(len(data))/float64(MiB), maxMB)
	}

	return io.NopCloser(bytes.NewReader(data)), nil
}

func (s *RedisStore) WriteRaw(ctx context.Context, reference DataReference, size int64, opts Options, raw io.Reader,
) error {
	key, err := s.key(reference)
	if err != nil {
		return err
	}

	rawBytes, err := io.ReadAll(raw)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, key, rawBytes, 0).Err()
}

func (s *RedisStore) Delete(ctx context.Context, reference DataReference) error {
	key, err := s.key(reference)
	if err != nil {
		return err
	}

	deleted, err := s.client.Del(ctx, key).Result()
	if err != nil {
		return err
	}

	if deleted == 0 {
		return os.ErrNotExist
	}

	return nil
}

// NewRedisRawStore creates a RawStore backed by the Redis server configured in cfg.Redis.
func NewRedisRawStore(_ context.Context, cfg *Config, metrics *dataStoreMetrics) (RawStore, error) {
	if len(cfg.Redis.Addr) == 0 {
		return nil, fmt.Errorf("storage type [%v] requires redis.addr to be set", TypeRedis)
	}

	self := &RedisStore{
		client: redis.NewClient(&redis.Options{
			Addr:     cfg.Redis.Addr,
			Username: cfg.Redis.Username,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		}),
		baseRef: DataReference(fmt.Sprintf("%s://%s", TypeRedis, cfg.Redis.Addr)),
	}

	self.copyImpl = newCopyImpl(self, metrics.copyMetrics)
	return self, nil
}
