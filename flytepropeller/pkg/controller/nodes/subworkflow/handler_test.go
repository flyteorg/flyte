package subworkflow

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	mocks2 "github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/subworkflow/launchplan"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/subworkflow/launchplan/mocks"
	"github.com/lyft/flytestdlib/contextutils"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/promutils/labeled"
	"github.com/lyft/flytestdlib/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWorkflowNodeHandler_StartNode_Launchplan(t *testing.T) {
	ctx := context.TODO()

	nodeID := "n1"
	attempts := uint32(1)

	lpID := &core.Identifier{
		Project:      "p",
		Domain:       "d",
		Name:         "n",
		Version:      "v",
		ResourceType: core.ResourceType_LAUNCH_PLAN,
	}
	mockWfNode := &mocks2.ExecutableWorkflowNode{}
	mockWfNode.On("GetLaunchPlanRefID").Return(&v1alpha1.Identifier{
		Identifier: lpID,
	})
	mockWfNode.On("GetSubWorkflowRef").Return(nil)

	mockNode := &mocks2.ExecutableNode{}
	mockNode.On("GetID").Return("n1")
	mockNode.On("GetWorkflowNode").Return(mockWfNode)

	mockNodeStatus := &mocks2.ExecutableNodeStatus{}
	mockNodeStatus.On("GetAttempts").Return(attempts)
	wfStatus := &mocks2.MutableWorkflowNodeStatus{}
	mockNodeStatus.On("GetOrCreateWorkflowStatus").Return(wfStatus)
	wfStatus.On("SetWorkflowExecutionName",
		mock.MatchedBy(func(name string) bool {
			return name == "x-n1-1"
		}),
	).Return()
	parentID := &core.WorkflowExecutionIdentifier{
		Name:    "x",
		Domain:  "y",
		Project: "z",
	}
	mockWf := &mocks2.ExecutableWorkflow{}
	mockWf.On("GetNodeExecutionStatus", nodeID).Return(mockNodeStatus)
	mockWf.On("GetExecutionID").Return(v1alpha1.WorkflowExecutionIdentifier{
		WorkflowExecutionIdentifier: parentID,
	})

	ni := &core.LiteralMap{}

	t.Run("happy", func(t *testing.T) {

		mockLPExec := &mocks.Executor{}

		h := New(nil, nil, mockLPExec, nil, nil, promutils.NewTestScope())
		mockLPExec.On("Launch",
			ctx,
			mock.MatchedBy(func(o launchplan.LaunchContext) bool {
				return o.ParentNodeExecution.NodeId == mockNode.GetID() &&
					o.ParentNodeExecution.ExecutionId == parentID
			}),
			mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
				return o.Project == parentID.Project && o.Domain == parentID.Domain
			}),
			mock.MatchedBy(func(o *core.Identifier) bool { return lpID == o }),
			mock.MatchedBy(func(o *core.LiteralMap) bool { return ni == o }),
		).Return(nil)

		s, err := h.StartNode(ctx, mockWf, mockNode, ni)
		assert.NoError(t, err)
		assert.Equal(t, s.Phase, handler.PhaseRunning)
	})
}

func TestWorkflowNodeHandler_CheckNodeStatus(t *testing.T) {
	ctx := context.TODO()

	nodeID := "n1"
	attempts := uint32(1)
	dataDir := storage.DataReference("data")

	lpID := &core.Identifier{
		Project:      "p",
		Domain:       "d",
		Name:         "n",
		Version:      "v",
		ResourceType: core.ResourceType_LAUNCH_PLAN,
	}
	mockWfNode := &mocks2.ExecutableWorkflowNode{}
	mockWfNode.On("GetLaunchPlanRefID").Return(&v1alpha1.Identifier{
		Identifier: lpID,
	})
	mockWfNode.On("GetSubWorkflowRef").Return(nil)

	mockNode := &mocks2.ExecutableNode{}
	mockNode.On("GetID").Return("n1")
	mockNode.On("GetWorkflowNode").Return(mockWfNode)

	mockNodeStatus := &mocks2.ExecutableNodeStatus{}
	mockNodeStatus.On("GetAttempts").Return(attempts)
	mockNodeStatus.On("GetDataDir").Return(dataDir)

	parentID := &core.WorkflowExecutionIdentifier{
		Name:    "x",
		Domain:  "y",
		Project: "z",
	}
	mockWf := &mocks2.ExecutableWorkflow{}
	mockWf.On("GetNodeExecutionStatus", nodeID).Return(mockNodeStatus)
	mockWf.On("GetExecutionID").Return(v1alpha1.WorkflowExecutionIdentifier{
		WorkflowExecutionIdentifier: parentID,
	})

	t.Run("stillRunning", func(t *testing.T) {

		mockLPExec := &mocks.Executor{}

		h := New(nil, nil, mockLPExec, nil, nil, promutils.NewTestScope())
		mockLPExec.On("GetStatus",
			ctx,
			mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
				return o.Project == parentID.Project && o.Domain == parentID.Domain
			}),
		).Return(&admin.ExecutionClosure{
			Phase: core.WorkflowExecution_RUNNING,
		}, nil)

		s, err := h.CheckNodeStatus(ctx, mockWf, mockNode, nil)
		assert.NoError(t, err)
		assert.Equal(t, s.Phase, handler.PhaseRunning)
	})
}

func TestWorkflowNodeHandler_AbortNode(t *testing.T) {
	ctx := context.TODO()

	nodeID := "n1"
	attempts := uint32(1)
	dataDir := storage.DataReference("data")

	lpID := &core.Identifier{
		Project:      "p",
		Domain:       "d",
		Name:         "n",
		Version:      "v",
		ResourceType: core.ResourceType_LAUNCH_PLAN,
	}
	mockWfNode := &mocks2.ExecutableWorkflowNode{}
	mockWfNode.On("GetLaunchPlanRefID").Return(&v1alpha1.Identifier{
		Identifier: lpID,
	})
	mockWfNode.On("GetSubWorkflowRef").Return(nil)

	mockNode := &mocks2.ExecutableNode{}
	mockNode.On("GetID").Return("n1")
	mockNode.On("GetWorkflowNode").Return(mockWfNode)

	mockNodeStatus := &mocks2.ExecutableNodeStatus{}
	mockNodeStatus.On("GetAttempts").Return(attempts)
	mockNodeStatus.On("GetDataDir").Return(dataDir)

	parentID := &core.WorkflowExecutionIdentifier{
		Name:    "x",
		Domain:  "y",
		Project: "z",
	}
	mockWf := &mocks2.ExecutableWorkflow{}
	mockWf.On("GetNodeExecutionStatus", nodeID).Return(mockNodeStatus)
	mockWf.On("GetName").Return("test")
	mockWf.On("GetExecutionID").Return(v1alpha1.WorkflowExecutionIdentifier{
		WorkflowExecutionIdentifier: parentID,
	})

	t.Run("abort", func(t *testing.T) {

		mockLPExec := &mocks.Executor{}

		h := New(nil, nil, mockLPExec, nil, nil, promutils.NewTestScope())
		mockLPExec.On("Kill",
			ctx,
			mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
				return o.Project == parentID.Project && o.Domain == parentID.Domain
			}),
			mock.AnythingOfType(reflect.String.String()),
		).Return(nil)

		err := h.AbortNode(ctx, mockWf, mockNode)
		assert.NoError(t, err)
	})

	t.Run("abort-fail", func(t *testing.T) {

		mockLPExec := &mocks.Executor{}
		expectedErr := fmt.Errorf("fail")
		h := New(nil, nil, mockLPExec, nil, nil, promutils.NewTestScope())
		mockLPExec.On("Kill",
			ctx,
			mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
				return o.Project == parentID.Project && o.Domain == parentID.Domain
			}),
			mock.AnythingOfType(reflect.String.String()),
		).Return(expectedErr)

		err := h.AbortNode(ctx, mockWf, mockNode)
		assert.Error(t, err)
		assert.Equal(t, err, expectedErr)
	})
}

func init() {
	labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey)
}
