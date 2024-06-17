package dynamic

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io"
	ioMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	flyteMocks "github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/config"
	executorMocks "github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/dynamic/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/handler"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces"
	nodeMocks "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces/mocks"
	lpMocks "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/subworkflow/launchplan/mocks"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

type dynamicNodeStateHolder struct {
	s handler.DynamicNodeState
}

func (t *dynamicNodeStateHolder) ClearNodeStatus() {
}

func (t *dynamicNodeStateHolder) PutTaskNodeState(s handler.TaskNodeState) error {
	panic("not implemented")
}

func (t dynamicNodeStateHolder) PutBranchNode(s handler.BranchNodeState) error {
	panic("not implemented")
}

func (t dynamicNodeStateHolder) PutWorkflowNodeState(s handler.WorkflowNodeState) error {
	panic("not implemented")
}

func (t *dynamicNodeStateHolder) PutDynamicNodeState(s handler.DynamicNodeState) error {
	t.s = s
	return nil
}

func (t dynamicNodeStateHolder) PutGateNodeState(s handler.GateNodeState) error {
	panic("not implemented")
}

func (t dynamicNodeStateHolder) PutArrayNodeState(s handler.ArrayNodeState) error {
	panic("not implemented")
}

var tID = "task-1"

var eventConfig = &config.EventConfig{
	RawOutputPolicy: config.RawOutputPolicyReference,
}

func Test_dynamicNodeHandler_Handle_Parent(t *testing.T) {
	createNodeContext := func(ttype string) *nodeMocks.NodeExecutionContext {
		wfExecID := &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		}

		res := &v12.ResourceRequirements{}
		n := &flyteMocks.ExecutableNode{}
		n.OnGetResources().Return(res)
		n.OnGetID().Return("n1")

		nm := &nodeMocks.NodeExecutionMetadata{}
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

		taskID := &core.Identifier{}
		tk := &core.TaskTemplate{
			Id:   taskID,
			Type: "test",
			Metadata: &core.TaskMetadata{
				Discoverable: true,
			},
			Interface: &core.TypedInterface{
				Outputs: &core.VariableMap{
					Variables: map[string]*core.Variable{
						"x": {
							Type: &core.LiteralType{
								Type: &core.LiteralType_Simple{
									Simple: core.SimpleType_BOOLEAN,
								},
							},
						},
					},
				},
			},
		}
		tr := &nodeMocks.TaskReader{}
		tr.OnGetTaskID().Return(taskID)
		tr.OnGetTaskType().Return(ttype)
		tr.OnReadMatch(mock.Anything).Return(tk, nil)

		ns := &flyteMocks.ExecutableNodeStatus{}
		ns.OnGetDataDir().Return(storage.DataReference("data-dir"))
		ns.OnGetOutputDir().Return(storage.DataReference("data-dir"))

		dataStore, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
		assert.NoError(t, err)

		ir := &ioMocks.InputReader{}
		nCtx := &nodeMocks.NodeExecutionContext{}
		nCtx.OnNodeExecutionMetadata().Return(nm)
		nCtx.OnNode().Return(n)
		nCtx.OnInputReader().Return(ir)
		nCtx.OnCurrentAttempt().Return(uint32(1))
		nCtx.OnTaskReader().Return(tr)
		nCtx.OnNodeStatus().Return(ns)
		nCtx.OnNodeID().Return("n1")
		nCtx.OnEnqueueOwnerFunc().Return(nil)
		nCtx.OnDataStore().Return(dataStore)

		r := &nodeMocks.NodeStateReader{}
		r.OnGetDynamicNodeState().Return(handler.DynamicNodeState{})
		nCtx.OnNodeStateReader().Return(r)
		return nCtx
	}

	i := &handler.ExecutionInfo{}
	type args struct {
		isDynamic bool
		trns      handler.Transition
		isErr     bool
	}
	type want struct {
		p     handler.EPhase
		info  *handler.ExecutionInfo
		isErr bool
		phase v1alpha1.DynamicNodePhase
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{"error", args{isErr: true}, want{isErr: true}},
		{"success-non-parent", args{trns: handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoSuccess(nil))}, want{p: handler.EPhaseSuccess, phase: v1alpha1.DynamicNodePhaseNone}},
		{"running-non-parent", args{trns: handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRunning(i))}, want{p: handler.EPhaseRunning, info: i}},
		{"retryfailure-non-parent", args{trns: handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRetryableFailure(core.ExecutionError_SYSTEM, "x", "y", i))}, want{p: handler.EPhaseRetryableFailure, info: i}},
		{"failure-non-parent", args{trns: handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_USER, "x", "y", i))}, want{p: handler.EPhaseFailed, info: i}},
		{"success-parent", args{isDynamic: true, trns: handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoSuccess(nil))}, want{p: handler.EPhaseRunning, phase: v1alpha1.DynamicNodePhaseParentFinalizing}},
		{"running-parent", args{isDynamic: true, trns: handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRunning(i))}, want{p: handler.EPhaseRunning, info: i}},
		{"retryfailure-parent", args{isDynamic: true, trns: handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoRetryableFailure(core.ExecutionError_SYSTEM, "x", "y", i))}, want{p: handler.EPhaseRetryableFailure, info: i}},
		{"failure-non-parent", args{isDynamic: true, trns: handler.DoTransition(handler.TransitionTypeEphemeral, handler.PhaseInfoFailure(core.ExecutionError_USER, "x", "y", i))}, want{p: handler.EPhaseFailed, info: i}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nCtx := createNodeContext("test")
			s := &dynamicNodeStateHolder{}
			nCtx.OnNodeStateWriter().Return(s)
			if tt.args.isDynamic {
				f, err := nCtx.DataStore().ConstructReference(context.TODO(), nCtx.NodeStatus().GetDataDir(), "futures.pb")
				assert.NoError(t, err)
				dj := &core.DynamicJobSpec{}
				assert.NoError(t, nCtx.DataStore().WriteProtobuf(context.TODO(), f, storage.Options{}, dj))
			}
			h := &mocks.TaskNodeHandler{}
			mockLPLauncher := &lpMocks.Reader{}
			n := &nodeMocks.Node{}
			if tt.args.isErr {
				h.OnHandleMatch(mock.Anything, mock.Anything).Return(handler.UnknownTransition, fmt.Errorf("error"))
			} else {
				h.OnHandleMatch(mock.Anything, mock.Anything).Return(tt.args.trns, nil)
			}
			d := New(h, n, mockLPLauncher, eventConfig, promutils.NewTestScope())
			got, err := d.Handle(context.TODO(), nCtx)
			if (err != nil) != tt.want.isErr {
				t.Errorf("Handle() error = %v, wantErr %v", err, tt.want.isErr)
				return
			}
			if err == nil {
				assert.Equal(t, tt.want.p.String(), got.Info().GetPhase().String())
				assert.Equal(t, tt.want.info, got.Info().GetInfo())
				assert.Equal(t, tt.want.phase, s.s.Phase)
			}
		})
	}

}

func Test_dynamicNodeHandler_Handle_ParentFinalize(t *testing.T) {
	createNodeContext := func(ttype string) *nodeMocks.NodeExecutionContext {
		wfExecID := &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		}

		nm := &nodeMocks.NodeExecutionMetadata{}
		nm.On("GetAnnotations").Return(map[string]string{})
		nm.On("GetExecutionID").Return(v1alpha1.WorkflowExecutionIdentifier{
			WorkflowExecutionIdentifier: wfExecID,
		})
		nm.On("GetK8sServiceAccount").Return("service-account")
		nm.On("GetLabels").Return(map[string]string{})
		nm.On("GetNamespace").Return("namespace")
		nm.On("GetOwnerID").Return(types.NamespacedName{Namespace: "namespace", Name: "name"})
		nm.On("GetOwnerReference").Return(v1.OwnerReference{
			Kind: "sample",
			Name: "name",
		})

		taskID := &core.Identifier{}
		tk := &core.TaskTemplate{
			Id:   taskID,
			Type: "test",
			Metadata: &core.TaskMetadata{
				Discoverable: true,
			},
			Interface: &core.TypedInterface{
				Outputs: &core.VariableMap{
					Variables: map[string]*core.Variable{
						"x": {
							Type: &core.LiteralType{
								Type: &core.LiteralType_Simple{
									Simple: core.SimpleType_BOOLEAN,
								},
							},
						},
					},
				},
			},
		}
		tr := &nodeMocks.TaskReader{}
		tr.On("GetTaskID").Return(taskID)
		tr.On("GetTaskType").Return(ttype)
		tr.On("Read", mock.Anything).Return(tk, nil)

		ns := &flyteMocks.ExecutableNodeStatus{}
		ns.On("GetDataDir").Return(storage.DataReference("data-dir"))
		ns.On("GetOutputDir").Return(storage.DataReference("data-dir"))

		res := &v12.ResourceRequirements{}
		n := &flyteMocks.ExecutableNode{}
		n.On("GetResources").Return(res)

		dataStore, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
		assert.NoError(t, err)

		ir := &ioMocks.InputReader{}
		nCtx := &nodeMocks.NodeExecutionContext{}
		nCtx.On("NodeExecutionMetadata").Return(nm)
		nCtx.On("Node").Return(n)
		nCtx.On("InputReader").Return(ir)
		nCtx.On("DataReferenceConstructor").Return(storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope()))
		nCtx.On("CurrentAttempt").Return(uint32(1))
		nCtx.On("TaskReader").Return(tr)
		nCtx.On("MaxDatasetSizeBytes").Return(int64(1))
		nCtx.On("NodeStatus").Return(ns)
		nCtx.On("NodeID").Return("n1")
		nCtx.On("EnqueueOwner").Return(nil)
		nCtx.OnDataStore().Return(dataStore)

		r := &nodeMocks.NodeStateReader{}
		r.On("GetDynamicNodeState").Return(handler.DynamicNodeState{
			Phase: v1alpha1.DynamicNodePhaseParentFinalizing,
		})
		nCtx.OnNodeStateReader().Return(r)

		return nCtx
	}

	t.Run("parent-finalize-success", func(t *testing.T) {
		nCtx := createNodeContext("test")
		s := &dynamicNodeStateHolder{
			s: handler.DynamicNodeState{Phase: v1alpha1.DynamicNodePhaseParentFinalizing},
		}
		nCtx.OnNodeStateWriter().Return(s)
		f, err := nCtx.DataStore().ConstructReference(context.TODO(), nCtx.NodeStatus().GetDataDir(), "futures.pb")
		assert.NoError(t, err)
		dj := &core.DynamicJobSpec{}
		mockLPLauncher := &lpMocks.Reader{}
		n := &nodeMocks.Node{}
		assert.NoError(t, nCtx.DataStore().WriteProtobuf(context.TODO(), f, storage.Options{}, dj))
		h := &mocks.TaskNodeHandler{}
		h.OnFinalizeMatch(mock.Anything, mock.Anything).Return(nil)
		d := New(h, n, mockLPLauncher, eventConfig, promutils.NewTestScope())
		got, err := d.Handle(context.TODO(), nCtx)
		assert.NoError(t, err)
		assert.Equal(t, handler.EPhaseRunning.String(), got.Info().GetPhase().String())
	})

	t.Run("parent-finalize-error", func(t *testing.T) {
		nCtx := createNodeContext("test")
		s := &dynamicNodeStateHolder{
			s: handler.DynamicNodeState{Phase: v1alpha1.DynamicNodePhaseParentFinalizing},
		}
		nCtx.OnNodeStateWriter().Return(s)
		f, err := nCtx.DataStore().ConstructReference(context.TODO(), nCtx.NodeStatus().GetDataDir(), "futures.pb")
		assert.NoError(t, err)
		dj := &core.DynamicJobSpec{}
		mockLPLauncher := &lpMocks.Reader{}
		n := &nodeMocks.Node{}
		assert.NoError(t, nCtx.DataStore().WriteProtobuf(context.TODO(), f, storage.Options{}, dj))
		h := &mocks.TaskNodeHandler{}
		h.OnFinalizeMatch(mock.Anything, mock.Anything).Return(fmt.Errorf("err"))
		d := New(h, n, mockLPLauncher, eventConfig, promutils.NewTestScope())
		_, err = d.Handle(context.TODO(), nCtx)
		assert.Error(t, err)
	})
}

func createDynamicJobSpec() *core.DynamicJobSpec {
	return &core.DynamicJobSpec{
		MinSuccesses: 2,
		Tasks: []*core.TaskTemplate{
			{
				Id:   &core.Identifier{Name: "task_1"},
				Type: "container",
				Interface: &core.TypedInterface{
					Outputs: &core.VariableMap{
						Variables: map[string]*core.Variable{
							"x": {
								Type: &core.LiteralType{
									Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
								},
							},
						},
					},
				},
			},
			{
				Id:   &core.Identifier{Name: "task_2"},
				Type: "container",
			},
		},
		Nodes: []*core.Node{
			{
				Id: "Node_1",
				Target: &core.Node_TaskNode{
					TaskNode: &core.TaskNode{
						Reference: &core.TaskNode_ReferenceId{
							ReferenceId: &core.Identifier{Name: "task_1"},
						},
					},
				},
			},
			{
				Id: "Node_2",
				Target: &core.Node_TaskNode{
					TaskNode: &core.TaskNode{
						Reference: &core.TaskNode_ReferenceId{
							ReferenceId: &core.Identifier{Name: "task_1"},
						},
					},
				},
			},
			{
				Id: "Node_3",
				Target: &core.Node_TaskNode{
					TaskNode: &core.TaskNode{
						Reference: &core.TaskNode_ReferenceId{
							ReferenceId: &core.Identifier{Name: "task_2"},
						},
					},
				},
			},
		},
		Outputs: []*core.Binding{
			{
				Var: "x",
				Binding: &core.BindingData{
					Value: &core.BindingData_Promise{
						Promise: &core.OutputReference{
							Var:    "x",
							NodeId: "Node_1",
						},
					},
				},
			},
		},
	}
}

func Test_dynamicNodeHandler_Handle_SubTaskV1(t *testing.T) {
	createNodeContext := func(ttype string, finalOutput storage.DataReference) *nodeMocks.NodeExecutionContext {
		ctx := context.TODO()
		nodeID := "n1"
		wfExecID := &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		}

		nm := &nodeMocks.NodeExecutionMetadata{}
		nm.OnGetAnnotations().Return(map[string]string{})
		nm.OnGetNodeExecutionID().Return(&core.NodeExecutionIdentifier{
			ExecutionId: wfExecID,
			NodeId:      nodeID,
		})
		nm.OnGetK8sServiceAccount().Return("service-account")
		nm.OnGetLabels().Return(map[string]string{})
		nm.OnGetNamespace().Return("namespace")
		nm.OnGetOwnerID().Return(types.NamespacedName{Namespace: "namespace", Name: "name"})
		nm.OnGetOwnerReference().Return(v1.OwnerReference{
			Kind: "sample",
			Name: "name",
		})

		taskID := &core.Identifier{}
		tk := &core.TaskTemplate{
			Id:   taskID,
			Type: "test",
			Metadata: &core.TaskMetadata{
				Discoverable: true,
			},
			Interface: &core.TypedInterface{
				Outputs: &core.VariableMap{
					Variables: map[string]*core.Variable{
						"x": {
							Type: &core.LiteralType{
								Type: &core.LiteralType_Simple{
									Simple: core.SimpleType_INTEGER,
								},
							},
						},
					},
				},
			},
		}
		tr := &nodeMocks.TaskReader{}
		tr.OnGetTaskID().Return(taskID)
		tr.OnGetTaskType().Return(ttype)
		tr.OnRead(ctx).Return(tk, nil)

		n := &flyteMocks.ExecutableNode{}
		n.OnGetTaskID().Return(&tID)

		dataStore, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
		assert.NoError(t, err)

		ir := &ioMocks.InputReader{}
		nCtx := &nodeMocks.NodeExecutionContext{}
		nCtx.OnNodeExecutionMetadata().Return(nm)
		nCtx.OnNode().Return(n)
		nCtx.OnInputReader().Return(ir)
		nCtx.OnCurrentAttempt().Return(uint32(1))
		nCtx.OnTaskReader().Return(tr)
		nCtx.OnNodeID().Return(nodeID)
		nCtx.OnEnqueueOwnerFunc().Return(func() error { return nil })
		nCtx.OnDataStore().Return(dataStore)

		endNodeStatus := &flyteMocks.ExecutableNodeStatus{}
		endNodeStatus.OnGetDataDir().Return("end-node")
		endNodeStatus.OnGetOutputDir().Return("end-node")

		subNs := &flyteMocks.ExecutableNodeStatus{}
		subNs.On("SetDataDir", mock.Anything).Return()
		subNs.On("SetOutputDir", mock.Anything).Return()
		subNs.On("SetParentNodeID", mock.Anything).Return()
		subNs.On("ResetDirty").Return()
		subNs.OnGetOutputDir().Return(finalOutput)
		subNs.On("SetParentTaskID", mock.Anything).Return()
		subNs.OnGetAttempts().Return(0)

		dynamicNS := &flyteMocks.ExecutableNodeStatus{}
		dynamicNS.On("SetDataDir", mock.Anything).Return()
		dynamicNS.On("SetOutputDir", mock.Anything).Return()
		dynamicNS.On("SetParentTaskID", mock.Anything).Return()
		dynamicNS.On("SetParentNodeID", mock.Anything).Return()
		dynamicNS.OnGetNodeExecutionStatus(ctx, "Node_1").Return(subNs)
		dynamicNS.OnGetNodeExecutionStatus(ctx, "Node_2").Return(subNs)
		dynamicNS.OnGetNodeExecutionStatus(ctx, "Node_3").Return(subNs)
		dynamicNS.OnGetNodeExecutionStatus(ctx, v1alpha1.EndNodeID).Return(endNodeStatus)

		ns := &flyteMocks.ExecutableNodeStatus{}
		ns.OnGetDataDir().Return("data-dir")
		ns.OnGetOutputDir().Return("output-dir")
		ns.OnGetNodeExecutionStatus(ctx, dynamicNodeID).Return(dynamicNS)
		nCtx.OnNodeStatus().Return(ns)

		w := &flyteMocks.ExecutableWorkflow{}
		ws := &flyteMocks.ExecutableWorkflowStatus{}
		ws.OnGetNodeExecutionStatus(ctx, nodeID).Return(ns)
		w.OnGetExecutionStatus().Return(ws)

		r := &nodeMocks.NodeStateReader{}
		r.OnGetDynamicNodeState().Return(handler.DynamicNodeState{
			Phase: v1alpha1.DynamicNodePhaseExecuting,
		})
		nCtx.OnNodeStateReader().Return(r)
		return nCtx
	}

	execInfoOutputOnly := &handler.ExecutionInfo{
		OutputInfo: &handler.OutputInfo{
			OutputURI: "output-dir/outputs.pb",
		},
	}

	type args struct {
		s               interfaces.NodeStatus
		isErr           bool
		dj              *core.DynamicJobSpec
		validErr        *io.ExecutionError
		generateOutputs bool
	}
	type want struct {
		p     handler.EPhase
		isErr bool
		phase v1alpha1.DynamicNodePhase
		info  *handler.ExecutionInfo
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{"error", args{isErr: true, dj: createDynamicJobSpec()}, want{isErr: true}},
		{"success", args{s: interfaces.NodeStatusSuccess, dj: createDynamicJobSpec()}, want{p: handler.EPhaseDynamicRunning, phase: v1alpha1.DynamicNodePhaseExecuting}},
		{"complete", args{s: interfaces.NodeStatusComplete, dj: createDynamicJobSpec(), generateOutputs: true}, want{p: handler.EPhaseSuccess, phase: v1alpha1.DynamicNodePhaseExecuting, info: execInfoOutputOnly}},
		{"complete-no-outputs", args{s: interfaces.NodeStatusComplete, dj: createDynamicJobSpec(), generateOutputs: false}, want{p: handler.EPhaseRetryableFailure, phase: v1alpha1.DynamicNodePhaseFailing}},
		{"complete-valid-error-retryable", args{s: interfaces.NodeStatusComplete, dj: createDynamicJobSpec(), validErr: &io.ExecutionError{IsRecoverable: true}, generateOutputs: true}, want{p: handler.EPhaseRetryableFailure, phase: v1alpha1.DynamicNodePhaseFailing, info: execInfoOutputOnly}},
		{"complete-valid-error", args{s: interfaces.NodeStatusComplete, dj: createDynamicJobSpec(), validErr: &io.ExecutionError{}, generateOutputs: true}, want{p: handler.EPhaseFailed, phase: v1alpha1.DynamicNodePhaseFailing, info: execInfoOutputOnly}},
		{"failed", args{s: interfaces.NodeStatusFailed(&core.ExecutionError{}), dj: createDynamicJobSpec()}, want{p: handler.EPhaseDynamicRunning, phase: v1alpha1.DynamicNodePhaseFailing}},
		{"running", args{s: interfaces.NodeStatusRunning, dj: createDynamicJobSpec()}, want{p: handler.EPhaseDynamicRunning, phase: v1alpha1.DynamicNodePhaseExecuting}},
		{"running-valid-err", args{s: interfaces.NodeStatusRunning, dj: createDynamicJobSpec(), validErr: &io.ExecutionError{}}, want{p: handler.EPhaseDynamicRunning, phase: v1alpha1.DynamicNodePhaseExecuting}},
		{"queued", args{s: interfaces.NodeStatusQueued, dj: createDynamicJobSpec()}, want{p: handler.EPhaseDynamicRunning, phase: v1alpha1.DynamicNodePhaseExecuting}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finalOutput := storage.DataReference("/subnode")
			nCtx := createNodeContext("test", finalOutput)
			s := &dynamicNodeStateHolder{}
			nCtx.OnNodeStateWriter().Return(s)
			f, err := nCtx.DataStore().ConstructReference(context.TODO(), nCtx.NodeStatus().GetOutputDir(), "futures.pb")
			assert.NoError(t, err)
			if tt.args.dj != nil {
				assert.NoError(t, nCtx.DataStore().WriteProtobuf(context.TODO(), f, storage.Options{}, tt.args.dj))
			}
			mockLPLauncher := &lpMocks.Reader{}
			h := &mocks.TaskNodeHandler{}
			if tt.args.validErr != nil {
				h.OnValidateOutputMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tt.args.validErr, nil)
			} else {
				h.OnValidateOutputMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
			}
			n := &nodeMocks.Node{}
			if tt.args.isErr {
				n.OnRecursiveNodeHandlerMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(interfaces.NodeStatusUndefined, fmt.Errorf("error"))
			} else {
				n.OnRecursiveNodeHandlerMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tt.args.s, nil)
			}
			if tt.args.generateOutputs {
				endF := v1alpha1.GetOutputsFile("end-node")
				assert.NoError(t, nCtx.DataStore().WriteProtobuf(context.TODO(), endF, storage.Options{}, &core.LiteralMap{}))
			}
			execContext := executorMocks.ExecutionContext{}
			execContext.EXPECT().GetEventVersion().Return(v1alpha1.EventVersion1)
			immutableParentInfo := executorMocks.ImmutableParentInfo{}
			immutableParentInfo.OnGetUniqueID().Return("c1")
			immutableParentInfo.OnCurrentAttempt().Return(uint32(2))
			execContext.EXPECT().GetParentInfo().Return(&immutableParentInfo)
			execContext.EXPECT().GetExecutionConfig().Return(v1alpha1.ExecutionConfig{})
			nCtx.OnExecutionContext().Return(&execContext)
			d := New(h, n, mockLPLauncher, eventConfig, promutils.NewTestScope())
			got, err := d.Handle(context.TODO(), nCtx)
			if tt.want.isErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if err == nil {
				assert.Equal(t, tt.want.p.String(), got.Info().GetPhase().String())
				assert.Equal(t, tt.want.phase, s.s.Phase)
				assert.Equal(t, tt.want.info, got.Info().GetInfo())
			}
		})
	}
}

func Test_dynamicNodeHandler_Handle_SubTask(t *testing.T) {
	createNodeContext := func(ttype string, finalOutput storage.DataReference) *nodeMocks.NodeExecutionContext {
		ctx := context.TODO()
		nodeID := "n1"
		wfExecID := &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		}

		nm := &nodeMocks.NodeExecutionMetadata{}
		nm.OnGetAnnotations().Return(map[string]string{})
		nm.OnGetNodeExecutionID().Return(&core.NodeExecutionIdentifier{
			ExecutionId: wfExecID,
			NodeId:      nodeID,
		})
		nm.OnGetK8sServiceAccount().Return("service-account")
		nm.OnGetLabels().Return(map[string]string{})
		nm.OnGetNamespace().Return("namespace")
		nm.OnGetOwnerID().Return(types.NamespacedName{Namespace: "namespace", Name: "name"})
		nm.OnGetOwnerReference().Return(v1.OwnerReference{
			Kind: "sample",
			Name: "name",
		})

		taskID := &core.Identifier{}
		tk := &core.TaskTemplate{
			Id:   taskID,
			Type: "test",
			Metadata: &core.TaskMetadata{
				Discoverable: true,
			},
			Interface: &core.TypedInterface{
				Outputs: &core.VariableMap{
					Variables: map[string]*core.Variable{
						"x": {
							Type: &core.LiteralType{
								Type: &core.LiteralType_Simple{
									Simple: core.SimpleType_INTEGER,
								},
							},
						},
					},
				},
			},
		}
		tr := &nodeMocks.TaskReader{}
		tr.OnGetTaskID().Return(taskID)
		tr.OnGetTaskType().Return(ttype)
		tr.OnRead(ctx).Return(tk, nil)

		n := &flyteMocks.ExecutableNode{}
		n.OnGetTaskID().Return(&tID)

		dataStore, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
		assert.NoError(t, err)

		ir := &ioMocks.InputReader{}
		nCtx := &nodeMocks.NodeExecutionContext{}
		nCtx.OnNodeExecutionMetadata().Return(nm)
		nCtx.OnNode().Return(n)
		nCtx.OnInputReader().Return(ir)
		nCtx.OnCurrentAttempt().Return(uint32(1))
		nCtx.OnTaskReader().Return(tr)
		nCtx.OnNodeID().Return(nodeID)
		nCtx.OnEnqueueOwnerFunc().Return(func() error { return nil })
		nCtx.OnDataStore().Return(dataStore)

		endNodeStatus := &flyteMocks.ExecutableNodeStatus{}
		endNodeStatus.OnGetDataDir().Return("end-node")
		endNodeStatus.OnGetOutputDir().Return("end-node")

		subNs := &flyteMocks.ExecutableNodeStatus{}
		subNs.On("SetDataDir", mock.Anything).Return()
		subNs.On("SetOutputDir", mock.Anything).Return()
		subNs.On("ResetDirty").Return()
		subNs.OnGetOutputDir().Return(finalOutput)
		subNs.On("SetParentTaskID", mock.Anything).Return()
		subNs.On("SetParentNodeID", mock.Anything).Return()
		subNs.OnGetAttempts().Return(0)

		dynamicNS := &flyteMocks.ExecutableNodeStatus{}
		dynamicNS.On("SetDataDir", mock.Anything).Return()
		dynamicNS.On("SetOutputDir", mock.Anything).Return()
		dynamicNS.On("SetParentTaskID", mock.Anything).Return()
		dynamicNS.On("SetParentNodeID", mock.Anything).Return()
		dynamicNS.OnGetNodeExecutionStatus(ctx, "n1-1-Node_1").Return(subNs)
		dynamicNS.OnGetNodeExecutionStatus(ctx, "n1-1-Node_2").Return(subNs)
		dynamicNS.OnGetNodeExecutionStatus(ctx, "n1-1-Node_3").Return(subNs)
		dynamicNS.OnGetNodeExecutionStatus(ctx, v1alpha1.EndNodeID).Return(endNodeStatus)

		ns := &flyteMocks.ExecutableNodeStatus{}
		ns.OnGetDataDir().Return("data-dir")
		ns.OnGetOutputDir().Return("output-dir")
		ns.OnGetNodeExecutionStatus(ctx, dynamicNodeID).Return(dynamicNS)
		nCtx.OnNodeStatus().Return(ns)

		w := &flyteMocks.ExecutableWorkflow{}
		ws := &flyteMocks.ExecutableWorkflowStatus{}
		ws.OnGetNodeExecutionStatus(ctx, nodeID).Return(ns)
		w.OnGetExecutionStatus().Return(ws)

		r := &nodeMocks.NodeStateReader{}
		r.OnGetDynamicNodeState().Return(handler.DynamicNodeState{
			Phase: v1alpha1.DynamicNodePhaseExecuting,
		})
		nCtx.OnNodeStateReader().Return(r)
		return nCtx
	}

	type args struct {
		s               interfaces.NodeStatus
		isErr           bool
		dj              *core.DynamicJobSpec
		validErr        *io.ExecutionError
		generateOutputs bool
	}
	type want struct {
		p     handler.EPhase
		isErr bool
		phase v1alpha1.DynamicNodePhase
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{"error", args{isErr: true, dj: createDynamicJobSpec()}, want{isErr: true}},
		{"success", args{s: interfaces.NodeStatusSuccess, dj: createDynamicJobSpec()}, want{p: handler.EPhaseDynamicRunning, phase: v1alpha1.DynamicNodePhaseExecuting}},
		{"complete", args{s: interfaces.NodeStatusComplete, dj: createDynamicJobSpec(), generateOutputs: true}, want{p: handler.EPhaseSuccess, phase: v1alpha1.DynamicNodePhaseExecuting}},
		{"complete-no-outputs", args{s: interfaces.NodeStatusComplete, dj: createDynamicJobSpec(), generateOutputs: false}, want{p: handler.EPhaseRetryableFailure, phase: v1alpha1.DynamicNodePhaseFailing}},
		{"complete-valid-error-retryable", args{s: interfaces.NodeStatusComplete, dj: createDynamicJobSpec(), validErr: &io.ExecutionError{IsRecoverable: true}, generateOutputs: true}, want{p: handler.EPhaseRetryableFailure, phase: v1alpha1.DynamicNodePhaseFailing}},
		{"complete-valid-error", args{s: interfaces.NodeStatusComplete, dj: createDynamicJobSpec(), validErr: &io.ExecutionError{}, generateOutputs: true}, want{p: handler.EPhaseFailed, phase: v1alpha1.DynamicNodePhaseFailing}},
		{"failed", args{s: interfaces.NodeStatusFailed(&core.ExecutionError{}), dj: createDynamicJobSpec()}, want{p: handler.EPhaseDynamicRunning, phase: v1alpha1.DynamicNodePhaseFailing}},
		{"running", args{s: interfaces.NodeStatusRunning, dj: createDynamicJobSpec()}, want{p: handler.EPhaseDynamicRunning, phase: v1alpha1.DynamicNodePhaseExecuting}},
		{"running-valid-err", args{s: interfaces.NodeStatusRunning, dj: createDynamicJobSpec(), validErr: &io.ExecutionError{}}, want{p: handler.EPhaseDynamicRunning, phase: v1alpha1.DynamicNodePhaseExecuting}},
		{"queued", args{s: interfaces.NodeStatusQueued, dj: createDynamicJobSpec()}, want{p: handler.EPhaseDynamicRunning, phase: v1alpha1.DynamicNodePhaseExecuting}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finalOutput := storage.DataReference("/subnode")
			nCtx := createNodeContext("test", finalOutput)
			s := &dynamicNodeStateHolder{}
			nCtx.OnNodeStateWriter().Return(s)
			f, err := nCtx.DataStore().ConstructReference(context.TODO(), nCtx.NodeStatus().GetOutputDir(), "futures.pb")
			assert.NoError(t, err)
			if tt.args.dj != nil {
				assert.NoError(t, nCtx.DataStore().WriteProtobuf(context.TODO(), f, storage.Options{}, tt.args.dj))
			}
			mockLPLauncher := &lpMocks.Reader{}
			h := &mocks.TaskNodeHandler{}
			if tt.args.validErr != nil {
				h.OnValidateOutputMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tt.args.validErr, nil)
			} else {
				h.OnValidateOutputMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
			}
			n := &nodeMocks.Node{}
			if tt.args.isErr {
				n.OnRecursiveNodeHandlerMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(interfaces.NodeStatusUndefined, fmt.Errorf("error"))
			} else {
				n.OnRecursiveNodeHandlerMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tt.args.s, nil)
			}
			if tt.args.generateOutputs {
				endF := v1alpha1.GetOutputsFile("end-node")
				assert.NoError(t, nCtx.DataStore().WriteProtobuf(context.TODO(), endF, storage.Options{}, &core.LiteralMap{}))
			}
			execContext := executorMocks.ExecutionContext{}
			execContext.EXPECT().GetEventVersion().Return(v1alpha1.EventVersion0)
			execContext.EXPECT().GetParentInfo().Return(nil)
			execContext.EXPECT().GetExecutionConfig().Return(v1alpha1.ExecutionConfig{})
			nCtx.OnExecutionContext().Return(&execContext)
			d := New(h, n, mockLPLauncher, eventConfig, promutils.NewTestScope())
			got, err := d.Handle(context.TODO(), nCtx)
			if tt.want.isErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if err == nil {
				assert.Equal(t, tt.want.p.String(), got.Info().GetPhase().String())
				assert.Equal(t, tt.want.phase, s.s.Phase)
			}
		})
	}
}

func createDynamicJobSpecWithLaunchPlans() *core.DynamicJobSpec {
	return &core.DynamicJobSpec{
		MinSuccesses: 1,
		Tasks:        []*core.TaskTemplate{},
		Nodes: []*core.Node{
			{
				Id: "Node_1",
				Target: &core.Node_WorkflowNode{
					WorkflowNode: &core.WorkflowNode{
						Reference: &core.WorkflowNode_LaunchplanRef{
							LaunchplanRef: &core.Identifier{
								ResourceType: core.ResourceType_LAUNCH_PLAN,
								Name:         "my_plan",
								Project:      "p",
								Domain:       "d",
							},
						},
					},
				},
			},
		},
		Outputs: []*core.Binding{
			{
				Var: "x",
				Binding: &core.BindingData{
					Value: &core.BindingData_Promise{
						Promise: &core.OutputReference{
							Var:    "x",
							NodeId: "Node_1",
						},
					},
				},
			},
		},
	}
}

func TestDynamicNodeTaskNodeHandler_Finalize(t *testing.T) {
	ctx := context.TODO()

	t.Run("dynamicnodephase-none", func(t *testing.T) {
		s := handler.DynamicNodeState{
			Phase:  v1alpha1.DynamicNodePhaseNone,
			Reason: "",
		}
		nCtx := &nodeMocks.NodeExecutionContext{}
		sr := &nodeMocks.NodeStateReader{}
		sr.OnGetDynamicNodeState().Return(s)
		nCtx.OnNodeStateReader().Return(sr)
		nCtx.OnCurrentAttempt().Return(0)

		mockLPLauncher := &lpMocks.Reader{}
		h := &mocks.TaskNodeHandler{}
		h.OnFinalize(ctx, nCtx).Return(nil)
		n := &nodeMocks.Node{}
		d := New(h, n, mockLPLauncher, eventConfig, promutils.NewTestScope())
		assert.NoError(t, d.Finalize(ctx, nCtx))
		assert.NotZero(t, len(h.ExpectedCalls))
		assert.Equal(t, "Finalize", h.ExpectedCalls[0].Method)
	})

	createNodeContext := func(ttype string, finalOutput storage.DataReference) *nodeMocks.NodeExecutionContext {
		ctx := context.TODO()
		nodeID := "n1"
		wfExecID := &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		}

		nm := &nodeMocks.NodeExecutionMetadata{}
		nm.OnGetAnnotations().Return(map[string]string{})
		nm.OnGetNodeExecutionID().Return(&core.NodeExecutionIdentifier{
			ExecutionId: wfExecID,
			NodeId:      nodeID,
		})
		nm.OnGetK8sServiceAccount().Return("service-account")
		nm.OnGetLabels().Return(map[string]string{})
		nm.OnGetNamespace().Return("namespace")
		nm.OnGetOwnerID().Return(types.NamespacedName{Namespace: "namespace", Name: "name"})
		nm.OnGetOwnerReference().Return(v1.OwnerReference{
			Kind: "sample",
			Name: "name",
		})

		taskID := &core.Identifier{}
		tk := &core.TaskTemplate{
			Id:   taskID,
			Type: "test",
			Metadata: &core.TaskMetadata{
				Discoverable: true,
			},
			Interface: &core.TypedInterface{
				Outputs: &core.VariableMap{
					Variables: map[string]*core.Variable{
						"x": {
							Type: &core.LiteralType{
								Type: &core.LiteralType_Simple{
									Simple: core.SimpleType_INTEGER,
								},
							},
						},
					},
				},
			},
		}
		tr := &nodeMocks.TaskReader{}
		tr.OnGetTaskID().Return(taskID)
		tr.OnGetTaskType().Return(ttype)
		tr.OnRead(ctx).Return(tk, nil)

		n := &flyteMocks.ExecutableNode{}
		n.OnGetTaskID().Return(&tID)

		dataStore, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
		assert.NoError(t, err)

		ir := &ioMocks.InputReader{}
		nCtx := &nodeMocks.NodeExecutionContext{}
		nCtx.OnNodeExecutionMetadata().Return(nm)
		nCtx.OnNode().Return(n)
		nCtx.OnInputReader().Return(ir)
		nCtx.OnCurrentAttempt().Return(uint32(1))
		nCtx.OnTaskReader().Return(tr)
		nCtx.OnNodeID().Return(nodeID)
		nCtx.OnEnqueueOwnerFunc().Return(func() error { return nil })
		nCtx.OnDataStore().Return(dataStore)
		execContext := executorMocks.ExecutionContext{}
		execContext.EXPECT().GetEventVersion().Return(v1alpha1.EventVersion0)
		execContext.EXPECT().GetParentInfo().Return(nil)
		nCtx.OnExecutionContext().Return(&execContext)

		endNodeStatus := &flyteMocks.ExecutableNodeStatus{}
		endNodeStatus.OnGetDataDir().Return("end-node")
		endNodeStatus.OnGetOutputDir().Return("end-node")

		subNs := &flyteMocks.ExecutableNodeStatus{}
		subNs.On("SetDataDir", mock.Anything).Return()
		subNs.On("SetOutputDir", mock.Anything).Return()
		subNs.On("ResetDirty").Return()
		subNs.OnGetOutputDir().Return(finalOutput)
		subNs.On("SetParentTaskID", mock.Anything).Return()
		subNs.On("SetParentNodeID", mock.Anything).Return()
		subNs.OnGetAttempts().Return(0)

		dynamicNS := &flyteMocks.ExecutableNodeStatus{}
		dynamicNS.On("SetDataDir", mock.Anything).Return()
		dynamicNS.On("SetOutputDir", mock.Anything).Return()
		dynamicNS.On("SetParentTaskID", mock.Anything).Return()
		dynamicNS.On("SetParentNodeID", mock.Anything).Return()
		dynamicNS.OnGetNodeExecutionStatus(ctx, "n1-1-Node_1").Return(subNs)
		dynamicNS.OnGetNodeExecutionStatus(ctx, "n1-1-Node_2").Return(subNs)
		dynamicNS.OnGetNodeExecutionStatus(ctx, "n1-1-Node_3").Return(subNs)
		dynamicNS.OnGetNodeExecutionStatus(ctx, v1alpha1.EndNodeID).Return(endNodeStatus)

		ns := &flyteMocks.ExecutableNodeStatus{}
		ns.OnGetDataDir().Return("data-dir")
		ns.OnGetOutputDir().Return("output-dir")
		ns.OnGetNodeExecutionStatus(ctx, dynamicNodeID).Return(dynamicNS)
		nCtx.OnNodeStatus().Return(ns)

		w := &flyteMocks.ExecutableWorkflow{}
		ws := &flyteMocks.ExecutableWorkflowStatus{}
		ws.OnGetNodeExecutionStatus(ctx, nodeID).Return(ns)
		w.OnGetExecutionStatus().Return(ws)

		r := &nodeMocks.NodeStateReader{}
		r.OnGetDynamicNodeState().Return(handler.DynamicNodeState{
			Phase: v1alpha1.DynamicNodePhaseExecuting,
		})
		nCtx.OnNodeStateReader().Return(r)
		return nCtx
	}
	t.Run("dynamicnodephase-executing", func(t *testing.T) {

		nCtx := createNodeContext("test", "x")
		f, err := nCtx.DataStore().ConstructReference(context.TODO(), nCtx.NodeStatus().GetOutputDir(), "futures.pb")
		assert.NoError(t, err)
		dj := createDynamicJobSpec()
		assert.NoError(t, nCtx.DataStore().WriteProtobuf(context.TODO(), f, storage.Options{}, dj))

		mockLPLauncher := &lpMocks.Reader{}
		h := &mocks.TaskNodeHandler{}
		h.OnFinalize(ctx, nCtx).Return(nil)
		n := &nodeMocks.Node{}
		n.OnFinalizeHandlerMatch(ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		d := New(h, n, mockLPLauncher, eventConfig, promutils.NewTestScope())
		assert.NoError(t, d.Finalize(ctx, nCtx))
		assert.NotZero(t, len(h.ExpectedCalls))
		assert.Equal(t, "Finalize", h.ExpectedCalls[0].Method)
		assert.NotZero(t, len(n.ExpectedCalls))
		assert.Equal(t, "FinalizeHandler", n.ExpectedCalls[0].Method)
	})

	t.Run("dynamicnodephase-executing-parenterror", func(t *testing.T) {

		nCtx := createNodeContext("test", "x")
		f, err := nCtx.DataStore().ConstructReference(context.TODO(), nCtx.NodeStatus().GetOutputDir(), "futures.pb")
		assert.NoError(t, err)
		dj := createDynamicJobSpec()
		assert.NoError(t, nCtx.DataStore().WriteProtobuf(context.TODO(), f, storage.Options{}, dj))

		mockLPLauncher := &lpMocks.Reader{}
		h := &mocks.TaskNodeHandler{}
		h.OnFinalize(ctx, nCtx).Return(fmt.Errorf("err"))
		n := &nodeMocks.Node{}
		n.OnFinalizeHandlerMatch(ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
		d := New(h, n, mockLPLauncher, eventConfig, promutils.NewTestScope())
		assert.Error(t, d.Finalize(ctx, nCtx))
		assert.NotZero(t, len(h.ExpectedCalls))
		assert.Equal(t, "Finalize", h.ExpectedCalls[0].Method)
		assert.NotZero(t, len(n.ExpectedCalls))
		assert.Equal(t, "FinalizeHandler", n.ExpectedCalls[0].Method)
	})

	t.Run("dynamicnodephase-executing-childerror", func(t *testing.T) {

		nCtx := createNodeContext("test", "x")
		f, err := nCtx.DataStore().ConstructReference(context.TODO(), nCtx.NodeStatus().GetOutputDir(), "futures.pb")
		assert.NoError(t, err)
		dj := createDynamicJobSpec()
		assert.NoError(t, nCtx.DataStore().WriteProtobuf(context.TODO(), f, storage.Options{}, dj))

		mockLPLauncher := &lpMocks.Reader{}
		h := &mocks.TaskNodeHandler{}
		h.OnFinalize(ctx, nCtx).Return(nil)
		n := &nodeMocks.Node{}
		n.OnFinalizeHandlerMatch(ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("err"))
		d := New(h, n, mockLPLauncher, eventConfig, promutils.NewTestScope())
		assert.Error(t, d.Finalize(ctx, nCtx))
		assert.NotZero(t, len(h.ExpectedCalls))
		assert.Equal(t, "Finalize", h.ExpectedCalls[0].Method)
		assert.NotZero(t, len(n.ExpectedCalls))
		assert.Equal(t, "FinalizeHandler", n.ExpectedCalls[0].Method)
	})
}

func init() {
	labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey,
		contextutils.TaskIDKey)
}
