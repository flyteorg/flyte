// Simple implementation of a KubeClient that caches reads and falls back
// to make direct API calls on failure. Write calls are not cached.
package k8s

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
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

type Options struct {
	MapperProvider func(*rest.Config) (meta.RESTMapper, error)
	CacheOptions   *cache.Options
	ClientOptions  *client.Options
}

// NewKubeClient creates a new KubeClient that caches reads and falls back to
// make API calls on failure. Write calls are not cached.
func NewKubeClient(config *rest.Config, options Options) (core.KubeClient, error) {
	httpClient, err := rest.HTTPClientFor(config)
	if err != nil {
		return nil, err
	}

	if options.MapperProvider == nil {
		options.MapperProvider = func(c *rest.Config) (meta.RESTMapper, error) {
			return apiutil.NewDynamicRESTMapper(config, httpClient)
		}
	}

	mapper, err := options.MapperProvider(config)
	if err != nil {
		return nil, err
	}

	if options.CacheOptions == nil {
		options.CacheOptions = &cache.Options{
			HTTPClient: httpClient,
			Mapper:     mapper,
		}
	}

	cache, err := cache.New(config, *options.CacheOptions)
	if err != nil {
		return nil, err
	}

	if options.ClientOptions == nil {
		options.ClientOptions = &client.Options{
			HTTPClient: httpClient,
			Mapper:     mapper,
		}
	}

	client, err := client.New(config, *options.ClientOptions)
	if err != nil {
		return nil, err
	}

	return newKubeClient(client, cache), nil
}

// NewDefaultKubeClient creates a new KubeClient with default options set.
// This client caches reads and falls back to make API calls on failure. Write calls are not cached.
func NewDefaultKubeClient(config *rest.Config) (core.KubeClient, error) {
	return NewKubeClient(config, Options{})
}
