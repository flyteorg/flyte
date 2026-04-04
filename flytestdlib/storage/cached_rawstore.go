package storage

import (
	"bytes"
	"context"
	"io"
	"runtime/debug"
	"time"

	"github.com/coocood/freecache"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/flyteorg/flyte/flytestdlib/errors"
	"github.com/flyteorg/flyte/flytestdlib/ioutils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/otelutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

const neverExpire = 0

// freecacheCollector implements prometheus.Collector to expose freecache internal snapshot metrics.
// These are registered when the in-memory cache is enabled and help identify memory pressure from
// the data cache (one of the main sources of Flytepropeller memory consumption).
type freecacheCollector struct {
	cache          *freecache.Cache
	entryCount     *prometheus.Desc
	evacuateCount  *prometheus.Desc
	overwriteCount *prometheus.Desc
	expiredCount   *prometheus.Desc
}

func (c *freecacheCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.entryCount
	ch <- c.evacuateCount
	ch <- c.overwriteCount
	ch <- c.expiredCount
}

func (c *freecacheCollector) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(c.entryCount, prometheus.GaugeValue, float64(c.cache.EntryCount()))
	ch <- prometheus.MustNewConstMetric(c.evacuateCount, prometheus.CounterValue, float64(c.cache.EvacuateCount()))
	ch <- prometheus.MustNewConstMetric(c.overwriteCount, prometheus.CounterValue, float64(c.cache.OverwriteCount()))
	ch <- prometheus.MustNewConstMetric(c.expiredCount, prometheus.CounterValue, float64(c.cache.ExpiredCount()))
}

type cacheMetrics struct {
	CacheHit        prometheus.Counter
	CacheMiss       prometheus.Counter
	CacheWriteError prometheus.Counter
	FetchLatency    promutils.StopWatch
	CacheReadBytes  prometheus.Counter
	CacheWriteBytes prometheus.Counter
	collector       *freecacheCollector
}

type cachedRawStore struct {
	RawStore
	cache   *freecache.Cache
	metrics *cacheMetrics
}

// Head gets metadata about the reference. This should generally be a lightweight operation.
func (s *cachedRawStore) Head(ctx context.Context, reference DataReference) (Metadata, error) {
	ctx, span := otelutils.NewSpan(ctx, otelutils.BlobstoreClientTracer, "flytestdlib.storage.cachedRawStore/Head")
	defer span.End()

	key := []byte(reference)
	if oRaw, err := s.cache.Get(key); err == nil {
		s.metrics.CacheHit.Inc()
		// Found, Cache hit
		size := int64(len(oRaw))
		// return size in metadata
		return StowMetadata{exists: true, size: size}, nil
	}
	s.metrics.CacheMiss.Inc()
	return s.RawStore.Head(ctx, reference)
}

// ReadRaw retrieves a byte array from the Blob store or an error
func (s *cachedRawStore) ReadRaw(ctx context.Context, reference DataReference) (io.ReadCloser, error) {
	ctx, span := otelutils.NewSpan(ctx, otelutils.BlobstoreClientTracer, "flytestdlib.storage.cachedRawStore/ReadRaw")
	defer span.End()

	key := []byte(reference)
	if oRaw, err := s.cache.Get(key); err == nil {
		// Found, Cache hit
		s.metrics.CacheHit.Inc()
		s.metrics.CacheReadBytes.Add(float64(len(oRaw)))
		return ioutils.NewBytesReadCloser(oRaw), nil
	}
	s.metrics.CacheMiss.Inc()
	reader, err := s.RawStore.ReadRaw(ctx, reference)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = reader.Close()
		if err != nil {
			logger.Warnf(ctx, "Failed to close reader [%v]. Error: %v", reference, err)
		}
	}()

	b, err := ioutils.ReadAll(reader, s.metrics.FetchLatency.Start())
	if err != nil {
		return nil, err
	}

	err = s.cache.Set(key, b, 0)
	if err != nil {
		logger.Debugf(ctx, "Failed to Cache the metadata")
		err = errors.Wrapf(ErrFailedToWriteCache, err, "Failed to Cache the metadata")
	}

	return ioutils.NewBytesReadCloser(b), err
}

// WriteRaw stores a raw byte array.
func (s *cachedRawStore) WriteRaw(ctx context.Context, reference DataReference, size int64, opts Options, raw io.Reader) error {
	ctx, span := otelutils.NewSpan(ctx, otelutils.BlobstoreClientTracer, "flytestdlib.storage.cachedRawStore/WriteRaw")
	defer span.End()

	var buf bytes.Buffer
	teeReader := io.TeeReader(raw, &buf)
	err := s.RawStore.WriteRaw(ctx, reference, size, opts, teeReader)
	if err != nil {
		return err
	}

	err = s.cache.Set([]byte(reference), buf.Bytes(), neverExpire)
	if err != nil {
		s.metrics.CacheWriteError.Inc()
		err = errors.Wrapf(ErrFailedToWriteCache, err, "Failed to Cache the metadata")
		logger.Warn(ctx, err.Error())
	} else {
		s.metrics.CacheWriteBytes.Add(float64(buf.Len()))
	}

	return err
}

// Delete removes the referenced data from the cache as well as underlying store.
func (s *cachedRawStore) Delete(ctx context.Context, reference DataReference) error {
	key := []byte(reference)
	if deleted := s.cache.Del(key); deleted {
		s.metrics.CacheHit.Inc()
	} else {
		s.metrics.CacheMiss.Inc()
	}

	return s.RawStore.Delete(ctx, reference)
}

func newCacheMetrics(scope promutils.Scope) *cacheMetrics {
	return &cacheMetrics{
		FetchLatency:    scope.MustNewStopWatch("remote_fetch", "Total Time to read from remote metastore", time.Millisecond),
		CacheHit:        scope.MustNewCounter("cache_hit", "Number of times metadata was found in cache"),
		CacheMiss:       scope.MustNewCounter("cache_miss", "Number of times metadata was not found in cache and remote fetch was required"),
		CacheWriteError: scope.MustNewCounter("cache_write_err", "Failed to write to cache"),
		CacheReadBytes:  scope.MustNewCounter("cache_read_bytes_total", "Bytes read from in-memory cache"),
		CacheWriteBytes: scope.MustNewCounter("cache_write_bytes_total", "Bytes written to in-memory cache"),
		collector: &freecacheCollector{
			entryCount:     prometheus.NewDesc(scope.NewScopedMetricName("entry_count"), "Current number of entries in the in-memory cache", nil, nil),
			evacuateCount:  prometheus.NewDesc(scope.NewScopedMetricName("evacuate_count_total"), "Number of entries evicted from the in-memory cache due to memory pressure", nil, nil),
			overwriteCount: prometheus.NewDesc(scope.NewScopedMetricName("overwrite_count_total"), "Number of times an entry was overwritten in the in-memory cache", nil, nil),
			expiredCount:   prometheus.NewDesc(scope.NewScopedMetricName("expired_count_total"), "Number of entries expired from the in-memory cache", nil, nil),
		},
	}
}

// Creates a CachedStore if Caching is enabled, otherwise returns a RawStore
func newCachedRawStore(cfg *Config, store RawStore, metrics *cacheMetrics) RawStore {
	if cfg.Cache.MaxSizeMegabytes > 0 {
		if cfg.Cache.TargetGCPercent > 0 {
			debug.SetGCPercent(cfg.Cache.TargetGCPercent)
		}
		fc := freecache.NewCache(cfg.Cache.MaxSizeMegabytes * 1024 * 1024)
		metrics.collector.cache = fc
		if err := prometheus.Register(metrics.collector); err != nil {
			// AlreadyRegisteredError occurs on config refresh; the cache pointer is already updated above.
			if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
				panic("Failed to register freecache collector: " + err.Error())
			}
		}
		return &cachedRawStore{
			RawStore: store,
			cache:    fc,
			metrics:  metrics,
		}
	}
	return store
}
