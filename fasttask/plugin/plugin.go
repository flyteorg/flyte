package plugin

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	idlcore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	flyteerrors "github.com/flyteorg/flyte/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/logs"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/template"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/tasklog"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler/transformers/k8s"

	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"

	"github.com/unionai/flyte/fasttask/plugin/interfaces"
	"github.com/unionai/flyte/fasttask/plugin/pb"
)

const actorTaskType = "actor"
const fastTaskType = "fast-task"
const maxErrorMessageLength = 102400 // 100kb

var (
	noCapacityAvailableError  = errors.New("NoCapacityAvailable")
	statusUpdateNotFoundError = errors.New("StatusUpdateNotFound")
	taskContextNotFoundError  = errors.New("TaskContextNotFound")
	podContainerNotFoundError = errors.New("PodContainerNotFound")

	taskStartTimeTemplateVar       = tasklog.MustCreateRegex("taskStartTime")
	taskStartTimeUnixMsTemplateVar = tasklog.MustCreateRegex("taskStartTimeUnixMs")
	taskEndTimeTemplateVar         = tasklog.MustCreateRegex("taskEndTime")
	taskEndTimeUnixMsTemplateVar   = tasklog.MustCreateRegex("taskEndTimeUnixMs")
)

type SubmissionPhase int

const (
	NotSubmitted SubmissionPhase = iota
	Submitted
)

// pluginMetrics is a collection of metrics for the plugin.
type pluginMetrics struct {
	allReplicasFailed             prometheus.Counter
	statusUpdateNotFoundTimeout   prometheus.Counter
	unexpectedEnvironmentDeletion prometheus.Counter
	unexpectedWorkerDeletion      prometheus.Counter
}

// newPluginMetrics creates a new pluginMetrics with the given scope.
func newPluginMetrics(scope promutils.Scope) pluginMetrics {
	return pluginMetrics{
		allReplicasFailed:             scope.MustNewCounter("all_replicas_failed", "Count of tasks that failed due to all environment replicas failing"),
		statusUpdateNotFoundTimeout:   scope.MustNewCounter("status_update_not_found_timeout", "Count of tasks that timed out waiting for status update from worker"),
		unexpectedEnvironmentDeletion: scope.MustNewCounter("unexpected_environment_deletion", "Count of environments that were deleted unexpectedly"),
		unexpectedWorkerDeletion:      scope.MustNewCounter("unexpected_worker_deletion", "Count of workers that were deleted unexpectedly"),
	}
}

// State maintains the current status of the task execution.
type State struct {
	SubmissionPhase SubmissionPhase
	PhaseVersion    uint32
	WorkerID        string
	SubmittedAt     time.Time
	LastUpdated     time.Time
}

// Plugin is a fast task plugin that offers task execution to a worker pool.
type Plugin struct {
	cfg             *Config
	fastTaskService interfaces.FastTaskService
	metrics         pluginMetrics
	builder         interfaces.EnvironmentBuilder
	store           interfaces.EnvironmentStore
	enqueueLabels   map[string]struct{}
}

// GetID returns the unique identifier for the plugin.
func (p *Plugin) GetID() string {
	return fastTaskType
}

// GetProperties returns the properties of the plugin.
func (p *Plugin) GetProperties() core.PluginProperties {
	return core.PluginProperties{}
}

// Handle is the main entrypoint for the plugin. It will offer the task to the worker pool and
// monitor the task until completion.
func (p *Plugin) Handle(ctx context.Context, tCtx core.TaskExecutionContext) (core.Transition, error) {
	// retrieve plugin state
	pluginState := &State{}
	if _, err := tCtx.PluginStateReader().Get(pluginState); err != nil {
		return core.UnknownTransition,
			flyteerrors.Wrapf(flyteerrors.CorruptedPluginState, err, "Failed to read unmarshal custom state")
	}

	phaseInfo := core.PhaseInfoUndefined
	var err error
	switch pluginState.SubmissionPhase {
	case NotSubmitted:
		// attempt to submit the task
		pluginState, phaseInfo, err = p.trySubmitTask(ctx, tCtx, pluginState)
		if err != nil {
			logger.Errorf(ctx, "failed to submit task: %v", err)
			return core.UnknownTransition, err
		}
	case Submitted:
		// monitor task status
		pluginState, phaseInfo, err = p.monitorTask(ctx, tCtx, pluginState)
		if err != nil {
			logger.Errorf(ctx, "failed to monitor task: %v", err)
			return core.UnknownTransition, err
		}
	}

	// update plugin state
	if err := tCtx.PluginStateWriter().Put(0, pluginState); err != nil {
		return core.UnknownTransition, err
	}

	return core.DoTransition(phaseInfo), nil
}

// Abort halts the specified task execution.
func (p *Plugin) Abort(ctx context.Context, tCtx core.TaskExecutionContext) error {
	// halting an execution is handled through sending a `DELETE` to the worker, which kills any
	// active executions. this is performed in the `Finalize` function which is _always_ called
	// during any abort. if this logic changes, we will need to add a call to
	// `fastTaskService.Cleanup` to ensure proper abort here.
	return nil
}

// Finalize is called when the task execution is complete, performing any necessary cleanup.
func (p *Plugin) Finalize(ctx context.Context, tCtx core.TaskExecutionContext) error {
	// read execution environment
	taskTemplate, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		return err
	}

	executionEnv := &idlcore.ExecutionEnv{}
	if err := utils.UnmarshalStruct(taskTemplate.GetCustom(), executionEnv); err != nil {
		return flyteerrors.Wrapf(flyteerrors.BadTaskSpecification, err, "failed to unmarshal environment")
	}

	executionEnvID := buildExecutionEnvID(tCtx, executionEnv)

	// retrieve plugin state
	pluginState := &State{}
	if _, err := tCtx.PluginStateReader().Get(pluginState); err != nil {
		return flyteerrors.Wrapf(flyteerrors.CorruptedPluginState, err, "Failed to read unmarshal custom state")
	}

	taskID, err := tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedNameWith(0, 50)
	if err != nil {
		return err
	}

	return p.fastTaskService.Cleanup(ctx, taskID, executionEnvID.String(), pluginState.WorkerID)
}

// trySubmitTask attempts to submit the task to the worker pool.
func (p *Plugin) trySubmitTask(ctx context.Context, tCtx core.TaskExecutionContext, initialState *State) (*State, core.PhaseInfo, error) {
	pluginState := *initialState

	// read execution environment
	taskTemplate, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		return nil, core.PhaseInfoUndefined, err
	}

	executionEnv := &idlcore.ExecutionEnv{}
	if err := utils.UnmarshalStruct(taskTemplate.GetCustom(), executionEnv); err != nil {
		return nil, core.PhaseInfoUndefined,
			flyteerrors.Wrapf(flyteerrors.BadTaskSpecification, err, "failed to unmarshal environment")
	}

	executionEnvID := buildExecutionEnvID(tCtx, executionEnv)
	taskID, err := tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedNameWith(0, 50)
	if err != nil {
		return nil, core.PhaseInfoUndefined, err
	}

	// retrieve environment
	env, err := p.builder.GetOrCreateEnvironment(ctx, tCtx, executionEnvID, executionEnv)
	if err != nil {
		return nil, core.PhaseInfoUndefined, err
	}

	if env.State() == interfaces.INITIALIZING {
		phaseInfo := core.PhaseInfoWaitingForResourcesInfo(time.Now(), pluginState.PhaseVersion, "environment is initializing", nil)
		return &pluginState, phaseInfo, nil
	}

	if env.State() == interfaces.TOMBSTONED {
		taskInfo, err := buildTaskInfo(executionEnvID, executionEnv, "")
		if err != nil {
			return nil, core.PhaseInfoUndefined, err
		}

		phaseInfo := core.PhaseInfoSystemFailure("unknown", fmt.Sprintf("failed to create any workers for actor '%s'\n%s",
			executionEnvID.String(), env.FailureMessage()), taskInfo)
		return &pluginState, phaseInfo, nil
	}

	// submit to worker
	cmd, envVars, err := buildTaskExecutionInfo(ctx, tCtx, taskTemplate)
	if err != nil {
		return nil, core.PhaseInfoUndefined, err
	}

	// offer to environment
	ownerID := tCtx.TaskExecutionMetadata().GetOwnerID()
	taskExecID := tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetID()
	// set enqueue labels to be passed in as part of the heartbeat response
	// include namespaced execution name for v1
	enqueueLabels := map[string]string{
		k8s.WorkflowID: ownerID.String(),
	}

	for label, value := range tCtx.TaskExecutionMetadata().GetLabels() {
		if _, ok := p.enqueueLabels[label]; ok {
			enqueueLabels[label] = value
		}
	}

	worker, err := p.fastTaskService.OfferTaskToEnvironment(ctx, taskExecID.GetNodeExecutionId().GetExecutionId(), executionEnvID.String(), taskID, ownerID.Namespace, ownerID.Name, cmd, envVars, enqueueLabels)
	if err != nil {
		if errors.Is(err, noCapacityAvailableError) {
			taskInfo, err := buildTaskInfo(executionEnvID, executionEnv, "")
			if err != nil {
				return nil, core.PhaseInfoUndefined, err
			}

			// check if "all" workers have failed
			failureMessage, err := p.builder.ValidateWorkerPods(ctx, executionEnvID.String(), taskInfo)
			if err != nil {
				return nil, core.PhaseInfoUndefined, err
			} else if len(failureMessage) > 0 {
				phaseInfo := core.PhaseInfoSystemFailure("unknown", fmt.Sprintf("all workers have failed for actor '%s'\n%s",
					executionEnvID.String(), failureMessage), taskInfo)
				p.metrics.allReplicasFailed.Inc()
				return &pluginState, phaseInfo, nil
			}

			// only init scale up if this task is newly added to pending (prevents duplicate scale-ups)
			if p.fastTaskService.AddPendingOwner(executionEnvID.String(), taskID, enqueueLabels) {
				p.builder.ScaleUp(ctx, executionEnvID.String())
			}

			phaseInfo := core.PhaseInfoWaitingForResourcesInfo(time.Now(), pluginState.PhaseVersion, "no workers available", taskInfo)
			return &pluginState, phaseInfo, nil
		}

		return nil, core.PhaseInfoUndefined, err
	}

	// persist worker assignment
	now := time.Now()
	taskInfo, err := p.buildTaskInfoWithLogs(ctx, executionEnvID, executionEnv, worker.ID(),
		tCtx.TaskExecutionMetadata().GetTaskExecutionID(), taskTemplate, now, now)
	if err != nil && !errors.Is(err, podContainerNotFoundError) {
		return nil, core.PhaseInfoUndefined, err
	}

	pluginState.WorkerID = worker.ID()
	pluginState.SubmissionPhase = Submitted
	pluginState.PhaseVersion = core.DefaultPhaseVersion
	pluginState.SubmittedAt = now
	pluginState.LastUpdated = now

	phaseInfo := core.PhaseInfoQueuedWithTaskInfo(now, pluginState.PhaseVersion,
		fmt.Sprintf("task offered to worker %s", worker.ID()), taskInfo)

	logger.Debugf(ctx, "task %s submitted to worker %s", taskID, worker.ID())
	return &pluginState, phaseInfo, nil
}

// monitorTask checks the status of the task execution.
func (p *Plugin) monitorTask(ctx context.Context, tCtx core.TaskExecutionContext, initialState *State) (*State, core.PhaseInfo, error) {
	// read execution environment
	taskTemplate, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		return nil, core.PhaseInfoUndefined, err
	}

	pluginState := *initialState

	executionEnv := &idlcore.ExecutionEnv{}
	if err := utils.UnmarshalStruct(taskTemplate.GetCustom(), executionEnv); err != nil {
		return nil, core.PhaseInfoUndefined,
			flyteerrors.Wrapf(flyteerrors.BadTaskSpecification, err, "failed to unmarshal environment")
	}

	executionEnvID := buildExecutionEnvID(tCtx, executionEnv)

	// short-circuit to handle cases where environment with active tasks has all of its workers terminated unexpectedly
	// GetOrCreate ensures an orphaned environment’s spec is initialized without requiring new tasks to be submitted
	env, err := p.builder.GetOrCreateEnvironment(ctx, tCtx, executionEnvID, executionEnv)
	if err != nil {
		return nil, core.PhaseInfoUndefined, err
	}
	if env == nil {
		p.metrics.unexpectedEnvironmentDeletion.Inc()
		logger.Errorf(ctx, "environment %s not found in store for running task", executionEnvID.String())
		return &pluginState, core.PhaseInfoSystemRetryableFailure("unknown", fmt.Sprintf("Environment %s not found in store for running task", executionEnvID.String()), nil), nil
	}
	workerID := initialState.WorkerID
	if worker := env.GetWorker(workerID); worker == nil {
		p.metrics.unexpectedWorkerDeletion.Inc()
		logger.Errorf(ctx, "worker %s not found in store for running task", workerID)
		return &pluginState, core.PhaseInfoSystemRetryableFailure("unknown", fmt.Sprintf("Worker %s not found in store for running task", workerID), nil), nil
	}

	taskID, err := tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedNameWith(0, 50)
	if err != nil {
		return nil, core.PhaseInfoUndefined, err
	}
	// build task info
	taskInfo, buildTaskInfoErr := p.buildTaskInfoWithLogs(ctx, executionEnvID, executionEnv, workerID,
		tCtx.TaskExecutionMetadata().GetTaskExecutionID(), taskTemplate, initialState.SubmittedAt, time.Now())
	if buildTaskInfoErr != nil && !errors.Is(buildTaskInfoErr, podContainerNotFoundError) {
		return nil, core.PhaseInfoUndefined, buildTaskInfoErr
	}

	// check the task status
	var phaseInfo core.PhaseInfo
	taskStatus, err := p.fastTaskService.CheckStatus(ctx, taskID, executionEnvID.String(), workerID)
	if err != nil {
		if errors.Is(err, statusUpdateNotFoundError) && time.Since(pluginState.LastUpdated) > GetConfig().GracePeriodStatusNotFound.Duration {
			// if task has not been updated within the grace period we should abort
			logger.Warnf(ctx, "task status update not reported within grace period for queue %s and worker %s", executionEnvID.String(), workerID)
			p.metrics.statusUpdateNotFoundTimeout.Inc()
			phaseInfo = core.PhaseInfoSystemRetryableFailure("unknown", fmt.Sprintf("task status update not reported within grace period for queue %s and worker %s", executionEnvID.String(), pluginState.WorkerID), taskInfo)
			return &pluginState, phaseInfo, nil
		} else if errors.Is(err, statusUpdateNotFoundError) || errors.Is(err, taskContextNotFoundError) {
			phaseInfo = core.PhaseInfoRunning(pluginState.PhaseVersion, taskInfo)
		} else {
			return nil, core.PhaseInfoUndefined, err
		}
	} else {
		pluginState.LastUpdated = time.Now()
		switch taskStatus.Phase {
		case core.PhaseSuccess:
			// gather outputs or errors
			outputReader := ioutils.NewRemoteFileOutputReader(ctx, tCtx.DataStore(), tCtx.OutputWriter(), 0)
			err = tCtx.OutputWriter().Put(ctx, outputReader)
			if err != nil {
				return nil, core.PhaseInfoUndefined, err
			}

			phaseInfo = core.PhaseInfoSuccess(taskInfo)
		case core.PhaseRetryableFailure:
			phaseInfo = core.PhaseInfoRetryableFailure("unknown", taskStatus.Reason, taskInfo)
		default:
			pluginState.PhaseVersion++
			phaseInfo = core.PhaseInfoRunning(pluginState.PhaseVersion, taskInfo)
		}
	}

	// if we could not find the pod / container (and the task is not attempting to failover to
	// another replica) we should retry for terminal phases to ensure that logs are correctly
	// populated before completing the task.
	if errors.Is(buildTaskInfoErr, podContainerNotFoundError) && !errors.Is(err, statusUpdateNotFoundError) &&
		(phaseInfo.Phase() == core.PhaseSuccess || phaseInfo.Phase() == core.PhaseRetryableFailure) {

		return nil, core.PhaseInfoUndefined, buildTaskInfoErr
	}

	return &pluginState, phaseInfo, nil
}

func (p *Plugin) buildTaskInfoWithLogs(ctx context.Context, executionEnvID interfaces.ExecutionEnvID,
	executionEnv *idlcore.ExecutionEnv, workerID string, taskExecutionID core.TaskExecutionID,
	taskTemplate *idlcore.TaskTemplate, start, end time.Time) (*core.TaskInfo, error) {

	// build task info
	taskInfo, err := buildTaskInfo(executionEnvID, executionEnv, workerID)
	if err != nil {
		return nil, err
	}

	// get worker pod
	pod, err := p.builder.GetWorkerPod(ctx, executionEnvID.String(), workerID)
	if err != nil {
		if k8serrors.IsNotFound(err) || k8serrors.IsGone(err) || k8serrors.IsResourceExpired(err) {
			return taskInfo, podContainerNotFoundError
		}

		return nil, err
	}

	containerIndex := -1
	for i := range pod.Spec.Containers {
		if pod.Spec.Containers[i].Name == pod.GetName() {
			containerIndex = i
			break
		}
	}
	if containerIndex == -1 {
		return taskInfo, podContainerNotFoundError
	}

	taskInfo.LogContext = &idlcore.LogContext{
		PrimaryPodName: pod.Name,
		Pods: []*idlcore.PodLogContext{
			{
				Namespace:            pod.Namespace,
				PodName:              pod.Name,
				PrimaryContainerName: pod.Spec.Containers[containerIndex].Name,
				Containers: []*idlcore.ContainerContext{
					{
						ContainerName: pod.Spec.Containers[containerIndex].Name,
						Process: &idlcore.ContainerContext_ProcessContext{
							ContainerStartTime: timestamppb.New(start),
							ContainerEndTime:   timestamppb.New(end),
						},
					},
				},
			},
		},
	}

	if len(pod.Status.ContainerStatuses) <= containerIndex {
		// no container id yet
		return taskInfo, podContainerNotFoundError
	}

	enableVscode := false
	for _, env := range pod.Spec.Containers[containerIndex].Env {
		if env.Name != logs.FlyteEnableVscode {
			continue
		}
		var err error
		enableVscode, err = strconv.ParseBool(env.Value)
		if err != nil {
			logger.Errorf(ctx, "failed to parse %s env var [%s] for pod [%s]", logs.FlyteEnableVscode, env.Value, pod.Name)
		}
	}

	in := tasklog.Input{
		Namespace:       pod.Namespace,
		PodName:         pod.Name,
		PodUID:          string(pod.GetUID()),
		ContainerName:   pod.Spec.Containers[containerIndex].Name,
		ContainerID:     pod.Status.ContainerStatuses[containerIndex].ContainerID,
		HostName:        pod.Spec.Hostname,
		TaskExecutionID: taskExecutionID,
		TaskTemplate:    taskTemplate,
		ExtraTemplateVars: []tasklog.TemplateVar{
			{
				Regex: taskStartTimeTemplateVar,
				Value: start.Format(time.RFC3339),
			},
			{
				Regex: taskStartTimeUnixMsTemplateVar,
				Value: strconv.FormatInt(start.UnixMilli(), 10),
			},
			{
				Regex: taskEndTimeTemplateVar,
				Value: end.Format(time.RFC3339),
			},
			{
				Regex: taskEndTimeUnixMsTemplateVar,
				Value: strconv.FormatInt(end.UnixMilli(), 10),
			},
		},
		EnableVscode: enableVscode,
	}
	logPlugin, err := logs.InitializeLogPlugins(&p.cfg.Logs)
	if err != nil {
		return nil, err
	}

	logs, err := logPlugin.GetTaskLogs(in)
	if err != nil {
		return nil, err
	}

	taskInfo.Logs = logs.TaskLogs
	return taskInfo, nil
}

func buildTaskExecutionInfo(ctx context.Context, tCtx core.TaskExecutionContext,
	taskTemplate *idlcore.TaskTemplate) ([]string, map[string]string, error) {

	taskContainer := taskTemplate.GetContainer()
	taskK8Pods := taskTemplate.GetK8SPod()

	var args []string
	if taskContainer != nil {
		args = taskContainer.GetArgs()
	} else if taskK8Pods != nil {
		var podSpec = v1.PodSpec{}
		err := utils.UnmarshalStructToObj(taskK8Pods.GetPodSpec(), &podSpec)
		if err != nil {
			return nil, nil, flyteerrors.Errorf(flyteerrors.BadTaskSpecification, "unable to read pod template")
		}

		var primaryContainerName string
		var ok bool
		if primaryContainerName, ok = taskTemplate.GetConfig()[flytek8s.PrimaryContainerKey]; !ok {
			return nil, nil, flyteerrors.Errorf(flyteerrors.BadTaskSpecification, "failed to find environment")
		}

		foundContainer := false
		for _, container := range podSpec.Containers {
			if container.Name == primaryContainerName {
				args = container.Args
				foundContainer = true
				break
			}
		}
		if !foundContainer {
			return nil, nil, flyteerrors.Errorf(flyteerrors.BadTaskSpecification, "unable to find primary container")
		}

	} else {
		return nil, nil, flyteerrors.Errorf(flyteerrors.BadTaskSpecification, "unable to create container with no definition in TaskTemplate")
	}

	templateParameters := template.Parameters{
		TaskExecMetadata: tCtx.TaskExecutionMetadata(),
		Inputs:           tCtx.InputReader(),
		OutputPath:       tCtx.OutputWriter(),
		Task:             tCtx.TaskReader(),
	}
	command, err := template.Render(ctx, args, templateParameters)
	if err != nil {
		return nil, nil, err
	}

	// compile environment variables
	envVars := make(map[string]string)
	for k, v := range tCtx.TaskExecutionMetadata().GetEnvironmentVariables() {
		envVars[k] = v
	}
	for _, envVar := range flytek8s.GetContextEnvVars(ctx) {
		envVars[envVar.Name] = envVar.Value
	}
	for _, envVar := range flytek8s.GetExecutionEnvVars(tCtx.TaskExecutionMetadata().GetTaskExecutionID(), "") {
		envVars[envVar.Name] = envVar.Value
	}

	return command, envVars, nil
}

// init registers the plugin with the plugin machinery.
func init() {
	pluginmachinery.PluginRegistry().RegisterCorePlugin(
		core.PluginEntry{
			ID:                  fastTaskType,
			RegisteredTaskTypes: []core.TaskType{actorTaskType, fastTaskType},
			LoadPlugin: func(ctx context.Context, iCtx core.SetupContext) (core.Plugin, error) {
				pluginMetrics := newPluginMetrics(iCtx.MetricsScope())

				cfg := GetConfig()

				// create environment store and builder
				store := newEnvironmentStore()
				envBuilderScope := iCtx.MetricsScope().NewSubScope("env_builder")
				builder := newEnvironmentBuilder(ctx, iCtx.KubeClient(), store, envBuilderScope.NewSubScope("fast_task"), cfg)

				// open tcp listener
				listener, err := net.Listen("tcp", GetConfig().Endpoint)
				if err != nil {
					return nil, err
				}

				// create and start grpc server
				fastTaskService := newFastTaskService(iCtx.EnqueueOwner(), builder, store, iCtx.MetricsScope(), cfg)
				go func() {
					grpcServer := grpc.NewServer()
					pb.RegisterFastTaskServer(grpcServer, fastTaskService)
					if err := grpcServer.Serve(listener); err != nil {
						panic("failed to start grpc fast task grpc server")
					}
				}()

				enqueueLabels := make(map[string]struct{})
				for _, label := range iCtx.IncludeEnqueueLabels() {
					enqueueLabels[label] = struct{}{}
				}

				return &Plugin{
					cfg:             GetConfig(),
					metrics:         pluginMetrics,
					builder:         builder,
					fastTaskService: fastTaskService,
					store:           store,
					enqueueLabels:   enqueueLabels,
				}, nil
			},
			IsDefault: false,
		},
	)
}
