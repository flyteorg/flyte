package create

import (
	"fmt"
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/flyteorg/flytectl/cmd/config"
	cmdCore "github.com/flyteorg/flytectl/cmd/core"
	"github.com/flyteorg/flytectl/cmd/testutils"
)

type createSuite struct {
	suite.Suite
	testutils.TestStruct
	originalExecConfig ExecutionConfig
}

func (s *createSuite) SetupTest() {
	s.TestStruct = setup()

	// TODO: migrate to new command context from testutils
	s.CmdCtx = cmdCore.NewCommandContext(s.MockClient, s.MockOutStream)
	s.originalExecConfig = *executionConfig
}

func (s *createSuite) TearDownTest() {
	orig := s.originalExecConfig
	executionConfig = &orig
	s.MockAdminClient.AssertExpectations(s.T())
}

func (s *createSuite) onGetTask() {
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
	s.MockAdminClient.OnGetTaskMatch(s.Ctx, mock.Anything).Return(task1, nil)
}

func (s *createSuite) onGetLaunchPlan() {
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
			Name:    "core.control_flow.merge_sort.merge_sort",
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
			Name:         "core.control_flow.merge_sort.merge_sort",
			Version:      "v3",
		},
	}
	s.MockAdminClient.OnGetLaunchPlanMatch(s.Ctx, objectGetRequest).Return(launchPlan1, nil).Once()
}

func (s *createSuite) Test_CreateTaskExecution() {
	s.onGetTask()
	executionCreateResponseTask := &admin.ExecutionCreateResponse{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "flytesnacks",
			Domain:  "development",
			Name:    "ff513c0e44b5b4a35aa5",
		},
	}
	expected := &admin.ExecutionCreateRequest{
		Project: "dummyProject",
		Domain:  "dummyDomain",
		Spec: &admin.ExecutionSpec{
			LaunchPlan: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      "dummyProject",
				Domain:       "dummyDomain",
				Name:         "task1",
				Version:      "v2",
			},
			Metadata: &admin.ExecutionMetadata{Mode: admin.ExecutionMetadata_MANUAL, Principal: "sdk", Nesting: 0},
			AuthRole: &admin.AuthRole{
				KubernetesServiceAccount: executionConfig.KubeServiceAcct,
				AssumableIamRole:         "iamRoleARN",
			},
			SecurityContext: &core.SecurityContext{
				RunAs: &core.Identity{
					K8SServiceAccount: executionConfig.KubeServiceAcct,
					IamRole:           "iamRoleARN",
				},
			},
			ClusterAssignment: &admin.ClusterAssignment{ClusterPoolName: "gpu"},
			Envs:              &admin.Envs{},
		},
	}
	s.MockAdminClient.
		OnCreateExecutionMatch(s.Ctx, mock.Anything).
		Run(func(args mock.Arguments) {
			actual := args.Get(1).(*admin.ExecutionCreateRequest)
			actual.Name = ""
			actual.Inputs = nil
			s.True(proto.Equal(expected, actual), actual.String())
		}).
		Return(executionCreateResponseTask, nil).
		Once()
	executionConfig.ExecFile = testDataFolder + "task_execution_spec_with_iamrole.yaml"

	err := createExecutionCommand(s.Ctx, nil, s.CmdCtx)

	s.NoError(err)
	tearDownAndVerify(s.T(), s.Writer, `execution identifier project:"flytesnacks" domain:"development" name:"ff513c0e44b5b4a35aa5" `)
}

func (s *createSuite) Test_CreateTaskExecution_GetTaskError() {
	expected := fmt.Errorf("error")
	s.MockAdminClient.OnGetTaskMatch(s.Ctx, mock.Anything).Return(nil, expected).Once()
	executionConfig.ExecFile = testDataFolder + "task_execution_spec.yaml"

	err := createExecutionCommand(s.Ctx, nil, s.CmdCtx)

	s.Equal(expected, err)
}

func (s *createSuite) Test_CreateTaskExecution_CreateExecutionError() {
	s.onGetTask()
	s.MockAdminClient.
		OnCreateExecutionMatch(s.Ctx, mock.Anything).
		Return(nil, fmt.Errorf("error launching task")).
		Once()
	executionConfig.ExecFile = testDataFolder + "task_execution_spec.yaml"

	err := createExecutionCommand(s.Ctx, nil, s.CmdCtx)

	s.EqualError(err, "error launching task")
}

func (s *createSuite) Test_CreateLaunchPlanExecution() {
	executionCreateResponseLP := &admin.ExecutionCreateResponse{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "flytesnacks",
			Domain:  "development",
			Name:    "f652ea3596e7f4d80a0e",
		},
	}
	s.onGetLaunchPlan()
	s.MockAdminClient.OnCreateExecutionMatch(s.Ctx, mock.Anything).Return(executionCreateResponseLP, nil)
	executionConfig.ExecFile = testDataFolder + "launchplan_execution_spec.yaml"

	err := createExecutionCommand(s.Ctx, nil, s.CmdCtx)

	s.NoError(err)
	tearDownAndVerify(s.T(), s.Writer, `execution identifier project:"flytesnacks" domain:"development" name:"f652ea3596e7f4d80a0e" `)
}

func (s *createSuite) Test_CreateLaunchPlan_GetLaunchPlanError() {
	expected := fmt.Errorf("error")
	s.MockAdminClient.OnGetLaunchPlanMatch(s.Ctx, mock.Anything).Return(nil, expected).Once()
	executionConfig.ExecFile = testDataFolder + "launchplan_execution_spec.yaml"

	err := createExecutionCommand(s.Ctx, nil, s.CmdCtx)

	s.Equal(expected, err)
}

func (s *createSuite) Test_CreateRelaunchExecution() {
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
	s.MockAdminClient.OnRelaunchExecutionMatch(s.Ctx, relaunchRequest).Return(relaunchExecResponse, nil).Once()

	err := createExecutionCommand(s.Ctx, nil, s.CmdCtx)

	s.NoError(err)
	tearDownAndVerify(s.T(), s.Writer, `execution identifier project:"flytesnacks" domain:"development" name:"f652ea3596e7f4d80a0e"`)
}

func (s *createSuite) Test_CreateRecoverExecution() {
	originalExecutionName := "abc123"
	recoverExecResponse := &admin.ExecutionCreateResponse{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "flytesnacks",
			Domain:  "development",
			Name:    "f652ea3596e7f4d80a0e",
		},
	}
	executionConfig.Recover = originalExecutionName
	recoverRequest := &admin.ExecutionRecoverRequest{
		Id: &core.WorkflowExecutionIdentifier{
			Name:    originalExecutionName,
			Project: config.GetConfig().Project,
			Domain:  config.GetConfig().Domain,
		},
	}
	s.MockAdminClient.OnRecoverExecutionMatch(s.Ctx, recoverRequest).Return(recoverExecResponse, nil).Once()

	err := createExecutionCommand(s.Ctx, nil, s.CmdCtx)

	s.NoError(err)
	tearDownAndVerify(s.T(), s.Writer, `execution identifier project:"flytesnacks" domain:"development" name:"f652ea3596e7f4d80a0e"`)
}

func (s *createSuite) TestCreateExecutionFuncInvalid() {
	executionConfig.Relaunch = ""
	executionConfig.ExecFile = ""
	err := createExecutionCommand(s.Ctx, nil, s.CmdCtx)
	s.EqualError(err, "executionConfig, relaunch and recover can't be empty. Run the flytectl get task/launchplan to generate the config")

	executionConfig.ExecFile = "Invalid-file"
	err = createExecutionCommand(s.Ctx, nil, s.CmdCtx)
	s.EqualError(err, fmt.Sprintf("unable to read from %v yaml file", executionConfig.ExecFile))

	executionConfig.ExecFile = testDataFolder + "invalid_execution_spec.yaml"
	err = createExecutionCommand(s.Ctx, nil, s.CmdCtx)
	s.EqualError(err, "either task or workflow name should be specified to launch an execution")
}

func (s *createSuite) Test_CreateTaskExecution_DryRun() {
	s.onGetTask()
	executionConfig.DryRun = true
	executionConfig.ExecFile = testDataFolder + "task_execution_spec_with_iamrole.yaml"

	err := createExecutionCommand(s.Ctx, []string{"target"}, s.CmdCtx)

	s.NoError(err)
}

func TestCreateSuite(t *testing.T) {
	suite.Run(t, &createSuite{originalExecConfig: *executionConfig})
}
