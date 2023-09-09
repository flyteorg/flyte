package executions

import (
	"context"
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"

	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/mocks"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	runtimeMocks "github.com/flyteorg/flyteadmin/pkg/runtime/mocks"
	"github.com/stretchr/testify/assert"
)

const testProject = "project"
const testDomain = "domain"
const testWorkflow = "name"

func TestGetQueue(t *testing.T) {
	executionQueues := []runtimeInterfaces.ExecutionQueue{
		{
			Dynamic:    "queue dynamic",
			Attributes: []string{"attribute"},
		},
	}
	db := mocks.NewMockRepository()
	db.ResourceRepo().(*mocks.MockResourceRepo).GetFunction = func(ctx context.Context, ID interfaces.ResourceID) (resource models.Resource, e error) {
		response := models.Resource{
			Project:      ID.Project,
			Domain:       ID.Domain,
			Workflow:     ID.Workflow,
			ResourceType: ID.ResourceType,
		}
		if ID.Project == testProject && ID.Domain == testDomain && ID.Workflow == testWorkflow {
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
			Dynamic:    "queue1 dynamic",
			Attributes: []string{"attr1"},
		},
		{
			Dynamic:    "queue2 dynamic",
			Attributes: []string{"attr2"},
		},
		{
			Dynamic:    "queue3 dynamic",
			Attributes: []string{"attr3"},
		},
		{
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
	db.ResourceRepo().(*mocks.MockResourceRepo).GetFunction = func(ctx context.Context, ID interfaces.ResourceID) (resource models.Resource, e error) {
		if ID.Project == testProject && ID.Domain == testDomain && ID.Workflow == "workflow" &&
			ID.ResourceType == admin.MatchableResource_EXECUTION_QUEUE.String() {
			matchingAttributes := &admin.MatchingAttributes{
				Target: &admin.MatchingAttributes_ExecutionQueueAttributes{
					ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{
						Tags: []string{"attr3"},
					},
				},
			}
			marshalledMatchingAttributes, _ := proto.Marshal(matchingAttributes)
			return models.Resource{
				Project:      ID.Project,
				Domain:       ID.Domain,
				Workflow:     ID.Workflow,
				ResourceType: ID.ResourceType,
				Attributes:   marshalledMatchingAttributes,
			}, nil
		}
		if ID.Project == testProject && ID.Domain == testDomain && ID.ResourceType == admin.MatchableResource_EXECUTION_QUEUE.String() {
			matchingAttributes := &admin.MatchingAttributes{
				Target: &admin.MatchingAttributes_ExecutionQueueAttributes{
					ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{
						Tags: []string{"attr2"},
					},
				},
			}
			marshalledMatchingAttributes, _ := proto.Marshal(matchingAttributes)
			return models.Resource{
				Project:      ID.Project,
				Domain:       ID.Domain,
				ResourceType: ID.ResourceType,
				Attributes:   marshalledMatchingAttributes,
			}, nil
		}

		if ID.Project == testProject && ID.ResourceType == admin.MatchableResource_EXECUTION_QUEUE.String() {
			matchingAttributes := &admin.MatchingAttributes{
				Target: &admin.MatchingAttributes_ExecutionQueueAttributes{
					ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{
						Tags: []string{"attr1"},
					},
				},
			}
			marshalledMatchingAttributes, _ := proto.Marshal(matchingAttributes)
			return models.Resource{
				Project:      ID.Project,
				ResourceType: ID.ResourceType,
				Attributes:   marshalledMatchingAttributes,
			}, nil
		}
		return models.Resource{}, errors.NewFlyteAdminError(codes.NotFound, "foo")
	}

	queueAllocator := NewQueueAllocator(runtimeMocks.NewMockConfigurationProvider(
		nil, runtimeMocks.NewMockQueueConfigurationProvider(executionQueues, workflowConfigs), nil,
		nil, nil, nil), db)
	assert.Equal(t, singleQueueConfiguration{
		DynamicQueue: "default dynamic",
	}, queueAllocator.GetQueue(
		context.Background(), core.Identifier{
			Project: "unmatched",
			Domain:  "domain",
			Name:    "workflow",
		}))
	assert.EqualValues(t, singleQueueConfiguration{
		DynamicQueue: "queue1 dynamic",
	}, queueAllocator.GetQueue(
		context.Background(), core.Identifier{
			Project: "project",
			Domain:  "UNMATCHED",
			Name:    "workflow",
		}))
	assert.EqualValues(t, singleQueueConfiguration{
		DynamicQueue: "queue2 dynamic",
	}, queueAllocator.GetQueue(
		context.Background(), core.Identifier{
			Project: "project",
			Domain:  "domain",
			Name:    "UNMATCHED",
		}))
	assert.Equal(t, singleQueueConfiguration{
		DynamicQueue: "queue3 dynamic",
	}, queueAllocator.GetQueue(
		context.Background(), core.Identifier{
			Project: "project",
			Domain:  "domain",
			Name:    "workflow",
		}))
}
