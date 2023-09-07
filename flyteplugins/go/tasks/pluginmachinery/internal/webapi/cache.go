package webapi

import (
	"context"

	"github.com/flyteorg/flytestdlib/promutils"
	"k8s.io/client-go/util/workqueue"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/webapi"

	"github.com/flyteorg/flytestdlib/cache"

	"github.com/flyteorg/flyteplugins/go/tasks/errors"
	stdErrors "github.com/flyteorg/flytestdlib/errors"

	"github.com/flyteorg/flytestdlib/logger"
)

//go:generate mockery -all -case=underscore

const (
	BadReturnCodeError stdErrors.ErrorCode = "RETURNED_UNKNOWN"
)

// Client interface needed for resource cache to fetch latest updates for resources.
type Client interface {
	// Get multiple resources that match all the keys. If the plugin hits any failure, it should stop and return
	// the failure. This batch will not be processed further.
	Get(ctx context.Context, tCtx webapi.GetContext) (latest webapi.Resource, err error)

	// Status checks the status of a given resource and translates it to a Flyte-understandable PhaseInfo. This API
	// should avoid making any network calls and should run very efficiently.
	Status(ctx context.Context, tCtx webapi.StatusContext) (phase core.PhaseInfo, err error)
}

// A generic AutoRefresh cache that uses a client to fetch items' status.
type ResourceCache struct {
	// AutoRefresh
	cache.AutoRefresh
	client Client
	cfg    webapi.CachingConfig
}

// A wrapper for each item in the cache.
type CacheItem struct {
	State

	Resource webapi.Resource
}

// This basically grab an updated status from Client and store it in the cache
// All other handling should be in the synchronous loop.
func (q *ResourceCache) SyncResource(ctx context.Context, batch cache.Batch) (
	updatedBatch []cache.ItemSyncResponse, err error) {

	resp := make([]cache.ItemSyncResponse, 0, len(batch))
	for _, resource := range batch {
		// Cast the item back to the thing we want to work with.
		cacheItem, ok := resource.GetItem().(CacheItem)
		if !ok {
			logger.Errorf(ctx, "Sync loop - Error casting cache object into CacheItem")
			return nil, errors.Errorf(errors.CacheFailed, "Failed to cast [%v]", batch[0].GetID())
		}

		if len(resource.GetID()) == 0 {
			logger.Warnf(ctx, "Sync loop - ResourceKey is blank for [%s] skipping", resource.GetID())
			resp = append(resp, cache.ItemSyncResponse{
				ID:     resource.GetID(),
				Item:   resource.GetItem(),
				Action: cache.Unchanged,
			})

			continue
		}

		logger.Debugf(ctx, "Sync loop - processing resource with cache key [%s]",
			resource.GetID())

		if cacheItem.SyncFailureCount > q.cfg.MaxSystemFailures {
			logger.Infof(ctx, "Sync loop - Item with key [%v] has failed to sync [%v] time(s). More than the allowed [%v] time(s). Marking as failure.",
				cacheItem.SyncFailureCount, q.cfg.MaxSystemFailures)
			cacheItem.State.Phase = PhaseSystemFailure
		}

		if cacheItem.State.Phase.IsTerminal() {
			logger.Debugf(ctx, "Sync loop - resource cache key [%v] in terminal state [%s]",
				resource.GetID())

			resp = append(resp, cache.ItemSyncResponse{
				ID:     resource.GetID(),
				Item:   resource.GetItem(),
				Action: cache.Unchanged,
			})

			continue
		}

		// Get an updated status
		logger.Debugf(ctx, "Querying AsyncPlugin for %s", resource.GetID())
		newResource, err := q.client.Get(ctx, newPluginContext(cacheItem.ResourceMeta, cacheItem.Resource, "", nil))
		if err != nil {
			logger.Infof(ctx, "Error retrieving resource [%s]. Error: %v", resource.GetID(), err)
			cacheItem.SyncFailureCount++

			// Make sure we don't return nil for the first argument, because that deletes it from the cache.
			resp = append(resp, cache.ItemSyncResponse{
				ID:     resource.GetID(),
				Item:   cacheItem,
				Action: cache.Update,
			})

			continue
		}

		cacheItem.Resource = newResource

		resp = append(resp, cache.ItemSyncResponse{
			ID:     resource.GetID(),
			Item:   cacheItem,
			Action: cache.Update,
		})
	}

	return resp, nil
}

// ToPluginPhase translates the more granular task phase into the webapi plugin phase.
func ToPluginPhase(s core.Phase) (Phase, error) {
	switch s {

	case core.PhaseUndefined:
		fallthrough
	case core.PhaseNotReady:
		return PhaseNotStarted, nil
	case core.PhaseInitializing:
		fallthrough
	case core.PhaseWaitingForResources:
		fallthrough
	case core.PhaseQueued:
		fallthrough
	case core.PhaseRunning:
		return PhaseResourcesCreated, nil
	case core.PhaseSuccess:
		return PhaseSucceeded, nil
	case core.PhasePermanentFailure:
		fallthrough
	case core.PhaseRetryableFailure:
		return PhaseUserFailure, nil
	default:
		return PhaseSystemFailure, errors.Errorf(BadReturnCodeError, "default fallthrough case")
	}
}

func NewResourceCache(ctx context.Context, name string, client Client, cfg webapi.CachingConfig,
	scope promutils.Scope) (ResourceCache, error) {

	q := ResourceCache{
		client: client,
		cfg:    cfg,
	}

	autoRefreshCache, err := cache.NewAutoRefreshCache(name, q.SyncResource,
		workqueue.DefaultControllerRateLimiter(), cfg.ResyncInterval.Duration, cfg.Workers, cfg.Size,
		scope.NewSubScope("cache"))

	if err != nil {
		logger.Errorf(ctx, "Could not create AutoRefreshCache. Error: [%s]", err)
		return q, errors.Wrapf(errors.CacheFailed, err, "Error creating AutoRefreshCache")
	}

	q.AutoRefresh = autoRefreshCache
	return q, nil
}
