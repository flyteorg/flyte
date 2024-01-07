package mocks

import (
	"context"

	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
)

var (
	_ interfaces.CacheInterface = &MockCacheManager{}
)

type EvictTaskExecutionCacheFunc func(ctx context.Context, req service.EvictTaskExecutionCacheRequest) (*service.EvictCacheResponse, error)

type MockCacheManager struct {
	evictTaskExecutionCacheFunc EvictTaskExecutionCacheFunc
}

func (m *MockCacheManager) SetEvictTaskExecutionCacheFunc(evictTaskExecutionCacheFunc EvictTaskExecutionCacheFunc) {
	m.evictTaskExecutionCacheFunc = evictTaskExecutionCacheFunc
}

func (m *MockCacheManager) EvictTaskExecutionCache(ctx context.Context, req service.EvictTaskExecutionCacheRequest) (*service.EvictCacheResponse, error) {
	if m.evictTaskExecutionCacheFunc != nil {
		return m.evictTaskExecutionCacheFunc(ctx, req)
	}
	return nil, nil
}
