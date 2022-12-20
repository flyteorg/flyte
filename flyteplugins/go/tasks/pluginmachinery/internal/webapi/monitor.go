package webapi

import (
	"context"

	"github.com/flyteorg/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flytestdlib/cache"
	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
)

func monitor(ctx context.Context, tCtx core.TaskExecutionContext, p Client, cache cache.AutoRefresh, state *State) (
	newState *State, phaseInfo core.PhaseInfo, err error) {
	newCacheItem := CacheItem{
		State: *state,
	}

	cacheItemID := tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()
	item, err := cache.GetOrCreate(cacheItemID, newCacheItem)
	if err != nil {
		return nil, core.PhaseInfo{}, err
	}

	cacheItem, ok := item.(CacheItem)
	if !ok {
		logger.Errorf(ctx, "Error casting cache object into ExecutionState")
		return nil, core.PhaseInfo{}, errors.Errorf(
			errors.CacheFailed, "Failed to cast [%v]", cacheItem)
	}

	// If the cache has not syncd yet, just return
	if cacheItem.Resource == nil {
		return state, core.PhaseInfoRunning(0, nil), nil
	}

	newPhase, err := p.Status(ctx, newPluginContext(cacheItem.ResourceMeta, cacheItem.Resource, "", tCtx))
	if err != nil {
		return nil, core.PhaseInfoUndefined, err
	}

	newPluginPhase, err := ToPluginPhase(newPhase.Phase())
	if err != nil {
		return nil, core.PhaseInfoUndefined, err
	}

	if cacheItem.Phase != newPluginPhase {
		logger.Infof(ctx, "Moving Phase for from %s to %s", cacheItem.Phase, newPluginPhase)
	}

	cacheItem.Phase = newPluginPhase

	if newPluginPhase.IsTerminal() {
		// Queue item for deletion in the cache.
		err = cache.DeleteDelayed(cacheItemID)
		if err != nil {
			logger.Warnf(ctx, "Failed to queue item for deletion in the cache with Item Id: [%v]. Error: %v",
				cacheItemID, err)
		}
	}

	// If there were updates made to the state, we'll have picked them up automatically. Nothing more to do.
	return &cacheItem.State, newPhase, nil
}
