package subworkflow

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	coreMocks "github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors"
	execMocks "github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/common"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces/mocks"
)

func TestGetSubWorkflow(t *testing.T) {
	ctx := context.TODO()

	t.Run("subworkflow", func(t *testing.T) {

		wfNode := &coreMocks.ExecutableWorkflowNode{}
		x := "x"
		wfNode.OnGetSubWorkflowRef().Return(&x)

		node := &coreMocks.ExecutableNode{}
		node.OnGetWorkflowNode().Return(wfNode)

		ectx := &execMocks.ExecutionContext{}

		swf := &coreMocks.ExecutableSubWorkflow{}
		ectx.EXPECT().FindSubWorkflow("x").Return(swf)

		nCtx := &mocks.NodeExecutionContext{}
		nCtx.OnNode().Return(node)
		nCtx.OnExecutionContext().Return(ectx)

		w, err := GetSubWorkflow(ctx, nCtx)
		assert.NoError(t, err)
		assert.Equal(t, swf, w)
	})

	t.Run("subworkflow with failure node", func(t *testing.T) {

		wfNode := &coreMocks.ExecutableWorkflowNode{}
		x := "x"
		wfNode.OnGetSubWorkflowRef().Return(&x)

		node := &coreMocks.ExecutableNode{}
		node.OnGetWorkflowNode().Return(wfNode)

		ectx := &execMocks.ExecutionContext{}

		wfFailureNode := &coreMocks.ExecutableWorkflowNode{}
		y := "y"
		wfFailureNode.OnGetSubWorkflowRef().Return(&y)
		failureNode := &coreMocks.ExecutableNode{}
		failureNode.OnGetWorkflowNode().Return(wfFailureNode)

		swf := &coreMocks.ExecutableSubWorkflow{}
		swf.OnGetOnFailureNode().Return(failureNode)
		ectx.EXPECT().FindSubWorkflow("x").Return(swf)

		nCtx := &mocks.NodeExecutionContext{}
		nCtx.OnNode().Return(node)
		nCtx.OnExecutionContext().Return(ectx)

		w, err := GetSubWorkflow(ctx, nCtx)
		assert.NoError(t, err)
		assert.Equal(t, swf, w)
	})

	t.Run("missing-subworkflow", func(t *testing.T) {

		wfNode := &coreMocks.ExecutableWorkflowNode{}
		x := "x"
		wfNode.OnGetSubWorkflowRef().Return(&x)

		node := &coreMocks.ExecutableNode{}
		node.OnGetWorkflowNode().Return(wfNode)

		ectx := &execMocks.ExecutionContext{}

		ectx.EXPECT().FindSubWorkflow("x").Return(nil)

		nCtx := &mocks.NodeExecutionContext{}
		nCtx.OnNode().Return(node)
		nCtx.OnExecutionContext().Return(ectx)

		_, err := GetSubWorkflow(ctx, nCtx)
		assert.Error(t, err)
	})
}

func Test_subworkflowHandler_HandleAbort(t *testing.T) {
	ctx := context.TODO()

	t.Run("missing-startNode", func(t *testing.T) {

		wfNode := &coreMocks.ExecutableWorkflowNode{}
		x := "x"
		wfNode.OnGetSubWorkflowRef().Return(&x)

		node := &coreMocks.ExecutableNode{}
		node.OnGetWorkflowNode().Return(wfNode)

		swf := &coreMocks.ExecutableSubWorkflow{}
		ectx := &execMocks.ExecutionContext{}
		ectx.EXPECT().FindSubWorkflow("x").Return(swf)

		ns := &coreMocks.ExecutableNodeStatus{}
		nCtx := &mocks.NodeExecutionContext{}
		nCtx.OnNode().Return(node)
		nCtx.OnExecutionContext().Return(ectx)
		nCtx.OnNodeStatus().Return(ns)
		nCtx.OnNodeID().Return("n1")

		nodeExec := &mocks.Node{}
		s := newSubworkflowHandler(nodeExec, eventConfig)
		n := &coreMocks.ExecutableNode{}
		swf.OnGetID().Return("swf")
		nodeExec.OnAbortHandlerMatch(mock.Anything, ectx, swf, mock.Anything, n, "reason").Return(nil)
		assert.Panics(t, func() {
			_ = s.HandleAbort(ctx, nCtx, "reason")
		})
	})

	t.Run("abort-error", func(t *testing.T) {
		wfNode := &coreMocks.ExecutableWorkflowNode{}
		x := "x"
		wfNode.OnGetSubWorkflowRef().Return(&x)

		node := &coreMocks.ExecutableNode{}
		node.OnGetWorkflowNode().Return(wfNode)

		swf := &coreMocks.ExecutableSubWorkflow{}
		swf.OnStartNode().Return(&coreMocks.ExecutableNode{})

		ectx := &execMocks.ExecutionContext{}
		ectx.EXPECT().FindSubWorkflow("x").Return(swf)
		ectx.EXPECT().GetParentInfo().Return(nil)

		ns := &coreMocks.ExecutableNodeStatus{}
		nCtx := &mocks.NodeExecutionContext{}
		nCtx.OnNode().Return(node)
		nCtx.OnExecutionContext().Return(ectx)
		nCtx.OnNodeStatus().Return(ns)
		nCtx.OnNodeID().Return("n1")
		nCtx.OnCurrentAttempt().Return(uint32(1))

		nodeExec := &mocks.Node{}
		s := newSubworkflowHandler(nodeExec, eventConfig)
		n := &coreMocks.ExecutableNode{}
		swf.OnGetID().Return("swf")
		newParentInfo, _ := common.CreateParentInfo(nil, nCtx.NodeID(), nCtx.CurrentAttempt())
		expectedExecContext := executors.NewExecutionContextWithParentInfo(nCtx.ExecutionContext(), newParentInfo)
		nodeExec.OnAbortHandlerMatch(mock.Anything, expectedExecContext, swf, mock.Anything, n, "reason").Return(fmt.Errorf("err"))
		assert.Error(t, s.HandleAbort(ctx, nCtx, "reason"))
	})

	t.Run("abort-success", func(t *testing.T) {

		wfNode := &coreMocks.ExecutableWorkflowNode{}
		x := "x"
		wfNode.OnGetSubWorkflowRef().Return(&x)

		node := &coreMocks.ExecutableNode{}
		node.OnGetWorkflowNode().Return(wfNode)

		swf := &coreMocks.ExecutableSubWorkflow{}
		swf.OnStartNode().Return(&coreMocks.ExecutableNode{})

		ectx := &execMocks.ExecutionContext{}
		ectx.EXPECT().FindSubWorkflow("x").Return(swf)
		ectx.EXPECT().GetParentInfo().Return(nil)

		ns := &coreMocks.ExecutableNodeStatus{}
		nCtx := &mocks.NodeExecutionContext{}
		nCtx.OnNode().Return(node)
		nCtx.OnExecutionContext().Return(ectx)
		nCtx.OnNodeStatus().Return(ns)
		nCtx.OnNodeID().Return("n1")
		nCtx.OnCurrentAttempt().Return(uint32(1))

		nodeExec := &mocks.Node{}
		s := newSubworkflowHandler(nodeExec, eventConfig)
		n := &coreMocks.ExecutableNode{}
		swf.OnGetID().Return("swf")
		newParentInfo, _ := common.CreateParentInfo(nil, nCtx.NodeID(), nCtx.CurrentAttempt())
		expectedExecContext := executors.NewExecutionContextWithParentInfo(nCtx.ExecutionContext(), newParentInfo)
		nodeExec.OnAbortHandlerMatch(mock.Anything, expectedExecContext, swf, mock.Anything, n, "reason").Return(nil)
		assert.NoError(t, s.HandleAbort(ctx, nCtx, "reason"))
	})
}
