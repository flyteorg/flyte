package impl

import (
	"context"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/lyft/flyteadmin/pkg/manager/impl/testutils"
	"github.com/lyft/flyteadmin/pkg/repositories/mocks"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"
)

func TestUpdateWorkflowAttributes(t *testing.T) {
	request := admin.WorkflowAttributesUpdateRequest{
		Attributes: &admin.WorkflowAttributes{
			Project:            "project",
			Domain:             "domain",
			Workflow:           "workflow",
			MatchingAttributes: testutils.ExecutionQueueAttributes,
		},
	}
	db := mocks.NewMockRepository()
	expectedSerializedAttrs, _ := proto.Marshal(testutils.ExecutionQueueAttributes)
	var createOrUpdateCalled bool
	db.WorkflowAttributesRepo().(*mocks.MockWorkflowAttributesRepo).CreateOrUpdateFunction = func(
		ctx context.Context, input models.WorkflowAttributes) error {
		assert.Equal(t, "project", input.Project)
		assert.Equal(t, "domain", input.Domain)
		assert.Equal(t, "workflow", input.Workflow)
		assert.Equal(t, admin.MatchableResource_EXECUTION_QUEUE.String(), input.Resource)
		assert.EqualValues(t, expectedSerializedAttrs, input.Attributes)
		createOrUpdateCalled = true
		return nil
	}
	manager := NewWorkflowAttributesManager(db)
	_, err := manager.UpdateWorkflowAttributes(context.Background(), request)
	assert.Nil(t, err)
	assert.True(t, createOrUpdateCalled)
}

func TestGetWorkflowAttributes(t *testing.T) {
	request := admin.WorkflowAttributesGetRequest{
		Project:      "project",
		Domain:       "domain",
		Workflow:     "workflow",
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE,
	}
	db := mocks.NewMockRepository()
	db.WorkflowAttributesRepo().(*mocks.MockWorkflowAttributesRepo).GetFunction = func(
		ctx context.Context, project, domain, workflow, resource string) (models.WorkflowAttributes, error) {
		assert.Equal(t, "project", project)
		assert.Equal(t, "domain", domain)
		assert.Equal(t, "workflow", workflow)
		assert.Equal(t, admin.MatchableResource_EXECUTION_QUEUE.String(), resource)
		expectedSerializedAttrs, _ := proto.Marshal(testutils.ExecutionQueueAttributes)
		return models.WorkflowAttributes{
			Project:    project,
			Domain:     domain,
			Workflow:   workflow,
			Resource:   resource,
			Attributes: expectedSerializedAttrs,
		}, nil
	}
	manager := NewWorkflowAttributesManager(db)
	response, err := manager.GetWorkflowAttributes(context.Background(), request)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&admin.WorkflowAttributesGetResponse{
		Attributes: &admin.WorkflowAttributes{
			Project:            "project",
			Domain:             "domain",
			Workflow:           "workflow",
			MatchingAttributes: testutils.ExecutionQueueAttributes,
		},
	}, response))
}

func TestDeleteWorkflowAttributes(t *testing.T) {
	request := admin.WorkflowAttributesDeleteRequest{
		Project:      "project",
		Domain:       "domain",
		Workflow:     "workflow",
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE,
	}
	db := mocks.NewMockRepository()
	db.WorkflowAttributesRepo().(*mocks.MockWorkflowAttributesRepo).DeleteFunction = func(
		ctx context.Context, project, domain, workflow, resource string) error {
		assert.Equal(t, "project", project)
		assert.Equal(t, "domain", domain)
		assert.Equal(t, "workflow", workflow)
		assert.Equal(t, admin.MatchableResource_EXECUTION_QUEUE.String(), resource)
		return nil
	}
	manager := NewWorkflowAttributesManager(db)
	_, err := manager.DeleteWorkflowAttributes(context.Background(), request)
	assert.Nil(t, err)
}
