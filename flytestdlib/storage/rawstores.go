package storage

import (
	"context"
	"fmt"
	"net/http"

	"github.com/flyteorg/flytestdlib/promutils"
)

type dataStoreCreateFn func(ctx context.Context, cfg *Config, metrics *dataStoreMetrics) (RawStore, error)

var stores = map[string]dataStoreCreateFn{
	TypeMemory: NewInMemoryRawStore,
	TypeLocal:  newStowRawStore,
	TypeMinio:  newStowRawStore,
	TypeS3:     newStowRawStore,
	TypeStow:   newStowRawStore,
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
	c := &http.Client{
		Timeout: cfg.Timeout.Duration,
	}

	if len(cfg.Headers) > 0 {
		c.Transport = &proxyTransport{
			RoundTripper:   http.DefaultTransport,
			defaultHeaders: cfg.Headers,
		}
	}

	return c
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
func (ds *DataStore) RefreshConfig(ctx context.Context, cfg *Config) error {
	defaultClient := http.DefaultClient
	defer func() {
		http.DefaultClient = defaultClient
	}()

	http.DefaultClient = createHTTPClient(cfg.DefaultHTTPClient)

	fn, found := stores[cfg.Type]
	if !found {
		return fmt.Errorf("type is of an invalid value [%v]", cfg.Type)
	}

	rawStore, err := fn(ctx, cfg, ds.metrics)
	if err != nil {
		return err
	}

	rawStore = newCachedRawStore(cfg, rawStore, ds.metrics.cacheMetrics)
	protoStore := NewDefaultProtobufStoreWithMetrics(rawStore, ds.metrics.protoMetrics)
	newDS := NewCompositeDataStore(NewURLPathConstructor(), protoStore)
	newDS.metrics = ds.metrics
	*ds = *newDS
	return nil
}
