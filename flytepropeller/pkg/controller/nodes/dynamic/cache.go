package dynamic

import (
	"context"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/encoding"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/common"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/task"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/otelutils"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

// computeCatalogReservationOwnerID constructs a unique identifier which includes the nodes
// parent information, node ID, and retry attempt number. This is used to uniquely identify a task
// when the cache reservation API to serialize cached executions.
func computeCatalogReservationOwnerID(nCtx interfaces.NodeExecutionContext) (string, error) {
	currentNodeUniqueID, err := common.GenerateUniqueID(nCtx.ExecutionContext().GetParentInfo(), nCtx.NodeID())
	if err != nil {
		return "", err
	}

	ownerID, err := encoding.FixedLengthUniqueIDForParts(task.IDMaxLength,
		[]string{nCtx.NodeExecutionMetadata().GetOwnerID().Name, currentNodeUniqueID, strconv.Itoa(int(nCtx.CurrentAttempt()))})
	if err != nil {
		return "", err
	}

	return ownerID, nil
}

// CheckCatalogFutureCache checks if cached outputs exist for the current node and copies them if found
func (d dynamicNodeTaskNodeHandler) CheckCatalogFutureCache(ctx context.Context, nCtx interfaces.NodeExecutionContext, cacheHandler interfaces.CacheableNodeHandler) (catalog.Entry, error) {
	ctx, span := otelutils.NewSpan(ctx, otelutils.FlytePropellerTracer, "pkg.controller.nodes.NodeExecutor/CheckFutureCatalogCache")
	defer span.End()

	// Get catalog entry
	entry, err := d.getFutureCatalogEntry(ctx, nCtx, cacheHandler)
	if err != nil {
		return catalog.Entry{}, err
	}

	// Only process cache hit if we actually got a hit
	if entry.GetStatus().GetCacheStatus() == core.CatalogCacheStatus_CACHE_HIT {
		logger.Debug(ctx, "Future cache hit successfully")
		if err := d.handleCacheHit(ctx, entry, nCtx); err != nil {
			return catalog.Entry{}, err
		}
	}

	return entry, nil
}

// getFutureCatalogEntry retrieves the catalog entry and handles the initial validation
func (d dynamicNodeTaskNodeHandler) getFutureCatalogEntry(ctx context.Context, nCtx interfaces.NodeExecutionContext, cacheHandler interfaces.CacheableNodeHandler) (catalog.Entry, error) {
	catalogKey, err := cacheHandler.GetCatalogKey(ctx, nCtx)
	if err != nil {
		return catalog.Entry{}, errors.Wrapf(err, "failed to initialize the catalogKey")
	}

	entry, err := d.catalog.GetFuture(ctx, catalogKey)
	if err != nil {
		return d.handleCatalogError(ctx, err)
	}

	logger.Infof(ctx, "Catalog CacheHit: for future of dynamic task [%s/%s/%s/%s]", catalogKey.Identifier.GetProject(),
		catalogKey.Identifier.GetDomain(), catalogKey.Identifier.GetName(), catalogKey.Identifier.GetVersion())
	d.metrics.catalogHitCount.Inc(ctx)

	return entry, nil
}

// handleCatalogError processes errors that occur during catalog operations
func (d dynamicNodeTaskNodeHandler) handleCatalogError(ctx context.Context, err error) (catalog.Entry, error) {
	causeErr := errors.Cause(err)
	if taskStatus, ok := status.FromError(causeErr); ok && taskStatus.Code() == codes.NotFound {
		d.metrics.catalogMissCount.Inc(ctx)
		logger.Infof(ctx, "Future Catalog CacheMiss: Future Artifact not found in Catalog. Executing Task.")
		return catalog.NewCatalogEntry(nil, catalog.NewStatus(core.CatalogCacheStatus_CACHE_MISS, nil)), nil
	}
	d.metrics.catalogGetFailureCount.Inc(ctx)
	logger.Errorf(ctx, "Catalog Failure: Future memoization check failed. err: %v", err.Error())
	return catalog.Entry{}, errors.Wrapf(err, "Failed to check Catalog for previous results")
}

// handleCacheHit processes successful cache hits and manages the data transformation
func (d dynamicNodeTaskNodeHandler) handleCacheHit(ctx context.Context, entry catalog.Entry, nCtx interfaces.NodeExecutionContext) error {
	// Read and convert the output
	djSpec, err := d.readAndConvertOutput(ctx, entry)
	if err != nil {
		return err
	}

	// Write to data store
	return d.writeToDynamicJobSpec(ctx, nCtx, djSpec)
}

// readAndConvertOutput reads the catalog entry output and converts it to DynamicJobSpec
func (d dynamicNodeTaskNodeHandler) readAndConvertOutput(ctx context.Context, entry catalog.Entry) (*core.DynamicJobSpec, error) {
	o, ee, err := entry.GetOutputs().Read(ctx)
	if err != nil {
		logger.Errorf(ctx, "failed to read future from catalog, err: %s", err.Error())
		return nil, err
	}
	if ee != nil {
		logger.Errorf(ctx, "got execution error from catalog future reader, err: %s", ee.String())
		return nil, errors.New("execution error from a cache future output")
	}

	djSpec, err := FromFutureLiteralMapToDynamicJobSpec(o)
	if err != nil {
		logger.Errorf(ctx, "got error while converting Future LiteralMap into DynamicJobSpec, err: %s", err.Error())
		return nil, errors.New("execution from converting Future LiteralMap into DynamicJobSpec")
	}
	return djSpec, nil
}

// writeToDynamicJobSpec writes the DynamicJobSpec to the data store
func (d dynamicNodeTaskNodeHandler) writeToDynamicJobSpec(ctx context.Context, nCtx interfaces.NodeExecutionContext, djSpec *core.DynamicJobSpec) error {
	f, err := task.NewRemoteFutureFileReader(ctx, nCtx.NodeStatus().GetOutputDir(), nCtx.DataStore())
	if err != nil {
		logger.Errorf(ctx, "got error while getting future file reader, err: %s", err)
		return errors.New("error from getting future file reader")
	}

	futureFile := f.GetLoc()
	if err := nCtx.DataStore().WriteProtobuf(ctx, futureFile, storage.Options{}, djSpec); err != nil {
		logger.Errorf(ctx, "failed to write cached future value to datastore, err: %s", err.Error())
		return err
	}
	return nil
}

func (d dynamicNodeTaskNodeHandler) WriteCatalogFutureCache(ctx context.Context, nCtx interfaces.NodeExecutionContext, cacheHandler interfaces.CacheableNodeHandler) (catalog.Status, error) {
	ctx, span := otelutils.NewSpan(ctx, otelutils.FlytePropellerTracer, "pkg.controller.nodes.NodeExecutor/WriteFutureCatalogCache")
	defer span.End()
	catalogKey, err := cacheHandler.GetCatalogKey(ctx, nCtx)
	if err != nil {
		return catalog.NewStatus(core.CatalogCacheStatus_CACHE_DISABLED, nil), errors.Wrapf(err, "failed to initialize the catalogKey")
	}
	metadata := catalog.Metadata{
		TaskExecutionIdentifier: task.GetTaskExecutionIdentifier(nCtx),
	}

	f, err := task.NewRemoteFutureFileReader(ctx, nCtx.NodeStatus().GetOutputDir(), nCtx.DataStore())
	if err != nil {
		return catalog.NewStatus(core.CatalogCacheStatus_CACHE_DISABLED, nil), errors.Wrapf(err, "failed to initiate the future file reader")
	}

	var status catalog.Status
	if nCtx.ExecutionContext().GetExecutionConfig().OverwriteCache {
		status, err = d.catalog.UpdateFuture(ctx, catalogKey, f, metadata)
	} else {
		status, err = d.catalog.PutFuture(ctx, catalogKey, f, metadata)
	}
	if err != nil {
		d.metrics.CacheError.Inc(ctx)
		logger.Errorf(ctx, "Failed to write results to catalog for Task [%v]. Error: %v", catalogKey.Identifier, err)
		return catalog.NewStatus(core.CatalogCacheStatus_CACHE_PUT_FAILURE, status.GetMetadata()), nil
	}

	d.metrics.catalogPutSuccessCount.Inc(ctx)
	logger.Infof(ctx, "Successfully cached future.pb to catalog - DynamicTask [%v]", catalogKey.Identifier)
	return status, nil
}

// GetOrExtendCatalogReservation attempts to acquire an artifact reservation if the task is
// cacheable and cache serializable. If the reservation already exists for this owner, the
// reservation is extended.
func (d *dynamicNodeTaskNodeHandler) GetOrExtendCatalogReservation(ctx context.Context, nCtx interfaces.NodeExecutionContext,
	cacheHandler interfaces.CacheableNodeHandler, heartbeatInterval time.Duration) (catalog.ReservationEntry, error) {

	catalogKey, err := cacheHandler.GetCatalogKey(ctx, nCtx)
	if err != nil {
		return catalog.NewReservationEntryStatus(core.CatalogReservation_RESERVATION_DISABLED),
			errors.Wrapf(err, "failed to initialize the future catalogKey")
	}

	ownerID, err := computeCatalogReservationOwnerID(nCtx)
	if err != nil {
		return catalog.NewReservationEntryStatus(core.CatalogReservation_RESERVATION_DISABLED),
			errors.Wrapf(err, "failed to initialize the cache reservation ownerID")
	}

	reservation, err := d.catalog.GetOrExtendReservation(ctx, catalogKey, ownerID, heartbeatInterval)
	if err != nil {
		d.metrics.reservationGetFailureCount.Inc(ctx)
		logger.Errorf(ctx, "Catalog Failure: future reservation get or extend failed. err: %v", err.Error())
		return catalog.NewReservationEntryStatus(core.CatalogReservation_RESERVATION_FAILURE), err
	}

	var status core.CatalogReservation_Status
	if reservation.GetOwnerId() == ownerID {
		status = core.CatalogReservation_RESERVATION_ACQUIRED
	} else {
		status = core.CatalogReservation_RESERVATION_EXISTS
	}

	d.metrics.reservationGetSuccessCount.Inc(ctx)
	return catalog.NewReservationEntry(reservation.GetExpiresAt().AsTime(),
		reservation.GetHeartbeatInterval().AsDuration(), reservation.GetOwnerId(), status), nil
}

// ReleaseCatalogReservation attempts to release an artifact reservation if the task is cacheable
// and cache serializable. If the reservation does not exist for this owner (e.x. it never existed
// or has been acquired by another owner) this call is still successful.
func (d *dynamicNodeTaskNodeHandler) ReleaseCatalogReservation(ctx context.Context, nCtx interfaces.NodeExecutionContext,
	cacheHandler interfaces.CacheableNodeHandler) (catalog.ReservationEntry, error) {

	catalogKey, err := cacheHandler.GetCatalogKey(ctx, nCtx)
	if err != nil {
		return catalog.NewReservationEntryStatus(core.CatalogReservation_RESERVATION_DISABLED),
			errors.Wrapf(err, "failed to initialize the catalogKey")
	}

	ownerID, err := computeCatalogReservationOwnerID(nCtx)
	if err != nil {
		return catalog.NewReservationEntryStatus(core.CatalogReservation_RESERVATION_DISABLED),
			errors.Wrapf(err, "failed to initialize the cache reservation ownerID")
	}

	err = d.catalog.ReleaseReservation(ctx, catalogKey, ownerID)
	if err != nil {
		d.metrics.reservationReleaseFailureCount.Inc(ctx)
		logger.Errorf(ctx, "Catalog Failure: release reservation failed. err: %v", err.Error())
		return catalog.NewReservationEntryStatus(core.CatalogReservation_RESERVATION_FAILURE), err
	}

	d.metrics.reservationReleaseSuccessCount.Inc(ctx)
	return catalog.NewReservationEntryStatus(core.CatalogReservation_RESERVATION_RELEASED), nil
}

func (d *dynamicNodeTaskNodeHandler) CheckFutureCache(ctx context.Context, nCtx interfaces.NodeExecutionContext, cacheHandler interfaces.CacheableNodeHandler) (*catalog.Status, error) {
	var cacheStatus *catalog.Status
	if nCtx.ExecutionContext().GetExecutionConfig().OverwriteCache {
		logger.Info(ctx, "execution config forced cache skip, not checking future catalog")
		status := catalog.NewStatus(core.CatalogCacheStatus_CACHE_SKIPPED, nil)
		cacheStatus = &status
		d.metrics.catalogSkipCount.Inc(ctx)
	} else {
		entry, err := d.CheckCatalogFutureCache(ctx, nCtx, cacheHandler)
		if err != nil {
			logger.Errorf(ctx, "failed to check the future catalog cache with err '%s'", err.Error())
			return nil, err
		}
		status := entry.GetStatus()
		cacheStatus = &status
	}
	return cacheStatus, nil
}

func (d *dynamicNodeTaskNodeHandler) WriteFutureCache(ctx context.Context, nCtx interfaces.NodeExecutionContext, cacheHandler interfaces.CacheableNodeHandler) (*catalog.Status, error) {
	status, err := d.WriteCatalogFutureCache(ctx, nCtx, cacheHandler)
	if err != nil {
		return nil, err
	}

	logger.Infof(ctx, "future data written to catalog cache with status %v", status.GetCacheStatus())
	return &status, nil
}
