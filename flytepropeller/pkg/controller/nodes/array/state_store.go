package array

import (
	"context"
	"fmt"
	"strconv"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/common"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/handler"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces"
	"github.com/flyteorg/flyte/flytestdlib/bitarray"
)

//go:generate mockery -all -case=underscore

type arrayNodeStateStore interface {
	initArrayNodeState(maxAttemptsValue int, maxSystemFailuresValue int, size int) error
	buildArrayNodeContext(ctx context.Context, nCtx interfaces.NodeExecutionContext, arrayNode v1alpha1.ExecutableArrayNode, subNodeIndex int, eventRecorder ArrayEventRecorder) (
		interfaces.Node, executors.ExecutionContext, executors.DAGStructure, executors.NodeLookup, *v1alpha1.NodeSpec, *v1alpha1.NodeStatus, error)
	persistArraySubNodeState(ctx context.Context, nCtx interfaces.NodeExecutionContext, subNodeStatus *v1alpha1.NodeStatus, index int)
	getAttempts(ctx context.Context, nCtx interfaces.NodeExecutionContext, index int) uint32
	getTaskPhase(ctx context.Context, nCtx interfaces.NodeExecutionContext, index int) int
}

func getSubNodeID(index int) string {
	return fmt.Sprintf("n%d", index)
}

type fullStateStore struct {
	arrayNodeHandler   *arrayNodeHandler
	arrayNodeStateCopy *handler.ArrayNodeState
}

func (f *fullStateStore) initArrayNodeState(maxAttemptsValue int, maxSystemFailuresValue int, size int) error {
	for _, item := range []struct {
		arrayReference *bitarray.CompactArray
		maxValue       int
	}{
		// we use NodePhaseRecovered for the `maxValue` of `SubNodePhases` because `Phase` is
		// defined as an `iota` so it is impossible to programmatically get largest value
		{arrayReference: &f.arrayNodeStateCopy.SubNodePhases, maxValue: int(v1alpha1.NodePhaseRecovered)},
	} {
		var err error
		*item.arrayReference, err = bitarray.NewCompactArray(uint(size), bitarray.Item(item.maxValue))
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *fullStateStore) persistArraySubNodeState(ctx context.Context, nCtx interfaces.NodeExecutionContext, subNodeStatus *v1alpha1.NodeStatus, index int) {
	f.arrayNodeStateCopy.SubNodePhases.SetItem(index, uint64(subNodeStatus.GetPhase()))

	subNodeStatusAddress := nCtx.NodeStatus().GetNodeExecutionStatus(ctx, getSubNodeID(index)).(*v1alpha1.NodeStatus)
	*subNodeStatusAddress = *subNodeStatus
}

func (f *fullStateStore) buildArrayNodeContext(ctx context.Context, nCtx interfaces.NodeExecutionContext, arrayNode v1alpha1.ExecutableArrayNode, subNodeIndex int, eventRecorder ArrayEventRecorder) (
	interfaces.Node, executors.ExecutionContext, executors.DAGStructure, executors.NodeLookup, *v1alpha1.NodeSpec, *v1alpha1.NodeStatus, error) {
	subNodeStatus := nCtx.NodeStatus().GetNodeExecutionStatus(ctx, getSubNodeID(subNodeIndex))
	if subNodeStatus == nil {
		subNodeStatus = &v1alpha1.NodeStatus{}
	}

	currentAttempt := subNodeStatus.GetAttempts()
	subDataDir, subOutputDir, err := constructOutputReferences(ctx, nCtx, strconv.Itoa(subNodeIndex), strconv.Itoa(int(currentAttempt)))
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	subNodeStatus.SetDataDir(subDataDir)
	subNodeStatus.SetOutputDir(subOutputDir)

	subNodeSpec := *arrayNode.GetSubNodeSpec()
	subNodeID := getSubNodeID(subNodeIndex)
	subNodeSpec.ID = subNodeID
	subNodeSpec.Name = subNodeID
	// mock the input bindings for the subNode to nil to bypass input resolution in the
	// `nodeExecutor.preExecute` function. this is required because this function is the entrypoint
	// for initial cache lookups. an alternative solution would be to mock the datastore to bypass
	// writing the inputFile.
	subNodeSpec.InputBindings = nil

	arrayNodeLookup := newArrayNodeLookup(nCtx.ContextualNodeLookup(), subNodeID, &subNodeSpec, subNodeStatus.(*v1alpha1.NodeStatus))

	newParentInfo, err := common.CreateParentInfo(nCtx.ExecutionContext().GetParentInfo(), nCtx.NodeID(), nCtx.CurrentAttempt(), false, true)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	// set new parent info and re-initialize the control flow to not inherit the parent's state
	arrayExecutionContext := newArrayExecutionContext(
		executors.NewExecutionContext(nCtx.ExecutionContext(), nCtx.ExecutionContext(), nCtx.ExecutionContext(), newParentInfo, executors.InitializeControlFlow()),
		subNodeIndex)

	// need to initialize the inputReader every time to ensure TaskHandler can access for cache lookups / population
	inputs, err := nCtx.InputReader().Get(ctx)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	inputLiteralMap, err := constructLiteralMap(inputs, subNodeIndex)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	inputReader := newStaticInputReader(nCtx.InputReader(), inputLiteralMap)

	arrayNodeExecutionContextBuilder := newArrayNodeExecutionContextBuilder(f.arrayNodeHandler.nodeExecutor.GetNodeExecutionContextBuilder(),
		subNodeID, subNodeIndex, subNodeStatus.(*v1alpha1.NodeStatus), inputReader, eventRecorder)
	arrayNodeExecutor := f.arrayNodeHandler.nodeExecutor.WithNodeExecutionContextBuilder(arrayNodeExecutionContextBuilder)

	return arrayNodeExecutor, arrayExecutionContext, &arrayNodeLookup, &arrayNodeLookup, &subNodeSpec, subNodeStatus.(*v1alpha1.NodeStatus), nil
}

func (f *fullStateStore) getAttempts(ctx context.Context, nCtx interfaces.NodeExecutionContext, index int) uint32 {
	subNodeStatus := nCtx.NodeStatus().GetNodeExecutionStatus(ctx, getSubNodeID(index))
	if subNodeStatus == nil {
		return 0
	}
	return subNodeStatus.GetAttempts()
}

func (f *fullStateStore) getTaskPhase(ctx context.Context, nCtx interfaces.NodeExecutionContext, index int) int {
	subNodeStatus := nCtx.NodeStatus().GetNodeExecutionStatus(ctx, getSubNodeID(index))
	if subNodeStatus == nil {
		return int(core.PhaseUndefined)
	}

	if subNodeStatus.GetWorkflowNodeStatus() != nil {
		return int(subNodeStatus.GetWorkflowNodeStatus().GetWorkflowNodePhase())
	} else if subNodeStatus.GetTaskNodeStatus() != nil {
		return subNodeStatus.GetTaskNodeStatus().GetPhase()
	}

	return int(core.PhaseUndefined)
}

type minStateStore struct {
	arrayNodeHandler   *arrayNodeHandler
	arrayNodeStateCopy *handler.ArrayNodeState
}

func (m *minStateStore) initArrayNodeState(maxAttemptsValue int, maxSystemFailuresValue int, size int) error {
	for _, item := range []struct {
		arrayReference *bitarray.CompactArray
		maxValue       int
	}{
		// we use NodePhaseRecovered for the `maxValue` of `SubNodePhases` because `Phase` is
		// defined as an `iota` so it is impossible to programmatically get largest value
		{arrayReference: &m.arrayNodeStateCopy.SubNodePhases, maxValue: int(v1alpha1.NodePhaseRecovered)},
		{arrayReference: &m.arrayNodeStateCopy.SubNodeTaskPhases, maxValue: len(core.Phases) - 1},
		{arrayReference: &m.arrayNodeStateCopy.SubNodeRetryAttempts, maxValue: maxAttemptsValue},
		{arrayReference: &m.arrayNodeStateCopy.SubNodeSystemFailures, maxValue: maxSystemFailuresValue},
		{arrayReference: &m.arrayNodeStateCopy.SubNodeDeltaTimestamps, maxValue: MAX_DELTA_TIMESTAMP},
	} {
		var err error
		*item.arrayReference, err = bitarray.NewCompactArray(uint(size), bitarray.Item(item.maxValue))
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *minStateStore) persistArraySubNodeState(ctx context.Context, nCtx interfaces.NodeExecutionContext, subNodeStatus *v1alpha1.NodeStatus, index int) {
	m.arrayNodeStateCopy.SubNodePhases.SetItem(index, uint64(subNodeStatus.GetPhase()))
	if subNodeStatus.GetTaskNodeStatus() == nil {
		// resetting task phase because during retries we clear the GetTaskNodeStatus
		m.arrayNodeStateCopy.SubNodeTaskPhases.SetItem(index, uint64(0))
	} else {
		m.arrayNodeStateCopy.SubNodeTaskPhases.SetItem(index, uint64(subNodeStatus.GetTaskNodeStatus().GetPhase()))
	}
	m.arrayNodeStateCopy.SubNodeRetryAttempts.SetItem(index, uint64(subNodeStatus.GetAttempts()))
	m.arrayNodeStateCopy.SubNodeSystemFailures.SetItem(index, uint64(subNodeStatus.GetSystemFailures()))

	if m.arrayNodeStateCopy.SubNodeDeltaTimestamps.BitSet != nil {
		startedAt := nCtx.NodeStatus().GetLastAttemptStartedAt()
		subNodeStartedAt := subNodeStatus.GetLastAttemptStartedAt()
		if subNodeStartedAt == nil {
			// subNodeStartedAt == nil indicates either (1) node has not started or (2) node status has
			// been reset (ex. retryable failure). in both cases we set the delta timestamp to 0
			m.arrayNodeStateCopy.SubNodeDeltaTimestamps.SetItem(index, 0)
		} else if startedAt != nil && m.arrayNodeStateCopy.SubNodeDeltaTimestamps.GetItem(index) == 0 {
			// otherwise if `SubNodeDeltaTimestamps` is unset, we compute the delta and set it
			deltaDuration := uint64(subNodeStartedAt.Time.Sub(startedAt.Time).Seconds())
			m.arrayNodeStateCopy.SubNodeDeltaTimestamps.SetItem(index, deltaDuration)
		}
	}
}

// buildArrayNodeContext creates a custom environment to execute the ArrayNode subnode. This is uniquely required for
// the arrayNodeHandler because we require the same node execution entrypoint (ie. recursiveNodeExecutor.RecursiveNodeHandler)
// but need many different execution details, for example setting input values as a singular item rather than a collection,
// injecting environment variables for flytekit maptask execution, aggregating eventing so that rather than tracking state for
// each subnode individually it sends a single event for the whole ArrayNode, and many more.
func (m *minStateStore) buildArrayNodeContext(ctx context.Context, nCtx interfaces.NodeExecutionContext, arrayNode v1alpha1.ExecutableArrayNode, subNodeIndex int, eventRecorder ArrayEventRecorder) (
	interfaces.Node, executors.ExecutionContext, executors.DAGStructure, executors.NodeLookup, *v1alpha1.NodeSpec, *v1alpha1.NodeStatus, error) {
	nodePhase := v1alpha1.NodePhase(m.arrayNodeStateCopy.SubNodePhases.GetItem(subNodeIndex))
	taskPhase := int(m.arrayNodeStateCopy.SubNodeTaskPhases.GetItem(subNodeIndex))

	// need to initialize the inputReader every time to ensure TaskHandler can access for cache lookups / population
	inputs, err := nCtx.InputReader().Get(ctx)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	inputLiteralMap, err := constructLiteralMap(inputs, subNodeIndex)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	inputReader := newStaticInputReader(nCtx.InputReader(), inputLiteralMap)

	// wrap node lookup
	subNodeSpec := *arrayNode.GetSubNodeSpec()

	subNodeID := fmt.Sprintf("n%d", subNodeIndex)
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
	pluginStateBytes := m.arrayNodeHandler.pluginStateBytesStarted
	if taskPhase == int(core.PhaseUndefined) || taskPhase == int(core.PhaseRetryableFailure) || taskPhase == int(core.PhaseWaitingForResources) {
		pluginStateBytes = m.arrayNodeHandler.pluginStateBytesNotStarted
	}

	// construct output references
	currentAttempt := uint32(m.arrayNodeStateCopy.SubNodeRetryAttempts.GetItem(subNodeIndex))
	subDataDir, subOutputDir, err := constructOutputReferences(ctx, nCtx, strconv.Itoa(subNodeIndex), strconv.Itoa(int(currentAttempt)))
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	// compute start time for subNode using delta timestamp from ArrayNode NodeStatus
	var startedAt *metav1.Time
	if nCtx.NodeStatus().GetLastAttemptStartedAt() != nil && m.arrayNodeStateCopy.SubNodeDeltaTimestamps.BitSet != nil {
		if deltaSeconds := m.arrayNodeStateCopy.SubNodeDeltaTimestamps.GetItem(subNodeIndex); deltaSeconds != 0 {
			startedAt = &metav1.Time{Time: nCtx.NodeStatus().GetLastAttemptStartedAt().Add(time.Duration(deltaSeconds) * time.Second)} // #nosec G115
		}
	}

	subNodeStatus := &v1alpha1.NodeStatus{
		Phase:          nodePhase,
		DataDir:        subDataDir,
		OutputDir:      subOutputDir,
		Attempts:       currentAttempt,
		SystemFailures: uint32(m.arrayNodeStateCopy.SubNodeSystemFailures.GetItem(subNodeIndex)),
		TaskNodeStatus: &v1alpha1.TaskNodeStatus{
			Phase:       taskPhase,
			PluginState: pluginStateBytes,
		},
		LastAttemptStartedAt: startedAt,
	}

	// initialize mocks
	arrayNodeLookup := newArrayNodeLookup(nCtx.ContextualNodeLookup(), subNodeID, &subNodeSpec, subNodeStatus)

	newParentInfo, err := common.CreateParentInfo(nCtx.ExecutionContext().GetParentInfo(), nCtx.NodeID(), nCtx.CurrentAttempt(), false, true)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	// set new parent info and re-initialize the sub-node's control flow to not share/update the parent wf state
	arrayExecutionContext := newArrayExecutionContext(
		executors.NewExecutionContext(nCtx.ExecutionContext(), nCtx.ExecutionContext(), nCtx.ExecutionContext(), newParentInfo, executors.InitializeControlFlow()),
		subNodeIndex)

	arrayNodeExecutionContextBuilder := newArrayNodeExecutionContextBuilder(m.arrayNodeHandler.nodeExecutor.GetNodeExecutionContextBuilder(),
		subNodeID, subNodeIndex, subNodeStatus, inputReader, eventRecorder)
	arrayNodeExecutor := m.arrayNodeHandler.nodeExecutor.WithNodeExecutionContextBuilder(arrayNodeExecutionContextBuilder)

	return arrayNodeExecutor, arrayExecutionContext, &arrayNodeLookup, &arrayNodeLookup, &subNodeSpec, subNodeStatus, nil
}

func (m *minStateStore) getAttempts(ctx context.Context, nCtx interfaces.NodeExecutionContext, index int) uint32 {
	return uint32(m.arrayNodeStateCopy.SubNodeRetryAttempts.GetItem(index))
}

func (m *minStateStore) getTaskPhase(ctx context.Context, nCtx interfaces.NodeExecutionContext, index int) int {
	return int(m.arrayNodeStateCopy.SubNodeTaskPhases.GetItem(index))
}
