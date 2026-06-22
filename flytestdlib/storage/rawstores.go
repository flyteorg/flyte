package storage

import (
	"context"
	"fmt"
	"net/http"

	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
)

type dataStoreCreateFn func(ctx context.Context, cfg *Config, metrics *dataStoreMetrics) (RawStore, error)

var stores = map[string]dataStoreCreateFn{
	TypeMemory: NewInMemoryRawStore,
	TypeLocal:  newStowRawStore,
	TypeMinio:  newStowRawStore,
	TypeS3:     newStowRawStore,
	TypeStow:   newStowRawStore,
	TypeRedis:  NewRedisRawStore,
}

type proxyTransport struct {
	http.RoundTripper
	defaultHeaders map[string][]string
}

func (p proxyTransport) RoundTrip(r *http.Request) (resp *http.Response, err error) {
	applyDefaultHeaders(r, p.defaultHeaders)
	return p.RoundTripper.RoundTrip(r)
}

func applyDefaultHeaders(r *http.Request, headers map[string][]string) {
	if r.Header == nil {
		r.Header = http.Header{}
	}

	for key, values := range headers {
		for _, val := range values {
			r.Header.Add(key, val)
		}
	}
}

func createHTTPClient(cfg HTTPClientConfig) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if cfg.MaxIdleConns > 0 {
		transport.MaxIdleConns = cfg.MaxIdleConns
	}
	if cfg.MaxIdleConnsPerHost > 0 {
		transport.MaxIdleConnsPerHost = cfg.MaxIdleConnsPerHost
	}
	if cfg.MaxConnsPerHost > 0 {
		transport.MaxConnsPerHost = cfg.MaxConnsPerHost
	}
	if cfg.IdleConnTimeout.Duration > 0 {
		transport.IdleConnTimeout = cfg.IdleConnTimeout.Duration
	}

	var rt http.RoundTripper = transport
	if len(cfg.Headers) > 0 {
		rt = &proxyTransport{
			RoundTripper:   transport,
			defaultHeaders: cfg.Headers,
		}
	}

	return &http.Client{
		Timeout:   cfg.Timeout.Duration,
		Transport: rt,
	}
}

type dataStoreMetrics struct {
	cacheMetrics *cacheMetrics
	protoMetrics *protoMetrics
	copyMetrics  *copyMetrics
	stowMetrics  *stowMetrics
}

// newDataStoreMetrics initialises all metrics required for DataStore
func newDataStoreMetrics(scope promutils.Scope) *dataStoreMetrics {
	return &dataStoreMetrics{
		cacheMetrics: newCacheMetrics(scope),
		protoMetrics: newProtoMetrics(scope),
		copyMetrics:  newCopyMetrics(scope.NewSubScope("copy")),
		stowMetrics:  newStowMetrics(scope),
	}
}

// NewDataStore creates a new Data Store with the supplied config.
func NewDataStore(cfg *Config, scope promutils.Scope) (s *DataStore, err error) {
	return NewDataStoreWithContext(context.Background(), cfg, scope)
}

// NewDataStoreWithContext creates a new Data Store with the supplied config and context.
func NewDataStoreWithContext(ctx context.Context, cfg *Config, scope promutils.Scope) (s *DataStore, err error) {
	ds := &DataStore{metrics: newDataStoreMetrics(scope)}
	return ds, ds.RefreshConfig(ctx, cfg)
}

// NewCompositeDataStore composes a new DataStore.
func NewCompositeDataStore(refConstructor ReferenceConstructor, composedProtobufStore ComposedProtobufStore) *DataStore {
	return &DataStore{
		ReferenceConstructor:  refConstructor,
		ComposedProtobufStore: composedProtobufStore,
	}
}

// RefreshConfig re-initialises the data store client leaving metrics untouched.
// This is NOT thread-safe!
//
// The DataStore is multi-scheme: the configured Type is built eagerly as the primary backend (it
// owns the InitContainer and GetBaseContainerFQN), then wrapped in a routing store that lazily
// instantiates a backend for any other scheme (s3://, gs://, abfs://, redis://, ...) the first time
// it is referenced and reuses it thereafter. See routingStore for details.
func (ds *DataStore) RefreshConfig(ctx context.Context, cfg *Config) error {
	httpClient := createHTTPClient(cfg.DefaultHTTPClient)

	fn, found := stores[cfg.Type]
	if !found {
		return fmt.Errorf("type is of an invalid value [%v]", cfg.Type)
	}

	// stow reads http.DefaultClient at dial time. Install the configured client for the eager primary
	// build while holding dialMu — the same lock lazy per-scheme dials take (via dialStow) — so a
	// concurrent lazy dial or another DataStore being constructed cannot interleave on the global
	// client and restore the wrong value. Lazy per-scheme dials thread the client explicitly through
	// the routing store, so the global only needs to be installed for this eager build.
	primaryStore, err := func() (RawStore, error) {
		dialMu.Lock()
		defer dialMu.Unlock()
		defaultClient := http.DefaultClient
		http.DefaultClient = httpClient
		defer func() { http.DefaultClient = defaultClient }()
		return fn(ctx, cfg, ds.metrics)
	}()
	if err != nil {
		return err
	}

	primaryScheme, err := primarySchemeForConfig(cfg)
	if err != nil {
		return err
	}

	rawStore := newRoutingStore(cfg, httpClient, primaryScheme, primaryStore, ds.metrics)

	rawStore = newCachedRawStore(cfg, rawStore, ds.metrics.cacheMetrics)
	protoStore := NewDefaultProtobufStoreWithMetrics(rawStore, ds.metrics.protoMetrics)
	newDS := NewCompositeDataStore(NewURLPathConstructor(), protoStore)
	newDS.metrics = ds.metrics
	*ds = *newDS
	return nil
}
