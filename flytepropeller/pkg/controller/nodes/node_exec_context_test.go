package nodes

import (
	"context"
	"testing"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/storage"
	"github.com/stretchr/testify/assert"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
)

type TaskReader struct{}

func (t TaskReader) Read(ctx context.Context) (*core.TaskTemplate, error) { return nil, nil }
func (t TaskReader) GetTaskType() v1alpha1.TaskType                       { return "" }
func (t TaskReader) GetTaskID() *core.Identifier {
	return &core.Identifier{Project: "p", Domain: "d", Name: "task-name"}
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
	nCtx := newNodeExecContext(context.TODO(), s, w1, n, nil, nil, false, 0, nil, TaskReader{}, nil, nil, "s3://bucket", ioutils.NewConstantShardSelector([]string{"x"}))
	assert.Equal(t, "id", nCtx.NodeExecutionMetadata().GetLabels()["node-id"])
	assert.Equal(t, "false", nCtx.NodeExecutionMetadata().GetLabels()["interruptible"])
	assert.Equal(t, "task-name", nCtx.NodeExecutionMetadata().GetLabels()["task-name"])
}
