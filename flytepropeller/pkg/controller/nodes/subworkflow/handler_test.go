package subworkflow

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	mocks4 "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	mocks2 "github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/config"
	execMocks "github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/handler"
	mocks3 "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces/mocks"
	mocks5 "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/recovery/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/subworkflow/launchplan"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/subworkflow/launchplan/mocks"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

type workflowNodeStateHolder struct {
	s handler.WorkflowNodeState
}

var eventConfig = &config.EventConfig{
	RawOutputPolicy: config.RawOutputPolicyReference,
}

func (t *workflowNodeStateHolder) ClearNodeStatus() {
}

func (t *workflowNodeStateHolder) PutTaskNodeState(s handler.TaskNodeState) error {
	panic("not implemented")
}

func (t workflowNodeStateHolder) PutBranchNode(s handler.BranchNodeState) error {
	panic("not implemented")
}

func (t *workflowNodeStateHolder) PutWorkflowNodeState(s handler.WorkflowNodeState) error {
	t.s = s
	return nil
}

func (t workflowNodeStateHolder) PutDynamicNodeState(s handler.DynamicNodeState) error {
	panic("not implemented")
}

func (t workflowNodeStateHolder) PutGateNodeState(s handler.GateNodeState) error {
	panic("not implemented")
}

func (t workflowNodeStateHolder) PutArrayNodeState(s handler.ArrayNodeState) error {
	panic("not implemented")
}

var wfExecID = &core.WorkflowExecutionIdentifier{
	Project: "project",
	Domain:  "domain",
	Name:    "name",
}

func createNodeContextWithVersion(phase v1alpha1.WorkflowNodePhase, n v1alpha1.ExecutableNode, s v1alpha1.ExecutableNodeStatus, version v1alpha1.EventVersion) *mocks3.NodeExecutionContext {

	wfNodeState := handler.WorkflowNodeState{}
	state := &workflowNodeStateHolder{s: wfNodeState}

	nm := &mocks3.NodeExecutionMetadata{}
	nm.OnGetAnnotations().Return(map[string]string{})
	nm.OnGetNodeExecutionID().Return(&core.NodeExecutionIdentifier{
		ExecutionId: wfExecID,
		NodeId:      n.GetID(),
	})
	nm.OnGetK8sServiceAccount().Return("service-account")
	nm.OnGetLabels().Return(map[string]string{})
	nm.OnGetNamespace().Return("namespace")
	nm.OnGetOwnerID().Return(types.NamespacedName{Namespace: "namespace", Name: "name"})
	nm.OnGetOwnerReference().Return(v1.OwnerReference{
		Kind: "sample",
		Name: "name",
	})

	ir := &mocks4.InputReader{}
	inputs := &core.LiteralMap{}
	ir.OnGetMatch(mock.Anything).Return(inputs, nil)

	nCtx := &mocks3.NodeExecutionContext{}
	nCtx.OnNode().Return(n)
	nCtx.OnNodeExecutionMetadata().Return(nm)
	nCtx.OnInputReader().Return(ir)
	nCtx.OnCurrentAttempt().Return(uint32(1))
	nCtx.OnNodeID().Return(n.GetID())
	nCtx.OnEnqueueOwnerFunc().Return(nil)
	nCtx.OnNodeStatus().Return(s)

	nr := &mocks3.NodeStateReader{}
	nr.OnGetWorkflowNodeState().Return(handler.WorkflowNodeState{
		Phase: phase,
	})
	nCtx.OnNodeStateReader().Return(nr)
	nCtx.OnNodeStateWriter().Return(state)

	ex := &execMocks.ExecutionContext{}
	ex.EXPECT().GetEventVersion().Return(version)
	ex.EXPECT().GetParentInfo().Return(nil)
	ex.EXPECT().GetName().Return("name")
	ex.EXPECT().GetExecutionConfig().Return(v1alpha1.ExecutionConfig{})
	ex.EXPECT().IncrementParallelism().Return(1)
	ex.EXPECT().GetSecurityContext().Return(core.SecurityContext{})
	ex.EXPECT().GetAnnotations().Return(nil)
	ex.EXPECT().GetLabels().Return(nil)
	ex.EXPECT().GetRawOutputDataConfig().Return(v1alpha1.RawOutputDataConfig{})
	ex.EXPECT().GetDefinitionVersion().Return(v1alpha1.WorkflowDefinitionVersion1)

	nCtx.OnExecutionContext().Return(ex)

	return nCtx
}

func createNodeContextV1(phase v1alpha1.WorkflowNodePhase, n v1alpha1.ExecutableNode, s v1alpha1.ExecutableNodeStatus) *mocks3.NodeExecutionContext {
	return createNodeContextWithVersion(phase, n, s, v1alpha1.EventVersion1)
}

func createNodeContext(phase v1alpha1.WorkflowNodePhase, n v1alpha1.ExecutableNode, s v1alpha1.ExecutableNodeStatus) *mocks3.NodeExecutionContext {
	return createNodeContextWithVersion(phase, n, s, v1alpha1.EventVersion0)
}

func TestWorkflowNodeHandler_StartNode_Launchplan(t *testing.T) {
	ctx := context.TODO()

	attempts := uint32(1)

	lpID := &core.Identifier{
		Project:      "p",
		Domain:       "d",
		Name:         "n",
		Version:      "v",
		ResourceType: core.ResourceType_LAUNCH_PLAN,
	}
	mockWfNode := &mocks2.ExecutableWorkflowNode{}
	mockWfNode.OnGetLaunchPlanRefID().Return(&v1alpha1.Identifier{
		Identifier: lpID,
	})
	mockWfNode.OnGetSubWorkflowRef().Return(nil)

	mockNode := &mocks2.ExecutableNode{}
	mockNode.OnGetID().Return("n1")
	mockNode.OnGetWorkflowNode().Return(mockWfNode)

	mockNodeStatus := &mocks2.ExecutableNodeStatus{}
	mockNodeStatus.OnGetAttempts().Return(attempts)
	wfStatus := &mocks2.MutableWorkflowNodeStatus{}
	mockNodeStatus.OnGetOrCreateWorkflowStatus().Return(wfStatus)
	recoveryClient := &mocks5.Client{}

	t.Run("happy v0", func(t *testing.T) {
		mockLPExec := &mocks.Executor{}
		h := New(nil, mockLPExec, recoveryClient, eventConfig, promutils.NewTestScope())
		mockLPExec.OnLaunchMatch(
			ctx,
			mock.MatchedBy(func(o launchplan.LaunchContext) bool {
				return o.ParentNodeExecution.NodeId == mockNode.GetID() &&
					o.ParentNodeExecution.ExecutionId == wfExecID
			}),
			mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
				return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
			}),
			mock.MatchedBy(func(o *core.Identifier) bool { return lpID == o }),
			mock.MatchedBy(func(o *core.LiteralMap) bool { return o.Literals == nil }),
		).Return(nil)

		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseUndefined, mockNode, mockNodeStatus)
		s, err := h.Handle(ctx, nCtx)
		assert.NoError(t, err)
		assert.Equal(t, handler.EPhaseRunning, s.Info().GetPhase())
		c := nCtx.ExecutionContext().(*execMocks.ExecutionContext)
		c.AssertCalled(t, "IncrementParallelism")
	})

	t.Run("happy v1", func(t *testing.T) {

		mockLPExec := &mocks.Executor{}
		h := New(nil, mockLPExec, recoveryClient, eventConfig, promutils.NewTestScope())
		mockLPExec.OnLaunchMatch(
			ctx,
			mock.MatchedBy(func(o launchplan.LaunchContext) bool {
				return o.ParentNodeExecution.NodeId == mockNode.GetID() &&
					o.ParentNodeExecution.ExecutionId == wfExecID
			}),
			mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
				return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
			}),
			mock.MatchedBy(func(o *core.Identifier) bool { return lpID == o }),
			mock.MatchedBy(func(o *core.LiteralMap) bool { return o.Literals == nil }),
		).Return(nil)

		nCtx := createNodeContextV1(v1alpha1.WorkflowNodePhaseUndefined, mockNode, mockNodeStatus)
		s, err := h.Handle(ctx, nCtx)
		assert.NoError(t, err)
		assert.Equal(t, handler.EPhaseRunning, s.Info().GetPhase())
		c := nCtx.ExecutionContext().(*execMocks.ExecutionContext)
		c.AssertCalled(t, "IncrementParallelism")
	})
}

func TestWorkflowNodeHandler_CheckNodeStatus(t *testing.T) {
	ctx := context.TODO()

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
	mockWfNode.OnGetLaunchPlanRefID().Return(&v1alpha1.Identifier{
		Identifier: lpID,
	})
	mockWfNode.OnGetSubWorkflowRef().Return(nil)

	mockNode := &mocks2.ExecutableNode{}
	mockNode.OnGetID().Return("n1")
	mockNode.OnGetWorkflowNode().Return(mockWfNode)

	mockNodeStatus := &mocks2.ExecutableNodeStatus{}
	mockNodeStatus.OnGetAttempts().Return(attempts)
	mockNodeStatus.OnGetDataDir().Return(dataDir)
	recoveryClient := &mocks5.Client{}

	t.Run("stillRunning V0", func(t *testing.T) {

		mockLPExec := &mocks.Executor{}

		h := New(nil, mockLPExec, recoveryClient, eventConfig, promutils.NewTestScope())
		mockLPExec.OnGetStatusMatch(
			ctx,
			mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
				return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
			}),
		).Return(&admin.ExecutionClosure{
			Phase: core.WorkflowExecution_RUNNING,
		}, &core.LiteralMap{}, nil)

		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseExecuting, mockNode, mockNodeStatus)
		s, err := h.Handle(ctx, nCtx)
		assert.NoError(t, err)
		assert.Equal(t, handler.EPhaseRunning, s.Info().GetPhase())
		c := nCtx.ExecutionContext().(*execMocks.ExecutionContext)
		c.AssertCalled(t, "IncrementParallelism")
	})
	t.Run("stillRunning V1", func(t *testing.T) {

		mockLPExec := &mocks.Executor{}

		h := New(nil, mockLPExec, recoveryClient, eventConfig, promutils.NewTestScope())
		mockLPExec.OnGetStatusMatch(
			ctx,
			mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
				return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
			}),
		).Return(&admin.ExecutionClosure{
			Phase: core.WorkflowExecution_RUNNING,
		}, &core.LiteralMap{}, nil)

		nCtx := createNodeContextV1(v1alpha1.WorkflowNodePhaseExecuting, mockNode, mockNodeStatus)
		s, err := h.Handle(ctx, nCtx)
		assert.NoError(t, err)
		assert.Equal(t, handler.EPhaseRunning, s.Info().GetPhase())
		c := nCtx.ExecutionContext().(*execMocks.ExecutionContext)
		c.AssertCalled(t, "IncrementParallelism")
	})
}

func TestWorkflowNodeHandler_AbortNode(t *testing.T) {
	ctx := context.TODO()

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
	mockWfNode.OnGetLaunchPlanRefID().Return(&v1alpha1.Identifier{
		Identifier: lpID,
	})
	mockWfNode.OnGetSubWorkflowRef().Return(nil)

	mockNode := &mocks2.ExecutableNode{}
	mockNode.OnGetID().Return("n1")
	mockNode.OnGetWorkflowNode().Return(mockWfNode)

	mockNodeStatus := &mocks2.ExecutableNodeStatus{}
	mockNodeStatus.OnGetAttempts().Return(attempts)
	mockNodeStatus.OnGetDataDir().Return(dataDir)
	recoveryClient := &mocks5.Client{}

	t.Run("abort v0", func(t *testing.T) {

		mockLPExec := &mocks.Executor{}
		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseExecuting, mockNode, mockNodeStatus)

		h := New(nil, mockLPExec, recoveryClient, eventConfig, promutils.NewTestScope())
		mockLPExec.OnKillMatch(
			ctx,
			mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
				return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
			}),
			mock.AnythingOfType(reflect.String.String()),
		).Return(nil)

		eCtx := &execMocks.ExecutionContext{}
		eCtx.EXPECT().GetDefinitionVersion().Return(v1alpha1.WorkflowDefinitionVersion1)
		eCtx.EXPECT().GetName().Return("test")
		nCtx.OnExecutionContext().Return(eCtx)
		err := h.Abort(ctx, nCtx, "test")
		assert.NoError(t, err)
	})

	t.Run("abort v1", func(t *testing.T) {

		mockLPExec := &mocks.Executor{}
		nCtx := createNodeContextV1(v1alpha1.WorkflowNodePhaseExecuting, mockNode, mockNodeStatus)

		h := New(nil, mockLPExec, recoveryClient, eventConfig, promutils.NewTestScope())
		mockLPExec.OnKillMatch(
			ctx,
			mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
				return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
			}),
			mock.AnythingOfType(reflect.String.String()),
		).Return(nil)

		eCtx := &execMocks.ExecutionContext{}
		eCtx.EXPECT().GetDefinitionVersion().Return(v1alpha1.WorkflowDefinitionVersion1)
		eCtx.EXPECT().GetName().Return("test")
		nCtx.OnExecutionContext().Return(eCtx)
		err := h.Abort(ctx, nCtx, "test")
		assert.NoError(t, err)
	})
	t.Run("abort-fail", func(t *testing.T) {

		mockLPExec := &mocks.Executor{}
		expectedErr := fmt.Errorf("fail")
		h := New(nil, mockLPExec, recoveryClient, eventConfig, promutils.NewTestScope())
		mockLPExec.OnKillMatch(
			ctx,
			mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
				return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
			}),
			mock.AnythingOfType(reflect.String.String()),
		).Return(expectedErr)

		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseExecuting, mockNode, mockNodeStatus)
		eCtx := &execMocks.ExecutionContext{}
		eCtx.EXPECT().GetDefinitionVersion().Return(v1alpha1.WorkflowDefinitionVersion1)
		eCtx.EXPECT().GetName().Return("test")
		nCtx.OnExecutionContext().Return(eCtx)

		err := h.Abort(ctx, nCtx, "test")
		assert.Error(t, err)
		assert.Equal(t, err, expectedErr)
	})
}

func init() {
	labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey)
}
