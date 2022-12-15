package cacheservice

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyteadmin/pkg/rpc/adminservice/util"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flytestdlib/logger"
)

func (s *CacheService) EvictExecutionCache(ctx context.Context, req *service.EvictExecutionCacheRequest) (*service.EvictCacheResponse, error) {
	defer s.interceptPanic(ctx, req)
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}

	var resp *service.EvictCacheResponse
	var err error
	s.Metrics.cacheEndpointMetrics.evictExecution.Time(func() {
		resp, err = s.CacheManager.EvictExecutionCache(ctx, *req)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &s.Metrics.cacheEndpointMetrics.evictExecution)
	}
	s.Metrics.cacheEndpointMetrics.evictExecution.Success()

	return resp, nil
}

func (s *CacheService) EvictTaskExecutionCache(ctx context.Context, req *service.EvictTaskExecutionCacheRequest) (*service.EvictCacheResponse, error) {
	defer s.interceptPanic(ctx, req)
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Incorrect request, nil requests not allowed")
	}

	// Calling the HTTP endpoint implicitly carries the resource type information in the URL, but does not set it in the
	// parsed protobuf message. Ensure task execution identifier is valid by manually setting task resource type below.
	if req.TaskExecutionId != nil && req.TaskExecutionId.TaskId != nil &&
		req.TaskExecutionId.TaskId.ResourceType == core.ResourceType_UNSPECIFIED {
		logger.Infof(ctx, "Adding task resource type for unspecified value in request: [%+v]", req)
		req.TaskExecutionId.TaskId.ResourceType = core.ResourceType_TASK
	}

	var resp *service.EvictCacheResponse
	var err error
	s.Metrics.cacheEndpointMetrics.evictTaskExecution.Time(func() {
		resp, err = s.CacheManager.EvictTaskExecutionCache(ctx, *req)
	})
	if err != nil {
		return nil, util.TransformAndRecordError(err, &s.Metrics.cacheEndpointMetrics.evictTaskExecution)
	}
	s.Metrics.cacheEndpointMetrics.evictTaskExecution.Success()

	return resp, nil
}
