package get

import (
	"fmt"
	"testing"
	"time"

	u "github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"

	"github.com/disiqueira/gotree"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
		expectedRoot := gotree.New("")
		treeRoot := createNodeDetailsTreeView(nil, nil)
		assert.Equal(t, expectedRoot, treeRoot)
	})

	t.Run("successful simple node execution full view", func(t *testing.T) {

		nodeExec1 := createDummyNodeWithID("start-node", false)
		nodeExec1Closure := NodeExecutionClosure{NodeExec: &NodeExecution{nodeExec1}}
		taskExec11 := createDummyTaskExecutionForNode("start-node", "task11")
		taskExec11Closure := TaskExecutionClosure{&TaskExecution{taskExec11}}
		taskExec12 := createDummyTaskExecutionForNode("start-node", "task12")
		taskExec12Closure := TaskExecutionClosure{&TaskExecution{taskExec12}}

		nodeExec1Closure.TaskExecutions = []*TaskExecutionClosure{&taskExec11Closure, &taskExec12Closure}

		nodeExec2 := createDummyNodeWithID("n0", false)
		nodeExec2Closure := NodeExecutionClosure{NodeExec: &NodeExecution{nodeExec2}}
		taskExec21 := createDummyTaskExecutionForNode("n0", "task21")
		taskExec21Closure := TaskExecutionClosure{&TaskExecution{taskExec21}}
		taskExec22 := createDummyTaskExecutionForNode("n0", "task22")
		taskExec22Closure := TaskExecutionClosure{&TaskExecution{taskExec22}}

		nodeExec2Closure.TaskExecutions = []*TaskExecutionClosure{&taskExec21Closure, &taskExec22Closure}

		wrapperNodeExecutions := []*NodeExecutionClosure{&nodeExec1Closure, &nodeExec2Closure}

		treeRoot := createNodeDetailsTreeView(nil, wrapperNodeExecutions)

		assert.Equal(t, 2, len(treeRoot.Items()))
	})
}

func createDummyNodeWithID(nodeID string, isParentNode bool) *admin.NodeExecution {
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
		Metadata: &admin.NodeExecutionMetaData{
			IsParentNode: isParentNode,
		},
		Closure: &admin.NodeExecutionClosure{
			OutputResult: &admin.NodeExecutionClosure_OutputUri{
				OutputUri: nodeID + "outputUri",
			},
			Phase:     core.NodeExecution_SUCCEEDED,
			StartedAt: timestamppb.Now(),
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

		nodeExecStart := createDummyNodeWithID("start-node", false)
		nodeExecN2 := createDummyNodeWithID("n2", true)
		nodeExec1 := createDummyNodeWithID("n0", false)
		taskExec1 := createDummyTaskExecutionForNode("n0", "task21")
		taskExec2 := createDummyTaskExecutionForNode("n0", "task22")

		nodeExecutions := []*admin.NodeExecution{nodeExecStart, nodeExecN2, nodeExec1}
		nodeExecList := &admin.NodeExecutionList{NodeExecutions: nodeExecutions}

		inputs := map[string]*core.Literal{
			"val1": &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Primitive{
							Primitive: &core.Primitive{
								Value: &core.Primitive_Integer{
									Integer: 100,
								},
							},
						},
					},
				},
			},
		}
		outputs := map[string]*core.Literal{
			"o2": &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Primitive{
							Primitive: &core.Primitive{
								Value: &core.Primitive_Integer{
									Integer: 120,
								},
							},
						},
					},
				},
			},
		}
		dataResp := &admin.NodeExecutionGetDataResponse{
			FullOutputs: &core.LiteralMap{
				Literals: inputs,
			},
			FullInputs: &core.LiteralMap{
				Literals: outputs,
			},
		}

		mockFetcherExt.OnFetchNodeExecutionDetailsMatch(ctx, dummyExec, dummyProject, dummyDomain, "").Return(nodeExecList, nil)
		mockFetcherExt.OnFetchNodeExecutionDetailsMatch(ctx, dummyExec, dummyProject, dummyDomain, "n2").Return(&admin.NodeExecutionList{}, nil)
		mockFetcherExt.OnFetchTaskExecutionsOnNodeMatch(ctx, mock.Anything, dummyExec, dummyProject, dummyDomain).Return(&admin.TaskExecutionList{
			TaskExecutions: []*admin.TaskExecution{taskExec1, taskExec2},
		}, nil)
		mockFetcherExt.OnFetchNodeExecutionDataMatch(ctx, mock.Anything, dummyExec, dummyProject, dummyDomain).Return(dataResp, nil)

		nodeExecWrappers, err := getExecutionDetails(ctx, dummyProject, dummyDomain, dummyExec, "", mockCmdCtx)
		assert.Nil(t, err)
		assert.NotNil(t, nodeExecWrappers)
	})

	t.Run("successful get details default view for node-id", func(t *testing.T) {
		setup()
		ctx := u.Ctx
		mockCmdCtx := u.CmdCtx
		mockFetcherExt := u.FetcherExt

		nodeExec1 := createDummyNodeWithID("n0", false)
		taskExec1 := createDummyTaskExecutionForNode("n0", "task21")
		taskExec2 := createDummyTaskExecutionForNode("n0", "task22")

		nodeExecutions := []*admin.NodeExecution{nodeExec1}
		nodeExecList := &admin.NodeExecutionList{NodeExecutions: nodeExecutions}

		inputs := map[string]*core.Literal{
			"val1": &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Primitive{
							Primitive: &core.Primitive{
								Value: &core.Primitive_Integer{
									Integer: 100,
								},
							},
						},
					},
				},
			},
		}
		outputs := map[string]*core.Literal{
			"o2": &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Primitive{
							Primitive: &core.Primitive{
								Value: &core.Primitive_Integer{
									Integer: 120,
								},
							},
						},
					},
				},
			},
		}
		dataResp := &admin.NodeExecutionGetDataResponse{
			FullOutputs: &core.LiteralMap{
				Literals: inputs,
			},
			FullInputs: &core.LiteralMap{
				Literals: outputs,
			},
		}

		mockFetcherExt.OnFetchNodeExecutionDetailsMatch(ctx, dummyExec, dummyProject, dummyDomain, "").Return(nodeExecList, nil)
		mockFetcherExt.OnFetchTaskExecutionsOnNodeMatch(ctx, "n0", dummyExec, dummyProject, dummyDomain).Return(&admin.TaskExecutionList{
			TaskExecutions: []*admin.TaskExecution{taskExec1, taskExec2},
		}, nil)
		mockFetcherExt.OnFetchNodeExecutionDataMatch(ctx, mock.Anything, dummyExec, dummyProject, dummyDomain).Return(dataResp, nil)

		nodeExecWrappers, err := getExecutionDetails(ctx, dummyProject, dummyDomain, dummyExec, "n0", mockCmdCtx)
		assert.Nil(t, err)
		assert.NotNil(t, nodeExecWrappers)
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

		mockFetcherExt.OnFetchNodeExecutionDetailsMatch(ctx, dummyExec, dummyProject, dummyDomain, "").Return(nodeExecList, nil)
		mockFetcherExt.OnFetchTaskExecutionsOnNodeMatch(ctx, "n0", dummyExec, dummyProject, dummyDomain).Return(nil, fmt.Errorf("unable to fetch task exec details"))
		_, err = getExecutionDetails(ctx, dummyProject, dummyDomain, dummyExec, "", mockCmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("unable to fetch task exec details"), err)
	})
}

func TestExtractLiteralMapError(t *testing.T) {
	literalMap, err := extractLiteralMap(nil)
	assert.Nil(t, err)
	assert.Equal(t, len(literalMap), 0)

	literalMap, err = extractLiteralMap(&core.LiteralMap{})
	assert.Nil(t, err)
	assert.Equal(t, len(literalMap), 0)
}
