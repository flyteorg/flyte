package subworkflow

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	coreMocks "github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
	execMocks "github.com/lyft/flytepropeller/pkg/controller/executors/mocks"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler/mocks"
)

func Test_subworkflowHandler_HandleAbort(t *testing.T) {
	ctx := context.TODO()

	t.Run("missing-subworkflow", func(t *testing.T) {
		nCtx := &mocks.NodeExecutionContext{}
		nodeExec := &execMocks.Node{}
		s := newSubworkflowHandler(nodeExec)
		wf := &coreMocks.ExecutableWorkflow{}
		wf.OnFindSubWorkflow("x").Return(nil)
		nCtx.OnNodeID().Return("n1")
		assert.Error(t, s.HandleAbort(ctx, nCtx, wf, "x", "reason"))
	})

	t.Run("missing-startNode", func(t *testing.T) {
		nCtx := &mocks.NodeExecutionContext{}
		nodeExec := &execMocks.Node{}
		s := newSubworkflowHandler(nodeExec)
		wf := &coreMocks.ExecutableWorkflow{}
		st := &coreMocks.ExecutableNodeStatus{}
		swf := &coreMocks.ExecutableSubWorkflow{}
		wf.OnFindSubWorkflow("x").Return(swf)
		wf.OnGetNodeExecutionStatus(ctx, "n1").Return(st)
		nCtx.OnNodeID().Return("n1")
		swf.OnStartNode().Return(nil)
		assert.Error(t, s.HandleAbort(ctx, nCtx, wf, "x", "reason"))
	})

	t.Run("abort-error", func(t *testing.T) {
		nCtx := &mocks.NodeExecutionContext{}
		nodeExec := &execMocks.Node{}
		s := newSubworkflowHandler(nodeExec)
		wf := &coreMocks.ExecutableWorkflow{}
		st := &coreMocks.ExecutableNodeStatus{}
		swf := &coreMocks.ExecutableSubWorkflow{}
		wf.OnFindSubWorkflow("x").Return(swf)
		wf.OnGetNodeExecutionStatus(ctx, "n1").Return(st)
		nCtx.OnNodeID().Return("n1")
		n := &coreMocks.ExecutableNode{}
		swf.OnStartNode().Return(n)
		nodeExec.OnAbortHandler(ctx, wf, n, "reason").Return(fmt.Errorf("err"))
		assert.Error(t, s.HandleAbort(ctx, nCtx, wf, "x", "reason"))
	})

	t.Run("abort-success", func(t *testing.T) {
		nCtx := &mocks.NodeExecutionContext{}
		nodeExec := &execMocks.Node{}
		s := newSubworkflowHandler(nodeExec)
		wf := &coreMocks.ExecutableWorkflow{}
		st := &coreMocks.ExecutableNodeStatus{}
		swf := &coreMocks.ExecutableSubWorkflow{}
		wf.OnFindSubWorkflow("x").Return(swf)
		wf.OnGetNodeExecutionStatus(ctx, "n1").Return(st)
		nCtx.OnNodeID().Return("n1")
		n := &coreMocks.ExecutableNode{}
		swf.OnStartNode().Return(n)
		swf.OnGetID().Return("swf")
		nodeExec.OnAbortHandlerMatch(mock.Anything, mock.MatchedBy(func(wf v1alpha1.ExecutableWorkflow) bool {
			return wf.GetID() == swf.GetID()
		}), n, mock.Anything).Return(nil)
		assert.NoError(t, s.HandleAbort(ctx, nCtx, wf, "x", "reason"))
	})
}
