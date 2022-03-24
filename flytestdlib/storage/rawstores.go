package storage

import (
	"fmt"
	"net/http"

	"github.com/flyteorg/flytestdlib/promutils"
)

type dataStoreCreateFn func(cfg *Config, metricsScope promutils.Scope) (RawStore, error)

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

// NewDataStore creates a new Data Store with the supplied config.
func NewDataStore(cfg *Config, metricsScope promutils.Scope) (s *DataStore, err error) {
	defaultClient := http.DefaultClient
	defer func() {
		http.DefaultClient = defaultClient
	}()

	http.DefaultClient = createHTTPClient(cfg.DefaultHTTPClient)

	var rawStore RawStore
	if fn, found := stores[cfg.Type]; found {
		rawStore, err = fn(cfg, metricsScope)
		if err != nil {
			return &emptyStore, err
		}

		protoStore := NewDefaultProtobufStore(newCachedRawStore(cfg, rawStore, metricsScope), metricsScope)
		return NewCompositeDataStore(NewURLPathConstructor(), protoStore), nil
	}

	return &emptyStore, fmt.Errorf("type is of an invalid value [%v]", cfg.Type)
}

// NewCompositeDataStore composes a new DataStore.
func NewCompositeDataStore(refConstructor ReferenceConstructor, composedProtobufStore ComposedProtobufStore) *DataStore {
	return &DataStore{
		ReferenceConstructor:  refConstructor,
		ComposedProtobufStore: composedProtobufStore,
	}
}
