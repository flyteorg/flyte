// Simple implementation of a KubeClient that caches reads and falls back
// to make direct API calls on failure. Write calls are not cached.
package k8s

import (
	"context"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/rest"

	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

type kubeClient struct {
	client client.Client
	cache  cache.Cache
}

func (k *kubeClient) GetClient() client.Client {
	return k.client
}

func (k *kubeClient) GetCache() cache.Cache {
	return k.cache
}

func newKubeClient(c client.Client, cache cache.Cache) core.KubeClient {
	return &kubeClient{client: c, cache: cache}
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

// ClientBuilder builder is the interface for the client builder.
type ClientBuilder interface {
	// WithUncached takes a list of runtime objects (plain or lists) that users don't want to cache
	// for this client. This function can be called multiple times, it should append to an internal slice.
	WithUncached(objs ...client.Object) ClientBuilder

	// Build returns a new client.
	Build(cache cache.Cache, config *rest.Config, options client.Options) (client.Client, error)
}

type fallbackClientBuilder struct {
	uncached []client.Object
}

func (f *fallbackClientBuilder) WithUncached(objs ...client.Object) ClientBuilder {
	f.uncached = append(f.uncached, objs...)
	return f
}

func (f fallbackClientBuilder) Build(cache cache.Cache, config *rest.Config, options client.Options) (client.Client, error) {
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
func NewFallbackClientBuilder() ClientBuilder {
	return &fallbackClientBuilder{}
}

type Options struct {
	MapperProvider func(*rest.Config) (meta.RESTMapper, error)
	CacheOptions   *cache.Options
	ClientOptions  *client.Options
}

// NewKubeClient creates a new KubeClient that caches reads and falls back to
// make API calls on failure. Write calls are not cached.
func NewKubeClient(config *rest.Config, options Options) (core.KubeClient, error) {
	if options.MapperProvider == nil {
		options.MapperProvider = func(c *rest.Config) (meta.RESTMapper, error) {
			return apiutil.NewDynamicRESTMapper(config)
		}
	}
	mapper, err := options.MapperProvider(config)
	if err != nil {
		return nil, err
	}

	if options.CacheOptions == nil {
		options.CacheOptions = &cache.Options{Mapper: mapper}
	}
	cache, err := cache.New(config, *options.CacheOptions)
	if err != nil {
		return nil, err
	}

	if options.ClientOptions == nil {
		options.ClientOptions = &client.Options{Mapper: mapper}
	}

	fallbackClient, err := NewFallbackClientBuilder().Build(cache, config, *options.ClientOptions)
	if err != nil {
		return nil, err
	}

	return newKubeClient(fallbackClient, cache), nil
}

// NewDefaultKubeClient creates a new KubeClient with default options set.
// This client caches reads and falls back to make API calls on failure. Write calls are not cached.
func NewDefaultKubeClient(config *rest.Config) (core.KubeClient, error) {
	return NewKubeClient(config, Options{})
}
