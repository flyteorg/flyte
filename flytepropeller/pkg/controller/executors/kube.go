package executors

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
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
	orderedClients []client.Client
}

func (c fallbackClientReader) Get(ctx context.Context, key client.ObjectKey, out runtime.Object) (err error) {
	for _, k8sClient := range c.orderedClients {
		if err = k8sClient.Get(ctx, key, out); err == nil {
			return nil
		}
	}

	return
}

func (c fallbackClientReader) List(ctx context.Context, list runtime.Object, opts ...client.ListOption) (err error) {
	for _, k8sClient := range c.orderedClients {
		if err = k8sClient.List(ctx, list, opts...); err == nil {
			return nil
		}
	}

	return
}

// Creates a new k8s client that uses the cached client for reads and falls back to making API
// calls if it failed. Write calls will always go to raw client directly.
func NewFallbackClient(cachedClient, rawClient client.Client) client.Client {
	return client.DelegatingClient{
		Reader: fallbackClientReader{
			orderedClients: []client.Client{cachedClient, rawClient},
		},
		StatusClient: rawClient,
		Writer:       rawClient,
	}
}
