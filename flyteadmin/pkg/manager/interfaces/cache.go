package interfaces

import (
	"context"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
)

type CacheInterface interface {
	EvictTaskExecutionCache(ctx context.Context, request service.EvictTaskExecutionCacheRequest) (*service.EvictCacheResponse, error)
}
