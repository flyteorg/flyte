package executions

import (
	"context"
	"testing"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"

	runtimeInterfaces "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	runtimeMocks "github.com/lyft/flyteadmin/pkg/runtime/mocks"
	"github.com/stretchr/testify/assert"
)

func TestGetAnyMapKey(t *testing.T) {
	fooConfig := singleQueueConfiguration{
		PrimaryQueue: "foo primary",
		DynamicQueue: "foo dynamic",
	}
	testMap := map[singleQueueConfiguration]bool{
		fooConfig: true,
	}
	key := getAnyMapKey(testMap)
	assert.Equal(t, fooConfig, key)

	barConfig := singleQueueConfiguration{
		PrimaryQueue: "bar primary",
		DynamicQueue: "bar dynamic",
	}
	testMap[barConfig] = true
	key = getAnyMapKey(testMap)
	assert.Contains(t, []string{"foo primary", "bar primary"}, key.PrimaryQueue)
}

func TestGetQueue(t *testing.T) {
	executionQueues := []runtimeInterfaces.ExecutionQueue{
		{
			Primary:    "queue primary",
			Dynamic:    "queue dynamic",
			Attributes: []string{"attribute"},
		},
	}
	workflowConfigs := []runtimeInterfaces.WorkflowConfig{
		{
			Project:      "project",
			Domain:       "domain",
			WorkflowName: "name",
			Tags:         []string{"attribute"},
		},
		{
			Project:      "project",
			Domain:       "domain",
			WorkflowName: "name2",
			Tags:         []string{"another attribute"},
		},
		{
			Project:      "project",
			Domain:       "domain2",
			WorkflowName: "name",
			Tags:         []string{"another attribute"},
		},
		{
			Project:      "project2",
			Domain:       "domain",
			WorkflowName: "name",
			Tags:         []string{"another attribute"},
		},
	}
	queueAllocator := NewQueueAllocator(runtimeMocks.NewMockConfigurationProvider(
		nil, runtimeMocks.NewMockQueueConfigurationProvider(executionQueues, workflowConfigs),
		nil, nil, nil, nil))
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
		{
			Project: "project",
			Tags:    []string{"attr1"},
		},
		{
			Project: "project",
			Domain:  "domain",
			Tags:    []string{"attr2"},
		},
		{
			Project:      "project",
			Domain:       "domain",
			WorkflowName: "workflow",
			Tags:         []string{"attr3"},
		},
	}
	queueAllocator := NewQueueAllocator(runtimeMocks.NewMockConfigurationProvider(
		nil, runtimeMocks.NewMockQueueConfigurationProvider(executionQueues, workflowConfigs), nil,
		nil, nil, nil))
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
