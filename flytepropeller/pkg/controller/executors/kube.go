package executors

import (
	"context"
	"fmt"

	"github.com/flyteorg/flytestdlib/fastcheck"
	"github.com/flyteorg/flytestdlib/promutils"

	"k8s.io/client-go/rest"

	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Client is a friendlier controller-runtime client that gets passed to executors
type Client interface {
	// GetClient returns a client configured with the Config
	GetClient() client.Client

	// GetCache returns a cache.Cache
	GetCache() cache.Cache
}

// fallbackClientReader reads from the cache first and if not found then reads from the configured reader, which
// directly reads from the API
type fallbackClientReader struct {
	orderedClients []client.Reader
}

func (c fallbackClientReader) Get(ctx context.Context, key client.ObjectKey, out client.Object) (err error) {
	for _, k8sClient := range c.orderedClients {
		if err = k8sClient.Get(ctx, key, out); err == nil {
			return nil
		}
	}

	return
}

func (c fallbackClientReader) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) (err error) {
	for _, k8sClient := range c.orderedClients {
		if err = k8sClient.List(ctx, list, opts...); err == nil {
			return nil
		}
	}

	return
}

// ClientBuilder builder is the interface for the client builder.
type ClientBuilder interface {
	// WithUncached takes a list of runtime objects (plain or lists) that users don't want to cache
	// for this client. This function can be called multiple times, it should append to an internal slice.
	WithUncached(objs ...client.Object) ClientBuilder

	// Build returns a new client.
	Build(cache cache.Cache, config *rest.Config, options client.Options) (client.Client, error)
}

type FallbackClientBuilder struct {
	uncached []client.Object
	scope    promutils.Scope
}

func (f *FallbackClientBuilder) WithUncached(objs ...client.Object) ClientBuilder {
	f.uncached = append(f.uncached, objs...)
	return f
}

func (f FallbackClientBuilder) Build(cache cache.Cache, config *rest.Config, options client.Options) (client.Client, error) {
	c, err := client.New(config, options)
	if err != nil {
		return nil, err
	}

	c, err = newWriteThroughCachingWriter(c, 50000, f.scope)
	if err != nil {
		return nil, err
	}

	return client.NewDelegatingClient(client.NewDelegatingClientInput{
		Client: c,
		CacheReader: fallbackClientReader{
			orderedClients: []client.Reader{cache, c},
		},
		UncachedObjects: f.uncached,
		// TODO figure out if this should be true?
		// CacheUnstructured: true,
	})
}

// NewFallbackClientBuilder Creates a new k8s client that uses the cached client for reads and falls back to making API
// calls if it failed. Write calls will always go to raw client directly.
func NewFallbackClientBuilder(scope promutils.Scope) *FallbackClientBuilder {
	return &FallbackClientBuilder{
		scope: scope,
	}
}

type writeThroughCachingWriter struct {
	client.Client
	filter fastcheck.Filter
}

func IDFromObject(obj client.Object, op string) []byte {
	return []byte(fmt.Sprintf("%s:%s:%s:%s", obj.GetObjectKind().GroupVersionKind().String(), obj.GetNamespace(), obj.GetName(), op))
}

// Create first checks the local cache if the object with id was previously successfully saved, if not then
// saves the object obj in the Kubernetes cluster
func (w writeThroughCachingWriter) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	// "c" represents create
	id := IDFromObject(obj, "c")
	if w.filter.Contains(ctx, id) {
		return nil
	}
	err := w.Client.Create(ctx, obj, opts...)
	if err != nil {
		return err
	}
	w.filter.Add(ctx, id)
	return nil
}

// Delete first checks the local cache if the object with id was previously successfully deleted, if not then
// deletes the given obj from Kubernetes cluster.
func (w writeThroughCachingWriter) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	// "d" represents delete
	id := IDFromObject(obj, "d")
	if w.filter.Contains(ctx, id) {
		return nil
	}
	err := w.Client.Delete(ctx, obj, opts...)
	if err != nil {
		return err
	}
	w.filter.Add(ctx, id)
	return nil
}

func newWriteThroughCachingWriter(c client.Client, cacheSize int, scope promutils.Scope) (writeThroughCachingWriter, error) {
	filter, err := fastcheck.NewOppoBloomFilter(cacheSize, scope.NewSubScope("kube_filter"))
	if err != nil {
		return writeThroughCachingWriter{}, err
	}
	return writeThroughCachingWriter{
		Client: c,
		filter: filter,
	}, nil
}
