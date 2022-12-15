package impl

import (
	"context"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/flyteorg/flyteadmin/pkg/manager/impl/shared"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/util"
	"github.com/flyteorg/flyteadmin/pkg/manager/impl/validation"
	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	repoInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteadmin/pkg/repositories/transformers"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytestdlib/catalog/datacatalog"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
)

var (
	_ interfaces.CacheInterface = &CacheManager{}
)

var sortByCreatedAtAsc = &admin.Sort{Key: shared.CreatedAt, Direction: admin.Sort_ASCENDING}

const (
	nodeExecutionLimit   = 10000
	taskExecutionLimit   = 10000
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

func (m *CacheManager) EvictExecutionCache(ctx context.Context, req service.EvictExecutionCacheRequest) (*service.EvictCacheResponse, error) {
	if err := validation.ValidateWorkflowExecutionIdentifier(req.GetWorkflowExecutionId()); err != nil {
		logger.Debugf(ctx, "EvictExecutionCache request [%+v] failed validation with err: %v", req, err)
		return nil, err
	}

	ctx = getExecutionContext(ctx, req.WorkflowExecutionId)

	executionModel, err := util.GetExecutionModel(ctx, m.db, *req.WorkflowExecutionId)
	if err != nil {
		logger.Debugf(ctx, "Failed to get workflow execution model for workflow execution ID %+v: %v", req.WorkflowExecutionId, err)
		return nil, err
	}

	logger.Debugf(ctx, "Starting to evict cache for execution %d of workflow %+v", executionModel.ID, req.WorkflowExecutionId)

	nodeExecutions, err := m.listAllNodeExecutionsForWorkflow(ctx, req.WorkflowExecutionId, "")
	if err != nil {
		logger.Debugf(ctx, "Failed to list all node executions for execution %d of workflow %+v: %v", executionModel.ID, req.WorkflowExecutionId, err)
		return nil, err
	}

	var evictionErrors []*core.CacheEvictionError
	for _, nodeExecution := range nodeExecutions {
		evictionErrors = m.evictNodeExecutionCache(ctx, nodeExecution, evictionErrors)
	}

	logger.Debugf(ctx, "Finished evicting cache for execution %d of workflow %+v", executionModel.ID, req.WorkflowExecutionId)
	return &service.EvictCacheResponse{
		Errors: &core.CacheEvictionErrorList{Errors: evictionErrors},
	}, nil
}

func (m *CacheManager) EvictTaskExecutionCache(ctx context.Context, req service.EvictTaskExecutionCacheRequest) (*service.EvictCacheResponse, error) {
	if err := validation.ValidateTaskExecutionIdentifier(req.GetTaskExecutionId()); err != nil {
		logger.Debugf(ctx, "EvictTaskExecutionCache request [%+v] failed validation with err: %v", req, err)
		return nil, err
	}

	ctx = getTaskExecutionContext(ctx, req.TaskExecutionId)

	// Sanity check to ensure referenced task execution exists although only encapsulated node execution is strictly required
	_, err := util.GetTaskExecutionModel(ctx, m.db, req.TaskExecutionId)
	if err != nil {
		logger.Debugf(ctx, "Failed to get task execution model for task execution ID %+v: %v", req.TaskExecutionId, err)
		return nil, err
	}

	logger.Debugf(ctx, "Starting to evict cache for execution %+v of task %+v", req.TaskExecutionId, req.TaskExecutionId.TaskId)

	nodeExecutionModel, err := util.GetNodeExecutionModel(ctx, m.db, req.TaskExecutionId.NodeExecutionId)
	if err != nil {
		logger.Debugf(ctx, "Failed to get node execution model for node execution ID %+v: %v", req.TaskExecutionId.NodeExecutionId, err)
		return nil, err
	}

	evictionErrors := m.evictNodeExecutionCache(ctx, *nodeExecutionModel, nil)

	logger.Debugf(ctx, "Finished evicting cache for execution %+v of task %+v", req.TaskExecutionId, req.TaskExecutionId.TaskId)
	return &service.EvictCacheResponse{
		Errors: &core.CacheEvictionErrorList{Errors: evictionErrors},
	}, nil
}

func (m *CacheManager) evictNodeExecutionCache(ctx context.Context, nodeExecutionModel models.NodeExecution,
	evictionErrors []*core.CacheEvictionError) []*core.CacheEvictionError {
	if strings.HasSuffix(nodeExecutionModel.NodeID, v1alpha1.StartNodeID) || strings.HasSuffix(nodeExecutionModel.NodeID, v1alpha1.EndNodeID) {
		return evictionErrors
	}

	nodeExecution, err := transformers.FromNodeExecutionModel(nodeExecutionModel)
	if err != nil {
		logger.Warnf(ctx, "Failed to transform node execution model %+v: %v",
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
		return evictionErrors
	}

	logger.Debugf(ctx, "Starting to evict cache for node execution %+v", nodeExecution.Id)
	t := m.metrics.CacheEvictionTime.Start()
	defer t.Stop()

	if nodeExecution.Metadata.IsParentNode {
		var childNodeExecutions []models.NodeExecution
		if nodeExecution.Metadata.IsDynamic {
			var err error
			childNodeExecutions, err = m.listAllNodeExecutionsForWorkflow(ctx, nodeExecution.Id.ExecutionId, nodeExecution.Id.NodeId)
			if err != nil {
				logger.Warnf(ctx, "Failed to list child node executions for dynamic node execution %+v: %v",
					nodeExecution.Id, err)
				m.metrics.CacheEvictionFailures.Inc()
				evictionErrors = append(evictionErrors, &core.CacheEvictionError{
					NodeExecutionId: nodeExecution.Id,
					Source: &core.CacheEvictionError_WorkflowExecutionId{
						WorkflowExecutionId: nodeExecution.Id.ExecutionId,
					},
					Code:    core.CacheEvictionError_INTERNAL,
					Message: "Failed to list child node executions",
				})
				return evictionErrors
			}
		} else {
			childNodeExecutions = nodeExecutionModel.ChildNodeExecutions
		}
		for _, childNodeExecution := range childNodeExecutions {
			evictionErrors = m.evictNodeExecutionCache(ctx, childNodeExecution, evictionErrors)
		}
	}

	taskExecutions, err := m.listAllTaskExecutions(ctx, nodeExecution.Id)
	if err != nil {
		logger.Warnf(ctx, "Failed to list task executions for node execution %+v: %v",
			nodeExecution.Id, err)
		m.metrics.CacheEvictionFailures.Inc()
		evictionErrors = append(evictionErrors, &core.CacheEvictionError{
			NodeExecutionId: nodeExecution.Id,
			Source: &core.CacheEvictionError_WorkflowExecutionId{
				WorkflowExecutionId: nodeExecution.Id.ExecutionId,
			},
			Code:    core.CacheEvictionError_INTERNAL,
			Message: "Failed to list task executions",
		})
		return evictionErrors
	}

	switch md := nodeExecution.GetClosure().GetTargetMetadata().(type) {
	case *admin.NodeExecutionClosure_TaskNodeMetadata:
		for _, taskExecutionModel := range taskExecutions {
			taskExecution, err := transformers.FromTaskExecutionModel(taskExecutionModel)
			if err != nil {
				logger.Warnf(ctx, "Failed to transform task execution model %+v: %v",
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
				return evictionErrors
			}

			evictionErrors = m.evictTaskNodeExecutionCache(ctx, nodeExecutionModel, nodeExecution, taskExecution,
				md.TaskNodeMetadata, evictionErrors)
		}
	case *admin.NodeExecutionClosure_WorkflowNodeMetadata:
		evictionErrors = m.evictWorkflowNodeExecutionCache(ctx, nodeExecution, md.WorkflowNodeMetadata, evictionErrors)
	default:
		logger.Errorf(ctx, "Invalid target metadata type %T for node execution closure %+v for node execution %+v",
			md, md, nodeExecution.Id)
		m.metrics.CacheEvictionFailures.Inc()
		evictionErrors = append(evictionErrors, &core.CacheEvictionError{
			NodeExecutionId: nodeExecution.Id,
			Source: &core.CacheEvictionError_WorkflowExecutionId{
				WorkflowExecutionId: nodeExecution.Id.ExecutionId,
			},
			Code:    core.CacheEvictionError_INTERNAL,
			Message: "Internal error",
		})
	}

	return evictionErrors
}

func (m *CacheManager) evictTaskNodeExecutionCache(ctx context.Context, nodeExecutionModel models.NodeExecution,
	nodeExecution *admin.NodeExecution, taskExecution *admin.TaskExecution,
	taskNodeMetadata *admin.TaskNodeMetadata, evictionErrors []*core.CacheEvictionError) []*core.CacheEvictionError {
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

	selectedFields := []string{shared.CreatedAt}
	nodeExecutionModel.CacheStatus = nil

	taskNodeMetadata.CacheStatus = core.CatalogCacheStatus_CACHE_DISABLED
	taskNodeMetadata.CatalogKey = nil
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
		logger.Warnf(ctx, "Failed to update node execution model %+v before deleting artifact: %v",
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

	if err := m.catalogClient.ReleaseReservationByArtifactTag(ctx, datasetID, artifactTag, ownerID); err != nil {
		logger.Warnf(ctx, "Failed to release reservation for artifact of node execution %+v, cannot evict cache",
			nodeExecution.Id)
		m.metrics.CacheEvictionFailures.Inc()
		evictionErrors = append(evictionErrors, &core.CacheEvictionError{
			NodeExecutionId: nodeExecution.Id,
			Source: &core.CacheEvictionError_TaskExecutionId{
				TaskExecutionId: taskExecution.Id,
			},
			Code:    core.CacheEvictionError_RESERVATION_NOT_RELEASED,
			Message: "Failed to release reservation for artifact",
		})
		return evictionErrors
	}

	logger.Debugf(ctx, "Successfully evicted cache for task node execution %+v", nodeExecution.Id)
	m.metrics.CacheEvictionSuccess.Inc()

	return evictionErrors
}

func (m *CacheManager) evictWorkflowNodeExecutionCache(ctx context.Context, nodeExecution *admin.NodeExecution,
	workflowNodeMetadata *admin.WorkflowNodeMetadata, evictionErrors []*core.CacheEvictionError) []*core.CacheEvictionError {
	if workflowNodeMetadata == nil {
		logger.Debugf(ctx, "Node execution %+v did not contain cached data, skipping cache eviction",
			nodeExecution.Id)
		return evictionErrors
	}

	childNodeExecutions, err := m.listAllNodeExecutionsForWorkflow(ctx, workflowNodeMetadata.GetExecutionId(), "")
	if err != nil {
		logger.Debugf(ctx, "Failed to list child executions for node execution %+v of workflow %+v: %v",
			nodeExecution.Id, workflowNodeMetadata.GetExecutionId(), err)
		m.metrics.CacheEvictionFailures.Inc()
		evictionErrors = append(evictionErrors, &core.CacheEvictionError{
			NodeExecutionId: nodeExecution.Id,
			Source: &core.CacheEvictionError_WorkflowExecutionId{
				WorkflowExecutionId: workflowNodeMetadata.GetExecutionId(),
			},
			Code:    core.CacheEvictionError_INTERNAL,
			Message: "Failed to evict child executions for workflow",
		})
		return evictionErrors
	}
	for _, childNodeExecution := range childNodeExecutions {
		evictionErrors = m.evictNodeExecutionCache(ctx, childNodeExecution, evictionErrors)
	}

	logger.Debugf(ctx, "Successfully evicted cache for workflow node execution %+v", nodeExecution.Id)
	m.metrics.CacheEvictionSuccess.Inc()

	return evictionErrors
}

func (m *CacheManager) listAllNodeExecutionsForWorkflow(ctx context.Context,
	workflowExecutionID *core.WorkflowExecutionIdentifier, uniqueParentID string) ([]models.NodeExecution, error) {
	var nodeExecutions []models.NodeExecution
	var token string
	for {
		executions, newToken, err := util.ListNodeExecutionsForWorkflow(ctx, m.db, workflowExecutionID,
			uniqueParentID, "", nodeExecutionLimit, token, sortByCreatedAtAsc)
		if err != nil {
			return nil, err
		}

		nodeExecutions = append(nodeExecutions, executions...)
		if len(newToken) == 0 {
			// empty token is returned once no more node executions are available
			break
		}
		token = newToken
	}

	return nodeExecutions, nil
}

func (m *CacheManager) listAllTaskExecutions(ctx context.Context, nodeExecutionID *core.NodeExecutionIdentifier) ([]models.TaskExecution, error) {
	var taskExecutions []models.TaskExecution
	var token string
	for {
		executions, newToken, err := util.ListTaskExecutions(ctx, m.db, nodeExecutionID, "",
			taskExecutionLimit, token, sortByCreatedAtAsc)
		if err != nil {
			return nil, err
		}

		taskExecutions = append(taskExecutions, executions...)
		if len(newToken) == 0 {
			// empty token is returned once no more task executions are available
			break
		}
		token = newToken
	}

	return taskExecutions, nil
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
