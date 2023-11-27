package plugins

import (
	"context"
	"fmt"
	"time"

	idlcore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/config"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/ioutils"
	flytestdconfig "github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

const (
	echoTaskType = "echo"
)

//go:generate pflags EchoPluginConfig --default-var=defaultEchoPluginConfig

var (
	defaultEchoPluginConfig = EchoPluginConfig{
		SleepDuration: flytestdconfig.Duration{Duration: 0 * time.Second},
	}

	EchoPluginConfigSection = config.MustRegisterSubSection(echoTaskType, &defaultEchoPluginConfig)
)

type EchoPluginConfig struct {
	// TODO @hamersaw - docs
	SleepDuration flytestdconfig.Duration `json:"sleep-duration" pflag:"-,TODO"`
}

type EchoPlugin struct {
	taskStartTimes map[string]time.Time
}

func (e *EchoPlugin) GetID() string {
	return echoTaskType
}

func (e *EchoPlugin) GetProperties() core.PluginProperties {
	return core.PluginProperties{}
}

func (e *EchoPlugin) Handle(ctx context.Context, tCtx core.TaskExecutionContext) (core.Transition, error) {
	var startTime time.Time
	var exists bool
	taskExecutionID := tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()
	if startTime, exists = e.taskStartTimes[taskExecutionID]; !exists {
		startTime = time.Now()
		e.taskStartTimes[taskExecutionID] = startTime

		/*if _, err := compileInputToOutputVariableMappings(ctx, tCtx); err != nil {
			return core.UnknownTransition, err
		}*/
	}

	echoConfig := EchoPluginConfigSection.GetConfig().(*EchoPluginConfig)
	if time.Since(startTime) >= echoConfig.SleepDuration.Duration {
		// copy inputs to outputs
		inputToOutputVariableMappings, err := compileInputToOutputVariableMappings(ctx, tCtx)
		if err != nil {
			return core.UnknownTransition, err
		}

		fmt.Printf("HAMERSAW %v\n", inputToOutputVariableMappings)
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

			or := ioutils.NewRemoteFileOutputReader(ctx, tCtx.DataStore(), tCtx.OutputWriter(), tCtx.MaxDatasetSizeBytes())
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
	for inputVariableName, _ := range inputs {
		firstCastableOutputName := ""
		for outputVariableName, _ := range outputs {
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
			LoadPlugin:          func(ctx context.Context, iCtx core.SetupContext) (core.Plugin, error) {
				return &EchoPlugin{
					taskStartTimes: make(map[string]time.Time),
				}, nil
			},
			IsDefault: true,
		},
	)
}
