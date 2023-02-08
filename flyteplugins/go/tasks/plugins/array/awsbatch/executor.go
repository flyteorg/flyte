package awsbatch

import (
	"context"

	arrayCore "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/core"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/awsbatch/definition"

	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array"
	batchConfig "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/awsbatch/config"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/utils"

	"github.com/flyteorg/flyteplugins/go/tasks/aws"

	"github.com/flyteorg/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"

	idlCore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

const (
	executorName              = "aws_array"
	defaultPluginStateVersion = 0
	arrayTaskType             = "container_array"
)

type Executor struct {
	jobStore           *JobStore
	jobDefinitionCache definition.Cache

	outputAssembler array.OutputAssembler
	errorAssembler  array.OutputAssembler
	metrics         ExecutorMetrics
}

func (e Executor) GetID() string {
	return executorName
}

func (e Executor) GetProperties() core.PluginProperties {
	return core.PluginProperties{
		DisableNodeLevelCaching: true,
	}
}

func (e Executor) Handle(ctx context.Context, tCtx core.TaskExecutionContext) (core.Transition, error) {
	pluginConfig := batchConfig.GetConfig()

	pluginState := &State{}
	if _, err := tCtx.PluginStateReader().Get(pluginState); err != nil {
		return core.UnknownTransition, errors.Wrapf(errors.CorruptedPluginState, err, "Failed to read unmarshal custom state")
	}

	if pluginState.State == nil {
		pluginState.State = &arrayCore.State{}
	}

	var err error
	var externalResources []*core.ExternalResource

	p, version := pluginState.GetPhase()
	logger.Infof(ctx, "Entering handle with phase [%v]", p)

	switch p {
	case arrayCore.PhaseStart:
		pluginState.State, err = array.DetermineDiscoverability(ctx, tCtx, pluginConfig.MaxArrayJobSize, pluginState.State)

	case arrayCore.PhasePreLaunch:
		pluginState, err = EnsureJobDefinition(ctx, tCtx, pluginConfig, e.jobStore.Client, e.jobDefinitionCache, pluginState, version+1)

	case arrayCore.PhaseWaitingForResources:
		fallthrough

	case arrayCore.PhaseLaunch:
		pluginState, err = LaunchSubTasks(ctx, tCtx, e.jobStore, pluginConfig, pluginState, e.metrics, version+1)

	case arrayCore.PhaseCheckingSubTaskExecutions:
		pluginState, err = CheckSubTasksState(ctx, tCtx, e.jobStore, pluginConfig, pluginState, e.metrics)

	case arrayCore.PhaseAssembleFinalOutput:
		pluginState.State, err = array.AssembleFinalOutputs(ctx, e.outputAssembler, tCtx, arrayCore.PhaseSuccess, version+1, pluginState.State)

	case arrayCore.PhaseWriteToDiscoveryThenFail:
		pluginState.State, externalResources, err = array.WriteToDiscovery(ctx, tCtx, pluginState.State, arrayCore.PhaseAssembleFinalError, version+1)

	case arrayCore.PhaseWriteToDiscovery:
		pluginState.State, externalResources, err = array.WriteToDiscovery(ctx, tCtx, pluginState.State, arrayCore.PhaseAssembleFinalOutput, version+1)

	case arrayCore.PhaseAssembleFinalError:
		pluginState.State, err = array.AssembleFinalOutputs(ctx, e.errorAssembler, tCtx, arrayCore.PhaseRetryableFailure, version+1, pluginState.State)
	}

	if err != nil {
		return core.UnknownTransition, err
	}

	if err := tCtx.PluginStateWriter().Put(defaultPluginStateVersion, pluginState); err != nil {
		return core.UnknownTransition, err
	}

	// Always attempt to augment phase with task logs.
	var logLinks []*idlCore.TaskLog

	nextPhase, _ := pluginState.GetPhase()
	if p == arrayCore.PhaseStart && nextPhase != arrayCore.PhaseStart {
		// if transitioning from PhaseStart to another phase then cache lookups have completed
		externalResources, err = arrayCore.InitializeExternalResources(ctx, tCtx, pluginState.State,
			func(tCtx core.TaskExecutionContext, childIndex int) string {
				// subTaskIDs for the the aws_batch are generated based on the job ID, therefore
				// to initialize we default to an empty string which will be updated later.
				return ""
			},
		)
	} else if p != arrayCore.PhaseStart && p != arrayCore.PhaseWriteToDiscovery && p != arrayCore.PhaseWriteToDiscoveryThenFail {
		// if externalResources is not otherwise being populated then attempt to get task log links
		logLinks, externalResources, err = GetTaskLinks(ctx, tCtx.TaskExecutionMetadata(), e.jobStore, pluginState)
	}

	if err != nil {
		return core.UnknownTransition, err
	}

	logger.Infof(ctx, "Exiting handle with phase [%v]", pluginState.State.CurrentPhase)

	// Determine transition information from the state
	phaseInfo, err := arrayCore.MapArrayStateToPluginPhase(ctx, pluginState.State, logLinks, externalResources)
	if err != nil {
		return core.UnknownTransition, err
	}

	return core.DoTransition(phaseInfo), nil
}

func (e Executor) Abort(ctx context.Context, tCtx core.TaskExecutionContext) error {
	return TerminateSubTasks(ctx, tCtx, e.jobStore.Client, "Aborted", e.metrics)
}

func (e Executor) Finalize(ctx context.Context, tCtx core.TaskExecutionContext) error {
	return TerminateSubTasks(ctx, tCtx, e.jobStore.Client, "Finalized", e.metrics)
}

func NewExecutor(ctx context.Context, awsClient aws.Client, cfg *batchConfig.Config,
	enqueueOwner core.EnqueueOwner, scope promutils.Scope) (Executor, error) {

	getRateLimiter := utils.NewRateLimiter("getRateLimiter", float64(cfg.GetRateLimiter.Rate),
		cfg.GetRateLimiter.Burst)
	defaultRateLimiter := utils.NewRateLimiter("defaultRateLimiter", float64(cfg.DefaultRateLimiter.Rate),
		cfg.DefaultRateLimiter.Burst)
	batchClient := NewBatchClient(awsClient, getRateLimiter, defaultRateLimiter)
	jobStore, err := NewJobStore(ctx, batchClient, cfg.JobStoreConfig, EventHandler{
		Updated: func(ctx context.Context, event Event) {
			err := enqueueOwner(event.NewJob.OwnerReference)
			if err != nil {
				logger.Warnf(ctx, "Failed to enqueue owner [%v] of job [%v]. Error: %v", event.NewJob.OwnerReference, event.NewJob.ID)
			}
		},
	}, scope)

	if err != nil {
		return Executor{}, err
	}

	outputAssembler, err := array.NewOutputAssembler(cfg.OutputAssembler, scope.NewSubScope("output_assembler"))
	if err != nil {
		return Executor{}, err
	}

	errorAssembler, err := array.NewErrorAssembler(cfg.MaxErrorStringLength, cfg.ErrorAssembler,
		scope.NewSubScope("error_assembler"))
	if err != nil {
		return Executor{}, err
	}

	return Executor{
		jobStore:           &jobStore,
		jobDefinitionCache: definition.NewCache(cfg.JobDefCacheSize),
		outputAssembler:    outputAssembler,
		errorAssembler:     errorAssembler,
		metrics:            getAwsBatchExecutorMetrics(scope.NewSubScope("awsbatch")),
	}, nil
}

func (e Executor) Start(ctx context.Context) error {
	if err := e.jobStore.Start(ctx); err != nil {
		return err
	}

	if err := e.outputAssembler.Start(ctx); err != nil {
		return err
	}

	if err := e.errorAssembler.Start(ctx); err != nil {
		return err
	}

	return nil
}

func init() {
	pluginmachinery.PluginRegistry().RegisterCorePlugin(
		core.PluginEntry{
			ID:                  executorName,
			RegisteredTaskTypes: []core.TaskType{arrayTaskType, array.AwsBatchTaskType},
			LoadPlugin:          createNewExecutorPlugin,
			IsDefault:           false,
		})
}

func createNewExecutorPlugin(ctx context.Context, iCtx core.SetupContext) (core.Plugin, error) {
	awsClient, err := aws.GetClient()
	if err != nil {
		return nil, err
	}

	exec, err := NewExecutor(ctx, awsClient, batchConfig.GetConfig(), iCtx.EnqueueOwner(), iCtx.MetricsScope().NewSubScope(executorName))
	if err != nil {
		return nil, err
	}

	if err = exec.Start(ctx); err != nil {
		return nil, err
	}

	return exec, nil
}
