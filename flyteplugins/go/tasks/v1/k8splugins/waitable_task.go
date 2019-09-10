package k8splugins

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	eventErrors "github.com/lyft/flyteidl/clients/go/events/errors"
	pluginsConfig "github.com/lyft/flyteplugins/go/tasks/v1/config"

	"github.com/lyft/flytestdlib/config"

	"github.com/golang/protobuf/jsonpb"
	"github.com/lyft/flyteidl/clients/go/admin"
	admin2 "github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/plugins"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/storage"
	utils2 "github.com/lyft/flytestdlib/utils"
	v1 "k8s.io/api/core/v1"

	tasksV1 "github.com/lyft/flyteplugins/go/tasks/v1"
	"github.com/lyft/flyteplugins/go/tasks/v1/errors"
	"github.com/lyft/flyteplugins/go/tasks/v1/events"
	"github.com/lyft/flyteplugins/go/tasks/v1/flytek8s"
	"github.com/lyft/flyteplugins/go/tasks/v1/types"
	"github.com/lyft/flyteplugins/go/tasks/v1/utils"
)

const (
	waitableTaskType = "waitable"
	waitablesKey     = "waitables"
	handoverKey      = "handover"
)

var marshaler = &jsonpb.Marshaler{}

var (
	defaultWaitableConfig = &WaitableConfig{
		LruCacheSize: 1000,
	}

	waitableConfigSection = pluginsConfig.MustRegisterSubSection("waitable", defaultWaitableConfig)
)

type WaitableConfig struct {
	// TODO: Reconsider this once we plug Console as a Log Provider plugin.
	ConsoleURI   config.URL `json:"console-uri" pflag:",URI for console. Used to expose links in the emitted events."`
	LruCacheSize int        `json:"lru-cache-size" pflag:",Size of the AutoRefreshCache"`
}

func GetWaitableConfig() *WaitableConfig {
	return waitableConfigSection.GetConfig().(*WaitableConfig)
}

// WaitableTaskExecutor plugin handle tasks of type "waitable". These tasks take as input plugins.Waitable protos. Each
// proto includes an execution id. The plugin asynchronously waits for all executions to reach a terminal state before
// passing the execution over to a ContainerTaskExecutor plugin.
type waitableTaskExecutor struct {
	*flytek8s.K8sTaskExecutor
	containerTaskExecutor
	adminClient     service.AdminServiceClient
	executionsCache utils2.AutoRefreshCache
	dataStore       *storage.DataStore
	recorder        types.EventRecorder
}

// Wrapper object for plugins.waitable to implement CacheItem interface required to use AutoRefreshCache.
type waitableWrapper struct {
	*plugins.Waitable
}

// A unique identifier for the Waitable instance used to dedupe AutoRefreshCache.
func (d *waitableWrapper) ID() string {
	if d.WfExecId != nil {
		return d.WfExecId.String()
	}

	return ""
}

// An override to the default json marshaler to use jsonPb.
func (d *waitableWrapper) MarshalJSON() ([]byte, error) {
	s, err := marshaler.MarshalToString(d.Waitable)
	if err != nil {
		return nil, err
	}

	return []byte(s), nil
}

// An override to the default json unmarshaler to use jsonPb.
func (d *waitableWrapper) UnmarshalJSON(b []byte) error {
	w := plugins.Waitable{}
	err := jsonpb.UnmarshalString(string(b), &w)
	if err != nil {
		return err
	}

	d.Waitable = &w
	return nil
}

// Traverses a literal to find all Waitable objects.
func discoverWaitableInputs(l *core.Literal) (literals []*core.Literal, waitables []*waitableWrapper) {
	switch o := l.Value.(type) {
	case *core.Literal_Collection:
		literals := make([]*core.Literal, 0, len(o.Collection.Literals))
		waitables := make([]*waitableWrapper, 0, len(o.Collection.Literals))
		for _, i := range o.Collection.Literals {
			ls, ws := discoverWaitableInputs(i)
			literals = append(literals, ls...)
			waitables = append(waitables, ws...)
		}

		return literals, waitables
	case *core.Literal_Map:
		literals := make([]*core.Literal, 0, len(o.Map.Literals))
		waitables := make([]*waitableWrapper, 0, len(o.Map.Literals))
		for _, i := range o.Map.Literals {
			ls, ws := discoverWaitableInputs(i)
			literals = append(literals, ls...)
			waitables = append(waitables, ws...)
		}

		return literals, waitables
	case *core.Literal_Scalar:
		switch v := o.Scalar.Value.(type) {
		case *core.Scalar_Generic:
			waitable := &plugins.Waitable{}
			err := utils.UnmarshalStruct(v.Generic, waitable)
			if err != nil {
				// skip, it's just a different type?
				return []*core.Literal{}, []*waitableWrapper{}
			}

			if err = validateWaitable(waitable); err != nil {
				// skip, it's just a different type?
				return []*core.Literal{}, []*waitableWrapper{}
			}

			return []*core.Literal{l}, []*waitableWrapper{{Waitable: waitable}}
		}
	}

	return []*core.Literal{}, []*waitableWrapper{}
}

// Generates workflow execution links as log links.
func generateWorkflowExecutionLinks(ctx context.Context, waitables []*waitableWrapper) []*core.TaskLog {
	cfg := GetWaitableConfig()
	if cfg == nil {
		logger.Info(ctx, "No console endpoint config, skipping the generation of execution links.")
		return []*core.TaskLog{}
	}

	logs := make([]*core.TaskLog, 0, len(waitables))
	for _, w := range waitables {
		logs = append(logs, &core.TaskLog{
			Name: fmt.Sprintf("Exec: %v (%v)", w.WfExecId.Name, w.Phase),
			Uri: fmt.Sprintf("%v/projects/%v/domains/%v/executions/%v", cfg.ConsoleURI.String(),
				w.WfExecId.Project, w.WfExecId.Domain, w.WfExecId.Name),
		})
	}

	return logs
}

// Shadows k8sExecutor initialize to capture some of the initialization params.
func (w *waitableTaskExecutor) Initialize(ctx context.Context, params types.ExecutorInitializationParameters) error {
	w.dataStore = params.DataStore
	w.recorder = params.EventRecorder
	err := w.K8sTaskExecutor.Initialize(ctx, params)
	if err != nil {
		return err
	}

	// We assign a mock one in tests, so let's not override if it's already there
	w.executionsCache, err = utils2.NewAutoRefreshCache(w.syncItem, utils2.NewRateLimiter(
		"admin-get-executions", 50, 50),
		flytek8s.DefaultInformerResyncDuration, GetWaitableConfig().LruCacheSize,
		params.MetricsScope.NewSubScope(waitableTaskType))
	if err != nil {
		return err
	}

	w.executionsCache.Start(ctx)
	return nil
}

// Shadows k8sExecutor
func (w waitableTaskExecutor) StartTask(ctx context.Context, taskCtx types.TaskContext, task *core.TaskTemplate, inputs *core.LiteralMap) (
	types.TaskStatus, error) {

	_, allWaitables := discoverWaitableInputs(&core.Literal{Value: &core.Literal_Map{Map: inputs}})
	for _, waitable := range allWaitables {
		_, err := w.executionsCache.GetOrCreate(waitable)
		if err != nil {
			// System failure
			return types.TaskStatusUndefined, err
		}
	}

	state := map[string]interface{}{
		waitablesKey: allWaitables,
	}

	status := types.TaskStatus{Phase: types.TaskPhaseQueued, PhaseVersion: taskCtx.GetPhaseVersion(), State: state}
	ev := events.CreateEvent(taskCtx, status, &events.TaskEventInfo{
		Logs: generateWorkflowExecutionLinks(ctx, allWaitables),
	})

	err := w.recorder.RecordTaskEvent(ctx, ev)
	if err != nil && eventErrors.IsEventAlreadyInTerminalStateError(err) {
		return types.TaskStatusPermanentFailure(errors.Wrapf(errors.TaskEventRecordingFailed, err,
			"failed to record task event. phase mis-match between Propeller %v and Control Plane.", &status.Phase)), nil
	} else if err != nil {
		logger.Errorf(ctx, "Failed to record task event [%v]. Error: %v", ev, err)
		return types.TaskStatusUndefined, errors.Wrapf(errors.TaskEventRecordingFailed, err,
			"failed to record start task event")
	}

	return status, nil
}

func isTerminalWorkflowPhase(phase core.WorkflowExecution_Phase) bool {
	return phase == core.WorkflowExecution_SUCCEEDED ||
		phase == core.WorkflowExecution_FAILED ||
		phase == core.WorkflowExecution_ABORTED
}

func toWaitableWrapperSlice(ctx context.Context, sliceInterface interface{}) ([]*waitableWrapper, error) {
	waitables, casted := sliceInterface.([]*waitableWrapper)
	if casted {
		return waitables, nil
	}

	waitablesIfaceSlice, casted := sliceInterface.([]interface{})
	if !casted {
		err := fmt.Errorf("failed to cast interface to watiableWrapper. Actual type: %v, Allowed types: [%v, %v]",
			reflect.TypeOf(sliceInterface),
			reflect.TypeOf([]interface{}{}),
			reflect.TypeOf([]*waitableWrapper{}))
		logger.Error(ctx, err)

		return []*waitableWrapper{}, err
	}

	waitables = make([]*waitableWrapper, 0, len(waitablesIfaceSlice))
	for _, item := range waitablesIfaceSlice {
		raw, err := json.Marshal(item)
		if err != nil {
			return nil, errors.Wrapf(errors.RuntimeFailure, err, "failed to marshal state item into json")
		}

		waitable := &waitableWrapper{}
		err = json.Unmarshal(raw, waitable)
		if err != nil {
			return nil, errors.Wrapf(errors.RuntimeFailure, err, "failed to unmarshal state into waitable wrapper")
		}

		waitables = append(waitables, waitable)
	}

	return waitables, nil
}

func (w waitableTaskExecutor) getUpdatedWaitables(ctx context.Context, taskCtx types.TaskContext) (
	updatedWaitables []*waitableWrapper, terminatedCount int, hasChanged bool, err error) {

	state := taskCtx.GetCustomState()
	if state == nil {
		return []*waitableWrapper{}, 0, false, nil
	}

	sliceInterface, found := state[waitablesKey]
	if !found {
		return []*waitableWrapper{}, 0, false, nil
	}

	allWaitables, err := toWaitableWrapperSlice(ctx, sliceInterface)
	if err != nil {
		return []*waitableWrapper{}, 0, false, err
	}

	updatedWaitables = make([]*waitableWrapper, 0, len(allWaitables))
	allDone := 0
	hasChanged = false
	for _, waitable := range allWaitables {
		if !isTerminalWorkflowPhase(waitable.GetPhase()) {
			w, err := w.executionsCache.GetOrCreate(waitable)
			if err != nil {
				return nil, 0, false, err
			}

			newWaitable := w.(*waitableWrapper)
			if newWaitable.Phase != waitable.Phase {
				hasChanged = true
			}

			waitable = newWaitable
		}

		if isTerminalWorkflowPhase(waitable.GetPhase()) {
			allDone++
		}

		updatedWaitables = append(updatedWaitables, waitable)
	}

	return updatedWaitables, allDone, hasChanged, nil
}

func validateWaitable(waitable *plugins.Waitable) error {
	if waitable == nil {
		return fmt.Errorf("empty waitable")
	}

	if waitable.WfExecId == nil {
		return fmt.Errorf("empty executionID")
	}

	if len(waitable.WfExecId.Name) == 0 {
		return fmt.Errorf("empty executionID Name")
	}

	return nil
}

func updateWaitableLiterals(literals []*core.Literal, waitables []*waitableWrapper) error {
	index := make(map[string]*plugins.Waitable, len(waitables))
	for _, w := range waitables {
		index[w.WfExecId.String()] = w.Waitable
	}

	for _, l := range literals {
		orig := &plugins.Waitable{}
		if err := utils.UnmarshalStruct(l.GetScalar().GetGeneric(), orig); err != nil {
			return err
		}

		if err := validateWaitable(orig); err != nil {
			return err
		}

		newW, found := index[orig.WfExecId.String()]
		if !found {
			return fmt.Errorf("couldn't find a waitable corresponding to literal WfID: %v", orig.WfExecId.String())
		}

		if err := utils.MarshalStruct(newW, l.GetScalar().GetGeneric()); err != nil {
			return err
		}
	}

	return nil
}

// Shadows K8sExecutor CheckTaskStatus to check for in-progress workflow executions and only schedule the container when
// they have reached a terminal state.
func (w waitableTaskExecutor) CheckTaskStatus(ctx context.Context, taskCtx types.TaskContext, task *core.TaskTemplate) (
	types.TaskStatus, error) {

	// Have we handed over execution to the container plugin already? if so, then go straight to K8sExecutor to check
	// status.
	_, handOver := taskCtx.GetCustomState()[handoverKey]
	logger.Debugf(ctx, "Hand over state: %v", handOver)

	if taskCtx.GetPhase() == types.TaskPhaseQueued && !handOver {
		logger.Infof(ctx, "Monitoring Launch Plan Execution.")
		// Unmarshal custom state and get latest known state for the watched executions.
		allWaitables, allDone, hasChanged, err := w.getUpdatedWaitables(ctx, taskCtx)
		if err != nil {
			return types.TaskStatusRetryableFailure(err), nil
		}

		// If all executions reached terminal states, hand over execution to K8sExecutor to launch the user container.
		if allDone == len(allWaitables) {
			logger.Infof(ctx, "All [%v] waitables have finished executing.", allDone)
			inputs := &core.LiteralMap{}
			// Rewrite inputs using the latest phases.
			// 1. Read inputs
			if err = w.dataStore.ReadProtobuf(ctx, taskCtx.GetInputsFile(), inputs); err != nil {
				logger.Errorf(ctx, "Failed to read inputs file [%v]. Error: %v", taskCtx.GetInputsFile(), err)
				return types.TaskStatusUndefined, err
			}

			// 2. Get pointers to literals that contain waitables.
			literals, _ := discoverWaitableInputs(&core.Literal{Value: &core.Literal_Map{Map: inputs}})
			logger.Debugf(ctx, "Discovered literals with waitables [%v].", len(literals))

			// 3. Find the corresponding waitables and update original literals.
			if err = updateWaitableLiterals(literals, allWaitables); err != nil {
				logger.Errorf(ctx, "Failed to update Waitable literals. Error: %v", err)
				return types.TaskStatusUndefined, err
			}

			// 4. Write back results into inputs path.
			// TODO: Read after update consistency?
			if err = w.dataStore.WriteProtobuf(ctx, taskCtx.GetInputsFile(), storage.Options{}, inputs); err != nil {
				logger.Errorf(ctx, "Failed to write inputs file back [%v]. Error: %v", taskCtx.GetInputsFile(), err)
				return types.TaskStatusUndefined, err
			}

			status, err := w.K8sTaskExecutor.StartTask(ctx, taskCtx, task, inputs)
			if err != nil {
				logger.Errorf(ctx, "Failed to start k8s task. Error: %v", err)
				return status, err
			}

			logger.Info(ctx, "Launched k8s task with status [%v]", status)

			// If launching the container resulted in queued state, then we need to advance phase version.
			if status.Phase == types.TaskPhaseQueued {
				logger.Debugf(ctx, "StartTask returned queued state.")
				status.PhaseVersion = taskCtx.GetPhaseVersion() + 1
			}

			// Indicate we are done waiting and have launched the task so that next time we go straight to the
			// container plugin.
			state := taskCtx.GetCustomState()
			state[handoverKey] = true

			return status.WithState(state), nil
		} else if hasChanged {
			logger.Infof(ctx, "Some waitable state has changed. Sending event.")
			status := types.TaskStatus{
				Phase:        types.TaskPhaseQueued,
				PhaseVersion: taskCtx.GetPhaseVersion() + 1,
			}

			ev := events.CreateEvent(taskCtx, status, &events.TaskEventInfo{
				Logs: generateWorkflowExecutionLinks(ctx, allWaitables),
			})

			err := w.recorder.RecordTaskEvent(ctx, ev)
			if err != nil && eventErrors.IsEventAlreadyInTerminalStateError(err) {
				return types.TaskStatusPermanentFailure(errors.Wrapf(errors.TaskEventRecordingFailed, err,
					"failed to record task event. phase mis-match between Propeller %v and Control Plane.", &status.Phase)), nil
			} else if err != nil {
				logger.Errorf(ctx, "Failed to record task event [%v]. Error: %v", ev, err)
				return types.TaskStatusUndefined, errors.Wrapf(errors.TaskEventRecordingFailed, err,
					"failed to record task event")
			}
			return status.WithState(map[string]interface{}{
				waitablesKey: allWaitables,
			}), nil
		}

		logger.Debugf(ctx, "No update to waitables in this round.")
		return types.TaskStatus{
			Phase:        types.TaskPhaseQueued,
			PhaseVersion: taskCtx.GetPhaseVersion(),
			State: map[string]interface{}{
				waitablesKey: allWaitables,
			}}, nil
	}

	logger.Infof(ctx, "Handing over task status check to K8s Task Executor.")
	status, err := w.K8sTaskExecutor.CheckTaskStatus(ctx, taskCtx, task)
	if err != nil {
		return status, err
	}

	return status.WithState(taskCtx.GetCustomState()), nil
}

func (w waitableTaskExecutor) GetTaskStatus(ctx context.Context, taskCtx types.TaskContext, r flytek8s.K8sResource) (
	types.TaskStatus, *events.TaskEventInfo, error) {

	status, eventInfo, err := w.containerTaskExecutor.GetTaskStatus(ctx, taskCtx, r)
	if err != nil {
		return status, eventInfo, err
	}

	// TODO: no need to check with the cache anymore since all executions are terminated already. Optimize!
	allWaitables, _, _, err := w.getUpdatedWaitables(ctx, taskCtx)
	if err != nil {
		return types.TaskStatusRetryableFailure(err), nil, nil
	}

	if eventInfo == nil {
		eventInfo = &events.TaskEventInfo{}
	}

	logs := generateWorkflowExecutionLinks(ctx, allWaitables)

	if eventInfo.Logs == nil {
		eventInfo.Logs = make([]*core.TaskLog, 0, len(logs))
	}

	eventInfo.Logs = append(eventInfo.Logs, logs...)

	return status, eventInfo, err
}

func (w waitableTaskExecutor) KillTask(ctx context.Context, taskCtx types.TaskContext, reason string) error {
	switch taskCtx.GetPhase() {
	case types.TaskPhaseQueued:
		// TODO: no need to check with the cache anymore since all executions are terminated already. Optimize!
		allWaitables, _, _, err := w.getUpdatedWaitables(ctx, taskCtx)
		if err != nil {
			return err
		}

		for _, waitable := range allWaitables {
			if !isTerminalWorkflowPhase(waitable.GetPhase()) {
				_, err := w.adminClient.TerminateExecution(ctx, &admin2.ExecutionTerminateRequest{
					Id:    waitable.WfExecId,
					Cause: reason,
				})

				if err != nil {
					return err
				}
			}
		}

		return nil
	default:
		return w.K8sTaskExecutor.KillTask(ctx, taskCtx, reason)
	}
}

func (w waitableTaskExecutor) syncItem(ctx context.Context, obj utils2.CacheItem) (
	utils2.CacheItem, utils2.CacheSyncAction, error) {

	waitable, casted := obj.(*waitableWrapper)
	if !casted {
		return nil, utils2.Unchanged, fmt.Errorf("wrong type. expected %v. got %v", reflect.TypeOf(&waitableWrapper{}), reflect.TypeOf(obj))
	}

	exec, err := w.adminClient.GetExecution(ctx, &admin2.WorkflowExecutionGetRequest{
		Id: waitable.WfExecId,
	})

	if err != nil {
		return nil, utils2.Unchanged, err
	}

	if waitable.Phase != exec.GetClosure().Phase {
		waitable.Phase = exec.GetClosure().Phase
		return waitable, utils2.Update, nil
	}

	return waitable, utils2.Unchanged, nil
}

func newWaitableTaskExecutor(ctx context.Context) (executor *waitableTaskExecutor, err error) {
	waitableExec := &waitableTaskExecutor{
		containerTaskExecutor: containerTaskExecutor{},
	}

	waitableExec.K8sTaskExecutor = flytek8s.NewK8sTaskExecutorForResource(waitableTaskType, &v1.Pod{},
		waitableExec, flytek8s.DefaultInformerResyncDuration)

	waitableExec.adminClient, err = admin.InitializeAdminClientFromConfig(ctx)
	if err != nil {
		return waitableExec, err
	}

	return waitableExec, nil
}

func init() {
	tasksV1.RegisterLoader(func(ctx context.Context) error {
		waitableExec, err := newWaitableTaskExecutor(ctx)
		if err != nil {
			return err
		}

		return tasksV1.RegisterForTaskTypes(waitableExec, waitableTaskType)
	})
}
