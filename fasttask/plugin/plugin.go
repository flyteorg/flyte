package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"

	idlcore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	flyteerrors "github.com/flyteorg/flyte/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/template"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils"

	"github.com/unionai/flyte/fasttask/plugin/pb"
)

const fastTaskType = "fast-task"

var statusUpdateNotFoundError = errors.New("StatusUpdateNotFound")

type Phase int

const (
	PhaseNotStarted Phase = iota
	PhaseRunning
)

// State maintains the current status of the task execution.
type State struct {
	Phase       Phase
	WorkerID    string
	LastUpdated time.Time
}

// Plugin is a fast task plugin that offers task execution to a worker pool.
type Plugin struct {
	fastTaskService *FastTaskService
}

// GetID returns the unique identifier for the plugin.
func (p *Plugin) GetID() string {
	return fastTaskType
}

// GetProperties returns the properties of the plugin.
func (p *Plugin) GetProperties() core.PluginProperties {
	return core.PluginProperties{}
}

// getExecutionEnv retrieves the execution environment for the task. If the environment does not
// exist, it will create it.
// this is here because we wanted uniformity within `TaskExecutionContext` where functions simply
// return an interface rather than doing any actual work. alternatively, we could bury this within
// `NodeExecutionContext` so other `ExecutionEnvironment` plugins do not need to duplicate this.
func (p *Plugin) getExecutionEnv(ctx context.Context, tCtx core.TaskExecutionContext) (*pb.FastTaskEnvironment, error) {
	taskTemplate, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		return nil, err
	}

	executionEnv := &idlcore.ExecutionEnv{}
	if err := utils.UnmarshalStruct(taskTemplate.GetCustom(), executionEnv); err != nil {
		return nil, flyteerrors.Wrapf(flyteerrors.BadTaskSpecification, err, "failed to unmarshal environment")
	}

	switch e := executionEnv.GetEnvironment().(type) {
	case *idlcore.ExecutionEnv_Spec:
		executionEnvClient := tCtx.GetExecutionEnvClient()

		// if environment already exists then return it
		if environment := executionEnvClient.Get(ctx, executionEnv.GetId()); environment != nil {
			fastTaskEnvironment := &pb.FastTaskEnvironment{}
			if err := utils.UnmarshalStruct(environment, fastTaskEnvironment); err != nil {
				return nil, flyteerrors.Wrapf(flyteerrors.BadTaskSpecification, err, "failed to unmarshal environment client")
			}

			return fastTaskEnvironment, nil
		}

		// create environment
		environmentSpec := e.Spec

		fastTaskEnvironmentSpec := &pb.FastTaskEnvironmentSpec{}
		if err := utils.UnmarshalStruct(environmentSpec, fastTaskEnvironmentSpec); err != nil {
			return nil, flyteerrors.Wrapf(flyteerrors.BadTaskSpecification, err, "failed to unmarshal environment spec")
		}

		// if podTemplateSpec is not popualated - then generate from tCtx
		if len(fastTaskEnvironmentSpec.GetPodTemplateSpec()) == 0 {
			podSpec, objectMeta, primaryContainerName, err := flytek8s.ToK8sPodSpec(ctx, tCtx)
			if err != nil {
				return nil, flyteerrors.Wrapf(flyteerrors.BadTaskSpecification, err, "failed to create environment")
			}

			podTemplateSpec := &v1.PodTemplateSpec{
				ObjectMeta: *objectMeta,
				Spec:       *podSpec,
			}
			podTemplateSpec.SetNamespace(tCtx.TaskExecutionMetadata().GetNamespace())

			// need to marshal as JSON to maintain container resources, proto serialization does
			// not persist these settings for `PodSpec`
			podTemplateSpecBytes, err := json.Marshal(podTemplateSpec)
			if err != nil {
				return nil, flyteerrors.Wrapf(flyteerrors.BadTaskSpecification, err, "failed to marshal pod template spec")
			}

			fastTaskEnvironmentSpec.PodTemplateSpec = podTemplateSpecBytes
			if err := utils.MarshalStruct(fastTaskEnvironmentSpec, environmentSpec); err != nil {
				return nil, fmt.Errorf("unable to marshal EnvironmentSpec [%v], Err: [%v]", fastTaskEnvironmentSpec, err.Error())
			}

			fastTaskEnvironmentSpec.PrimaryContainerName = primaryContainerName
		}

		environment, err := executionEnvClient.Create(ctx, executionEnv.GetId(), environmentSpec)
		if err != nil {
			return nil, flyteerrors.Wrapf(flyteerrors.BadTaskSpecification, err, "failed to create environment")
		}

		fastTaskEnvironment := &pb.FastTaskEnvironment{}
		if err := utils.UnmarshalStruct(environment, fastTaskEnvironment); err != nil {
			return nil, flyteerrors.Wrapf(flyteerrors.BadTaskSpecification, err, "failed to unmarshal environment extant")
		}

		return fastTaskEnvironment, nil
	case *idlcore.ExecutionEnv_Extant:
		fastTaskEnvironment := &pb.FastTaskEnvironment{}
		if err := utils.UnmarshalStruct(e.Extant, fastTaskEnvironment); err != nil {
			return nil, flyteerrors.Wrapf(flyteerrors.BadTaskSpecification, err, "failed to unmarshal environment extant")
		}

		return fastTaskEnvironment, nil
	}

	return nil, nil
}

// Handle is the main entrypoint for the plugin. It will offer the task to the worker pool and
// monitor the task until completion.
func (p *Plugin) Handle(ctx context.Context, tCtx core.TaskExecutionContext) (core.Transition, error) {
	fastTaskEnvironment, err := p.getExecutionEnv(ctx, tCtx)
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

	phaseInfo := core.PhaseInfoUndefined
	switch pluginState.Phase {
	case PhaseNotStarted:
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
		workerID, err := p.fastTaskService.OfferOnQueue(ctx, fastTaskEnvironment.GetQueueId(), taskID, ownerID.Namespace, ownerID.Name, command)
		if err != nil {
			return core.UnknownTransition, err
		}

		if len(workerID) > 0 {
			pluginState.Phase = PhaseRunning
			pluginState.WorkerID = workerID
			pluginState.LastUpdated = time.Now()

			phaseInfo = core.PhaseInfoRunning(core.DefaultPhaseVersion, nil)
		} else {
			if pluginState.LastUpdated.IsZero() {
				pluginState.LastUpdated = time.Now()
			}

			// fail if no worker available within grace period
			if time.Since(pluginState.LastUpdated) > GetConfig().GracePeriodWorkersUnavailable.Duration {
				phaseInfo = core.PhaseInfoSystemFailure("unknown", "timed out waiting for worker availability", nil)
			} else {
				phaseInfo = core.PhaseInfoNotReady(time.Now(), core.DefaultPhaseVersion, "no workers available")
			}
		}
	case PhaseRunning:
		// check the task status
		phase, reason, err := p.fastTaskService.CheckStatus(ctx, taskID, fastTaskEnvironment.GetQueueId(), pluginState.WorkerID)

		now := time.Now()
		if err != nil && !errors.Is(err, statusUpdateNotFoundError) {
			return core.UnknownTransition, err
		} else if errors.Is(err, statusUpdateNotFoundError) && now.Sub(pluginState.LastUpdated) > GetConfig().GracePeriodStatusNotFound.Duration {
			// if task has not been updated within the grace period we should abort
			return core.DoTransition(core.PhaseInfoSystemRetryableFailure("unknown", "task status update not reported within grace period", nil)), nil
		} else if phase == core.PhaseSuccess {
			taskTemplate, err := tCtx.TaskReader().Read(ctx)
			if err != nil {
				return core.UnknownTransition, err
			}

			// gather outputs if they exist
			if taskTemplate.GetInterface() != nil && taskTemplate.GetInterface().GetOutputs() != nil && taskTemplate.GetInterface().GetOutputs().GetVariables() != nil {
				outputReader := ioutils.NewRemoteFileOutputReader(ctx, tCtx.DataStore(), tCtx.OutputWriter(), tCtx.MaxDatasetSizeBytes())
				err = tCtx.OutputWriter().Put(ctx, outputReader)
				if err != nil {
					return core.UnknownTransition, err
				}
			}

			phaseInfo = core.PhaseInfoSuccess(nil)
		} else if phase == core.PhaseRetryableFailure {
			return core.DoTransition(core.PhaseInfoRetryableFailure("unknown", reason, nil)), nil
		} else {
			pluginState.LastUpdated = now
			phaseInfo = core.PhaseInfoRunning(core.DefaultPhaseVersion, nil)
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
	// TODO: add this since the service may now hold pending tasks without a worker up
	return nil
}

// Finalize is called when the task execution is complete, performing any necessary cleanup.
func (p *Plugin) Finalize(ctx context.Context, tCtx core.TaskExecutionContext) error {
	fastTaskEnvironment, err := p.getExecutionEnv(ctx, tCtx)
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
				fastTaskService := NewFastTaskService(iCtx.EnqueueOwner())
				go func() {
					grpcServer := grpc.NewServer()
					pb.RegisterFastTaskServer(grpcServer, fastTaskService)
					if err := grpcServer.Serve(listener); err != nil {
						panic("failed to start grpc fast task grpc server")
					}
				}()

				return &Plugin{
					fastTaskService: fastTaskService,
				}, nil
			},
			IsDefault: false,
		},
	)
}
