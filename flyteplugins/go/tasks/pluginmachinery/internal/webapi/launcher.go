package webapi

import (
	"context"
	"time"

	"github.com/flyteorg/flytestdlib/cache"
	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/webapi"
)

func launch(ctx context.Context, p webapi.AsyncPlugin, tCtx core.TaskExecutionContext, cache cache.AutoRefresh,
	state *State) (newState *State, phaseInfo core.PhaseInfo, err error) {
	rMeta, r, err := p.Create(ctx, tCtx)
	if err != nil {
		logger.Errorf(ctx, "Failed to create resource. Error: %v", err)
		return nil, core.PhaseInfo{}, err
	}

	// If the plugin also returned the created resource, check to see if it's already in a terminal state.
	logger.Infof(ctx, "Created Resource Name [%s] and Meta [%v]", tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), rMeta)
	if r != nil {
		phase, err := p.Status(ctx, newPluginContext(rMeta, r, "", tCtx))
		if err != nil {
			logger.Errorf(ctx, "Failed to check resource status. Error: %v", err)
			return nil, core.PhaseInfo{}, err
		}

		if phase.Phase().IsTerminal() {
			logger.Infof(ctx, "Resource has already terminated ID:[%s], Phase:[%s]",
				tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), phase.Phase())
			return state, phase, nil
		}
	}

	// Store the created resource name, and update our state.
	state.ResourceMeta = rMeta
	state.Phase = PhaseResourcesCreated
	cacheItem := CacheItem{
		State: *state,
	}

	// Also, add to the AutoRefreshCache so we start getting updates through background refresh.
	_, err = cache.GetOrCreate(tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), cacheItem)
	if err != nil {
		logger.Errorf(ctx, "Failed to add item to cache. Error: %v", err)
		return nil, core.PhaseInfo{}, err
	}

	return state, core.PhaseInfoQueued(time.Now(), 2, "launched"), nil
}
