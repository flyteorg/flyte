package create

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flytectl/cmd/config"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TestStruct struct {
	executionConfig *ExecutionConfig
	args            []string
}

// This function needs to be called after testutils.Steup()
func createExecutionSetup(s *testutils.TestStruct) (t TestStruct) {
	ctx := s.Ctx
	t.executionConfig = &ExecutionConfig{}
	// TODO: migrate to new command context from testutils
	s.CmdCtx = cmdCore.NewCommandContext(s.MockClient, s.MockOutStream)
	sortedListLiteralType := core.Variable{
		Type: &core.LiteralType{
			Type: &core.LiteralType_CollectionType{
				CollectionType: &core.LiteralType{
					Type: &core.LiteralType_Simple{
						Simple: core.SimpleType_INTEGER,
					},
				},
			},
		},
	}
	variableMap := map[string]*core.Variable{
		"sorted_list1": &sortedListLiteralType,
		"sorted_list2": &sortedListLiteralType,
	}

	task1 := &admin.Task{
		Id: &core.Identifier{
			Name:    "task1",
			Version: "v2",
		},
		Closure: &admin.TaskClosure{
			CreatedAt: &timestamppb.Timestamp{Seconds: 1, Nanos: 0},
			CompiledTask: &core.CompiledTask{
				Template: &core.TaskTemplate{
					Interface: &core.TypedInterface{
						Inputs: &core.VariableMap{
							Variables: variableMap,
						},
					},
				},
			},
		},
	}
	s.MockAdminClient.OnGetTaskMatch(ctx, mock.Anything).Return(task1, nil)
	parameterMap := map[string]*core.Parameter{
		"numbers": {
			Var: &core.Variable{
				Type: &core.LiteralType{
					Type: &core.LiteralType_CollectionType{
						CollectionType: &core.LiteralType{
							Type: &core.LiteralType_Simple{
								Simple: core.SimpleType_INTEGER,
							},
						},
					},
				},
			},
		},
		"numbers_count": {
			Var: &core.Variable{
				Type: &core.LiteralType{
					Type: &core.LiteralType_Simple{
						Simple: core.SimpleType_INTEGER,
					},
				},
			},
		},
		"run_local_at_count": {
			Var: &core.Variable{
				Type: &core.LiteralType{
					Type: &core.LiteralType_Simple{
						Simple: core.SimpleType_INTEGER,
					},
				},
			},
			Behavior: &core.Parameter_Default{
				Default: &core.Literal{
					Value: &core.Literal_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Primitive{
								Primitive: &core.Primitive{
									Value: &core.Primitive_Integer{
										Integer: 10,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	launchPlan1 := &admin.LaunchPlan{
		Id: &core.Identifier{
			Name:    "core.advanced.run_merge_sort.merge_sort",
			Version: "v3",
		},
		Spec: &admin.LaunchPlanSpec{
			DefaultInputs: &core.ParameterMap{
				Parameters: parameterMap,
			},
		},
		Closure: &admin.LaunchPlanClosure{
			CreatedAt: &timestamppb.Timestamp{Seconds: 0, Nanos: 0},
			ExpectedInputs: &core.ParameterMap{
				Parameters: parameterMap,
			},
		},
	}
	objectGetRequest := &admin.ObjectGetRequest{
		Id: &core.Identifier{
			ResourceType: core.ResourceType_LAUNCH_PLAN,
			Project:      config.GetConfig().Project,
			Domain:       config.GetConfig().Domain,
			Name:         "core.advanced.run_merge_sort.merge_sort",
			Version:      "v2",
		},
	}
	s.MockAdminClient.OnGetLaunchPlanMatch(ctx, objectGetRequest).Return(launchPlan1, nil)

	return TestStruct{
		executionConfig: executionConfig,
		args:            []string{},
	}
}

func TestCreateTaskExecutionFunc(t *testing.T) {
	s := setup()
	ts := createExecutionSetup(&s)
	executionCreateResponseTask := &admin.ExecutionCreateResponse{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "flytesnacks",
			Domain:  "development",
			Name:    "ff513c0e44b5b4a35aa5",
		},
	}

	ctx := s.Ctx
	s.MockAdminClient.OnCreateExecutionMatch(ctx, mock.Anything).Return(executionCreateResponseTask, nil)
	ts.executionConfig.ExecFile = testDataFolder + "task_execution_spec_with_iamrole.yaml"
	err := createExecutionCommand(ctx, ts.args, s.CmdCtx)
	assert.Nil(t, err)
	s.MockAdminClient.AssertCalled(t, "CreateExecution", ctx, mock.Anything)
	tearDownAndVerify(t, s.Writer, `execution identifier project:"flytesnacks" domain:"development" name:"ff513c0e44b5b4a35aa5" `)
}

func TestCreateTaskExecutionFuncError(t *testing.T) {
	s := setup()
	ts := createExecutionSetup(&s)
	ctx := s.Ctx
	s.MockAdminClient.OnCreateExecutionMatch(ctx, mock.Anything).Return(nil, fmt.Errorf("error launching task"))
	ts.executionConfig.ExecFile = testDataFolder + "task_execution_spec.yaml"
	err := createExecutionCommand(ctx, ts.args, s.CmdCtx)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf("error launching task"), err)
	s.MockAdminClient.AssertCalled(t, "CreateExecution", ctx, mock.Anything)
}

func TestCreateLaunchPlanExecutionFunc(t *testing.T) {
	s := setup()
	ts := createExecutionSetup(&s)
	executionCreateResponseLP := &admin.ExecutionCreateResponse{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "flytesnacks",
			Domain:  "development",
			Name:    "f652ea3596e7f4d80a0e",
		},
	}

	ctx := s.Ctx
	s.MockAdminClient.OnCreateExecutionMatch(ctx, mock.Anything).Return(executionCreateResponseLP, nil)
	ts.executionConfig.ExecFile = testDataFolder + "launchplan_execution_spec.yaml"
	err := createExecutionCommand(ctx, ts.args, s.CmdCtx)
	assert.Nil(t, err)
	s.MockAdminClient.AssertCalled(t, "CreateExecution", ctx, mock.Anything)
	tearDownAndVerify(t, s.Writer, `execution identifier project:"flytesnacks" domain:"development" name:"f652ea3596e7f4d80a0e" `)
}

func TestCreateRelaunchExecutionFunc(t *testing.T) {
	s := setup()
	ts := createExecutionSetup(&s)
	defer func() { ts.executionConfig.Relaunch = "" }()
	relaunchExecResponse := &admin.ExecutionCreateResponse{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "flytesnacks",
			Domain:  "development",
			Name:    "f652ea3596e7f4d80a0e",
		},
	}

	ts.executionConfig.Relaunch = relaunchExecResponse.Id.Name
	relaunchRequest := &admin.ExecutionRelaunchRequest{
		Id: &core.WorkflowExecutionIdentifier{
			Name:    executionConfig.Relaunch,
			Project: config.GetConfig().Project,
			Domain:  config.GetConfig().Domain,
		},
	}
	ctx := s.Ctx
	s.MockAdminClient.OnRelaunchExecutionMatch(ctx, relaunchRequest).Return(relaunchExecResponse, nil)
	err := createExecutionCommand(ctx, ts.args, s.CmdCtx)
	assert.Nil(t, err)
	s.MockAdminClient.AssertCalled(t, "RelaunchExecution", ctx, relaunchRequest)
	tearDownAndVerify(t, s.Writer, `execution identifier project:"flytesnacks" domain:"development" name:"f652ea3596e7f4d80a0e"`)
}

func TestCreateRecoverExecutionFunc(t *testing.T) {
	s := setup()
	ts := createExecutionSetup(&s)
	defer func() { ts.executionConfig.Recover = "" }()

	originalExecutionName := "abc123"
	recoverExecResponse := &admin.ExecutionCreateResponse{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "flytesnacks",
			Domain:  "development",
			Name:    "f652ea3596e7f4d80a0e",
		},
	}

	ts.executionConfig.Recover = originalExecutionName
	recoverRequest := &admin.ExecutionRecoverRequest{
		Id: &core.WorkflowExecutionIdentifier{
			Name:    originalExecutionName,
			Project: config.GetConfig().Project,
			Domain:  config.GetConfig().Domain,
		},
	}

	ctx := s.Ctx
	s.MockAdminClient.OnRecoverExecutionMatch(ctx, recoverRequest).Return(recoverExecResponse, nil)
	err := createExecutionCommand(ctx, ts.args, s.CmdCtx)
	assert.Nil(t, err)
	s.MockAdminClient.AssertCalled(t, "RecoverExecution", ctx, recoverRequest)
	tearDownAndVerify(t, s.Writer, `execution identifier project:"flytesnacks" domain:"development" name:"f652ea3596e7f4d80a0e"`)
	ts.executionConfig.Relaunch = ""
}

func TestCreateExecutionFuncInvalid(t *testing.T) {
	s := setup()
	ts := createExecutionSetup(&s)
	executionConfig := ts.executionConfig
	executionConfig.Relaunch = ""
	executionConfig.ExecFile = ""
	err := createExecutionCommand(s.Ctx, ts.args, s.CmdCtx)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf("executionConfig, relaunch and recover can't be empty. Run the flytectl get task/launchplan to generate the config"), err)
	executionConfig.ExecFile = "Invalid-file"
	err = createExecutionCommand(s.Ctx, ts.args, s.CmdCtx)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf("unable to read from %v yaml file", executionConfig.ExecFile), err)
	executionConfig.ExecFile = testDataFolder + "invalid_execution_spec.yaml"
	err = createExecutionCommand(s.Ctx, ts.args, s.CmdCtx)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf("either task or workflow name should be specified to launch an execution"), err)
}
