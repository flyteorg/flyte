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

func TestUpdateProjectDomainAttributes(t *testing.T) {
	request := admin.ProjectDomainAttributesUpdateRequest{
		Attributes: &admin.ProjectDomainAttributes{
			Project:            "project",
			Domain:             "domain",
			MatchingAttributes: testutils.ExecutionQueueAttributes,
		},
	}
	db := mocks.NewMockRepository()
	expectedSerializedAttrs, _ := proto.Marshal(testutils.ExecutionQueueAttributes)
	var createOrUpdateCalled bool
	db.ProjectDomainAttributesRepo().(*mocks.MockProjectDomainAttributesRepo).CreateOrUpdateFunction = func(
		ctx context.Context, input models.ProjectDomainAttributes) error {
		assert.Equal(t, "project", input.Project)
		assert.Equal(t, "domain", input.Domain)
		assert.Equal(t, admin.MatchableResource_EXECUTION_QUEUE.String(), input.Resource)
		assert.EqualValues(t, expectedSerializedAttrs, input.Attributes)
		createOrUpdateCalled = true
		return nil
	}
	manager := NewProjectDomainAttributesManager(db)
	_, err := manager.UpdateProjectDomainAttributes(context.Background(), request)
	assert.Nil(t, err)
	assert.True(t, createOrUpdateCalled)
}

func TestGetProjectDomainAttributes(t *testing.T) {
	request := admin.ProjectDomainAttributesGetRequest{
		Project:      "project",
		Domain:       "domain",
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE,
	}
	db := mocks.NewMockRepository()
	db.ProjectDomainAttributesRepo().(*mocks.MockProjectDomainAttributesRepo).GetFunction = func(
		ctx context.Context, project, domain, resource string) (models.ProjectDomainAttributes, error) {
		assert.Equal(t, "project", project)
		assert.Equal(t, "domain", domain)
		assert.Equal(t, admin.MatchableResource_EXECUTION_QUEUE.String(), resource)
		expectedSerializedAttrs, _ := proto.Marshal(testutils.ExecutionQueueAttributes)
		return models.ProjectDomainAttributes{
			Project:    project,
			Domain:     domain,
			Resource:   resource,
			Attributes: expectedSerializedAttrs,
		}, nil
	}
	manager := NewProjectDomainAttributesManager(db)
	response, err := manager.GetProjectDomainAttributes(context.Background(), request)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&admin.ProjectDomainAttributesGetResponse{
		Attributes: &admin.ProjectDomainAttributes{
			Project:            "project",
			Domain:             "domain",
			MatchingAttributes: testutils.ExecutionQueueAttributes,
		},
	}, response))
}

func TestDeleteProjectDomainAttributes(t *testing.T) {
	request := admin.ProjectDomainAttributesDeleteRequest{
		Project:      "project",
		Domain:       "domain",
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE,
	}
	db := mocks.NewMockRepository()
	db.ProjectDomainAttributesRepo().(*mocks.MockProjectDomainAttributesRepo).DeleteFunction = func(
		ctx context.Context, project, domain, resource string) error {
		assert.Equal(t, "project", project)
		assert.Equal(t, "domain", domain)
		assert.Equal(t, admin.MatchableResource_EXECUTION_QUEUE.String(), resource)
		return nil
	}
	manager := NewProjectDomainAttributesManager(db)
	_, err := manager.DeleteProjectDomainAttributes(context.Background(), request)
	assert.Nil(t, err)
}
