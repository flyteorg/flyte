package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
)

// backendFactory lazily builds the RawStore that serves a single URL scheme. It receives the scheme,
// the reference that triggered the creation (some backends, e.g. redis, derive connection details
// from the reference host), the full Config, the configured http client, and the shared metrics.
type backendFactory func(ctx context.Context, scheme string, ref DataReference, cfg *Config, httpClient *http.Client, metrics *dataStoreMetrics) (RawStore, error)

// routingStore is a RawStore that dispatches each operation by the DataReference's scheme to a
// per-scheme backend, instantiating that backend lazily on first use and memoizing it for reuse.
// This lets a single DataStore serve any scheme (s3://, gs://, abfs://, redis://, ...) and any
// container within each, instead of being pinned to one configured backend.
//
// The primary scheme (derived from Config.Type) is built eagerly by RefreshConfig — it owns the
// configured InitContainer and GetBaseContainerFQN — and seeded into the live map. Every other
// scheme in the registry is dialed on demand the first time a matching reference appears. A scheme
// not present in the registry falls back to the primary store, preserving the pre-routing behavior
// where a single configured store handled every reference.
//
// CopyRaw is implemented via copyImpl pointed at the router itself, so cross-scheme copies (e.g.
// redis -> s3, s3 -> gs) work through plain ReadRaw/WriteRaw against the appropriate backends.
type routingStore struct {
	copyImpl
	cfg          *Config
	httpClient   *http.Client
	metrics      *dataStoreMetrics
	registry     map[string]backendFactory
	primary      string
	primaryStore RawStore
	live         sync.Map // scheme -> RawStore
	createMu     sync.Mutex
}

// storeFor resolves the backend that should serve reference, creating it lazily if needed.
func (s *routingStore) storeFor(ctx context.Context, reference DataReference) (RawStore, error) {
	scheme, _, _, err := reference.Split()
	if err != nil || scheme == "" {
		scheme = s.primary
	}

	if _, ok := s.registry[scheme]; !ok {
		// Unknown scheme: fall back to the primary store, which is how a single configured store
		// handled every reference before scheme routing existed.
		scheme = s.primary
	}

	return s.getOrCreate(ctx, scheme, reference)
}

func (s *routingStore) getOrCreate(ctx context.Context, scheme string, reference DataReference) (RawStore, error) {
	key := s.cacheKey(scheme, reference)
	if v, ok := s.live.Load(key); ok {
		return v.(RawStore), nil
	}

	// Serialize creation so a backend is dialed at most once even under concurrent first references.
	s.createMu.Lock()
	defer s.createMu.Unlock()
	if v, ok := s.live.Load(key); ok {
		return v.(RawStore), nil
	}

	factory, ok := s.registry[scheme]
	if !ok {
		return nil, fmt.Errorf("unsupported storage scheme [%v]", scheme)
	}

	store, err := factory(ctx, scheme, reference, s.cfg, s.httpClient, s.metrics)
	if err != nil {
		return nil, err
	}

	s.live.Store(key, store)
	return store, nil
}

// cacheKey returns the live-map key for the backend that serves reference under scheme. Backends are
// normally memoized per scheme. Redis is the exception when no address is configured: redisFactory
// then dials the host taken from the reference, so distinct hosts must map to distinct backends —
// keying by scheme alone would make the first host win and silently serve later, different hosts from
// the wrong server. When a redis address IS configured it is authoritative and the reference host is
// advisory (see RedisStore), so a single per-scheme backend is correct; this also keeps the eagerly
// seeded primary (Type: redis always has a configured addr) reachable under its scheme key.
func (s *routingStore) cacheKey(scheme string, reference DataReference) string {
	if scheme == TypeRedis && !redisAddrConfigured(s.cfg) {
		if _, host, _, err := reference.Split(); err == nil && host != "" {
			return scheme + "://" + host
		}
	}
	return scheme
}

// GetBaseContainerFQN returns the primary store's base container; other schemes are reachable only
// through explicit references (they have no configured base).
func (s *routingStore) GetBaseContainerFQN(ctx context.Context) DataReference {
	return s.primaryStore.GetBaseContainerFQN(ctx)
}

func (s *routingStore) CreateSignedURL(
	ctx context.Context, reference DataReference, properties SignedURLProperties) (SignedURLResponse, error) {
	store, err := s.storeFor(ctx, reference)
	if err != nil {
		return SignedURLResponse{}, err
	}
	return store.CreateSignedURL(ctx, reference, properties)
}

func (s *routingStore) Head(ctx context.Context, reference DataReference) (Metadata, error) {
	store, err := s.storeFor(ctx, reference)
	if err != nil {
		return nil, err
	}
	return store.Head(ctx, reference)
}

func (s *routingStore) List(ctx context.Context, reference DataReference, maxItems int, cursor Cursor) (
	[]DataReference, Cursor, error) {
	store, err := s.storeFor(ctx, reference)
	if err != nil {
		return nil, NewCursorAtEnd(), err
	}
	return store.List(ctx, reference, maxItems, cursor)
}

func (s *routingStore) ReadRaw(ctx context.Context, reference DataReference) (io.ReadCloser, error) {
	store, err := s.storeFor(ctx, reference)
	if err != nil {
		return nil, err
	}
	return store.ReadRaw(ctx, reference)
}

func (s *routingStore) WriteRaw(
	ctx context.Context, reference DataReference, size int64, opts Options, raw io.Reader) error {
	store, err := s.storeFor(ctx, reference)
	if err != nil {
		return err
	}
	return store.WriteRaw(ctx, reference, size, opts, raw)
}

func (s *routingStore) Delete(ctx context.Context, reference DataReference) error {
	store, err := s.storeFor(ctx, reference)
	if err != nil {
		return err
	}
	return store.Delete(ctx, reference)
}

// memFactory adapts NewInMemoryRawStore to the backendFactory signature.
func memFactory(ctx context.Context, _ string, _ DataReference, cfg *Config, _ *http.Client, metrics *dataStoreMetrics) (RawStore, error) {
	return NewInMemoryRawStore(ctx, cfg, metrics)
}

// defaultBackendRegistry maps every known scheme to the factory that serves it. Stow-backed schemes
// are taken from schemeToStowKind (so out-of-tree schemes registered via RegisterStowScheme are
// included), plus the in-memory and redis backends.
func defaultBackendRegistry() map[string]backendFactory {
	registry := map[string]backendFactory{
		TypeMemory: memFactory,
		TypeRedis:  redisFactory,
	}

	for scheme := range schemeToStowKind {
		registry[scheme] = stowFactory
	}

	return registry
}

// newRoutingStore builds a routing RawStore seeded with an eagerly-constructed primary store
// registered under primaryScheme. CopyRaw is wired to route through the store itself.
func newRoutingStore(cfg *Config, httpClient *http.Client, primaryScheme string, primaryStore RawStore, metrics *dataStoreMetrics) RawStore {
	self := &routingStore{
		cfg:          cfg,
		httpClient:   httpClient,
		metrics:      metrics,
		registry:     defaultBackendRegistry(),
		primary:      primaryScheme,
		primaryStore: primaryStore,
	}

	self.live.Store(primaryScheme, primaryStore)
	self.copyImpl = newCopyImpl(self, metrics.copyMetrics)
	return self
}
