package impl

import (
	"context"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/shared"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/util"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/validation"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/interfaces"
	repoInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/transformers"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/flyteorg/flyte/flytestdlib/catalog/datacatalog"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

var (
	_ interfaces.CacheInterface = &CacheManager{}
)

const (
	reservationHeartbeat = 10 * time.Second
)

type cacheMetrics struct {
	Scope                 promutils.Scope
	CacheEvictionTime     promutils.StopWatch
	CacheEvictionSuccess  prometheus.Counter
	CacheEvictionFailures prometheus.Counter
}

type CacheManager struct {
	db            repoInterfaces.Repository
	config        runtimeInterfaces.Configuration
	catalogClient catalog.Client
	metrics       cacheMetrics
}

func (m *CacheManager) EvictTaskExecutionCache(ctx context.Context, req service.EvictTaskExecutionCacheRequest) (*service.EvictCacheResponse, error) {
	if err := validation.ValidateTaskExecutionIdentifier(req.GetTaskExecutionId()); err != nil {
		logger.Debugf(ctx, "EvictTaskExecutionCache request [%+v] failed validation with err: %v", req, err)
		return nil, err
	}

	ctx = getTaskExecutionContext(ctx, req.TaskExecutionId)
	var evictionErrors []*core.CacheEvictionError

	taskExecutionModel, err := util.GetTaskExecutionModel(ctx, m.db, req.TaskExecutionId)
	if err != nil {
		logger.Debugf(ctx, "Failed to get task execution model for task execution ID %+v: %v", req.TaskExecutionId, err)
		return nil, err
	}

	nodeExecutionModel, err := util.GetNodeExecutionModel(ctx, m.db, req.TaskExecutionId.NodeExecutionId)
	if err != nil {
		logger.Debugf(ctx, "Failed to get node execution model for node execution ID %+v: %v", req.TaskExecutionId.NodeExecutionId, err)
		return nil, err
	}

	nodeExecution, err := transformers.FromNodeExecutionModel(*nodeExecutionModel, transformers.DefaultExecutionTransformerOptions)
	if err != nil {
		logger.Debugf(ctx, "Failed to transform node execution model %+v: %v",
			nodeExecutionModel.NodeExecutionKey, err)
		m.metrics.CacheEvictionFailures.Inc()
		evictionErrors = append(evictionErrors, &core.CacheEvictionError{
			NodeExecutionId: &core.NodeExecutionIdentifier{
				NodeId: nodeExecutionModel.NodeID,
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: nodeExecutionModel.NodeExecutionKey.ExecutionKey.Project,
					Domain:  nodeExecutionModel.NodeExecutionKey.ExecutionKey.Domain,
					Name:    nodeExecutionModel.NodeExecutionKey.ExecutionKey.Name,
				},
			},
			Code:    core.CacheEvictionError_INTERNAL,
			Message: "Internal error",
		})
		return &service.EvictCacheResponse{
			Errors: &core.CacheEvictionErrorList{Errors: evictionErrors},
		}, nil
	}

	taskExecution, err := transformers.FromTaskExecutionModel(*taskExecutionModel, transformers.DefaultExecutionTransformerOptions)
	if err != nil {
		logger.Debugf(ctx, "Failed to transform task execution model %+v: %v",
			taskExecutionModel.TaskExecutionKey, err)
		m.metrics.CacheEvictionFailures.Inc()
		evictionErrors = append(evictionErrors, &core.CacheEvictionError{
			NodeExecutionId: nodeExecution.Id,
			Source: &core.CacheEvictionError_TaskExecutionId{
				TaskExecutionId: taskExecution.Id,
			},
			Code:    core.CacheEvictionError_INTERNAL,
			Message: "Internal error",
		})
		return &service.EvictCacheResponse{
			Errors: &core.CacheEvictionErrorList{Errors: evictionErrors},
		}, nil
	}

	logger.Debugf(ctx, "Starting to evict cache for execution %+v of task %+v", req.TaskExecutionId, req.TaskExecutionId.TaskId)

	metadata, ok := nodeExecution.GetClosure().GetTargetMetadata().(*admin.NodeExecutionClosure_TaskNodeMetadata)
	if !ok {
		logger.Debugf(ctx, "Node execution %+v did not contain task node metadata, skipping cache eviction",
			nodeExecution.Id)
		m.metrics.CacheEvictionFailures.Inc()
		evictionErrors = append(evictionErrors, &core.CacheEvictionError{
			NodeExecutionId: nodeExecution.Id,
			Source: &core.CacheEvictionError_WorkflowExecutionId{
				WorkflowExecutionId: nodeExecution.Id.ExecutionId,
			},
			Code:    core.CacheEvictionError_INTERNAL,
			Message: "Internal error",
		})
		return &service.EvictCacheResponse{
			Errors: &core.CacheEvictionErrorList{Errors: evictionErrors},
		}, nil
	}

	evictionErrors = m.evictTaskNodeExecutionCache(ctx, *nodeExecutionModel, nodeExecution, taskExecution, metadata.TaskNodeMetadata, evictionErrors)
	logger.Debugf(ctx, "Finished evicting cache for execution %+v of task %+v", req.TaskExecutionId, req.TaskExecutionId.TaskId)
	return &service.EvictCacheResponse{
		Errors: &core.CacheEvictionErrorList{Errors: evictionErrors},
	}, nil
}

func (m *CacheManager) evictTaskNodeExecutionCache(ctx context.Context, nodeExecutionModel models.NodeExecution,
	nodeExecution *admin.NodeExecution, taskExecution *admin.TaskExecution,
	taskNodeMetadata *admin.TaskNodeMetadata, errors []*core.CacheEvictionError) (evictionErrors []*core.CacheEvictionError) {
	evictionErrors = errors
	if taskNodeMetadata == nil || (taskNodeMetadata.GetCacheStatus() != core.CatalogCacheStatus_CACHE_HIT &&
		taskNodeMetadata.GetCacheStatus() != core.CatalogCacheStatus_CACHE_POPULATED) {
		logger.Debugf(ctx, "Node execution %+v did not contain cached data, skipping cache eviction",
			nodeExecution.Id)
		return evictionErrors
	}

	datasetID := datacatalog.IdentifierToDatasetID(taskNodeMetadata.GetCatalogKey().GetDatasetId())
	artifactTag := taskNodeMetadata.GetCatalogKey().GetArtifactTag().GetName()
	artifactID := taskNodeMetadata.GetCatalogKey().GetArtifactTag().GetArtifactId()
	ownerID := taskExecution.GetClosure().GetMetadata().GetGeneratedName()

	reservation, err := m.catalogClient.GetOrExtendReservationByArtifactTag(ctx, datasetID, artifactTag, ownerID,
		reservationHeartbeat)
	if err != nil {
		logger.Warnf(ctx,
			"Failed to acquire reservation for artifact with tag %q for dataset %+v of node execution %+v: %v",
			artifactTag, datasetID, nodeExecution.Id, err)
		m.metrics.CacheEvictionFailures.Inc()
		evictionErrors = append(evictionErrors, &core.CacheEvictionError{
			NodeExecutionId: nodeExecution.Id,
			Source: &core.CacheEvictionError_TaskExecutionId{
				TaskExecutionId: taskExecution.Id,
			},
			Code:    core.CacheEvictionError_RESERVATION_NOT_ACQUIRED,
			Message: "Failed to acquire reservation for artifact",
		})
		return evictionErrors
	}

	defer func() {
		if err := m.catalogClient.ReleaseReservationByArtifactTag(ctx, datasetID, artifactTag, ownerID); err != nil {
			logger.Warnf(ctx, "Failed to release reservation for artifact of node execution %+v",
				nodeExecution.Id)
			evictionErrors = append(evictionErrors, &core.CacheEvictionError{
				NodeExecutionId: nodeExecution.Id,
				Source: &core.CacheEvictionError_TaskExecutionId{
					TaskExecutionId: taskExecution.Id,
				},
				Code:    core.CacheEvictionError_RESERVATION_NOT_RELEASED,
				Message: "Failed to release reservation for artifact",
			})
		}
	}()

	if len(reservation.GetOwnerId()) == 0 {
		logger.Debugf(ctx, "Received empty owner ID for artifact of node execution %+v, assuming NOOP catalog, skipping cache eviction",
			nodeExecution.Id)
		return evictionErrors
	} else if reservation.GetOwnerId() != ownerID {
		logger.Debugf(ctx, "Artifact of node execution %+v is reserved by different owner, cannot evict cache",
			nodeExecution.Id)
		m.metrics.CacheEvictionFailures.Inc()
		evictionErrors = append(evictionErrors, &core.CacheEvictionError{
			NodeExecutionId: nodeExecution.Id,
			Source: &core.CacheEvictionError_TaskExecutionId{
				TaskExecutionId: taskExecution.Id,
			},
			Code:    core.CacheEvictionError_RESERVATION_NOT_ACQUIRED,
			Message: "Artifact is reserved by different owner",
		})
		return evictionErrors
	}

	if err := m.catalogClient.DeleteByArtifactID(ctx, datasetID, artifactID); err != nil {
		if catalog.IsNotFound(err) {
			logger.Debugf(ctx, "Artifact with ID %s for dataset %+v of node execution %+v not found, assuming already deleted by previous eviction. Skipping...",
				artifactID, datasetID, nodeExecution.Id)
		} else {
			logger.Warnf(ctx, "Failed to delete artifact with ID %s for dataset %+v of node execution %+v: %v",
				artifactID, datasetID, nodeExecution.Id, err)
			m.metrics.CacheEvictionFailures.Inc()
			evictionErrors = append(evictionErrors, &core.CacheEvictionError{
				NodeExecutionId: nodeExecution.Id,
				Source: &core.CacheEvictionError_TaskExecutionId{
					TaskExecutionId: taskExecution.Id,
				},
				Code:    core.CacheEvictionError_ARTIFACT_DELETE_FAILED,
				Message: "Failed to delete artifact",
			})
			return evictionErrors
		}
	}

	selectedFields := []string{shared.CreatedAt}
	nodeExecutionModel.CacheStatus = nil

	taskNodeMetadata.CacheStatus = core.CatalogCacheStatus_CACHE_DISABLED
	nodeExecution.Closure.TargetMetadata = &admin.NodeExecutionClosure_TaskNodeMetadata{
		TaskNodeMetadata: taskNodeMetadata,
	}

	marshaledClosure, err := proto.Marshal(nodeExecution.Closure)
	if err != nil {
		logger.Warnf(ctx, "Failed to marshal updated closure for node execution %+v: err",
			nodeExecution.Id, err)
		m.metrics.CacheEvictionFailures.Inc()
		evictionErrors = append(evictionErrors, &core.CacheEvictionError{
			NodeExecutionId: nodeExecution.Id,
			Source: &core.CacheEvictionError_TaskExecutionId{
				TaskExecutionId: taskExecution.Id,
			},
			Code:    core.CacheEvictionError_INTERNAL,
			Message: "Failed to marshal updated node execution closure",
		})
		return evictionErrors
	}

	nodeExecutionModel.Closure = marshaledClosure
	selectedFields = append(selectedFields, shared.Closure)

	if err := m.db.NodeExecutionRepo().UpdateSelected(ctx, &nodeExecutionModel, selectedFields); err != nil {
		logger.Warnf(ctx, "Failed to update node execution model %+v before after artifact: %v",
			nodeExecution.Id, err)
		m.metrics.CacheEvictionFailures.Inc()
		evictionErrors = append(evictionErrors, &core.CacheEvictionError{
			NodeExecutionId: nodeExecution.Id,
			Source: &core.CacheEvictionError_TaskExecutionId{
				TaskExecutionId: taskExecution.Id,
			},
			Code:    core.CacheEvictionError_DATABASE_UPDATE_FAILED,
			Message: "Failed to update node execution",
		})
		return evictionErrors
	}

	logger.Debugf(ctx, "Successfully evicted cache for task node execution %+v", nodeExecution.Id)
	m.metrics.CacheEvictionSuccess.Inc()

	return evictionErrors
}

func NewCacheManager(db repoInterfaces.Repository, config runtimeInterfaces.Configuration, catalogClient catalog.Client,
	scope promutils.Scope) interfaces.CacheInterface {
	metrics := cacheMetrics{
		Scope: scope,
		CacheEvictionTime: scope.MustNewStopWatch("cache_eviction_duration",
			"duration taken for evicting the cache of a node execution", time.Millisecond),
		CacheEvictionSuccess: scope.MustNewCounter("cache_eviction_success",
			"number of times cache eviction for a node execution succeeded"),
		CacheEvictionFailures: scope.MustNewCounter("cache_eviction_failures",
			"number of times cache eviction for a node execution failed"),
	}

	return &CacheManager{
		db:            db,
		config:        config,
		catalogClient: catalogClient,
		metrics:       metrics,
	}
}
