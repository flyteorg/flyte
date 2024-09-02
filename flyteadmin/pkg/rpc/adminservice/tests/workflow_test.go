package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/mocks"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/utils"
)

var workflowIdentifier = core.Identifier{
	ResourceType: core.ResourceType_WORKFLOW,
	Name:         "Name",
	Domain:       "Domain",
	Project:      "Project",
	Version:      "Version",
}

func TestCreateWorkflowHappyCase(t *testing.T) {
	ctx := context.Background()

	mockWorkflowManager := mocks.MockWorkflowManager{}
	mockWorkflowManager.SetCreateCallback(
		func(ctx context.Context,
			request *admin.WorkflowCreateRequest) (*admin.WorkflowCreateResponse, error) {
			return &admin.WorkflowCreateResponse{}, nil
		},
	)
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		workflowManager: &mockWorkflowManager,
	})

	resp, err := mockServer.CreateWorkflow(ctx, &admin.WorkflowCreateRequest{
		Id: &workflowIdentifier,
	})
	assert.NotNil(t, resp)
	assert.NoError(t, err)
}

func TestCreateWorkflowError(t *testing.T) {
	ctx := context.Background()

	mockWorkflowManager := mocks.MockWorkflowManager{}
	mockWorkflowManager.SetCreateCallback(
		func(ctx context.Context,
			request *admin.WorkflowCreateRequest) (*admin.WorkflowCreateResponse, error) {
			return nil, errors.GetMissingEntityError(core.ResourceType_WORKFLOW.String(), request.Id)
		},
	)
	mockServer := NewMockAdminServer(NewMockAdminServerInput{
		workflowManager: &mockWorkflowManager,
	})

	resp, err := mockServer.CreateWorkflow(ctx, &admin.WorkflowCreateRequest{
		Id: &workflowIdentifier,
	})
	assert.Nil(t, resp)
	utils.AssertEqualWithSanitizedRegex(t, "missing entity of type WORKFLOW with "+
		"identifier resource_type:WORKFLOW project:\"Project\" domain:\"Domain\" name:\"Name\" version:\"Version\"", err.Error())
}
