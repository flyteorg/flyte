package nodes

import (
	"context"
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytepropeller/pkg/controller/executors"
	mocks2 "github.com/flyteorg/flytepropeller/pkg/controller/executors/mocks"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
)

type TaskReader struct{}

func (t TaskReader) Read(ctx context.Context) (*core.TaskTemplate, error) { return nil, nil }
func (t TaskReader) GetTaskType() v1alpha1.TaskType                       { return "" }
func (t TaskReader) GetTaskID() *core.Identifier {
	return &core.Identifier{Project: "p", Domain: "d", Name: "task-name"}
}

type parentInfo struct {
	executors.ImmutableParentInfo
}

func Test_NodeContext(t *testing.T) {
	ns := mocks.ExecutableNodeStatus{}
	ns.On("GetDataDir").Return(storage.DataReference("data-dir"))
	ns.On("GetPhase").Return(v1alpha1.NodePhaseNotYetStarted)

	childDatadir := v1alpha1.DataReference("test")
	dataStore, _ := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	w1 := &v1alpha1.FlyteWorkflow{
		Status: v1alpha1.WorkflowStatus{
			NodeStatus: map[v1alpha1.NodeID]*v1alpha1.NodeStatus{
				"childNodeID": {
					DataDir: childDatadir,
				},
			},
		},
		DataReferenceConstructor: dataStore,
	}

	taskID := "taskID"
	n := &v1alpha1.NodeSpec{
		ID:      "id",
		TaskRef: &taskID,
		Kind:    v1alpha1.NodeKindTask,
	}
	s, _ := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	p := parentInfo{}
	execContext := executors.NewExecutionContext(w1, nil, nil, p, nil)
	nCtx := newNodeExecContext(context.TODO(), s, execContext, w1, n, nil, nil, false, 0, 2, nil, TaskReader{}, nil, nil, "s3://bucket", ioutils.NewConstantShardSelector([]string{"x"}))
	assert.Equal(t, "id", nCtx.NodeExecutionMetadata().GetLabels()["node-id"])
	assert.Equal(t, "false", nCtx.NodeExecutionMetadata().GetLabels()["interruptible"])
	assert.Equal(t, "task-name", nCtx.NodeExecutionMetadata().GetLabels()["task-name"])
	assert.Equal(t, p, nCtx.ExecutionContext().GetParentInfo())
}

func Test_NodeContextDefault(t *testing.T) {
	ctx := context.Background()

	w1 := &v1alpha1.FlyteWorkflow{
		NodeDefaults: v1alpha1.NodeDefaults{Interruptible: false},
		RawOutputDataConfig: v1alpha1.RawOutputDataConfig{RawOutputDataConfig: &admin.RawOutputDataConfig{
			OutputLocationPrefix: ""},
		},
		WorkflowSpec: &v1alpha1.WorkflowSpec{
			ID: "some.workflow",
		},
		Tasks: map[v1alpha1.TaskID]*v1alpha1.TaskSpec{
			"taskID": {
				TaskTemplate: &core.TaskTemplate{
					Id: &core.Identifier{
						ResourceType: 1,
						Project:      "proj",
						Domain:       "domain",
						Name:         "taskID",
						Version:      "abc",
					},
				},
			},
		},
	}
	dataStore, _ := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	taskID := "taskID"
	n := &v1alpha1.NodeSpec{
		ID:      "id",
		TaskRef: &taskID,
		Kind:    v1alpha1.NodeKindTask,
	}
	nodeLookup := &mocks2.NodeLookup{}
	nodeLookup.OnGetNode("node-a").Return(n, true)
	nodeLookup.OnGetNodeExecutionStatus(ctx, "node-a").Return(&v1alpha1.NodeStatus{
		SystemFailures: 0,
	})

	nodeExecutor := nodeExecutor{
		interruptibleFailureThreshold: 0,
		maxDatasetSizeBytes:           0,
		defaultDataSandbox:            "s3://bucket-a",
		store:                         dataStore,
		shardSelector:                 ioutils.NewConstantShardSelector([]string{"x"}),
		enqueueWorkflow:               func(workflowID v1alpha1.WorkflowID) {},
	}
	p := parentInfo{}
	execContext := executors.NewExecutionContext(w1, w1, w1, p, nil)
	nodeExecContext, err := nodeExecutor.newNodeExecContextDefault(context.Background(), "node-a", execContext, nodeLookup)
	assert.NoError(t, err)
	assert.Equal(t, "s3://bucket-a", nodeExecContext.rawOutputPrefix.String())

	w1.RawOutputDataConfig.OutputLocationPrefix = "s3://bucket-b"
	nodeExecContext, err = nodeExecutor.newNodeExecContextDefault(context.Background(), "node-a", execContext, nodeLookup)
	assert.NoError(t, err)
	assert.Equal(t, "s3://bucket-b", nodeExecContext.rawOutputPrefix.String())
}
