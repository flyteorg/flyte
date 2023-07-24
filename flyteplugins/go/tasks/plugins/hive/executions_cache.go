package hive

import (
	"context"
	"time"

	"k8s.io/client-go/util/workqueue"

	"github.com/flyteorg/flytestdlib/cache"

	"github.com/flyteorg/flyteplugins/go/tasks/errors"
	stdErrors "github.com/flyteorg/flytestdlib/errors"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/hive/client"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/hive/config"

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
)

const ResyncDuration = 30 * time.Second

const (
	BadQuboleReturnCodeError stdErrors.ErrorCode = "QUBOLE_RETURNED_UNKNOWN"
)

type QuboleHiveExecutionsCache struct {
	cache.AutoRefresh
	quboleClient  client.QuboleClient
	secretManager core.SecretManager
	scope         promutils.Scope
	cfg           *config.Config
}

func NewQuboleHiveExecutionsCache(ctx context.Context, quboleClient client.QuboleClient,
	secretManager core.SecretManager, cfg *config.Config, scope promutils.Scope) (QuboleHiveExecutionsCache, error) {

	q := QuboleHiveExecutionsCache{
		quboleClient:  quboleClient,
		secretManager: secretManager,
		scope:         scope,
		cfg:           cfg,
	}
	autoRefreshCache, err := cache.NewAutoRefreshCache("qubole", q.SyncQuboleQuery, workqueue.DefaultControllerRateLimiter(), ResyncDuration, cfg.Workers, cfg.LruCacheSize, scope)
	if err != nil {
		logger.Errorf(ctx, "Could not create AutoRefreshCache in QuboleHiveExecutor. [%s]", err)
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

// This basically grab an updated status from the Qubole API and store it in the cache
// All other handling should be in the synchronous loop.
func (q *QuboleHiveExecutionsCache) SyncQuboleQuery(ctx context.Context, batch cache.Batch) (
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

		logger.Debugf(ctx, "Sync loop - processing Hive job [%s] - cache key [%s]",
			executionStateCacheItem.CommandID, executionStateCacheItem.Identifier)

		quboleAPIKey, err := q.secretManager.Get(ctx, q.cfg.TokenKey)
		if err != nil {
			return nil, err
		}

		if InTerminalState(executionStateCacheItem.ExecutionState) {
			logger.Debugf(ctx, "Sync loop - Qubole id [%s] in terminal state [%s]",
				executionStateCacheItem.CommandID, executionStateCacheItem.Identifier)

			resp = append(resp, cache.ItemSyncResponse{
				ID:     query.GetID(),
				Item:   query.GetItem(),
				Action: cache.Unchanged,
			})

			continue
		}

		// Get an updated status from Qubole
		logger.Debugf(ctx, "Querying Qubole for %s - %s", executionStateCacheItem.CommandID, executionStateCacheItem.Identifier)
		commandStatus, err := q.quboleClient.GetCommandStatus(ctx, executionStateCacheItem.CommandID, quboleAPIKey)
		if err != nil {
			logger.Errorf(ctx, "Error from Qubole command %s", executionStateCacheItem.CommandID)
			executionStateCacheItem.SyncFailureCount++
			// Make sure we don't return nil for the first argument, because that deletes it from the cache.
			resp = append(resp, cache.ItemSyncResponse{
				ID:     query.GetID(),
				Item:   executionStateCacheItem,
				Action: cache.Update,
			})

			continue
		}

		newExecutionPhase, err := QuboleStatusToExecutionPhase(commandStatus)
		if err != nil {
			return nil, err
		}

		if newExecutionPhase > executionStateCacheItem.Phase {
			logger.Infof(ctx, "Moving ExecutionPhase for %s %s from %s to %s", executionStateCacheItem.CommandID,
				executionStateCacheItem.Identifier, executionStateCacheItem.Phase, newExecutionPhase)

			executionStateCacheItem.Phase = newExecutionPhase

			resp = append(resp, cache.ItemSyncResponse{
				ID:     query.GetID(),
				Item:   executionStateCacheItem,
				Action: cache.Update,
			})
		}
	}

	return resp, nil
}

// We need some way to translate results we get from Qubole, into a plugin phase
// NB: This function should only return plugin phases that are greater than (">") phases that represent states before
//
//	the query was kicked off. That is, it will never make sense to go back to PhaseNotStarted, after we've
//	submitted the query to Qubole.
func QuboleStatusToExecutionPhase(s client.QuboleStatus) (ExecutionPhase, error) {
	switch s {
	case client.QuboleStatusDone:
		return PhaseWriteOutputFile, nil
	case client.QuboleStatusCancelled:
		return PhaseQueryFailed, nil
	case client.QuboleStatusError:
		return PhaseQueryFailed, nil
	case client.QuboleStatusWaiting:
		return PhaseSubmitted, nil
	case client.QuboleStatusRunning:
		return PhaseSubmitted, nil
	case client.QuboleStatusUnknown:
		return PhaseQueryFailed, errors.Errorf(BadQuboleReturnCodeError, "Qubole returned status Unknown")
	default:
		return PhaseQueryFailed, errors.Errorf(BadQuboleReturnCodeError, "default fallthrough case")
	}
}
