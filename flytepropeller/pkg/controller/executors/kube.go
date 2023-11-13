package executors

import (
	"context"
	"fmt"

	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flyte/flytestdlib/fastcheck"
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

// ClientBuilder builder is the interface for the client builder.
type ClientBuilder interface {
	// Build returns a new client.
	Build(cache cache.Cache, config *rest.Config, options client.Options) (client.Client, error)
}

type FallbackClientBuilder struct {
	scope promutils.Scope
}

func (f *FallbackClientBuilder) Build(_ cache.Cache, config *rest.Config, options client.Options) (client.Client, error) {
	return client.New(config, options)
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
