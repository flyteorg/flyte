package task

import (
	"context"
	"time"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/ioutils"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/catalog"
	pluginCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	errors2 "github.com/flyteorg/flytepropeller/pkg/controller/nodes/errors"
)

var cacheDisabled = catalog.NewStatus(core.CatalogCacheStatus_CACHE_DISABLED, nil)

func (t *Handler) CheckCatalogCache(ctx context.Context, tr pluginCore.TaskReader, inputReader io.InputReader, outputWriter io.OutputWriter) (catalog.Entry, error) {
	tk, err := tr.Read(ctx)
	if err != nil {
		logger.Errorf(ctx, "Failed to read TaskTemplate, error :%s", err.Error())
		return catalog.Entry{}, err
	}

	if tk.Metadata.Discoverable {
		logger.Infof(ctx, "Catalog CacheEnabled: Looking up catalog Cache.")
		key := catalog.Key{
			Identifier:     *tk.Id,
			CacheVersion:   tk.Metadata.DiscoveryVersion,
			TypedInterface: *tk.Interface,
			InputReader:    inputReader,
		}

		resp, err := t.catalog.Get(ctx, key)
		if err != nil {
			causeErr := errors.Cause(err)
			if taskStatus, ok := status.FromError(causeErr); ok && taskStatus.Code() == codes.NotFound {
				t.metrics.catalogMissCount.Inc(ctx)
				logger.Infof(ctx, "Catalog CacheMiss: Artifact not found in Catalog. Executing Task.")
				return catalog.NewCatalogEntry(nil, catalog.NewStatus(core.CatalogCacheStatus_CACHE_MISS, nil)), nil
			}

			t.metrics.catalogGetFailureCount.Inc(ctx)
			logger.Errorf(ctx, "Catalog Failure: memoization check failed. err: %v", err.Error())
			return catalog.Entry{}, errors.Wrapf(err, "Failed to check Catalog for previous results")
		}

		if resp.GetStatus().GetCacheStatus() != core.CatalogCacheStatus_CACHE_HIT {
			logger.Errorf(ctx, "No CacheHIT and no Error received. Illegal state, Cache State: %s", resp.GetStatus().GetCacheStatus().String())
			// TODO should this be an error?
			return resp, nil
		}

		logger.Infof(ctx, "Catalog CacheHit: for task [%s/%s/%s/%s]", tk.Id.Project, tk.Id.Domain, tk.Id.Name, tk.Id.Version)
		t.metrics.catalogHitCount.Inc(ctx)
		if iface := tk.Interface; iface != nil && iface.Outputs != nil && len(iface.Outputs.Variables) > 0 {
			if err := outputWriter.Put(ctx, resp.GetOutputs()); err != nil {
				logger.Errorf(ctx, "failed to write data to Storage, err: %v", err.Error())
				return catalog.Entry{}, errors.Wrapf(err, "failed to copy cached results for task.")
			}
		}
		// SetCached.
		return resp, nil
	}
	logger.Infof(ctx, "Catalog CacheDisabled: for Task [%s/%s/%s/%s]", tk.Id.Project, tk.Id.Domain, tk.Id.Name, tk.Id.Version)
	return catalog.NewCatalogEntry(nil, cacheDisabled), nil
}

// GetOrExtendCatalogReservation attempts to acquire an artifact reservation if the task is
// cachable and cache serializable. If the reservation already exists for this owner, the
// reservation is extended.
func (t *Handler) GetOrExtendCatalogReservation(ctx context.Context, ownerID string, heartbeatInterval time.Duration, tr pluginCore.TaskReader, inputReader io.InputReader) (catalog.ReservationEntry, error) {
	tk, err := tr.Read(ctx)
	if err != nil {
		logger.Errorf(ctx, "Failed to read TaskTemplate, error :%s", err.Error())
		return catalog.NewReservationEntryStatus(core.CatalogReservation_RESERVATION_FAILURE), err
	}

	if tk.Metadata.Discoverable && tk.Metadata.CacheSerializable {
		logger.Infof(ctx, "Catalog CacheSerializeEnabled: creating catalog reservation.")
		key := catalog.Key{
			Identifier:     *tk.Id,
			CacheVersion:   tk.Metadata.DiscoveryVersion,
			TypedInterface: *tk.Interface,
			InputReader:    inputReader,
		}

		reservation, err := t.catalog.GetOrExtendReservation(ctx, key, ownerID, heartbeatInterval)
		if err != nil {
			t.metrics.reservationGetFailureCount.Inc(ctx)
			logger.Errorf(ctx, "Catalog Failure: reservation get or extend failed. err: %v", err.Error())
			return catalog.NewReservationEntryStatus(core.CatalogReservation_RESERVATION_FAILURE), err
		}

		expiresAt := reservation.ExpiresAt.AsTime()
		heartbeatInterval := reservation.HeartbeatInterval.AsDuration()

		var status core.CatalogReservation_Status
		if reservation.OwnerId == ownerID {
			status = core.CatalogReservation_RESERVATION_ACQUIRED
		} else {
			status = core.CatalogReservation_RESERVATION_EXISTS
		}

		t.metrics.reservationGetSuccessCount.Inc(ctx)
		return catalog.NewReservationEntry(expiresAt, heartbeatInterval, reservation.OwnerId, status), nil
	}
	logger.Infof(ctx, "Catalog CacheSerializeDisabled: for Task [%s/%s/%s/%s]", tk.Id.Project, tk.Id.Domain, tk.Id.Name, tk.Id.Version)
	return catalog.NewReservationEntryStatus(core.CatalogReservation_RESERVATION_DISABLED), nil
}

func (t *Handler) ValidateOutputAndCacheAdd(ctx context.Context, nodeID v1alpha1.NodeID, i io.InputReader,
	r io.OutputReader, outputCommitter io.OutputWriter, executionConfig v1alpha1.ExecutionConfig,
	tr ioutils.SimpleTaskReader, m catalog.Metadata) (catalog.Status, *io.ExecutionError, error) {

	tk, err := tr.Read(ctx)
	if err != nil {
		logger.Errorf(ctx, "Failed to read TaskTemplate, error :%s", err.Error())
		return cacheDisabled, nil, err
	}

	iface := tk.Interface
	outputsDeclared := iface != nil && iface.Outputs != nil && len(iface.Outputs.Variables) > 0

	if r == nil {
		if outputsDeclared {
			// Whack! plugin did not return any outputs for this task
			// Also When an error is observed, cache is automatically disabled
			return cacheDisabled, &io.ExecutionError{
				ExecutionError: &core.ExecutionError{
					Code:    "OutputsNotGenerated",
					Message: "Output Reader was nil. Plugin/Platform problem.",
				},
				IsRecoverable: true,
			}, nil
		}
		return cacheDisabled, nil, nil
	}
	// Reader exists, we can check for error, even if this task may not have any outputs declared
	y, err := r.IsError(ctx)
	if err != nil {
		return cacheDisabled, nil, err
	}
	if y {
		taskErr, err := r.ReadError(ctx)
		if err != nil {
			return cacheDisabled, nil, err
		}

		if taskErr.ExecutionError == nil {
			taskErr.ExecutionError = &core.ExecutionError{Kind: core.ExecutionError_UNKNOWN, Code: "Unknown", Message: "Unknown"}
		}
		// Errors can be arbitrary long since they are written by containers/potentially 3rd party plugins. This ensures
		// the error message length will never be big enough to cause write failures to Etcd. or spam Admin DB with huge
		// objects.
		taskErr.Message = trimErrorMessage(taskErr.Message, t.cfg.MaxErrorMessageLength)
		return cacheDisabled, &taskErr, nil
	}

	// Do this if we have outputs declared for the Handler interface!
	if !outputsDeclared {
		return cacheDisabled, nil, nil
	}
	ok, err := r.Exists(ctx)
	if err != nil {
		logger.Errorf(ctx, "Failed to check if the output file exists. Error: %s", err.Error())
		return cacheDisabled, nil, err
	}

	if !ok {
		// Does not exist
		return cacheDisabled,
			&io.ExecutionError{
				ExecutionError: &core.ExecutionError{
					Code:    "OutputsNotFound",
					Message: "Outputs not generated by task execution",
				},
				IsRecoverable: true,
			}, nil
	}

	if !r.IsFile(ctx) {
		// Read output and write to file
		// No need to check for Execution Error here as we have done so above this block.
		err = outputCommitter.Put(ctx, r)
		if err != nil {
			logger.Errorf(ctx, "Failed to commit output to remote location. Error: %v", err)
			return cacheDisabled, nil, err
		}
	}

	p, err := t.ResolvePlugin(ctx, tk.Type, executionConfig)
	if err != nil {
		return cacheDisabled, nil, errors2.Wrapf(errors2.UnsupportedTaskTypeError, nodeID, err, "unable to resolve plugin")
	}
	writeToCatalog := !p.GetProperties().DisableNodeLevelCaching

	if !tk.Metadata.Discoverable || !writeToCatalog {
		if !writeToCatalog {
			logger.Infof(ctx, "Node level caching is disabled. Skipping catalog write.")
		}
		return cacheDisabled, nil, nil
	}

	cacheVersion := "0"
	if tk.Metadata != nil {
		cacheVersion = tk.Metadata.DiscoveryVersion
	}

	key := catalog.Key{
		Identifier:     *tk.Id,
		CacheVersion:   cacheVersion,
		TypedInterface: *tk.Interface,
		InputReader:    i,
	}

	logger.Infof(ctx, "Catalog CacheEnabled. recording execution [%s/%s/%s/%s]", tk.Id.Project, tk.Id.Domain, tk.Id.Name, tk.Id.Version)
	// ignores discovery write failures
	s, err2 := t.catalog.Put(ctx, key, r, m)
	if err2 != nil {
		t.metrics.catalogPutFailureCount.Inc(ctx)
		logger.Errorf(ctx, "Failed to write results to catalog for Task [%v]. Error: %v", tk.GetId(), err2)
		return catalog.NewStatus(core.CatalogCacheStatus_CACHE_PUT_FAILURE, s.GetMetadata()), nil, nil
	}
	t.metrics.catalogPutSuccessCount.Inc(ctx)
	logger.Infof(ctx, "Successfully cached results to catalog - Task [%v]", tk.GetId())
	return s, nil, nil
}

// ReleaseCatalogReservation attempts to release an artifact reservation if the task is cachable
// and cache serializable. If the reservation does not exist for this owner (e.x. it never existed
// or has been acquired by another owner) this call is still successful.
func (t *Handler) ReleaseCatalogReservation(ctx context.Context, ownerID string, tr pluginCore.TaskReader, inputReader io.InputReader) (catalog.ReservationEntry, error) {
	tk, err := tr.Read(ctx)
	if err != nil {
		logger.Errorf(ctx, "Failed to read TaskTemplate, error :%s", err.Error())
		return catalog.NewReservationEntryStatus(core.CatalogReservation_RESERVATION_FAILURE), err
	}

	if tk.Metadata.Discoverable && tk.Metadata.CacheSerializable {
		logger.Infof(ctx, "Catalog CacheSerializeEnabled: releasing catalog reservation.")
		key := catalog.Key{
			Identifier:     *tk.Id,
			CacheVersion:   tk.Metadata.DiscoveryVersion,
			TypedInterface: *tk.Interface,
			InputReader:    inputReader,
		}

		err := t.catalog.ReleaseReservation(ctx, key, ownerID)
		if err != nil {
			t.metrics.reservationReleaseFailureCount.Inc(ctx)
			logger.Errorf(ctx, "Catalog Failure: release reservation failed. err: %v", err.Error())
			return catalog.NewReservationEntryStatus(core.CatalogReservation_RESERVATION_FAILURE), err
		}

		t.metrics.reservationReleaseSuccessCount.Inc(ctx)
		return catalog.NewReservationEntryStatus(core.CatalogReservation_RESERVATION_RELEASED), nil
	}
	logger.Infof(ctx, "Catalog CacheSerializeDisabled: for Task [%s/%s/%s/%s]", tk.Id.Project, tk.Id.Domain, tk.Id.Name, tk.Id.Version)
	return catalog.NewReservationEntryStatus(core.CatalogReservation_RESERVATION_DISABLED), nil
}
