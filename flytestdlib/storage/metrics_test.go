package storage

import (
	"context"
	"sync"
	"testing"

	"github.com/coocood/freecache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
)

func useStorageTestRegistry(t *testing.T) *prometheus.Registry {
	t.Helper()
	registerer := prometheus.DefaultRegisterer
	gatherer := prometheus.DefaultGatherer
	registry := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = registry
	prometheus.DefaultGatherer = registry
	t.Cleanup(func() {
		prometheus.DefaultRegisterer = registerer
		prometheus.DefaultGatherer = gatherer
	})
	return registry
}

func storageMetricNames(t *testing.T, registry *prometheus.Registry) map[string]struct{} {
	t.Helper()
	families, err := registry.Gather()
	require.NoError(t, err)
	names := make(map[string]struct{}, len(families))
	for _, family := range families {
		names[family.GetName()] = struct{}{}
	}
	return names
}

func TestDataStoreMetricScopes(t *testing.T) {
	t.Run("CanonicalOnlyByDefault", func(t *testing.T) {
		registry := useStorageTestRegistry(t)
		newDataStoreMetrics(promutils.NewScope("storage"), false)
		names := storageMetricNames(t, registry)

		require.Contains(t, names, "storage:cache:cache_hit")
		require.Contains(t, names, "storage:proto:proto_fetch_ms")
		require.Contains(t, names, "storage:stow:bad_key_unlabeled")
		require.Contains(t, names, "storage:copy:overall_unlabeled_ms")
		require.NotContains(t, names, "storage:cache_hit")
		require.NotContains(t, names, "storage:proto_fetch_ms")
		require.NotContains(t, names, "storage:bad_key_unlabeled")
	})

	t.Run("LegacyAliasesExcludeNewMetrics", func(t *testing.T) {
		registry := useStorageTestRegistry(t)
		newDataStoreMetrics(promutils.NewScope("storage"), true)
		names := storageMetricNames(t, registry)

		for _, name := range []string{
			"storage:cache:cache_hit", "storage:cache_hit",
			"storage:proto:proto_fetch_ms", "storage:proto_fetch_ms",
			"storage:stow:bad_key_unlabeled", "storage:bad_key_unlabeled",
		} {
			require.Contains(t, names, name)
		}
		require.Contains(t, names, "storage:cache:read_bytes_total")
		require.Contains(t, names, "storage:cache:write_bytes_total")
		require.Contains(t, names, "storage:proto:read_bytes_total")
		require.Contains(t, names, "storage:proto:written_bytes_total")
		require.NotContains(t, names, "storage:read_bytes_total")
		require.NotContains(t, names, "storage:write_bytes_total")
		require.NotContains(t, names, "storage:written_bytes_total")
	})
}

func TestLegacyMetricConfigurationRequiresNewDataStore(t *testing.T) {
	t.Run("Enable", func(t *testing.T) {
		registry := useStorageTestRegistry(t)
		scope := promutils.NewScope("storage")
		store, err := NewDataStore(&Config{Type: TypeMemory}, scope)
		require.NoError(t, err)

		err = store.RefreshConfig(context.Background(), &Config{Type: TypeMemory, EnableLegacyMetrics: true})
		require.NoError(t, err)
		names := storageMetricNames(t, registry)
		require.Contains(t, names, "storage:cache:cache_hit")
		require.NotContains(t, names, "storage:cache_hit")
	})

	t.Run("Disable", func(t *testing.T) {
		registry := useStorageTestRegistry(t)
		scope := promutils.NewScope("storage")
		store, err := NewDataStore(&Config{Type: TypeMemory, EnableLegacyMetrics: true}, scope)
		require.NoError(t, err)

		err = store.RefreshConfig(context.Background(), &Config{Type: TypeMemory})
		require.NoError(t, err)
		names := storageMetricNames(t, registry)
		require.Contains(t, names, "storage:cache:cache_hit")
		require.Contains(t, names, "storage:cache_hit")
	})
}

func TestFreecacheCollectorLifecycle(t *testing.T) {
	registry := useStorageTestRegistry(t)
	scope := promutils.NewScope("storage:cache")
	cacheMetrics := newCacheMetrics(scope, scope)

	newCachedRawStore(&Config{}, nil, cacheMetrics)
	require.NotContains(t, storageMetricNames(t, registry), "storage:cache:entry_count")

	config := &Config{Cache: CachingConfig{MaxSizeMegabytes: 1}}
	newCachedRawStore(config, nil, cacheMetrics)
	names := storageMetricNames(t, registry)
	for _, name := range []string{
		"storage:cache:entry_count",
		"storage:cache:evacuate_count_total",
		"storage:cache:overwrite_count_total",
		"storage:cache:expired_count_total",
	} {
		require.Contains(t, names, name)
	}

	newCachedRawStore(&Config{}, nil, cacheMetrics)
	require.NotContains(t, storageMetricNames(t, registry), "storage:cache:entry_count")

	newCachedRawStore(config, nil, cacheMetrics)
	require.Contains(t, storageMetricNames(t, registry), "storage:cache:entry_count")
}

func TestFreecacheCollectorConcurrentRefresh(t *testing.T) {
	registry := useStorageTestRegistry(t)
	scope := promutils.NewScope("storage:cache")
	cacheMetrics := newCacheMetrics(scope, scope)
	cache := freecache.NewCache(1024 * 1024)
	cacheMetrics.collector.cache.Store(cache)
	cacheMetrics.collector.registerOnce.Do(func() {
		registry.MustRegister(cacheMetrics.collector)
	})

	errors := make(chan error, 4)
	var waitGroup sync.WaitGroup
	for range 4 {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			for range 100 {
				_, err := registry.Gather()
				if err != nil {
					errors <- err
					return
				}
			}
		}()
	}
	for range 100 {
		cacheMetrics.collector.cache.Store(nil)
		cacheMetrics.collector.cache.Store(cache)
	}
	waitGroup.Wait()
	close(errors)
	for err := range errors {
		require.NoError(t, err)
	}
}
