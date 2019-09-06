package task

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	mocks2 "github.com/lyft/flytepropeller/pkg/controller/executors/mocks"

	"github.com/lyft/flytestdlib/promutils"

	"github.com/lyft/flyteidl/clients/go/events"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	pluginsV1 "github.com/lyft/flyteplugins/go/tasks/v1/types"
	"github.com/lyft/flyteplugins/go/tasks/v1/types/mocks"
	"github.com/lyft/flytestdlib/storage"
	regErrors "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	typesV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/controller/catalog"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/errors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
)

const DataDir = storage.DataReference("test-data")
const NodeID = "n1"

var (
	enqueueWfFunc  = func(id string) {}
	fakeKubeClient = mocks2.NewFakeKubeClient()
)

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
						"out1": &core.Variable{
							Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}},
						},
					},
				},
			},
		},
	}
}

func createDummyExec() *mocks.Executor {
	dummyExec := &mocks.Executor{}
	dummyExec.On("Initialize",
		mock.AnythingOfType(reflect.TypeOf(context.TODO()).String()),
		mock.AnythingOfType(reflect.TypeOf(pluginsV1.ExecutorInitializationParameters{}).String()),
	).Return(nil)
	dummyExec.On("GetID").Return("test")

	return dummyExec
}

func TestTaskHandler_Initialize(t *testing.T) {
	ctx := context.TODO()
	t.Run("NoHandlers", func(t *testing.T) {
		d := &FactoryFuncs{}
		th := NewTaskHandlerForFactory(events.NewMockEventSink(), nil, enqueueWfFunc, d, mockCatalogClient(), fakeKubeClient, promutils.NewTestScope())
		assert.NoError(t, th.Initialize(context.TODO()))
	})

	t.Run("SomeHandler", func(t *testing.T) {
		d := &FactoryFuncs{
			ListAllTaskExecutorsCb: func() []pluginsV1.Executor {
				return []pluginsV1.Executor{
					createDummyExec(),
				}
			},
		}
		th := NewTaskHandlerForFactory(events.NewMockEventSink(), nil, enqueueWfFunc, d, mockCatalogClient(), fakeKubeClient, promutils.NewTestScope())
		assert.NoError(t, th.Initialize(ctx))
	})
}

func TestTaskHandler_HandleFailingNode(t *testing.T) {
	ctx := context.Background()
	d := &FactoryFuncs{}
	th := NewTaskHandlerForFactory(events.NewMockEventSink(), nil, enqueueWfFunc, d, mockCatalogClient(), fakeKubeClient, promutils.NewTestScope())

	w := createWf("w1", "w2-exec", "project", "domain", "execName1")
	n := createStartNode()
	s, err := th.HandleFailingNode(ctx, w, n)
	assert.NoError(t, err)
	assert.Equal(t, handler.PhaseFailed, s.Phase)
	assert.Error(t, s.Err)
}

func TestTaskHandler_GetTaskExecutorContext(t *testing.T) {
	ctx := context.Background()
	const execName = "w1-exec"
	t.Run("NoTaskId", func(t *testing.T) {
		w := createWf("w1", execName, "project", "domain", "execName1")
		n := createStartNode()
		d := &FactoryFuncs{}
		th := NewTaskHandlerForFactory(events.NewMockEventSink(), nil, enqueueWfFunc, d, mockCatalogClient(), fakeKubeClient, promutils.NewTestScope()).(*taskHandler)

		_, _, _, err := th.GetTaskExecutorContext(ctx, w, n)
		assert.Error(t, err)
		assert.True(t, errors.Matches(err, errors.BadSpecificationError))
	})

	t.Run("NoTaskMatch", func(t *testing.T) {
		taskID := "t1"
		w := createWf("w1", execName, "project", "domain", "execName1")
		n := createStartNode()
		n.TaskRef = &taskID

		d := &FactoryFuncs{}
		th := NewTaskHandlerForFactory(events.NewMockEventSink(), nil, enqueueWfFunc, d, mockCatalogClient(), fakeKubeClient, promutils.NewTestScope()).(*taskHandler)
		_, _, _, err := th.GetTaskExecutorContext(ctx, w, n)
		assert.Error(t, err)
		assert.True(t, errors.Matches(err, errors.BadSpecificationError))
	})

	t.Run("TaskMatchNoExecutor", func(t *testing.T) {
		taskID := "t1"
		task := createTask(taskID, "dynamic", false)

		w := createWf("w1", execName, "project", "domain", "execName1")
		w.Tasks = map[v1alpha1.TaskID]*v1alpha1.TaskSpec{
			taskID: task,
		}

		n := createStartNode()
		n.TaskRef = &taskID

		d := &FactoryFuncs{}
		th := NewTaskHandlerForFactory(events.NewMockEventSink(), nil, enqueueWfFunc, d, mockCatalogClient(), fakeKubeClient, promutils.NewTestScope()).(*taskHandler)
		_, _, _, err := th.GetTaskExecutorContext(ctx, w, n)
		assert.Error(t, err)
		assert.True(t, errors.Matches(err, errors.UnsupportedTaskTypeError))
	})

	t.Run("TaskMatch", func(t *testing.T) {
		taskID := "t1"
		task := createTask(taskID, "container", false)
		w := createWf("w1", execName, "project", "domain", "execName1")
		w.Tasks = map[v1alpha1.TaskID]*v1alpha1.TaskSpec{
			taskID: task,
		}
		w.ServiceAccountName = "service-account"
		n := createStartNode()
		n.TaskRef = &taskID

		taskExec := &mocks.Executor{}
		d := &FactoryFuncs{
			GetTaskExecutorCb: func(taskType v1alpha1.TaskType) (pluginsV1.Executor, error) {
				if taskType == task.Type {
					return taskExec, nil
				}

				return nil, regErrors.New("No match")
			},
		}
		th := NewTaskHandlerForFactory(events.NewMockEventSink(), nil, enqueueWfFunc, d, mockCatalogClient(), fakeKubeClient, promutils.NewTestScope()).(*taskHandler)
		te, receivedTask, tc, err := th.GetTaskExecutorContext(ctx, w, n)
		if assert.NoError(t, err) {
			assert.Equal(t, taskExec, te)
			if assert.NotNil(t, tc) {
				assert.Equal(t, "execName1-n1-0", tc.GetTaskExecutionID().GetGeneratedName())
				assert.Equal(t, DataDir, tc.GetDataDir())
				assert.NotNil(t, tc.GetOverrides())
				assert.NotNil(t, tc.GetOverrides().GetResources())
				assert.NotEmpty(t, tc.GetOverrides().GetResources().Requests)
				assert.Equal(t, "service-account", tc.GetK8sServiceAccount())
			}
			assert.Equal(t, task, receivedTask)
		}
	})

	t.Run("TaskMatchAttempt>0", func(t *testing.T) {
		taskID := "t1"
		task := createTask(taskID, "container", false)
		w := createWf("w1", execName, "project", "domain", "execName1")
		w.Tasks = map[v1alpha1.TaskID]*v1alpha1.TaskSpec{
			taskID: task,
		}
		n := createStartNode()
		n.TaskRef = &taskID

		status := w.Status.GetNodeExecutionStatus(n.ID).(*v1alpha1.NodeStatus)
		status.Attempts = 2

		taskExec := &mocks.Executor{}
		d := &FactoryFuncs{
			GetTaskExecutorCb: func(taskType v1alpha1.TaskType) (pluginsV1.Executor, error) {
				if taskType == task.Type {
					return taskExec, nil
				}
				return nil, regErrors.New("No match")
			},
		}
		th := NewTaskHandlerForFactory(events.NewMockEventSink(), nil, enqueueWfFunc, d, mockCatalogClient(), fakeKubeClient, promutils.NewTestScope()).(*taskHandler)
		te, receivedTask, tc, err := th.GetTaskExecutorContext(ctx, w, n)
		if assert.NoError(t, err) {
			assert.Equal(t, taskExec, te)
			if assert.NotNil(t, tc) {
				assert.Equal(t, "execName1-n1-2", tc.GetTaskExecutionID().GetGeneratedName())
				assert.Equal(t, DataDir, tc.GetDataDir())
				assert.NotNil(t, tc.GetOverrides())
				assert.NotNil(t, tc.GetOverrides().GetResources())
				assert.NotEmpty(t, tc.GetOverrides().GetResources().Requests)
			}
			assert.Equal(t, task, receivedTask)
		}
	})

}

func TestTaskHandler_StartNode(t *testing.T) {
	ctx := context.Background()
	taskID := "t1"
	task := createTask(taskID, "container", false)
	w := createWf("w2", "w2-exec", "project", "domain", "execName")
	w.Tasks = map[v1alpha1.TaskID]*v1alpha1.TaskSpec{
		taskID: task,
	}
	n := createStartNode()
	n.TaskRef = &taskID

	t.Run("NoTaskExec", func(t *testing.T) {
		taskExec := &mocks.Executor{}
		d := &FactoryFuncs{
			GetTaskExecutorCb: func(taskType v1alpha1.TaskType) (pluginsV1.Executor, error) {
				if taskType == task.Type {
					return nil, regErrors.New("No match")
				}
				return taskExec, nil
			},
		}
		th := NewTaskHandlerForFactory(events.NewMockEventSink(), nil, enqueueWfFunc, d, mockCatalogClient(), fakeKubeClient, promutils.NewTestScope())

		s, err := th.StartNode(ctx, w, n, nil)
		assert.NoError(t, err)
		assert.Error(t, s.Err)
		assert.True(t, errors.Matches(s.Err, errors.CausedByError))
	})

	t.Run("TaskExecStartFail", func(t *testing.T) {
		taskExec := &mocks.Executor{}
		taskExec.On("GetProperties").Return(pluginsV1.ExecutorProperties{})
		taskExec.On("StartTask",
			ctx,
			mock.MatchedBy(func(o pluginsV1.TaskContext) bool { return true }),
			mock.MatchedBy(func(o *core.TaskTemplate) bool { return true }),
			mock.MatchedBy(func(o *core.LiteralMap) bool { return true }),
		).Return(pluginsV1.TaskStatusPermanentFailure(regErrors.New("Failed")), nil)
		d := &FactoryFuncs{
			GetTaskExecutorCb: func(taskType v1alpha1.TaskType) (pluginsV1.Executor, error) {
				if taskType == task.Type {
					return taskExec, nil
				}
				return nil, regErrors.New("No match")
			},
		}
		th := NewTaskHandlerForFactory(events.NewMockEventSink(), nil, enqueueWfFunc, d, mockCatalogClient(), fakeKubeClient, promutils.NewTestScope())

		s, err := th.StartNode(ctx, w, n, nil)
		assert.NoError(t, err)
		assert.Equal(t, handler.PhaseFailed, s.Phase)
	})

	t.Run("TaskExecStartPanic", func(t *testing.T) {
		taskExec := &mocks.Executor{}
		taskExec.On("GetProperties").Return(pluginsV1.ExecutorProperties{})
		taskExec.On("StartTask",
			ctx,
			mock.MatchedBy(func(o pluginsV1.TaskContext) bool { return true }),
			mock.MatchedBy(func(o *core.TaskTemplate) bool { return true }),
			mock.MatchedBy(func(o *core.LiteralMap) bool { return true }),
		).Return(
			func(ctx context.Context, taskCtx pluginsV1.TaskContext, task *core.TaskTemplate, inputs *core.LiteralMap) (pluginsV1.TaskStatus, error) {
				panic("failed in execution")
			},
		)
		d := &FactoryFuncs{
			GetTaskExecutorCb: func(taskType v1alpha1.TaskType) (pluginsV1.Executor, error) {
				if taskType == task.Type {
					return taskExec, nil
				}
				return nil, regErrors.New("No match")
			},
		}
		th := NewTaskHandlerForFactory(events.NewMockEventSink(), nil, enqueueWfFunc, d, mockCatalogClient(), fakeKubeClient, promutils.NewTestScope())
		s, err := th.StartNode(ctx, w, n, nil)
		assert.Error(t, err)
		assert.Equal(t, handler.PhaseUndefined, s.Phase)
	})

	t.Run("TaskExecStarted", func(t *testing.T) {
		taskExec := &mocks.Executor{}
		taskExec.On("GetProperties").Return(pluginsV1.ExecutorProperties{})
		taskExec.On("StartTask",
			ctx,
			mock.MatchedBy(func(o pluginsV1.TaskContext) bool { return true }),
			mock.MatchedBy(func(o *core.TaskTemplate) bool { return true }),
			mock.MatchedBy(func(o *core.LiteralMap) bool { return true }),
		).Return(pluginsV1.TaskStatusRunning, nil)
		d := &FactoryFuncs{
			GetTaskExecutorCb: func(taskType v1alpha1.TaskType) (pluginsV1.Executor, error) {
				if taskType == task.Type {
					return taskExec, nil
				}
				return nil, regErrors.New("No match")
			},
		}
		th := NewTaskHandlerForFactory(events.NewMockEventSink(), nil, enqueueWfFunc, d, mockCatalogClient(), fakeKubeClient, promutils.NewTestScope())

		s, err := th.StartNode(ctx, w, n, nil)
		assert.NoError(t, err)
		assert.Equal(t, handler.StatusRunning, s)
	})
}

func TestTaskHandler_StartNodeDiscoverable(t *testing.T) {
	ctx := context.Background()
	taskID := "t1"
	task := createTask(taskID, "container", true)
	task.Id.Project = "flytekit"
	w := createWf("w2", "w2-exec", "flytekit", "domain", "execName")
	w.Tasks = map[v1alpha1.TaskID]*v1alpha1.TaskSpec{
		taskID: task,
	}
	n := createStartNode()
	n.TaskRef = &taskID

	t.Run("TaskExecStartNodeDiscoveryFail", func(t *testing.T) {
		taskExec := &mocks.Executor{}
		taskExec.On("GetProperties").Return(pluginsV1.ExecutorProperties{})
		taskExec.On("StartTask",
			ctx,
			mock.MatchedBy(func(o pluginsV1.TaskContext) bool { return true }),
			mock.MatchedBy(func(o *core.TaskTemplate) bool { return true }),
			mock.MatchedBy(func(o *core.LiteralMap) bool { return true }),
		).Return(pluginsV1.TaskStatusRunning, nil)
		d := &FactoryFuncs{
			GetTaskExecutorCb: func(taskType v1alpha1.TaskType) (pluginsV1.Executor, error) {
				if taskType == task.Type {
					return taskExec, nil
				}
				return nil, regErrors.New("No match")
			},
		}
		mockCatalog := catalog.MockCatalogClient{
			GetFunc: func(ctx context.Context, task *core.TaskTemplate, inputPath storage.DataReference) (*core.LiteralMap, error) {
				return nil, regErrors.Errorf("error")
			},
			PutFunc: func(ctx context.Context, task *core.TaskTemplate, execId *core.TaskExecutionIdentifier, inputPath storage.DataReference, outputPath storage.DataReference) error {
				return nil
			},
		}
		th := NewTaskHandlerForFactory(events.NewMockEventSink(), nil, enqueueWfFunc, d, &mockCatalog, fakeKubeClient, promutils.NewTestScope())

		s, err := th.StartNode(ctx, w, n, nil)
		assert.NoError(t, err)
		assert.Equal(t, handler.StatusRunning, s)
	})

	t.Run("TaskExecStartNodeDiscoveryMiss", func(t *testing.T) {
		taskExec := &mocks.Executor{}
		taskExec.On("GetProperties").Return(pluginsV1.ExecutorProperties{})
		taskExec.On("StartTask",
			ctx,
			mock.MatchedBy(func(o pluginsV1.TaskContext) bool { return true }),
			mock.MatchedBy(func(o *core.TaskTemplate) bool { return true }),
			mock.MatchedBy(func(o *core.LiteralMap) bool { return true }),
		).Return(pluginsV1.TaskStatusRunning, nil)
		d := &FactoryFuncs{
			GetTaskExecutorCb: func(taskType v1alpha1.TaskType) (pluginsV1.Executor, error) {
				if taskType == task.Type {
					return taskExec, nil
				}
				return nil, regErrors.New("No match")
			},
		}
		mockCatalog := catalog.MockCatalogClient{
			GetFunc: func(ctx context.Context, task *core.TaskTemplate, inputPath storage.DataReference) (*core.LiteralMap, error) {
				return nil, status.Errorf(codes.NotFound, "not found")
			},
			PutFunc: func(ctx context.Context, task *core.TaskTemplate, execId *core.TaskExecutionIdentifier, inputPath storage.DataReference, outputPath storage.DataReference) error {
				return nil
			},
		}
		th := NewTaskHandlerForFactory(events.NewMockEventSink(), nil, enqueueWfFunc, d, &mockCatalog, fakeKubeClient, promutils.NewTestScope())

		s, err := th.StartNode(ctx, w, n, nil)
		assert.NoError(t, err)
		assert.Equal(t, handler.StatusRunning, s)
	})

	t.Run("TaskExecStartNodeDiscoveryHit", func(t *testing.T) {
		taskExec := &mocks.Executor{}
		taskExec.On("GetProperties").Return(pluginsV1.ExecutorProperties{})
		taskExec.On("StartTask",
			ctx,
			mock.MatchedBy(func(o pluginsV1.TaskContext) bool { return true }),
			mock.MatchedBy(func(o *core.TaskTemplate) bool { return true }),
			mock.MatchedBy(func(o *core.LiteralMap) bool { return true }),
		).Return(pluginsV1.TaskStatusRunning, nil)
		d := &FactoryFuncs{
			GetTaskExecutorCb: func(taskType v1alpha1.TaskType) (pluginsV1.Executor, error) {
				if taskType == task.Type {
					return taskExec, nil
				}
				return nil, regErrors.New("No match")
			},
		}
		mockCatalog := catalog.MockCatalogClient{
			GetFunc: func(ctx context.Context, task *core.TaskTemplate, inputPath storage.DataReference) (*core.LiteralMap, error) {
				paramsMap := make(map[string]*core.Literal)
				paramsMap["out1"] = newIntegerLiteral(100)

				return &core.LiteralMap{
					Literals: paramsMap,
				}, nil
			},
			PutFunc: func(ctx context.Context, task *core.TaskTemplate, execId *core.TaskExecutionIdentifier, inputPath storage.DataReference, outputPath storage.DataReference) error {
				return nil
			},
		}
		store := createInmemoryDataStore(t, testScope.NewSubScope("12"))
		th := NewTaskHandlerForFactory(events.NewMockEventSink(), store, enqueueWfFunc, d, &mockCatalog, fakeKubeClient, promutils.NewTestScope())

		s, err := th.StartNode(ctx, w, n, nil)
		assert.NoError(t, err)
		assert.Equal(t, handler.StatusSuccess, s)
	})
}

func TestTaskHandler_AbortNode(t *testing.T) {
	ctx := context.Background()
	taskID := "t1"
	task := createTask(taskID, "container", false)
	w := createWf("w2", "w2-exec", "project", "domain", "execName")
	w.Tasks = map[v1alpha1.TaskID]*v1alpha1.TaskSpec{
		taskID: task,
	}
	n := createStartNode()
	n.TaskRef = &taskID

	t.Run("NoTaskExec", func(t *testing.T) {
		taskExec := &mocks.Executor{}
		taskExec.On("GetProperties").Return(pluginsV1.ExecutorProperties{})
		d := &FactoryFuncs{
			GetTaskExecutorCb: func(taskType v1alpha1.TaskType) (pluginsV1.Executor, error) {
				if taskType == task.Type {
					return nil, regErrors.New("No match")
				}
				return taskExec, nil
			},
		}
		th := NewTaskHandlerForFactory(events.NewMockEventSink(), nil, enqueueWfFunc, d, mockCatalogClient(), fakeKubeClient, promutils.NewTestScope())

		err := th.AbortNode(ctx, w, n)
		assert.Error(t, err)
		assert.True(t, errors.Matches(err, errors.CausedByError))
	})

	t.Run("TaskExecKillFail", func(t *testing.T) {
		taskExec := &mocks.Executor{}
		taskExec.On("GetProperties").Return(pluginsV1.ExecutorProperties{})
		taskExec.On("KillTask",
			ctx,
			mock.MatchedBy(func(o pluginsV1.TaskContext) bool { return true }),
			mock.Anything,
		).Return(regErrors.New("Failed"))
		d := &FactoryFuncs{
			GetTaskExecutorCb: func(taskType v1alpha1.TaskType) (pluginsV1.Executor, error) {
				if taskType == task.Type {
					return taskExec, nil
				}
				return nil, regErrors.New("No match")
			},
		}
		th := NewTaskHandlerForFactory(events.NewMockEventSink(), nil, enqueueWfFunc, d, mockCatalogClient(), fakeKubeClient, promutils.NewTestScope())

		err := th.AbortNode(ctx, w, n)
		assert.Error(t, err)
		assert.True(t, errors.Matches(err, errors.CausedByError))
	})

	t.Run("TaskExecKilled", func(t *testing.T) {
		taskExec := &mocks.Executor{}
		taskExec.On("GetProperties").Return(pluginsV1.ExecutorProperties{})
		taskExec.On("KillTask",
			ctx,
			mock.MatchedBy(func(o pluginsV1.TaskContext) bool { return true }),
			mock.Anything,
		).Return(nil)
		d := &FactoryFuncs{
			GetTaskExecutorCb: func(taskType v1alpha1.TaskType) (pluginsV1.Executor, error) {
				if taskType == task.Type {
					return taskExec, nil
				}
				return nil, regErrors.New("No match")
			},
		}
		th := NewTaskHandlerForFactory(events.NewMockEventSink(), nil, enqueueWfFunc, d, mockCatalogClient(), fakeKubeClient, promutils.NewTestScope())

		err := th.AbortNode(ctx, w, n)
		assert.NoError(t, err)
	})
}

func createInmemoryDataStore(t testing.TB, scope promutils.Scope) *storage.DataStore {
	cfg := storage.Config{
		Type: storage.TypeMemory,
	}
	d, err := storage.NewDataStore(&cfg, scope)
	assert.NoError(t, err)
	return d
}

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

var testScope = promutils.NewScope("test_wfexec")

func TestTaskHandler_CheckNodeStatus(t *testing.T) {
	ctx := context.Background()

	taskID := "t1"
	task := createTask(taskID, "container", false)
	w := createWf("w1", "w2-exec", "projTest", "domainTest", "checkNodeTestName")
	w.Tasks = map[v1alpha1.TaskID]*v1alpha1.TaskSpec{
		taskID: task,
	}
	n := createStartNode()
	n.TaskRef = &taskID

	t.Run("NoTaskExec", func(t *testing.T) {
		taskExec := &mocks.Executor{}
		taskExec.On("GetProperties").Return(pluginsV1.ExecutorProperties{})
		d := &FactoryFuncs{
			GetTaskExecutorCb: func(taskType v1alpha1.TaskType) (pluginsV1.Executor, error) {
				if taskType == task.Type {
					return nil, regErrors.New("No match")
				}
				return taskExec, nil
			},
		}
		th := NewTaskHandlerForFactory(events.NewMockEventSink(), nil, enqueueWfFunc, d, mockCatalogClient(), fakeKubeClient, promutils.NewTestScope())

		prevNodeStatus := &v1alpha1.NodeStatus{Phase: v1alpha1.NodePhaseNotYetStarted}
		s, err := th.CheckNodeStatus(ctx, w, n, prevNodeStatus)
		assert.NoError(t, err)
		assert.True(t, errors.Matches(s.Err, errors.CausedByError))
	})

	t.Run("TaskExecStartFail", func(t *testing.T) {
		taskExec := &mocks.Executor{}
		taskExec.On("GetProperties").Return(pluginsV1.ExecutorProperties{})
		taskExec.On("CheckTaskStatus",
			ctx,
			mock.MatchedBy(func(o pluginsV1.TaskContext) bool { return true }),
			mock.MatchedBy(func(o *core.TaskTemplate) bool { return true }),
		).Return(pluginsV1.TaskStatusPermanentFailure(regErrors.New("Failed")), nil)
		d := &FactoryFuncs{
			GetTaskExecutorCb: func(taskType v1alpha1.TaskType) (pluginsV1.Executor, error) {
				if taskType == task.Type {
					return taskExec, nil
				}
				return nil, regErrors.New("No match")
			},
		}
		th := NewTaskHandlerForFactory(events.NewMockEventSink(), nil, enqueueWfFunc, d, mockCatalogClient(), fakeKubeClient, promutils.NewTestScope())

		prevNodeStatus := &v1alpha1.NodeStatus{Phase: v1alpha1.NodePhaseRunning}
		s, err := th.CheckNodeStatus(ctx, w, n, prevNodeStatus)
		assert.NoError(t, err)
		assert.Equal(t, handler.PhaseFailed, s.Phase)
	})

	t.Run("TaskExecCheckPanic", func(t *testing.T) {
		taskExec := &mocks.Executor{}
		taskExec.On("GetProperties").Return(pluginsV1.ExecutorProperties{})
		taskExec.On("CheckTaskStatus",
			ctx,
			mock.MatchedBy(func(o pluginsV1.TaskContext) bool { return true }),
			mock.MatchedBy(func(o *core.TaskTemplate) bool { return true }),
		).Return(func(ctx context.Context, taskCtx pluginsV1.TaskContext, task *core.TaskTemplate) (status pluginsV1.TaskStatus, err error) {
			panic("failed in execution")
		})
		d := &FactoryFuncs{
			GetTaskExecutorCb: func(taskType v1alpha1.TaskType) (pluginsV1.Executor, error) {
				if taskType == task.Type {
					return taskExec, nil
				}
				return nil, regErrors.New("No match")
			},
		}
		th := NewTaskHandlerForFactory(events.NewMockEventSink(), nil, enqueueWfFunc, d, mockCatalogClient(), fakeKubeClient, promutils.NewTestScope())
		prevNodeStatus := &v1alpha1.NodeStatus{Phase: v1alpha1.NodePhaseRunning}
		s, err := th.CheckNodeStatus(ctx, w, n, prevNodeStatus)
		assert.Error(t, err)
		assert.Equal(t, handler.PhaseUndefined, s.Phase)
	})

	t.Run("TaskExecRunning", func(t *testing.T) {
		taskExec := &mocks.Executor{}
		taskExec.On("GetProperties").Return(pluginsV1.ExecutorProperties{})
		taskExec.On("CheckTaskStatus",
			ctx,
			mock.MatchedBy(func(o pluginsV1.TaskContext) bool { return true }),
			mock.MatchedBy(func(o *core.TaskTemplate) bool { return true }),
		).Return(pluginsV1.TaskStatusRunning, nil)
		d := &FactoryFuncs{
			GetTaskExecutorCb: func(taskType v1alpha1.TaskType) (pluginsV1.Executor, error) {
				if taskType == task.Type {
					return taskExec, nil
				}
				return nil, regErrors.New("No match")
			},
		}
		th := NewTaskHandlerForFactory(events.NewMockEventSink(), nil, enqueueWfFunc, d, mockCatalogClient(), fakeKubeClient, promutils.NewTestScope())

		prevNodeStatus := &v1alpha1.NodeStatus{Phase: v1alpha1.NodePhaseRunning}
		s, err := th.CheckNodeStatus(ctx, w, n, prevNodeStatus)
		assert.NoError(t, err)
		assert.Equal(t, handler.StatusRunning, s)
	})

	t.Run("TaskExecDone", func(t *testing.T) {
		taskExec := &mocks.Executor{}
		taskExec.On("GetProperties").Return(pluginsV1.ExecutorProperties{})
		taskExec.On("CheckTaskStatus",
			ctx,
			mock.MatchedBy(func(o pluginsV1.TaskContext) bool { return true }),
			mock.MatchedBy(func(o *core.TaskTemplate) bool { return true }),
		).Return(pluginsV1.TaskStatusSucceeded, nil)
		d := &FactoryFuncs{
			GetTaskExecutorCb: func(taskType v1alpha1.TaskType) (pluginsV1.Executor, error) {
				if taskType == task.Type {
					return taskExec, nil
				}
				return nil, regErrors.New("No match")
			},
		}

		store := createInmemoryDataStore(t, testScope.NewSubScope("4"))
		paramsMap := make(map[string]*core.Literal)
		paramsMap["out1"] = newIntegerLiteral(100)
		err1 := store.WriteProtobuf(ctx, "test-data/inputs.pb", storage.Options{}, &core.LiteralMap{Literals: paramsMap})
		err2 := store.WriteProtobuf(ctx, "test-data/outputs.pb", storage.Options{}, &core.LiteralMap{Literals: paramsMap})
		assert.NoError(t, err1)
		assert.NoError(t, err2)

		th := NewTaskHandlerForFactory(events.NewMockEventSink(), store, enqueueWfFunc, d, mockCatalogClient(), fakeKubeClient, promutils.NewTestScope())

		prevNodeStatus := &v1alpha1.NodeStatus{Phase: v1alpha1.NodePhaseRunning}

		s, err := th.CheckNodeStatus(ctx, w, n, prevNodeStatus)
		assert.NoError(t, err)
		assert.Equal(t, handler.StatusSuccess, s)
	})
}

func TestConvertTaskPhaseToHandlerStatus(t *testing.T) {
	expectedErr := fmt.Errorf("failed")
	tests := []struct {
		name    string
		status  pluginsV1.TaskStatus
		hs      handler.Status
		isError bool
	}{
		{"undefined", pluginsV1.TaskStatusUndefined, handler.StatusUndefined, true},
		{"running", pluginsV1.TaskStatusRunning, handler.StatusRunning, false},
		{"queued", pluginsV1.TaskStatusQueued, handler.StatusRunning, false},
		{"succeeded", pluginsV1.TaskStatusSucceeded, handler.StatusSuccess, false},
		{"unknown", pluginsV1.TaskStatusUnknown, handler.StatusUndefined, true},
		{"retryable", pluginsV1.TaskStatusRetryableFailure(expectedErr), handler.StatusRetryableFailure(expectedErr), false},
		{"failed", pluginsV1.TaskStatusPermanentFailure(expectedErr), handler.StatusFailed(expectedErr), false},
		{"undefined", pluginsV1.TaskStatusUndefined, handler.StatusUndefined, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			hs, err := ConvertTaskPhaseToHandlerStatus(test.status)
			assert.Equal(t, hs, test.hs)
			if test.isError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
