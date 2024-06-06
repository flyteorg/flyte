package v1alpha1

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/bitarray"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

// The intention of these interfaces is to decouple the algorithm and usage from the actual CRD definition.
// this would help in ease of changes underneath without affecting the code.

//go:generate mockery -all

var nilJSON, _ = json.Marshal(nil)

type CustomState map[string]interface{}
type WorkflowID = string
type TaskID = string
type NodeID = string
type LaunchPlanRefID = Identifier
type ExecutionID = WorkflowExecutionIdentifier

// NodeKind refers to the type of Node.
type NodeKind string

func (n NodeKind) String() string {
	return string(n)
}

type DataReference = storage.DataReference

const (
	// TODO Should we default a NodeKindTask to empty? thus we can assume all unspecified nodetypes as task
	NodeKindTask     NodeKind = "task"
	NodeKindBranch   NodeKind = "branch"   // A Branch node with conditions
	NodeKindWorkflow NodeKind = "workflow" // Either an inline workflow or a remote workflow definition
	NodeKindGate     NodeKind = "gate"     // A Gate node with a condition
	NodeKindArray    NodeKind = "array"    // An array node with a subtask Node
	NodeKindStart    NodeKind = "start"    // Start node is a special node
	NodeKindEnd      NodeKind = "end"
)

// NodePhase indicates the current state of the Node (phase). A node progresses through these states
type NodePhase int

const (
	NodePhaseNotYetStarted NodePhase = iota
	NodePhaseQueued
	NodePhaseRunning
	NodePhaseFailing
	NodePhaseSucceeding
	NodePhaseSucceeded
	NodePhaseFailed
	NodePhaseSkipped
	NodePhaseRetryableFailure
	NodePhaseTimingOut
	NodePhaseTimedOut
	NodePhaseDynamicRunning
	NodePhaseRecovered
)

func (p NodePhase) String() string {
	switch p {
	case NodePhaseNotYetStarted:
		return "NotYetStarted"
	case NodePhaseQueued:
		return "Queued"
	case NodePhaseRunning:
		return "Running"
	case NodePhaseTimingOut:
		return "NodePhaseTimingOut"
	case NodePhaseTimedOut:
		return "NodePhaseTimedOut"
	case NodePhaseSucceeding:
		return "Succeeding"
	case NodePhaseSucceeded:
		return "Succeeded"
	case NodePhaseFailed:
		return "Failed"
	case NodePhaseFailing:
		return "Failing"
	case NodePhaseSkipped:
		return "Skipped"
	case NodePhaseRetryableFailure:
		return "RetryableFailure"
	case NodePhaseDynamicRunning:
		return "DynamicRunning"
	case NodePhaseRecovered:
		return "NodePhaseRecovered"
	}

	return "Unknown"
}

// WorkflowPhase indicates current state of the Workflow.
type WorkflowPhase int

const (
	WorkflowPhaseReady WorkflowPhase = iota
	WorkflowPhaseRunning
	WorkflowPhaseSucceeding
	WorkflowPhaseSuccess
	WorkflowPhaseFailing
	WorkflowPhaseFailed
	WorkflowPhaseAborted
	// WorkflowPhaseHandlingFailureNode is the phase the workflow will enter when a failure is detected in the workflow,
	// the workflow has finished cleaning up (aborted running nodes... etc.) and a failure node is declared in the
	// workflow spec. We enter this explicit phase so as to ensure we do not attempt to repeatedly clean up old nodes
	// when handling a workflow event which might yield to seemingly random failures. This phase ensure we are handling,
	// and only so, the failure node until it's done executing or it fails itself.
	// If a failure node fails to execute (a real possibility), the final failure output of the workflow will only include
	// its failure reason. In other words, its failure will mask the original failure for the workflow. It's imperative
	// failure nodes should be very simple, very resilient and very well tested.
	WorkflowPhaseHandlingFailureNode
)

func (p WorkflowPhase) String() string {
	switch p {
	case WorkflowPhaseReady:
		return "Ready"
	case WorkflowPhaseRunning:
		return "Running"
	case WorkflowPhaseSuccess:
		return "Succeeded"
	case WorkflowPhaseFailed:
		return "Failed"
	case WorkflowPhaseFailing:
		return "Failing"
	case WorkflowPhaseSucceeding:
		return "Succeeding"
	case WorkflowPhaseAborted:
		return "Aborted"
	case WorkflowPhaseHandlingFailureNode:
		return "HandlingFailureNode"
	}
	return "Unknown"
}

// A branchNode has its own Phases. These are used by the child nodes to ensure that the branch node is in the right state
type BranchNodePhase int

const (
	BranchNodeNotYetEvaluated BranchNodePhase = iota
	BranchNodeSuccess
	BranchNodeError
)

func (b BranchNodePhase) String() string {
	switch b {
	case BranchNodeNotYetEvaluated:
		return "NotYetEvaluated"
	case BranchNodeSuccess:
		return "BranchEvalSuccess"
	case BranchNodeError:
		return "BranchEvalFailed"
	}
	return "Undefined"
}

// Failure Handling Policy
type WorkflowOnFailurePolicy core.WorkflowMetadata_OnFailurePolicy

func (in WorkflowOnFailurePolicy) MarshalJSON() ([]byte, error) {
	return json.Marshal(proto.EnumName(core.WorkflowMetadata_OnFailurePolicy_name, int32(in)))
}

func (in *WorkflowOnFailurePolicy) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("WorkflowOnFailurePolicy should be a string, got %s", data)
	}

	var err error
	*in, err = WorkflowOnFailurePolicyString(s)
	return err
}

func WorkflowOnFailurePolicyString(policy string) (WorkflowOnFailurePolicy, error) {
	if val, found := core.WorkflowMetadata_OnFailurePolicy_value[policy]; found {
		return WorkflowOnFailurePolicy(val), nil
	}

	return WorkflowOnFailurePolicy(0), fmt.Errorf("%s does not belong to WorkflowOnFailurePolicy values", policy)
}

// TaskType is a dynamic enumeration, that is defined by configuration
type TaskType = string

// Interface for a Task that can be executed
type ExecutableTask interface {
	TaskType() TaskType
	CoreTask() *core.TaskTemplate
}

// Interface for the executable If block
type ExecutableIfBlock interface {
	GetCondition() *core.BooleanExpression
	GetThenNode() *NodeID
}

// Interface for branch node status. This is the mutable API for a branch node
type ExecutableBranchNodeStatus interface {
	GetPhase() BranchNodePhase
	GetFinalizedNode() *NodeID
}

type MutableBranchNodeStatus interface {
	Mutable
	ExecutableBranchNodeStatus

	SetBranchNodeError()
	SetBranchNodeSuccess(id NodeID)
}

// Interface for dynamic node status.
type ExecutableDynamicNodeStatus interface {
	GetDynamicNodePhase() DynamicNodePhase
	GetDynamicNodeReason() string
	GetExecutionError() *core.ExecutionError
	GetIsFailurePermanent() bool
}

type MutableDynamicNodeStatus interface {
	Mutable
	ExecutableDynamicNodeStatus

	SetDynamicNodePhase(phase DynamicNodePhase)
	SetDynamicNodeReason(reason string)
	SetExecutionError(executionError *core.ExecutionError)
	SetIsFailurePermanent(isFailurePermanent bool)
}

// ExecutableBranchNode is an interface for Branch node. All the methods are purely read only except for the
// GetExecutionStatus. p returns ExecutableBranchNodeStatus, which permits some mutations
type ExecutableBranchNode interface {
	GetIf() ExecutableIfBlock
	GetElse() *NodeID
	GetElseIf() []ExecutableIfBlock
	GetElseFail() *core.Error
}

type ExecutableGateNode interface {
	GetKind() ConditionKind
	GetApprove() *core.ApproveCondition
	GetSignal() *core.SignalCondition
	GetSleep() *core.SleepCondition
}

type ExecutableArrayNode interface {
	GetSubNodeSpec() *NodeSpec
	GetParallelism() *uint32
	GetMinSuccesses() *uint32
	GetMinSuccessRatio() *float32
}

type ExecutableWorkflowNodeStatus interface {
	GetWorkflowNodePhase() WorkflowNodePhase
	GetExecutionError() *core.ExecutionError
}

type MutableWorkflowNodeStatus interface {
	Mutable
	ExecutableWorkflowNodeStatus
	SetWorkflowNodePhase(phase WorkflowNodePhase)
	SetExecutionError(executionError *core.ExecutionError)
}

type ExecutableGateNodeStatus interface {
	GetGateNodePhase() GateNodePhase
}

type MutableGateNodeStatus interface {
	Mutable
	ExecutableGateNodeStatus
	SetGateNodePhase(phase GateNodePhase)
}

type ExecutableArrayNodeStatus interface {
	GetArrayNodePhase() ArrayNodePhase
	GetExecutionError() *core.ExecutionError
	GetSubNodePhases() bitarray.CompactArray
	GetSubNodeTaskPhases() bitarray.CompactArray
	GetSubNodeRetryAttempts() bitarray.CompactArray
	GetSubNodeSystemFailures() bitarray.CompactArray
	GetTaskPhaseVersion() uint32
}

type MutableArrayNodeStatus interface {
	Mutable
	ExecutableArrayNodeStatus
	SetArrayNodePhase(phase ArrayNodePhase)
	SetExecutionError(executionError *core.ExecutionError)
	SetSubNodePhases(subNodePhases bitarray.CompactArray)
	SetSubNodeTaskPhases(subNodeTaskPhases bitarray.CompactArray)
	SetSubNodeRetryAttempts(subNodeRetryAttempts bitarray.CompactArray)
	SetSubNodeSystemFailures(subNodeSystemFailures bitarray.CompactArray)
	SetTaskPhaseVersion(taskPhaseVersion uint32)
}

type Mutable interface {
	IsDirty() bool
}

type MutableNodeStatus interface {
	Mutable
	// Mutation API's
	SetDataDir(DataReference)
	SetOutputDir(d DataReference)
	SetParentNodeID(n *NodeID)
	SetParentTaskID(t *core.TaskExecutionIdentifier)
	UpdatePhase(phase NodePhase, occurredAt metav1.Time, reason string, enableCRDebugMetadata bool, err *core.ExecutionError)
	IncrementAttempts() uint32
	IncrementSystemFailures() uint32
	SetCached()
	ResetDirty()

	GetBranchStatus() MutableBranchNodeStatus
	GetOrCreateBranchStatus() MutableBranchNodeStatus
	GetWorkflowStatus() MutableWorkflowNodeStatus
	GetOrCreateWorkflowStatus() MutableWorkflowNodeStatus

	ClearWorkflowStatus()
	GetOrCreateTaskStatus() MutableTaskNodeStatus
	GetTaskStatus() MutableTaskNodeStatus
	ClearTaskStatus()
	GetOrCreateDynamicNodeStatus() MutableDynamicNodeStatus
	GetDynamicNodeStatus() MutableDynamicNodeStatus
	ClearDynamicNodeStatus()
	ClearLastAttemptStartedAt()
	ClearSubNodeStatus()

	GetGateNodeStatus() MutableGateNodeStatus
	GetOrCreateGateNodeStatus() MutableGateNodeStatus
	ClearGateNodeStatus()

	GetArrayNodeStatus() MutableArrayNodeStatus
	GetOrCreateArrayNodeStatus() MutableArrayNodeStatus
	ClearArrayNodeStatus()
}

type ExecutionTimeInfo interface {
	GetStoppedAt() *metav1.Time
	GetStartedAt() *metav1.Time
	GetLastUpdatedAt() *metav1.Time
}

// ExecutableNodeStatus interface for a Node p. This provides a mutable API.
type ExecutableNodeStatus interface {
	NodeStatusGetter
	MutableNodeStatus
	NodeStatusVisitor
	ExecutionTimeInfo
	GetPhase() NodePhase
	GetQueuedAt() *metav1.Time
	GetLastAttemptStartedAt() *metav1.Time
	GetParentNodeID() *NodeID
	GetParentTaskID() *core.TaskExecutionIdentifier
	GetDataDir() DataReference
	GetOutputDir() DataReference
	GetMessage() string
	GetExecutionError() *core.ExecutionError
	GetAttempts() uint32
	GetSystemFailures() uint32
	GetWorkflowNodeStatus() ExecutableWorkflowNodeStatus
	GetTaskNodeStatus() ExecutableTaskNodeStatus

	IsCached() bool
}

type ExecutableSubWorkflowNodeStatus interface {
	GetPhase() WorkflowPhase
}

type MutableSubWorkflowNodeStatus interface {
	Mutable
	ExecutableSubWorkflowNodeStatus
	SetPhase(phase WorkflowPhase)
}

type ExecutableTaskNodeStatus interface {
	GetPhase() int
	GetPhaseVersion() uint32
	GetPluginState() []byte
	GetPluginStateVersion() uint32
	GetBarrierClockTick() uint32
	GetLastPhaseUpdatedAt() time.Time
	GetPreviousNodeExecutionCheckpointPath() DataReference
	GetCleanupOnFailure() bool
}

type MutableTaskNodeStatus interface {
	Mutable
	ExecutableTaskNodeStatus
	SetPhase(phase int)
	SetLastPhaseUpdatedAt(updatedAt time.Time)
	SetPhaseVersion(version uint32)
	SetPluginState([]byte)
	SetPluginStateVersion(uint32)
	SetBarrierClockTick(tick uint32)
	SetPreviousNodeExecutionCheckpointPath(DataReference)
	SetCleanupOnFailure(bool)
}

// ExecutableWorkflowNode is an interface for a Child Workflow Node
type ExecutableWorkflowNode interface {
	GetLaunchPlanRefID() *LaunchPlanRefID
	GetSubWorkflowRef() *WorkflowID
}

type BaseNode interface {
	GetID() NodeID
	GetKind() NodeKind
}

// ExecutableNode is an interface for the Executable Node
type ExecutableNode interface {
	BaseNode
	IsStartNode() bool
	IsEndNode() bool
	GetTaskID() *TaskID
	GetBranchNode() ExecutableBranchNode
	GetWorkflowNode() ExecutableWorkflowNode
	GetGateNode() ExecutableGateNode
	GetArrayNode() ExecutableArrayNode
	GetOutputAlias() []Alias
	GetInputBindings() []*Binding
	GetResources() *v1.ResourceRequirements
	GetExtendedResources() *core.ExtendedResources
	GetConfig() *v1.ConfigMap
	GetRetryStrategy() *RetryStrategy
	GetExecutionDeadline() *time.Duration
	GetActiveDeadline() *time.Duration
	IsInterruptible() *bool
	GetName() string
	GetContainerImage() string
}

// ExecutableWorkflowStatus is an interface for the Workflow p. This is the mutable portion for a Workflow
type ExecutableWorkflowStatus interface {
	NodeStatusGetter
	ExecutionTimeInfo
	UpdatePhase(p WorkflowPhase, msg string, err *core.ExecutionError)
	GetPhase() WorkflowPhase
	IsTerminated() bool
	GetMessage() string
	GetExecutionError() *core.ExecutionError
	SetDataDir(DataReference)
	GetDataDir() DataReference
	GetOutputReference() DataReference
	SetOutputReference(reference DataReference)
	IncFailedAttempts()
	SetMessage(msg string)
	ConstructNodeDataDir(ctx context.Context, name NodeID) (storage.DataReference, error)
}

type NodeGetter interface {
	GetNode(nodeID NodeID) (ExecutableNode, bool)
}

type BaseWorkflow interface {
	NodeGetter
	StartNode() ExecutableNode
	GetID() WorkflowID
	// FromNode returns all nodes that can be reached directly
	// from the node with the given unique name.
	FromNode(name NodeID) ([]NodeID, error)
	ToNode(name NodeID) ([]NodeID, error)
}

type BaseWorkflowWithStatus interface {
	BaseWorkflow
	NodeStatusGetter
}

// ExecutableSubWorkflow interface captures the methods available on any workflow (top level or child). The Meta section is available
// only for the top level workflow
type ExecutableSubWorkflow interface {
	BaseWorkflow
	GetOutputBindings() []*Binding
	GetOnFailureNode() ExecutableNode
	GetNodes() []NodeID
	GetConnections() *Connections
	GetOutputs() *OutputVarMap
	GetOnFailurePolicy() WorkflowOnFailurePolicy
}

// Meta provides an interface to retrieve labels, annotations and other concepts that are declared only once
// for the top level workflow
type Meta interface {
	GetExecutionID() ExecutionID
	GetK8sWorkflowID() types.NamespacedName
	GetOwnerReference() metav1.OwnerReference
	GetNamespace() string
	GetCreationTimestamp() metav1.Time
	GetAnnotations() map[string]string
	GetLabels() map[string]string
	GetName() string
	GetServiceAccountName() string
	GetSecurityContext() core.SecurityContext
	IsInterruptible() bool
	GetEventVersion() EventVersion
	GetDefinitionVersion() WorkflowDefinitionVersion
	GetRawOutputDataConfig() RawOutputDataConfig
	GetConsoleURL() string
}

type TaskDetailsGetter interface {
	GetTask(id TaskID) (ExecutableTask, error)
}

type SubWorkflowGetter interface {
	FindSubWorkflow(subID WorkflowID) ExecutableSubWorkflow
}

type MetaExtended interface {
	Meta
	TaskDetailsGetter
	SubWorkflowGetter
	GetExecutionStatus() ExecutableWorkflowStatus
}

// A Top level Workflow is a combination of Meta and an ExecutableSubWorkflow
type ExecutableWorkflow interface {
	ExecutableSubWorkflow
	MetaExtended
	NodeStatusGetter
	GetExecutionConfig() ExecutionConfig
	GetConsoleURL() string
}

type NodeStatusGetter interface {
	GetNodeExecutionStatus(ctx context.Context, id NodeID) ExecutableNodeStatus
}

type NodeStatusMap = map[NodeID]ExecutableNodeStatus

type NodeStatusVisitFn = func(node NodeID, status ExecutableNodeStatus)

type NodeStatusVisitor interface {
	VisitNodeStatuses(visitor NodeStatusVisitFn)
}

// Simple callback that can be used to indicate that the workflow with WorkflowID should be re-enqueued for examination.
type EnqueueWorkflow func(workflowID WorkflowID)

func GetOutputsFile(outputDir DataReference) DataReference {
	return outputDir + "/outputs.pb"
}

func GetInputsFile(inputDir DataReference) DataReference {
	return inputDir + "/inputs.pb"
}

func GetDeckFile(inputDir DataReference) DataReference {
	return inputDir + "/deck.html"
}
