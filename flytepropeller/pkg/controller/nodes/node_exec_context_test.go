package nodes

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"

	"github.com/flyteorg/flytepropeller/events"
	eventsErr "github.com/flyteorg/flytepropeller/events/errors"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
	"github.com/flyteorg/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flytepropeller/pkg/controller/executors"
	mocks2 "github.com/flyteorg/flytepropeller/pkg/controller/executors/mocks"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/ioutils"

	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flytestdlib/storage"

	"github.com/stretchr/testify/assert"
)

type TaskReader struct{}

func (t TaskReader) Read(ctx context.Context) (*core.TaskTemplate, error) { return nil, nil }
func (t TaskReader) GetTaskType() v1alpha1.TaskType                       { return "" }
func (t TaskReader) GetTaskID() *core.Identifier {
	return &core.Identifier{Project: "p", Domain: "d", Name: "task-name"}
}

type fakeEventRecorder struct {
	nodeErr error
	taskErr error
}

func (f fakeEventRecorder) RecordNodeEvent(ctx context.Context, event *event.NodeExecutionEvent, eventConfig *config.EventConfig) error {
	return f.nodeErr
}

func (f fakeEventRecorder) RecordTaskEvent(ctx context.Context, event *event.TaskExecutionEvent, eventConfig *config.EventConfig) error {
	return f.taskErr
}

type parentInfo struct {
	executors.ImmutableParentInfo
}

func getTestNodeSpec(interruptible *bool) *v1alpha1.NodeSpec {
	taskID := "taskID"
	return &v1alpha1.NodeSpec{
		ID:            "id",
		TaskRef:       &taskID,
		Kind:          v1alpha1.NodeKindTask,
		Interruptible: interruptible,
	}
}

func getTestFlyteWorkflow() *v1alpha1.FlyteWorkflow {
	interruptible := false
	return &v1alpha1.FlyteWorkflow{
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
		ExecutionConfig: v1alpha1.ExecutionConfig{Interruptible: &interruptible},
	}
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

	s, _ := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	p := parentInfo{}
	execContext := executors.NewExecutionContext(w1, nil, nil, p, nil)
	nCtx := newNodeExecContext(context.TODO(), s, execContext, w1, getTestNodeSpec(nil), nil, nil, false, 0, 2, nil, nil, TaskReader{}, nil, nil, "s3://bucket", ioutils.NewConstantShardSelector([]string{"x"}))
	assert.Equal(t, "id", nCtx.NodeExecutionMetadata().GetLabels()["node-id"])
	assert.Equal(t, "false", nCtx.NodeExecutionMetadata().GetLabels()["interruptible"])
	assert.Equal(t, "task-name", nCtx.NodeExecutionMetadata().GetLabels()["task-name"])
	assert.Equal(t, p, nCtx.ExecutionContext().GetParentInfo())
}

func Test_NodeContextDefault(t *testing.T) {
	ctx := context.Background()

	w1 := getTestFlyteWorkflow()
	dataStore, _ := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	nodeLookup := &mocks2.NodeLookup{}
	nodeLookup.OnGetNode("node-a").Return(getTestNodeSpec(nil), true)
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
	nodeExecContext, err := nodeExecutor.BuildNodeExecutionContext(context.Background(), execContext, nodeLookup, "node-a")
	assert.NoError(t, err)
	assert.Equal(t, "s3://bucket-a", nodeExecContext.RawOutputPrefix().String())

	w1.RawOutputDataConfig.OutputLocationPrefix = "s3://bucket-b"
	nodeExecContext, err = nodeExecutor.BuildNodeExecutionContext(context.Background(), execContext, nodeLookup, "node-a")
	assert.NoError(t, err)
	assert.Equal(t, "s3://bucket-b", nodeExecContext.RawOutputPrefix().String())
}

func Test_NodeContextDefaultInterruptible(t *testing.T) {
	ctx := context.Background()
	scope := promutils.NewTestScope()

	dataStore, _ := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, scope.NewSubScope("dataStore"))
	nodeExecutor := nodeExecutor{
		interruptibleFailureThreshold: 10,
		maxDatasetSizeBytes:           0,
		defaultDataSandbox:            "s3://bucket-a",
		store:                         dataStore,
		shardSelector:                 ioutils.NewConstantShardSelector([]string{"x"}),
		enqueueWorkflow:               func(workflowID v1alpha1.WorkflowID) {},
		metrics: &nodeMetrics{
			InterruptibleNodesRunning:    labeled.NewCounter("running", "xyz", scope.NewSubScope("interruptible1")),
			InterruptibleNodesTerminated: labeled.NewCounter("terminated", "xyz", scope.NewSubScope("interruptible2")),
			InterruptedThresholdHit:      labeled.NewCounter("thresholdHit", "xyz", scope.NewSubScope("interruptible3")),
		},
	}

	verifyNodeExecContext := func(t *testing.T, executionContext executors.ExecutionContext, nl executors.NodeLookup, shouldBeInterruptible bool) {
		nodeExecContext, err := nodeExecutor.BuildNodeExecutionContext(context.Background(), executionContext, nl, "node-a")
		assert.NoError(t, err)
		assert.Equal(t, shouldBeInterruptible, nodeExecContext.NodeExecutionMetadata().IsInterruptible())
		labels := nodeExecContext.NodeExecutionMetadata().GetLabels()
		assert.Contains(t, labels, NodeInterruptibleLabel)
		assert.Equal(t, strconv.FormatBool(shouldBeInterruptible), labels[NodeInterruptibleLabel])
	}

	t.Run("NodeSpec interruptible nil", func(t *testing.T) {
		w := getTestFlyteWorkflow()

		nodeLookup := &mocks2.NodeLookup{}
		nodeLookup.OnGetNode("node-a").Return(getTestNodeSpec(nil), true)
		nodeLookup.OnGetNodeExecutionStatus(ctx, "node-a").Return(&v1alpha1.NodeStatus{
			SystemFailures: 0,
		})

		p := parentInfo{}
		execContext := executors.NewExecutionContext(w, w, w, p, nil)

		// node spec, exec config and node defaults have no interruptible flag -> false
		verifyNodeExecContext(t, execContext, nodeLookup, false)

		// both exec config and node defaults have interruptible flag, node spec defines no override -> true
		execConfigInterruptible := true
		w.ExecutionConfig.Interruptible = &execConfigInterruptible
		w.NodeDefaults.Interruptible = true
		verifyNodeExecContext(t, execContext, nodeLookup, true)

		// node defaults set interruptible flag, but exec config overwrites it -> false
		execConfigInterruptible = false
		w.NodeDefaults.Interruptible = true
		verifyNodeExecContext(t, execContext, nodeLookup, false)

		// node defaults do not have interruptible flags, but exec config enables it -> true
		execConfigInterruptible = true
		w.NodeDefaults.Interruptible = false
		verifyNodeExecContext(t, execContext, nodeLookup, true)

		// exec config does not specify interruptible flag, node defaults contain false, node spec defines no override -> false
		w.ExecutionConfig.Interruptible = nil
		w.NodeDefaults.Interruptible = false
		verifyNodeExecContext(t, execContext, nodeLookup, false)

		// exec config does not specify interruptible flag, node defaults contain true, node spec defines no override -> true
		w.ExecutionConfig.Interruptible = nil
		w.NodeDefaults.Interruptible = true
		verifyNodeExecContext(t, execContext, nodeLookup, true)
	})

	t.Run("NodeSpec interruptible true", func(t *testing.T) {
		w := getTestFlyteWorkflow()

		interruptible := true
		nodeLookup := &mocks2.NodeLookup{}
		nodeLookup.OnGetNode("node-a").Return(getTestNodeSpec(&interruptible), true)
		nodeLookup.OnGetNodeExecutionStatus(ctx, "node-a").Return(&v1alpha1.NodeStatus{
			SystemFailures: 0,
		})

		p := parentInfo{}
		execContext := executors.NewExecutionContext(w, w, w, p, nil)

		// exec config and node defaults have no interruptible flag, node spec defines true -> true
		verifyNodeExecContext(t, execContext, nodeLookup, true)

		// both exec config and node defaults have interruptible flag, node spec defines true -> true
		execConfigInterruptible := true
		w.ExecutionConfig.Interruptible = &execConfigInterruptible
		w.NodeDefaults.Interruptible = true
		verifyNodeExecContext(t, execContext, nodeLookup, true)

		// node defaults set interruptible flag, exec config overwrites it, but node spec defines true -> true
		execConfigInterruptible = false
		w.NodeDefaults.Interruptible = true
		verifyNodeExecContext(t, execContext, nodeLookup, true)

		// node defaults do not have interruptible flags, but exec config enables it -> true
		execConfigInterruptible = true
		w.NodeDefaults.Interruptible = false
		verifyNodeExecContext(t, execContext, nodeLookup, true)

		// exec config does not specify interruptible flag, node defaults contain false, node spec defines true -> true
		w.ExecutionConfig.Interruptible = nil
		w.NodeDefaults.Interruptible = false
		verifyNodeExecContext(t, execContext, nodeLookup, true)

		// exec config does not specify interruptible flag, node defaults contain true, node spec defines true -> true
		w.ExecutionConfig.Interruptible = nil
		w.NodeDefaults.Interruptible = true
		verifyNodeExecContext(t, execContext, nodeLookup, true)
	})

	t.Run("NodeSpec interruptible false", func(t *testing.T) {
		w := getTestFlyteWorkflow()

		interruptible := false
		nodeLookup := &mocks2.NodeLookup{}
		nodeLookup.OnGetNode("node-a").Return(getTestNodeSpec(&interruptible), true)
		nodeLookup.OnGetNodeExecutionStatus(ctx, "node-a").Return(&v1alpha1.NodeStatus{
			SystemFailures: 0,
		})

		p := parentInfo{}
		execContext := executors.NewExecutionContext(w, w, w, p, nil)

		// exec config and node defaults have no interruptible flag, node spec defines false -> false
		verifyNodeExecContext(t, execContext, nodeLookup, false)

		// both exec config and node defaults have interruptible flag, node spec defines false -> true
		execConfigInterruptible := true
		w.ExecutionConfig.Interruptible = &execConfigInterruptible
		w.NodeDefaults.Interruptible = true
		verifyNodeExecContext(t, execContext, nodeLookup, false)

		// node defaults set interruptible flag, exec config overwrites it, node spec defines false -> false
		execConfigInterruptible = false
		w.NodeDefaults.Interruptible = true
		verifyNodeExecContext(t, execContext, nodeLookup, false)

		// node defaults do not have interruptible flags, exec config enables it, but node spec defines false -> false
		execConfigInterruptible = true
		w.NodeDefaults.Interruptible = false
		verifyNodeExecContext(t, execContext, nodeLookup, false)

		// exec config does not specify interruptible flag, node defaults contain false, node spec defines false -> false
		w.ExecutionConfig.Interruptible = nil
		w.NodeDefaults.Interruptible = false
		verifyNodeExecContext(t, execContext, nodeLookup, false)

		// exec config does not specify interruptible flag, node defaults contain true, node spec defines false -> false
		w.ExecutionConfig.Interruptible = nil
		w.NodeDefaults.Interruptible = true
		verifyNodeExecContext(t, execContext, nodeLookup, false)
	})
}

func Test_NodeContext_RecordNodeEvent(t *testing.T) {
	noErrRecorder := fakeEventRecorder{}
	alreadyExistsError := fakeEventRecorder{nodeErr: &eventsErr.EventError{Code: eventsErr.AlreadyExists, Cause: fmt.Errorf("err")}}
	inTerminalError := fakeEventRecorder{nodeErr: &eventsErr.EventError{Code: eventsErr.EventAlreadyInTerminalStateError, Cause: fmt.Errorf("err")}}
	otherError := fakeEventRecorder{nodeErr: &eventsErr.EventError{Code: eventsErr.ResourceExhausted, Cause: fmt.Errorf("err")}}

	tests := []struct {
		name    string
		rec     events.NodeEventRecorder
		p       core.NodeExecution_Phase
		wantErr bool
	}{
		{"aborted-success", noErrRecorder, core.NodeExecution_ABORTED, false},
		{"aborted-failure", otherError, core.NodeExecution_ABORTED, true},
		{"aborted-already", alreadyExistsError, core.NodeExecution_ABORTED, false},
		{"aborted-terminal", inTerminalError, core.NodeExecution_ABORTED, false},
		{"running-terminal", inTerminalError, core.NodeExecution_RUNNING, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventRecorder := &eventRecorder{
				nodeEventRecorder: tt.rec,
			}

			ev := &event.NodeExecutionEvent{
				Id:         &core.NodeExecutionIdentifier{},
				Phase:      tt.p,
				ProducerId: "propeller",
			}
			if err := eventRecorder.RecordNodeEvent(context.TODO(), ev, &config.EventConfig{}); (err != nil) != tt.wantErr {
				t.Errorf("RecordNodeEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_NodeContext_RecordTaskEvent(t1 *testing.T) {
	noErrRecorder := fakeEventRecorder{}
	alreadyExistsError := fakeEventRecorder{taskErr: &eventsErr.EventError{Code: eventsErr.AlreadyExists, Cause: fmt.Errorf("err")}}
	inTerminalError := fakeEventRecorder{taskErr: &eventsErr.EventError{Code: eventsErr.EventAlreadyInTerminalStateError, Cause: fmt.Errorf("err")}}
	otherError := fakeEventRecorder{taskErr: &eventsErr.EventError{Code: eventsErr.ResourceExhausted, Cause: fmt.Errorf("err")}}

	tests := []struct {
		name    string
		rec     events.TaskEventRecorder
		p       core.TaskExecution_Phase
		wantErr bool
	}{
		{"aborted-success", noErrRecorder, core.TaskExecution_ABORTED, false},
		{"aborted-failure", otherError, core.TaskExecution_ABORTED, true},
		{"aborted-already", alreadyExistsError, core.TaskExecution_ABORTED, false},
		{"aborted-terminal", inTerminalError, core.TaskExecution_ABORTED, false},
		{"running-terminal", inTerminalError, core.TaskExecution_RUNNING, true},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &eventRecorder{
				taskEventRecorder: tt.rec,
			}
			ev := &event.TaskExecutionEvent{
				Phase: tt.p,
			}
			if err := t.RecordTaskEvent(context.TODO(), ev, &config.EventConfig{
				RawOutputPolicy: config.RawOutputPolicyReference,
			}); (err != nil) != tt.wantErr {
				t1.Errorf("RecordTaskEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
