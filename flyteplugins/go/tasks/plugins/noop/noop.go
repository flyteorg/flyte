package noop

import (
	"context"
	pluginErrors "github.com/flyteorg/flyte/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/ioutils"
)

const (
	noopTaskType = "noop"
)

type NoopPlugin struct{}

func (n NoopPlugin) GetID() string {
	return noopTaskType
}

func (n NoopPlugin) GetProperties() core.PluginProperties {
	return core.PluginProperties{}
}

func (n NoopPlugin) Handle(ctx context.Context, tCtx core.TaskExecutionContext) (core.Transition, error) {
	inputs, err := tCtx.InputReader().Get(ctx)
	if err != nil {
		return core.DoTransition(core.PhaseInfoFailure(pluginErrors.TaskFailedWithError, "Failed to read the inputs.", nil)), nil
	}

	err = tCtx.OutputWriter().Put(ctx, ioutils.NewInMemoryOutputReader(
		inputs, nil, nil))
	if err != nil {
		return core.DoTransition(core.PhaseInfoFailure(pluginErrors.TaskFailedWithError, "Failed to write outputs", nil)), nil
	}

	return core.DoTransition(core.PhaseInfoSuccess(nil)), nil
}

func (n NoopPlugin) Abort(ctx context.Context, tCtx core.TaskExecutionContext) error {
	return nil
}

func (n NoopPlugin) Finalize(ctx context.Context, tCtx core.TaskExecutionContext) error {
	return nil
}

func init() {
	pluginmachinery.PluginRegistry().RegisterCorePlugin(
		core.PluginEntry{
			ID:                  noopTaskType,
			RegisteredTaskTypes: []core.TaskType{noopTaskType},
			LoadPlugin: func(ctx context.Context, iCtx core.SetupContext) (core.Plugin, error) {
				return NoopPlugin{}, nil
			},
			IsDefault: true,
		},
	)
}
