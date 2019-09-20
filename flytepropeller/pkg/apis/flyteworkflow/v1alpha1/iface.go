package v1alpha1

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	types2 "github.com/lyft/flyteplugins/go/tasks/v1/types"
	"github.com/lyft/flytestdlib/storage"
)

// The intention of these interfaces is to decouple the algorithm and usage from the actual CRD definition.
// this would help in ease of changes underneath without affecting the code.

//go:generate mockery -all

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
)

func (p NodePhase) String() string {
	switch p {
	case NodePhaseNotYetStarted:
		return "NotYetStarted"
	case NodePhaseQueued:
		return "Queued"
	case NodePhaseRunning:
		return "Running"
	case NodePhaseSucceeding:
		return "Succeeding"
	case NodePhaseSucceeded:
		return "Succeeded"
	case NodePhaseFailed:
		return "Failed"
	case NodePhaseSkipped:
		return "Skipped"
	case NodePhaseRetryableFailure:
		return "RetryableFailure"
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
	ExecutableBranchNodeStatus

	SetBranchNodeError()
	SetBranchNodeSuccess(id NodeID)
}

// Interface for dynamic node status.
type ExecutableDynamicNodeStatus interface {
	GetDynamicNodePhase() DynamicNodePhase
}

type MutableDynamicNodeStatus interface {
	ExecutableDynamicNodeStatus

	SetDynamicNodePhase(phase DynamicNodePhase)
}

// Interface for Branch node. All the methods are purely read only except for the GetExecutionStatus.
// Phase returns ExecutableBranchNodeStatus, which permits some mutations
type ExecutableBranchNode interface {
	GetIf() ExecutableIfBlock
	GetElse() *NodeID
	GetElseIf() []ExecutableIfBlock
	GetElseFail() *core.Error
}

type ExecutableWorkflowNodeStatus interface {
	// Name of the child execution. We only store name since the project and domain will be
	// the same as the parent workflow execution.
	GetWorkflowExecutionName() string
}

type MutableWorkflowNodeStatus interface {
	ExecutableWorkflowNodeStatus

	// Sets the name of the child execution. We only store name since the project and domain
	// will be the same as the parent workflow execution.
	SetWorkflowExecutionName(name string)
}

type MutableNodeStatus interface {
	// Mutation API's
	SetDataDir(DataReference)
	SetParentNodeID(n *NodeID)
	SetParentTaskID(t *core.TaskExecutionIdentifier)
	UpdatePhase(phase NodePhase, occurredAt metav1.Time, reason string)
	IncrementAttempts() uint32
	SetCached()
	ResetDirty()

	GetOrCreateBranchStatus() MutableBranchNodeStatus
	GetOrCreateWorkflowStatus() MutableWorkflowNodeStatus
	ClearWorkflowStatus()
	GetOrCreateTaskStatus() MutableTaskNodeStatus
	ClearTaskStatus()
	GetOrCreateSubWorkflowStatus() MutableSubWorkflowNodeStatus
	ClearSubWorkflowStatus()
	GetOrCreateDynamicNodeStatus() MutableDynamicNodeStatus
	ClearDynamicNodeStatus()
}

// Interface for a Node Phase. This provides a mutable API.
type ExecutableNodeStatus interface {
	NodeStatusGetter
	MutableNodeStatus
	NodeStatusVisitor
	GetPhase() NodePhase
	GetQueuedAt() *metav1.Time
	GetStoppedAt() *metav1.Time
	GetStartedAt() *metav1.Time
	GetLastUpdatedAt() *metav1.Time
	GetParentNodeID() *NodeID
	GetParentTaskID() *core.TaskExecutionIdentifier
	GetDataDir() DataReference
	GetMessage() string
	GetAttempts() uint32
	GetWorkflowNodeStatus() ExecutableWorkflowNodeStatus
	GetTaskNodeStatus() ExecutableTaskNodeStatus
	GetSubWorkflowNodeStatus() ExecutableSubWorkflowNodeStatus

	IsCached() bool
	IsDirty() bool
}

type ExecutableSubWorkflowNodeStatus interface {
	GetPhase() WorkflowPhase
}

type MutableSubWorkflowNodeStatus interface {
	ExecutableSubWorkflowNodeStatus
	SetPhase(phase WorkflowPhase)
}

type ExecutableTaskNodeStatus interface {
	GetPhase() types2.TaskPhase
	GetPhaseVersion() uint32
	GetCustomState() types2.CustomState
}

type MutableTaskNodeStatus interface {
	ExecutableTaskNodeStatus
	SetPhase(phase types2.TaskPhase)
	SetPhaseVersion(version uint32)
	SetCustomState(state types2.CustomState)
}

// Interface for a Child Workflow Node
type ExecutableWorkflowNode interface {
	GetLaunchPlanRefID() *LaunchPlanRefID
	GetSubWorkflowRef() *WorkflowID
}

type BaseNode interface {
	GetID() NodeID
	GetKind() NodeKind
}

// Interface for the Executable Node
type ExecutableNode interface {
	BaseNode
	IsStartNode() bool
	IsEndNode() bool
	GetTaskID() *TaskID
	GetBranchNode() ExecutableBranchNode
	GetWorkflowNode() ExecutableWorkflowNode
	GetOutputAlias() []Alias
	GetInputBindings() []*Binding
	GetResources() *v1.ResourceRequirements
	GetConfig() *v1.ConfigMap
	GetRetryStrategy() *RetryStrategy
}

// Interface for the Workflow Phase. This is the mutable portion for a Workflow
type ExecutableWorkflowStatus interface {
	NodeStatusGetter
	UpdatePhase(p WorkflowPhase, msg string)
	GetPhase() WorkflowPhase
	GetStoppedAt() *metav1.Time
	GetStartedAt() *metav1.Time
	GetLastUpdatedAt() *metav1.Time
	IsTerminated() bool
	GetMessage() string
	SetDataDir(DataReference)
	GetDataDir() DataReference
	GetOutputReference() DataReference
	SetOutputReference(reference DataReference)
	IncFailedAttempts()
	SetMessage(msg string)
	ConstructNodeDataDir(ctx context.Context, constructor storage.ReferenceConstructor, name NodeID) (storage.DataReference, error)
}

type BaseWorkflow interface {
	StartNode() ExecutableNode
	GetID() WorkflowID
	// From returns all nodes that can be reached directly
	// from the node with the given unique name.
	FromNode(name NodeID) ([]NodeID, error)
	GetNode(nodeID NodeID) (ExecutableNode, bool)
}

type BaseWorkflowWithStatus interface {
	BaseWorkflow
	NodeStatusGetter
}

// This interface captures the methods available on any workflow (top level or child). The Meta section is available
// only for the top level workflow
type ExecutableSubWorkflow interface {
	BaseWorkflow
	GetOutputBindings() []*Binding
	GetOnFailureNode() ExecutableNode
	GetNodes() []NodeID
	GetConnections() *Connections
	GetOutputs() *OutputVarMap
}

// WorkflowMeta provides an interface to retrieve labels, annotations and other concepts that are declared only once
// for the top level workflow
type WorkflowMeta interface {
	GetExecutionID() ExecutionID
	GetK8sWorkflowID() types.NamespacedName
	NewControllerRef() metav1.OwnerReference
	GetNamespace() string
	GetCreationTimestamp() metav1.Time
	GetAnnotations() map[string]string
	GetLabels() map[string]string
	GetName() string
	GetServiceAccountName() string
}

type WorkflowMetaExtended interface {
	WorkflowMeta
	GetTask(id TaskID) (ExecutableTask, error)
	FindSubWorkflow(subID WorkflowID) ExecutableSubWorkflow
	GetExecutionStatus() ExecutableWorkflowStatus
}

// A Top level Workflow is a combination of WorkflowMeta and an ExecutableSubWorkflow
type ExecutableWorkflow interface {
	ExecutableSubWorkflow
	WorkflowMetaExtended
	NodeStatusGetter
}

type NodeStatusGetter interface {
	GetNodeExecutionStatus(id NodeID) ExecutableNodeStatus
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

func GetOutputErrorFile(inputDir DataReference) DataReference {
	return inputDir + "/error.pb"
}

func GetFutureFile() string {
	return "futures.pb"
}

func GetCompiledFutureFile() string {
	return "futures_compiled.pb"
}
