package webapi

import (
	"context"
	"encoding/gob"
	"fmt"
	"time"

	"k8s.io/utils/clock"

	stdErrs "github.com/flyteorg/flytestdlib/errors"

	"github.com/flyteorg/flytestdlib/cache"

	"github.com/flyteorg/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/webapi"
)

const (
	pluginStateVersion = 1
	minCacheSize       = 10
	maxCacheSize       = 500000
	minWorkers         = 1
	maxWorkers         = 100
	minSyncDuration    = 5 * time.Second
	maxSyncDuration    = time.Hour
	minBurst           = 5
	maxBurst           = 10000
	minQPS             = 1
	maxQPS             = 100000
)

type CorePlugin struct {
	id             string
	p              webapi.AsyncPlugin
	cache          cache.AutoRefresh
	tokenAllocator tokenAllocator
	metrics        Metrics
}

func (c CorePlugin) unmarshalState(ctx context.Context, stateReader core.PluginStateReader) (State, error) {
	t := c.metrics.SucceededUnmarshalState.Start(ctx)
	existingState := State{}

	// We assume here that the first time this function is called, the custom state we get back is whatever we passed in,
	// namely the zero-value of our struct.
	if _, err := stateReader.Get(&existingState); err != nil {
		c.metrics.FailedUnmarshalState.Inc(ctx)
		logger.Errorf(ctx, "AsyncPlugin [%v] failed to unmarshal custom state. Error: %v",
			c.GetID(), err)

		return State{}, errors.Wrapf(errors.CorruptedPluginState, err,
			"Failed to unmarshal custom state in Handle")
	}

	t.Stop()
	return existingState, nil
}

func (c CorePlugin) GetID() string {
	return c.id
}

func (c CorePlugin) GetProperties() core.PluginProperties {
	return core.PluginProperties{}
}

func (c CorePlugin) Handle(ctx context.Context, tCtx core.TaskExecutionContext) (core.Transition, error) {
	incomingState, err := c.unmarshalState(ctx, tCtx.PluginStateReader())
	if err != nil {
		return core.UnknownTransition, err
	}

	var nextState *State
	var phaseInfo core.PhaseInfo
	switch incomingState.Phase {
	case PhaseNotStarted:
		if len(c.p.GetConfig().ResourceQuotas) > 0 {
			nextState, phaseInfo, err = c.tokenAllocator.allocateToken(ctx, c.p, tCtx, &incomingState, c.metrics)
		} else {
			nextState, phaseInfo, err = launch(ctx, c.p, tCtx, c.cache, &incomingState)
		}
	case PhaseAllocationTokenAcquired:
		nextState, phaseInfo, err = launch(ctx, c.p, tCtx, c.cache, &incomingState)
	case PhaseResourcesCreated:
		nextState, phaseInfo, err = monitor(ctx, tCtx, c.p, c.cache, &incomingState)
	}

	if err != nil {
		return core.UnknownTransition, err
	}

	if err := tCtx.PluginStateWriter().Put(pluginStateVersion, nextState); err != nil {
		return core.UnknownTransition, err
	}

	return core.DoTransition(phaseInfo), nil
}

func (c CorePlugin) Abort(ctx context.Context, tCtx core.TaskExecutionContext) error {
	incomingState, err := c.unmarshalState(ctx, tCtx.PluginStateReader())
	if err != nil {
		return err
	}

	logger.Infof(ctx, "Attempting to abort resource [%v].", tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetID())

	err = c.p.Delete(ctx, newPluginContext(incomingState.ResourceMeta, nil, "Aborted", tCtx))
	if err != nil {
		logger.Errorf(ctx, "Failed to abort some resources [%v]. Error: %v",
			tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), err)
		return err
	}

	return nil
}

func (c CorePlugin) Finalize(ctx context.Context, tCtx core.TaskExecutionContext) error {
	if len(c.p.GetConfig().ResourceQuotas) == 0 {
		// If there are no defined quotas, there is nothing to cleanup.
		return nil
	}

	logger.Infof(ctx, "Attempting to finalize resource [%v].",
		tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName())
	return c.tokenAllocator.releaseToken(ctx, c.p, tCtx, c.metrics)
}

func validateRangeInt(fieldName string, min, max, provided int) error {
	if provided > max || provided < min {
		return fmt.Errorf("%v is expected to be between %v and %v. Provided value is %v",
			fieldName, min, max, provided)
	}

	return nil
}

func validateRangeFloat64(fieldName string, min, max, provided float64) error {
	if provided > max || provided < min {
		return fmt.Errorf("%v is expected to be between %v and %v. Provided value is %v",
			fieldName, min, max, provided)
	}

	return nil
}
func validateConfig(cfg webapi.PluginConfig) error {
	errs := stdErrs.ErrorCollection{}
	errs.Append(validateRangeInt("cache size", minCacheSize, maxCacheSize, cfg.Caching.Size))
	errs.Append(validateRangeInt("workers count", minWorkers, maxWorkers, cfg.Caching.Workers))
	errs.Append(validateRangeFloat64("resync interval", minSyncDuration.Seconds(), maxSyncDuration.Seconds(), cfg.Caching.ResyncInterval.Seconds()))
	errs.Append(validateRangeInt("read burst", minBurst, maxBurst, cfg.ReadRateLimiter.Burst))
	errs.Append(validateRangeInt("read qps", minQPS, maxQPS, cfg.ReadRateLimiter.QPS))
	errs.Append(validateRangeInt("write burst", minBurst, maxBurst, cfg.WriteRateLimiter.Burst))
	errs.Append(validateRangeInt("write qps", minQPS, maxQPS, cfg.WriteRateLimiter.QPS))

	return errs.ErrorOrDefault()
}

func createRemotePlugin(pluginEntry webapi.PluginEntry, c clock.Clock) core.PluginEntry {
	return core.PluginEntry{
		ID:                  pluginEntry.ID,
		RegisteredTaskTypes: pluginEntry.SupportedTaskTypes,
		LoadPlugin: func(ctx context.Context, iCtx core.SetupContext) (
			core.Plugin, error) {
			p, err := pluginEntry.PluginLoader(ctx, iCtx)
			if err != nil {
				return nil, err
			}

			err = validateConfig(p.GetConfig())
			if err != nil {
				return nil, fmt.Errorf("config validation failed. Error: %w", err)
			}

			// If the plugin will use a custom state, register it to be able to
			// serialize/deserialize interfaces later.
			if customState := p.GetConfig().ResourceMeta; customState != nil {
				gob.Register(customState)
			}

			if quotas := p.GetConfig().ResourceQuotas; len(quotas) > 0 {
				for ns, quota := range quotas {
					err := iCtx.ResourceRegistrar().RegisterResourceQuota(ctx, ns, quota)
					if err != nil {
						return nil, err
					}
				}
			}

			resourceCache, err := NewResourceCache(ctx, pluginEntry.ID, p, p.GetConfig().Caching,
				iCtx.MetricsScope().NewSubScope("cache"))

			if err != nil {
				return nil, err
			}

			err = resourceCache.Start(ctx)
			if err != nil {
				return nil, err
			}

			return CorePlugin{
				id:             pluginEntry.ID,
				p:              p,
				cache:          resourceCache,
				metrics:        newMetrics(iCtx.MetricsScope()),
				tokenAllocator: newTokenAllocator(c),
			}, nil
		},
	}
}

func CreateRemotePlugin(pluginEntry webapi.PluginEntry) core.PluginEntry {
	return createRemotePlugin(pluginEntry, clock.RealClock{})
}
