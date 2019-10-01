package dynamic

import (
	"context"
	"fmt"
	"testing"

	"github.com/lyft/flyteidl/clients/go/events"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteplugins/go/tasks/v1/types/mocks"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pluginsV1 "github.com/lyft/flyteplugins/go/tasks/v1/types"
	typesV1 "k8s.io/api/core/v1"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/controller/catalog"
	mocks2 "github.com/lyft/flytepropeller/pkg/controller/executors/mocks"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/task"
)

const DataDir = storage.DataReference("test-data")
const NodeID = "n1"

var (
	enqueueWfFunc  = func(id string) {}
	fakeKubeClient = mocks2.NewFakeKubeClient()
)

func newIntegerPrimitive(value int64) *core.Primitive {
	return &core.Primitive{Value: &core.Primitive_Integer{Integer: value}}
}

func newScalarInteger(value int64) *core.Scalar {
	return &core.Scalar{
		Value: &core.Scalar_Primitive{
			Primitive: newIntegerPrimitive(value),
		},
	}
}

func newIntegerLiteral(value int64) *core.Literal {
	return &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: newScalarInteger(value),
		},
	}
}

func createTask(id string, ttype string, discoverable bool) *v1alpha1.TaskSpec {
	return &v1alpha1.TaskSpec{
		TaskTemplate: &core.TaskTemplate{
			Id:       &core.Identifier{Name: id},
			Type:     ttype,
			Metadata: &core.TaskMetadata{Discoverable: discoverable},
			Interface: &core.TypedInterface{
				Inputs: &core.VariableMap{},
				Outputs: &core.VariableMap{
					Variables: map[string]*core.Variable{
						"out1": {
							Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}},
						},
					},
				},
			},
		},
	}
}

func mockCatalogClient() catalog.Client {
	return &catalog.MockCatalogClient{
		GetFunc: func(ctx context.Context, task *core.TaskTemplate, inputPath storage.DataReference) (*core.LiteralMap, error) {
			return nil, nil
		},
		PutFunc: func(ctx context.Context, task *core.TaskTemplate, execId *core.TaskExecutionIdentifier, inputPath storage.DataReference, outputPath storage.DataReference) error {
			return nil
		},
	}
}

func createWf(id string, execID string, project string, domain string, name string) *v1alpha1.FlyteWorkflow {
	return &v1alpha1.FlyteWorkflow{
		ExecutionID: v1alpha1.WorkflowExecutionIdentifier{
			WorkflowExecutionIdentifier: &core.WorkflowExecutionIdentifier{
				Project: project,
				Domain:  domain,
				Name:    execID,
			},
		},
		Status: v1alpha1.WorkflowStatus{
			NodeStatus: map[v1alpha1.NodeID]*v1alpha1.NodeStatus{
				NodeID: {
					DataDir: DataDir,
				},
			},
		},
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		WorkflowSpec: &v1alpha1.WorkflowSpec{
			ID: id,
		},
	}
}

func createStartNode() *v1alpha1.NodeSpec {
	return &v1alpha1.NodeSpec{
		ID:   NodeID,
		Kind: v1alpha1.NodeKindStart,
		Resources: &typesV1.ResourceRequirements{
			Requests: typesV1.ResourceList{
				typesV1.ResourceCPU: resource.MustParse("1"),
			},
		},
	}
}

func createInmemoryDataStore(t testing.TB, scope promutils.Scope) *storage.DataStore {
	cfg := storage.Config{
		Type: storage.TypeMemory,
	}
	d, err := storage.NewDataStore(&cfg, scope)
	assert.NoError(t, err)
	return d
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

func TestTaskHandler_CheckNodeStatusDiscovery(t *testing.T) {
	ctx := context.Background()

	taskID := "t1"
	tk := createTask(taskID, "container", true)
	tk.Id.Project = "flytekit"

	w := createWf("w1", "w2-exec", "projTest", "domainTest", "checkNodeTestName")
	w.Tasks = map[v1alpha1.TaskID]*v1alpha1.TaskSpec{
		taskID: tk,
	}

	n := createStartNode()
	n.TaskRef = &taskID

	t.Run("TaskExecDoneDiscoveryWriteFail", func(t *testing.T) {
		taskExec := &mocks.Executor{}
		taskExec.On("GetProperties").Return(pluginsV1.ExecutorProperties{})
		taskExec.On("CheckTaskStatus",
			ctx,
			mock.MatchedBy(func(o pluginsV1.TaskContext) bool { return true }),
			mock.MatchedBy(func(o *core.TaskTemplate) bool { return true }),
		).Return(pluginsV1.TaskStatusSucceeded, nil)
		d := &task.FactoryFuncs{
			GetTaskExecutorCb: func(taskType v1alpha1.TaskType) (pluginsV1.Executor, error) {
				if taskType == tk.Type {
					return taskExec, nil
				}
				return nil, fmt.Errorf("no match")
			},
		}
		mockCatalog := catalog.MockCatalogClient{
			GetFunc: func(ctx context.Context, task *core.TaskTemplate, inputPath storage.DataReference) (*core.LiteralMap, error) {
				return nil, nil
			},
			PutFunc: func(ctx context.Context, task *core.TaskTemplate, execId *core.TaskExecutionIdentifier, inputPath storage.DataReference, outputPath storage.DataReference) error {
				return status.Errorf(codes.DeadlineExceeded, "")
			},
		}
		store := createInmemoryDataStore(t, promutils.NewTestScope())
		paramsMap := make(map[string]*core.Literal)
		paramsMap["out1"] = newIntegerLiteral(1)
		err1 := store.WriteProtobuf(ctx, "test-data/inputs.pb", storage.Options{}, &core.LiteralMap{Literals: paramsMap})
		err2 := store.WriteProtobuf(ctx, "test-data/outputs.pb", storage.Options{}, &core.LiteralMap{Literals: paramsMap})
		assert.NoError(t, err1)
		assert.NoError(t, err2)

		th := New(
			task.NewTaskHandlerForFactory(events.NewMockEventSink(), store, enqueueWfFunc, d, &mockCatalog, fakeKubeClient, promutils.NewTestScope()),
			nil,
			enqueueWfFunc,
			store,
			promutils.NewTestScope(),
		)

		prevNodeStatus := &v1alpha1.NodeStatus{Phase: v1alpha1.NodePhaseRunning}
		s, err := th.CheckNodeStatus(ctx, w, n, prevNodeStatus)
		assert.NoError(t, err)
		assert.Equal(t, handler.StatusSuccess, s)
	})

	t.Run("TaskExecDoneDiscoveryMissingOutputs", func(t *testing.T) {
		taskExec := &mocks.Executor{}
		taskExec.On("GetProperties").Return(pluginsV1.ExecutorProperties{})
		taskExec.On("CheckTaskStatus",
			ctx,
			mock.MatchedBy(func(o pluginsV1.TaskContext) bool { return true }),
			mock.MatchedBy(func(o *core.TaskTemplate) bool { return true }),
		).Return(pluginsV1.TaskStatusSucceeded, nil)
		d := &task.FactoryFuncs{
			GetTaskExecutorCb: func(taskType v1alpha1.TaskType) (pluginsV1.Executor, error) {
				if taskType == tk.Type {
					return taskExec, nil
				}
				return nil, fmt.Errorf("no match")
			},
		}
		store := createInmemoryDataStore(t, promutils.NewTestScope())
		th := New(
			task.NewTaskHandlerForFactory(events.NewMockEventSink(), store, enqueueWfFunc, d, mockCatalogClient(), fakeKubeClient, promutils.NewTestScope()),
			nil,
			enqueueWfFunc,
			store,
			promutils.NewTestScope(),
		)

		prevNodeStatus := &v1alpha1.NodeStatus{Phase: v1alpha1.NodePhaseRunning}
		s, err := th.CheckNodeStatus(ctx, w, n, prevNodeStatus)
		assert.NoError(t, err)
		assert.Equal(t, handler.PhaseRetryableFailure, s.Phase, "received: %s", s.Phase.String())
	})

	t.Run("TaskExecDoneDiscoveryWriteSuccess", func(t *testing.T) {
		taskExec := &mocks.Executor{}
		taskExec.On("GetProperties").Return(pluginsV1.ExecutorProperties{})
		taskExec.On("CheckTaskStatus",
			ctx,
			mock.MatchedBy(func(o pluginsV1.TaskContext) bool { return true }),
			mock.MatchedBy(func(o *core.TaskTemplate) bool { return true }),
		).Return(pluginsV1.TaskStatusSucceeded, nil)
		d := &task.FactoryFuncs{
			GetTaskExecutorCb: func(taskType v1alpha1.TaskType) (pluginsV1.Executor, error) {
				if taskType == tk.Type {
					return taskExec, nil
				}
				return nil, fmt.Errorf("no match")
			},
		}
		store := createInmemoryDataStore(t, promutils.NewTestScope())
		paramsMap := make(map[string]*core.Literal)
		paramsMap["out1"] = newIntegerLiteral(100)
		err1 := store.WriteProtobuf(ctx, "test-data/inputs.pb", storage.Options{}, &core.LiteralMap{Literals: paramsMap})
		err2 := store.WriteProtobuf(ctx, "test-data/outputs.pb", storage.Options{}, &core.LiteralMap{Literals: paramsMap})
		assert.NoError(t, err1)
		assert.NoError(t, err2)
		th := New(
			task.NewTaskHandlerForFactory(events.NewMockEventSink(), store, enqueueWfFunc, d, mockCatalogClient(), fakeKubeClient, promutils.NewTestScope()),
			nil,
			enqueueWfFunc,
			store,
			promutils.NewTestScope(),
		)

		prevNodeStatus := &v1alpha1.NodeStatus{Phase: v1alpha1.NodePhaseRunning}
		s, err := th.CheckNodeStatus(ctx, w, n, prevNodeStatus)
		assert.NoError(t, err)
		assert.Equal(t, handler.StatusSuccess, s)
	})
}

func TestBuildFlyteWorkflow(t *testing.T) {
	ctx := context.Background()

	dynamicSpec := createDynamicJobSpec()
	tk := dynamicSpec.Tasks[0]
	w := createWf("w1", "w2-exec", "projTest", "domainTest", "checkNodeTestName")
	w.Tasks = map[v1alpha1.TaskID]*v1alpha1.TaskSpec{
		tk.Id.String(): {TaskTemplate: tk},
	}

	n := createStartNode()
	taskID := tk.Id.String()
	n.TaskRef = &taskID

	w.Nodes = map[v1alpha1.NodeID]*v1alpha1.NodeSpec{
		NodeID: n,
	}

	taskExec := &mocks.Executor{}
	taskExec.On("GetProperties").Return(pluginsV1.ExecutorProperties{})
	taskExec.On("CheckTaskStatus",
		ctx,
		mock.MatchedBy(func(o pluginsV1.TaskContext) bool { return true }),
		mock.MatchedBy(func(o *core.TaskTemplate) bool { return true }),
	).Return(pluginsV1.TaskStatusSucceeded, nil)
	d := &task.FactoryFuncs{
		GetTaskExecutorCb: func(taskType v1alpha1.TaskType) (pluginsV1.Executor, error) {
			if taskType == tk.Type {
				return taskExec, nil
			}
			return nil, fmt.Errorf("no match")
		},
	}
	store := createInmemoryDataStore(t, promutils.NewTestScope())
	assert.NoError(t, store.WriteProtobuf(ctx, "/test-data/futures.pb", storage.Options{}, dynamicSpec))

	paramsMap := make(map[string]*core.Literal)
	paramsMap["out1"] = newIntegerLiteral(100)
	assert.NoError(t, store.WriteProtobuf(ctx, "/test-data/inputs.pb", storage.Options{}, &core.LiteralMap{Literals: paramsMap}))
	assert.NoError(t, store.WriteProtobuf(ctx, "/test-data/outputs.pb", storage.Options{}, &core.LiteralMap{Literals: paramsMap}))

	th := New(
		task.NewTaskHandlerForFactory(events.NewMockEventSink(), store, enqueueWfFunc, d, mockCatalogClient(), fakeKubeClient, promutils.NewTestScope()),
		nil,
		enqueueWfFunc,
		store,
		promutils.NewTestScope(),
	).(dynamicNodeHandler)

	prevNodeStatus := &v1alpha1.NodeStatus{Phase: v1alpha1.NodePhaseRunning}

	n2, found := w.GetNode(NodeID)
	assert.True(t, found)

	_, isdynamic, err := th.buildFlyteWorkflow(ctx, w, n2, "/test-data/", prevNodeStatus)
	assert.NoError(t, err)
	assert.True(t, isdynamic)
}
