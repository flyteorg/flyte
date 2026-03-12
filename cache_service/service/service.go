package service

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	cacheservicepb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/cacheservice"
	cacheservicev2 "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/cacheservice/v2"
)

type CacheService struct{}

func NewCacheService() *CacheService {
	return &CacheService{}
}

func (s *CacheService) Get(context.Context, *connect.Request[cacheservicev2.GetCacheRequest]) (*connect.Response[cacheservicepb.GetCacheResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("cache service not implemented"))
}

func (s *CacheService) Put(context.Context, *connect.Request[cacheservicev2.PutCacheRequest]) (*connect.Response[cacheservicepb.PutCacheResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("cache service not implemented"))
}

func (s *CacheService) Delete(context.Context, *connect.Request[cacheservicev2.DeleteCacheRequest]) (*connect.Response[cacheservicepb.DeleteCacheResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("cache service not implemented"))
}

func (s *CacheService) GetOrExtendReservation(context.Context, *connect.Request[cacheservicev2.GetOrExtendReservationRequest]) (*connect.Response[cacheservicepb.GetOrExtendReservationResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("cache service not implemented"))
}

func (s *CacheService) ReleaseReservation(context.Context, *connect.Request[cacheservicev2.ReleaseReservationRequest]) (*connect.Response[cacheservicepb.ReleaseReservationResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("cache service not implemented"))
}
