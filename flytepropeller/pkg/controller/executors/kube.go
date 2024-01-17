package executors

import (
	"context"
	"fmt"

	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flyte/flytestdlib/fastcheck"
	"github.com/flyteorg/flyte/flytestdlib/otelutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

//go:generate mockery -name Client -case=underscore

// Client is a friendlier controller-runtime client that gets passed to executors
type Client interface {
	// GetClient returns a client configured with the Config
	GetClient() client.Client

	// GetCache returns a cache.Cache
	GetCache() cache.Cache
}

var NewCache = func(config *rest.Config, options cache.Options) (cache.Cache, error) {
	k8sCache, err := cache.New(config, options)
	if err != nil {
		return k8sCache, err
	}

	return otelutils.WrapK8sCache(k8sCache), nil
}

var NewClient = func(config *rest.Config, options client.Options) (client.Client, error) {
	var reader *fallbackClientReader
	if options.Cache != nil && options.Cache.Reader != nil {
		// if caching is enabled we create a fallback reader so we can attempt the client if the cache
		// reader does not have the object
		reader = &fallbackClientReader{
			orderedClients: []client.Reader{options.Cache.Reader},
		}

		options.Cache.Reader = reader
	}

	// create the k8s client
	k8sClient, err := client.New(config, options)
	if err != nil {
		return k8sClient, err
	}

	// TODO - should we wrap this in a writeThroughCachingWriter as well?
	k8sOtelClient := otelutils.WrapK8sClient(k8sClient)
	if reader != nil {
		// once the k8s client is created we set the fallback reader's client to the k8s client
		reader.orderedClients = append([]client.Reader{k8sOtelClient}, reader.orderedClients...)
	}

	return k8sOtelClient, nil
}

// fallbackClientReader reads from the cache first and if not found then reads from the configured reader, which
// directly reads from the API
type fallbackClientReader struct {
	orderedClients []client.Reader
}


func (c fallbackClientReader) Get(ctx context.Context, key client.ObjectKey, out client.Object, opts ...client.GetOption) (err error) {
	for _, k8sClient := range c.orderedClients {
		if err = k8sClient.Get(ctx, key, out, opts...); err == nil {
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
