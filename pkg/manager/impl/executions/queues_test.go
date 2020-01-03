package executions

import (
	"context"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteadmin/pkg/repositories/mocks"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"
	"google.golang.org/grpc/codes"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"

	runtimeInterfaces "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	runtimeMocks "github.com/lyft/flyteadmin/pkg/runtime/mocks"
	"github.com/stretchr/testify/assert"
)

const testProject = "project"
const testDomain = "domain"
const testWorkflow = "name"

func TestGetQueue(t *testing.T) {
	executionQueues := []runtimeInterfaces.ExecutionQueue{
		{
			Primary:    "queue primary",
			Dynamic:    "queue dynamic",
			Attributes: []string{"attribute"},
		},
	}
	db := mocks.NewMockRepository()
	db.WorkflowAttributesRepo().(*mocks.MockWorkflowAttributesRepo).GetFunction = func(
		ctx context.Context, project, domain, workflow, resource string) (
		models.WorkflowAttributes, error) {
		response := models.WorkflowAttributes{
			Project:  project,
			Domain:   domain,
			Workflow: workflow,
			Resource: resource,
		}
		if project == testProject && domain == testDomain && workflow == testWorkflow {
			matchingAttributes := &admin.MatchingAttributes{
				Target: &admin.MatchingAttributes_ExecutionQueueAttributes{
					ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{
						Tags: []string{"attribute"},
					},
				},
			}
			marshalledMatchingAttributes, _ := proto.Marshal(matchingAttributes)
			response.Attributes = marshalledMatchingAttributes
		} else {
			matchingAttributes := &admin.MatchingAttributes{
				Target: &admin.MatchingAttributes_ExecutionQueueAttributes{
					ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{
						Tags: []string{"another attribute"},
					},
				},
			}
			marshalledMatchingAttributes, _ := proto.Marshal(matchingAttributes)
			response.Attributes = marshalledMatchingAttributes
		}
		return response, nil
	}

	queueAllocator := NewQueueAllocator(runtimeMocks.NewMockConfigurationProvider(
		nil, runtimeMocks.NewMockQueueConfigurationProvider(executionQueues, nil),
		nil, nil, nil, nil), db)
	queueConfig := singleQueueConfiguration{
		PrimaryQueue: "queue primary",
		DynamicQueue: "queue dynamic",
	}
	assert.Equal(t, queueConfig, queueAllocator.GetQueue(context.Background(), core.Identifier{
		Project: "project",
		Domain:  "domain",
		Name:    "name",
	}))
	assert.EqualValues(t, singleQueueConfiguration{}, queueAllocator.GetQueue(context.Background(), core.Identifier{
		Project: "project",
		Domain:  "domain",
		Name:    "name2",
	}))
	assert.EqualValues(t, singleQueueConfiguration{}, queueAllocator.GetQueue(context.Background(), core.Identifier{
		Project: "project",
		Domain:  "domain2",
		Name:    "name",
	}))
	assert.EqualValues(t, singleQueueConfiguration{}, queueAllocator.GetQueue(context.Background(), core.Identifier{
		Project: "project2",
		Domain:  "domain",
		Name:    "name",
	}))
}

func TestGetQueueDefaults(t *testing.T) {
	executionQueues := []runtimeInterfaces.ExecutionQueue{
		{
			Primary:    "queue1 primary",
			Dynamic:    "queue1 dynamic",
			Attributes: []string{"attr1"},
		},
		{
			Primary:    "queue2 primary",
			Dynamic:    "queue2 dynamic",
			Attributes: []string{"attr2"},
		},
		{
			Primary:    "queue3 primary",
			Dynamic:    "queue3 dynamic",
			Attributes: []string{"attr3"},
		},
		{
			Primary:    "default primary",
			Dynamic:    "default dynamic",
			Attributes: []string{"default"},
		},
	}
	workflowConfigs := []runtimeInterfaces.WorkflowConfig{
		{
			Tags: []string{"default"},
		},
	}
	db := mocks.NewMockRepository()
	db.WorkflowAttributesRepo().(*mocks.MockWorkflowAttributesRepo).GetFunction = func(
		ctx context.Context, project, domain, workflow, resource string) (
		models.WorkflowAttributes, error) {
		if project == testProject && domain == testDomain && workflow == "workflow" &&
			resource == admin.MatchableResource_EXECUTION_QUEUE.String() {
			matchingAttributes := &admin.MatchingAttributes{
				Target: &admin.MatchingAttributes_ExecutionQueueAttributes{
					ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{
						Tags: []string{"attr3"},
					},
				},
			}
			marshalledMatchingAttributes, _ := proto.Marshal(matchingAttributes)
			return models.WorkflowAttributes{
				Project:    project,
				Domain:     domain,
				Workflow:   workflow,
				Resource:   resource,
				Attributes: marshalledMatchingAttributes,
			}, nil
		}
		return models.WorkflowAttributes{}, errors.NewFlyteAdminError(codes.NotFound, "foo")
	}
	db.ProjectDomainAttributesRepo().(*mocks.MockProjectDomainAttributesRepo).GetFunction = func(
		ctx context.Context, project, domain, resource string) (models.ProjectDomainAttributes, error) {
		if project == testProject && domain == testDomain && resource == admin.MatchableResource_EXECUTION_QUEUE.String() {
			matchingAttributes := &admin.MatchingAttributes{
				Target: &admin.MatchingAttributes_ExecutionQueueAttributes{
					ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{
						Tags: []string{"attr2"},
					},
				},
			}
			marshalledMatchingAttributes, _ := proto.Marshal(matchingAttributes)
			return models.ProjectDomainAttributes{
				Project:    project,
				Domain:     domain,
				Resource:   resource,
				Attributes: marshalledMatchingAttributes,
			}, nil
		}
		return models.ProjectDomainAttributes{}, errors.NewFlyteAdminError(codes.NotFound, "foo")
	}
	db.ProjectAttributesRepo().(*mocks.MockProjectAttributesRepo).GetFunction = func(
		ctx context.Context, project, resource string) (models.ProjectAttributes, error) {
		if project == testProject && resource == admin.MatchableResource_EXECUTION_QUEUE.String() {
			matchingAttributes := &admin.MatchingAttributes{
				Target: &admin.MatchingAttributes_ExecutionQueueAttributes{
					ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{
						Tags: []string{"attr1"},
					},
				},
			}
			marshalledMatchingAttributes, _ := proto.Marshal(matchingAttributes)
			return models.ProjectAttributes{
				Project:    project,
				Resource:   resource,
				Attributes: marshalledMatchingAttributes,
			}, nil
		}
		return models.ProjectAttributes{}, errors.NewFlyteAdminError(codes.NotFound, "foo")
	}

	queueAllocator := NewQueueAllocator(runtimeMocks.NewMockConfigurationProvider(
		nil, runtimeMocks.NewMockQueueConfigurationProvider(executionQueues, workflowConfigs), nil,
		nil, nil, nil), db)
	assert.Equal(t, singleQueueConfiguration{
		PrimaryQueue: "default primary",
		DynamicQueue: "default dynamic",
	}, queueAllocator.GetQueue(
		context.Background(), core.Identifier{
			Project: "unmatched",
			Domain:  "domain",
			Name:    "workflow",
		}))
	assert.EqualValues(t, singleQueueConfiguration{
		PrimaryQueue: "queue1 primary",
		DynamicQueue: "queue1 dynamic",
	}, queueAllocator.GetQueue(
		context.Background(), core.Identifier{
			Project: "project",
			Domain:  "UNMATCHED",
			Name:    "workflow",
		}))
	assert.EqualValues(t, singleQueueConfiguration{
		PrimaryQueue: "queue2 primary",
		DynamicQueue: "queue2 dynamic",
	}, queueAllocator.GetQueue(
		context.Background(), core.Identifier{
			Project: "project",
			Domain:  "domain",
			Name:    "UNMATCHED",
		}))
	assert.Equal(t, singleQueueConfiguration{
		PrimaryQueue: "queue3 primary",
		DynamicQueue: "queue3 dynamic",
	}, queueAllocator.GetQueue(
		context.Background(), core.Identifier{
			Project: "project",
			Domain:  "domain",
			Name:    "workflow",
		}))
}
