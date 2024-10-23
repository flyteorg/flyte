package util

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"

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
	nodeID, err := generateNodeNameFromTask("foo@`!=+- 1,2$3(4)5")
	assert.NoError(t, err)
	assert.EqualValues(t, "foo-12345", nodeID)

	nodeID, err = generateNodeNameFromTask("app.long.path.name.to.a.defined.task.somewhere.in.code.very.very.nested.crazy")
	assert.NoError(t, err)
	assert.EqualValues(t, "fch4xs5i", nodeID)
}

func TestGenerateWorkflowNameFromTask(t *testing.T) {
	assert.EqualValues(t, ".flytegen.SingleTask", generateWorkflowNameFromTask("SingleTask"))
}

func TestGenerateWorkflowNameFromNode(t *testing.T) {
	tests := []struct {
		name         string
		node         *core.Node
		expectedName string
		expectError  bool
	}{
		{
			name: "TaskNode",
			node: &core.Node{
				Id: "n1",
				Target: &core.Node_TaskNode{
					TaskNode: &core.TaskNode{
						Reference: &core.TaskNode_ReferenceId{
							ReferenceId: &core.Identifier{Name: "TaskId"},
						},
					},
				},
			},
			expectedName: ".flytegen.TaskId",
		},
		{
			name: "ArrayNode",
			node: &core.Node{
				Id: "n1",
				Target: &core.Node_ArrayNode{
					ArrayNode: &core.ArrayNode{
						Node: &core.Node{
							Id: "subnode",
							Target: &core.Node_TaskNode{
								TaskNode: &core.TaskNode{
									Reference: &core.TaskNode_ReferenceId{
										ReferenceId: &core.Identifier{Name: "TaskId"},
									},
								},
							},
						},
					},
				},
			},
			expectedName: ".flytegen.subnode",
		},
		{
			name: "WorkflowNode",
			node: &core.Node{
				Id:     "n1",
				Target: &core.Node_WorkflowNode{},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, err := generateWorkflowNameFromNode(tt.node)
			if tt.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.EqualValues(t, tt.expectedName, name)
			}
		})
	}
}

func TestGenerateBindings(t *testing.T) {
	nodeID := "nodeID"
	outputs := core.VariableMap{
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

	mockNamedEntityManager := managerMocks.NamedEntityManager{}
	mockNamedEntityManager.UpdateNamedEntityFunc = func(ctx context.Context, request admin.NamedEntityUpdateRequest) (*admin.NamedEntityUpdateResponse, error) {
		assert.Equal(t, request.ResourceType, core.ResourceType_WORKFLOW)
		assert.True(t, proto.Equal(request.Id, &admin.NamedEntityIdentifier{
			Project: "flytekit",
			Domain:  "production",
			Name:    ".flytegen.app.workflows.MyWorkflow.my_task",
		}), fmt.Sprintf("%+v", request.Id))
		assert.True(t, proto.Equal(request.Metadata, &admin.NamedEntityMetadata{
			State: admin.NamedEntityState_SYSTEM_GENERATED,
		}))
		return &admin.NamedEntityUpdateResponse{}, nil
	}

	mockWorkflowManager := managerMocks.MockWorkflowManager{}
	mockWorkflowManager.SetCreateCallback(func(ctx context.Context, request admin.WorkflowCreateRequest) (*admin.WorkflowCreateResponse, error) {
		assert.True(t, proto.Equal(request.Id, &core.Identifier{
			ResourceType: core.ResourceType_WORKFLOW,
			Project:      "flytekit",
			Domain:       "production",
			Name:         ".flytegen.app.workflows.MyWorkflow.my_task",
			Version:      "12345",
		}), fmt.Sprintf("%+v", request.Id))
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
	workflowModel, err := CreateOrGetWorkflowModel(context.Background(), admin.ExecutionCreateRequest{
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
		context.Background(), repository, config, taskIdentifier, workflowInterface, workflowID, spec.AuthRole, spec.SecurityContext)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(&core.Identifier{
		ResourceType: core.ResourceType_LAUNCH_PLAN,
		Project:      "flytekit",
		Domain:       "production",
		Name:         ".flytegen.app.workflows.MyWorkflow.my_task",
		Version:      "12345",
	}, launchPlan.Id))
	assert.True(t, proto.Equal(launchPlan.Closure.ExpectedOutputs, workflowInterface.Outputs))
	assert.True(t, proto.Equal(launchPlan.Spec.AuthRole, spec.AuthRole))
}

func TestCreateOrGetWorkflowFromNode(t *testing.T) {

	wfIdentifier := &core.Identifier{
		ResourceType: core.ResourceType_WORKFLOW,
		Project:      "flytekit",
		Domain:       "production",
		Name:         "app.workflows.MyWorkflow.my_task",
		Version:      "12345",
		Org:          "org",
	}
	workflowName := "wf_name"
	tests := []struct {
		name           string
		node           *core.Node
		workflowName   string
		expectedName   string
		expectError    bool
		typedInterface *core.TypedInterface
	}{
		{
			name: "TaskNode",
			node: &core.Node{
				Id: "node_id",
				Metadata: &core.NodeMetadata{
					Name: "simple_task",
				},
				Target: &core.Node_TaskNode{
					TaskNode: &core.TaskNode{
						Reference: &core.TaskNode_ReferenceId{
							ReferenceId: &core.Identifier{Name: "simple_task"},
						},
					},
				},
			},
			expectedName: ".flytegen.simple_task",
			typedInterface: &core.TypedInterface{
				Inputs: &core.VariableMap{
					Variables: map[string]*core.Variable{
						"a": {
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
						"b": {
							Type: &core.LiteralType{
								Type: &core.LiteralType_Simple{
									Simple: core.SimpleType_INTEGER,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "TaskNode - set name",
			node: &core.Node{
				Id: "node_id",
				Metadata: &core.NodeMetadata{
					Name: "simple_task",
				},
				Target: &core.Node_TaskNode{
					TaskNode: &core.TaskNode{
						Reference: &core.TaskNode_ReferenceId{
							ReferenceId: &core.Identifier{Name: "simple_task"},
						},
					},
				},
			},
			workflowName: workflowName,
			expectedName: fmt.Sprintf(systemNamePrefix, workflowName),
			typedInterface: &core.TypedInterface{
				Inputs: &core.VariableMap{
					Variables: map[string]*core.Variable{
						"a": {
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
						"b": {
							Type: &core.LiteralType{
								Type: &core.LiteralType_Simple{
									Simple: core.SimpleType_INTEGER,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "ArrayNode with TaskNode",
			node: &core.Node{
				Id: "node_id",
				Metadata: &core.NodeMetadata{
					Name: "simple_task",
				},
				Target: &core.Node_ArrayNode{
					ArrayNode: &core.ArrayNode{
						Node: &core.Node{
							Id: "subnodeID",
							Target: &core.Node_TaskNode{
								TaskNode: &core.TaskNode{
									Reference: &core.TaskNode_ReferenceId{
										ReferenceId: &core.Identifier{Name: "ref_1"},
									},
								},
							},
						},
					},
				},
			},
			expectedName: ".flytegen.subnodeID",
			typedInterface: &core.TypedInterface{
				Inputs: &core.VariableMap{
					Variables: map[string]*core.Variable{
						"a": {
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
				},
				Outputs: &core.VariableMap{
					Variables: map[string]*core.Variable{
						"b": {
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
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			taskIdentifier := &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      wfIdentifier.Project,
				Domain:       wfIdentifier.Domain,
				Name:         "simple_task",
				Version:      wfIdentifier.Version,
				Org:          wfIdentifier.Org,
			}
			repository.TaskRepo().(*repositoryMocks.MockTaskRepo).SetGetCallback(
				func(input interfaces.Identifier) (models.Task, error) {
					createdAt := time.Now()
					createdAtProto, _ := ptypes.TimestampProto(createdAt)
					taskClosure := &admin.TaskClosure{
						CompiledTask: &core.CompiledTask{
							Template: &core.TaskTemplate{
								Id:   taskIdentifier,
								Type: "python-task",
								Metadata: &core.TaskMetadata{
									Runtime: &core.RuntimeMetadata{
										Type:    core.RuntimeMetadata_FLYTE_SDK,
										Version: "0.6.2",
										Flavor:  "python",
									},
									Timeout: ptypes.DurationProto(time.Second),
								},
								Interface: tt.typedInterface,
								Custom:    nil,
							},
						},
						CreatedAt: createdAtProto,
					}
					serializedTaskClosure, err := proto.Marshal(taskClosure)
					assert.NoError(t, err)
					return models.Task{
						TaskKey: models.TaskKey{
							Project: taskIdentifier.Project,
							Domain:  taskIdentifier.Domain,
							Name:    taskIdentifier.Name,
							Version: taskIdentifier.Version,
							Org:     taskIdentifier.Org,
						},
						Closure: serializedTaskClosure,
						Digest:  []byte("simple_task"),
						Type:    "python",
					}, nil
				})

			mockWorkflowManager := managerMocks.MockWorkflowManager{}
			mockWorkflowManager.SetCreateCallback(func(ctx context.Context, request admin.WorkflowCreateRequest) (*admin.WorkflowCreateResponse, error) {
				assert.True(t, proto.Equal(request.Id, &core.Identifier{
					ResourceType: core.ResourceType_WORKFLOW,
					Project:      wfIdentifier.Project,
					Domain:       wfIdentifier.Domain,
					Name:         tt.expectedName,
					Version:      wfIdentifier.Version,
					Org:          wfIdentifier.Org,
				}), fmt.Sprintf("%+v", request.Id))
				assert.True(t, proto.Equal(tt.typedInterface, request.Spec.Template.Interface))
				nodes := request.GetSpec().GetTemplate().GetNodes()
				assert.Equal(t, 1, len(nodes))
				node := nodes[0]
				assert.Equal(t, tt.node, node)
				return &admin.WorkflowCreateResponse{}, nil
			})
			mockNamedEntityManager := managerMocks.NamedEntityManager{}
			mockNamedEntityManager.UpdateNamedEntityFunc = func(ctx context.Context, request admin.NamedEntityUpdateRequest) (*admin.NamedEntityUpdateResponse, error) {
				assert.Equal(t, request.ResourceType, core.ResourceType_WORKFLOW)
				assert.True(t, proto.Equal(request.Id, &admin.NamedEntityIdentifier{
					Project: wfIdentifier.Project,
					Domain:  wfIdentifier.Domain,
					Name:    tt.expectedName,
					Org:     wfIdentifier.Org,
				}), fmt.Sprintf("%+v", request.Id))
				assert.True(t, proto.Equal(request.Metadata, &admin.NamedEntityMetadata{
					State: admin.NamedEntityState_SYSTEM_GENERATED,
				}))
				return &admin.NamedEntityUpdateResponse{}, nil
			}

			workflowModel, err := CreateOrGetWorkflowFromNode(context.Background(), tt.node, repository, &mockWorkflowManager, &mockNamedEntityManager, wfIdentifier, tt.workflowName)
			if tt.expectError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, workflowModel)
			}
		})
	}
}
