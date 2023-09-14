package array

import (
	"context"
	"fmt"
	"math"
	"strconv"

	idlcore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/errorcollector"

	"github.com/flyteorg/flytepropeller/events"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/compiler/validators"
	"github.com/flyteorg/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flytepropeller/pkg/controller/executors"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/errors"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/handler"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/interfaces"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/task"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/k8s"

	"github.com/flyteorg/flytestdlib/bitarray"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/storage"
)

var (
	nilLiteral = &idlcore.Literal{
		Value: &idlcore.Literal_Scalar{
			Scalar: &idlcore.Scalar{
				Value: &idlcore.Scalar_NoneType{
					NoneType: &idlcore.Void{},
				},
			},
		},
	}
)

//go:generate mockery -all -case=underscore

// arrayNodeHandler is a handle implementation for processing array nodes
type arrayNodeHandler struct {
	eventConfig                *config.EventConfig
	metrics                    metrics
	nodeExecutor               interfaces.Node
	pluginStateBytesNotStarted []byte
	pluginStateBytesStarted    []byte
}

// metrics encapsulates the prometheus metrics for this handler
type metrics struct {
	scope promutils.Scope
}

// newMetrics initializes a new metrics struct
func newMetrics(scope promutils.Scope) metrics {
	return metrics{
		scope: scope,
	}
}

// Abort stops the array node defined in the NodeExecutionContext
func (a *arrayNodeHandler) Abort(ctx context.Context, nCtx interfaces.NodeExecutionContext, reason string) error {
	arrayNode := nCtx.Node().GetArrayNode()
	arrayNodeState := nCtx.NodeStateReader().GetArrayNodeState()

	externalResources := make([]*event.ExternalResourceInfo, 0, len(arrayNodeState.SubNodePhases.GetItems()))
	messageCollector := errorcollector.NewErrorMessageCollector()
	switch arrayNodeState.Phase {
	case v1alpha1.ArrayNodePhaseExecuting, v1alpha1.ArrayNodePhaseFailing:
		currentParallelism := uint32(0)
		for i, nodePhaseUint64 := range arrayNodeState.SubNodePhases.GetItems() {
			nodePhase := v1alpha1.NodePhase(nodePhaseUint64)

			// do not process nodes that have not started or are in a terminal state
			if nodePhase == v1alpha1.NodePhaseNotYetStarted || isTerminalNodePhase(nodePhase) {
				continue
			}

			// create array contexts
			arrayNodeExecutor, arrayExecutionContext, arrayDAGStructure, arrayNodeLookup, subNodeSpec, _, _, err :=
				a.buildArrayNodeContext(ctx, nCtx, &arrayNodeState, arrayNode, i, &currentParallelism)
			if err != nil {
				return err
			}

			// abort subNode
			err = arrayNodeExecutor.AbortHandler(ctx, arrayExecutionContext, arrayDAGStructure, arrayNodeLookup, subNodeSpec, reason)
			if err != nil {
				messageCollector.Collect(i, err.Error())
			} else {
				externalResources = append(externalResources, &event.ExternalResourceInfo{
					ExternalId:   buildSubNodeID(nCtx, i, 0),
					Index:        uint32(i),
					Logs:         nil,
					RetryAttempt: 0,
					Phase:        idlcore.TaskExecution_ABORTED,
				})
			}
		}
	}

	if messageCollector.Length() > 0 {
		return fmt.Errorf(messageCollector.Summary(events.MaxErrorMessageLength))
	}

	// update aborted state for subNodes
	taskExecutionEvent, err := buildTaskExecutionEvent(ctx, nCtx, idlcore.TaskExecution_ABORTED, 0, externalResources)
	if err != nil {
		return err
	}

	if err := nCtx.EventsRecorder().RecordTaskEvent(ctx, taskExecutionEvent, a.eventConfig); err != nil {
		logger.Errorf(ctx, "ArrayNode event recording failed: [%s]", err.Error())
		return err
	}

	return nil
}

// Finalize completes the array node defined in the NodeExecutionContext
func (a *arrayNodeHandler) Finalize(ctx context.Context, nCtx interfaces.NodeExecutionContext) error {
	arrayNode := nCtx.Node().GetArrayNode()
	arrayNodeState := nCtx.NodeStateReader().GetArrayNodeState()

	messageCollector := errorcollector.NewErrorMessageCollector()
	switch arrayNodeState.Phase {
	case v1alpha1.ArrayNodePhaseExecuting, v1alpha1.ArrayNodePhaseFailing, v1alpha1.ArrayNodePhaseSucceeding:
		currentParallelism := uint32(0)
		for i, nodePhaseUint64 := range arrayNodeState.SubNodePhases.GetItems() {
			nodePhase := v1alpha1.NodePhase(nodePhaseUint64)

			// do not process nodes that have not started or are in a terminal state
			if nodePhase == v1alpha1.NodePhaseNotYetStarted || isTerminalNodePhase(nodePhase) {
				continue
			}

			// create array contexts
			arrayNodeExecutor, arrayExecutionContext, arrayDAGStructure, arrayNodeLookup, subNodeSpec, _, _, err :=
				a.buildArrayNodeContext(ctx, nCtx, &arrayNodeState, arrayNode, i, &currentParallelism)
			if err != nil {
				return err
			}

			// finalize subNode
			err = arrayNodeExecutor.FinalizeHandler(ctx, arrayExecutionContext, arrayDAGStructure, arrayNodeLookup, subNodeSpec)
			if err != nil {
				messageCollector.Collect(i, err.Error())
			}
		}
	}

	if messageCollector.Length() > 0 {
		return fmt.Errorf(messageCollector.Summary(events.MaxErrorMessageLength))
	}

	return nil
}

// FinalizeRequired defines whether or not this handler requires finalize to be called on node
// completion
func (a *arrayNodeHandler) FinalizeRequired() bool {
	// must return true because we can't determine if finalize is required for the subNode
	return true
}

// Handle is responsible for transitioning and reporting node state to complete the node defined
// by the NodeExecutionContext
func (a *arrayNodeHandler) Handle(ctx context.Context, nCtx interfaces.NodeExecutionContext) (handler.Transition, error) {
	arrayNode := nCtx.Node().GetArrayNode()
	arrayNodeState := nCtx.NodeStateReader().GetArrayNodeState()
	currentArrayNodePhase := arrayNodeState.Phase

	var externalResources []*event.ExternalResourceInfo
	taskPhaseVersion := arrayNodeState.TaskPhaseVersion

	switch currentArrayNodePhase {
	case v1alpha1.ArrayNodePhaseNone:
		// identify and validate array node input value lengths
		literalMap, err := nCtx.InputReader().Get(ctx)
		if err != nil {
			return handler.UnknownTransition, err
		}

		size := -1
		for _, variable := range literalMap.Literals {
			literalType := validators.LiteralTypeForLiteral(variable)
			switch literalType.Type.(type) {
			case *idlcore.LiteralType_CollectionType:
				collectionLength := len(variable.GetCollection().Literals)

				if size == -1 {
					size = collectionLength
				} else if size != collectionLength {
					return handler.DoTransition(handler.TransitionTypeEphemeral,
						handler.PhaseInfoFailure(idlcore.ExecutionError_USER, errors.InvalidArrayLength,
							fmt.Sprintf("input arrays have different lengths: expecting '%d' found '%d'", size, collectionLength), nil),
					), nil
				}
			}
		}

		if size == -1 {
			return handler.DoTransition(handler.TransitionTypeEphemeral,
				handler.PhaseInfoFailure(idlcore.ExecutionError_USER, errors.InvalidArrayLength, "no input array provided", nil),
			), nil
		}

		// initialize ArrayNode state
		maxAttempts := task.DefaultMaxAttempts
		subNodeSpec := *arrayNode.GetSubNodeSpec()
		if subNodeSpec.GetRetryStrategy() != nil && subNodeSpec.GetRetryStrategy().MinAttempts != nil {
			maxAttempts = *subNodeSpec.GetRetryStrategy().MinAttempts
		}

		for _, item := range []struct {
			arrayReference *bitarray.CompactArray
			maxValue       int
		}{
			// we use NodePhaseRecovered for the `maxValue` of `SubNodePhases` because `Phase` is
			// defined as an `iota` so it is impossible to programmatically get largest value
			{arrayReference: &arrayNodeState.SubNodePhases, maxValue: int(v1alpha1.NodePhaseRecovered)},
			{arrayReference: &arrayNodeState.SubNodeTaskPhases, maxValue: len(core.Phases) - 1},
			{arrayReference: &arrayNodeState.SubNodeRetryAttempts, maxValue: maxAttempts},
			{arrayReference: &arrayNodeState.SubNodeSystemFailures, maxValue: maxAttempts},
		} {

			*item.arrayReference, err = bitarray.NewCompactArray(uint(size), bitarray.Item(item.maxValue))
			if err != nil {
				return handler.UnknownTransition, err
			}
		}

		// initialize externalResources
		externalResources = make([]*event.ExternalResourceInfo, 0, size)
		for i := 0; i < size; i++ {
			externalResources = append(externalResources, &event.ExternalResourceInfo{
				ExternalId:   buildSubNodeID(nCtx, i, 0),
				Index:        uint32(i),
				Logs:         nil,
				RetryAttempt: 0,
				Phase:        idlcore.TaskExecution_QUEUED,
			})
		}

		// transition ArrayNode to `ArrayNodePhaseExecuting`
		arrayNodeState.Phase = v1alpha1.ArrayNodePhaseExecuting
	case v1alpha1.ArrayNodePhaseExecuting:
		// process array node subNodes
		currentParallelism := uint32(0)
		messageCollector := errorcollector.NewErrorMessageCollector()
		externalResources = make([]*event.ExternalResourceInfo, 0)
		for i, nodePhaseUint64 := range arrayNodeState.SubNodePhases.GetItems() {
			nodePhase := v1alpha1.NodePhase(nodePhaseUint64)

			// do not process nodes in terminal state
			if isTerminalNodePhase(nodePhase) {
				continue
			}

			// create array contexts
			arrayNodeExecutor, arrayExecutionContext, arrayDAGStructure, arrayNodeLookup, subNodeSpec, subNodeStatus, arrayEventRecorder, err :=
				a.buildArrayNodeContext(ctx, nCtx, &arrayNodeState, arrayNode, i, &currentParallelism)
			if err != nil {
				return handler.UnknownTransition, err
			}

			// execute subNode
			_, err = arrayNodeExecutor.RecursiveNodeHandler(ctx, arrayExecutionContext, arrayDAGStructure, arrayNodeLookup, subNodeSpec)
			if err != nil {
				return handler.UnknownTransition, err
			}

			// capture subNode error if exists
			if subNodeStatus.Error != nil {
				messageCollector.Collect(i, subNodeStatus.Error.Message)
			}

			// process events
			cacheStatus := idlcore.CatalogCacheStatus_CACHE_DISABLED
			for _, nodeExecutionEvent := range arrayEventRecorder.NodeEvents() {
				switch target := nodeExecutionEvent.TargetMetadata.(type) {
				case *event.NodeExecutionEvent_TaskNodeMetadata:
					if target.TaskNodeMetadata != nil {
						cacheStatus = target.TaskNodeMetadata.CacheStatus
					}
				}
			}

			retryAttempt := subNodeStatus.GetAttempts()

			// fastcache will not emit task events for cache hits. we need to manually detect a
			// transition to `SUCCEEDED` and add an `ExternalResourceInfo` for it.
			if cacheStatus == idlcore.CatalogCacheStatus_CACHE_HIT && len(arrayEventRecorder.TaskEvents()) == 0 {
				externalResources = append(externalResources, &event.ExternalResourceInfo{
					ExternalId:   buildSubNodeID(nCtx, i, retryAttempt),
					Index:        uint32(i),
					RetryAttempt: retryAttempt,
					Phase:        idlcore.TaskExecution_SUCCEEDED,
					CacheStatus:  cacheStatus,
				})
			}

			for _, taskExecutionEvent := range arrayEventRecorder.TaskEvents() {
				for _, log := range taskExecutionEvent.Logs {
					log.Name = fmt.Sprintf("%s-%d", log.Name, i)
				}

				externalResources = append(externalResources, &event.ExternalResourceInfo{
					ExternalId:   buildSubNodeID(nCtx, i, retryAttempt),
					Index:        uint32(i),
					Logs:         taskExecutionEvent.Logs,
					RetryAttempt: retryAttempt,
					Phase:        taskExecutionEvent.Phase,
					CacheStatus:  cacheStatus,
				})
			}

			// update subNode state
			arrayNodeState.SubNodePhases.SetItem(i, uint64(subNodeStatus.GetPhase()))
			if subNodeStatus.GetTaskNodeStatus() == nil {
				// resetting task phase because during retries we clear the GetTaskNodeStatus
				arrayNodeState.SubNodeTaskPhases.SetItem(i, uint64(0))
			} else {
				arrayNodeState.SubNodeTaskPhases.SetItem(i, uint64(subNodeStatus.GetTaskNodeStatus().GetPhase()))
			}
			arrayNodeState.SubNodeRetryAttempts.SetItem(i, uint64(subNodeStatus.GetAttempts()))
			arrayNodeState.SubNodeSystemFailures.SetItem(i, uint64(subNodeStatus.GetSystemFailures()))
		}

		// process phases of subNodes to determine overall `ArrayNode` phase
		successCount := 0
		failedCount := 0
		failingCount := 0
		runningCount := 0
		for _, nodePhaseUint64 := range arrayNodeState.SubNodePhases.GetItems() {
			nodePhase := v1alpha1.NodePhase(nodePhaseUint64)
			switch nodePhase {
			case v1alpha1.NodePhaseSucceeded, v1alpha1.NodePhaseRecovered, v1alpha1.NodePhaseSkipped:
				successCount++
			case v1alpha1.NodePhaseFailing:
				failingCount++
			case v1alpha1.NodePhaseFailed, v1alpha1.NodePhaseTimedOut:
				failedCount++
			default:
				runningCount++
			}
		}

		// calculate minimum number of successes to succeed the ArrayNode
		minSuccesses := len(arrayNodeState.SubNodePhases.GetItems())
		if arrayNode.GetMinSuccesses() != nil {
			minSuccesses = int(*arrayNode.GetMinSuccesses())
		} else if minSuccessRatio := arrayNode.GetMinSuccessRatio(); minSuccessRatio != nil {
			minSuccesses = int(math.Ceil(float64(*minSuccessRatio) * float64(minSuccesses)))
		}

		// if there is a failing node set the error message if it has not been previous set
		if failingCount > 0 && arrayNodeState.Error == nil {
			arrayNodeState.Error = &idlcore.ExecutionError{
				Message: messageCollector.Summary(events.MaxErrorMessageLength),
			}
		}

		if len(arrayNodeState.SubNodePhases.GetItems())-failedCount < minSuccesses {
			// no chance to reach the mininum number of successes
			arrayNodeState.Phase = v1alpha1.ArrayNodePhaseFailing
		} else if successCount >= minSuccesses && runningCount == 0 {
			// wait until all tasks have completed before declaring success
			arrayNodeState.Phase = v1alpha1.ArrayNodePhaseSucceeding
		}
	case v1alpha1.ArrayNodePhaseFailing:
		if err := a.Abort(ctx, nCtx, "ArrayNodeFailing"); err != nil {
			return handler.UnknownTransition, err
		}

		// fail with reported error if one exists
		if arrayNodeState.Error != nil {
			return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailureErr(arrayNodeState.Error, nil)), nil
		}

		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(
			idlcore.ExecutionError_UNKNOWN,
			"ArrayNodeFailing",
			"Unknown reason",
			nil,
		)), nil
	case v1alpha1.ArrayNodePhaseSucceeding:
		outputLiterals := make(map[string]*idlcore.Literal)
		for i, nodePhaseUint64 := range arrayNodeState.SubNodePhases.GetItems() {
			nodePhase := v1alpha1.NodePhase(nodePhaseUint64)

			if nodePhase != v1alpha1.NodePhaseSucceeded {
				// retrieve output variables from task template
				var outputVariables map[string]*idlcore.Variable
				task, err := nCtx.ExecutionContext().GetTask(*arrayNode.GetSubNodeSpec().TaskRef)
				if err != nil {
					// Should never happen
					return handler.UnknownTransition, err
				}

				if task.CoreTask() != nil && task.CoreTask().Interface != nil && task.CoreTask().Interface.Outputs != nil {
					outputVariables = task.CoreTask().Interface.Outputs.Variables
				}

				// append nil literal for all output variables
				for name := range outputVariables {
					appendLiteral(name, nilLiteral, outputLiterals, len(arrayNodeState.SubNodePhases.GetItems()))
				}
			} else {
				// initialize subNode reader
				currentAttempt := uint32(arrayNodeState.SubNodeRetryAttempts.GetItem(i))
				subDataDir, subOutputDir, err := constructOutputReferences(ctx, nCtx, strconv.Itoa(i), strconv.Itoa(int(currentAttempt)))
				if err != nil {
					return handler.UnknownTransition, err
				}

				// checkpoint paths are not computed here because this function is only called when writing
				// existing cached outputs. if this functionality changes this will need to be revisited.
				outputPaths := ioutils.NewCheckpointRemoteFilePaths(ctx, nCtx.DataStore(), subOutputDir, ioutils.NewRawOutputPaths(ctx, subDataDir), "")
				reader := ioutils.NewRemoteFileOutputReader(ctx, nCtx.DataStore(), outputPaths, nCtx.MaxDatasetSizeBytes())

				// read outputs
				outputs, executionErr, err := reader.Read(ctx)
				if err != nil {
					return handler.UnknownTransition, err
				} else if executionErr != nil {
					return handler.UnknownTransition, errors.Errorf(errors.IllegalStateError, nCtx.NodeID(),
						"execution error ArrayNode output, bad state: %s", executionErr.String())
				}

				// copy individual subNode output literals into a collection of output literals
				for name, literal := range outputs.GetLiterals() {
					appendLiteral(name, literal, outputLiterals, len(arrayNodeState.SubNodePhases.GetItems()))
				}
			}
		}

		outputLiteralMap := &idlcore.LiteralMap{
			Literals: outputLiterals,
		}

		outputFile := v1alpha1.GetOutputsFile(nCtx.NodeStatus().GetOutputDir())
		if err := nCtx.DataStore().WriteProtobuf(ctx, outputFile, storage.Options{}, outputLiteralMap); err != nil {
			return handler.UnknownTransition, err
		}

		return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoSuccess(
			&handler.ExecutionInfo{
				OutputInfo: &handler.OutputInfo{
					OutputURI: outputFile,
				},
			},
		)), nil
	default:
		return handler.UnknownTransition, errors.Errorf(errors.IllegalStateError, nCtx.NodeID(), "invalid ArrayNode phase %+v", arrayNodeState.Phase)
	}

	// if there were changes to subNode status externalResources will be populated and must be
	// reported to admin through a TaskExecutionEvent.
	if len(externalResources) > 0 {
		// determine task phase from ArrayNodePhase
		taskPhase := idlcore.TaskExecution_UNDEFINED
		switch currentArrayNodePhase {
		case v1alpha1.ArrayNodePhaseNone:
			taskPhase = idlcore.TaskExecution_QUEUED
		case v1alpha1.ArrayNodePhaseExecuting:
			taskPhase = idlcore.TaskExecution_RUNNING
		case v1alpha1.ArrayNodePhaseSucceeding:
			taskPhase = idlcore.TaskExecution_SUCCEEDED
		case v1alpha1.ArrayNodePhaseFailing:
			taskPhase = idlcore.TaskExecution_FAILED
		}

		// need to increment taskPhaseVersion if arrayNodeState.Phase does not change, otherwise
		// reset to 0. by incrementing this always we report an event and ensure processing
		// everytime the ArrayNode is evaluated. if this overhead becomes too large, we will need
		// to revisit and only increment when any subNode state changes.
		if currentArrayNodePhase != arrayNodeState.Phase {
			arrayNodeState.TaskPhaseVersion = 0
		} else {
			arrayNodeState.TaskPhaseVersion = taskPhaseVersion + 1
		}

		taskExecutionEvent, err := buildTaskExecutionEvent(ctx, nCtx, taskPhase, taskPhaseVersion, externalResources)
		if err != nil {
			return handler.UnknownTransition, err
		}

		if err := nCtx.EventsRecorder().RecordTaskEvent(ctx, taskExecutionEvent, a.eventConfig); err != nil {
			logger.Errorf(ctx, "ArrayNode event recording failed: [%s]", err.Error())
			return handler.UnknownTransition, err
		}
	}

	// update array node status
	if err := nCtx.NodeStateWriter().PutArrayNodeState(arrayNodeState); err != nil {
		logger.Errorf(ctx, "failed to store ArrayNode state with err [%s]", err.Error())
		return handler.UnknownTransition, err
	}

	return handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRunning(&handler.ExecutionInfo{})), nil
}

// Setup handles any initialization requirements for this handler
func (a *arrayNodeHandler) Setup(_ context.Context, _ interfaces.SetupContext) error {
	return nil
}

// New initializes a new arrayNodeHandler
func New(nodeExecutor interfaces.Node, eventConfig *config.EventConfig, scope promutils.Scope) (interfaces.NodeHandler, error) {
	// create k8s PluginState byte mocks to reuse instead of creating for each subNode evaluation
	pluginStateBytesNotStarted, err := bytesFromK8sPluginState(k8s.PluginState{Phase: k8s.PluginPhaseNotStarted})
	if err != nil {
		return nil, err
	}

	pluginStateBytesStarted, err := bytesFromK8sPluginState(k8s.PluginState{Phase: k8s.PluginPhaseStarted})
	if err != nil {
		return nil, err
	}

	arrayScope := scope.NewSubScope("array")
	return &arrayNodeHandler{
		eventConfig:                eventConfig,
		metrics:                    newMetrics(arrayScope),
		nodeExecutor:               nodeExecutor,
		pluginStateBytesNotStarted: pluginStateBytesNotStarted,
		pluginStateBytesStarted:    pluginStateBytesStarted,
	}, nil
}

// buildArrayNodeContext creates a custom environment to execute the ArrayNode subnode. This is uniquely required for
// the arrayNodeHandler because we require the same node execution entrypoint (ie. recursiveNodeExecutor.RecursiveNodeHandler)
// but need many different execution details, for example setting input values as a singular item rather than a collection,
// injecting environment variables for flytekit maptask execution, aggregating eventing so that rather than tracking state for
// each subnode individually it sends a single event for the whole ArrayNode, and many more.
func (a *arrayNodeHandler) buildArrayNodeContext(ctx context.Context, nCtx interfaces.NodeExecutionContext, arrayNodeState *handler.ArrayNodeState, arrayNode v1alpha1.ExecutableArrayNode, subNodeIndex int, currentParallelism *uint32) (
	interfaces.Node, executors.ExecutionContext, executors.DAGStructure, executors.NodeLookup, *v1alpha1.NodeSpec, *v1alpha1.NodeStatus, *arrayEventRecorder, error) {

	nodePhase := v1alpha1.NodePhase(arrayNodeState.SubNodePhases.GetItem(subNodeIndex))
	taskPhase := int(arrayNodeState.SubNodeTaskPhases.GetItem(subNodeIndex))

	// need to initialize the inputReader everytime to ensure TaskHandler can access for cache lookups / population
	inputLiteralMap, err := constructLiteralMap(ctx, nCtx.InputReader(), subNodeIndex)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}

	inputReader := newStaticInputReader(nCtx.InputReader(), inputLiteralMap)

	// wrap node lookup
	subNodeSpec := *arrayNode.GetSubNodeSpec()

	subNodeID := fmt.Sprintf("%s-n%d", nCtx.NodeID(), subNodeIndex)
	subNodeSpec.ID = subNodeID
	subNodeSpec.Name = subNodeID
	// mock the input bindings for the subNode to nil to bypass input resolution in the
	// `nodeExecutor.preExecute` function. this is required because this function is the entrypoint
	// for initial cache lookups. an alternative solution would be to mock the datastore to bypass
	// writing the inputFile.
	subNodeSpec.InputBindings = nil

	// TODO - if we want to support more plugin types we need to figure out the best way to store plugin state
	// currently just mocking based on node phase -> which works for all k8s plugins
	// we can not pre-allocated a bit array because max size is 256B and with 5k fanout node state = 1.28MB
	pluginStateBytes := a.pluginStateBytesStarted
	if taskPhase == int(core.PhaseUndefined) || taskPhase == int(core.PhaseRetryableFailure) {
		pluginStateBytes = a.pluginStateBytesNotStarted
	}

	// construct output references
	currentAttempt := uint32(arrayNodeState.SubNodeRetryAttempts.GetItem(subNodeIndex))
	subDataDir, subOutputDir, err := constructOutputReferences(ctx, nCtx, strconv.Itoa(subNodeIndex), strconv.Itoa(int(currentAttempt)))
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}

	subNodeStatus := &v1alpha1.NodeStatus{
		Phase:          nodePhase,
		DataDir:        subDataDir,
		OutputDir:      subOutputDir,
		Attempts:       currentAttempt,
		SystemFailures: uint32(arrayNodeState.SubNodeSystemFailures.GetItem(subNodeIndex)),
		TaskNodeStatus: &v1alpha1.TaskNodeStatus{
			Phase:       taskPhase,
			PluginState: pluginStateBytes,
		},
	}

	// initialize mocks
	arrayNodeLookup := newArrayNodeLookup(nCtx.ContextualNodeLookup(), subNodeID, &subNodeSpec, subNodeStatus)

	arrayExecutionContext := newArrayExecutionContext(nCtx.ExecutionContext(), subNodeIndex, currentParallelism, arrayNode.GetParallelism())

	arrayEventRecorder := newArrayEventRecorder()
	arrayNodeExecutionContextBuilder := newArrayNodeExecutionContextBuilder(a.nodeExecutor.GetNodeExecutionContextBuilder(),
		subNodeID, subNodeIndex, subNodeStatus, inputReader, arrayEventRecorder, currentParallelism, arrayNode.GetParallelism())
	arrayNodeExecutor := a.nodeExecutor.WithNodeExecutionContextBuilder(arrayNodeExecutionContextBuilder)

	return arrayNodeExecutor, arrayExecutionContext, &arrayNodeLookup, &arrayNodeLookup, &subNodeSpec, subNodeStatus, arrayEventRecorder, nil
}
