package mocks

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
)

var (
	_ interfaces.CacheInterface = &MockCacheManager{}
)

type EvictExecutionCacheFunc func(ctx context.Context, req service.EvictExecutionCacheRequest) (*service.EvictCacheResponse, error)
type EvictTaskExecutionCacheFunc func(ctx context.Context, req service.EvictTaskExecutionCacheRequest) (*service.EvictCacheResponse, error)

type MockCacheManager struct {
	evictExecutionCacheFunc     EvictExecutionCacheFunc
	evictTaskExecutionCacheFunc EvictTaskExecutionCacheFunc
}

func (m *MockCacheManager) SetEvictExecutionCacheFunc(evictExecutionCacheFunc EvictExecutionCacheFunc) {
	m.evictExecutionCacheFunc = evictExecutionCacheFunc
}

func (m *MockCacheManager) EvictExecutionCache(ctx context.Context, req service.EvictExecutionCacheRequest) (*service.EvictCacheResponse, error) {
	if m.evictExecutionCacheFunc != nil {
		return m.evictExecutionCacheFunc(ctx, req)
	}
	return nil, nil
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
