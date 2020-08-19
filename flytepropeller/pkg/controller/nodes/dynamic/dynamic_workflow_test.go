package dynamic

import (
	"context"
	"testing"

	"github.com/lyft/flytepropeller/pkg/utils"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	mocks3 "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	mocks2 "github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
	mocks4 "github.com/lyft/flytepropeller/pkg/controller/executors/mocks"
	mocks6 "github.com/lyft/flytepropeller/pkg/controller/nodes/dynamic/mocks"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler/mocks"
	mocks5 "github.com/lyft/flytepropeller/pkg/controller/nodes/subworkflow/launchplan/mocks"
)

func Test_dynamicNodeHandler_buildContextualDynamicWorkflow_withLaunchPlans(t *testing.T) {
	createNodeContext := func(ttype string, finalOutput storage.DataReference) *mocks.NodeExecutionContext {
		ctx := context.TODO()

		wfExecID := &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		}

		nodeID := "n1"
		nm := &mocks.NodeExecutionMetadata{}
		nm.OnGetAnnotations().Return(map[string]string{})
		nm.OnGetNodeExecutionID().Return(&core.NodeExecutionIdentifier{ExecutionId: wfExecID, NodeId: nodeID})
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
		tr := &mocks.TaskReader{}
		tr.OnGetTaskID().Return(taskID)
		tr.OnGetTaskType().Return(ttype)
		tr.OnReadMatch(mock.Anything).Return(tk, nil)

		n := &mocks2.ExecutableNode{}
		tID := "dyn-task-1"
		n.OnGetTaskID().Return(&tID)

		dataStore, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
		assert.NoError(t, err)

		ir := &mocks3.InputReader{}
		nCtx := &mocks.NodeExecutionContext{}
		nCtx.OnNodeExecutionMetadata().Return(nm)
		nCtx.OnNode().Return(n)
		nCtx.OnInputReader().Return(ir)
		nCtx.OnCurrentAttempt().Return(uint32(1))
		nCtx.OnTaskReader().Return(tr)
		nCtx.OnMaxDatasetSizeBytes().Return(int64(1))
		nCtx.OnNodeID().Return("n1")
		nCtx.OnEnqueueOwnerFunc().Return(func() error { return nil })
		nCtx.OnDataStore().Return(dataStore)

		endNodeStatus := &mocks2.ExecutableNodeStatus{}
		endNodeStatus.OnGetDataDir().Return(storage.DataReference("end-node"))
		endNodeStatus.OnGetOutputDir().Return(storage.DataReference("end-node"))

		subNs := &mocks2.ExecutableNodeStatus{}
		subNs.On("SetDataDir", mock.Anything).Return()
		subNs.On("SetOutputDir", mock.Anything).Return()
		subNs.On("ResetDirty").Return()
		subNs.OnGetOutputDir().Return(finalOutput)
		subNs.On("SetParentTaskID", mock.Anything).Return()
		subNs.OnGetAttempts().Return(0)

		dynamicNS := &mocks2.ExecutableNodeStatus{}
		dynamicNS.On("SetDataDir", mock.Anything).Return()
		dynamicNS.On("SetOutputDir", mock.Anything).Return()
		dynamicNS.On("SetParentTaskID", mock.Anything).Return()
		dynamicNS.OnGetNodeExecutionStatus(ctx, "n1-1-Node_1").Return(subNs)
		dynamicNS.OnGetNodeExecutionStatus(ctx, "Node_1").Return(subNs)
		dynamicNS.OnGetNodeExecutionStatus(ctx, v1alpha1.EndNodeID).Return(endNodeStatus)

		ns := &mocks2.ExecutableNodeStatus{}
		ns.OnGetDataDir().Return(storage.DataReference("data-dir"))
		ns.OnGetOutputDir().Return(storage.DataReference("output-dir"))
		ns.OnGetNodeExecutionStatus(ctx, dynamicNodeID).Return(dynamicNS)
		ns.OnGetNodeExecutionStatus(ctx, dynamicNodeID).Return(dynamicNS)
		nCtx.OnNodeStatus().Return(ns)

		w := &mocks2.ExecutableWorkflow{}
		ws := &mocks2.ExecutableWorkflowStatus{}
		ws.OnGetNodeExecutionStatus(ctx, "n1").Return(ns)
		w.OnGetExecutionStatus().Return(ws)

		r := &mocks.NodeStateReader{}
		r.OnGetDynamicNodeState().Return(handler.DynamicNodeState{
			Phase: v1alpha1.DynamicNodePhaseExecuting,
		})
		nCtx.OnNodeStateReader().Return(r)
		return nCtx
	}

	t.Run("launch plan interfaces match parent task interface", func(t *testing.T) {
		ctx := context.Background()
		lpID := &core.Identifier{
			ResourceType: core.ResourceType_LAUNCH_PLAN,
			Name:         "my_plan",
			Project:      "p",
			Domain:       "d",
		}
		djSpec := createDynamicJobSpecWithLaunchPlans()
		finalOutput := storage.DataReference("/subnode")
		nCtx := createNodeContext("test", finalOutput)
		s := &dynamicNodeStateHolder{}
		nCtx.On("NodeStateWriter").Return(s)
		f, err := nCtx.DataStore().ConstructReference(ctx, nCtx.NodeStatus().GetOutputDir(), "futures.pb")
		assert.NoError(t, err)
		assert.NoError(t, nCtx.DataStore().WriteProtobuf(context.TODO(), f, storage.Options{}, djSpec))

		mockLPLauncher := &mocks5.Reader{}
		var callsAdmin = false
		mockLPLauncher.OnGetLaunchPlanMatch(ctx, lpID).Run(func(args mock.Arguments) {
			// When a launch plan node is detected, a call should be made to Admin to fetch the interface for the LP
			callsAdmin = true
		}).Return(&admin.LaunchPlan{
			Id: lpID,
			Closure: &admin.LaunchPlanClosure{
				ExpectedInputs: &core.ParameterMap{},
				ExpectedOutputs: &core.VariableMap{
					Variables: map[string]*core.Variable{
						"x": {
							Type: &core.LiteralType{
								Type: &core.LiteralType_Simple{
									Simple: core.SimpleType_INTEGER,
								},
							},
							Description: "output of the launch plan",
						},
					},
				},
			},
		}, nil)
		h := &mocks6.TaskNodeHandler{}
		n := &mocks4.Node{}
		d := dynamicNodeTaskNodeHandler{
			TaskNodeHandler: h,
			nodeExecutor:    n,
			lpReader:        mockLPLauncher,
			metrics:         newMetrics(promutils.NewTestScope()),
		}

		execContext := &mocks4.ExecutionContext{}
		immutableParentInfo := mocks4.ImmutableParentInfo{}
		immutableParentInfo.OnGetUniqueID().Return("c1")
		immutableParentInfo.OnCurrentAttempt().Return(uint32(2))
		execContext.OnGetParentInfo().Return(&immutableParentInfo)
		execContext.OnGetEventVersion().Return(v1alpha1.EventVersion1)
		nCtx.OnExecutionContext().Return(execContext)

		dCtx, err := d.buildContextualDynamicWorkflow(ctx, nCtx)
		assert.NoError(t, err)
		assert.True(t, callsAdmin)
		assert.True(t, dCtx.isDynamic)
		assert.NotNil(t, dCtx.subWorkflow)
		assert.NotNil(t, dCtx.execContext)
		assert.NotNil(t, dCtx.execContext.GetParentInfo())
		expectedParentUniqueID, err := utils.FixedLengthUniqueIDForParts(20, "c1", "2", "n1")
		assert.Nil(t, err)
		assert.Equal(t, expectedParentUniqueID, dCtx.execContext.GetParentInfo().GetUniqueID())
		assert.Equal(t, uint32(1), dCtx.execContext.GetParentInfo().CurrentAttempt())
		assert.NotNil(t, dCtx.nodeLookup)
	})

	t.Run("launch plan interfaces match parent task interface no parent", func(t *testing.T) {
		ctx := context.Background()
		lpID := &core.Identifier{
			ResourceType: core.ResourceType_LAUNCH_PLAN,
			Name:         "my_plan",
			Project:      "p",
			Domain:       "d",
		}
		djSpec := createDynamicJobSpecWithLaunchPlans()
		finalOutput := storage.DataReference("/subnode")
		nCtx := createNodeContext("test", finalOutput)
		s := &dynamicNodeStateHolder{}
		nCtx.On("NodeStateWriter").Return(s)
		f, err := nCtx.DataStore().ConstructReference(ctx, nCtx.NodeStatus().GetOutputDir(), "futures.pb")
		assert.NoError(t, err)
		assert.NoError(t, nCtx.DataStore().WriteProtobuf(context.TODO(), f, storage.Options{}, djSpec))

		mockLPLauncher := &mocks5.Reader{}
		var callsAdmin = false
		mockLPLauncher.OnGetLaunchPlanMatch(ctx, lpID).Run(func(args mock.Arguments) {
			// When a launch plan node is detected, a call should be made to Admin to fetch the interface for the LP
			callsAdmin = true
		}).Return(&admin.LaunchPlan{
			Id: lpID,
			Closure: &admin.LaunchPlanClosure{
				ExpectedInputs: &core.ParameterMap{},
				ExpectedOutputs: &core.VariableMap{
					Variables: map[string]*core.Variable{
						"x": {
							Type: &core.LiteralType{
								Type: &core.LiteralType_Simple{
									Simple: core.SimpleType_INTEGER,
								},
							},
							Description: "output of the launch plan",
						},
					},
				},
			},
		}, nil)
		h := &mocks6.TaskNodeHandler{}
		n := &mocks4.Node{}
		d := dynamicNodeTaskNodeHandler{
			TaskNodeHandler: h,
			nodeExecutor:    n,
			lpReader:        mockLPLauncher,
			metrics:         newMetrics(promutils.NewTestScope()),
		}

		execContext := &mocks4.ExecutionContext{}
		execContext.OnGetParentInfo().Return(nil)
		execContext.OnGetEventVersion().Return(v1alpha1.EventVersion0)
		nCtx.OnExecutionContext().Return(execContext)

		dCtx, err := d.buildContextualDynamicWorkflow(ctx, nCtx)
		assert.NoError(t, err)
		assert.True(t, callsAdmin)
		assert.True(t, dCtx.isDynamic)
		assert.NotNil(t, dCtx.subWorkflow)
		assert.NotNil(t, dCtx.execContext)
		assert.NotNil(t, dCtx.execContext.GetParentInfo())
		expectedParentUniqueID, err := utils.FixedLengthUniqueIDForParts(20, "", "", "n1")
		assert.Nil(t, err)
		assert.Equal(t, expectedParentUniqueID, dCtx.execContext.GetParentInfo().GetUniqueID())
		assert.Equal(t, uint32(1), dCtx.execContext.GetParentInfo().CurrentAttempt())
		assert.NotNil(t, dCtx.nodeLookup)
	})

	t.Run("launch plan interfaces do not parent task interface", func(t *testing.T) {
		ctx := context.Background()
		lpID := &core.Identifier{
			ResourceType: core.ResourceType_LAUNCH_PLAN,
			Name:         "my_plan",
			Project:      "p",
			Domain:       "d",
		}
		djSpec := createDynamicJobSpecWithLaunchPlans()
		finalOutput := storage.DataReference("/subnode")
		nCtx := createNodeContext("test", finalOutput)
		s := &dynamicNodeStateHolder{}
		nCtx.OnNodeStateWriter().Return(s)
		f, err := nCtx.DataStore().ConstructReference(ctx, nCtx.NodeStatus().GetOutputDir(), "futures.pb")
		assert.NoError(t, err)
		assert.NoError(t, nCtx.DataStore().WriteProtobuf(context.TODO(), f, storage.Options{}, djSpec))

		mockLPLauncher := &mocks5.Reader{}
		var callsAdmin = false
		mockLPLauncher.OnGetLaunchPlanMatch(ctx, lpID).Run(func(args mock.Arguments) {
			// When a launch plan node is detected, a call should be made to Admin to fetch the interface for the LP
			callsAdmin = true
		}).Return(&admin.LaunchPlan{
			Id: lpID,
			Closure: &admin.LaunchPlanClosure{
				ExpectedInputs: &core.ParameterMap{},
				ExpectedOutputs: &core.VariableMap{
					Variables: map[string]*core.Variable{
						"d": {
							Type: &core.LiteralType{
								Type: &core.LiteralType_Simple{
									Simple: core.SimpleType_STRING,
								},
							},
							Description: "output of the launch plan",
						},
					},
				},
			},
		}, nil)
		h := &mocks6.TaskNodeHandler{}
		n := &mocks4.Node{}
		d := dynamicNodeTaskNodeHandler{
			TaskNodeHandler: h,
			nodeExecutor:    n,
			lpReader:        mockLPLauncher,
			metrics:         newMetrics(promutils.NewTestScope()),
		}
		execContext := &mocks4.ExecutionContext{}
		execContext.OnGetParentInfo().Return(nil)
		execContext.OnGetEventVersion().Return(v1alpha1.EventVersion0)
		nCtx.OnExecutionContext().Return(execContext)

		_, err = d.buildContextualDynamicWorkflow(ctx, nCtx)
		assert.Error(t, err)
		assert.True(t, callsAdmin)
	})
}
