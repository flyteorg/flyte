package plugins

import (
	"context"
	"fmt"
	"time"

	idlcore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/config"
	flyteerrors "github.com/flyteorg/flyte/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
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

type State struct {
	InputToOuputVariableNames map[string]string
	Started                   bool 
	StartTime                 time.Time
}

type EchoPlugin struct {
}

func (e *EchoPlugin) GetID() string {
	return echoTaskType
}

func (e *EchoPlugin) GetProperties() core.PluginProperties {
	return core.PluginProperties{}
}

func (e *EchoPlugin) Handle(ctx context.Context, tCtx core.TaskExecutionContext) (core.Transition, error) {
	// retrieve plugin state
	pluginState := &State{}
	if _, err := tCtx.PluginStateReader().Get(pluginState); err != nil {
		return core.UnknownTransition, flyteerrors.Wrapf(flyteerrors.CorruptedPluginState, err, "Failed to read unmarshal custom state")
	}

	if !pluginState.Started {
		pluginState.Started = true
		pluginState.StartTime = time.Now()

		// validate outputs are castable from inputs otherwise error as this plugin is not applicable
		taskTemplate, err := tCtx.TaskReader().Read(ctx)
		if err != nil {
			return core.UnknownTransition, fmt.Errorf("failed to read TaskTemplate: [%w]", err)
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
			return core.UnknownTransition, fmt.Errorf("the number of input [%d] and output [%d] variables does not match", len(inputs), len(outputs))
		} else if len(inputs) > 1 {
			return core.UnknownTransition, fmt.Errorf("this plugin does not currently support more than one input variable")
		}

		pluginState.InputToOuputVariableNames = make(map[string]string, len(inputs))
		for inputName, _ := range inputs {
			firstCastableOutputName := ""
			for outputName, _ := range outputs {
				// TODO - need to check if types are castable to support multiple values
				firstCastableOutputName = outputName
				break
			}

			if len(firstCastableOutputName) == 0 {
				return core.UnknownTransition, fmt.Errorf("no castable output variable found for input variable [%s]", inputName)
			}

			delete(outputs, firstCastableOutputName)
			pluginState.InputToOuputVariableNames[inputName] = firstCastableOutputName
		}
	}

	echoConfig := EchoPluginConfigSection.GetConfig().(*EchoPluginConfig)
	if time.Since(pluginState.StartTime) > echoConfig.SleepDuration.Duration {
		// copy inputs to outputs
		if len(pluginState.InputToOuputVariableNames) > 0 {
			inputLiterals, err := tCtx.InputReader().Get(ctx)
			if err != nil {
				return core.UnknownTransition, err
			}

			outputLiterals := make(map[string]*idlcore.Literal, len(pluginState.InputToOuputVariableNames))
			for inputName, outputName := range pluginState.InputToOuputVariableNames {
				outputLiterals[outputName] = inputLiterals.Literals[inputName]
			}

			outputLiteralMap := &idlcore.LiteralMap{
				Literals: outputLiterals,
			}

			outputFile := tCtx.OutputWriter().GetOutputPath()
			if err := tCtx.DataStore().WriteProtobuf(ctx, outputFile, storage.Options{}, outputLiteralMap); err != nil {
				return core.UnknownTransition, err
			}
		}

		return core.DoTransition(core.PhaseInfoSuccess(nil)), nil
	}

	// update plugin state
	if err := tCtx.PluginStateWriter().Put(0, pluginState); err != nil {
		return core.UnknownTransition, err
	}

	return core.DoTransition(core.PhaseInfoRunning(core.DefaultPhaseVersion, nil)), nil
}

func (e *EchoPlugin) Abort(ctx context.Context, tCtx core.TaskExecutionContext) error {
	return nil
}

func (e *EchoPlugin) Finalize(ctx context.Context, tCtx core.TaskExecutionContext) error {
	return nil
}

func init() {
	pluginmachinery.PluginRegistry().RegisterCorePlugin(
		core.PluginEntry{
			ID:                  echoTaskType,
			RegisteredTaskTypes: []core.TaskType{echoTaskType},
			LoadPlugin:          func(ctx context.Context, iCtx core.SetupContext) (core.Plugin, error) {
				return &EchoPlugin{}, nil
			},
			IsDefault: true,
		},
	)
}
