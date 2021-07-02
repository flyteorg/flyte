package get

import (
	"fmt"
	"testing"
	"time"

	"github.com/flyteorg/flytectl/cmd/config"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/execution"
	u "github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"

	"github.com/disiqueira/gotree"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	dummyProject = "dummyProject"
	dummyDomain  = "dummyDomain"
	dummyExec    = "dummyExec"
)

func TestCreateNodeDetailsTreeView(t *testing.T) {

	t.Run("empty node execution", func(t *testing.T) {
		var nodeExecutions []*admin.NodeExecution
		var nodeExecToTaskExec map[string]*admin.TaskExecutionList
		expectedRoot := gotree.New("")
		treeRoot := createNodeDetailsTreeView(nodeExecutions, nodeExecToTaskExec)
		assert.Equal(t, expectedRoot, treeRoot)
	})

	t.Run("successful simple node execution full view", func(t *testing.T) {
		nodeExecToTaskExec := map[string]*admin.TaskExecutionList{}

		nodeExec1 := createDummyNodeWithID("start-node", true)
		taskExec11 := createDummyTaskExecutionForNode("start-node", "task11")
		taskExec12 := createDummyTaskExecutionForNode("start-node", "task12")

		nodeExecToTaskExec["start-node"] = &admin.TaskExecutionList{
			TaskExecutions: []*admin.TaskExecution{taskExec11, taskExec12},
		}

		nodeExec2 := createDummyNodeWithID("n0", false)
		taskExec21 := createDummyTaskExecutionForNode("n0", "task21")
		taskExec22 := createDummyTaskExecutionForNode("n0", "task22")

		nodeExecToTaskExec["n0"] = &admin.TaskExecutionList{
			TaskExecutions: []*admin.TaskExecution{taskExec21, taskExec22},
		}

		nodeExec3 := createDummyNodeWithID("n1", false)
		taskExec31 := createDummyTaskExecutionForNode("n1", "task31")
		taskExec32 := createDummyTaskExecutionForNode("n1", "task32")

		nodeExecToTaskExec["n0"] = &admin.TaskExecutionList{
			TaskExecutions: []*admin.TaskExecution{taskExec31, taskExec32},
		}

		nodeExecutions := []*admin.NodeExecution{nodeExec1, nodeExec2, nodeExec3}

		treeRoot := createNodeDetailsTreeView(nodeExecutions, nodeExecToTaskExec)

		assert.Equal(t, 3, len(treeRoot.Items()))
	})

	t.Run("empty task execution only view", func(t *testing.T) {
		taskExecutionList := &admin.TaskExecutionList{}
		treeRoot := createNodeTaskExecTreeView(nil, taskExecutionList)
		assert.Equal(t, 0, len(treeRoot.Items()))
	})

	t.Run("successful task execution only view", func(t *testing.T) {
		taskExec31 := createDummyTaskExecutionForNode("n1", "task31")
		taskExec32 := createDummyTaskExecutionForNode("n1", "task32")

		taskExecutionList := &admin.TaskExecutionList{
			TaskExecutions: []*admin.TaskExecution{taskExec31, taskExec32},
		}

		treeRoot := createNodeTaskExecTreeView(nil, taskExecutionList)
		assert.Equal(t, 2, len(treeRoot.Items()))
	})
}

func createDummyNodeWithID(nodeID string, startedAtEmpty bool) *admin.NodeExecution {
	// Remove this param  startedAtEmpty and code once admin code is fixed
	startedAt := timestamppb.Now()
	if startedAtEmpty {
		startedAt = nil
	}

	nodeExecution := &admin.NodeExecution{
		Id: &core.NodeExecutionIdentifier{
			NodeId: nodeID,
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: dummyProject,
				Domain:  dummyDomain,
				Name:    dummyExec,
			},
		},
		InputUri: nodeID + "inputUri",
		Closure: &admin.NodeExecutionClosure{
			OutputResult: &admin.NodeExecutionClosure_OutputUri{
				OutputUri: nodeID + "outputUri",
			},
			Phase:     core.NodeExecution_SUCCEEDED,
			StartedAt: startedAt,
			Duration:  &durationpb.Duration{Seconds: 100},
			CreatedAt: timestamppb.Now(),
			UpdatedAt: timestamppb.Now(),
			TargetMetadata: &admin.NodeExecutionClosure_WorkflowNodeMetadata{
				WorkflowNodeMetadata: &admin.WorkflowNodeMetadata{
					ExecutionId: &core.WorkflowExecutionIdentifier{
						Project: dummyProject,
						Domain:  dummyDomain,
						Name:    dummyExec,
					},
				},
			},
		},
	}
	return nodeExecution
}

func createDummyTaskExecutionForNode(nodeID string, taskID string) *admin.TaskExecution {
	taskLog1 := &core.TaskLog{
		Uri:           nodeID + taskID + "logUri1",
		Name:          nodeID + taskID + "logName1",
		MessageFormat: core.TaskLog_JSON,
		Ttl:           &durationpb.Duration{Seconds: 100},
	}

	taskLog2 := &core.TaskLog{
		Uri:           nodeID + taskID + "logUri2",
		Name:          nodeID + taskID + "logName2",
		MessageFormat: core.TaskLog_JSON,
		Ttl:           &durationpb.Duration{Seconds: 100},
	}

	taskLogs := []*core.TaskLog{taskLog1, taskLog2}

	extResourceInfo := &event.ExternalResourceInfo{
		ExternalId: nodeID + taskID + "externalId",
	}
	extResourceInfos := []*event.ExternalResourceInfo{extResourceInfo}

	resourcePoolInfo := &event.ResourcePoolInfo{
		AllocationToken: nodeID + taskID + "allocationToken",
		Namespace:       nodeID + taskID + "namespace",
	}
	resourcePoolInfos := []*event.ResourcePoolInfo{resourcePoolInfo}

	taskExec := &admin.TaskExecution{
		Id: &core.TaskExecutionIdentifier{
			TaskId: &core.Identifier{
				Project:      dummyProject,
				Domain:       dummyDomain,
				Name:         dummyExec,
				ResourceType: core.ResourceType_TASK,
			},
		},
		InputUri: nodeID + taskID + "inputUrlForTask",
		Closure: &admin.TaskExecutionClosure{
			OutputResult: &admin.TaskExecutionClosure_OutputUri{
				OutputUri: nodeID + taskID + "outputUri-task",
			},
			Phase:     core.TaskExecution_SUCCEEDED,
			Logs:      taskLogs,
			StartedAt: timestamppb.Now(),
			Duration:  &durationpb.Duration{Seconds: 100},
			CreatedAt: timestamppb.Now(),
			UpdatedAt: timestamppb.New(time.Now()),
			Reason:    nodeID + taskID + "reason",
			TaskType:  nodeID + taskID + "taskType",
			Metadata: &event.TaskExecutionMetadata{
				GeneratedName:     nodeID + taskID + "generatedName",
				ExternalResources: extResourceInfos,
				ResourcePoolInfo:  resourcePoolInfos,
				PluginIdentifier:  nodeID + taskID + "pluginId",
			},
		},
	}
	return taskExec
}

func TestGetExecutionDetails(t *testing.T) {
	t.Run("successful get details default view", func(t *testing.T) {
		setup()
		ctx := u.Ctx
		mockCmdCtx := u.CmdCtx
		mockFetcherExt := u.FetcherExt
		nodeExecToTaskExec := map[string]*admin.TaskExecutionList{}

		nodeExec1 := createDummyNodeWithID("n0", false)
		taskExec1 := createDummyTaskExecutionForNode("n0", "task21")
		taskExec2 := createDummyTaskExecutionForNode("n0", "task22")

		nodeExecToTaskExec["n0"] = &admin.TaskExecutionList{
			TaskExecutions: []*admin.TaskExecution{taskExec1, taskExec2},
		}

		nodeExecutions := []*admin.NodeExecution{nodeExec1}
		nodeExecList := &admin.NodeExecutionList{NodeExecutions: nodeExecutions}

		mockFetcherExt.OnFetchNodeExecutionDetailsMatch(ctx, dummyExec, dummyProject, dummyDomain).Return(nodeExecList, nil)
		mockFetcherExt.OnFetchTaskExecutionsOnNodeMatch(ctx, "n0", dummyExec, dummyProject, dummyDomain).Return(nodeExecToTaskExec["n0"], nil)

		err = getExecutionDetails(ctx, dummyProject, dummyDomain, dummyExec, mockCmdCtx)
		assert.Nil(t, err)
	})

	t.Run("failure node details fetch", func(t *testing.T) {
		setup()
		ctx := u.Ctx
		mockCmdCtx := u.CmdCtx
		mockFetcherExt := u.FetcherExt
		mockFetcherExt.OnFetchNodeExecutionDetailsMatch(ctx, dummyExec, dummyProject, dummyDomain).Return(nil, fmt.Errorf("unable to fetch details"))
		err = getExecutionDetails(ctx, dummyProject, dummyDomain, dummyExec, mockCmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("unable to fetch details"), err)
	})

	t.Run("failure task exec fetch", func(t *testing.T) {
		setup()
		ctx := u.Ctx
		mockCmdCtx := u.CmdCtx
		mockFetcherExt := u.FetcherExt
		nodeExecToTaskExec := map[string]*admin.TaskExecutionList{}

		nodeExec1 := createDummyNodeWithID("n0", false)
		taskExec1 := createDummyTaskExecutionForNode("n0", "task21")
		taskExec2 := createDummyTaskExecutionForNode("n0", "task22")

		nodeExecToTaskExec["n0"] = &admin.TaskExecutionList{
			TaskExecutions: []*admin.TaskExecution{taskExec1, taskExec2},
		}

		nodeExecutions := []*admin.NodeExecution{nodeExec1}
		nodeExecList := &admin.NodeExecutionList{NodeExecutions: nodeExecutions}

		mockFetcherExt.OnFetchNodeExecutionDetailsMatch(ctx, dummyExec, dummyProject, dummyDomain).Return(nodeExecList, nil)
		mockFetcherExt.OnFetchTaskExecutionsOnNodeMatch(ctx, "n0", dummyExec, dummyProject, dummyDomain).Return(nil, fmt.Errorf("unable to fetch task exec details"))
		err = getExecutionDetails(ctx, dummyProject, dummyDomain, dummyExec, mockCmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("unable to fetch task exec details"), err)
	})

	t.Run("successful get details non default view", func(t *testing.T) {
		setup()
		config.GetConfig().Output = "table"
		execution.DefaultConfig.NodeID = "n0"

		ctx := u.Ctx
		mockCmdCtx := u.CmdCtx
		mockFetcherExt := u.FetcherExt
		nodeExecToTaskExec := map[string]*admin.TaskExecutionList{}

		nodeExec1 := createDummyNodeWithID("n0", false)
		taskExec1 := createDummyTaskExecutionForNode("n0", "task21")
		taskExec2 := createDummyTaskExecutionForNode("n0", "task22")

		nodeExecToTaskExec["n0"] = &admin.TaskExecutionList{
			TaskExecutions: []*admin.TaskExecution{taskExec1, taskExec2},
		}

		nodeExecutions := []*admin.NodeExecution{nodeExec1}
		nodeExecList := &admin.NodeExecutionList{NodeExecutions: nodeExecutions}

		mockFetcherExt.OnFetchNodeExecutionDetailsMatch(ctx, dummyExec, dummyProject, dummyDomain).Return(nodeExecList, nil)
		mockFetcherExt.OnFetchTaskExecutionsOnNodeMatch(ctx, "n0", dummyExec, dummyProject, dummyDomain).Return(nodeExecToTaskExec["n0"], nil)

		err = getExecutionDetails(ctx, dummyProject, dummyDomain, dummyExec, mockCmdCtx)
		assert.Nil(t, err)
	})

	t.Run("Table test successful cases", func(t *testing.T) {
		tests := []struct {
			outputFormat string
			nodeID       string
			want         error
		}{
			{outputFormat: "table", nodeID: "", want: nil},
			{outputFormat: "table", nodeID: "n0", want: nil},
			{outputFormat: "yaml", nodeID: "", want: nil},
			{outputFormat: "yaml", nodeID: "n0", want: nil},
			{outputFormat: "yaml", nodeID: "n1", want: nil},
		}

		for _, tt := range tests {
			setup()
			config.GetConfig().Output = tt.outputFormat
			execution.DefaultConfig.NodeID = tt.nodeID

			ctx := u.Ctx
			mockCmdCtx := u.CmdCtx
			mockFetcherExt := u.FetcherExt
			nodeExecToTaskExec := map[string]*admin.TaskExecutionList{}

			nodeExec1 := createDummyNodeWithID("n0", false)
			taskExec1 := createDummyTaskExecutionForNode("n0", "task21")
			taskExec2 := createDummyTaskExecutionForNode("n0", "task22")

			nodeExecToTaskExec["n0"] = &admin.TaskExecutionList{
				TaskExecutions: []*admin.TaskExecution{taskExec1, taskExec2},
			}

			nodeExecutions := []*admin.NodeExecution{nodeExec1}
			nodeExecList := &admin.NodeExecutionList{NodeExecutions: nodeExecutions}

			mockFetcherExt.OnFetchNodeExecutionDetailsMatch(ctx, dummyExec, dummyProject, dummyDomain).Return(nodeExecList, nil)
			mockFetcherExt.OnFetchTaskExecutionsOnNodeMatch(ctx, "n0", dummyExec, dummyProject, dummyDomain).Return(nodeExecToTaskExec["n0"], nil)

			got := getExecutionDetails(ctx, dummyProject, dummyDomain, dummyExec, mockCmdCtx)
			assert.Equal(t, tt.want, got)
		}
	})
}
