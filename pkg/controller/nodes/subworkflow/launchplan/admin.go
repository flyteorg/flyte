package launchplan

import (
	"context"
	"fmt"
	"runtime/pprof"
	"time"

	"github.com/lyft/flytestdlib/logger"

	"github.com/lyft/flytestdlib/contextutils"

	"github.com/lyft/flytestdlib/promutils"

	"github.com/lyft/flytestdlib/utils"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Executor for Launchplans that executes on a remote FlyteAdmin service (if configured)
type adminLaunchPlanExecutor struct {
	adminClient service.AdminServiceClient
	cache       utils.AutoRefreshCache
}

type executionCacheItem struct {
	core.WorkflowExecutionIdentifier
	ExecutionClosure *admin.ExecutionClosure
	SyncError        error
}

func (e executionCacheItem) ID() string {
	return e.String()
}

func (a *adminLaunchPlanExecutor) Launch(ctx context.Context, launchCtx LaunchContext, executionID *core.WorkflowExecutionIdentifier, launchPlanRef *core.Identifier, inputs *core.LiteralMap) error {
	req := &admin.ExecutionCreateRequest{
		Project: executionID.Project,
		Domain:  executionID.Domain,
		Name:    executionID.Name,
		Spec: &admin.ExecutionSpec{
			LaunchPlan: launchPlanRef,
			Metadata: &admin.ExecutionMetadata{
				Mode:                admin.ExecutionMetadata_SYSTEM,
				Nesting:             launchCtx.NestingLevel + 1,
				Principal:           launchCtx.Principal,
				ParentNodeExecution: launchCtx.ParentNodeExecution,
			},
			Inputs: inputs,
		},
	}
	_, err := a.adminClient.CreateExecution(ctx, req)
	if err != nil {
		statusCode := status.Code(err)
		switch statusCode {
		case codes.AlreadyExists:
			_, err := a.cache.GetOrCreate(executionCacheItem{WorkflowExecutionIdentifier: *executionID})
			if err != nil {
				logger.Errorf(ctx, "Failed to add ExecID [%v] to auto refresh cache", executionID)
			}

			return Wrapf(RemoteErrorAlreadyExists, err, "ExecID %s already exists", executionID.Name)
		case codes.DataLoss, codes.DeadlineExceeded, codes.Internal, codes.Unknown, codes.Canceled:
			return Wrapf(RemoteErrorSystem, err, "failed to launch workflow [%s], system error", launchPlanRef.Name)
		default:
			return Wrapf(RemoteErrorUser, err, "failed to launch workflow")
		}
	}

	_, err = a.cache.GetOrCreate(executionCacheItem{WorkflowExecutionIdentifier: *executionID})
	if err != nil {
		logger.Info(ctx, "Failed to add ExecID [%v] to auto refresh cache", executionID)
	}

	return nil
}

func (a *adminLaunchPlanExecutor) GetStatus(ctx context.Context, executionID *core.WorkflowExecutionIdentifier) (*admin.ExecutionClosure, error) {
	if executionID == nil {
		return nil, fmt.Errorf("nil executionID")
	}

	obj, err := a.cache.GetOrCreate(executionCacheItem{WorkflowExecutionIdentifier: *executionID})
	if err != nil {
		return nil, err
	}

	item := obj.(executionCacheItem)

	return item.ExecutionClosure, item.SyncError
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
		return Wrapf(RemoteErrorSystem, err, "system error")
	}
	return nil
}

func (a *adminLaunchPlanExecutor) Initialize(ctx context.Context) error {
	go func() {
		// Set goroutine-label...
		ctx = contextutils.WithGoroutineLabel(ctx, "admin-launcher")
		pprof.SetGoroutineLabels(ctx)
		a.cache.Start(ctx)
	}()

	return nil
}

func (a *adminLaunchPlanExecutor) syncItem(ctx context.Context, obj utils.CacheItem) (
	newItem utils.CacheItem, result utils.CacheSyncAction, err error) {
	exec := obj.(executionCacheItem)
	req := &admin.WorkflowExecutionGetRequest{
		Id: &exec.WorkflowExecutionIdentifier,
	}

	res, err := a.adminClient.GetExecution(ctx, req)
	if err != nil {
		// TODO: Define which error codes are system errors (and return the error) vs user errors.

		if status.Code(err) == codes.NotFound {
			err = Wrapf(RemoteErrorNotFound, err, "execID [%s] not found on remote", exec.WorkflowExecutionIdentifier.Name)
		} else {
			err = Wrapf(RemoteErrorSystem, err, "system error")
		}

		return executionCacheItem{
			WorkflowExecutionIdentifier: exec.WorkflowExecutionIdentifier,
			SyncError:                   err,
		}, utils.Update, nil
	}

	return executionCacheItem{
		WorkflowExecutionIdentifier: exec.WorkflowExecutionIdentifier,
		ExecutionClosure:            res.Closure,
	}, utils.Update, nil
}

func NewAdminLaunchPlanExecutor(_ context.Context, client service.AdminServiceClient,
	syncPeriod time.Duration, cfg *AdminConfig, scope promutils.Scope) (Executor, error) {
	exec := &adminLaunchPlanExecutor{
		adminClient: client,
	}

	// TODO: make tps/burst/size configurable
	cache, err := utils.NewAutoRefreshCache(exec.syncItem, utils.NewRateLimiter("adminSync",
		float64(cfg.TPS), cfg.Burst), syncPeriod, cfg.MaxCacheSize, scope)
	if err != nil {
		return nil, err
	}

	exec.cache = cache
	return exec, nil
}
