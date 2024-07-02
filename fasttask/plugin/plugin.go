package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/samber/lo"
	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	idlcore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	flyteerrors "github.com/flyteorg/flyte/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/template"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils/secrets"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/array/errorcollector"
	podplugin "github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/k8s/pod"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler/transformers/k8s"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"

	"github.com/unionai/flyte/fasttask/plugin/pb"
)

const fastTaskType = "fast-task"
const maxErrorMessageLength = 102400 // 100kb

var (
	statusUpdateNotFoundError = errors.New("StatusUpdateNotFound")
	taskContextNotFoundError  = errors.New("TaskContextNotFound")
)

type SubmissionPhase int

const (
	NotSubmitted SubmissionPhase = iota
	Submitted
)

// pluginMetrics is a collection of metrics for the plugin.
type pluginMetrics struct {
	allReplicasFailed           prometheus.Counter
	statusUpdateNotFoundTimeout prometheus.Counter
}

// newPluginMetrics creates a new pluginMetrics with the given scope.
func newPluginMetrics(scope promutils.Scope) pluginMetrics {
	return pluginMetrics{
		allReplicasFailed:           scope.MustNewCounter("all_replicas_failed", "Count of tasks that failed due to all environment replicas failing"),
		statusUpdateNotFoundTimeout: scope.MustNewCounter("status_update_not_found_timeout", "Count of tasks that timed out waiting for status update from worker"),
	}
}

// State maintains the current status of the task execution.
type State struct {
	SubmissionPhase SubmissionPhase
	PhaseVersion    uint32
	WorkerID        string
	LastUpdated     time.Time
}

// Plugin is a fast task plugin that offers task execution to a worker pool.
type Plugin struct {
	fastTaskService FastTaskService
	metrics         pluginMetrics
}

// GetID returns the unique identifier for the plugin.
func (p *Plugin) GetID() string {
	return fastTaskType
}

// GetProperties returns the properties of the plugin.
func (p *Plugin) GetProperties() core.PluginProperties {
	return core.PluginProperties{}
}

// buildExecutionEnvID creates an `ExecutionEnvID` from a task ID and an `ExecutionEnv`. This
// collection of attributes is used to uniquely identify an execution environment.
func buildExecutionEnvID(taskID *idlcore.Identifier, executionEnv *idlcore.ExecutionEnv) core.ExecutionEnvID {
	return core.ExecutionEnvID{
		Org:     taskID.GetOrg(),
		Project: taskID.GetProject(),
		Domain:  taskID.GetDomain(),
		Name:    executionEnv.GetName(),
		Version: executionEnv.GetVersion(),
	}
}

// getExecutionEnv retrieves the execution environment for the task. If the environment does not
// exist, it will create it.
// this is here because we wanted uniformity within `TaskExecutionContext` where functions simply
// return an interface rather than doing any actual work. alternatively, we could bury this within
// `NodeExecutionContext` so other `ExecutionEnvironment` plugins do not need to duplicate this.
func (p *Plugin) getExecutionEnv(ctx context.Context, tCtx core.TaskExecutionContext) (*idlcore.ExecutionEnv, *pb.FastTaskEnvironment, error) {
	taskTemplate, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		return nil, nil, err
	}

	executionEnv := &idlcore.ExecutionEnv{}
	if err := utils.UnmarshalStruct(taskTemplate.GetCustom(), executionEnv); err != nil {
		return nil, nil, flyteerrors.Wrapf(flyteerrors.BadTaskSpecification, err, "failed to unmarshal environment")
	}

	switch e := executionEnv.GetEnvironment().(type) {
	case *idlcore.ExecutionEnv_Spec:
		executionEnvClient := tCtx.GetExecutionEnvClient()
		taskExecutionID := tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetID()
		executionEnvID := buildExecutionEnvID(taskExecutionID.GetTaskId(), executionEnv)

		// if environment already exists then return it
		if environment := executionEnvClient.Get(ctx, executionEnvID); environment != nil {
			fastTaskEnvironment := &pb.FastTaskEnvironment{}
			if err := utils.UnmarshalStruct(environment, fastTaskEnvironment); err != nil {
				return nil, nil, flyteerrors.Wrapf(flyteerrors.BadTaskSpecification, err, "failed to unmarshal environment client")
			}

			return executionEnv, fastTaskEnvironment, nil
		}

		// otherwise create the environment
		fastTaskEnvironment, err := p.createExecutionEnv(ctx, tCtx, executionEnvID, e)
		if err != nil {
			return nil, nil, err
		}

		return executionEnv, fastTaskEnvironment, nil
	case *idlcore.ExecutionEnv_Extant:
		fastTaskEnvironment := &pb.FastTaskEnvironment{}
		if err := utils.UnmarshalStruct(e.Extant, fastTaskEnvironment); err != nil {
			return nil, nil, flyteerrors.Wrapf(flyteerrors.BadTaskSpecification, err, "failed to unmarshal environment extant")
		}

		return executionEnv, fastTaskEnvironment, nil
	}

	return nil, nil, nil
}

// createExecutionEnv creates a new execution environment based on the specified parameters.
func (p *Plugin) createExecutionEnv(ctx context.Context, tCtx core.TaskExecutionContext,
	executionEnvID core.ExecutionEnvID, envSpec *idlcore.ExecutionEnv_Spec) (*pb.FastTaskEnvironment, error) {

	environmentSpec := envSpec.Spec

	fastTaskEnvironmentSpec := &pb.FastTaskEnvironmentSpec{}
	if err := utils.UnmarshalStruct(environmentSpec, fastTaskEnvironmentSpec); err != nil {
		return nil, flyteerrors.Wrapf(flyteerrors.BadTaskSpecification, err, "failed to unmarshal environment spec")
	}
	var podTemplateSpec v1.PodTemplateSpec
	if len(fastTaskEnvironmentSpec.GetPodTemplateSpec()) > 0 {
		if err := json.Unmarshal(fastTaskEnvironmentSpec.GetPodTemplateSpec(), &podTemplateSpec); err != nil {
			return nil, flyteerrors.Wrapf(flyteerrors.BadTaskSpecification, err, "failed to unmarshal pod template spec")
		}
	} else {
		podSpec, objectMeta, primaryContainerName, err := flytek8s.ToK8sPodSpec(ctx, tCtx)
		if err != nil {
			return nil, flyteerrors.Wrapf(flyteerrors.BadTaskSpecification, err, "failed to create environment")
		}

		podTemplateSpec = v1.PodTemplateSpec{
			ObjectMeta: *objectMeta,
			Spec:       *podSpec,
		}
		fastTaskEnvironmentSpec.PrimaryContainerName = primaryContainerName
	}
	if err := p.addObjectMetadata(ctx, tCtx, &podTemplateSpec, config.GetK8sPluginConfig()); err != nil {
		return nil, flyteerrors.Wrapf(flyteerrors.BadTaskSpecification, err, "failed to add object metadata")
	}

	// need to marshal as JSON to maintain container resources, proto serialization does
	// not persist these settings for `PodSpec`
	podTemplateSpecBytes, err := json.Marshal(podTemplateSpec)
	if err != nil {
		return nil, flyteerrors.Wrapf(flyteerrors.BadTaskSpecification, err, "failed to marshal pod template spec")
	}

	fastTaskEnvironmentSpec.PodTemplateSpec = podTemplateSpecBytes
	if err := utils.MarshalStruct(fastTaskEnvironmentSpec, environmentSpec); err != nil {
		return nil, flyteerrors.Wrapf(flyteerrors.BadTaskSpecification, err, "failed to marshal environment spec")
	}

	executionEnvClient := tCtx.GetExecutionEnvClient()
	environment, err := executionEnvClient.Create(ctx, executionEnvID, environmentSpec)
	if err != nil {
		return nil, flyteerrors.Wrapf(flyteerrors.BadTaskSpecification, err, "failed to create environment")
	}

	fastTaskEnvironment := &pb.FastTaskEnvironment{}
	if err := utils.UnmarshalStruct(environment, fastTaskEnvironment); err != nil {
		return nil, flyteerrors.Wrapf(flyteerrors.BadTaskSpecification, err, "failed to unmarshal environment extant")
	}

	return fastTaskEnvironment, nil
}

func (p *Plugin) addObjectMetadata(ctx context.Context, tCtx core.TaskExecutionContext, spec *v1.PodTemplateSpec, cfg *config.K8sPluginConfig) error {
	annotations := tCtx.TaskExecutionMetadata().GetAnnotations()
	// Omit some execution specific labels that don't make sense for a reusable env
	labels := lo.OmitByKeys(tCtx.TaskExecutionMetadata().GetLabels(), []string{
		k8s.ExecutionIDLabel, k8s.WorkflowNameLabel, nodes.NodeIDLabel, nodes.TaskNameLabel})

	tmpl, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		return flyteerrors.Wrapf(flyteerrors.BadTaskSpecification, err, "failed to read task template")
	}
	if len(tmpl.GetSecurityContext().GetSecrets()) > 0 {
		secretsMap, err := secrets.MarshalSecretsToMapStrings(tmpl.GetSecurityContext().GetSecrets())
		if err != nil {
			return flyteerrors.Wrapf(flyteerrors.BadTaskSpecification, err, "failed to marshal secrets")
		}
		annotations = utils.UnionMaps(annotations, secretsMap)
		labels[secrets.PodLabel] = secrets.PodLabelValue
	}

	spec.SetAnnotations(utils.UnionMaps(cfg.DefaultAnnotations, spec.GetAnnotations(), annotations))
	spec.SetLabels(utils.UnionMaps(cfg.DefaultLabels, spec.GetLabels(), labels))
	spec.SetNamespace(tCtx.TaskExecutionMetadata().GetNamespace())

	// don't set owner references for fast tasks, as they are intended to outlive a single task execution
	spec.SetOwnerReferences([]metav1.OwnerReference{})

	if cfg.InjectFinalizer { // nolint: staticcheck
		// TODO: add finalizer
	}

	return nil
}

// Handle is the main entrypoint for the plugin. It will offer the task to the worker pool and
// monitor the task until completion.
func (p *Plugin) Handle(ctx context.Context, tCtx core.TaskExecutionContext) (core.Transition, error) {
	executionEnv, fastTaskEnvironment, err := p.getExecutionEnv(ctx, tCtx)
	if err != nil {
		return core.UnknownTransition, err
	}

	// retrieve plugin state
	pluginState := &State{}
	if _, err := tCtx.PluginStateReader().Get(pluginState); err != nil {
		return core.UnknownTransition, flyteerrors.Wrapf(flyteerrors.CorruptedPluginState, err, "Failed to read unmarshal custom state")
	}

	taskID, err := tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedNameWith(0, 50)
	if err != nil {
		return core.UnknownTransition, err
	}

	queueID := fastTaskEnvironment.GetQueueId()
	phaseInfo := core.PhaseInfoUndefined
	switch pluginState.SubmissionPhase {
	case NotSubmitted:
		// read task template
		taskTemplate, err := tCtx.TaskReader().Read(ctx)
		if err != nil {
			return core.UnknownTransition, err
		}

		taskContainer := taskTemplate.GetContainer()
		if taskContainer == nil {
			return core.UnknownTransition, flyteerrors.Errorf(flyteerrors.BadTaskSpecification, "unable to create container with no definition in TaskTemplate")
		}

		templateParameters := template.Parameters{
			TaskExecMetadata: tCtx.TaskExecutionMetadata(),
			Inputs:           tCtx.InputReader(),
			OutputPath:       tCtx.OutputWriter(),
			Task:             tCtx.TaskReader(),
		}
		command, err := template.Render(ctx, taskContainer.GetArgs(), templateParameters)
		if err != nil {
			return core.UnknownTransition, err
		}

		// offer the work to the queue
		ownerID := tCtx.TaskExecutionMetadata().GetOwnerID()
		workerID, err := p.fastTaskService.OfferOnQueue(ctx, queueID, taskID, ownerID.Namespace, ownerID.Name, command)
		if err != nil {
			return core.UnknownTransition, err
		}

		if len(workerID) > 0 {
			pluginState.SubmissionPhase = Submitted
			pluginState.PhaseVersion = core.DefaultPhaseVersion

			pluginState.WorkerID = workerID
			pluginState.LastUpdated = time.Now()

			phaseInfo = core.PhaseInfoQueued(time.Now(), pluginState.PhaseVersion, fmt.Sprintf("task offered to worker %s", workerID))
		} else {
			if pluginState.LastUpdated.IsZero() {
				pluginState.LastUpdated = time.Now()
			}

			//  fail if all replicas for this environment are in a failed state
			taskExecutionID := tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetID()
			executionEnvID := buildExecutionEnvID(taskExecutionID.GetTaskId(), executionEnv)
			statuses, err := tCtx.GetExecutionEnvClient().Status(ctx, executionEnvID)
			if err != nil {
				return core.UnknownTransition, err
			}

			statusesMap := statuses.(map[string]*v1.Pod)

			allReplicasFailed := true
			messageCollector := errorcollector.NewErrorMessageCollector()

			now := time.Now()
			index := 0
			for _, pod := range statusesMap {
				if pod == nil {
					// pod does not exist because it has not yet been populated in the kubeclient
					// cache or was deleted. to be safe, we treat both as a non-failure state.
					allReplicasFailed = false
					break
				}

				phaseInfo, err := podplugin.DemystifyPodStatus(pod, core.TaskInfo{OccurredAt: &now})
				if err != nil {
					return core.UnknownTransition, err
				}

				switch phaseInfo.Phase() {
				case core.PhasePermanentFailure, core.PhaseRetryableFailure:
					if phaseInfo.Err() != nil {
						messageCollector.Collect(index, phaseInfo.Err().GetMessage())
					} else {
						messageCollector.Collect(index, phaseInfo.Reason())
					}
				default:
					allReplicasFailed = false
				}

				index++
			}

			if allReplicasFailed {
				logger.Infof(ctx, "all workers have failed for queue %s", queueID)
				p.metrics.allReplicasFailed.Inc()

				phaseInfo = core.PhaseInfoSystemFailure("unknown", fmt.Sprintf("all workers have failed for queue %s\n%s",
					queueID, messageCollector.Summary(maxErrorMessageLength)), nil)
			} else {
				pluginState.PhaseVersion = core.DefaultPhaseVersion
				phaseInfo = core.PhaseInfoWaitingForResourcesInfo(time.Now(), pluginState.PhaseVersion, "no workers available", nil)
			}
		}
	case Submitted:
		// check the task status
		phase, reason, err := p.fastTaskService.CheckStatus(ctx, taskID, fastTaskEnvironment.GetQueueId(), pluginState.WorkerID)

		now := time.Now()
		if err != nil {
			if errors.Is(err, statusUpdateNotFoundError) && now.Sub(pluginState.LastUpdated) > GetConfig().GracePeriodStatusNotFound.Duration {
				// if task has not been updated within the grace period we should abort
				logger.Errorf(ctx, "Task status update not reported within grace period for queue %s and worker %s", queueID, pluginState.WorkerID)
				p.metrics.statusUpdateNotFoundTimeout.Inc()

				return core.DoTransition(core.PhaseInfoSystemRetryableFailure("unknown",
					fmt.Sprintf("task status update not reported within grace period for queue %s and worker %s", queueID, pluginState.WorkerID), nil)), nil
			} else if errors.Is(err, statusUpdateNotFoundError) || errors.Is(err, taskContextNotFoundError) {
				phaseInfo = core.PhaseInfoRunning(pluginState.PhaseVersion, nil)
			} else {
				return core.UnknownTransition, err
			}
		} else if phase == core.PhaseSuccess {
			taskTemplate, err := tCtx.TaskReader().Read(ctx)
			if err != nil {
				return core.UnknownTransition, err
			}

			// gather outputs if they exist
			if taskTemplate.GetInterface() != nil && taskTemplate.GetInterface().GetOutputs() != nil && taskTemplate.GetInterface().GetOutputs().GetVariables() != nil {
				outputReader := ioutils.NewRemoteFileOutputReader(ctx, tCtx.DataStore(), tCtx.OutputWriter(), 0)
				err = tCtx.OutputWriter().Put(ctx, outputReader)
				if err != nil {
					return core.UnknownTransition, err
				}
			}

			phaseInfo = core.PhaseInfoSuccess(nil)
		} else if phase == core.PhaseRetryableFailure {
			return core.DoTransition(core.PhaseInfoRetryableFailure("unknown", reason, nil)), nil
		} else {
			pluginState.PhaseVersion++
			pluginState.LastUpdated = now
			phaseInfo = core.PhaseInfoRunning(pluginState.PhaseVersion, nil)
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
	_, fastTaskEnvironment, err := p.getExecutionEnv(ctx, tCtx)
	if err != nil {
		return err
	}

	// retrieve plugin state
	pluginState := &State{}
	if _, err := tCtx.PluginStateReader().Get(pluginState); err != nil {
		return flyteerrors.Wrapf(flyteerrors.CorruptedPluginState, err, "Failed to read unmarshal custom state")
	}

	taskID, err := tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedNameWith(0, 50)
	if err != nil {
		return err
	}

	return p.fastTaskService.Cleanup(ctx, taskID, fastTaskEnvironment.GetQueueId(), pluginState.WorkerID)
}

// init registers the plugin with the plugin machinery.
func init() {
	pluginmachinery.PluginRegistry().RegisterCorePlugin(
		core.PluginEntry{
			ID:                  fastTaskType,
			RegisteredTaskTypes: []core.TaskType{fastTaskType},
			LoadPlugin: func(ctx context.Context, iCtx core.SetupContext) (core.Plugin, error) {
				// open tcp listener
				listener, err := net.Listen("tcp", GetConfig().Endpoint)
				if err != nil {
					return nil, err
				}

				// create and start grpc server
				fastTaskService := newFastTaskService(iCtx.EnqueueOwner(), iCtx.MetricsScope())
				go func() {
					grpcServer := grpc.NewServer()
					pb.RegisterFastTaskServer(grpcServer, fastTaskService)
					if err := grpcServer.Serve(listener); err != nil {
						panic("failed to start grpc fast task grpc server")
					}
				}()

				return &Plugin{
					fastTaskService: fastTaskService,
					metrics:         newPluginMetrics(iCtx.MetricsScope()),
				}, nil
			},
			IsDefault: false,
		},
	)
}
