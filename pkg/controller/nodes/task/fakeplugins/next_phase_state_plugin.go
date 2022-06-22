package fakeplugins

import (
	"context"
	"fmt"
	"time"

	pluginCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/stretchr/testify/mock"
)

type NextPhaseState struct {
	Phase        pluginCore.Phase
	PhaseVersion uint32
	TaskInfo     *pluginCore.TaskInfo
	TaskErr      *io.ExecutionError
	DeckExists   bool
	OutputExists bool
	OrError      bool
}

type NextPhaseStatePlugin struct {
	id    string
	props pluginCore.PluginProperties
}

func (n NextPhaseStatePlugin) GetID() string {
	return n.id
}

func (n NextPhaseStatePlugin) GetProperties() pluginCore.PluginProperties {
	return n.props
}

func (n NextPhaseStatePlugin) Handle(ctx context.Context, tCtx pluginCore.TaskExecutionContext) (pluginCore.Transition, error) {
	s := &NextPhaseState{}
	if _, err := tCtx.PluginStateReader().Get(s); err != nil {
		return pluginCore.UnknownTransition, err
	}
	if s.OrError {
		return pluginCore.UnknownTransition, fmt.Errorf("state requests error")
	}
	switch s.Phase {
	case pluginCore.PhaseSuccess:
		r := &mocks.OutputReader{}
		isErr := false
		if s.TaskErr != nil {
			isErr = true
			r.On("ReadError", mock.Anything).Return(*s.TaskErr, nil)
		}

		r.OnDeckExistsMatch(mock.Anything).Return(s.DeckExists, nil)
		r.On("IsError", mock.Anything).Return(isErr, nil)
		r.On("IsFile", mock.Anything).Return(true)
		r.On("Exists", mock.Anything).Return(s.OutputExists, nil)
		if err := tCtx.OutputWriter().Put(ctx, r); err != nil {
			return pluginCore.UnknownTransition, err
		}
		return pluginCore.DoTransition(pluginCore.PhaseInfoSuccess(s.TaskInfo)), nil
	case pluginCore.PhasePermanentFailure:
		return pluginCore.DoTransition(pluginCore.PhaseInfoFailure("failed", "message", s.TaskInfo)), nil
	case pluginCore.PhaseRetryableFailure:
		return pluginCore.DoTransition(pluginCore.PhaseInfoRetryableFailure("failed", "message", s.TaskInfo)), nil
	case pluginCore.PhaseNotReady:
		return pluginCore.DoTransition(pluginCore.PhaseInfoNotReady(time.Now(), s.PhaseVersion, "not-ready")), nil
	case pluginCore.PhaseInitializing:
		return pluginCore.DoTransition(pluginCore.PhaseInfoInitializing(time.Now(), s.PhaseVersion, "initializing", s.TaskInfo)), nil
	case pluginCore.PhaseQueued:
		return pluginCore.DoTransition(pluginCore.PhaseInfoQueued(time.Now(), s.PhaseVersion, "queued")), nil
	case pluginCore.PhaseRunning:
		return pluginCore.DoTransition(pluginCore.PhaseInfoRunning(s.PhaseVersion, s.TaskInfo)), nil
	case pluginCore.PhaseWaitingForResources:
		return pluginCore.DoTransition(pluginCore.PhaseInfoWaitingForResourcesInfo(time.Now(), s.PhaseVersion, "waiting", nil)), nil
	}
	return pluginCore.UnknownTransition, nil
}

func (n NextPhaseStatePlugin) Abort(ctx context.Context, tCtx pluginCore.TaskExecutionContext) error {
	return nil
}

func (n NextPhaseStatePlugin) Finalize(ctx context.Context, tCtx pluginCore.TaskExecutionContext) error {
	return nil
}

func NewPhaseBasedPlugin() NextPhaseStatePlugin {
	return NextPhaseStatePlugin{
		id:    "next_phase_plugin",
		props: pluginCore.PluginProperties{},
	}
}
