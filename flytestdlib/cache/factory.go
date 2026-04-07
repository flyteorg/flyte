// Package cache provides a cache implementation that wraps eko/gocache and provides a simple interface for caching
// in memory as well as through redis. Ideally we move all cache references/implementations to use this package. This way
// we can provide a single namespaced cache for all services and make it easier to add caching for new features.
// For the in-memory implementation, it uses freecache which is a simple in-memory cache that uses a fixed size memory
// pool. For redis, it uses the redis client from the redis/go-redis package.
// Future enhancements could include adding priority/cost per namespace to allow us to evict less important namespaces'
// cache entries when memory is low. rositto has this feature but it's not clear if it's worth the complexity.

package cache

import (
	"context"
	"fmt"
	"reflect"

	"github.com/coocood/freecache"
	"github.com/eko/gocache/lib/v4/cache"
	promMetrics "github.com/eko/gocache/lib/v4/metrics"
	"github.com/eko/gocache/lib/v4/store"
	freecacheStore "github.com/eko/gocache/store/freecache/v4"
	redisStore "github.com/eko/gocache/store/redis/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/extra/redisprometheus/v9"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
)

type Factory struct {
	cfg             *Config
	inMemoryCache   store.StoreInterface
	redisCacheStore store.StoreInterface
}

type StringCache = cache.CacheInterface[string]
type UInt64Cache = cache.CacheInterface[uint64]

func (f Factory) Type() Type {
	return f.cfg.Type
}

type SecretManager interface {
	Get(ctx context.Context, key string) (string, error)
}

// NewRedisClient initializes a new redis client with the given options.
func NewRedisClient(ctx context.Context, cfg RedisOptions, secretManager SecretManager, scope promutils.Scope) (*redis.Client, error) {
	redisOptions, err := cfg.GetOptions(ctx, secretManager)
	if err != nil {
		return nil, fmt.Errorf("failed to get redis options: %w", err)
	}

	redisClient := redis.NewClient(redisOptions)
	redisClientScope := scope.NewSubScope("redis")
	collector := redisprometheus.NewCollector(redisClientScope.CurrentScope(), "", redisClient)
	prometheus.MustRegister(collector)
	redisClient.AddHook(NewRedisInstrumentationHook(redisClientScope))

	return redisClient, nil
}

// NewFactory initializes a new cache factory that should be used whenever typed caches are created.
func NewFactory(ctx context.Context, c *Config, secretManager SecretManager, scope promutils.Scope) (Factory, error) {
	f := Factory{
		cfg: c,
		inMemoryCache: freecacheStore.NewFreecache(freecache.NewCache(
			int(c.InMemoryFixedSize.Size.Value())),
			store.WithExpiration(c.InMemoryFixedSize.DefaultExpiration.Duration)),
	}

	switch c.Type {
	case TypeInMemoryFixedSize:
	case TypeRedis:
		redisClient, err := NewRedisClient(ctx, c.Redis.Options, secretManager, scope)
		if err != nil {
			return Factory{}, fmt.Errorf("failed to create redis client: %w", err)
		}

		f.redisCacheStore = redisStore.NewRedis(
			redisClient,
			store.WithExpiration(c.Redis.DefaultExpiration.Duration))
	}

	return f, nil
}

// New creates a new cache with the given name and load function.
func New[T any](name string, cacheType Type, f Factory, loadFunc cache.LoadFunction[T], scope promutils.Scope) (cache.CacheInterface[T], error) {
	scope = scope.NewSubScope(name)
	var cacheManager cache.CacheInterface[any]
	inMemoryStoreCache := cache.New[any](f.inMemoryCache)

	switch cacheType {
	case TypeInMemoryFixedSize:
		cacheManager = inMemoryStoreCache
	case TypeRedis:
		redisStoreCache := cache.New[any](f.redisCacheStore)
		cacheManager = cache.NewChain[any](inMemoryStoreCache, redisStoreCache)
	}

	promMetrics := promMetrics.NewPrometheus(name,
		promMetrics.WithRegisterer(prometheus.DefaultRegisterer),
		promMetrics.WithNamespace(scope.CurrentScope()))
	cacheManager = cache.NewMetric[any](promMetrics, cacheManager)

	var marshalerCache *Marshaler
	if reflect.TypeOf(new(T)).Elem().Implements(reflect.TypeOf((*proto.Message)(nil)).Elem()) {
		marshalerCache = NewProtoMarshaler(cacheManager)
	} else {
		marshalerCache = NewMsgPackMarshaler(cacheManager)
	}

	namespacedCache := NewNamespacedCache[T](name, NewTypedMarshaler[T](marshalerCache))
	if loadFunc != nil {
		return cache.NewLoadable[T](loadFunc, namespacedCache), nil
	}

	return namespacedCache, nil
}
