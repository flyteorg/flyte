package launchplan

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes/wrappers"
	"golang.org/x/time/rate"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/client-go/util/workqueue"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	evtErr "github.com/flyteorg/flyte/flytepropeller/events/errors"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler/transformers/k8s"
	"github.com/flyteorg/flyte/flytestdlib/cache"
	stdErr "github.com/flyteorg/flyte/flytestdlib/errors"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

var isRecovery = true

// IsWorkflowTerminated returns a true if the Workflow Phase is in a Terminal Phase, else returns a false
func IsWorkflowTerminated(p core.WorkflowExecution_Phase) bool {
	return p == core.WorkflowExecution_ABORTED || p == core.WorkflowExecution_FAILED ||
		p == core.WorkflowExecution_SUCCEEDED || p == core.WorkflowExecution_TIMED_OUT
}

// Executor for Launchplans that executes on a remote FlyteAdmin service (if configured)
type adminLaunchPlanExecutor struct {
	adminClient     service.AdminServiceClient
	cache           cache.AutoRefresh
	store           *storage.DataStore
	enqueueWorkflow v1alpha1.EnqueueWorkflow
}

type executionCacheItem struct {
	core.WorkflowExecutionIdentifier
	ExecutionClosure *admin.ExecutionClosure
	SyncError        error
	ExecutionOutputs *core.LiteralMap
	ParentWorkflowID v1alpha1.WorkflowID
}

func (e executionCacheItem) IsTerminal() bool {
	if e.ExecutionClosure == nil {
		return false
	}
	return e.ExecutionClosure.GetPhase() == core.WorkflowExecution_ABORTED || e.ExecutionClosure.GetPhase() == core.WorkflowExecution_FAILED || e.ExecutionClosure.GetPhase() == core.WorkflowExecution_SUCCEEDED
}

func (e executionCacheItem) ID() string {
	return e.String()
}

func (a *adminLaunchPlanExecutor) handleLaunchError(ctx context.Context, isRecovery bool,
	executionID *core.WorkflowExecutionIdentifier, launchPlanRef *core.Identifier, err error) error {

	statusCode := status.Code(err)
	if isRecovery && statusCode == codes.NotFound {
		logger.Warnf(ctx, "failed to recover workflow [%s] with err %+v. will attempt to launch instead", launchPlanRef.GetName(), err)
		return nil
	}
	switch statusCode {
	case codes.AlreadyExists:
		_, err := a.cache.GetOrCreate(executionID.String(), executionCacheItem{WorkflowExecutionIdentifier: *executionID})
		if err != nil {
			logger.Errorf(ctx, "Failed to add ExecID [%v] to auto refresh cache", executionID)
		}

		return stdErr.Wrapf(RemoteErrorAlreadyExists, err, "ExecID %s already exists", executionID.GetName())
	case codes.DataLoss, codes.DeadlineExceeded, codes.Internal, codes.Unknown, codes.Canceled:
		return stdErr.Wrapf(RemoteErrorSystem, err, "failed to launch workflow [%s], system error", launchPlanRef.GetName())
	default:
		return stdErr.Wrapf(RemoteErrorUser, err, "failed to launch workflow")
	}
}

func (a *adminLaunchPlanExecutor) Launch(ctx context.Context, launchCtx LaunchContext, executionID *core.WorkflowExecutionIdentifier,
	launchPlanRef *core.Identifier, inputs *core.LiteralMap, parentWorkflowID v1alpha1.WorkflowID) error {

	var err error
	if launchCtx.RecoveryExecution != nil {
		_, err = a.adminClient.RecoverExecution(ctx, &admin.ExecutionRecoverRequest{
			Id:   launchCtx.RecoveryExecution,
			Name: executionID.GetName(),
			Metadata: &admin.ExecutionMetadata{
				ParentNodeExecution: launchCtx.ParentNodeExecution,
			},
		})
		if err != nil {
			launchErr := a.handleLaunchError(ctx, isRecovery, executionID, launchPlanRef, err)
			if launchErr != nil {
				return launchErr
			}
		} else {
			return nil
		}
	}

	var interruptible *wrappers.BoolValue
	if launchCtx.Interruptible != nil {
		interruptible = &wrappers.BoolValue{
			Value: *launchCtx.Interruptible,
		}
	}

	environmentVariables := make([]*core.KeyValuePair, 0, len(launchCtx.EnvironmentVariables))
	for k, v := range launchCtx.EnvironmentVariables {
		environmentVariables = append(environmentVariables, &core.KeyValuePair{
			Key:   k,
			Value: v,
		})
	}

	// Make a copy of the labels with shard-key removed. This ensures that the shard-key is re-computed for each
	// instead of being copied from the parent.
	labels := make(map[string]string)
	for key, value := range launchCtx.Labels {
		if key != k8s.ShardKeyLabel {
			labels[key] = value
		}
	}

	req := &admin.ExecutionCreateRequest{
		Project: executionID.GetProject(),
		Domain:  executionID.GetDomain(),
		Name:    executionID.GetName(),
		Inputs:  inputs,
		Spec: &admin.ExecutionSpec{
			LaunchPlan: launchPlanRef,
			Metadata: &admin.ExecutionMetadata{
				Mode:                admin.ExecutionMetadata_CHILD_WORKFLOW,
				Nesting:             launchCtx.NestingLevel + 1,
				Principal:           launchCtx.Principal,
				ParentNodeExecution: launchCtx.ParentNodeExecution,
			},
			Labels:              &admin.Labels{Values: labels},
			Annotations:         &admin.Annotations{Values: launchCtx.Annotations},
			SecurityContext:     &launchCtx.SecurityContext,
			MaxParallelism:      int32(launchCtx.MaxParallelism), // #nosec G115
			RawOutputDataConfig: launchCtx.RawOutputDataConfig,
			Interruptible:       interruptible,
			OverwriteCache:      launchCtx.OverwriteCache,
			Envs:                &admin.Envs{Values: environmentVariables},
		},
	}

	_, err = a.adminClient.CreateExecution(ctx, req)
	if err != nil {
		launchErr := a.handleLaunchError(ctx, !isRecovery, executionID, launchPlanRef, err)
		if launchErr != nil {
			return launchErr
		}
	}

	_, err = a.cache.GetOrCreate(executionID.String(), executionCacheItem{WorkflowExecutionIdentifier: *executionID, ParentWorkflowID: parentWorkflowID})
	if err != nil {
		logger.Infof(ctx, "Failed to add ExecID [%v] to auto refresh cache", executionID)
	}

	return nil
}

func (a *adminLaunchPlanExecutor) GetStatus(ctx context.Context, executionID *core.WorkflowExecutionIdentifier) (*admin.ExecutionClosure, *core.LiteralMap, error) {
	if executionID == nil {
		return nil, nil, fmt.Errorf("nil executionID")
	}

	obj, err := a.cache.GetOrCreate(executionID.String(), executionCacheItem{WorkflowExecutionIdentifier: *executionID})
	if err != nil {
		return nil, nil, err
	}

	item := obj.(executionCacheItem)

	return item.ExecutionClosure, item.ExecutionOutputs, item.SyncError
}

func (a *adminLaunchPlanExecutor) GetLaunchPlan(ctx context.Context, launchPlanRef *core.Identifier) (*admin.LaunchPlan, error) {
	if launchPlanRef == nil {
		return nil, fmt.Errorf("launch plan reference is nil")
	}
	logger.Debugf(ctx, "Retrieving launch plan %v", *launchPlanRef)
	getObjectRequest := admin.ObjectGetRequest{
		Id: launchPlanRef,
	}

	lp, err := a.adminClient.GetLaunchPlan(ctx, &getObjectRequest)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, stdErr.Wrapf(RemoteErrorNotFound, err, "No launch plan retrieved from Admin")
		}
		return nil, stdErr.Wrapf(RemoteErrorSystem, err, "Could not fetch launch plan definition from Admin")
	}

	return lp, nil
}

func (a *adminLaunchPlanExecutor) Kill(ctx context.Context, executionID *core.WorkflowExecutionIdentifier, reason string) error {
	req := &admin.ExecutionTerminateRequest{
		Id:    executionID,
		Cause: reason,
	}
	_, err := a.adminClient.TerminateExecution(ctx, req)
	if err != nil {
		err := evtErr.WrapError(err)
		if evtErr.IsNotFound(err) {
			return nil
		}

		if evtErr.IsEventAlreadyInTerminalStateError(err) {
			return nil
		}

		return stdErr.Wrapf(RemoteErrorSystem, err, "system error")
	}
	return nil
}

func (a *adminLaunchPlanExecutor) Initialize(ctx context.Context) error {
	return a.cache.Start(ctx)
}

func (a *adminLaunchPlanExecutor) syncItem(ctx context.Context, batch cache.Batch) (
	resp []cache.ItemSyncResponse, err error) {
	resp = make([]cache.ItemSyncResponse, 0, len(batch))
	for _, obj := range batch {
		exec := obj.GetItem().(executionCacheItem)

		// Is workflow already terminated, then no need to fetch information, also the item can be dropped from the cache
		if exec.ExecutionClosure != nil {
			if IsWorkflowTerminated(exec.ExecutionClosure.GetPhase()) {
				logger.Debugf(ctx, "Workflow [%s] is already completed, will not fetch execution information", exec.ExecutionClosure.GetWorkflowId())
				resp = append(resp, cache.ItemSyncResponse{
					ID:     obj.GetID(),
					Item:   exec,
					Action: cache.Unchanged,
				})
				continue
			}
		}

		// Workflow is not already terminated, lets check the status
		req := &admin.WorkflowExecutionGetRequest{
			Id: &exec.WorkflowExecutionIdentifier,
		}

		res, err := a.adminClient.GetExecution(ctx, req)
		if err != nil {
			// TODO: Define which error codes are system errors (and return the error) vs user stdErr.

			if status.Code(err) == codes.NotFound {
				err = stdErr.Wrapf(RemoteErrorNotFound, err, "execID [%s] not found on remote", exec.WorkflowExecutionIdentifier.GetName())
			} else {
				err = stdErr.Wrapf(RemoteErrorSystem, err, "system error")
			}

			resp = append(resp, cache.ItemSyncResponse{
				ID: obj.GetID(),
				Item: executionCacheItem{
					WorkflowExecutionIdentifier: exec.WorkflowExecutionIdentifier,
					SyncError:                   err,
					ParentWorkflowID:            exec.ParentWorkflowID,
				},
				Action: cache.Update,
			})

			continue
		}

		var outputs = &core.LiteralMap{}
		// Retrieve potential outputs only when the workflow succeeded.
		// TODO: We can optimize further by only retrieving the outputs when the workflow has output variables in the
		// 	interface.
		if res.GetClosure().GetPhase() == core.WorkflowExecution_SUCCEEDED {
			execData, err := a.adminClient.GetExecutionData(ctx, &admin.WorkflowExecutionGetDataRequest{
				Id: &exec.WorkflowExecutionIdentifier,
			})
			if err != nil || execData.GetFullOutputs() == nil || execData.GetFullOutputs().GetLiterals() == nil {
				outputURI := res.GetClosure().GetOutputs().GetUri()
				// attempt remote storage read on GetExecutionData failure
				if outputURI != "" {
					err = a.store.ReadProtobuf(ctx, storage.DataReference(outputURI), outputs)
					if err != nil {
						logger.Errorf(ctx, "Failed to read outputs from URI [%s] with err: %v", outputURI, err)
					}
				}
				if err != nil {
					resp = append(resp, cache.ItemSyncResponse{
						ID: obj.GetID(),
						Item: executionCacheItem{
							WorkflowExecutionIdentifier: exec.WorkflowExecutionIdentifier,
							SyncError:                   err,
							ParentWorkflowID:            exec.ParentWorkflowID,
						},
						Action: cache.Update,
					})

					continue
				}

			} else {
				outputs = execData.GetFullOutputs()
			}
		}

		// Update the cache with the retrieved status
		resp = append(resp, cache.ItemSyncResponse{
			ID: obj.GetID(),
			Item: executionCacheItem{
				WorkflowExecutionIdentifier: exec.WorkflowExecutionIdentifier,
				ExecutionClosure:            res.GetClosure(),
				ExecutionOutputs:            outputs,
				ParentWorkflowID:            exec.ParentWorkflowID,
			},
			Action: cache.Update,
		})
	}

	// wait until all responses have been processed to enqueue parent workflows. if we do it
	// prematurely, there is a chance the parent workflow evaluates before the cache is updated.
	for _, itemSyncResponse := range resp {
		exec := itemSyncResponse.Item.(executionCacheItem)
		if exec.ExecutionClosure != nil && IsWorkflowTerminated(exec.ExecutionClosure.GetPhase()) {
			a.enqueueWorkflow(exec.ParentWorkflowID)
		}
	}

	return resp, nil
}

func NewAdminLaunchPlanExecutor(_ context.Context, client service.AdminServiceClient,
	cfg *AdminConfig, scope promutils.Scope, store *storage.DataStore, enqueueWorkflow v1alpha1.EnqueueWorkflow) (FlyteAdmin, error) {
	exec := &adminLaunchPlanExecutor{
		adminClient:     client,
		store:           store,
		enqueueWorkflow: enqueueWorkflow,
	}

	rateLimiter := &workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(cfg.TPS), cfg.Burst)}
	// #nosec G115
	c, err := cache.NewAutoRefreshCache("admin-launcher", exec.syncItem, rateLimiter, cfg.CacheResyncDuration.Duration, uint(cfg.Workers), uint(cfg.MaxCacheSize), scope)
	if err != nil {
		return nil, err
	}

	exec.cache = c
	return exec, nil
}
