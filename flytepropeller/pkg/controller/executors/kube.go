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

func BuildNewClientFunc(scope promutils.Scope) func(config *rest.Config, options client.Options) (client.Client, error) {
	return func(config *rest.Config, options client.Options) (client.Client, error) {
		var cacheReader client.Reader
		cachelessOptions := options
		if options.Cache != nil && options.Cache.Reader != nil {
			cacheReader = options.Cache.Reader
			cachelessOptions.Cache = nil
		}

		kubeClient, err := client.New(config, cachelessOptions)
		if err != nil {
			return nil, err
		}

		return newFlyteK8sClient(kubeClient, cacheReader, scope)
	}
}

type flyteK8sClient struct {
	client.Client
	cacheReader client.Reader
	writeFilter fastcheck.Filter
}

func (f flyteK8sClient) Get(ctx context.Context, key client.ObjectKey, out client.Object, opts ...client.GetOption) (err error) {
	if f.cacheReader != nil {
		if err = f.cacheReader.Get(ctx, key, out, opts...); err == nil {
			return nil
		}
	}

	return f.Client.Get(ctx, key, out, opts...)
}

func (f flyteK8sClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) (err error) {
	if f.cacheReader != nil {
		if err = f.cacheReader.List(ctx, list, opts...); err == nil {
			return nil
		}
	}

	return f.Client.List(ctx, list, opts...)
}

// Create first checks the local cache if the object with id was previously successfully saved, if not then
// saves the object obj in the Kubernetes cluster
func (f flyteK8sClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	// "c" represents create
	id := idFromObject(obj, "c")
	if f.writeFilter.Contains(ctx, id) {
		return nil
	}
	err := f.Client.Create(ctx, obj, opts...)
	if err != nil {
		return err
	}
	f.writeFilter.Add(ctx, id)
	return nil
}

// Delete first checks the local cache if the object with id was previously successfully deleted, if not then
// deletes the given obj from Kubernetes cluster.
func (f flyteK8sClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	// "d" represents delete
	id := idFromObject(obj, "d")
	if f.writeFilter.Contains(ctx, id) {
		return nil
	}
	err := f.Client.Delete(ctx, obj, opts...)
	if err != nil {
		return err
	}
	f.writeFilter.Add(ctx, id)
	return nil
}

func idFromObject(obj client.Object, op string) []byte {
	return []byte(fmt.Sprintf("%s:%s:%s:%s", obj.GetObjectKind().GroupVersionKind().String(), obj.GetNamespace(), obj.GetName(), op))
}

func newFlyteK8sClient(kubeClient client.Client, cacheReader client.Reader, scope promutils.Scope) (flyteK8sClient, error) {
	writeFilter, err := fastcheck.NewOppoBloomFilter(50000, scope.NewSubScope("kube_filter"))
	if err != nil {
		return flyteK8sClient{}, err
	}

	return flyteK8sClient{
		Client:      kubeClient,
		cacheReader: cacheReader,
		writeFilter: writeFilter,
	}, nil
}
