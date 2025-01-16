package launchplan

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/time/rate"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/client-go/util/workqueue"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/watch"
	evtErr "github.com/flyteorg/flyte/flytepropeller/events/errors"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler/transformers/k8s"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flyte/flytestdlib/cache"
	stdErr "github.com/flyteorg/flyte/flytestdlib/errors"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

var isRecovery = true

var terminalExecutionPhases = map[core.WorkflowExecution_Phase]bool{
	core.WorkflowExecution_SUCCEEDED: true,
	core.WorkflowExecution_FAILED:    true,
	core.WorkflowExecution_TIMED_OUT: true,
	core.WorkflowExecution_ABORTED:   true,
}

// Executor for Launchplans that executes on a remote FlyteAdmin service (if configured)
type adminLaunchPlanExecutor struct {
	cfg                *config.Config
	adminClient        service.AdminServiceClient
	watchServiceClient service.WatchServiceClient
	cache              cache.AutoRefreshWithUpdate
	watchCfg           WatchConfig
	itemSyncs          prometheus.Counter
	watchAPIErrors     prometheus.Counter
	watchAPIEvents     prometheus.Counter
	store              *storage.DataStore
	enqueueWorkflow    v1alpha1.EnqueueWorkflow
}

type executionCacheItem struct {
	core.WorkflowExecutionIdentifier
	SyncError        error
	Phase            core.WorkflowExecution_Phase
	ExecutionError   *core.ExecutionError
	HasOutputs       bool
	ExecutionOutputs *core.LiteralMap
	ParentWorkflowID v1alpha1.WorkflowID
	UpdatedAt        time.Time
}

func (e executionCacheItem) IsTerminal() bool {
	return terminalExecutionPhases[e.Phase]
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
		_, err := a.cache.GetOrCreate(executionID.String(), executionCacheItem{WorkflowExecutionIdentifier: *executionID, UpdatedAt: time.Now()})
		if err != nil {
			logger.Errorf(ctx, "Failed to add ExecID [%v] to auto refresh cache", executionID)
		}

		return stdErr.Wrapf(RemoteErrorAlreadyExists, err, "ExecID %s already exists", executionID.Name)
	case codes.DataLoss, codes.DeadlineExceeded, codes.Internal, codes.Unknown, codes.Canceled,
		// In case the remote system is under heavy load, we should retry
		codes.ResourceExhausted, codes.Unavailable:

		return stdErr.Wrapf(RemoteErrorSystem, err, "failed to launch workflow [%s], system error", launchPlanRef.Name)
	default:
		return stdErr.Wrapf(RemoteErrorUser, err, "failed to launch workflow")
	}
}

func (a *adminLaunchPlanExecutor) Launch(ctx context.Context, launchCtx LaunchContext, executionID *core.WorkflowExecutionIdentifier,
	launchPlan v1alpha1.ExecutableLaunchPlan, inputs *core.LiteralMap, parentWorkflowID v1alpha1.WorkflowID) error {

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
			launchErr := a.handleLaunchError(ctx, isRecovery, executionID, launchPlan.GetId(), err)
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

	if a.cfg.ClusterID != "" {
		labels[k8s.ParentClusterLabel] = a.cfg.ClusterID
	}

	parentShardLabel := launchCtx.Labels[k8s.ShardKeyLabel]
	if parentShardLabel != "" {
		labels[k8s.ParentShardLabel] = parentShardLabel
	}

	req := &admin.ExecutionCreateRequest{
		Project: executionID.Project,
		Domain:  executionID.Domain,
		Name:    executionID.Name,
		Org:     executionID.Org,
		Inputs:  inputs,
		Spec: &admin.ExecutionSpec{
			LaunchPlan: launchPlan.GetId(),
			Metadata: &admin.ExecutionMetadata{
				Mode:                admin.ExecutionMetadata_CHILD_WORKFLOW,
				Nesting:             launchCtx.NestingLevel + 1,
				Principal:           launchCtx.Principal,
				ParentNodeExecution: launchCtx.ParentNodeExecution,
			},
			Labels:              &admin.Labels{Values: labels},
			Annotations:         &admin.Annotations{Values: launchCtx.Annotations},
			SecurityContext:     &launchCtx.SecurityContext,
			MaxParallelism:      int32(launchCtx.MaxParallelism),
			RawOutputDataConfig: launchCtx.RawOutputDataConfig,
			Interruptible:       interruptible,
			OverwriteCache:      launchCtx.OverwriteCache,
			Envs:                &admin.Envs{Values: environmentVariables},
			ClusterAssignment:   launchCtx.ClusterAssignment,
		},
	}

	if logger.IsLoggable(ctx, logger.DebugLevel) {
		logger.Debugf(ctx, "Creating execution [%+v]", req)
	}

	resp, err := a.adminClient.CreateExecution(ctx, req)
	if err != nil {
		launchErr := a.handleLaunchError(ctx, !isRecovery, executionID, launchPlan.GetId(), err)
		if launchErr != nil {
			return launchErr
		}
	}

	if logger.IsLoggable(ctx, logger.DebugLevel) {
		logger.Debugf(ctx, "Create execution response [%+v]", resp)
	}

	hasOutputs := len(launchPlan.GetInterface().GetOutputs().GetVariables()) > 0
	_, err = a.cache.GetOrCreate(executionID.String(), executionCacheItem{
		WorkflowExecutionIdentifier: *executionID,
		HasOutputs:                  hasOutputs,
		ParentWorkflowID:            parentWorkflowID,
		UpdatedAt:                   time.Now(),
	})

	if err != nil {
		logger.Infof(ctx, "Failed to add ExecID [%v] to auto refresh cache", executionID)
	}

	return nil
}

type ExecutionStatus struct {
	Phase   core.WorkflowExecution_Phase
	Outputs *core.LiteralMap
	Error   *core.ExecutionError
}

func (a *adminLaunchPlanExecutor) GetStatus(ctx context.Context, executionID *core.WorkflowExecutionIdentifier,
	launchPlan v1alpha1.ExecutableLaunchPlan, parentWorkflowID v1alpha1.WorkflowID) (ExecutionStatus, error) {
	if executionID == nil {
		return ExecutionStatus{}, fmt.Errorf("nil executionID")
	}

	hasOutputs := len(launchPlan.GetInterface().GetOutputs().GetVariables()) > 0
	obj, err := a.cache.GetOrCreate(executionID.String(), executionCacheItem{
		WorkflowExecutionIdentifier: *executionID,
		HasOutputs:                  hasOutputs,
		ParentWorkflowID:            parentWorkflowID,
		UpdatedAt:                   time.Now(),
	})
	if err != nil {
		return ExecutionStatus{}, err
	}

	item := obj.(executionCacheItem)

	execStatus := ExecutionStatus{
		Phase:   item.Phase,
		Outputs: item.ExecutionOutputs,
		Error:   item.ExecutionError,
	}
	return execStatus, item.SyncError
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
	err := a.cache.Start(ctx)
	if err != nil {
		return err
	}

	if a.watchCfg.Enabled {
		go a.watchExecutionStatusUpdates(ctx)
	}

	return nil
}

func (a *adminLaunchPlanExecutor) syncItem(ctx context.Context, batch cache.Batch) (
	resp []cache.ItemSyncResponse, err error) {
	resp = make([]cache.ItemSyncResponse, 0, len(batch))
	for _, obj := range batch {
		exec := obj.GetItem().(executionCacheItem)
		execID := &exec.WorkflowExecutionIdentifier

		// Is workflow already terminated, then no need to fetch information, also the item can be dropped from the cache
		if exec.IsTerminal() {
			logger.Debugf(ctx, "workflow execution [%s] is already completed, will not fetch execution information", execID)
			continue
		}

		if a.watchCfg.Enabled {
			if time.Since(exec.UpdatedAt) < a.watchCfg.FreshnessDuration.Duration {
				logger.Debugf(ctx, "skipping syncItem for [%s] because its fresh", execID)
				// don't perform single call yet, give watch execution status updates API enough time to populate cache
				// use previous value for now
				continue
			}

			logger.Debugf(ctx, "detected stale cache item [%s], calling syncItem", execID)
		}
		a.itemSyncs.Inc()

		// Workflow is not already terminated, let's check the status
		req := &admin.WorkflowExecutionGetRequest{Id: execID}
		res, err := a.adminClient.GetExecution(ctx, req)
		if err != nil {
			// TODO: Define which error codes are system errors (and return the error) vs user stdErr.
			if status.Code(err) == codes.NotFound {
				err = stdErr.Wrapf(RemoteErrorNotFound, err, "execID [%s] not found on remote", execID.Name)
			} else {
				err = stdErr.Wrapf(RemoteErrorSystem, err, "system error")
			}
		}

		closure := res.GetClosure()
		exec.Phase = closure.GetPhase()
		exec.ExecutionError = closure.GetError()
		if err == nil {
			exec.ExecutionOutputs, err = a.fetchExecOutputs(ctx, &exec, closure.GetOutputs().GetUri())
		}
		exec.SyncError = err
		// don't update exec.UpdatedAt intentionally, item should be considered fresh only when updated by Watch API

		// Update the cache with the retrieved status
		resp = append(resp, cache.ItemSyncResponse{
			ID:     obj.GetID(),
			Item:   exec,
			Action: cache.Update,
		})
	}

	// wait until all responses have been processed to enqueue parent workflows. if we do it
	// prematurely, there is a chance the parent workflow evaluates before the cache is updated.
	for _, itemSyncResponse := range resp {
		exec := itemSyncResponse.Item.(executionCacheItem)
		if exec.IsTerminal() {
			a.enqueueWorkflow(exec.ParentWorkflowID)
		}
	}

	return resp, nil
}

func (a *adminLaunchPlanExecutor) watchExecutionStatusUpdates(ctx context.Context) {
	// we must pass cluster name to propeller configs so it can request execution updates designated for this particular cluster
	logger.Infof(ctx, "starting watch execution status updates stream with cluster id %q", a.cfg.ClusterID)
	stream, err := a.watchServiceClient.WatchExecutionStatusUpdates(ctx, &watch.WatchExecutionStatusUpdatesRequest{
		Cluster: a.cfg.ClusterID,
	})
	if err != nil {
		logger.Fatalf(ctx, "failed to start execution status updates stream: %v", err)
	}
	logger.Infof(ctx, "execution status updates stream started")

	for {
		select {
		case <-ctx.Done():
			err := stream.CloseSend()
			if err != nil {
				logger.Errorf(ctx, "failed to close execution status updates stream: %v", err)
			} else {
				logger.Debugf(ctx, "closed execution status updates stream")
			}
			return

		default:
			resp, err := stream.Recv()
			if err != nil {
				logger.Errorf(ctx, "failed to receive execution status update: %v. Reconnecting with cluster id %q", err, a.cfg.ClusterID)
				a.watchAPIErrors.Inc()

				time.Sleep(a.watchCfg.ReconnectDelay.Duration)
				stream, err = a.watchServiceClient.WatchExecutionStatusUpdates(ctx, &watch.WatchExecutionStatusUpdatesRequest{
					Cluster: a.cfg.ClusterID,
				})
				if err != nil {
					logger.Fatalf(ctx, "failed to start execution status updates stream: %v", err)
				}
				continue
			}
			logger.Debugf(ctx, "received execution status update [%v]", resp)
			a.watchAPIEvents.Inc()

			execID := resp.GetId()

			item, err := a.cache.Get(execID.String())
			if err != nil {
				logger.Warnf(ctx, "failed to get child execution item from auto-refresh cache: %v", err)
				continue
			}

			if item.IsTerminal() {
				logger.Debugf(ctx, "workflow execution [%s] is already completed, will not fetch execution information", execID)
				continue
			}

			exec := item.(executionCacheItem)
			exec.Phase = resp.GetPhase()
			exec.ExecutionError = resp.GetError()
			if exec.IsTerminal() {
				exec.ExecutionOutputs, exec.SyncError = a.fetchExecOutputs(ctx, &exec, resp.GetOutputUri())
				if exec.SyncError != nil {
					logger.Warnf(ctx, "fetch exec outputs error: %v", exec.SyncError)
				}
				if len(exec.ExecutionOutputs.GetLiterals()) > 0 {
					logger.Debugf(ctx, "fetched exec outputs [%v]", exec.ExecutionOutputs)
				}
			}
			exec.UpdatedAt = time.Now()
			exists := a.cache.Update(execID.String(), exec)
			if exists {
				if exec.IsTerminal() {
					a.enqueueWorkflow(exec.ParentWorkflowID)
				}
			} else {
				logger.Warnf(ctx, "execution cache item was not found during update")
			}
		}
	}
}

func (a *adminLaunchPlanExecutor) fetchExecOutputs(ctx context.Context,
	execItem *executionCacheItem,
	outputURI string,
) (*core.LiteralMap, error) {
	outputs := &core.LiteralMap{}
	id := &execItem.WorkflowExecutionIdentifier
	phase := execItem.Phase
	hasOutputs := execItem.HasOutputs
	if phase != core.WorkflowExecution_SUCCEEDED || !hasOutputs {
		return outputs, nil
	}

	execData, err := a.adminClient.GetExecutionData(ctx, &admin.WorkflowExecutionGetDataRequest{Id: id})

	// detect both scenarios where we want to read the outputURI within propeller:
	//  (1) gRPC ResourceExhausted code indicating the output literals were too large
	//  (2) output literals were not sent. this can be either intended if flyteadmin only
	//  stores the offloaded data or a consequence of flyteadmin failing to read the
	//  outputURI because of read limit configuration values, for example `maxDownloadMBs`.
	errorPreamble := ""
	fetchOutputsFromURI := false
	if status.Code(err) == codes.ResourceExhausted {
		errorPreamble = fmt.Sprintf("unable to retrieve output literals: [%v]", err)
		fetchOutputsFromURI = true
	} else if err == nil {
		if execData.GetFullOutputs().GetLiterals() != nil {
			outputs = execData.GetFullOutputs()
		} else {
			errorPreamble = "received empty output literals, if this is unexpected consult flyteadmin `GetExecutionData` size configuration"
			fetchOutputsFromURI = true
		}
	}

	if fetchOutputsFromURI && len(outputURI) > 0 {
		err = a.store.ReadProtobuf(ctx, storage.DataReference(outputURI), outputs)
		if err != nil {
			logger.Errorf(ctx, "failed to read outputs from URI [%s] with err: %v", outputURI, err)
			err = fmt.Errorf("%s %w", errorPreamble, err)
		}
	}

	return outputs, err
}

func NewAdminLaunchPlanExecutor(_ context.Context,
	cfg *config.Config,
	client service.AdminServiceClient,
	watchServiceClient service.WatchServiceClient,
	adminCfg *AdminConfig,
	scope promutils.Scope,
	store *storage.DataStore,
	enqueueWorkflow v1alpha1.EnqueueWorkflow,
) (FlyteAdmin, error) {
	exec := &adminLaunchPlanExecutor{
		adminClient:        client,
		cfg:                cfg,
		watchServiceClient: watchServiceClient,
		store:              store,
		enqueueWorkflow:    enqueueWorkflow,
		watchCfg:           adminCfg.WatchConfig,
		watchAPIEvents:     scope.MustNewCounter("watch_api_events", "Number of Watch API events received over stream"),
		watchAPIErrors:     scope.MustNewCounter("watch_api_errors", "Watch API stream receive errors"),
		itemSyncs:          scope.MustNewCounter("cache_item_syncs", "Number of synchronous single cache item syncs. Should be zero or close to zero"),
	}

	rateLimiter := &workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(adminCfg.TPS), adminCfg.Burst)}
	c, err := cache.NewAutoRefreshCache("admin-launcher", exec.syncItem, rateLimiter, adminCfg.CacheResyncDuration.Duration, adminCfg.Workers, adminCfg.MaxCacheSize, scope)
	if err != nil {
		return nil, err
	}

	cacheWithUpdate, ok := c.(cache.AutoRefreshWithUpdate)
	if !ok {
		return nil, fmt.Errorf("expected to create auto-refresh cache with update, but got %T", c)
	}
	exec.cache = cacheWithUpdate
	return exec, nil
}
