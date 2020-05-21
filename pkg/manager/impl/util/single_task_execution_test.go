package util

import (
	"context"
	"fmt"
	"testing"

	"github.com/lyft/flyteadmin/pkg/manager/impl/testutils"

	runtimeMocks "github.com/lyft/flyteadmin/pkg/runtime/mocks"

	"github.com/golang/protobuf/proto"
	managerMocks "github.com/lyft/flyteadmin/pkg/manager/mocks"
	"github.com/lyft/flyteadmin/pkg/repositories/interfaces"
	repositoryMocks "github.com/lyft/flyteadmin/pkg/repositories/mocks"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"

	flyteAdminErrors "github.com/lyft/flyteadmin/pkg/errors"
)

func TestGenerateNodeNameFromTask(t *testing.T) {
	assert.EqualValues(t, "foo-12345", generateNodeNameFromTask("foo@`!=+- 1,2$3(4)5"))
	assert.EqualValues(t, "nametoadefinedtasksomewhereincodeveryverynestedcrazy",
		generateNodeNameFromTask("app.long.path.name.to.a.defined.task.somewhere.in.code.very.very.nested.crazy"))
}

func TestGenerateWorkflowNameFromTask(t *testing.T) {
	assert.EqualValues(t, ".flytegen.SingleTask", generateWorkflowNameFromTask("SingleTask"))
}

func TestGetBinding(t *testing.T) {
	t.Run("scalar", func(t *testing.T) {
		primitiveValue := &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Primitive{
					Primitive: &core.Primitive{
						Value: &core.Primitive_StringValue{
							StringValue: "value",
						},
					},
				},
			},
		}
		assert.True(t, proto.Equal(&core.BindingData{
			Value: &core.BindingData_Scalar{
				Scalar: primitiveValue.Scalar,
			},
		}, getBinding(&core.Literal{
			Value: primitiveValue,
		})))
	})

	t.Run("collection", func(t *testing.T) {
		collection := &core.LiteralCollection{
			Literals: []*core.Literal{
				{
					Value: &core.Literal_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Primitive{
								Primitive: &core.Primitive{
									Value: &core.Primitive_Integer{
										Integer: 1,
									},
								},
							},
						},
					},
				},
				{
					Value: &core.Literal_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Primitive{
								Primitive: &core.Primitive{
									Value: &core.Primitive_Integer{
										Integer: 1,
									},
								},
							},
						},
					},
				},
			},
		}
		assert.True(t, proto.Equal(&core.BindingData{
			Value: &core.BindingData_Collection{
				Collection: &core.BindingDataCollection{
					Bindings: []*core.BindingData{
						{
							Value: &core.BindingData_Scalar{
								Scalar: collection.Literals[0].GetScalar(),
							},
						},
						{
							Value: &core.BindingData_Scalar{
								Scalar: collection.Literals[1].GetScalar(),
							},
						},
					},
				},
			},
		}, getBinding(&core.Literal{
			Value: &core.Literal_Collection{
				Collection: collection,
			},
		})))
	})
	t.Run("map", func(t *testing.T) {
		literalMap := &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"L": {
					Value: &core.Literal_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Primitive{
								Primitive: &core.Primitive{
									Value: &core.Primitive_FloatValue{
										FloatValue: 1.00,
									},
								},
							},
						},
					},
				},
				"C": {
					Value: &core.Literal_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Primitive{
								Primitive: &core.Primitive{
									Value: &core.Primitive_FloatValue{
										FloatValue: 2.00,
									},
								},
							},
						},
					},
				},
			},
		}
		assert.True(t, proto.Equal(&core.BindingData{
			Value: &core.BindingData_Map{
				Map: &core.BindingDataMap{
					Bindings: map[string]*core.BindingData{
						"L": {
							Value: &core.BindingData_Scalar{
								Scalar: literalMap.Literals["L"].GetScalar(),
							},
						},
						"C": {
							Value: &core.BindingData_Scalar{
								Scalar: literalMap.Literals["C"].GetScalar(),
							},
						},
					},
				},
			},
		}, getBinding(&core.Literal{
			Value: &core.Literal_Map{
				Map: literalMap,
			},
		})))
	})
}

func TestGenerateBindingsFromOutputs(t *testing.T) {
	nodeID := "nodeID"
	outputs := core.VariableMap{
		Variables: map[string]*core.Variable{
			"output1": {},
			"output2": {},
		},
	}
	generatedBindings := generateBindingsFromOutputs(outputs, nodeID)
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

func TestGenerateBindingsFromInputs(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		inputTemplate := core.VariableMap{
			Variables: map[string]*core.Variable{
				"simple": {
					Type: &core.LiteralType{
						Type: &core.LiteralType_Simple{
							Simple: core.SimpleType_INTEGER,
						},
					},
				},
			},
		}
		inputs := core.LiteralMap{
			Literals: map[string]*core.Literal{
				"simple": {
					Value: &core.Literal_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Primitive{
								Primitive: &core.Primitive{
									Value: &core.Primitive_Integer{
										Integer: 999,
									},
								},
							},
						},
					},
				},
			},
		}
		bindings, err := generateBindingsFromInputs(inputTemplate, inputs)
		assert.NoError(t, err)
		assert.EqualValues(t, []*core.Binding{
			{
				Var: "simple",
				Binding: &core.BindingData{
					Value: &core.BindingData_Scalar{
						Scalar: inputs.Literals["simple"].GetScalar(),
					},
				},
			},
		}, bindings)
	})
	t.Run("schema", func(t *testing.T) {
		inputTemplate := core.VariableMap{
			Variables: map[string]*core.Variable{
				"schema": {
					Type: &core.LiteralType{
						Type: &core.LiteralType_Schema{
							Schema: &core.SchemaType{
								Columns: []*core.SchemaType_SchemaColumn{
									{
										Name: "float column",
										Type: core.SchemaType_SchemaColumn_FLOAT,
									},
									{
										Name: "string column",
										Type: core.SchemaType_SchemaColumn_STRING,
									},
								},
							},
						},
					},
				},
			},
		}
		inputs := core.LiteralMap{
			Literals: map[string]*core.Literal{
				"schema": {
					Value: &core.Literal_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Schema{
								Schema: &core.Schema{
									Uri: "s3://flyte/myschema",
									Type: &core.SchemaType{
										Columns: []*core.SchemaType_SchemaColumn{
											{
												Name: "float column",
												Type: core.SchemaType_SchemaColumn_FLOAT,
											},
											{
												Name: "string column",
												Type: core.SchemaType_SchemaColumn_STRING,
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
		bindings, err := generateBindingsFromInputs(inputTemplate, inputs)
		assert.NoError(t, err)
		assert.EqualValues(t, []*core.Binding{
			{
				Var: "schema",
				Binding: &core.BindingData{
					Value: &core.BindingData_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Schema{
								Schema: &core.Schema{
									Uri: "s3://flyte/myschema",
									Type: &core.SchemaType{
										Columns: []*core.SchemaType_SchemaColumn{
											{
												Name: "float column",
												Type: core.SchemaType_SchemaColumn_FLOAT,
											},
											{
												Name: "string column",
												Type: core.SchemaType_SchemaColumn_STRING,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}, bindings)
	})
	t.Run("collection", func(t *testing.T) {
		inputTemplate := core.VariableMap{
			Variables: map[string]*core.Variable{
				"collection": {
					Type: &core.LiteralType{
						Type: &core.LiteralType_CollectionType{
							CollectionType: &core.LiteralType{
								Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
							},
						},
					},
				},
			},
		}
		inputs := core.LiteralMap{
			Literals: map[string]*core.Literal{
				"collection": {
					Value: &core.Literal_Collection{
						Collection: &core.LiteralCollection{
							Literals: []*core.Literal{
								{
									Value: &core.Literal_Scalar{
										Scalar: &core.Scalar{
											Value: &core.Scalar_Primitive{
												Primitive: &core.Primitive{
													Value: &core.Primitive_Integer{
														Integer: 999,
													},
												},
											},
										},
									},
								},
								{
									Value: &core.Literal_Scalar{
										Scalar: &core.Scalar{
											Value: &core.Scalar_Primitive{
												Primitive: &core.Primitive{
													Value: &core.Primitive_Integer{
														Integer: 1000,
													},
												},
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
		bindings, err := generateBindingsFromInputs(inputTemplate, inputs)
		assert.NoError(t, err)
		assert.EqualValues(t, []*core.Binding{
			{
				Var: "collection",
				Binding: &core.BindingData{
					Value: &core.BindingData_Collection{
						Collection: &core.BindingDataCollection{
							Bindings: []*core.BindingData{
								{
									Value: &core.BindingData_Scalar{
										Scalar: &core.Scalar{
											Value: &core.Scalar_Primitive{
												Primitive: &core.Primitive{
													Value: &core.Primitive_Integer{
														Integer: 999,
													},
												},
											},
										},
									},
								},
								{
									Value: &core.BindingData_Scalar{
										Scalar: &core.Scalar{
											Value: &core.Scalar_Primitive{
												Primitive: &core.Primitive{
													Value: &core.Primitive_Integer{
														Integer: 1000,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}, bindings)
	})
	t.Run("map", func(t *testing.T) {
		inputTemplate := core.VariableMap{
			Variables: map[string]*core.Variable{
				"map": {
					Type: &core.LiteralType{
						Type: &core.LiteralType_MapValueType{
							MapValueType: &core.LiteralType{
								Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
							},
						},
					},
				},
			},
		}
		inputs := core.LiteralMap{
			Literals: map[string]*core.Literal{
				"map": {
					Value: &core.Literal_Map{
						Map: &core.LiteralMap{
							Literals: map[string]*core.Literal{
								"foo": {
									Value: &core.Literal_Scalar{
										Scalar: &core.Scalar{
											Value: &core.Scalar_Primitive{
												Primitive: &core.Primitive{
													Value: &core.Primitive_Integer{
														Integer: 999,
													},
												},
											},
										},
									},
								},
								"bar": {
									Value: &core.Literal_Scalar{
										Scalar: &core.Scalar{
											Value: &core.Scalar_Primitive{
												Primitive: &core.Primitive{
													Value: &core.Primitive_Integer{
														Integer: 1000,
													},
												},
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
		bindings, err := generateBindingsFromInputs(inputTemplate, inputs)
		assert.NoError(t, err)
		assert.EqualValues(t, []*core.Binding{
			{
				Var: "map",
				Binding: &core.BindingData{
					Value: &core.BindingData_Map{
						Map: &core.BindingDataMap{
							Bindings: map[string]*core.BindingData{
								"foo": {
									Value: &core.BindingData_Scalar{
										Scalar: &core.Scalar{
											Value: &core.Scalar_Primitive{
												Primitive: &core.Primitive{
													Value: &core.Primitive_Integer{
														Integer: 999,
													},
												},
											},
										},
									},
								},
								"bar": {
									Value: &core.BindingData_Scalar{
										Scalar: &core.Scalar{
											Value: &core.Scalar_Primitive{
												Primitive: &core.Primitive{
													Value: &core.Primitive_Integer{
														Integer: 1000,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}, bindings)
	})
	t.Run("blob", func(t *testing.T) {
		inputTemplate := core.VariableMap{
			Variables: map[string]*core.Variable{
				"blob": {
					Type: &core.LiteralType{
						Type: &core.LiteralType_Blob{
							Blob: &core.BlobType{
								Format: "parquet courts",
							},
						},
					},
				},
			},
		}
		inputs := core.LiteralMap{
			Literals: map[string]*core.Literal{
				"blob": {
					Value: &core.Literal_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Blob{
								Blob: &core.Blob{
									Metadata: &core.BlobMetadata{
										Type: &core.BlobType{
											Format: "parquet courts",
										},
									},
									Uri: "s3://flyte/myblob",
								},
							},
						},
					},
				},
			},
		}
		bindings, err := generateBindingsFromInputs(inputTemplate, inputs)
		assert.NoError(t, err)
		assert.EqualValues(t, []*core.Binding{
			{
				Var: "blob",
				Binding: &core.BindingData{
					Value: &core.BindingData_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Blob{
								Blob: &core.Blob{
									Metadata: &core.BlobMetadata{
										Type: &core.BlobType{
											Format: "parquet courts",
										},
									},
									Uri: "s3://flyte/myblob",
								},
							},
						},
					},
				},
			},
		}, bindings)
	})
}

func TestCreateOrGetWorkflowModel(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	var getCalledCount = 0
	var newlyCreatedWorkflow models.Workflow
	workflowcreateFunc := func(input models.Workflow) error {
		newlyCreatedWorkflow = input
		return nil
	}
	repository.WorkflowRepo().(*repositoryMocks.MockWorkflowRepo).SetCreateCallback(workflowcreateFunc)

	workflowGetFunc := func(input interfaces.GetResourceInput) (models.Workflow, error) {
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

	launchPlanGetFunc := func(input interfaces.GetResourceInput) (models.LaunchPlan, error) {
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
			Method: &admin.AuthRole_AssumableIamRole{AssumableIamRole: "assumable_role"},
		},
	}
	launchPlan, err := CreateOrGetLaunchPlan(
		context.Background(), repository, config, taskIdentifier, workflowInterface, workflowID, &spec)
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
