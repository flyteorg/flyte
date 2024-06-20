package dynamic

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/encoding"
	mocks3 "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	mocks2 "github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
	mocks4 "github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors/mocks"
	mocks6 "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/dynamic/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/handler"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces/mocks"
	mocks5 "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/subworkflow/launchplan/mocks"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/storage"
	storageMocks "github.com/flyteorg/flyte/flytestdlib/storage/mocks"
)

func Test_dynamicNodeHandler_buildContextualDynamicWorkflow_withLaunchPlans(t *testing.T) {
	createNodeContext := func(ttype string, finalOutput storage.DataReference, dataStore *storage.DataStore) *mocks.NodeExecutionContext {
		ctx := context.Background()

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

		if dataStore == nil {
			var err error
			dataStore, err = storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
			assert.NoError(t, err)
		}

		ir := &mocks3.InputReader{}
		nCtx := &mocks.NodeExecutionContext{}
		nCtx.OnNodeExecutionMetadata().Return(nm)
		nCtx.OnNode().Return(n)
		nCtx.OnInputReader().Return(ir)
		nCtx.OnCurrentAttempt().Return(uint32(1))
		nCtx.OnTaskReader().Return(tr)
		nCtx.OnNodeID().Return("n1")
		nCtx.OnEnqueueOwnerFunc().Return(func() error { return nil })
		nCtx.OnDataStore().Return(dataStore)

		endNodeStatus := &mocks2.ExecutableNodeStatus{}
		endNodeStatus.OnGetDataDir().Return("end-node")
		endNodeStatus.OnGetOutputDir().Return("end-node")

		subNs := &mocks2.ExecutableNodeStatus{}
		subNs.On("SetDataDir", mock.Anything).Return()
		subNs.On("SetOutputDir", mock.Anything).Return()
		subNs.On("ResetDirty").Return()
		subNs.OnGetOutputDir().Return(finalOutput)
		subNs.On("SetParentTaskID", mock.Anything).Return()
		subNs.On("SetParentNodeID", mock.Anything).Return()
		subNs.OnGetAttempts().Return(0)

		dynamicNS := &mocks2.ExecutableNodeStatus{}
		dynamicNS.On("SetDataDir", mock.Anything).Return()
		dynamicNS.On("SetOutputDir", mock.Anything).Return()
		dynamicNS.On("SetParentTaskID", mock.Anything).Return()
		dynamicNS.On("SetParentNodeID", mock.Anything).Return()
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
		nCtx := createNodeContext("test", finalOutput, nil)
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
		n := &mocks.Node{}
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
		execContext.EXPECT().GetParentInfo().Return(&immutableParentInfo)
		execContext.EXPECT().GetEventVersion().Return(v1alpha1.EventVersion1)
		nCtx.OnExecutionContext().Return(execContext)

		dCtx, err := d.buildContextualDynamicWorkflow(ctx, nCtx)
		assert.NoError(t, err)
		assert.True(t, callsAdmin)
		assert.True(t, dCtx.isDynamic)
		assert.NotNil(t, dCtx.subWorkflow)
		assert.NotNil(t, dCtx.subWorkflowClosure)
		assert.NotNil(t, dCtx.execContext)
		assert.NotNil(t, dCtx.execContext.GetParentInfo())
		expectedParentUniqueID, err := encoding.FixedLengthUniqueIDForParts(20, []string{"c1", "2", "n1"})
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
		nCtx := createNodeContext("test", finalOutput, nil)
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
		n := &mocks.Node{}
		d := dynamicNodeTaskNodeHandler{
			TaskNodeHandler: h,
			nodeExecutor:    n,
			lpReader:        mockLPLauncher,
			metrics:         newMetrics(promutils.NewTestScope()),
		}

		execContext := &mocks4.ExecutionContext{}
		execContext.EXPECT().GetParentInfo().Return(nil)
		execContext.EXPECT().GetEventVersion().Return(v1alpha1.EventVersion0)
		nCtx.OnExecutionContext().Return(execContext)

		dCtx, err := d.buildContextualDynamicWorkflow(ctx, nCtx)
		assert.NoError(t, err)
		assert.True(t, callsAdmin)
		assert.True(t, dCtx.isDynamic)
		assert.NotNil(t, dCtx.subWorkflow)
		assert.NotNil(t, dCtx.subWorkflowClosure)
		assert.NotNil(t, dCtx.execContext)
		assert.NotNil(t, dCtx.execContext.GetParentInfo())
		expectedParentUniqueID, err := encoding.FixedLengthUniqueIDForParts(20, []string{"", "", "n1"})
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
		nCtx := createNodeContext("test", finalOutput, nil)
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
		n := &mocks.Node{}
		d := dynamicNodeTaskNodeHandler{
			TaskNodeHandler: h,
			nodeExecutor:    n,
			lpReader:        mockLPLauncher,
			metrics:         newMetrics(promutils.NewTestScope()),
		}
		execContext := &mocks4.ExecutionContext{}
		execContext.EXPECT().GetParentInfo().Return(nil)
		execContext.EXPECT().GetEventVersion().Return(v1alpha1.EventVersion0)
		nCtx.OnExecutionContext().Return(execContext)

		_, err = d.buildContextualDynamicWorkflow(ctx, nCtx)
		assert.Error(t, err)
		assert.True(t, callsAdmin)
	})
	t.Run("dynamic wf cached", func(t *testing.T) {
		ctx := context.Background()
		lpID := &core.Identifier{
			ResourceType: core.ResourceType_LAUNCH_PLAN,
			Name:         "my_plan",
			Project:      "p",
			Domain:       "d",
		}
		djSpec := createDynamicJobSpecWithLaunchPlans()
		finalOutput := storage.DataReference("/subnode")
		nCtx := createNodeContext("test", finalOutput, nil)

		s := &dynamicNodeStateHolder{}
		nCtx.On("NodeStateWriter").Return(s)

		// Create a k8s Flyte workflow and store that in the cache
		dynamicWf := &v1alpha1.FlyteWorkflow{
			ServiceAccountName: "sa",
		}

		rawDynamicWf, err := json.Marshal(dynamicWf)
		assert.NoError(t, err)
		_, err = nCtx.DataStore().ConstructReference(ctx, nCtx.NodeStatus().GetOutputDir(), "futures_compiled.pb")
		assert.NoError(t, err)
		assert.NoError(t, nCtx.DataStore().WriteRaw(context.TODO(), "/output-dir/futures_compiled.pb", int64(len(rawDynamicWf)), storage.Options{}, bytes.NewReader(rawDynamicWf)))

		f, err := nCtx.DataStore().ConstructReference(ctx, nCtx.NodeStatus().GetOutputDir(), "futures.pb")
		assert.NoError(t, err)
		assert.NoError(t, nCtx.DataStore().WriteProtobuf(context.TODO(), f, storage.Options{}, djSpec))

		f, err = nCtx.DataStore().ConstructReference(ctx, nCtx.NodeStatus().GetOutputDir(), "dynamic_compiled.pb")
		assert.NoError(t, err)
		assert.NoError(t, nCtx.DataStore().WriteProtobuf(context.TODO(), f, storage.Options{}, &core.CompiledWorkflowClosure{
			Primary: &core.CompiledWorkflow{
				Template: &core.WorkflowTemplate{
					Id: &core.Identifier{
						ResourceType: core.ResourceType_WORKFLOW,
					},
				},
			},
		}))

		mockLPLauncher := &mocks5.Reader{}
		var callsAdmin = false
		mockLPLauncher.OnGetLaunchPlanMatch(ctx, lpID).Run(func(args mock.Arguments) {
			// When a launch plan node is detected, a call should be made to Admin to fetch the interface for the LP.
			// However in the cached case no such call should be necessary.
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
		n := &mocks.Node{}
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
		execContext.EXPECT().GetParentInfo().Return(&immutableParentInfo)
		execContext.EXPECT().GetEventVersion().Return(v1alpha1.EventVersion1)
		nCtx.OnExecutionContext().Return(execContext)

		dCtx, err := d.buildContextualDynamicWorkflow(ctx, nCtx)
		assert.NoError(t, err)
		assert.False(t, callsAdmin)
		assert.True(t, dCtx.isDynamic)
		assert.NotNil(t, dCtx.subWorkflow)
		assert.NotNil(t, dCtx.execContext)
		assert.NotNil(t, dCtx.execContext.GetParentInfo())
		expectedParentUniqueID, err := encoding.FixedLengthUniqueIDForParts(20, []string{"c1", "2", "n1"})
		assert.Nil(t, err)
		assert.Equal(t, expectedParentUniqueID, dCtx.execContext.GetParentInfo().GetUniqueID())
		assert.Equal(t, uint32(1), dCtx.execContext.GetParentInfo().CurrentAttempt())
		assert.NotNil(t, dCtx.nodeLookup)
	})

	t.Run("dynamic wf cache read fails", func(t *testing.T) {
		ctx := context.Background()
		finalOutput := storage.DataReference("/subnode")

		composedPBStore := storageMocks.ComposedProtobufStore{}
		composedPBStore.On("Head", mock.MatchedBy(func(ctx context.Context) bool { return true }), storage.DataReference("s3://my-s3-bucket/foo/bar/futures_compiled.pb")).
			Return(nil, errors.New("foo"))
		referenceConstructor := storageMocks.ReferenceConstructor{}
		referenceConstructor.On("ConstructReference", mock.MatchedBy(func(ctx context.Context) bool { return true }), storage.DataReference("output-dir"), "futures.pb").Return(
			storage.DataReference("s3://my-s3-bucket/foo/bar/futures.pb"), nil)
		referenceConstructor.On("ConstructReference", mock.MatchedBy(func(ctx context.Context) bool { return true }), storage.DataReference("output-dir"), "futures_compiled.pb").Return(
			storage.DataReference("s3://my-s3-bucket/foo/bar/futures_compiled.pb"), nil)
		referenceConstructor.On("ConstructReference", mock.MatchedBy(func(ctx context.Context) bool { return true }), storage.DataReference("output-dir"), "dynamic_compiled.pb").Return(
			storage.DataReference("s3://my-s3-bucket/foo/bar/dynamic_compiled.pb"), nil)
		dataStore := &storage.DataStore{
			ComposedProtobufStore: &composedPBStore,
			ReferenceConstructor:  &referenceConstructor,
		}

		nCtx := createNodeContext("test", finalOutput, dataStore)
		nCtx.OnCurrentAttempt().Return(uint32(1))
		mockLPLauncher := &mocks5.Reader{}

		h := &mocks6.TaskNodeHandler{}
		n := &mocks.Node{}
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
		execContext.EXPECT().GetParentInfo().Return(&immutableParentInfo)
		execContext.EXPECT().GetEventVersion().Return(v1alpha1.EventVersion1)
		nCtx.OnExecutionContext().Return(execContext)

		_, err := d.buildContextualDynamicWorkflow(ctx, nCtx)
		assert.EqualError(t, err, "[system] Failed to do HEAD on compiled workflow files., caused by: Failed to do HEAD on futures file.: foo")
	})
	t.Run("dynamic wf cache write fails", func(t *testing.T) {
		ctx := context.Background()
		finalOutput := storage.DataReference("/subnode")

		metadata := existsMetadata{}
		composedPBStore := storageMocks.ComposedProtobufStore{}
		composedPBStore.OnHeadMatch(mock.MatchedBy(func(ctx context.Context) bool { return true }), storage.DataReference("s3://my-s3-bucket/foo/bar/futures_compiled.pb")).
			Return(&metadata, nil)

		djSpec := createDynamicJobSpecWithLaunchPlans()
		composedPBStore.OnReadProtobufMatch(mock.MatchedBy(func(ctx context.Context) bool { return true }),
			storage.DataReference("s3://my-s3-bucket/foo/bar/futures.pb"), &core.DynamicJobSpec{}).Return(nil).Run(func(args mock.Arguments) {
			djSpecPtr := args.Get(2).(*core.DynamicJobSpec)
			*djSpecPtr = *djSpec
		})
		composedPBStore.OnWriteRawMatch(
			mock.MatchedBy(func(ctx context.Context) bool { return true }),
			storage.DataReference("s3://my-s3-bucket/foo/bar/futures_compiled.pb"),
			int64(1501),
			storage.Options{},
			mock.MatchedBy(func(rdr *bytes.Reader) bool { return true })).Return(errors.New("foo"))

		referenceConstructor := storageMocks.ReferenceConstructor{}
		referenceConstructor.On("ConstructReference", mock.MatchedBy(func(ctx context.Context) bool { return true }), storage.DataReference("output-dir"), "futures.pb").Return(
			storage.DataReference("s3://my-s3-bucket/foo/bar/futures.pb"), nil)
		referenceConstructor.On("ConstructReference", mock.MatchedBy(func(ctx context.Context) bool { return true }), storage.DataReference("output-dir"), "dynamic_compiled.pb").Return(
			storage.DataReference("s3://my-s3-bucket/foo/bar/dynamic_compiled.pb"), nil)
		referenceConstructor.On("ConstructReference", mock.MatchedBy(func(ctx context.Context) bool { return true }), storage.DataReference("output-dir"), "Node_1").Return(
			storage.DataReference("s3://my-s3-bucket/foo/bar/Node_1"), nil)
		referenceConstructor.On("ConstructReference", mock.MatchedBy(func(ctx context.Context) bool { return true }), storage.DataReference("s3://my-s3-bucket/foo/bar/Node_1"), "0").Return(
			storage.DataReference("s3://my-s3-bucket/foo/bar/Node_1/0"), nil)
		referenceConstructor.On("ConstructReference", mock.MatchedBy(func(ctx context.Context) bool { return true }), storage.DataReference("output-dir"), "futures_compiled.pb").Return(
			storage.DataReference("s3://my-s3-bucket/foo/bar/futures_compiled.pb"), nil)
		dataStore := &storage.DataStore{
			ComposedProtobufStore: &composedPBStore,
			ReferenceConstructor:  &referenceConstructor,
		}

		nCtx := createNodeContext("test", finalOutput, dataStore)
		nCtx.OnCurrentAttempt().Return(uint32(1))

		lpID := &core.Identifier{
			ResourceType: core.ResourceType_LAUNCH_PLAN,
			Name:         "my_plan",
			Project:      "p",
			Domain:       "d",
		}
		mockLPLauncher := &mocks5.Reader{}
		mockLPLauncher.OnGetLaunchPlanMatch(mock.Anything, mock.MatchedBy(func(id *core.Identifier) bool {
			return lpID.Name == id.Name && lpID.Domain == id.Domain && lpID.Project == id.Project && lpID.ResourceType == id.ResourceType
		})).Return(&admin.LaunchPlan{
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
		n := &mocks.Node{}
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
		execContext.EXPECT().GetParentInfo().Return(&immutableParentInfo)
		execContext.EXPECT().GetEventVersion().Return(v1alpha1.EventVersion1)
		execContext.EXPECT().GetExecutionConfig().Return(v1alpha1.ExecutionConfig{
			RecoveryExecution: v1alpha1.WorkflowExecutionIdentifier{},
		})
		nCtx.OnExecutionContext().Return(execContext)

		dCtx, err := d.buildContextualDynamicWorkflow(ctx, nCtx)
		assert.NoError(t, err)
		assert.True(t, dCtx.isDynamic)
		assert.NotNil(t, dCtx.subWorkflow)
		assert.NotNil(t, dCtx.execContext)
		assert.NotNil(t, dCtx.execContext.GetParentInfo())
		expectedParentUniqueID, err := encoding.FixedLengthUniqueIDForParts(20, []string{"c1", "2", "n1"})
		assert.Nil(t, err)
		assert.Equal(t, expectedParentUniqueID, dCtx.execContext.GetParentInfo().GetUniqueID())
		assert.Equal(t, uint32(1), dCtx.execContext.GetParentInfo().CurrentAttempt())
		assert.NotNil(t, dCtx.nodeLookup)
	})
}

type existsMetadata struct{}

func (e existsMetadata) ContentMD5() string {
	return ""
}

func (e existsMetadata) Exists() bool {
	return false
}

func (e existsMetadata) Size() int64 {
	return int64(1)
}

func (e existsMetadata) Etag() string {
	return ""
}
