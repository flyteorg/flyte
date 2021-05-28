package nodes

import (
	"context"
	"fmt"
	"testing"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
	"github.com/stretchr/testify/assert"
)

func TestCanExecute(t *testing.T) {
	nodeN0 := "n0"
	nodeN1 := "n1"
	nodeN2 := "n2"
	ctx := context.Background()
	upstreamN2 := []v1alpha1.NodeID{nodeN0, nodeN1}

	// Table tests are not really helpful here, so we decided against it

	t.Run("startNode", func(t *testing.T) {
		mockNode := &mocks.BaseNode{}
		mockNode.On("GetID").Return(v1alpha1.StartNodeID)
		p, err := CanExecute(ctx, nil, nil, mockNode)
		assert.NoError(t, err)
		assert.Equal(t, PredicatePhaseReady, p)
	})

	t.Run("noUpstreamConnection", func(t *testing.T) {
		// Setup
		mockNodeStatus := &mocks.ExecutableNodeStatus{}
		// No parent node
		mockNodeStatus.OnGetParentNodeID().Return(nil)
		mockNode := &mocks.BaseNode{}
		mockNode.OnGetID().Return(nodeN2)
		mockWf := &mocks.ExecutableWorkflow{}
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN2).Return(mockNodeStatus)
		mockWf.OnGetID().Return("w1")
		mockWf.OnToNode("n2").Return(nil, fmt.Errorf("not found"))

		p, err := CanExecute(ctx, mockWf, mockWf, mockNode)
		assert.Error(t, err)
		assert.Equal(t, PredicatePhaseUndefined, p)
	})

	t.Run("upstreamConnectionsNotReady", func(t *testing.T) {
		// Setup
		mockN2Status := &mocks.ExecutableNodeStatus{}
		// No parent node
		mockN2Status.OnGetParentNodeID().Return(nil)
		mockN2Status.OnIsDirty().Return(false)
		mockNode := &mocks.BaseNode{}
		mockNode.OnGetID().Return(nodeN2)

		mockN0Status := &mocks.ExecutableNodeStatus{}
		mockN0Status.OnGetPhase().Return(v1alpha1.NodePhaseRunning)
		mockN0Status.OnIsDirty().Return(false)

		mockN1Status := &mocks.ExecutableNodeStatus{}
		mockN1Status.OnGetPhase().Return(v1alpha1.NodePhaseRunning)
		mockN1Status.OnIsDirty().Return(false)

		mockWf := &mocks.ExecutableWorkflow{}
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN0).Return(mockN0Status)
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN1).Return(mockN1Status)
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN2).Return(mockN2Status)
		mockWf.OnToNode(nodeN2).Return(upstreamN2, nil)
		mockWf.OnGetID().Return("w1")

		p, err := CanExecute(ctx, mockWf, mockWf, mockNode)
		assert.NoError(t, err)
		assert.Equal(t, PredicatePhaseNotReady, p)
	})

	t.Run("upstreamConnectionsPartialReady", func(t *testing.T) {
		// Setup
		mockN2Status := &mocks.ExecutableNodeStatus{}
		// No parent node
		mockN2Status.On("GetParentNodeID").Return(nil)
		mockN2Status.On("IsDirty").Return(false)

		mockNode := &mocks.BaseNode{}
		mockNode.On("GetID").Return(nodeN2)

		mockN0Status := &mocks.ExecutableNodeStatus{}
		mockN0Status.On("GetPhase").Return(v1alpha1.NodePhaseRunning)
		mockN0Status.On("IsDirty").Return(false)

		mockN1Status := &mocks.ExecutableNodeStatus{}
		mockN1Status.On("GetPhase").Return(v1alpha1.NodePhaseSucceeded)
		mockN1Status.On("IsDirty").Return(false)

		mockWf := &mocks.ExecutableWorkflow{}
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN0).Return(mockN0Status)
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN1).Return(mockN1Status)
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN2).Return(mockN2Status)
		mockWf.OnToNode(nodeN2).Return(upstreamN2, nil)
		mockWf.On("GetID").Return("w1")

		p, err := CanExecute(ctx, mockWf, mockWf, mockNode)
		assert.NoError(t, err)
		assert.Equal(t, PredicatePhaseNotReady, p)
	})

	t.Run("upstreamConnectionsCompletelyReady", func(t *testing.T) {
		// Setup
		mockN2Status := &mocks.ExecutableNodeStatus{}
		// No parent node
		mockN2Status.On("GetParentNodeID").Return(nil)
		mockN2Status.On("IsDirty").Return(false)

		mockNode := &mocks.BaseNode{}
		mockNode.On("GetID").Return(nodeN2)

		mockN0Status := &mocks.ExecutableNodeStatus{}
		mockN0Status.On("GetPhase").Return(v1alpha1.NodePhaseSucceeded)
		mockN0Status.On("IsDirty").Return(false)

		mockN1Status := &mocks.ExecutableNodeStatus{}
		mockN1Status.On("GetPhase").Return(v1alpha1.NodePhaseSucceeded)
		mockN1Status.On("IsDirty").Return(false)

		mockWf := &mocks.ExecutableWorkflow{}
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN0).Return(mockN0Status)
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN1).Return(mockN1Status)
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN2).Return(mockN2Status)
		mockWf.OnToNode(nodeN2).Return(upstreamN2, nil)

		mockWf.On("GetID").Return("w1")

		p, err := CanExecute(ctx, mockWf, mockWf, mockNode)
		assert.NoError(t, err)
		assert.Equal(t, PredicatePhaseReady, p)
	})

	t.Run("upstreamConnectionsDirty", func(t *testing.T) {
		// Setup
		mockN2Status := &mocks.ExecutableNodeStatus{}
		// No parent node
		mockN2Status.On("GetParentNodeID").Return(nil)
		mockN2Status.On("IsDirty").Return(false)

		mockNode := &mocks.BaseNode{}
		mockNode.On("GetID").Return(nodeN2)

		mockN0Status := &mocks.ExecutableNodeStatus{}
		mockN0Status.On("GetPhase").Return(v1alpha1.NodePhaseSucceeded)
		mockN0Status.On("IsDirty").Return(false)

		mockN1Status := &mocks.ExecutableNodeStatus{}
		mockN1Status.On("GetPhase").Return(v1alpha1.NodePhaseSucceeded)
		mockN1Status.On("IsDirty").Return(true)

		mockWf := &mocks.ExecutableWorkflow{}
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN0).Return(mockN0Status)
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN1).Return(mockN1Status)
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN2).Return(mockN2Status)
		mockWf.OnToNode(nodeN2).Return(upstreamN2, nil)
		mockWf.On("GetID").Return("w1")

		p, err := CanExecute(ctx, mockWf, mockWf, mockNode)
		assert.NoError(t, err)
		assert.Equal(t, PredicatePhaseNotReady, p)
	})

	t.Run("upstreamConnectionsPartialSkipped", func(t *testing.T) {
		// Setup
		mockN2Status := &mocks.ExecutableNodeStatus{}
		// No parent node
		mockN2Status.On("GetParentNodeID").Return(nil)
		mockN2Status.On("IsDirty").Return(false)

		mockNode := &mocks.BaseNode{}
		mockNode.On("GetID").Return(nodeN2)

		mockN0Status := &mocks.ExecutableNodeStatus{}
		mockN0Status.On("GetPhase").Return(v1alpha1.NodePhaseRunning)
		mockN0Status.On("IsDirty").Return(false)

		mockN1Status := &mocks.ExecutableNodeStatus{}
		mockN1Status.On("GetPhase").Return(v1alpha1.NodePhaseSkipped)
		mockN1Status.On("IsDirty").Return(false)

		mockWf := &mocks.ExecutableWorkflow{}
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN0).Return(mockN0Status)
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN1).Return(mockN1Status)
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN2).Return(mockN2Status)
		mockWf.OnToNode(nodeN2).Return(upstreamN2, nil)
		mockWf.On("GetID").Return("w1")

		p, err := CanExecute(ctx, mockWf, mockWf, mockNode)
		assert.NoError(t, err)
		assert.Equal(t, PredicatePhaseNotReady, p)
	})

	t.Run("upstreamConnectionsOneSkipped", func(t *testing.T) {
		// Setup
		mockN2Status := &mocks.ExecutableNodeStatus{}
		// No parent node
		mockN2Status.On("GetParentNodeID").Return(nil)
		mockN2Status.On("IsDirty").Return(false)

		mockNode := &mocks.BaseNode{}
		mockNode.On("GetID").Return(nodeN2)

		mockN0Status := &mocks.ExecutableNodeStatus{}
		mockN0Status.On("GetPhase").Return(v1alpha1.NodePhaseSucceeded)
		mockN0Status.On("IsDirty").Return(false)

		mockN1Status := &mocks.ExecutableNodeStatus{}
		mockN1Status.On("GetPhase").Return(v1alpha1.NodePhaseSkipped)
		mockN1Status.On("IsDirty").Return(false)

		mockWf := &mocks.ExecutableWorkflow{}
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN0).Return(mockN0Status)
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN1).Return(mockN1Status)
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN2).Return(mockN2Status)
		mockWf.OnToNode(nodeN2).Return(upstreamN2, nil)
		mockWf.On("GetID").Return("w1")

		p, err := CanExecute(ctx, mockWf, mockWf, mockNode)
		assert.NoError(t, err)
		assert.Equal(t, PredicatePhaseSkip, p)
	})

	t.Run("upstreamConnectionsAllSkipped", func(t *testing.T) {
		// Setup
		mockN2Status := &mocks.ExecutableNodeStatus{}
		// No parent node
		mockN2Status.On("GetParentNodeID").Return(nil)
		mockN2Status.On("IsDirty").Return(false)

		mockNode := &mocks.BaseNode{}
		mockNode.On("GetID").Return(nodeN2)

		mockN0Status := &mocks.ExecutableNodeStatus{}
		mockN0Status.On("GetPhase").Return(v1alpha1.NodePhaseSkipped)
		mockN0Status.On("IsDirty").Return(false)

		mockN1Status := &mocks.ExecutableNodeStatus{}
		mockN1Status.On("GetPhase").Return(v1alpha1.NodePhaseSkipped)
		mockN1Status.On("IsDirty").Return(false)

		mockWf := &mocks.ExecutableWorkflow{}
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN0).Return(mockN0Status)
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN1).Return(mockN1Status)
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN2).Return(mockN2Status)
		mockWf.OnToNode(nodeN2).Return(upstreamN2, nil)
		mockWf.On("GetID").Return("w1")

		p, err := CanExecute(ctx, mockWf, mockWf, mockNode)
		assert.NoError(t, err)
		assert.Equal(t, PredicatePhaseSkip, p)
	})

	// Failed should never happen for predicate check. Hence we return not ready
	t.Run("upstreamConnectionsFailed", func(t *testing.T) {
		// Setup
		mockN2Status := &mocks.ExecutableNodeStatus{}
		// No parent node
		mockN2Status.On("GetParentNodeID").Return(nil)
		mockN2Status.On("IsDirty").Return(false)

		mockNode := &mocks.BaseNode{}
		mockNode.On("GetID").Return(nodeN2)

		mockN0Status := &mocks.ExecutableNodeStatus{}
		mockN0Status.On("GetPhase").Return(v1alpha1.NodePhaseFailed)
		mockN0Status.On("IsDirty").Return(false)

		mockN1Status := &mocks.ExecutableNodeStatus{}
		mockN1Status.On("GetPhase").Return(v1alpha1.NodePhaseFailed)
		mockN1Status.On("IsDirty").Return(false)

		mockWf := &mocks.ExecutableWorkflow{}
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN0).Return(mockN0Status)
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN1).Return(mockN1Status)
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN2).Return(mockN2Status)
		mockWf.OnToNode(nodeN2).Return(upstreamN2, nil)
		mockWf.On("GetID").Return("w1")

		p, err := CanExecute(ctx, mockWf, mockWf, mockNode)
		assert.NoError(t, err)
		assert.Equal(t, PredicatePhaseSkip, p)
	})

	// Branch node tests

	// ParentNode not found?
	t.Run("upstreamConnectionsParentNodeNotFound", func(t *testing.T) {
		// Setup
		mockN2Status := &mocks.ExecutableNodeStatus{}
		// No parent node
		mockN2Status.On("GetParentNodeID").Return(&nodeN0)
		mockN2Status.On("IsDirty").Return(false)

		mockNode := &mocks.BaseNode{}
		mockNode.On("GetID").Return(nodeN2)

		mockN0Status := &mocks.ExecutableNodeStatus{}
		mockN0Status.On("GetPhase").Return(v1alpha1.NodePhaseSucceeded)
		mockN0Status.On("IsDirty").Return(false)

		mockN1Status := &mocks.ExecutableNodeStatus{}
		mockN1Status.On("GetPhase").Return(v1alpha1.NodePhaseSucceeded)
		mockN1Status.On("IsDirty").Return(false)

		mockWf := &mocks.ExecutableWorkflow{}
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN0).Return(mockN0Status)
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN1).Return(mockN1Status)
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN2).Return(mockN2Status)
		mockWf.OnToNode(nodeN2).Return(upstreamN2, nil)
		mockWf.On("GetNode", nodeN0).Return(nil, false)
		mockWf.On("GetID").Return("w1")

		p, err := CanExecute(ctx, mockWf, mockWf, mockNode)
		assert.Error(t, err)
		assert.Equal(t, PredicatePhaseUndefined, p)
	})

	// ParentNode branch ready
	t.Run("upstreamConnectionsBranchSuccessOtherSuccess", func(t *testing.T) {
		// Setup
		mockN2Status := &mocks.ExecutableNodeStatus{}
		// No parent node
		mockN2Status.On("GetParentNodeID").Return(&nodeN0)
		mockN2Status.On("IsDirty").Return(false)

		mockNode := &mocks.BaseNode{}
		mockNode.On("GetID").Return(nodeN2)

		mockN0BranchStatus := &mocks.MutableBranchNodeStatus{}
		mockN0BranchStatus.On("GetPhase").Return(v1alpha1.BranchNodeSuccess)
		mockN0BranchNode := &mocks.ExecutableBranchNode{}

		mockN0Node := &mocks.ExecutableNode{}
		mockN0Node.On("GetBranchNode").Return(mockN0BranchNode)
		mockN0Status := &mocks.ExecutableNodeStatus{}
		mockN0Status.On("GetPhase").Return(v1alpha1.NodePhaseSucceeded)
		mockN0Status.On("GetBranchStatus").Return(mockN0BranchStatus)
		mockN0Status.On("IsDirty").Return(false)

		mockN1Status := &mocks.ExecutableNodeStatus{}
		mockN1Status.On("GetPhase").Return(v1alpha1.NodePhaseSucceeded)
		mockN1Status.On("IsDirty").Return(false)

		mockWf := &mocks.ExecutableWorkflow{}
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN0).Return(mockN0Status)
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN1).Return(mockN1Status)
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN2).Return(mockN2Status)
		mockWf.OnToNode(nodeN2).Return(upstreamN2, nil)
		mockWf.On("GetNode", nodeN0).Return(mockN0Node, true)
		mockWf.On("GetID").Return("w1")

		p, err := CanExecute(ctx, mockWf, mockWf, mockNode)
		assert.NoError(t, err)
		assert.Equal(t, PredicatePhaseReady, p)
	})

	// ParentNode branch ready
	t.Run("upstreamConnectionsBranchSuccessOtherSkipped", func(t *testing.T) {
		// Setup
		mockN2Status := &mocks.ExecutableNodeStatus{}
		// No parent node
		mockN2Status.On("GetParentNodeID").Return(&nodeN0)
		mockN2Status.On("IsDirty").Return(false)

		mockNode := &mocks.BaseNode{}
		mockNode.On("GetID").Return(nodeN2)

		mockN0BranchStatus := &mocks.MutableBranchNodeStatus{}
		mockN0BranchStatus.On("GetPhase").Return(v1alpha1.BranchNodeSuccess)

		mockN0BranchNode := &mocks.ExecutableBranchNode{}
		mockN0Node := &mocks.ExecutableNode{}
		mockN0Node.On("GetBranchNode").Return(mockN0BranchNode)
		mockN0Status := &mocks.ExecutableNodeStatus{}
		mockN0Status.On("GetPhase").Return(v1alpha1.NodePhaseSucceeded)
		mockN0Status.On("GetBranchStatus").Return(mockN0BranchStatus)
		mockN0Status.On("IsDirty").Return(false)

		mockN1Status := &mocks.ExecutableNodeStatus{}
		mockN1Status.On("GetPhase").Return(v1alpha1.NodePhaseSkipped)
		mockN1Status.On("IsDirty").Return(false)

		mockWf := &mocks.ExecutableWorkflow{}
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN0).Return(mockN0Status)
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN1).Return(mockN1Status)
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN2).Return(mockN2Status)
		mockWf.OnToNode(nodeN2).Return(upstreamN2, nil)
		mockWf.On("GetNode", nodeN0).Return(mockN0Node, true)
		mockWf.On("GetID").Return("w1")

		p, err := CanExecute(ctx, mockWf, mockWf, mockNode)
		assert.NoError(t, err)
		assert.Equal(t, PredicatePhaseSkip, p)
	})

	// ParentNode branch ready
	t.Run("upstreamConnectionsBranchSuccessOtherRunning", func(t *testing.T) {
		// Setup
		mockN2Status := &mocks.ExecutableNodeStatus{}
		// No parent node
		mockN2Status.On("GetParentNodeID").Return(&nodeN0)
		mockN2Status.On("IsDirty").Return(false)

		mockNode := &mocks.BaseNode{}
		mockNode.On("GetID").Return(nodeN2)

		mockN0BranchStatus := &mocks.MutableBranchNodeStatus{}
		mockN0BranchStatus.On("GetPhase").Return(v1alpha1.BranchNodeSuccess)

		mockN0BranchNode := &mocks.ExecutableBranchNode{}
		mockN0Node := &mocks.ExecutableNode{}
		mockN0Node.On("GetBranchNode").Return(mockN0BranchNode)
		mockN0Status := &mocks.ExecutableNodeStatus{}
		mockN0Status.On("GetPhase").Return(v1alpha1.NodePhaseSucceeded)
		mockN0Status.On("GetBranchStatus").Return(mockN0BranchStatus)
		mockN0Status.On("IsDirty").Return(false)

		mockN1Status := &mocks.ExecutableNodeStatus{}
		mockN1Status.On("GetPhase").Return(v1alpha1.NodePhaseRunning)
		mockN1Status.On("IsDirty").Return(false)

		mockWf := &mocks.ExecutableWorkflow{}
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN0).Return(mockN0Status)
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN1).Return(mockN1Status)
		mockWf.OnGetNodeExecutionStatus(ctx, nodeN2).Return(mockN2Status)
		mockWf.OnToNode(nodeN2).Return(upstreamN2, nil)
		mockWf.On("GetNode", nodeN0).Return(mockN0Node, true)
		mockWf.On("GetID").Return("w1")

		p, err := CanExecute(ctx, mockWf, mockWf, mockNode)
		assert.NoError(t, err)
		assert.Equal(t, PredicatePhaseNotReady, p)
	})
}
