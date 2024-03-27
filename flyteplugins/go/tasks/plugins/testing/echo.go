package testing

import (
	"context"
	"fmt"
	"time"

	idlcore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

const (
	echoTaskType = "echo"
)

type EchoPlugin struct {
	enqueueOwner   core.EnqueueOwner
	taskStartTimes map[string]time.Time
}

func (e *EchoPlugin) GetID() string {
	return echoTaskType
}

func (e *EchoPlugin) GetProperties() core.PluginProperties {
	return core.PluginProperties{}
}

func (e *EchoPlugin) Handle(ctx context.Context, tCtx core.TaskExecutionContext) (core.Transition, error) {
	echoConfig := ConfigSection.GetConfig().(*Config)

	var startTime time.Time
	var exists bool
	taskExecutionID := tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()
	if startTime, exists = e.taskStartTimes[taskExecutionID]; !exists {
		startTime = time.Now()
		e.taskStartTimes[taskExecutionID] = startTime

		// start timer to enqueue owner once task sleep duration has elapsed
		go func() {
			time.Sleep(echoConfig.SleepDuration.Duration)
			if err := e.enqueueOwner(tCtx.TaskExecutionMetadata().GetOwnerID()); err != nil {
				logger.Warnf(ctx, "failed to enqueue owner [%s]: %v", tCtx.TaskExecutionMetadata().GetOwnerID(), err)
			}
		}()
	}

	if time.Since(startTime) >= echoConfig.SleepDuration.Duration {
		// copy inputs to outputs
		inputToOutputVariableMappings, err := compileInputToOutputVariableMappings(ctx, tCtx)
		if err != nil {
			return core.UnknownTransition, err
		}

		if len(inputToOutputVariableMappings) > 0 {
			inputLiterals, err := tCtx.InputReader().Get(ctx)
			if err != nil {
				return core.UnknownTransition, err
			}

			outputLiterals := make(map[string]*idlcore.Literal, len(inputToOutputVariableMappings))
			for inputVariableName, outputVariableName := range inputToOutputVariableMappings {
				outputLiterals[outputVariableName] = inputLiterals.Literals[inputVariableName]
			}

			outputLiteralMap := &idlcore.LiteralMap{
				Literals: outputLiterals,
			}

			outputFile := tCtx.OutputWriter().GetOutputPath()
			if err := tCtx.DataStore().WriteProtobuf(ctx, outputFile, storage.Options{}, outputLiteralMap); err != nil {
				return core.UnknownTransition, err
			}

			or := ioutils.NewRemoteFileOutputReader(ctx, tCtx.DataStore(), tCtx.OutputWriter(), 0)
			if err = tCtx.OutputWriter().Put(ctx, or); err != nil {
				return core.UnknownTransition, err
			}
		}

		return core.DoTransition(core.PhaseInfoSuccess(nil)), nil
	}

	return core.DoTransition(core.PhaseInfoRunning(core.DefaultPhaseVersion, nil)), nil
}

func (e *EchoPlugin) Abort(ctx context.Context, tCtx core.TaskExecutionContext) error {
	return nil
}

func (e *EchoPlugin) Finalize(ctx context.Context, tCtx core.TaskExecutionContext) error {
	taskExecutionID := tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()
	delete(e.taskStartTimes, taskExecutionID)
	return nil
}

func compileInputToOutputVariableMappings(ctx context.Context, tCtx core.TaskExecutionContext) (map[string]string, error) {
	// validate outputs are castable from inputs otherwise error as this plugin is not applicable
	taskTemplate, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read TaskTemplate: [%w]", err)
	}

	var inputs, outputs map[string]*idlcore.Variable
	if taskTemplate.Interface != nil {
		if taskTemplate.Interface.Inputs != nil {
			inputs = taskTemplate.Interface.Inputs.Variables
		}
		if taskTemplate.Interface.Outputs != nil {
			outputs = taskTemplate.Interface.Outputs.Variables
		}
	}

	if len(inputs) != len(outputs) {
		return nil, fmt.Errorf("the number of input [%d] and output [%d] variables does not match", len(inputs), len(outputs))
	} else if len(inputs) > 1 {
		return nil, fmt.Errorf("this plugin does not currently support more than one input variable")
	}

	inputToOutputVariableMappings := make(map[string]string)
	outputVariableNameUsed := make(map[string]struct{})
	for inputVariableName := range inputs {
		firstCastableOutputName := ""
		for outputVariableName := range outputs {
			// TODO - need to check if types are castable to support multiple values
			if _, ok := outputVariableNameUsed[outputVariableName]; !ok {
				firstCastableOutputName = outputVariableName
				break
			}
		}

		if len(firstCastableOutputName) == 0 {
			return nil, fmt.Errorf("no castable output variable found for input variable [%s]", inputVariableName)
		}

		outputVariableNameUsed[firstCastableOutputName] = struct{}{}
		inputToOutputVariableMappings[inputVariableName] = firstCastableOutputName
	}

	return inputToOutputVariableMappings, nil
}

func init() {
	pluginmachinery.PluginRegistry().RegisterCorePlugin(
		core.PluginEntry{
			ID:                  echoTaskType,
			RegisteredTaskTypes: []core.TaskType{echoTaskType},
			LoadPlugin: func(ctx context.Context, iCtx core.SetupContext) (core.Plugin, error) {
				return &EchoPlugin{
					enqueueOwner:   iCtx.EnqueueOwner(),
					taskStartTimes: make(map[string]time.Time),
				}, nil
			},
			IsDefault: true,
		},
	)
}
