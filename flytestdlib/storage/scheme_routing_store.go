package storage

import (
	"context"
	"io"
)

// schemeRoutingStore is a RawStore that dispatches each operation by the DataReference's scheme:
// redis:// references go to the redis store, everything else to the default store. It is installed
// automatically by RefreshConfig when redis.addr is configured alongside a non-redis storage type,
// so metadata can live in Redis while raw data, signed URLs, and everything else keep using the
// blob store. CopyRaw is implemented via copyImpl pointed at the router itself, which makes
// cross-backend copies (e.g. redis -> s3) work through plain ReadRaw/WriteRaw.
type schemeRoutingStore struct {
	copyImpl
	defaultStore RawStore
	redisStore   RawStore
}

func (s *schemeRoutingStore) route(reference DataReference) RawStore {
	if scheme, _, _, err := reference.Split(); err == nil && scheme == TypeRedis {
		return s.redisStore
	}
	return s.defaultStore
}

// GetBaseContainerFQN returns the default store's base container; the redis base is only reachable
// through explicit redis:// references (e.g. a metadata bucket prefix configured as redis://...).
func (s *schemeRoutingStore) GetBaseContainerFQN(ctx context.Context) DataReference {
	return s.defaultStore.GetBaseContainerFQN(ctx)
}

func (s *schemeRoutingStore) CreateSignedURL(
	ctx context.Context, reference DataReference, properties SignedURLProperties) (SignedURLResponse, error) {
	return s.route(reference).CreateSignedURL(ctx, reference, properties)
}

func (s *schemeRoutingStore) Head(ctx context.Context, reference DataReference) (Metadata, error) {
	return s.route(reference).Head(ctx, reference)
}

func (s *schemeRoutingStore) List(ctx context.Context, reference DataReference, maxItems int, cursor Cursor) (
	[]DataReference, Cursor, error) {
	return s.route(reference).List(ctx, reference, maxItems, cursor)
}

func (s *schemeRoutingStore) ReadRaw(ctx context.Context, reference DataReference) (io.ReadCloser, error) {
	return s.route(reference).ReadRaw(ctx, reference)
}

func (s *schemeRoutingStore) WriteRaw(
	ctx context.Context, reference DataReference, size int64, opts Options, raw io.Reader) error {
	return s.route(reference).WriteRaw(ctx, reference, size, opts, raw)
}

func (s *schemeRoutingStore) Delete(ctx context.Context, reference DataReference) error {
	return s.route(reference).Delete(ctx, reference)
}

func newSchemeRoutingStore(defaultStore, redisStore RawStore, metrics *copyMetrics) RawStore {
	self := &schemeRoutingStore{
		defaultStore: defaultStore,
		redisStore:   redisStore,
	}
	self.copyImpl = newCopyImpl(self, metrics)
	return self
}
