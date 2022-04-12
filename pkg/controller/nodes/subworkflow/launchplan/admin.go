package launchplan

import (
	"context"
	"fmt"
	"time"

	"github.com/flyteorg/flytestdlib/cache"
	"golang.org/x/time/rate"
	"k8s.io/client-go/util/workqueue"

	"github.com/flyteorg/flytestdlib/errors"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flytestdlib/promutils"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var isRecovery = true

// IsWorkflowTerminated returns a true if the Workflow Phase is in a Terminal Phase, else returns a false
func IsWorkflowTerminated(p core.WorkflowExecution_Phase) bool {
	return p == core.WorkflowExecution_ABORTED || p == core.WorkflowExecution_FAILED ||
		p == core.WorkflowExecution_SUCCEEDED || p == core.WorkflowExecution_TIMED_OUT
}

// Executor for Launchplans that executes on a remote FlyteAdmin service (if configured)
type adminLaunchPlanExecutor struct {
	adminClient service.AdminServiceClient
	cache       cache.AutoRefresh
}

type executionCacheItem struct {
	core.WorkflowExecutionIdentifier
	ExecutionClosure *admin.ExecutionClosure
	SyncError        error
}

func (e executionCacheItem) ID() string {
	return e.String()
}

func (a *adminLaunchPlanExecutor) handleLaunchError(ctx context.Context, isRecovery bool,
	executionID *core.WorkflowExecutionIdentifier, launchPlanRef *core.Identifier, err error) error {

	statusCode := status.Code(err)
	if isRecovery && statusCode == codes.NotFound {
		logger.Warnf(ctx, "failed to recover workflow [%s] with err %+v. will attempt to launch instead", launchPlanRef.Name, err)
		return nil
	}
	switch statusCode {
	case codes.AlreadyExists:
		_, err := a.cache.GetOrCreate(executionID.String(), executionCacheItem{WorkflowExecutionIdentifier: *executionID})
		if err != nil {
			logger.Errorf(ctx, "Failed to add ExecID [%v] to auto refresh cache", executionID)
		}

		return errors.Wrapf(RemoteErrorAlreadyExists, err, "ExecID %s already exists", executionID.Name)
	case codes.DataLoss, codes.DeadlineExceeded, codes.Internal, codes.Unknown, codes.Canceled:
		return errors.Wrapf(RemoteErrorSystem, err, "failed to launch workflow [%s], system error", launchPlanRef.Name)
	default:
		return errors.Wrapf(RemoteErrorUser, err, "failed to launch workflow")
	}
}

func (a *adminLaunchPlanExecutor) Launch(ctx context.Context, launchCtx LaunchContext,
	executionID *core.WorkflowExecutionIdentifier, launchPlanRef *core.Identifier, inputs *core.LiteralMap) error {
	var err error
	if launchCtx.RecoveryExecution != nil {
		_, err = a.adminClient.RecoverExecution(ctx, &admin.ExecutionRecoverRequest{
			Id:   launchCtx.RecoveryExecution,
			Name: executionID.Name,
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

	req := &admin.ExecutionCreateRequest{
		Project: executionID.Project,
		Domain:  executionID.Domain,
		Name:    executionID.Name,
		Inputs:  inputs,
		Spec: &admin.ExecutionSpec{
			LaunchPlan: launchPlanRef,
			Metadata: &admin.ExecutionMetadata{
				Mode:                admin.ExecutionMetadata_CHILD_WORKFLOW,
				Nesting:             launchCtx.NestingLevel + 1,
				Principal:           launchCtx.Principal,
				ParentNodeExecution: launchCtx.ParentNodeExecution,
			},
			Labels:              &admin.Labels{Values: launchCtx.Labels},
			Annotations:         &admin.Annotations{Values: launchCtx.Annotations},
			SecurityContext:     &launchCtx.SecurityContext,
			MaxParallelism:      int32(launchCtx.MaxParallelism),
			RawOutputDataConfig: launchCtx.RawOutputDataConfig,
		},
	}

	_, err = a.adminClient.CreateExecution(ctx, req)
	if err != nil {
		launchErr := a.handleLaunchError(ctx, !isRecovery, executionID, launchPlanRef, err)
		if launchErr != nil {
			return launchErr
		}
	}

	_, err = a.cache.GetOrCreate(executionID.String(), executionCacheItem{WorkflowExecutionIdentifier: *executionID})
	if err != nil {
		logger.Info(ctx, "Failed to add ExecID [%v] to auto refresh cache", executionID)
	}

	return nil
}

func (a *adminLaunchPlanExecutor) GetStatus(ctx context.Context, executionID *core.WorkflowExecutionIdentifier) (*admin.ExecutionClosure, error) {
	if executionID == nil {
		return nil, fmt.Errorf("nil executionID")
	}

	obj, err := a.cache.GetOrCreate(executionID.String(), executionCacheItem{WorkflowExecutionIdentifier: *executionID})
	if err != nil {
		return nil, err
	}

	item := obj.(executionCacheItem)

	return item.ExecutionClosure, item.SyncError
}

func (a *adminLaunchPlanExecutor) GetLaunchPlan(ctx context.Context, launchPlanRef *core.Identifier) (*admin.LaunchPlan, error) {
	if launchPlanRef == nil {
		return nil, fmt.Errorf("launch plan reference is nil")
	}
	logger.Debugf(ctx, "Retrieving launch plan %s", *launchPlanRef)
	getObjectRequest := admin.ObjectGetRequest{
		Id: launchPlanRef,
	}

	lp, err := a.adminClient.GetLaunchPlan(ctx, &getObjectRequest)
	if err != nil {
		return nil, errors.Wrapf(RemoteErrorSystem, err, "Could not fetch launch plan definition from Admin")
	}
	if lp == nil {
		return nil, errors.Wrapf(RemoteErrorNotFound, err, "No launch plan retrieved from Admin")
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
		if status.Code(err) == codes.NotFound {
			return nil
		}
		return errors.Wrapf(RemoteErrorSystem, err, "system error")
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
			if IsWorkflowTerminated(exec.ExecutionClosure.Phase) {
				logger.Debugf(ctx, "Workflow [%s] is already completed, will not fetch execution information", exec.ExecutionClosure.WorkflowId)
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
			// TODO: Define which error codes are system errors (and return the error) vs user errors.

			if status.Code(err) == codes.NotFound {
				err = errors.Wrapf(RemoteErrorNotFound, err, "execID [%s] not found on remote", exec.WorkflowExecutionIdentifier.Name)
			} else {
				err = errors.Wrapf(RemoteErrorSystem, err, "system error")
			}

			resp = append(resp, cache.ItemSyncResponse{
				ID: obj.GetID(),
				Item: executionCacheItem{
					WorkflowExecutionIdentifier: exec.WorkflowExecutionIdentifier,
					SyncError:                   err,
				},
				Action: cache.Update,
			})

			continue
		}

		// Update the cache with the retrieved status
		resp = append(resp, cache.ItemSyncResponse{
			ID: obj.GetID(),
			Item: executionCacheItem{
				WorkflowExecutionIdentifier: exec.WorkflowExecutionIdentifier,
				ExecutionClosure:            res.Closure,
			},
			Action: cache.Update,
		})
	}

	return resp, nil
}

func NewAdminLaunchPlanExecutor(_ context.Context, client service.AdminServiceClient,
	syncPeriod time.Duration, cfg *AdminConfig, scope promutils.Scope) (FlyteAdmin, error) {
	exec := &adminLaunchPlanExecutor{
		adminClient: client,
	}

	rateLimiter := &workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(cfg.TPS), cfg.Burst)}
	c, err := cache.NewAutoRefreshCache("admin-launcher", exec.syncItem, rateLimiter, syncPeriod, cfg.Workers, cfg.MaxCacheSize, scope)
	if err != nil {
		return nil, err
	}

	exec.cache = c
	return exec, nil
}
