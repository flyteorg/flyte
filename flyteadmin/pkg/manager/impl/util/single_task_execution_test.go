package util

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/durationpb"

	flyteAdminErrors "github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/testutils"
	managerMocks "github.com/flyteorg/flyte/flyteadmin/pkg/manager/mocks"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	repositoryMocks "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/mocks"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	runtimeMocks "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/mocks"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func TestGenerateNodeNameFromTask(t *testing.T) {
	assert.EqualValues(t, "foo-12345", generateNodeNameFromTask("foo@`!=+- 1,2$3(4)5"))
	assert.EqualValues(t, "nametoadefinedtasksomewhereincodeveryverynestedcrazy",
		generateNodeNameFromTask("app.long.path.name.to.a.defined.task.somewhere.in.code.very.very.nested.crazy"))
}

func TestGenerateWorkflowNameFromTask(t *testing.T) {
	assert.EqualValues(t, ".flytegen.SingleTask", generateWorkflowNameFromTask("SingleTask"))
}

func TestGenerateBindings(t *testing.T) {
	nodeID := "nodeID"
	outputs := &core.VariableMap{
		Variables: map[string]*core.Variable{
			"output1": {},
			"output2": {},
		},
	}
	generatedBindings := generateBindings(outputs, nodeID)
	assert.Contains(t, generatedBindings,
		&core.Binding{
			Var: "output1",
			Binding: &core.BindingData{
				Value: &core.BindingData_Promise{
					Promise: &core.OutputReference{
						NodeId: nodeID,
						Var:    "output1",
					},
				},
			},
		},
	)
	assert.Contains(t, generatedBindings,
		&core.Binding{
			Var: "output2",
			Binding: &core.BindingData{
				Value: &core.BindingData_Promise{
					Promise: &core.OutputReference{
						NodeId: nodeID,
						Var:    "output2",
					},
				},
			},
		})
}

func TestCreateOrGetWorkflowModel(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	var getCalledCount = 0
	var newlyCreatedWorkflow models.Workflow
	workflowcreateFunc := func(input models.Workflow, descriptionEntity *models.DescriptionEntity) error {
		newlyCreatedWorkflow = input
		return nil
	}
	repository.WorkflowRepo().(*repositoryMocks.MockWorkflowRepo).SetCreateCallback(workflowcreateFunc)

	workflowGetFunc := func(input interfaces.Identifier) (models.Workflow, error) {
		if getCalledCount == 0 {
			getCalledCount++
			return models.Workflow{}, flyteAdminErrors.NewFlyteAdminErrorf(codes.NotFound, "not found")
		}
		getCalledCount++
		return newlyCreatedWorkflow, nil
	}
	repository.WorkflowRepo().(*repositoryMocks.MockWorkflowRepo).SetGetCallback(workflowGetFunc)

	mockNamedEntityManager := managerMocks.NamedEntityInterface{}
	mockNamedEntityManager.EXPECT().UpdateNamedEntity(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, request *admin.NamedEntityUpdateRequest) (*admin.NamedEntityUpdateResponse, error) {
		assert.Equal(t, request.GetResourceType(), core.ResourceType_WORKFLOW)
		assert.True(t, proto.Equal(request.GetId(), &admin.NamedEntityIdentifier{
			Project: "flytekit",
			Domain:  "production",
			Name:    ".flytegen.app.workflows.MyWorkflow.my_task",
		}), fmt.Sprintf("%+v", request.GetId()))
		assert.True(t, proto.Equal(request.GetMetadata(), &admin.NamedEntityMetadata{
			State: admin.NamedEntityState_SYSTEM_GENERATED,
		}))
		return &admin.NamedEntityUpdateResponse{}, nil
	})

	mockWorkflowManager := managerMocks.WorkflowInterface{}
	mockWorkflowManager.EXPECT().CreateWorkflow(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, request *admin.WorkflowCreateRequest) (*admin.WorkflowCreateResponse, error) {
		assert.True(t, proto.Equal(request.GetId(), &core.Identifier{
			ResourceType: core.ResourceType_WORKFLOW,
			Project:      "flytekit",
			Domain:       "production",
			Name:         ".flytegen.app.workflows.MyWorkflow.my_task",
			Version:      "12345",
		}), fmt.Sprintf("%+v", request.GetId()))
		assert.Len(t, request.GetSpec().GetTemplate().GetNodes(), 1)
		assert.Equal(t, request.GetSpec().GetTemplate().GetNodes()[0].GetMetadata().GetRetries().GetRetries(), uint32(2))

		return &admin.WorkflowCreateResponse{}, nil
	})
	taskIdentifier := &core.Identifier{
		ResourceType: core.ResourceType_TASK,
		Project:      "flytekit",
		Domain:       "production",
		Name:         "app.workflows.MyWorkflow.my_task",
		Version:      "12345",
	}
	var task = &admin.Task{
		Id: taskIdentifier,
		Closure: &admin.TaskClosure{
			CompiledTask: &core.CompiledTask{
				Template: &core.TaskTemplate{
					Id: taskIdentifier,
					Interface: &core.TypedInterface{
						Inputs: &core.VariableMap{
							Variables: map[string]*core.Variable{
								"an_int": {
									Type: &core.LiteralType{
										Type: &core.LiteralType_Simple{
											Simple: core.SimpleType_INTEGER,
										},
									},
								},
							},
						},
						Outputs: &core.VariableMap{
							Variables: map[string]*core.Variable{
								"an_output": {
									Type: &core.LiteralType{
										Type: &core.LiteralType_Simple{
											Simple: core.SimpleType_DURATION,
										},
									},
								},
							},
						},
					},
					Metadata: &core.TaskMetadata{
						Retries: &core.RetryStrategy{
							Retries: 2,
						},
					},
				},
			},
		},
	}
	workflowModel, err := CreateOrGetWorkflowModel(context.Background(), &admin.ExecutionCreateRequest{
		Project: "flytekit",
		Domain:  "production",
		Name:    "SingleTaskExecution",
		Spec: &admin.ExecutionSpec{
			LaunchPlan: taskIdentifier,
		},
	}, repository, &mockWorkflowManager, &mockNamedEntityManager, taskIdentifier, task)
	assert.NoError(t, err)
	assert.EqualValues(t, workflowModel, &newlyCreatedWorkflow)
}

func TestCreateOrGetLaunchPlan(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	var getCalledCount = 0
	var newlyCreatedLP models.LaunchPlan
	launchPlanCreateFunc := func(input models.LaunchPlan) error {
		newlyCreatedLP = input
		return nil
	}
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetCreateCallback(launchPlanCreateFunc)

	launchPlanGetFunc := func(input interfaces.Identifier) (models.LaunchPlan, error) {
		if getCalledCount == 0 {
			getCalledCount++
			return models.LaunchPlan{}, flyteAdminErrors.NewFlyteAdminErrorf(codes.NotFound, "not found")
		}
		getCalledCount++
		return newlyCreatedLP, nil
	}
	repository.LaunchPlanRepo().(*repositoryMocks.MockLaunchPlanRepo).SetGetCallback(launchPlanGetFunc)

	workflowInterface := &core.TypedInterface{
		Inputs: &core.VariableMap{
			Variables: map[string]*core.Variable{
				"an_int": {
					Type: &core.LiteralType{
						Type: &core.LiteralType_Simple{
							Simple: core.SimpleType_INTEGER,
						},
					},
				},
			},
		},
		Outputs: &core.VariableMap{
			Variables: map[string]*core.Variable{
				"an_output": {
					Type: &core.LiteralType{
						Type: &core.LiteralType_Simple{
							Simple: core.SimpleType_DURATION,
						},
					},
				},
			},
		},
	}
	workflowID := uint(12)

	mockNamedEntityManager := managerMocks.NamedEntityInterface{}
	mockNamedEntityManager.EXPECT().UpdateNamedEntity(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, request *admin.NamedEntityUpdateRequest) (*admin.NamedEntityUpdateResponse, error) {
		assert.Equal(t, request.GetResourceType(), core.ResourceType_LAUNCH_PLAN)
		assert.True(t, proto.Equal(request.GetId(), &admin.NamedEntityIdentifier{
			Project: "flytekit",
			Domain:  "production",
			Name:    ".flytegen.app.workflows.MyWorkflow.my_task",
		}), fmt.Sprintf("%+v", request.GetId()))
		assert.True(t, proto.Equal(request.GetMetadata(), &admin.NamedEntityMetadata{
			State: admin.NamedEntityState_SYSTEM_GENERATED,
		}))
		return &admin.NamedEntityUpdateResponse{}, nil
	})

	taskIdentifier := &core.Identifier{
		ResourceType: core.ResourceType_TASK,
		Project:      "flytekit",
		Domain:       "production",
		Name:         "app.workflows.MyWorkflow.my_task",
		Version:      "12345",
	}
	config := runtimeMocks.NewMockConfigurationProvider(
		testutils.GetApplicationConfigWithDefaultDomains(), nil, nil, nil, nil, nil)
	spec := admin.ExecutionSpec{
		LaunchPlan: taskIdentifier,
		AuthRole: &admin.AuthRole{
			AssumableIamRole: "assumable_role",
		},
	}
	launchPlan, err := CreateOrGetLaunchPlan(
		context.Background(), repository, config, &mockNamedEntityManager, taskIdentifier, workflowInterface, workflowID, &spec)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(&core.Identifier{
		ResourceType: core.ResourceType_LAUNCH_PLAN,
		Project:      "flytekit",
		Domain:       "production",
		Name:         ".flytegen.app.workflows.MyWorkflow.my_task",
		Version:      "12345",
	}, launchPlan.GetId()))
	assert.True(t, proto.Equal(launchPlan.GetClosure().GetExpectedOutputs(), workflowInterface.GetOutputs()))
	assert.True(t, proto.Equal(launchPlan.GetSpec().GetAuthRole(), spec.GetAuthRole()))
}

func TestRetryStrategyHandling(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	var getCalledCount = 0
	var newlyCreatedWorkflow models.Workflow
	workflowcreateFunc := func(input models.Workflow, descriptionEntity *models.DescriptionEntity) error {
		newlyCreatedWorkflow = input
		return nil
	}
	repository.WorkflowRepo().(*repositoryMocks.MockWorkflowRepo).SetCreateCallback(workflowcreateFunc)

	workflowGetFunc := func(input interfaces.Identifier) (models.Workflow, error) {
		if getCalledCount == 0 {
			getCalledCount++
			return models.Workflow{}, flyteAdminErrors.NewFlyteAdminErrorf(codes.NotFound, "not found")
		}
		getCalledCount++
		return newlyCreatedWorkflow, nil
	}
	repository.WorkflowRepo().(*repositoryMocks.MockWorkflowRepo).SetGetCallback(workflowGetFunc)

	mockNamedEntityManager := managerMocks.NamedEntityInterface{}
	mockNamedEntityManager.EXPECT().UpdateNamedEntity(mock.Anything, mock.Anything).Return(&admin.NamedEntityUpdateResponse{}, nil)

	taskIdentifier := &core.Identifier{
		ResourceType: core.ResourceType_TASK,
		Project:      "flytekit",
		Domain:       "production",
		Name:         "app.workflows.MyWorkflow.my_task",
		Version:      "12345",
	}

	t.Run("with no retry strategy", func(t *testing.T) {
		mockWorkflowManager := managerMocks.WorkflowInterface{}
		mockWorkflowManager.EXPECT().CreateWorkflow(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, request *admin.WorkflowCreateRequest) (*admin.WorkflowCreateResponse, error) {
			// Verify default retry strategy is used
			retries := request.GetSpec().GetTemplate().GetNodes()[0].GetMetadata().GetRetries()
			assert.Equal(t, uint32(3), retries.GetRetries())
			assert.NotNil(t, retries.GetOnOom())
			assert.NotNil(t, retries.GetOnOom().GetBackoff())
			assert.Equal(t, uint32(1), retries.GetOnOom().GetBackoff().GetMaxExponent())
			assert.Equal(t, int64(0), retries.GetOnOom().GetBackoff().GetMax().GetSeconds())
			return &admin.WorkflowCreateResponse{}, nil
		})

		task := &admin.Task{
			Id: taskIdentifier,
			Closure: &admin.TaskClosure{
				CompiledTask: &core.CompiledTask{
					Template: &core.TaskTemplate{
						Id: taskIdentifier,
						Interface: &core.TypedInterface{
							Inputs:  &core.VariableMap{Variables: map[string]*core.Variable{}},
							Outputs: &core.VariableMap{Variables: map[string]*core.Variable{}},
						},
						// No retry strategy defined
						Metadata: &core.TaskMetadata{},
					},
				},
			},
		}

		_, err := CreateOrGetWorkflowModel(context.Background(), &admin.ExecutionCreateRequest{
			Project: "flytekit",
			Domain:  "production",
			Name:    "SingleTaskExecution",
			Spec: &admin.ExecutionSpec{
				LaunchPlan: taskIdentifier,
			},
		}, repository, &mockWorkflowManager, &mockNamedEntityManager, taskIdentifier, task)
		assert.NoError(t, err)
	})

	t.Run("with retry strategy but no OnOOM", func(t *testing.T) {
		mockWorkflowManager := managerMocks.WorkflowInterface{}
		mockWorkflowManager.EXPECT().CreateWorkflow(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, request *admin.WorkflowCreateRequest) (*admin.WorkflowCreateResponse, error) {
			// Verify retry strategy with default OnOOM
			retries := request.GetSpec().GetTemplate().GetNodes()[0].GetMetadata().GetRetries()
			assert.Equal(t, uint32(5), retries.GetRetries())
			assert.NotNil(t, retries.GetOnOom())
			assert.NotNil(t, retries.GetOnOom().GetBackoff())
			assert.Equal(t, uint32(1), retries.GetOnOom().GetBackoff().GetMaxExponent())
			assert.Equal(t, int64(0), retries.GetOnOom().GetBackoff().GetMax().GetSeconds())
			return &admin.WorkflowCreateResponse{}, nil
		})

		task := &admin.Task{
			Id: taskIdentifier,
			Closure: &admin.TaskClosure{
				CompiledTask: &core.CompiledTask{
					Template: &core.TaskTemplate{
						Id: taskIdentifier,
						Interface: &core.TypedInterface{
							Inputs:  &core.VariableMap{Variables: map[string]*core.Variable{}},
							Outputs: &core.VariableMap{Variables: map[string]*core.Variable{}},
						},
						Metadata: &core.TaskMetadata{
							Retries: &core.RetryStrategy{
								Retries: 5,
								// No OnOOM defined
							},
						},
					},
				},
			},
		}

		_, err := CreateOrGetWorkflowModel(context.Background(), &admin.ExecutionCreateRequest{
			Project: "flytekit",
			Domain:  "production",
			Name:    "SingleTaskExecution",
			Spec: &admin.ExecutionSpec{
				LaunchPlan: taskIdentifier,
			},
		}, repository, &mockWorkflowManager, &mockNamedEntityManager, taskIdentifier, task)
		assert.NoError(t, err)
	})

	t.Run("with retry strategy and OnOOM but no Backoff", func(t *testing.T) {
		mockWorkflowManager := managerMocks.WorkflowInterface{}
		mockWorkflowManager.EXPECT().CreateWorkflow(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, request *admin.WorkflowCreateRequest) (*admin.WorkflowCreateResponse, error) {
			// Verify retry strategy with default Backoff
			retries := request.GetSpec().GetTemplate().GetNodes()[0].GetMetadata().GetRetries()
			assert.Equal(t, uint32(4), retries.GetRetries())
			assert.NotNil(t, retries.GetOnOom())
			assert.Equal(t, float32(1.5), retries.GetOnOom().GetFactor())
			assert.Equal(t, "2Gi", retries.GetOnOom().GetLimit())
			assert.NotNil(t, retries.GetOnOom().GetBackoff())
			assert.Equal(t, uint32(1), retries.GetOnOom().GetBackoff().GetMaxExponent())
			assert.Equal(t, int64(0), retries.GetOnOom().GetBackoff().GetMax().GetSeconds())
			return &admin.WorkflowCreateResponse{}, nil
		})

		task := &admin.Task{
			Id: taskIdentifier,
			Closure: &admin.TaskClosure{
				CompiledTask: &core.CompiledTask{
					Template: &core.TaskTemplate{
						Id: taskIdentifier,
						Interface: &core.TypedInterface{
							Inputs:  &core.VariableMap{Variables: map[string]*core.Variable{}},
							Outputs: &core.VariableMap{Variables: map[string]*core.Variable{}},
						},
						Metadata: &core.TaskMetadata{
							Retries: &core.RetryStrategy{
								Retries: 4,
								OnOom: &core.RetryOnOOM{
									Factor: 1.5,
									Limit:  "2Gi",
									// No Backoff defined
								},
							},
						},
					},
				},
			},
		}

		_, err := CreateOrGetWorkflowModel(context.Background(), &admin.ExecutionCreateRequest{
			Project: "flytekit",
			Domain:  "production",
			Name:    "SingleTaskExecution",
			Spec: &admin.ExecutionSpec{
				LaunchPlan: taskIdentifier,
			},
		}, repository, &mockWorkflowManager, &mockNamedEntityManager, taskIdentifier, task)
		assert.NoError(t, err)
	})

	t.Run("with complete retry strategy", func(t *testing.T) {
		mockWorkflowManager := managerMocks.WorkflowInterface{}
		mockWorkflowManager.EXPECT().CreateWorkflow(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, request *admin.WorkflowCreateRequest) (*admin.WorkflowCreateResponse, error) {
			// Verify custom retry strategy is preserved
			retries := request.GetSpec().GetTemplate().GetNodes()[0].GetMetadata().GetRetries()
			assert.Equal(t, uint32(6), retries.GetRetries())
			assert.NotNil(t, retries.GetOnOom())
			assert.Equal(t, float32(2.0), retries.GetOnOom().GetFactor())
			assert.Equal(t, "4Gi", retries.GetOnOom().GetLimit())
			assert.NotNil(t, retries.GetOnOom().GetBackoff())
			assert.Equal(t, uint32(2), retries.GetOnOom().GetBackoff().GetMaxExponent())
			assert.Equal(t, int64(30), retries.GetOnOom().GetBackoff().GetMax().GetSeconds())
			return &admin.WorkflowCreateResponse{}, nil
		})

		customBackoff := &core.ExponentialBackoff{
			MaxExponent: 2,
			Max: &durationpb.Duration{
				Seconds: 30,
			},
		}

		task := &admin.Task{
			Id: taskIdentifier,
			Closure: &admin.TaskClosure{
				CompiledTask: &core.CompiledTask{
					Template: &core.TaskTemplate{
						Id: taskIdentifier,
						Interface: &core.TypedInterface{
							Inputs:  &core.VariableMap{Variables: map[string]*core.Variable{}},
							Outputs: &core.VariableMap{Variables: map[string]*core.Variable{}},
						},
						Metadata: &core.TaskMetadata{
							Retries: &core.RetryStrategy{
								Retries: 6,
								OnOom: &core.RetryOnOOM{
									Factor:  2.0,
									Limit:   "4Gi",
									Backoff: customBackoff,
								},
							},
						},
					},
				},
			},
		}

		_, err := CreateOrGetWorkflowModel(context.Background(), &admin.ExecutionCreateRequest{
			Project: "flytekit",
			Domain:  "production",
			Name:    "SingleTaskExecution",
			Spec: &admin.ExecutionSpec{
				LaunchPlan: taskIdentifier,
			},
		}, repository, &mockWorkflowManager, &mockNamedEntityManager, taskIdentifier, task)
		assert.NoError(t, err)
	})
}
