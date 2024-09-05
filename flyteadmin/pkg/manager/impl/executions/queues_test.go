package executions

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	managerInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/manager/interfaces"
	managerMocks "github.com/flyteorg/flyte/flyteadmin/pkg/manager/mocks"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/mocks"
	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	runtimeMocks "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/mocks"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

const testProject = "project"
const testDomain = "domain"
const testWorkflow = "name"
const UNMATCHED = "UNMATCHED"

func TestGetQueue(t *testing.T) {
	executionQueues := []runtimeInterfaces.ExecutionQueue{
		{
			Dynamic:    "queue dynamic",
			Attributes: []string{"attribute"},
		},
	}
	db := mocks.NewMockRepository()
	mockResourceManager := new(managerMocks.ResourceInterface)
	mockResourceManager.On("GetResource", context.Background(), mock.MatchedBy(func(request managerInterfaces.ResourceRequest) bool {
		return request.Project == testProject && request.Domain == testDomain && request.Workflow == testWorkflow && request.ResourceType == admin.MatchableResource_EXECUTION_QUEUE
	})).Return(&managerInterfaces.ResourceResponse{
		Project:      testProject,
		Domain:       testDomain,
		Workflow:     testWorkflow,
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE.String(),
		Attributes: &admin.MatchingAttributes{
			Target: &admin.MatchingAttributes_ExecutionQueueAttributes{
				ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{
					Tags: []string{"attribute"},
				},
			},
		},
	}, nil)
	mockResourceManager.On("GetResource", context.Background(), mock.MatchedBy(func(request managerInterfaces.ResourceRequest) bool {
		return request.ResourceType == admin.MatchableResource_EXECUTION_QUEUE
	})).Return(&managerInterfaces.ResourceResponse{
		Project:      testProject,
		Domain:       testDomain,
		Workflow:     testWorkflow,
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE.String(),
		Attributes: &admin.MatchingAttributes{
			Target: &admin.MatchingAttributes_ExecutionQueueAttributes{
				ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{
					Tags: []string{"another attribute"},
				},
			},
		},
	}, nil)

	queueAllocator := NewQueueAllocator(runtimeMocks.NewMockConfigurationProvider(
		nil, runtimeMocks.NewMockQueueConfigurationProvider(executionQueues, nil),
		nil, nil, nil, nil), db, mockResourceManager)
	queueConfig := singleQueueConfiguration{
		DynamicQueue: "queue dynamic",
	}
	assert.Equal(t, queueConfig, queueAllocator.GetQueue(context.Background(), &core.Identifier{
		Project: testProject,
		Domain:  testDomain,
		Name:    testWorkflow,
	}))
	assert.EqualValues(t, singleQueueConfiguration{}, queueAllocator.GetQueue(context.Background(), &core.Identifier{
		Project: testProject,
		Domain:  testDomain,
		Name:    UNMATCHED,
	}))
	assert.EqualValues(t, singleQueueConfiguration{}, queueAllocator.GetQueue(context.Background(), &core.Identifier{
		Project: testProject,
		Domain:  UNMATCHED,
		Name:    testWorkflow,
	}))
	assert.EqualValues(t, singleQueueConfiguration{}, queueAllocator.GetQueue(context.Background(), &core.Identifier{
		Project: UNMATCHED,
		Domain:  testDomain,
		Name:    testWorkflow,
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
	mockResourceManager := new(managerMocks.ResourceInterface)
	mockResourceManager.On("GetResource", context.Background(), mock.MatchedBy(func(request managerInterfaces.ResourceRequest) bool {
		return request.Project == testProject && request.Domain == testDomain && request.Workflow == testWorkflow && request.ResourceType == admin.MatchableResource_EXECUTION_QUEUE
	})).Return(&managerInterfaces.ResourceResponse{
		Project:      testProject,
		Domain:       testDomain,
		Workflow:     testWorkflow,
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE.String(),
		Attributes: &admin.MatchingAttributes{
			Target: &admin.MatchingAttributes_ExecutionQueueAttributes{
				ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{
					Tags: []string{"attr3"},
				},
			},
		},
	}, nil)
	mockResourceManager.On("GetResource", context.Background(), mock.MatchedBy(func(request managerInterfaces.ResourceRequest) bool {
		return request.Project == testProject && request.Domain == testDomain && request.ResourceType == admin.MatchableResource_EXECUTION_QUEUE
	})).Return(&managerInterfaces.ResourceResponse{
		Project:      testProject,
		Domain:       testDomain,
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE.String(),
		Attributes: &admin.MatchingAttributes{
			Target: &admin.MatchingAttributes_ExecutionQueueAttributes{
				ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{
					Tags: []string{"attr2"},
				},
			},
		},
	}, nil)
	mockResourceManager.On("GetResource", context.Background(), mock.MatchedBy(func(request managerInterfaces.ResourceRequest) bool {
		return request.Project == testProject && request.ResourceType == admin.MatchableResource_EXECUTION_QUEUE
	})).Return(&managerInterfaces.ResourceResponse{
		Project:      testProject,
		ResourceType: admin.MatchableResource_EXECUTION_QUEUE.String(),
		Attributes: &admin.MatchingAttributes{
			Target: &admin.MatchingAttributes_ExecutionQueueAttributes{
				ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{
					Tags: []string{"attr1"},
				},
			},
		},
	}, nil)
	mockResourceManager.On("GetResource", context.Background(), mock.MatchedBy(func(request managerInterfaces.ResourceRequest) bool {
		return request.ResourceType == admin.MatchableResource_EXECUTION_QUEUE
	})).Return(&managerInterfaces.ResourceResponse{}, errors.NewFlyteAdminError(codes.NotFound, "foo"))

	queueAllocator := NewQueueAllocator(runtimeMocks.NewMockConfigurationProvider(
		nil, runtimeMocks.NewMockQueueConfigurationProvider(executionQueues, workflowConfigs), nil,
		nil, nil, nil), db, mockResourceManager)
	assert.Equal(t, singleQueueConfiguration{
		DynamicQueue: "default dynamic",
	}, queueAllocator.GetQueue(
		context.Background(), &core.Identifier{
			Project: UNMATCHED,
			Domain:  testDomain,
			Name:    testWorkflow,
		}))
	assert.EqualValues(t, singleQueueConfiguration{
		DynamicQueue: "queue1 dynamic",
	}, queueAllocator.GetQueue(
		context.Background(), &core.Identifier{
			Project: testProject,
			Domain:  UNMATCHED,
			Name:    testWorkflow,
		}))
	assert.EqualValues(t, singleQueueConfiguration{
		DynamicQueue: "queue2 dynamic",
	}, queueAllocator.GetQueue(
		context.Background(), &core.Identifier{
			Project: testProject,
			Domain:  testDomain,
			Name:    UNMATCHED,
		}))
	assert.Equal(t, singleQueueConfiguration{
		DynamicQueue: "queue3 dynamic",
	}, queueAllocator.GetQueue(
		context.Background(), &core.Identifier{
			Project: testProject,
			Domain:  testDomain,
			Name:    testWorkflow,
		}))
}
