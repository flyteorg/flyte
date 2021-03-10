package fakeplugins

import (
	"context"
	"fmt"

	pluginCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
)

type HandleResponse struct {
	T   pluginCore.Transition
	Err error
}

func NewHandleTransition(transition pluginCore.Transition) HandleResponse {
	return HandleResponse{
		T:   transition,
		Err: nil,
	}
}

func NewHandleError(err error) HandleResponse {
	return HandleResponse{
		T:   pluginCore.UnknownTransition,
		Err: err,
	}
}

type taskReplayer struct {
	nextOnHandleResponseIdx   int
	nextOnAbortResponseIdx    int
	nextOnFinalizeResponseIdx int
}

// This is a test plugin and can be used to play any scenario responses from a plugin, (exceptions: panic)
// The plugin is to be invoked within a single thread (not thread safe) and is very simple in terms of usage
// It does not use any state and does not drive the state machine using that state. It drives the state machine constantly forward for a taskID
type ReplayerPlugin struct {
	id                       string
	props                    pluginCore.PluginProperties
	orderedOnHandleResponses []HandleResponse
	orderedAbortResponses    []error
	orderedFinalizeResponses []error
	taskReplayState          map[string]*taskReplayer
}

func (r ReplayerPlugin) GetID() string {
	return r.id
}

func (r ReplayerPlugin) GetProperties() pluginCore.PluginProperties {
	return r.props
}

func (r ReplayerPlugin) Handle(_ context.Context, tCtx pluginCore.TaskExecutionContext) (pluginCore.Transition, error) {
	n := tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()
	s, ok := r.taskReplayState[n]
	if !ok {
		s = &taskReplayer{}
		r.taskReplayState[n] = s
	}
	defer func() {
		s.nextOnHandleResponseIdx++
	}()
	if s.nextOnHandleResponseIdx > len(r.orderedOnHandleResponses) {
		return pluginCore.UnknownTransition, fmt.Errorf("plugin Handle Invoked [%d] times, expected [%d] for task [%s]", s.nextOnHandleResponseIdx, len(r.orderedOnHandleResponses), n)
	}
	hr := r.orderedOnHandleResponses[s.nextOnHandleResponseIdx]
	return hr.T, hr.Err
}

func (r ReplayerPlugin) Abort(_ context.Context, tCtx pluginCore.TaskExecutionContext) error {
	n := tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()
	s, ok := r.taskReplayState[n]
	if !ok {
		s = &taskReplayer{}
		r.taskReplayState[n] = s
	}
	defer func() {
		s.nextOnAbortResponseIdx++
	}()
	if s.nextOnAbortResponseIdx > len(r.orderedAbortResponses) {
		return fmt.Errorf("plugin Abort Invoked [%d] times, expected [%d] for task [%s]", s.nextOnAbortResponseIdx, len(r.orderedAbortResponses), n)
	}
	return r.orderedAbortResponses[s.nextOnAbortResponseIdx]
}

func (r ReplayerPlugin) Finalize(_ context.Context, tCtx pluginCore.TaskExecutionContext) error {
	n := tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()
	s, ok := r.taskReplayState[n]
	if !ok {
		s = &taskReplayer{}
		r.taskReplayState[n] = s
	}
	defer func() {
		s.nextOnAbortResponseIdx++
	}()
	if s.nextOnAbortResponseIdx > len(r.orderedAbortResponses) {
		return fmt.Errorf("plugin Finalize Invoked [%d] times, expected [%d] for task [%s]", s.nextOnAbortResponseIdx, len(r.orderedFinalizeResponses), n)
	}
	return r.orderedFinalizeResponses[s.nextOnFinalizeResponseIdx]
}

func (r ReplayerPlugin) VerifyAllCallsCompleted(taskExecID string) error {
	s, ok := r.taskReplayState[taskExecID]
	if !ok {
		s = &taskReplayer{}
		r.taskReplayState[taskExecID] = s
	}
	if s.nextOnFinalizeResponseIdx != len(r.orderedFinalizeResponses)+1 {
		return fmt.Errorf("finalize method expected invocations [%d], actual invocations [%d]", len(r.orderedFinalizeResponses), s.nextOnFinalizeResponseIdx)
	}
	if s.nextOnAbortResponseIdx != len(r.orderedAbortResponses)+1 {
		return fmt.Errorf("abort method expected invocations [%d], actual invocations [%d]", len(r.orderedAbortResponses), s.nextOnAbortResponseIdx)
	}
	if s.nextOnHandleResponseIdx != len(r.orderedOnHandleResponses)+1 {
		return fmt.Errorf("handle method expected invocations [%d], actual invocations [%d]", len(r.orderedOnHandleResponses), s.nextOnHandleResponseIdx)
	}
	return nil
}

func NewReplayer(forPluginID string, props pluginCore.PluginProperties, orderedOnHandleResponses []HandleResponse, orderedAbortResponses, orderedFinalizeResponses []error) *ReplayerPlugin {
	return &ReplayerPlugin{
		id:                       fmt.Sprintf("replayer-for-%s", forPluginID),
		props:                    props,
		orderedOnHandleResponses: orderedOnHandleResponses,
		orderedAbortResponses:    orderedAbortResponses,
		orderedFinalizeResponses: orderedFinalizeResponses,
		taskReplayState:          make(map[string]*taskReplayer),
	}
}
