package task

import (
	"context"
	"time"

	"github.com/flyteorg/flytestdlib/logger"
	"k8s.io/apimachinery/pkg/util/cache"

	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/config"
)

type BarrierKey = string

type PluginCallLog struct {
	PluginTransition *pluginRequestedTransition
}

type BarrierTransition struct {
	BarrierClockTick uint32
	CallLog          PluginCallLog
}

var NoBarrierTransition = BarrierTransition{BarrierClockTick: 0}

type barrier struct {
	barrierCacheExpiration time.Duration
	barrierTransitions     *cache.LRUExpireCache
	barrierEnabled         bool
}

func (b *barrier) RecordBarrierTransition(ctx context.Context, k BarrierKey, bt BarrierTransition) {
	if b.barrierEnabled {
		b.barrierTransitions.Add(k, bt, b.barrierCacheExpiration)
	}
}

func (b *barrier) GetPreviousBarrierTransition(ctx context.Context, k BarrierKey) BarrierTransition {
	if b.barrierEnabled {
		if v, ok := b.barrierTransitions.Get(k); ok {
			f, casted := v.(BarrierTransition)
			if !casted {
				logger.Errorf(ctx, "Failed to cast recorded value to BarrierTransition")
				return NoBarrierTransition
			}
			return f
		}
	}
	return NoBarrierTransition
}

func newLRUBarrier(_ context.Context, cfg config.BarrierConfig) *barrier {
	b := &barrier{
		barrierEnabled: cfg.Enabled,
	}
	if cfg.Enabled {
		b.barrierCacheExpiration = cfg.CacheTTL.Duration
		b.barrierTransitions = cache.NewLRUExpireCache(cfg.CacheSize)
	}
	return b
}
