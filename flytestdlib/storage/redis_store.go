package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strconv"
	"sync"

	goerrors "errors"

	"github.com/redis/go-redis/v9"

	"github.com/flyteorg/flyte/v2/flytestdlib/errors"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

// redisMaxValueBytes is Redis's hard cap on a string value (512 MiB).
const redisMaxValueBytes = 512 * MiB

// listScanBatchSize is the COUNT hint per SCAN round trip when collecting a listing.
const listScanBatchSize = int64(1000)

// RedisStore implements RawStore on top of a Redis server. Each object is a single Redis string
// value whose key is the path portion of the DataReference, verbatim: redis://<addr>/<key>.
// Directories are emulated as key prefixes (a SCAN match on "<prefix>/*"). It is intended for
// metadata-sized objects (inputs/outputs protobufs); Redis caps a string value at 512MiB and keeps
// it in memory, so large raw data should stay in a blob store.
//
// The configured redis.addr is authoritative for which server is contacted; the host portion of a
// reference is advisory. The same server is commonly reachable under different addresses from
// different network vantage points (in-cluster service DNS vs. host port mapping), so references
// written elsewhere may legitimately carry a host that differs from this process's config — a
// mismatch is logged once for observability rather than rejected.
type RedisStore struct {
	copyImpl
	client       redis.UniversalClient
	baseRef      DataReference
	addr         string
	warnMismatch sync.Once
}

// key extracts the Redis key from reference. An empty key is only valid for List (the container
// root); every other operation needs a concrete key.
func (s *RedisStore) key(ctx context.Context, reference DataReference, allowEmpty bool) (string, error) {
	scheme, container, key, err := reference.Split()
	if err != nil {
		return "", err
	}

	if scheme != TypeRedis {
		return "", fmt.Errorf("reference [%v] is not a redis reference", reference)
	}

	if len(key) == 0 && !allowEmpty {
		return "", fmt.Errorf("reference [%v] has an empty key", reference)
	}

	if container != "" && container != s.addr {
		s.warnMismatch.Do(func() {
			logger.Warnf(ctx,
				"redis reference host [%v] differs from configured redis.addr [%v]; the configured "+
					"address is authoritative (further mismatches will not be logged)", container, s.addr)
		})
	}

	return key, nil
}

func (s *RedisStore) GetBaseContainerFQN(ctx context.Context) DataReference {
	return s.baseRef
}

// CreateSignedURL is not supported; Redis has no notion of pre-signed access.
func (s *RedisStore) CreateSignedURL(ctx context.Context, reference DataReference, properties SignedURLProperties) (
	SignedURLResponse, error) {
	return SignedURLResponse{}, fmt.Errorf("signed URLs are unsupported by redis storage")
}

func (s *RedisStore) Head(ctx context.Context, reference DataReference) (Metadata, error) {
	key, err := s.key(ctx, reference, false)
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

// List returns up to maxItems keys under the reference prefix (or the whole container when the
// reference has no key), paginated by an index cursor.
//
// It deliberately collects the full match set per call instead of exposing the SCAN cursor: SCAN's
// COUNT is only a hint (a batch may exceed maxItems) and its cursor cannot be rewound, so per-batch
// pagination either over-returns or silently drops the keys trimmed off a batch. A MATCH scan walks
// the whole keyspace server-side anyway, so full collection costs the same order as one page, and
// the keyspaces this store is intended for (run metadata) are small. Sorting makes pages
// deterministic across calls.
func (s *RedisStore) List(ctx context.Context, reference DataReference, maxItems int, cursor Cursor) (
	[]DataReference, Cursor, error) {
	key, err := s.key(ctx, reference, true)
	if err != nil {
		return nil, NewCursorAtEnd(), err
	}

	match := "*"
	if key != "" {
		match = key + "/*"
	}

	var offset int
	switch cursor.cursorState {
	case AtStartCursorState:
		offset = 0
	case AtEndCursorState:
		return nil, NewCursorAtEnd(), fmt.Errorf("cursor cannot be at end for the List call")
	default:
		offset, err = strconv.Atoi(cursor.customPosition)
		if err != nil || offset < 0 {
			return nil, NewCursorAtEnd(), fmt.Errorf("invalid redis list cursor [%v]: %w", cursor.customPosition, err)
		}
	}

	var keys []string
	var scanCursor uint64
	for {
		batch, next, err := s.client.Scan(ctx, scanCursor, match, listScanBatchSize).Result()
		if err != nil {
			return nil, NewCursorAtEnd(), err
		}

		keys = append(keys, batch...)
		if next == 0 {
			break
		}

		scanCursor = next
	}

	// SCAN may return duplicates across batches; sort + compact yields a stable, unique listing.
	slices.Sort(keys)
	keys = slices.Compact(keys)

	offset = min(offset, len(keys))
	end := len(keys)
	if maxItems > 0 {
		end = min(offset+maxItems, len(keys))
	}

	_, container, _, err := s.baseRef.Split()
	if err != nil {
		return nil, NewCursorAtEnd(), err
	}

	results := make([]DataReference, 0, end-offset)
	for _, k := range keys[offset:end] {
		results = append(results, NewDataReference(TypeRedis, container, k))
	}

	if end >= len(keys) {
		return results, NewCursorAtEnd(), nil
	}

	return results, NewCursorFromCustomPosition(strconv.Itoa(end)), nil
}

func (s *RedisStore) ReadRaw(ctx context.Context, reference DataReference) (io.ReadCloser, error) {
	key, err := s.key(ctx, reference, false)
	if err != nil {
		return nil, err
	}

	// Enforce the download limit before transferring the value; checking after GET would defeat
	// the limit's purpose of bounding the allocation. STRLEN on a missing key returns 0, so absent
	// keys fall through to GET's redis.Nil handling.
	if maxMB := GetConfig().Limits.GetLimitMegabytes; maxMB != 0 {
		size, err := s.client.StrLen(ctx, key).Result()
		if err != nil {
			return nil, err
		}

		if size > maxMB*MiB {
			return nil, errors.Errorf(ErrExceedsLimit,
				"limit exceeded. %.6fmb > %vmb. You can increase the limit by setting maxDownloadMBs.",
				float64(size)/float64(MiB), maxMB)
		}
	}

	data, err := s.client.Get(ctx, key).Bytes()
	if goerrors.Is(err, redis.Nil) {
		return nil, os.ErrNotExist
	}

	if err != nil {
		return nil, err
	}

	return io.NopCloser(bytes.NewReader(data)), nil
}

func (s *RedisStore) WriteRaw(ctx context.Context, reference DataReference, size int64, opts Options, raw io.Reader,
) error {
	key, err := s.key(ctx, reference, false)
	if err != nil {
		return err
	}

	if size > redisMaxValueBytes {
		return fmt.Errorf("declared size %d exceeds the redis string value cap of %d bytes", size, redisMaxValueBytes)
	}

	// The value must be buffered to SET it, so bound the read: by the declared size when one is
	// given (callers pass the accurate length, as the stow store requires), else by the Redis value
	// cap. The +1 byte detects a stream that overruns its bound instead of silently truncating it.
	limit := redisMaxValueBytes
	if size >= 0 {
		limit = size
	}

	rawBytes, err := io.ReadAll(io.LimitReader(raw, limit+1))
	if err != nil {
		return err
	}

	if int64(len(rawBytes)) > limit {
		return fmt.Errorf("stream for reference [%v] exceeds its bound of %d bytes", reference, limit)
	}

	return s.client.Set(ctx, key, rawBytes, 0).Err()
}

func (s *RedisStore) Delete(ctx context.Context, reference DataReference) error {
	key, err := s.key(ctx, reference, false)
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

// NewRedisRawStore creates a RawStore backed by the Redis server configured in cfg.Redis. It is the
// primary-scheme builder (type: redis) and requires redis.addr to be set.
func NewRedisRawStore(_ context.Context, cfg *Config, metrics *dataStoreMetrics) (RawStore, error) {
	if len(cfg.Redis.Addr) == 0 {
		return nil, fmt.Errorf("storage type [%v] requires redis.addr to be set", TypeRedis)
	}

	return buildRedisStore(cfg.Redis, metrics), nil
}

// buildRedisStore constructs a RedisStore from a resolved RedisConfig.
func buildRedisStore(cfg RedisConfig, metrics *dataStoreMetrics) *RedisStore {
	self := &RedisStore{
		client: redis.NewClient(&redis.Options{
			Addr:     cfg.Addr,
			Username: cfg.Username,
			Password: cfg.Password,
			DB:       cfg.DB,
		}),
		baseRef: DataReference(fmt.Sprintf("%s://%s", TypeRedis, cfg.Addr)),
		addr:    cfg.Addr,
	}

	self.copyImpl = newCopyImpl(self, metrics.copyMetrics)
	return self
}

// redisFactory lazily builds a redis-backed RawStore for a secondary redis:// scheme. The address is
// resolved by precedence: an explicit Schemes["redis"].Redis override, then the top-level Redis
// config, and finally the host portion of the triggering reference. The last step is what lets a
// bare redis:// reference "just work" — the DataStore instantiates a client pointed at the reference
// host on demand. It satisfies backendFactory.
func redisFactory(_ context.Context, _ string, ref DataReference, cfg *Config, _ *http.Client, metrics *dataStoreMetrics) (RawStore, error) {
	var redisCfg RedisConfig
	switch {
	case cfg.Schemes[TypeRedis].Redis != nil:
		redisCfg = *cfg.Schemes[TypeRedis].Redis
	case len(cfg.Redis.Addr) > 0:
		redisCfg = cfg.Redis
	}

	if redisCfg.Addr == "" {
		_, host, _, err := ref.Split()
		if err != nil {
			return nil, err
		}
		if host == "" {
			return nil, fmt.Errorf("cannot resolve a redis address for reference [%v]: set redis.addr or include a host in the reference", ref)
		}
		redisCfg.Addr = host
	}

	return buildRedisStore(redisCfg, metrics), nil
}

// redisAddrConfigured reports whether a redis address is configured, meaning redisFactory will NOT
// fall back to the reference host. It mirrors the address-resolution precedence in redisFactory. The
// routing store uses this to decide whether redis backends can be memoized per scheme (configured
// addr: one authoritative server) or must be memoized per host (no addr: each reference host is its
// own server).
func redisAddrConfigured(cfg *Config) bool {
	if rc := cfg.Schemes[TypeRedis].Redis; rc != nil {
		return rc.Addr != ""
	}
	return cfg.Redis.Addr != ""
}
