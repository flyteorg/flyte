package sleep

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	idlcore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery"
	core "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
)

const sleepTaskType = "core-sleep"

type invalidInputError struct {
	message string
}

func (e *invalidInputError) Error() string {
	return e.message
}

type Plugin struct {
	taskStartTimes map[string]time.Time
	sync.Mutex
}

func (p *Plugin) GetID() string {
	return sleepTaskType
}

func (p *Plugin) GetProperties() core.PluginProperties {
	return core.PluginProperties{}
}

func (p *Plugin) getOrAddTaskStartTime(tCtx core.TaskExecutionContext) time.Time {
	p.Lock()
	defer p.Unlock()

	taskExecutionID := tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()
	if startTime, exists := p.taskStartTimes[taskExecutionID]; exists {
		return startTime
	}

	startTime := time.Now()
	p.taskStartTimes[taskExecutionID] = startTime
	return startTime
}

func (p *Plugin) removeTask(taskExecutionID string) {
	p.Lock()
	defer p.Unlock()
	delete(p.taskStartTimes, taskExecutionID)
}

func (p *Plugin) Handle(ctx context.Context, tCtx core.TaskExecutionContext) (core.Transition, error) {
	sleepDuration, err := resolveSleepDuration(ctx, tCtx)
	if err != nil {
		var invalidErr *invalidInputError
		if errors.As(err, &invalidErr) {
			return core.DoTransition(core.PhaseInfoFailure("BadTaskSpecification", invalidErr.Error(), nil)), nil
		}
		return core.UnknownTransition, err
	}

	if sleepDuration == 0 {
		return core.DoTransition(core.PhaseInfoSuccess(nil)), nil
	}

	startTime := p.getOrAddTaskStartTime(tCtx)
	if time.Since(startTime) >= sleepDuration {
		return core.DoTransition(core.PhaseInfoSuccess(nil)), nil
	}

	return core.DoTransition(core.PhaseInfoRunning(core.DefaultPhaseVersion, nil)), nil
}

func (p *Plugin) Abort(ctx context.Context, tCtx core.TaskExecutionContext) error {
	return nil
}

func (p *Plugin) Finalize(ctx context.Context, tCtx core.TaskExecutionContext) error {
	taskExecutionID := tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()
	p.removeTask(taskExecutionID)
	return nil
}

func resolveSleepDuration(ctx context.Context, tCtx core.TaskExecutionContext) (time.Duration, error) {
	taskTemplate, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to read task template: %w", err)
	}
	if taskTemplate == nil {
		return 0, fmt.Errorf("nil task template")
	}

	iface := taskTemplate.GetInterface()
	if iface == nil || iface.GetInputs() == nil || len(iface.GetInputs().GetVariables()) != 1 {
		return 0, &invalidInputError{message: fmt.Sprintf("task type [%s] requires exactly one duration input", sleepTaskType)}
	}
	if iface.GetOutputs() != nil && len(iface.GetOutputs().GetVariables()) != 0 {
		return 0, &invalidInputError{message: fmt.Sprintf("task type [%s] does not support outputs", sleepTaskType)}
	}

	// v1 GetVariables() returns map[string]*Variable (v2 returns []*VariableEntry)
	inputName := ""
	for name, v := range iface.GetInputs().GetVariables() {
		if v == nil || v.GetType() == nil || v.GetType().GetSimple() != idlcore.SimpleType_DURATION {
			return 0, &invalidInputError{message: fmt.Sprintf("input [%s] must be typed as duration", name)}
		}
		inputName = name
	}

	inputs, err := tCtx.InputReader().Get(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to read task inputs: %w", err)
	}
	if inputs == nil {
		return 0, &invalidInputError{message: fmt.Sprintf("task type [%s] requires a duration input value", sleepTaskType)}
	}

	literal, ok := inputs.GetLiterals()[inputName]
	if !ok || literal == nil {
		return 0, &invalidInputError{message: fmt.Sprintf("duration input [%s] is missing", inputName)}
	}

	durationValue := literal.GetScalar().GetPrimitive().GetDuration()
	if durationValue == nil {
		return 0, &invalidInputError{message: fmt.Sprintf("duration input [%s] must be a duration literal", inputName)}
	}

	sleepDuration := durationValue.AsDuration()
	if sleepDuration < 0 {
		return 0, &invalidInputError{message: fmt.Sprintf("duration input [%s] must be non-negative", inputName)}
	}

	return sleepDuration, nil
}

func init() {
	pluginmachinery.PluginRegistry().RegisterCorePlugin(
		core.PluginEntry{
			ID:                  sleepTaskType,
			RegisteredTaskTypes: []core.TaskType{sleepTaskType},
			LoadPlugin: func(ctx context.Context, iCtx core.SetupContext) (core.Plugin, error) {
				return &Plugin{
					taskStartTimes: make(map[string]time.Time),
				}, nil
			},
			IsDefault: false,
		},
	)
}
