package validation

import (
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"
)

var testExecutionID = core.WorkflowExecutionIdentifier{
	Project: "project",
	Domain:  "domain",
	Name:    "name",
}

var dynamicWfID = core.Identifier{
	ResourceType: core.ResourceType_WORKFLOW,
	Project:      "project",
	Domain:       "domain",
	Name:         "name",
	Version:      "abc123",
}

func TestValidateNodeExecutionIdentifier(t *testing.T) {
	err := ValidateNodeExecutionIdentifier(&core.NodeExecutionIdentifier{
		ExecutionId: &testExecutionID,
		NodeId:      "node id",
	})
	assert.Nil(t, err)
}

func TestValidateNodeExecutionIdentifier_MissingFields(t *testing.T) {
	err := ValidateNodeExecutionIdentifier(&core.NodeExecutionIdentifier{
		NodeId: "node id",
	})
	assert.EqualError(t, err, "missing execution_id")

	err = ValidateNodeExecutionIdentifier(&core.NodeExecutionIdentifier{
		ExecutionId: &testExecutionID,
	})
	assert.EqualError(t, err, "missing node_id")
}

func TestValidateNodeExecutionEventRequest(t *testing.T) {
	request := admin.NodeExecutionEventRequest{
		Event: &event.NodeExecutionEvent{
			Id: &core.NodeExecutionIdentifier{
				ExecutionId: &testExecutionID,
				NodeId:      "node id",
			},
		},
	}
	assert.NoError(t, ValidateNodeExecutionEventRequest(&request, maxOutputSizeInBytes))

	request = admin.NodeExecutionEventRequest{
		Event: &event.NodeExecutionEvent{
			Id: &core.NodeExecutionIdentifier{
				ExecutionId: &testExecutionID,
				NodeId:      "node id",
			},
			TargetMetadata: &event.NodeExecutionEvent_TaskNodeMetadata{
				TaskNodeMetadata: &event.TaskNodeMetadata{
					DynamicWorkflow: &event.DynamicWorkflowNodeMetadata{
						Id: &dynamicWfID,
						CompiledWorkflow: &core.CompiledWorkflowClosure{
							Primary: &core.CompiledWorkflow{
								Template: &core.WorkflowTemplate{
									Id: &dynamicWfID,
								},
							},
						},
					},
				},
			},
		},
	}
	assert.NoError(t, ValidateNodeExecutionEventRequest(&request, maxOutputSizeInBytes))
}

func TestValidateNodeExecutionEventRequest_Invalid(t *testing.T) {
	t.Run("missing dynamic workflow id", func(t *testing.T) {
		request := admin.NodeExecutionEventRequest{
			Event: &event.NodeExecutionEvent{
				Id: &core.NodeExecutionIdentifier{
					ExecutionId: &testExecutionID,
					NodeId:      "node id",
				},
				TargetMetadata: &event.NodeExecutionEvent_TaskNodeMetadata{
					TaskNodeMetadata: &event.TaskNodeMetadata{
						DynamicWorkflow: &event.DynamicWorkflowNodeMetadata{},
					},
				},
			},
		}
		assert.EqualError(t, ValidateNodeExecutionEventRequest(&request, maxOutputSizeInBytes), "missing id")
	})

	t.Run("missing dynamic compiled workflow", func(t *testing.T) {
		request := admin.NodeExecutionEventRequest{
			Event: &event.NodeExecutionEvent{
				Id: &core.NodeExecutionIdentifier{
					ExecutionId: &testExecutionID,
					NodeId:      "node id",
				},
				TargetMetadata: &event.NodeExecutionEvent_TaskNodeMetadata{
					TaskNodeMetadata: &event.TaskNodeMetadata{
						DynamicWorkflow: &event.DynamicWorkflowNodeMetadata{
							Id: &dynamicWfID,
						},
					},
				},
			},
		}
		assert.EqualError(t, ValidateNodeExecutionEventRequest(&request, maxOutputSizeInBytes), "missing compiled dynamic workflow")
	})

	t.Run("missing dynamic compiled workflow primary", func(t *testing.T) {
		request := admin.NodeExecutionEventRequest{
			Event: &event.NodeExecutionEvent{
				Id: &core.NodeExecutionIdentifier{
					ExecutionId: &testExecutionID,
					NodeId:      "node id",
				},
				TargetMetadata: &event.NodeExecutionEvent_TaskNodeMetadata{
					TaskNodeMetadata: &event.TaskNodeMetadata{
						DynamicWorkflow: &event.DynamicWorkflowNodeMetadata{
							Id:               &dynamicWfID,
							CompiledWorkflow: &core.CompiledWorkflowClosure{},
						},
					},
				},
			},
		}
		assert.EqualError(t, ValidateNodeExecutionEventRequest(&request, maxOutputSizeInBytes), "missing primary dynamic workflow")
	})

	t.Run("missing dynamic compiled primary template", func(t *testing.T) {
		request := admin.NodeExecutionEventRequest{
			Event: &event.NodeExecutionEvent{
				Id: &core.NodeExecutionIdentifier{
					ExecutionId: &testExecutionID,
					NodeId:      "node id",
				},
				TargetMetadata: &event.NodeExecutionEvent_TaskNodeMetadata{
					TaskNodeMetadata: &event.TaskNodeMetadata{
						DynamicWorkflow: &event.DynamicWorkflowNodeMetadata{
							Id: &dynamicWfID,
							CompiledWorkflow: &core.CompiledWorkflowClosure{
								Primary: &core.CompiledWorkflow{},
							},
						},
					},
				},
			},
		}
		assert.EqualError(t, ValidateNodeExecutionEventRequest(&request, maxOutputSizeInBytes), "missing primary dynamic workflow template")
	})

	t.Run("missing dynamic compiled workflow primary identifier", func(t *testing.T) {
		request := admin.NodeExecutionEventRequest{
			Event: &event.NodeExecutionEvent{
				Id: &core.NodeExecutionIdentifier{
					ExecutionId: &testExecutionID,
					NodeId:      "node id",
				},
				TargetMetadata: &event.NodeExecutionEvent_TaskNodeMetadata{
					TaskNodeMetadata: &event.TaskNodeMetadata{
						DynamicWorkflow: &event.DynamicWorkflowNodeMetadata{
							Id: &dynamicWfID,
							CompiledWorkflow: &core.CompiledWorkflowClosure{
								Primary: &core.CompiledWorkflow{
									Template: &core.WorkflowTemplate{
										Id: &core.Identifier{},
									},
								},
							},
						},
					},
				},
			},
		}
		assert.EqualError(t, ValidateNodeExecutionEventRequest(&request, maxOutputSizeInBytes), "unexpected resource type unspecified for identifier [], expected workflow instead")
	})
}

func TestValidateNodeExecutionListRequest(t *testing.T) {
	err := ValidateNodeExecutionListRequest(admin.NodeExecutionListRequest{
		WorkflowExecutionId: &testExecutionID,
		Filters:             "foo",
		Limit:               2,
	})
	assert.Nil(t, err)
}

func TestValidateNodeExecutionListRequest_MissingFields(t *testing.T) {
	err := ValidateNodeExecutionListRequest(admin.NodeExecutionListRequest{
		Limit: 2,
	})
	assert.EqualError(t, err, "missing execution_id")

	err = ValidateNodeExecutionListRequest(admin.NodeExecutionListRequest{
		WorkflowExecutionId: &testExecutionID,
	})
	assert.EqualError(t, err, "invalid value for limit")
}

func TestValidateNodeExecutionForTaskListRequest(t *testing.T) {
	err := ValidateNodeExecutionForTaskListRequest(admin.NodeExecutionForTaskListRequest{
		TaskExecutionId: &core.TaskExecutionIdentifier{
			NodeExecutionId: &core.NodeExecutionIdentifier{
				ExecutionId: &testExecutionID,
				NodeId:      "nodey",
			},
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      "project",
				Domain:       "domain",
				Name:         "name",
				Version:      "version",
			},
		},
		Filters: "foo",
		Limit:   2,
	})
	assert.Nil(t, err)
}

func TestValidateNodeExecutionForTaskListRequest_MissingFields(t *testing.T) {
	err := ValidateNodeExecutionForTaskListRequest(admin.NodeExecutionForTaskListRequest{
		TaskExecutionId: &core.TaskExecutionIdentifier{
			NodeExecutionId: &core.NodeExecutionIdentifier{
				ExecutionId: &testExecutionID,
				NodeId:      "nodey",
			},
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      "project",
				Domain:       "domain",
				Name:         "name",
				Version:      "version",
			},
		},
	})
	assert.EqualError(t, err, "invalid value for limit")

	err = ValidateNodeExecutionForTaskListRequest(admin.NodeExecutionForTaskListRequest{
		TaskExecutionId: &core.TaskExecutionIdentifier{
			NodeExecutionId: &core.NodeExecutionIdentifier{
				ExecutionId: &testExecutionID,
			},
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      "project",
				Domain:       "domain",
				Name:         "name",
				Version:      "version",
			},
		},
		Limit: 2,
	})
	assert.EqualError(t, err, "missing node_id")

	err = ValidateNodeExecutionForTaskListRequest(admin.NodeExecutionForTaskListRequest{
		TaskExecutionId: &core.TaskExecutionIdentifier{
			NodeExecutionId: &core.NodeExecutionIdentifier{
				ExecutionId: &testExecutionID,
				NodeId:      "nodey",
			},
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Domain:       "domain",
				Name:         "name",
				Version:      "version",
			},
		},
		Limit: 2,
	})
	assert.EqualError(t, err, "missing project")
}
