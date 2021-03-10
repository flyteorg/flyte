package executors

import (
	"context"

	"k8s.io/client-go/rest"

	"sigs.k8s.io/controller-runtime/pkg/cluster"

	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// A friendly controller-runtime client that gets passed to executors
type Client interface {
	// GetClient returns a client configured with the Config
	GetClient() client.Client

	// GetCache returns a cache.Cache
	GetCache() cache.Cache
}

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

type FallbackClientBuilder struct {
	uncached []client.Object
}

func (f *FallbackClientBuilder) WithUncached(objs ...client.Object) cluster.ClientBuilder {
	f.uncached = append(f.uncached, objs...)
	return f
}

func (f FallbackClientBuilder) Build(cache cache.Cache, config *rest.Config, options client.Options) (client.Client, error) {
	c, err := client.New(config, options)
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

// Creates a new k8s client that uses the cached client for reads and falls back to making API
// calls if it failed. Write calls will always go to raw client directly.
func NewFallbackClientBuilder() *FallbackClientBuilder {
	return &FallbackClientBuilder{}
}
