package presto

import (
	"context"

	"k8s.io/client-go/util/workqueue"

	"github.com/flyteorg/flytestdlib/cache"

	"github.com/flyteorg/flyteplugins/go/tasks/errors"
	stdErrors "github.com/flyteorg/flytestdlib/errors"

	"github.com/flyteorg/flyteplugins/go/tasks/plugins/presto/client"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/presto/config"

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
)

const (
	BadPrestoReturnCodeError stdErrors.ErrorCode = "PRESTO_RETURNED_UNKNOWN"
)

type ExecutionsCache struct {
	cache.AutoRefresh
	prestoClient client.PrestoClient
	scope        promutils.Scope
	cfg          *config.Config
}

func NewPrestoExecutionsCache(
	ctx context.Context,
	prestoClient client.PrestoClient,
	cfg *config.Config,
	scope promutils.Scope) (ExecutionsCache, error) {

	q := ExecutionsCache{
		prestoClient: prestoClient,
		scope:        scope,
		cfg:          cfg,
	}
	autoRefreshCache, err := cache.NewAutoRefreshCache(cfg.RefreshCacheConfig.Name, q.SyncPrestoQuery, workqueue.DefaultControllerRateLimiter(), cfg.RefreshCacheConfig.SyncPeriod.Duration, cfg.RefreshCacheConfig.Workers, cfg.RefreshCacheConfig.LruCacheSize, scope)
	if err != nil {
		logger.Errorf(ctx, "Could not create AutoRefreshCache in Executor. [%s]", err)
		return q, errors.Wrapf(errors.CacheFailed, err, "Error creating AutoRefreshCache")
	}
	q.AutoRefresh = autoRefreshCache
	return q, nil
}

type ExecutionStateCacheItem struct {
	ExecutionState

	// This ID is the cache key and so will need to be unique across all objects in the cache (it will probably be
	// unique across all of Flyte) and needs to be deterministic.
	// This will also be used as the allocation token for now.
	Identifier string `json:"id"`
}

func (e ExecutionStateCacheItem) ID() string {
	return e.Identifier
}

// This basically grab an updated status from the Presto API and stores it in the cache
// All other handling should be in the synchronous loop.
func (p *ExecutionsCache) SyncPrestoQuery(ctx context.Context, batch cache.Batch) (
	updatedBatch []cache.ItemSyncResponse, err error) {

	resp := make([]cache.ItemSyncResponse, 0, len(batch))
	for _, query := range batch {
		// Cast the item back to the thing we want to work with.
		executionStateCacheItem, ok := query.GetItem().(ExecutionStateCacheItem)
		if !ok {
			logger.Errorf(ctx, "Sync loop - Error casting cache object into ExecutionState")
			return nil, errors.Errorf(errors.CacheFailed, "Failed to cast [%v]", batch[0].GetID())
		}

		if executionStateCacheItem.CommandID == "" {
			logger.Warnf(ctx, "Sync loop - CommandID is blank for [%s] skipping", executionStateCacheItem.Identifier)
			resp = append(resp, cache.ItemSyncResponse{
				ID:     query.GetID(),
				Item:   query.GetItem(),
				Action: cache.Unchanged,
			})

			continue
		}

		logger.Debugf(ctx, "Sync loop - processing Presto job [%s] - cache key [%s]",
			executionStateCacheItem.CommandID, executionStateCacheItem.Identifier)

		if InTerminalState(executionStateCacheItem.ExecutionState) {
			logger.Debugf(ctx, "Sync loop - Presto id [%s] in terminal state [%s]",
				executionStateCacheItem.CommandID, executionStateCacheItem.Identifier)

			resp = append(resp, cache.ItemSyncResponse{
				ID:     query.GetID(),
				Item:   query.GetItem(),
				Action: cache.Unchanged,
			})

			continue
		}

		// Get an updated status from Presto
		logger.Debugf(ctx, "Querying Presto for %s - %s", executionStateCacheItem.CommandID, executionStateCacheItem.Identifier)
		commandStatus, err := p.prestoClient.GetCommandStatus(ctx, executionStateCacheItem.CommandID)
		if err != nil {
			logger.Errorf(ctx, "Error from Presto command %s", executionStateCacheItem.CommandID)
			executionStateCacheItem.SyncFailureCount++
			// Make sure we don't return nil for the first argument, because that deletes it from the cache.
			resp = append(resp, cache.ItemSyncResponse{
				ID:     query.GetID(),
				Item:   executionStateCacheItem,
				Action: cache.Update,
			})

			continue
		}

		newExecutionPhase, err := StatusToExecutionPhase(commandStatus)
		if err != nil {
			return nil, err
		}

		if newExecutionPhase > executionStateCacheItem.CurrentPhase {
			logger.Infof(ctx, "Moving ExecutionPhase for %s %s from %s to %s", executionStateCacheItem.CommandID,
				executionStateCacheItem.Identifier, executionStateCacheItem.CurrentPhase, newExecutionPhase)

			executionStateCacheItem.PreviousPhase = executionStateCacheItem.CurrentPhase
			executionStateCacheItem.CurrentPhase = newExecutionPhase

			resp = append(resp, cache.ItemSyncResponse{
				ID:     query.GetID(),
				Item:   executionStateCacheItem,
				Action: cache.Update,
			})
		}
	}

	return resp, nil
}

// We need some way to translate results we get from Presto, into a plugin phase
func StatusToExecutionPhase(s client.PrestoStatus) (ExecutionPhase, error) {
	switch s {
	case client.PrestoStatusFinished:
		return PhaseQuerySucceeded, nil
	case client.PrestoStatusCancelled:
		return PhaseQueryFailed, nil
	case client.PrestoStatusFailed:
		return PhaseQueryFailed, nil
	case client.PrestoStatusWaiting:
		return PhaseSubmitted, nil
	case client.PrestoStatusRunning:
		return PhaseSubmitted, nil
	case client.PrestoStatusUnknown:
		return PhaseQueryFailed, errors.Errorf(BadPrestoReturnCodeError, "Presto returned status Unknown")
	default:
		return PhaseQueryFailed, errors.Errorf(BadPrestoReturnCodeError, "default fallthrough case")
	}
}
