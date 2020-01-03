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
