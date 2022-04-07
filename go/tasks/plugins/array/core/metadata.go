package core

import (
	"context"

	idlCore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
)

// InitializeExternalResources constructs an ExternalResource array where each element describes the
// initial state of the subtask. This involves labeling all cached subtasks as successful with a
// cache hit and initializing others to undefined state.
func InitializeExternalResources(ctx context.Context, tCtx core.TaskExecutionContext, state *State,
	generateSubTaskID func(core.TaskExecutionContext, int) string) ([]*core.ExternalResource, error) {
	externalResources := make([]*core.ExternalResource, state.GetOriginalArraySize())

	taskTemplate, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		return externalResources, err
	} else if taskTemplate == nil {
		return externalResources, errors.Errorf(errors.BadTaskSpecification, "Required value not set, taskTemplate is nil")
	}

	executeSubTaskCount := 0
	cachedSubTaskCount := 0
	for i := 0; i < int(state.GetOriginalArraySize()); i++ {
		var cacheStatus idlCore.CatalogCacheStatus
		var childIndex int
		var phase core.Phase

		if state.IndexesToCache.IsSet(uint(i)) {
			// if not cached set to PhaseUndefined and set cacheStatus according to Discoverable
			phase = core.PhaseUndefined
			if taskTemplate.Metadata == nil || !taskTemplate.Metadata.Discoverable {
				cacheStatus = idlCore.CatalogCacheStatus_CACHE_DISABLED
			} else {
				cacheStatus = idlCore.CatalogCacheStatus_CACHE_MISS
			}

			childIndex = executeSubTaskCount
			executeSubTaskCount++
		} else {
			// if cached set to PhaseSuccess and mark as CACHE_HIT
			phase = core.PhaseSuccess
			cacheStatus = idlCore.CatalogCacheStatus_CACHE_HIT

			// child index is computed as a pseudo-value to ensure non-overlapping subTaskIDs
			childIndex = state.GetExecutionArraySize() + cachedSubTaskCount
			cachedSubTaskCount++
		}

		subTaskID := generateSubTaskID(tCtx, childIndex)
		externalResources[i] = &core.ExternalResource{
			ExternalID:   subTaskID,
			CacheStatus:  cacheStatus,
			Index:        uint32(i),
			Logs:         nil,
			RetryAttempt: 0,
			Phase:        phase,
		}
	}

	return externalResources, nil
}
