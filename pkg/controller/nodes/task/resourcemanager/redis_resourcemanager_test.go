package resourcemanager

import (
	"context"
	"sync"
	"testing"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/resourcemanager/mocks"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func createMockNamespacedResourcesMap(mScope promutils.Scope) map[core.ResourceNamespace]*Resource {
	mockScope1 := mScope.NewSubScope("test_resource1")
	mockScope2 := mScope.NewSubScope("test_resource2")
	mockMetric1 := NewRedisResourceManagerMetrics(mockScope1)
	mockMetric2 := NewRedisResourceManagerMetrics(mockScope2)
	mockNamespacedResourcesMap := map[core.ResourceNamespace]*Resource{
		core.ResourceNamespace("test-resource1"): &Resource{
			quota:          BaseResourceConstraint{Value: 3},
			metrics:        mockMetric1,
			rejectedTokens: sync.Map{},
		},
		core.ResourceNamespace("test-resource2"): &Resource{
			quota:          BaseResourceConstraint{Value: 4},
			metrics:        mockMetric2,
			rejectedTokens: sync.Map{},
		},
	}
	return mockNamespacedResourcesMap
}

func createMockComposedResourceConstraintList() []FullyQualifiedResourceConstraint {
	return []FullyQualifiedResourceConstraint{
		{
			TargetedPrefixString: "ns1",
			Value:                1,
		},
	}
}

func TestRedisResourceManager_AllocateResource(t *testing.T) {

	t.Run("Namespace cap is enforced. Namespace1 should not be granted while namespace2 should", func(t *testing.T) {
		mockScope := promutils.NewTestScope()
		mockRedisClient := &mocks.RedisClient{}
		mockContext := context.TODO()
		r := &RedisResourceManager{
			client:                 mockRedisClient,
			MetricsScope:           mockScope,
			namespacedResourcesMap: createMockNamespacedResourcesMap(mockScope),
		}
		allocatedTokens := []string{"ns1-token1"}
		mockRedisClient.OnSIsMember("test-resource1", mock.Anything).Return(false, nil) // how can I assign val?
		mockRedisClient.OnSMembers("test-resource1").Return(allocatedTokens, nil)
		mockRedisClient.OnSCard("test-resource1").Return(int64(len(allocatedTokens)), nil)
		mockRedisClient.OnSAdd(mock.Anything, mock.Anything).Return(0, nil)

		mockComposedResourceConstraintList := createMockComposedResourceConstraintList()
		ns1Got, err := r.AllocateResource(mockContext, "test-resource1", "ns1-token2", mockComposedResourceConstraintList)
		assert.Nil(t, err)
		assert.Equal(t, core.AllocationStatusExhausted, ns1Got)

		ns2Got, err := r.AllocateResource(mockContext, "test-resource1", "ns2-token1", []FullyQualifiedResourceConstraint{})
		assert.Nil(t, err)
		assert.Equal(t, core.AllocationStatusGranted, ns2Got)
	})

	t.Run("Namespace cap is not enforced. Namespace1 should be granted the 4th token of resource 2.", func(t *testing.T) {
		mockScope := promutils.NewTestScope()
		mockRedisClient := &mocks.RedisClient{}
		mockContext := context.TODO()
		r := &RedisResourceManager{
			client:                 mockRedisClient,
			MetricsScope:           mockScope,
			namespacedResourcesMap: createMockNamespacedResourcesMap(mockScope),
		}
		allocatedTokens := []string{"ns1-token1", "ns1-token2", "ns1-token3"}
		mockRedisClient.OnSIsMember("test-resource2", mock.Anything).Return(false, nil) // how can I assign val?
		mockRedisClient.OnSMembers("test-resource2").Return(allocatedTokens, nil)
		mockRedisClient.OnSCard("test-resource2").Return(int64(len(allocatedTokens)), nil)
		mockRedisClient.OnSAdd(mock.Anything, mock.Anything).Return(0, nil)
		got, err := r.AllocateResource(mockContext, "test-resource2", "ns1-token4", []FullyQualifiedResourceConstraint{})
		assert.Nil(t, err)
		assert.Equal(t, core.AllocationStatusGranted, got)
	})

	t.Run("Namespace cap is not enforced but resource 2 is exhausted. No namespace should be granted any token of resource 2.", func(t *testing.T) {
		mockScope := promutils.NewTestScope()
		mockRedisClient := &mocks.RedisClient{}
		mockContext := context.TODO()
		r := &RedisResourceManager{
			client:                 mockRedisClient,
			MetricsScope:           mockScope,
			namespacedResourcesMap: createMockNamespacedResourcesMap(mockScope),
		}
		allocatedTokens := []string{"ns1-token1", "ns1-token2", "ns1-token3", "ns1-token4"}
		mockRedisClient.OnSIsMember("test-resource2", mock.Anything).Return(false, nil) // how can I assign val?
		mockRedisClient.OnSMembers("test-resource2").Return(allocatedTokens, nil)
		mockRedisClient.OnSCard("test-resource2").Return(int64(len(allocatedTokens)), nil)
		mockRedisClient.OnSAdd(mock.Anything, mock.Anything).Return(0, nil)
		mockComposedResourceConstraintList := createMockComposedResourceConstraintList()
		got, err := r.AllocateResource(mockContext, "test-resource2", "ns1-token5", mockComposedResourceConstraintList)
		assert.Nil(t, err)
		assert.Equal(t, core.AllocationStatusExhausted, got)

		got, err = r.AllocateResource(mockContext, "test-resource2", "ns2-token1", mockComposedResourceConstraintList)
		assert.Nil(t, err)
		assert.Equal(t, core.AllocationStatusExhausted, got)
	})
}

func TestRedisResourceManager_checkAgainstConstraints(t *testing.T) {
	t.Run("The violated constraint with the largest scope should be pinpointed", func(t *testing.T) {
		mockScope := promutils.NewTestScope()
		mockRedisClient := &mocks.RedisClient{}
		mockContext := context.TODO()
		r := &RedisResourceManager{
			client:                 mockRedisClient,
			MetricsScope:           mockScope,
			namespacedResourcesMap: createMockNamespacedResourcesMap(mockScope),
		}
		allocatedTokens := []string{"ns1-token1"}
		mockRedisClient.OnSIsMember("test-resource1", mock.Anything).Return(false, nil) // how can I assign val?
		mockRedisClient.OnSMembers("test-resource1").Return(allocatedTokens, nil)
		mockRedisClient.OnSCard("test-resource1").Return(int64(len(allocatedTokens)), nil)
		mockRedisClient.OnSAdd(mock.Anything, mock.Anything).Return(0, nil)
		mockComposedResourceConstraintList := createMockComposedResourceConstraintList()
		ok, idx, err := r.checkAgainstConstraints(mockContext, mockRedisClient, "test-resource1", mockComposedResourceConstraintList)
		assert.NoError(t, err)
		assert.False(t, ok)
		assert.Equal(t, 0, idx)
	})
}
