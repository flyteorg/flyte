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

func TestUpdateProjectAttributes(t *testing.T) {
	request := admin.ProjectAttributesUpdateRequest{
		Attributes: &admin.ProjectAttributes{
			Project:            "project",
			MatchingAttributes: testutils.ExecutionQueueAttributes,
		},
	}
	db := mocks.NewMockRepository()
	expectedSerializedAttrs, _ := proto.Marshal(testutils.ExecutionQueueAttributes)
	var createOrUpdateCalled bool
	db.ProjectAttributesRepo().(*mocks.MockProjectAttributesRepo).CreateOrUpdateFunction = func(
		ctx context.Context, input models.ProjectAttributes) error {
		assert.Equal(t, "project", input.Project)
		assert.Equal(t, admin.MatchableResource_EXECUTION_QUEUE.String(), input.Resource)
		assert.EqualValues(t, expectedSerializedAttrs, input.Attributes)
		createOrUpdateCalled = true
		return nil
	}
	manager := NewProjectAttributesManager(db)
	_, err := manager.UpdateProjectAttributes(context.Background(), request)
	assert.Nil(t, err)
	assert.True(t, createOrUpdateCalled)
}
