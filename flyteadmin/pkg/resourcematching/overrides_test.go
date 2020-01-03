package resourcematching

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteadmin/pkg/repositories/mocks"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

const testProject = "project"
const testDomain = "domain"
const testWorkflow = "workflow"

func TestIsNotFoundErr(t *testing.T) {
	isNotFound := errors.NewFlyteAdminError(codes.NotFound, "foo")
	assert.True(t, isNotFoundErr(isNotFound))

	invalidArgs := errors.NewFlyteAdminErrorf(codes.InvalidArgument, "bar")
	assert.False(t, isNotFoundErr(invalidArgs))
}

func TestGetOverrideValuesToApply(t *testing.T) {
	db := mocks.NewMockRepository()
	matchingWorkflowAttributes := &admin.MatchingAttributes{
		Target: &admin.MatchingAttributes_ExecutionQueueAttributes{
			ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{
				Tags: []string{"attr3"},
			},
		},
	}
	db.WorkflowAttributesRepo().(*mocks.MockWorkflowAttributesRepo).GetFunction = func(
		ctx context.Context, project, domain, workflow, resource string) (
		models.WorkflowAttributes, error) {
		if project == testProject && domain == testDomain && workflow == testWorkflow &&
			resource == admin.MatchableResource_EXECUTION_QUEUE.String() {

			marshalledMatchingAttributes, _ := proto.Marshal(matchingWorkflowAttributes)
			return models.WorkflowAttributes{
				Project:    project,
				Domain:     domain,
				Workflow:   workflow,
				Resource:   resource,
				Attributes: marshalledMatchingAttributes,
			}, nil
		}
		if workflow == "error" {
			return models.WorkflowAttributes{}, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "bar")
		}
		return models.WorkflowAttributes{}, errors.NewFlyteAdminError(codes.NotFound, "foo")
	}
	matchingProjectDomainAttributes := &admin.MatchingAttributes{
		Target: &admin.MatchingAttributes_ExecutionQueueAttributes{
			ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{
				Tags: []string{"attr2"},
			},
		},
	}
	db.ProjectDomainAttributesRepo().(*mocks.MockProjectDomainAttributesRepo).GetFunction = func(
		ctx context.Context, project, domain, resource string) (models.ProjectDomainAttributes, error) {
		if project == testProject && domain == testDomain && resource == admin.MatchableResource_EXECUTION_QUEUE.String() {
			marshalledMatchingAttributes, _ := proto.Marshal(matchingProjectDomainAttributes)
			return models.ProjectDomainAttributes{
				Project:    project,
				Domain:     domain,
				Resource:   resource,
				Attributes: marshalledMatchingAttributes,
			}, nil
		}
		return models.ProjectDomainAttributes{}, errors.NewFlyteAdminError(codes.NotFound, "foo")
	}
	matchingProjectAttributes := &admin.MatchingAttributes{
		Target: &admin.MatchingAttributes_ExecutionQueueAttributes{
			ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{
				Tags: []string{"attr1"},
			},
		},
	}
	db.ProjectAttributesRepo().(*mocks.MockProjectAttributesRepo).GetFunction = func(
		ctx context.Context, project, resource string) (models.ProjectAttributes, error) {
		if project == testProject && resource == admin.MatchableResource_EXECUTION_QUEUE.String() {
			marshalledMatchingAttributes, _ := proto.Marshal(matchingProjectAttributes)
			return models.ProjectAttributes{
				Project:    project,
				Resource:   resource,
				Attributes: marshalledMatchingAttributes,
			}, nil
		}
		return models.ProjectAttributes{}, errors.NewFlyteAdminError(codes.NotFound, "foo")
	}

	testCases := []struct {
		input                      GetOverrideValuesInput
		expectedMatchingAttributes *admin.MatchingAttributes
		expectedErr                error
	}{
		{
			GetOverrideValuesInput{
				Db:       db,
				Project:  "project",
				Domain:   "domain",
				Workflow: "workflow",
				Resource: admin.MatchableResource_EXECUTION_QUEUE,
			},
			matchingWorkflowAttributes,
			nil,
		},
		{
			GetOverrideValuesInput{
				Db:       db,
				Project:  "project",
				Domain:   "domain",
				Workflow: "workflow2",
				Resource: admin.MatchableResource_EXECUTION_QUEUE,
			},
			matchingProjectDomainAttributes,
			nil,
		},
		{
			GetOverrideValuesInput{
				Db:       db,
				Project:  "project",
				Domain:   "domain2",
				Workflow: "workflow",
				Resource: admin.MatchableResource_EXECUTION_QUEUE,
			},
			matchingProjectAttributes,
			nil,
		},
		{
			GetOverrideValuesInput{
				Db:       db,
				Project:  "project2",
				Domain:   "domain",
				Workflow: "workflow",
				Resource: admin.MatchableResource_EXECUTION_QUEUE,
			},
			nil,
			nil,
		},
		{GetOverrideValuesInput{
			Db:       db,
			Project:  "project",
			Domain:   "domain",
			Workflow: "error",
			Resource: admin.MatchableResource_EXECUTION_QUEUE,
		},
			nil,
			errors.NewFlyteAdminErrorf(codes.InvalidArgument, "bar"),
		},
	}
	for _, tc := range testCases {
		matchingAttributes, err := GetOverrideValuesToApply(context.Background(), tc.input)
		assert.True(t, proto.Equal(tc.expectedMatchingAttributes, matchingAttributes),
			fmt.Sprintf("invalid value for [%+v]", tc.input))
		assert.EqualValues(t, tc.expectedErr, err)
	}
}
