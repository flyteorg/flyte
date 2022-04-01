package get

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/flyteorg/flytectl/cmd/testutils"

	"github.com/flyteorg/flytectl/cmd/config"
	"github.com/flyteorg/flytectl/cmd/config/subcommand/execution"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	nodeID = "node-id"
)

func getExecutionSetup() {
	config.GetConfig().Project = projectValue
	config.GetConfig().Domain = domainValue
	config.GetConfig().Output = output
	execution.DefaultConfig.Details = false
	execution.DefaultConfig.NodeID = ""
}

func TestListExecutionFunc(t *testing.T) {
	getExecutionSetup()
	s := setup()
	ctx := s.Ctx
	execListRequest := &admin.ResourceListRequest{
		Limit: 100,
		SortBy: &admin.Sort{
			Key:       "created_at",
			Direction: admin.Sort_DESCENDING,
		},
		Id: &admin.NamedEntityIdentifier{
			Project: projectValue,
			Domain:  domainValue,
		},
	}
	executionResponse := &admin.Execution{
		Id: &core.WorkflowExecutionIdentifier{
			Project: projectValue,
			Domain:  domainValue,
			Name:    executionNameValue,
		},
		Spec: &admin.ExecutionSpec{
			LaunchPlan: &core.Identifier{
				Project: projectValue,
				Domain:  domainValue,
				Name:    launchPlanNameValue,
				Version: launchPlanVersionValue,
			},
		},
		Closure: &admin.ExecutionClosure{
			WorkflowId: &core.Identifier{
				Project: projectValue,
				Domain:  domainValue,
				Name:    workflowNameValue,
				Version: workflowVersionValue,
			},
			Phase: core.WorkflowExecution_SUCCEEDED,
		},
	}
	executions := []*admin.Execution{executionResponse}
	executionList := &admin.ExecutionList{
		Executions: executions,
	}
	s.MockAdminClient.OnListExecutionsMatch(mock.Anything, execListRequest).Return(executionList, nil)
	err := getExecutionFunc(s.Ctx, []string{}, s.CmdCtx)
	assert.Nil(t, err)
	s.MockAdminClient.AssertCalled(t, "ListExecutions", ctx, execListRequest)
}

func TestListExecutionFuncWithError(t *testing.T) {
	ctx := context.Background()
	getExecutionSetup()
	execListRequest := &admin.ResourceListRequest{
		Limit: 100,
		SortBy: &admin.Sort{
			Key: "created_at",
		},
		Id: &admin.NamedEntityIdentifier{
			Project: projectValue,
			Domain:  domainValue,
		},
	}

	_ = &admin.Execution{
		Id: &core.WorkflowExecutionIdentifier{
			Project: projectValue,
			Domain:  domainValue,
			Name:    executionNameValue,
		},
		Spec: &admin.ExecutionSpec{
			LaunchPlan: &core.Identifier{
				Project: projectValue,
				Domain:  domainValue,
				Name:    launchPlanNameValue,
				Version: launchPlanVersionValue,
			},
		},
		Closure: &admin.ExecutionClosure{
			WorkflowId: &core.Identifier{
				Project: projectValue,
				Domain:  domainValue,
				Name:    workflowNameValue,
				Version: workflowVersionValue,
			},
			Phase: core.WorkflowExecution_SUCCEEDED,
		},
	}
	s := setup()
	s.FetcherExt.OnListExecutionMatch(s.Ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("executions NotFound"))
	s.MockAdminClient.OnListExecutionsMatch(mock.Anything, execListRequest).Return(nil, errors.New("executions NotFound"))
	err := getExecutionFunc(s.Ctx, []string{}, s.CmdCtx)
	assert.NotNil(t, err)
	assert.Equal(t, err, errors.New("executions NotFound"))
	s.MockAdminClient.AssertCalled(t, "ListExecutions", ctx, execListRequest)
}

func TestGetExecutionFunc(t *testing.T) {
	ctx := context.Background()
	getExecutionSetup()
	execGetRequest := &admin.WorkflowExecutionGetRequest{
		Id: &core.WorkflowExecutionIdentifier{
			Project: projectValue,
			Domain:  domainValue,
			Name:    executionNameValue,
		},
	}
	executionResponse := &admin.Execution{
		Id: &core.WorkflowExecutionIdentifier{
			Project: projectValue,
			Domain:  domainValue,
			Name:    executionNameValue,
		},
		Spec: &admin.ExecutionSpec{
			LaunchPlan: &core.Identifier{
				Project: projectValue,
				Domain:  domainValue,
				Name:    launchPlanNameValue,
				Version: launchPlanVersionValue,
			},
		},
		Closure: &admin.ExecutionClosure{
			WorkflowId: &core.Identifier{
				Project: projectValue,
				Domain:  domainValue,
				Name:    workflowNameValue,
				Version: workflowVersionValue,
			},
			Phase: core.WorkflowExecution_SUCCEEDED,
		},
	}
	args := []string{executionNameValue}
	s := setup()
	//executionList := &admin.ExecutionList{
	//	Executions: []*admin.Execution{executionResponse},
	//}
	s.MockAdminClient.OnGetExecutionMatch(ctx, execGetRequest).Return(executionResponse, nil)

	err := getExecutionFunc(s.Ctx, args, s.CmdCtx)
	assert.Nil(t, err)
	s.MockAdminClient.AssertCalled(t, "GetExecution", ctx, execGetRequest)
}

func TestGetExecutionFuncForDetails(t *testing.T) {
	s := testutils.SetupWithExt()
	getExecutionSetup()
	ctx := s.Ctx
	mockCmdCtx := s.CmdCtx
	mockFetcherExt := s.FetcherExt
	execution.DefaultConfig.Details = true
	args := []string{dummyExec}
	mockFetcherExt.OnFetchExecutionMatch(ctx, dummyExec, dummyProject, dummyDomain).Return(&admin.Execution{}, nil)
	mockFetcherExt.OnFetchNodeExecutionDetailsMatch(ctx, dummyExec, dummyProject, dummyDomain, "").Return(nil, fmt.Errorf("unable to fetch details"))
	err := getExecutionFunc(ctx, args, mockCmdCtx)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf("unable to fetch details"), err)
}

func TestGetExecutionFuncWithIOData(t *testing.T) {
	t.Run("successful inputs outputs", func(t *testing.T) {
		s := testutils.SetupWithExt()
		getExecutionSetup()
		ctx := s.Ctx
		mockCmdCtx := s.CmdCtx
		mockFetcherExt := s.FetcherExt
		execution.DefaultConfig.NodeID = nodeID
		args := []string{dummyExec}

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
									Integer: 110,
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
		mockFetcherExt.OnFetchExecutionMatch(ctx, dummyExec, dummyProject, dummyDomain).Return(&admin.Execution{}, nil)
		mockFetcherExt.OnFetchNodeExecutionDetailsMatch(ctx, dummyExec, dummyProject, dummyDomain, "").Return(nodeExecList, nil)
		mockFetcherExt.OnFetchTaskExecutionsOnNodeMatch(ctx, "n0", dummyExec, dummyProject, dummyDomain).Return(&admin.TaskExecutionList{
			TaskExecutions: []*admin.TaskExecution{taskExec1, taskExec2},
		}, nil)
		mockFetcherExt.OnFetchNodeExecutionDataMatch(ctx, mock.Anything, dummyExec, dummyProject, dummyDomain).Return(dataResp, nil)

		err := getExecutionFunc(ctx, args, mockCmdCtx)
		assert.Nil(t, err)
	})
	t.Run("fetch data error from admin", func(t *testing.T) {
		s := testutils.SetupWithExt()
		getExecutionSetup()
		ctx := s.Ctx
		mockCmdCtx := s.CmdCtx
		mockFetcherExt := s.FetcherExt
		execution.DefaultConfig.NodeID = nodeID
		args := []string{dummyExec}

		nodeExec1 := createDummyNodeWithID("n0", false)
		taskExec1 := createDummyTaskExecutionForNode("n0", "task21")
		taskExec2 := createDummyTaskExecutionForNode("n0", "task22")

		nodeExecutions := []*admin.NodeExecution{nodeExec1}
		nodeExecList := &admin.NodeExecutionList{NodeExecutions: nodeExecutions}
		mockFetcherExt.OnFetchExecutionMatch(ctx, dummyExec, dummyProject, dummyDomain).Return(&admin.Execution{}, nil)
		mockFetcherExt.OnFetchNodeExecutionDetailsMatch(ctx, dummyExec, dummyProject, dummyDomain, "").Return(nodeExecList, nil)
		mockFetcherExt.OnFetchTaskExecutionsOnNodeMatch(ctx, mock.Anything, dummyExec, dummyProject, dummyDomain).Return(&admin.TaskExecutionList{
			TaskExecutions: []*admin.TaskExecution{taskExec1, taskExec2},
		}, nil)
		mockFetcherExt.OnFetchNodeExecutionDataMatch(ctx, mock.Anything, dummyExec, dummyProject, dummyDomain).Return(nil, fmt.Errorf("error in fetching data"))

		err := getExecutionFunc(ctx, args, mockCmdCtx)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("error in fetching data"), err)
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

		args := []string{dummyExec}
		for _, tt := range tests {
			s := testutils.SetupWithExt()
			config.GetConfig().Output = tt.outputFormat
			execution.DefaultConfig.NodeID = tt.nodeID

			ctx := s.Ctx
			mockCmdCtx := s.CmdCtx
			mockFetcherExt := s.FetcherExt
			nodeExecToTaskExec := map[string]*admin.TaskExecutionList{}

			nodeExec1 := createDummyNodeWithID("n0", false)
			taskExec1 := createDummyTaskExecutionForNode("n0", "task21")
			taskExec2 := createDummyTaskExecutionForNode("n0", "task22")

			nodeExecToTaskExec["n0"] = &admin.TaskExecutionList{
				TaskExecutions: []*admin.TaskExecution{taskExec1, taskExec2},
			}

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

			mockFetcherExt.OnFetchExecutionMatch(ctx, dummyExec, dummyProject, dummyDomain).Return(&admin.Execution{}, nil)
			mockFetcherExt.OnFetchNodeExecutionDetailsMatch(ctx, dummyExec, dummyProject, dummyDomain, "").Return(nodeExecList, nil)
			mockFetcherExt.OnFetchTaskExecutionsOnNodeMatch(ctx, "n0", dummyExec, dummyProject, dummyDomain).Return(&admin.TaskExecutionList{
				TaskExecutions: []*admin.TaskExecution{taskExec1, taskExec2},
			}, nil)
			mockFetcherExt.OnFetchNodeExecutionDataMatch(ctx, mock.Anything, dummyExec, dummyProject, dummyDomain).Return(dataResp, nil)
			got := getExecutionFunc(ctx, args, mockCmdCtx)
			assert.Equal(t, tt.want, got)
		}
	})
}

func TestGetExecutionFuncWithError(t *testing.T) {
	ctx := context.Background()
	getExecutionSetup()
	_ = &admin.Execution{
		Id: &core.WorkflowExecutionIdentifier{
			Project: projectValue,
			Domain:  domainValue,
			Name:    executionNameValue,
		},
		Spec: &admin.ExecutionSpec{
			LaunchPlan: &core.Identifier{
				Project: projectValue,
				Domain:  domainValue,
				Name:    launchPlanNameValue,
				Version: launchPlanVersionValue,
			},
		},
		Closure: &admin.ExecutionClosure{
			WorkflowId: &core.Identifier{
				Project: projectValue,
				Domain:  domainValue,
				Name:    workflowNameValue,
				Version: workflowVersionValue,
			},
			Phase: core.WorkflowExecution_SUCCEEDED,
		},
	}

	args := []string{executionNameValue}
	s := testutils.SetupWithExt()
	s.FetcherExt.OnFetchExecutionMatch(s.Ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("execution NotFound"))
	err := getExecutionFunc(s.Ctx, args, s.CmdCtx)
	assert.NotNil(t, err)
	assert.Equal(t, err, errors.New("execution NotFound"))
	s.FetcherExt.AssertCalled(t, "FetchExecution", ctx, "e124", "dummyProject", "dummyDomain")
}
