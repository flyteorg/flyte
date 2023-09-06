package nodes

import (
	"context"
	"strconv"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/encoding"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/ioutils"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/common"
	nodeserrors "github.com/flyteorg/flytepropeller/pkg/controller/nodes/errors"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/handler"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/interfaces"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/task"

	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/storage"

	"github.com/pkg/errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

// updatePhaseCacheInfo adds the cache and catalog reservation metadata to the PhaseInfo. This
// ensures this information is reported in events and available within FlyteAdmin.
func updatePhaseCacheInfo(phaseInfo handler.PhaseInfo, cacheStatus *catalog.Status, reservationStatus *core.CatalogReservation_Status) handler.PhaseInfo {
	if cacheStatus == nil && reservationStatus == nil {
		return phaseInfo
	}

	info := phaseInfo.GetInfo()
	if info == nil {
		info = &handler.ExecutionInfo{}
	}

	if info.TaskNodeInfo == nil {
		info.TaskNodeInfo = &handler.TaskNodeInfo{}
	}

	if info.TaskNodeInfo.TaskNodeMetadata == nil {
		info.TaskNodeInfo.TaskNodeMetadata = &event.TaskNodeMetadata{}
	}

	if cacheStatus != nil {
		info.TaskNodeInfo.TaskNodeMetadata.CacheStatus = cacheStatus.GetCacheStatus()
		info.TaskNodeInfo.TaskNodeMetadata.CatalogKey = cacheStatus.GetMetadata()
	}

	if reservationStatus != nil {
		info.TaskNodeInfo.TaskNodeMetadata.ReservationStatus = *reservationStatus
	}

	return phaseInfo.WithInfo(info)
}

// CheckCatalogCache uses the handler and contexts to check if cached outputs for the current node
// exist. If the exist, this function also copies the outputs to this node.
func (n *nodeExecutor) CheckCatalogCache(ctx context.Context, nCtx interfaces.NodeExecutionContext, cacheHandler interfaces.CacheableNodeHandler) (catalog.Entry, error) {
	catalogKey, err := cacheHandler.GetCatalogKey(ctx, nCtx)
	if err != nil {
		return catalog.Entry{}, errors.Wrapf(err, "failed to initialize the catalogKey")
	}

	entry, err := n.catalog.Get(ctx, catalogKey)
	if err != nil {
		causeErr := errors.Cause(err)
		if taskStatus, ok := status.FromError(causeErr); ok && taskStatus.Code() == codes.NotFound {
			n.metrics.catalogMissCount.Inc(ctx)
			logger.Infof(ctx, "Catalog CacheMiss: Artifact not found in Catalog. Executing Task.")
			return catalog.NewCatalogEntry(nil, catalog.NewStatus(core.CatalogCacheStatus_CACHE_MISS, nil)), nil
		}

		n.metrics.catalogGetFailureCount.Inc(ctx)
		logger.Errorf(ctx, "Catalog Failure: memoization check failed. err: %v", err.Error())
		return catalog.Entry{}, errors.Wrapf(err, "Failed to check Catalog for previous results")
	}

	if entry.GetStatus().GetCacheStatus() != core.CatalogCacheStatus_CACHE_HIT {
		logger.Errorf(ctx, "No CacheHIT and no Error received. Illegal state, Cache State: %s", entry.GetStatus().GetCacheStatus().String())
		// TODO should this be an error?
		return entry, nil
	}

	logger.Infof(ctx, "Catalog CacheHit: for task [%s/%s/%s/%s]", catalogKey.Identifier.Project,
		catalogKey.Identifier.Domain, catalogKey.Identifier.Name, catalogKey.Identifier.Version)
	n.metrics.catalogHitCount.Inc(ctx)

	iface := catalogKey.TypedInterface
	if iface.Outputs != nil && len(iface.Outputs.Variables) > 0 {
		// copy cached outputs to node outputs
		o, ee, err := entry.GetOutputs().Read(ctx)
		if err != nil {
			logger.Errorf(ctx, "failed to read from catalog, err: %s", err.Error())
			return catalog.Entry{}, err
		} else if ee != nil {
			logger.Errorf(ctx, "got execution error from catalog output reader? This should not happen, err: %s", ee.String())
			return catalog.Entry{}, nodeserrors.Errorf(nodeserrors.IllegalStateError, nCtx.NodeID(), "execution error from a cache output, bad state: %s", ee.String())
		}

		outputFile := v1alpha1.GetOutputsFile(nCtx.NodeStatus().GetOutputDir())
		if err := nCtx.DataStore().WriteProtobuf(ctx, outputFile, storage.Options{}, o); err != nil {
			logger.Errorf(ctx, "failed to write cached value to datastore, err: %s", err.Error())
			return catalog.Entry{}, err
		}
	}

	return entry, nil
}

// GetOrExtendCatalogReservation attempts to acquire an artifact reservation if the task is
// cachable and cache serializable. If the reservation already exists for this owner, the
// reservation is extended.
func (n *nodeExecutor) GetOrExtendCatalogReservation(ctx context.Context, nCtx interfaces.NodeExecutionContext,
	cacheHandler interfaces.CacheableNodeHandler, heartbeatInterval time.Duration) (catalog.ReservationEntry, error) {

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

	reservation, err := n.catalog.GetOrExtendReservation(ctx, catalogKey, ownerID, heartbeatInterval)
	if err != nil {
		n.metrics.reservationGetFailureCount.Inc(ctx)
		logger.Errorf(ctx, "Catalog Failure: reservation get or extend failed. err: %v", err.Error())
		return catalog.NewReservationEntryStatus(core.CatalogReservation_RESERVATION_FAILURE), err
	}

	var status core.CatalogReservation_Status
	if reservation.OwnerId == ownerID {
		status = core.CatalogReservation_RESERVATION_ACQUIRED
	} else {
		status = core.CatalogReservation_RESERVATION_EXISTS
	}

	n.metrics.reservationGetSuccessCount.Inc(ctx)
	return catalog.NewReservationEntry(reservation.ExpiresAt.AsTime(),
		reservation.HeartbeatInterval.AsDuration(), reservation.OwnerId, status), nil
}

// ReleaseCatalogReservation attempts to release an artifact reservation if the task is cachable
// and cache serializable. If the reservation does not exist for this owner (e.x. it never existed
// or has been acquired by another owner) this call is still successful.
func (n *nodeExecutor) ReleaseCatalogReservation(ctx context.Context, nCtx interfaces.NodeExecutionContext,
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

	err = n.catalog.ReleaseReservation(ctx, catalogKey, ownerID)
	if err != nil {
		n.metrics.reservationReleaseFailureCount.Inc(ctx)
		logger.Errorf(ctx, "Catalog Failure: release reservation failed. err: %v", err.Error())
		return catalog.NewReservationEntryStatus(core.CatalogReservation_RESERVATION_FAILURE), err
	}

	n.metrics.reservationReleaseSuccessCount.Inc(ctx)
	return catalog.NewReservationEntryStatus(core.CatalogReservation_RESERVATION_RELEASED), nil
}

// WriteCatalogCache relays the outputs of this node to the cache. This allows future executions
// to reuse these data to avoid recomputation.
func (n *nodeExecutor) WriteCatalogCache(ctx context.Context, nCtx interfaces.NodeExecutionContext, cacheHandler interfaces.CacheableNodeHandler) (catalog.Status, error) {
	catalogKey, err := cacheHandler.GetCatalogKey(ctx, nCtx)
	if err != nil {
		return catalog.NewStatus(core.CatalogCacheStatus_CACHE_DISABLED, nil), errors.Wrapf(err, "failed to initialize the catalogKey")
	}

	iface := catalogKey.TypedInterface
	if iface.Outputs != nil && len(iface.Outputs.Variables) == 0 {
		return catalog.NewStatus(core.CatalogCacheStatus_CACHE_DISABLED, nil), nil
	}

	logger.Infof(ctx, "Catalog CacheEnabled. recording execution [%s/%s/%s/%s]", catalogKey.Identifier.Project,
		catalogKey.Identifier.Domain, catalogKey.Identifier.Name, catalogKey.Identifier.Version)

	outputPaths := ioutils.NewReadOnlyOutputFilePaths(ctx, nCtx.DataStore(), nCtx.NodeStatus().GetOutputDir())
	outputReader := ioutils.NewRemoteFileOutputReader(ctx, nCtx.DataStore(), outputPaths, nCtx.MaxDatasetSizeBytes())
	metadata := catalog.Metadata{
		TaskExecutionIdentifier: task.GetTaskExecutionIdentifier(nCtx),
	}

	// ignores discovery write failures
	status, err := n.catalog.Put(ctx, catalogKey, outputReader, metadata)
	if err != nil {
		n.metrics.catalogPutFailureCount.Inc(ctx)
		logger.Errorf(ctx, "Failed to write results to catalog for Task [%v]. Error: %v", catalogKey.Identifier, err)
		return catalog.NewStatus(core.CatalogCacheStatus_CACHE_PUT_FAILURE, status.GetMetadata()), nil
	}

	n.metrics.catalogPutSuccessCount.Inc(ctx)
	logger.Infof(ctx, "Successfully cached results to catalog - Task [%v]", catalogKey.Identifier)
	return status, nil
}
