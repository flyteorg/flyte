package create

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flytectl/cmd/config"
	"github.com/flyteorg/flytectl/cmd/testutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// This function needs to be called after testutils.Steup()
func createExecutionSetup() {
	ctx = testutils.Ctx
	cmdCtx = testutils.CmdCtx
	mockClient = testutils.MockClient
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
	mockClient.OnGetTaskMatch(ctx, mock.Anything).Return(task1, nil)
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
			Version:      "v3",
		},
	}
	mockClient.OnGetLaunchPlanMatch(ctx, objectGetRequest).Return(launchPlan1, nil)
}

func TestCreateTaskExecutionFunc(t *testing.T) {
	setup()
	createExecutionSetup()
	executionCreateResponseTask := &admin.ExecutionCreateResponse{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "flytesnacks",
			Domain:  "development",
			Name:    "ff513c0e44b5b4a35aa5",
		},
	}
	mockClient.OnCreateExecutionMatch(ctx, mock.Anything).Return(executionCreateResponseTask, nil)
	executionConfig.ExecFile = testDataFolder + "task_execution_spec.yaml"
	err = createExecutionCommand(ctx, args, cmdCtx)
	assert.Nil(t, err)
	mockClient.AssertCalled(t, "CreateExecution", ctx, mock.Anything)
	tearDownAndVerify(t, `execution identifier project:"flytesnacks" domain:"development" name:"ff513c0e44b5b4a35aa5" `)
}

func TestCreateTaskExecutionFuncError(t *testing.T) {
	setup()
	createExecutionSetup()
	mockClient.OnCreateExecutionMatch(ctx, mock.Anything).Return(nil, fmt.Errorf("error launching task"))
	executionConfig.ExecFile = testDataFolder + "task_execution_spec.yaml"
	err = createExecutionCommand(ctx, args, cmdCtx)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf("error launching task"), err)
	mockClient.AssertCalled(t, "CreateExecution", ctx, mock.Anything)
}

func TestCreateLaunchPlanExecutionFunc(t *testing.T) {
	setup()
	createExecutionSetup()
	executionCreateResponseLP := &admin.ExecutionCreateResponse{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "flytesnacks",
			Domain:  "development",
			Name:    "f652ea3596e7f4d80a0e",
		},
	}
	mockClient.OnCreateExecutionMatch(ctx, mock.Anything).Return(executionCreateResponseLP, nil)
	executionConfig.ExecFile = testDataFolder + "launchplan_execution_spec.yaml"
	err = createExecutionCommand(ctx, args, cmdCtx)
	assert.Nil(t, err)
	mockClient.AssertCalled(t, "CreateExecution", ctx, mock.Anything)
	tearDownAndVerify(t, `execution identifier project:"flytesnacks" domain:"development" name:"f652ea3596e7f4d80a0e" `)
}

func TestCreateRelaunchExecutionFunc(t *testing.T) {
	setup()
	createExecutionSetup()
	relaunchExecResponse := &admin.ExecutionCreateResponse{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "flytesnacks",
			Domain:  "development",
			Name:    "f652ea3596e7f4d80a0e",
		},
	}

	executionConfig.Relaunch = relaunchExecResponse.Id.Name
	relaunchRequest := &admin.ExecutionRelaunchRequest{
		Id: &core.WorkflowExecutionIdentifier{
			Name:    executionConfig.Relaunch,
			Project: config.GetConfig().Project,
			Domain:  config.GetConfig().Domain,
		},
	}
	mockClient.OnRelaunchExecutionMatch(ctx, relaunchRequest).Return(relaunchExecResponse, nil)
	err = createExecutionCommand(ctx, args, cmdCtx)
	assert.Nil(t, err)
	mockClient.AssertCalled(t, "RelaunchExecution", ctx, relaunchRequest)
	tearDownAndVerify(t, `execution identifier project:"flytesnacks" domain:"development" name:"f652ea3596e7f4d80a0e"`)
}

func TestCreateExecutionFuncInvalid(t *testing.T) {
	setup()
	createExecutionSetup()
	executionConfig.Relaunch = ""
	executionConfig.ExecFile = ""
	err = createExecutionCommand(ctx, args, cmdCtx)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf("executionConfig or relaunch can't be empty. Run the flytectl get task/launchplan to generate the config"), err)
	executionConfig.ExecFile = "Invalid-file"
	err = createExecutionCommand(ctx, args, cmdCtx)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf("unable to read from %v yaml file", executionConfig.ExecFile), err)
	executionConfig.ExecFile = testDataFolder + "invalid_execution_spec.yaml"
	err = createExecutionCommand(ctx, args, cmdCtx)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf("either one of task or workflow name should be specified to launch an execution"), err)
}
