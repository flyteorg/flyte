package task

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/ioutils"

	"github.com/golang/protobuf/ptypes"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/event"
	pluginMachinery "github.com/lyft/flyteplugins/go/tasks/pluginmachinery"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/catalog"
	pluginCore "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/io"
	pluginK8s "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/lyft/flytestdlib/contextutils"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/promutils/labeled"
	"github.com/lyft/flytestdlib/storage"
	regErrors "github.com/pkg/errors"

	"github.com/lyft/flytepropeller/pkg/controller/nodes/task/resourcemanager"
	rmConfig "github.com/lyft/flytepropeller/pkg/controller/nodes/task/resourcemanager/config"

	"github.com/lyft/flytepropeller/pkg/controller/executors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/errors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/task/config"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/task/secretmanager"
)

const pluginContextKey = contextutils.Key("plugin")

type metrics struct {
	pluginPanics           labeled.Counter
	unsupportedTaskType    labeled.Counter
	catalogPutFailureCount labeled.Counter
	catalogGetFailureCount labeled.Counter
	catalogPutSuccessCount labeled.Counter
	catalogMissCount       labeled.Counter
	catalogHitCount        labeled.Counter
	pluginExecutionLatency labeled.StopWatch
	pluginQueueLatency     labeled.StopWatch

	// TODO We should have a metric to capture custom state size
	scope promutils.Scope
}

type pluginRequestedTransition struct {
	previouslyObserved bool
	ttype              handler.TransitionType
	pInfo              pluginCore.PhaseInfo
	execInfo           handler.ExecutionInfo
	pluginState        []byte
	pluginStateVersion uint32
}

func (p *pluginRequestedTransition) CacheHit(outputPath storage.DataReference) {
	p.ttype = handler.TransitionTypeEphemeral
	p.pInfo = pluginCore.PhaseInfoSuccess(nil)
	if p.execInfo.TaskNodeInfo == nil {
		p.execInfo.TaskNodeInfo = &handler.TaskNodeInfo{}
	}
	p.execInfo.TaskNodeInfo.CacheHit = true
	p.ObserveSuccess(outputPath)
}

func (p *pluginRequestedTransition) ObservedTransitionAndState(trns pluginCore.Transition, pluginStateVersion uint32, pluginState []byte) {
	p.ttype = ToTransitionType(trns.Type())
	p.pInfo = trns.Info()
	p.pluginState = pluginState
	p.pluginStateVersion = pluginStateVersion
}

func (p *pluginRequestedTransition) ObservedExecutionError(executionError *io.ExecutionError) {
	if executionError.IsRecoverable {
		p.pInfo = pluginCore.PhaseInfoFailed(pluginCore.PhaseRetryableFailure, executionError.ExecutionError, p.pInfo.Info())
	} else {
		p.pInfo = pluginCore.PhaseInfoFailed(pluginCore.PhasePermanentFailure, executionError.ExecutionError, p.pInfo.Info())
	}
}

func (p *pluginRequestedTransition) IsPreviouslyObserved() bool {
	return p.previouslyObserved
}

func (p *pluginRequestedTransition) TransitionPreviouslyRecorded() {
	p.previouslyObserved = true
}

func (p *pluginRequestedTransition) FinalTaskEvent(id *core.TaskExecutionIdentifier, in io.InputFilePaths, out io.OutputFilePaths) (*event.TaskExecutionEvent, error) {
	if p.previouslyObserved {
		return nil, nil
	}

	return ToTaskExecutionEvent(id, in, out, p.pInfo)
}

func (p *pluginRequestedTransition) ObserveSuccess(outputPath storage.DataReference) {
	p.execInfo.OutputInfo = &handler.OutputInfo{OutputURI: outputPath}
}

func (p *pluginRequestedTransition) FinalTransition(ctx context.Context) (handler.Transition, error) {
	switch p.pInfo.Phase() {
	case pluginCore.PhaseSuccess:
		logger.Debugf(ctx, "Transitioning to Success")
		return handler.DoTransition(p.ttype, handler.PhaseInfoSuccess(&p.execInfo)), nil
	case pluginCore.PhaseRetryableFailure:
		logger.Debugf(ctx, "Transitioning to RetryableFailure")
		return handler.DoTransition(p.ttype, handler.PhaseInfoRetryableFailureErr(p.pInfo.Err(), nil)), nil
	case pluginCore.PhasePermanentFailure:
		logger.Debugf(ctx, "Transitioning to Failure")
		return handler.DoTransition(p.ttype, handler.PhaseInfoFailureErr(p.pInfo.Err(), nil)), nil
	case pluginCore.PhaseUndefined:
		return handler.UnknownTransition, fmt.Errorf("error converting plugin phase, received [Undefined]")
	}

	logger.Debugf(ctx, "Task still running")
	return handler.DoTransition(p.ttype, handler.PhaseInfoRunning(nil)), nil
}

// The plugin interface available especially for testing.
type PluginRegistryIface interface {
	GetCorePlugins() []pluginCore.PluginEntry
	GetK8sPlugins() []pluginK8s.PluginEntry
}

type Handler struct {
	catalog         catalog.Client
	asyncCatalog    catalog.AsyncClient
	plugins         map[pluginCore.TaskType]pluginCore.Plugin
	defaultPlugin   pluginCore.Plugin
	metrics         *metrics
	pluginRegistry  PluginRegistryIface
	kubeClient      pluginCore.KubeClient
	secretManager   pluginCore.SecretManager
	resourceManager pluginCore.ResourceManager
	barrierCache    *barrier
	cfg             *config.Config
}

func (t *Handler) FinalizeRequired() bool {
	return true
}

func (t *Handler) setDefault(ctx context.Context, p pluginCore.Plugin) error {
	if t.defaultPlugin != nil {
		logger.Errorf(ctx, "cannot set plugin [%s] as default as plugin [%s] is already configured as default", p.GetID(), t.defaultPlugin.GetID())
	} else {
		logger.Infof(ctx, "Plugin [%s] registered as default plugin", p.GetID())
		t.defaultPlugin = p
	}
	return nil
}

func (t *Handler) Setup(ctx context.Context, sCtx handler.SetupContext) error {
	tSCtx := t.newSetupContext(sCtx)

	// Create a new base resource negotiator
	resourceManagerConfig := rmConfig.GetConfig()
	newResourceManagerBuilder, err := resourcemanager.GetResourceManagerBuilderByType(ctx, resourceManagerConfig.Type, t.metrics.scope)
	if err != nil {
		return err
	}

	// Create the resource negotiator here
	// and then convert it to proxies later and pass them to plugins
	enabledPlugins, err := WranglePluginsAndGenerateFinalList(ctx, &t.cfg.TaskPlugins, t.pluginRegistry)
	if err != nil {
		logger.Errorf(ctx, "Failed to finalize enabled plugins. Err: %s", err)
		return err
	}

	for _, p := range enabledPlugins {
		// create a new resource registrar proxy for each plugin, and pass it into the plugin's LoadPlugin() via a setup context
		pluginResourceNamespacePrefix := pluginCore.ResourceNamespace(newResourceManagerBuilder.GetID()).CreateSubNamespace(pluginCore.ResourceNamespace(p.ID))
		sCtxFinal := newNameSpacedSetupCtx(
			tSCtx, newResourceManagerBuilder.GetResourceRegistrar(pluginResourceNamespacePrefix))
		logger.Infof(ctx, "Loading Plugin [%s] ENABLED", p.ID)
		// cp, err := p.LoadPlugin(ctx, tSCtx)
		cp, err := p.LoadPlugin(ctx, sCtxFinal)
		if err != nil {
			return regErrors.Wrapf(err, "failed to load plugin - %s", p.ID)
		}
		for _, tt := range p.RegisteredTaskTypes {
			logger.Infof(ctx, "Plugin [%s] registered for TaskType [%s]", p.ID, tt)
			t.plugins[tt] = cp
		}
		if p.IsDefault {
			if err := t.setDefault(ctx, cp); err != nil {
				return err
			}
		}
	}

	rm, err := newResourceManagerBuilder.BuildResourceManager(ctx)
	if err != nil {
		logger.Errorf(ctx, "Failed to build a resource manager")
		return err
	}

	t.resourceManager = rm

	return nil
}

func (t Handler) ResolvePlugin(ctx context.Context, ttype string) (pluginCore.Plugin, error) {
	p, ok := t.plugins[ttype]
	if ok {
		logger.Debugf(ctx, "Plugin [%s] resolved for Handler type [%s]", p.GetID(), ttype)
		return p, nil
	}
	if t.defaultPlugin != nil {
		logger.Warnf(ctx, "No plugin found for Handler-type [%s], defaulting to [%s]", ttype, t.defaultPlugin.GetID())
		return t.defaultPlugin, nil
	}
	return nil, fmt.Errorf("no plugin defined for Handler type [%s] and no defaultPlugin configured", ttype)
}

func validateTransition(transition pluginCore.Transition) error {
	if info := transition.Info(); info.Err() == nil && info.Info() == nil {
		return fmt.Errorf("transition doesn't have task info nor an execution error filled [%v]", transition)
	}

	return nil
}

func (t Handler) invokePlugin(ctx context.Context, p pluginCore.Plugin, tCtx *taskExecutionContext, ts handler.TaskNodeState) (*pluginRequestedTransition, error) {
	pluginTrns := &pluginRequestedTransition{}

	trns, err := func() (trns pluginCore.Transition, err error) {
		defer func() {
			if r := recover(); r != nil {
				t.metrics.pluginPanics.Inc(ctx)
				stack := debug.Stack()
				logger.Errorf(ctx, "Panic in plugin[%s]", p.GetID())
				err = fmt.Errorf("panic when executing a plugin [%s]. Stack: [%s]", p.GetID(), string(stack))
				trns = pluginCore.UnknownTransition
			}
		}()
		childCtx := context.WithValue(ctx, pluginContextKey, p.GetID())
		trns, err = p.Handle(childCtx, tCtx)
		return
	}()
	if err != nil {
		logger.Warnf(ctx, "Runtime error from plugin [%s]. Error: %s", p.GetID(), err.Error())
		return nil, regErrors.Wrapf(err, "failed to execute handle for plugin [%s]", p.GetID())
	}

	err = validateTransition(trns)
	if err != nil {
		logger.Errorf(ctx, "Invalid transition from plugin [%s]. Error: %s", p.GetID(), err.Error())
		return nil, regErrors.Wrapf(err, "Invalid transition for plugin [%s]", p.GetID())
	}

	var b []byte
	var v uint32
	if tCtx.psm.newState != nil {
		b = tCtx.psm.newState.Bytes()
		v = uint32(tCtx.psm.newStateVersion)
	} else {
		// New state was not mutated, so we should write back the existing state
		b = ts.PluginState
		v = ts.PluginPhaseVersion
	}
	pluginTrns.ObservedTransitionAndState(trns, v, b)

	// Emit the queue latency if the task has just transitioned from Queued to Running.
	if ts.PluginPhase == pluginCore.PhaseQueued &&
		(pluginTrns.pInfo.Phase() == pluginCore.PhaseInitializing || pluginTrns.pInfo.Phase() == pluginCore.PhaseRunning) {
		if !ts.LastPhaseUpdatedAt.IsZero() {
			t.metrics.pluginQueueLatency.Observe(ctx, ts.LastPhaseUpdatedAt, time.Now())
		}
	}

	if pluginTrns.pInfo.Phase() == ts.PluginPhase {
		if pluginTrns.pInfo.Version() == ts.PluginPhaseVersion {
			logger.Debugf(ctx, "p+Version previously seen .. no event will be sent")
			pluginTrns.TransitionPreviouslyRecorded()
			return pluginTrns, nil
		}
		if pluginTrns.pInfo.Version() > uint32(t.cfg.MaxPluginPhaseVersions) {
			logger.Errorf(ctx, "Too many Plugin p versions for plugin [%s]. p versions [%d/%d]", p.GetID(), pluginTrns.pInfo.Version(), t.cfg.MaxPluginPhaseVersions)
			pluginTrns.ObservedExecutionError(&io.ExecutionError{
				ExecutionError: &core.ExecutionError{
					Code: "TooManyPluginPhaseVersions",
					Message: fmt.Sprintf("Total number of phase versions exceeded for phase [%s] in Plugin "+
						"[%s]. Attempted to set version to [%v], max allowed [%d]",
						pluginTrns.pInfo.Phase().String(), p.GetID(), pluginTrns.pInfo.Version(), t.cfg.MaxPluginPhaseVersions),
				},
				IsRecoverable: false,
			})
			return pluginTrns, nil
		}
	}

	if pluginTrns.pInfo.Phase() == pluginCore.PhaseSuccess {
		// -------------------------------------
		// TODO: @kumare create Issue# Remove the code after we use closures to handle dynamic nodes
		// This code only exists to support Dynamic tasks. Eventually dynamic tasks will use closure nodes to execute
		// Until then we have to check if the Handler executed resulted in a dynamic node being generated, if so, then
		// we will not check for outputs or call onTaskSuccess. The reason is that outputs have not yet been materialized.
		// Outputs for the parent node will only get generated after the subtasks complete. We have to wait for the completion
		// the dynamic.handler will call onTaskSuccess for the parent node

		f, err := NewRemoteFutureFileReader(ctx, tCtx.ow.GetOutputPrefixPath(), tCtx.DataStore())
		if err != nil {
			return nil, regErrors.Wrapf(err, "failed to create remote file reader")
		}
		if ok, err := f.Exists(ctx); err != nil {
			logger.Errorf(ctx, "failed to check existence of futures file")
			return nil, regErrors.Wrapf(err, "failed to check existence of futures file")
		} else if ok {
			logger.Infof(ctx, "Futures file exists, this is a dynamic parent-Handler will not run onTaskSuccess")
			return pluginTrns, nil
		}
		// End TODO
		// -------------------------------------
		logger.Debugf(ctx, "Task success detected, calling on Task success")
		outputCommitter := ioutils.NewRemoteFileOutputWriter(ctx, tCtx.DataStore(), tCtx.OutputWriter())
		execID := tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetID()
		ee, err := t.ValidateOutputAndCacheAdd(ctx, tCtx.NodeID(), tCtx.InputReader(), tCtx.ow.GetReader(), outputCommitter, tCtx.tr, catalog.Metadata{
			TaskExecutionIdentifier: &execID,
		})
		if err != nil {
			return nil, err
		}
		if ee != nil {
			pluginTrns.ObservedExecutionError(ee)
		} else {
			pluginTrns.ObserveSuccess(tCtx.ow.GetOutputPath())
		}
	}

	return pluginTrns, nil
}

func (t Handler) Handle(ctx context.Context, nCtx handler.NodeExecutionContext) (handler.Transition, error) {
	ttype := nCtx.TaskReader().GetTaskType()
	ctx = contextutils.WithTaskType(ctx, ttype)
	p, err := t.ResolvePlugin(ctx, ttype)
	if err != nil {
		return handler.UnknownTransition, errors.Wrapf(errors.UnsupportedTaskTypeError, nCtx.NodeID(), err, "unable to resolve plugin")
	}

	checkCatalog := !p.GetProperties().DisableNodeLevelCaching
	if !checkCatalog {
		logger.Debug(ctx, "Node level caching is disabled. Skipping catalog read.")
	}

	tCtx, err := t.newTaskExecutionContext(ctx, nCtx, p.GetID())
	if err != nil {
		return handler.UnknownTransition, errors.Wrapf(errors.IllegalStateError, nCtx.NodeID(), err, "unable to create Handler execution context")
	}

	ts := nCtx.NodeStateReader().GetTaskNodeState()

	var pluginTrns *pluginRequestedTransition

	// NOTE: Ideally we should use a taskExecution state for this handler. But, doing that will make it completely backwards incompatible
	// So now we will derive this from the plugin phase
	// TODO @kumare re-evaluate this decision

	// STEP 1: Check Cache
	if ts.PluginPhase == pluginCore.PhaseUndefined && checkCatalog {
		// This is assumed to be first time. we will check catalog and call handle
		if ok, err := t.CheckCatalogCache(ctx, tCtx.tr, nCtx.InputReader(), tCtx.ow); err != nil {
			logger.Errorf(ctx, "failed to check catalog cache with error")
			return handler.UnknownTransition, err
		} else if ok {
			r := tCtx.ow.GetReader()
			if r != nil {
				// TODO @kumare this can be optimized, if we have paths then the reader could be pipelined to a sink
				o, ee, err := r.Read(ctx)
				if err != nil {
					logger.Errorf(ctx, "failed to read from catalog, err: %s", err.Error())
					return handler.UnknownTransition, err
				}
				if ee != nil {
					logger.Errorf(ctx, "got execution error from catalog output reader? This should not happen, err: %s", ee.String())
					return handler.UnknownTransition, errors.Errorf(errors.IllegalStateError, nCtx.NodeID(), "execution error from a cache output, bad state: %s", ee.String())
				}
				if err := nCtx.DataStore().WriteProtobuf(ctx, tCtx.ow.GetOutputPath(), storage.Options{}, o); err != nil {
					logger.Errorf(ctx, "failed to write cached value to datastore, err: %s", err.Error())
					return handler.UnknownTransition, err
				}
				pluginTrns = &pluginRequestedTransition{}
				pluginTrns.CacheHit(tCtx.ow.GetOutputPath())
			} else {
				logger.Errorf(ctx, "no output reader found after a catalog cache hit!")
			}
		}
	}

	barrierTick := uint32(0)
	// STEP 2: If no cache-hit, then lets invoke the plugin and wait for a transition out of undefined
	if pluginTrns == nil {
		prevBarrier := t.barrierCache.GetPreviousBarrierTransition(ctx, tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName())
		// Lets start with the current barrierTick (the value to be stored) same as the barrierTick in the cache
		barrierTick = prevBarrier.BarrierClockTick
		// Lets check if this value in cache is less than or equal to one in the store
		if barrierTick <= ts.BarrierClockTick {
			var err error
			pluginTrns, err = t.invokePlugin(ctx, p, tCtx, ts)
			if err != nil {
				return handler.UnknownTransition, errors.Wrapf(errors.RuntimeExecutionError, nCtx.NodeID(), err, "failed during plugin execution")
			}
			if pluginTrns.IsPreviouslyObserved() {
				logger.Debugf(ctx, "No state change for Task, previously observed same transition. Short circuiting.")
				return pluginTrns.FinalTransition(ctx)
			}
			// Now no matter what we should update the barrierTick (stored in state)
			// This is because the state is ahead of the inmemory representation
			// This can happen in the case where the process restarted or the barrier cache got reset
			barrierTick = ts.BarrierClockTick
			// Now if the transition is of type barrier, lets tick the clock by one from the prev known value
			// store that in the cache
			if pluginTrns.ttype == handler.TransitionTypeBarrier {
				logger.Infof(ctx, "Barrier transition observed for Plugin [%s], TaskExecID [%s]. recording: [%s]", p.GetID(), tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), pluginTrns.pInfo.String())
				barrierTick = barrierTick + 1
				t.barrierCache.RecordBarrierTransition(ctx, tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), BarrierTransition{
					BarrierClockTick: barrierTick,
					CallLog: PluginCallLog{
						PluginTransition: pluginTrns,
					},
				})

			}
		} else {
			// Barrier tick will remain to be the one in cache.
			// Now it may happen that the cache may get reset before we store the barrier tick
			// this will cause us to lose that information and potentially replaying.
			logger.Infof(ctx, "Replaying Barrier transition for cache tick [%d] < stored tick [%d], Plugin [%s], TaskExecID [%s]. recording: [%s]", barrierTick, ts.BarrierClockTick, p.GetID(), tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName(), prevBarrier.CallLog.PluginTransition.pInfo.String())
			pluginTrns = prevBarrier.CallLog.PluginTransition
		}
	}

	// STEP 3: Sanity check
	if pluginTrns == nil {
		// Still nil, this should never happen!!!
		return handler.UnknownTransition, errors.Errorf(errors.IllegalStateError, nCtx.NodeID(), "plugin transition is not observed and no error as well.")
	}

	execID := tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetID()
	// STEP 4: Send buffered events!
	logger.Debugf(ctx, "Sending buffered Task events.")
	for _, ev := range tCtx.ber.GetAll(ctx) {
		evInfo, err := ToTaskExecutionEvent(&execID, nCtx.InputReader(), tCtx.ow, ev)
		if err != nil {
			return handler.UnknownTransition, err
		}
		if err := nCtx.EventsRecorder().RecordTaskEvent(ctx, evInfo); err != nil {
			logger.Errorf(ctx, "Event recording failed for Plugin [%s], eventPhase [%s], error :%s", p.GetID(), evInfo.Phase.String(), err.Error())
			// Check for idempotency
			// Check for terminate state error
			return handler.UnknownTransition, err
		}
	}

	// STEP 5: Send Transition events
	logger.Debugf(ctx, "Sending transition event for plugin phase [%s]", pluginTrns.pInfo.Phase().String())
	evInfo, err := pluginTrns.FinalTaskEvent(&execID, nCtx.InputReader(), tCtx.ow)
	if err != nil {
		logger.Errorf(ctx, "failed to convert plugin transition to TaskExecutionEvent. Error: %s", err.Error())
		return handler.UnknownTransition, err
	}
	if evInfo != nil {
		if err := nCtx.EventsRecorder().RecordTaskEvent(ctx, evInfo); err != nil {
			// Check for idempotency
			// Check for terminate state error
			logger.Errorf(ctx, "failed to send event to Admin. error: %s", err.Error())
			return handler.UnknownTransition, err
		}
	} else {
		logger.Debugf(ctx, "Received no event to record.")
	}

	// STEP 6: Persist the plugin state
	err = nCtx.NodeStateWriter().PutTaskNodeState(handler.TaskNodeState{
		PluginState:        pluginTrns.pluginState,
		PluginStateVersion: pluginTrns.pluginStateVersion,
		PluginPhase:        pluginTrns.pInfo.Phase(),
		PluginPhaseVersion: pluginTrns.pInfo.Version(),
		BarrierClockTick:   barrierTick,
		LastPhaseUpdatedAt: time.Now(),
	})
	if err != nil {
		logger.Errorf(ctx, "Failed to store TaskNode state, err :%s", err.Error())
		return handler.UnknownTransition, err
	}

	return pluginTrns.FinalTransition(ctx)
}

func (t Handler) Abort(ctx context.Context, nCtx handler.NodeExecutionContext, reason string) error {
	currentPhase := nCtx.NodeStateReader().GetTaskNodeState().PluginPhase
	logger.Debugf(ctx, "Abort invoked with phase [%v]", currentPhase)

	if currentPhase.IsTerminal() {
		logger.Debugf(ctx, "Returning immediately from Abort since task is already in terminal phase.", currentPhase)
		return nil
	}

	ttype := nCtx.TaskReader().GetTaskType()
	p, err := t.ResolvePlugin(ctx, ttype)
	if err != nil {
		return errors.Wrapf(errors.UnsupportedTaskTypeError, nCtx.NodeID(), err, "unable to resolve plugin")
	}

	tCtx, err := t.newTaskExecutionContext(ctx, nCtx, p.GetID())
	if err != nil {
		return errors.Wrapf(errors.IllegalStateError, nCtx.NodeID(), err, "unable to create Handler execution context")
	}

	err = func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				t.metrics.pluginPanics.Inc(ctx)
				stack := debug.Stack()
				logger.Errorf(ctx, "Panic in plugin.Abort for TaskType [%s]", tCtx.tr.GetTaskType())
				err = fmt.Errorf("panic when executing a plugin for TaskType [%s]. Stack: [%s]", tCtx.tr.GetTaskType(), string(stack))
			}
		}()

		childCtx := context.WithValue(ctx, pluginContextKey, p.GetID())
		err = p.Abort(childCtx, tCtx)
		return
	}()

	if err != nil {
		logger.Errorf(ctx, "Abort failed when calling plugin abort.")
		return err
	}
	taskExecID := tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetID()
	evRecorder := nCtx.EventsRecorder()
	if err := evRecorder.RecordTaskEvent(ctx, &event.TaskExecutionEvent{
		TaskId:                taskExecID.TaskId,
		ParentNodeExecutionId: taskExecID.NodeExecutionId,
		RetryAttempt:          nCtx.CurrentAttempt(),
		Phase:                 core.TaskExecution_ABORTED,
		OccurredAt:            ptypes.TimestampNow(),
		OutputResult: &event.TaskExecutionEvent_Error{
			Error: &core.ExecutionError{
				Code:    "Task Aborted",
				Message: reason,
			}},
	}); err != nil {
		logger.Errorf(ctx, "failed to send event to Admin. error: %s", err.Error())
		return err
	}
	return nil
}

func (t Handler) Finalize(ctx context.Context, nCtx handler.NodeExecutionContext) error {
	logger.Debugf(ctx, "Finalize invoked.")
	ttype := nCtx.TaskReader().GetTaskType()
	p, err := t.ResolvePlugin(ctx, ttype)
	if err != nil {
		return errors.Wrapf(errors.UnsupportedTaskTypeError, nCtx.NodeID(), err, "unable to resolve plugin")
	}

	tCtx, err := t.newTaskExecutionContext(ctx, nCtx, p.GetID())
	if err != nil {
		return errors.Wrapf(errors.IllegalStateError, nCtx.NodeID(), err, "unable to create Handler execution context")
	}

	return func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				t.metrics.pluginPanics.Inc(ctx)
				stack := debug.Stack()
				logger.Errorf(ctx, "Panic in plugin.Finalize for TaskType [%s]", tCtx.tr.GetTaskType())
				err = fmt.Errorf("panic when executing a plugin for TaskType [%s]. Stack: [%s]", tCtx.tr.GetTaskType(), string(stack))
			}
		}()
		childCtx := context.WithValue(ctx, pluginContextKey, p.GetID())
		err = p.Finalize(childCtx, tCtx)
		return
	}()
}

func New(ctx context.Context, kubeClient executors.Client, client catalog.Client, scope promutils.Scope) (*Handler, error) {
	// TODO New should take a pointer
	async, err := catalog.NewAsyncClient(client, *catalog.GetConfig(), scope.NewSubScope("async_catalog"))
	if err != nil {
		return nil, err
	}

	if err = async.Start(ctx); err != nil {
		return nil, err
	}

	cfg := config.GetConfig()
	return &Handler{
		pluginRegistry: pluginMachinery.PluginRegistry(),
		plugins:        make(map[pluginCore.TaskType]pluginCore.Plugin),
		metrics: &metrics{
			pluginPanics:           labeled.NewCounter("plugin_panic", "Task plugin paniced when trying to execute a Handler.", scope),
			unsupportedTaskType:    labeled.NewCounter("unsupported_tasktype", "No Handler plugin configured for Handler type", scope),
			catalogHitCount:        labeled.NewCounter("discovery_hit_count", "Task cached in Discovery", scope),
			catalogMissCount:       labeled.NewCounter("discovery_miss_count", "Task not cached in Discovery", scope),
			catalogPutSuccessCount: labeled.NewCounter("discovery_put_success_count", "Discovery Put success count", scope),
			catalogPutFailureCount: labeled.NewCounter("discovery_put_failure_count", "Discovery Put failure count", scope),
			catalogGetFailureCount: labeled.NewCounter("discovery_get_failure_count", "Discovery Get faillure count", scope),
			pluginExecutionLatency: labeled.NewStopWatch("plugin_exec_latecny", "Time taken to invoke plugin for one round", time.Microsecond, scope),
			pluginQueueLatency:     labeled.NewStopWatch("plugin_queue_latecny", "Time spent by plugin in queued phase", time.Microsecond, scope),
			scope:                  scope,
		},
		kubeClient:      kubeClient,
		catalog:         client,
		asyncCatalog:    async,
		resourceManager: nil,
		secretManager:   secretmanager.NewFileEnvSecretManager(secretmanager.GetConfig()),
		barrierCache:    newLRUBarrier(ctx, cfg.BarrierConfig),
		cfg:             cfg,
	}, nil
}
