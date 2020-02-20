package resourcemanager

import (
	"context"
	"sync"
	"testing"

	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/task/resourcemanager/mocks"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRedisResourceManager_AllocateResource(t *testing.T) {

	mockRedisClient := &mocks.RedisClient{}

	mockScope := promutils.NewScope("testscope")
	mockScope1 := mockScope.NewSubScope("test_resource1")
	mockScope2 := mockScope.NewSubScope("test_resource2")
	mockMetric1 := NewRedisResourceManagerMetrics(mockScope1)
	mockMetric2 := NewRedisResourceManagerMetrics(mockScope2)
	mockNamespacedResourcesMap := map[core.ResourceNamespace]*Resource{
		core.ResourceNamespace("test-resource1"): &Resource{
			quota:             3,
			namespaceQuotaCap: 1,
			metrics:           mockMetric1,
			rejectedTokens:    sync.Map{},
		},
		core.ResourceNamespace("test-resource2"): &Resource{
			quota:             4,
			namespaceQuotaCap: 0,
			metrics:           mockMetric2,
			rejectedTokens:    sync.Map{},
		},
	}
	mockContext := context.TODO()
	t.Run("Namespace cap is enforced. Namespace1 should not be granted while namespace2 should", func(t *testing.T) {
		r := &RedisResourceManager{
			client:                 mockRedisClient,
			MetricsScope:           mockScope,
			namespacedResourcesMap: mockNamespacedResourcesMap,
		}
		allocatedTokens := []string{"ns1-token1"}
		mockRedisClient.OnSIsMember("test-resource1", mock.Anything).Return(false, nil) // how can I assign val?
		mockRedisClient.OnSMembers("test-resource1").Return(allocatedTokens, nil)
		mockRedisClient.OnSCard("test-resource1").Return(int64(len(allocatedTokens)), nil)
		mockRedisClient.OnSAdd(mock.Anything, mock.Anything).Return(0, nil)
		ns1Got, err := r.AllocateResource(mockContext, "test-resource1", "ns1-token2")
		assert.Nil(t, err)
		assert.Equal(t, core.AllocationStatusExhausted, ns1Got)

		ns2Got, err := r.AllocateResource(mockContext, "test-resource1", "ns2-token1")
		assert.Nil(t, err)
		assert.Equal(t, core.AllocationStatusGranted, ns2Got)
	})

	t.Run("Namespace cap is not enforced. Namespace1 should be granted the 4th token of resource 2.", func(t *testing.T) {
		r := &RedisResourceManager{
			client:                 mockRedisClient,
			MetricsScope:           mockScope,
			namespacedResourcesMap: mockNamespacedResourcesMap,
		}
		allocatedTokens := []string{"ns1-token1", "ns1-token2", "ns1-token3"}
		mockRedisClient.OnSIsMember("test-resource2", mock.Anything).Return(false, nil) // how can I assign val?
		mockRedisClient.OnSMembers("test-resource2").Return(allocatedTokens, nil)
		mockRedisClient.OnSCard("test-resource2").Return(int64(len(allocatedTokens)), nil)
		mockRedisClient.OnSAdd(mock.Anything, mock.Anything).Return(0, nil)

		got, err := r.AllocateResource(mockContext, "test-resource2", "ns1-token4")
		assert.Nil(t, err)
		assert.Equal(t, core.AllocationStatusGranted, got)
	})

	t.Run("Namespace cap is not enforced but resource 2 is exhausted. No namespace should be granted any token of resource 2.", func(t *testing.T) {
		r := &RedisResourceManager{
			client:                 mockRedisClient,
			MetricsScope:           mockScope,
			namespacedResourcesMap: mockNamespacedResourcesMap,
		}
		allocatedTokens := []string{"ns1-token1", "ns1-token2", "ns1-token3", "ns1-token4"}
		mockRedisClient.OnSIsMember("test-resource2", mock.Anything).Return(false, nil) // how can I assign val?
		mockRedisClient.OnSMembers("test-resource2").Return(allocatedTokens, nil)
		mockRedisClient.OnSCard("test-resource2").Return(int64(len(allocatedTokens)), nil)
		mockRedisClient.OnSAdd(mock.Anything, mock.Anything).Return(0, nil)

		got, err := r.AllocateResource(mockContext, "test-resource2", "ns1-token5")
		assert.Nil(t, err)
		assert.Equal(t, core.AllocationStatusExhausted, got)

		got, err = r.AllocateResource(mockContext, "test-resource2", "ns2-token1")
		assert.Nil(t, err)
		assert.Equal(t, core.AllocationStatusExhausted, got)
	})
}

func TestRedisResourceManager_getNamespaceAllocatedCount(t *testing.T) {
	mockRedisClient := &mocks.RedisClient{}

	mockScope := promutils.NewScope("testscope2")
	mockScope1 := mockScope.NewSubScope("test_resource1")
	mockScope2 := mockScope.NewSubScope("test_resource2")
	mockMetric1 := NewRedisResourceManagerMetrics(mockScope1)
	mockMetric2 := NewRedisResourceManagerMetrics(mockScope2)
	mockNamespacedResourcesMap := map[core.ResourceNamespace]*Resource{
		core.ResourceNamespace("test-resource1"): &Resource{
			quota:             3,
			namespaceQuotaCap: 1,
			metrics:           mockMetric1,
			rejectedTokens:    sync.Map{},
		},
		core.ResourceNamespace("test-resource2"): &Resource{
			quota:             4,
			namespaceQuotaCap: 0,
			metrics:           mockMetric2,
			rejectedTokens:    sync.Map{},
		},
	}
	mockContext := context.TODO()
	t.Run("Malformed token should not trigger an error but the count should be 0", func(t *testing.T) {
		r := &RedisResourceManager{
			client:                 mockRedisClient,
			MetricsScope:           mockScope,
			namespacedResourcesMap: mockNamespacedResourcesMap,
		}
		allocatedTokens := []string{"ns1-token1"}
		mockRedisClient.OnSIsMember("test-resource1", mock.Anything).Return(false, nil) // how can I assign val?
		mockRedisClient.OnSMembers("test-resource1").Return(allocatedTokens, nil)
		mockRedisClient.OnSCard("test-resource1").Return(int64(len(allocatedTokens)), nil)
		mockRedisClient.OnSAdd(mock.Anything, mock.Anything).Return(0, nil)
		count, err := r.getNamespaceAllocatedCount(mockContext, mockRedisClient, "test-resource1", "token2")
		assert.Nil(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("Wellformed token should get the corresponding count", func(t *testing.T) {
		r := &RedisResourceManager{
			client:                 mockRedisClient,
			MetricsScope:           mockScope,
			namespacedResourcesMap: mockNamespacedResourcesMap,
		}
		allocatedTokens := []string{"ns1-token1"}
		mockRedisClient.OnSIsMember("test-resource1", mock.Anything).Return(false, nil) // how can I assign val?
		mockRedisClient.OnSMembers("test-resource1").Return(allocatedTokens, nil)
		mockRedisClient.OnSCard("test-resource1").Return(int64(len(allocatedTokens)), nil)
		mockRedisClient.OnSAdd(mock.Anything, mock.Anything).Return(0, nil)
		count, err := r.getNamespaceAllocatedCount(mockContext, mockRedisClient, "test-resource1", "ns1-token2")
		assert.Nil(t, err)
		assert.Equal(t, 1, count)
	})
}
