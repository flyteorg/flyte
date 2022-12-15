package interfaces

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
)

type CacheInterface interface {
	EvictExecutionCache(ctx context.Context, request service.EvictExecutionCacheRequest) (*service.EvictCacheResponse, error)
	EvictTaskExecutionCache(ctx context.Context, request service.EvictTaskExecutionCacheRequest) (*service.EvictCacheResponse, error)
}
