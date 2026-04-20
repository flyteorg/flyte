package service

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/hashicorp/golang-lru/v2/expirable"

	flyteapp "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/app"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/app/appconnect"
)

// AppService is the control plane implementation of AppServiceHandler.
// It proxies all RPCs to InternalAppService (data plane) and maintains a
// per-instance TTL cache to reduce cross-plane RPC calls on Get.
type AppService struct {
	appconnect.UnimplementedAppServiceHandler
	internalClient appconnect.AppServiceClient
	// cache is nil when cacheTTL=0 (caching disabled).
	cache *expirable.LRU[string, *flyteapp.App]
}

// NewAppService creates a new AppService.
// cacheTTL=0 disables caching (every Get calls InternalAppService).
func NewAppService(internalClient appconnect.AppServiceClient, cacheTTL time.Duration) *AppService {
	var cache *expirable.LRU[string, *flyteapp.App]
	if cacheTTL > 0 {
		cache = expirable.NewLRU[string, *flyteapp.App](0, nil, cacheTTL)
	}
	return &AppService{
		internalClient: internalClient,
		cache:          cache,
	}
}

// Ensure AppService satisfies the generated handler interface.
var _ appconnect.AppServiceHandler = (*AppService)(nil)

// Create forwards to InternalAppService and invalidates the cache entry.
func (s *AppService) Create(
	ctx context.Context,
	req *connect.Request[flyteapp.CreateRequest],
) (*connect.Response[flyteapp.CreateResponse], error) {
	resp, err := s.internalClient.Create(ctx, req)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		s.cache.Remove(cacheKey(req.Msg.GetApp().GetMetadata().GetId()))
	}
	return resp, nil
}

// Get returns the app, using the cache on hit and calling InternalAppService on miss.
func (s *AppService) Get(
	ctx context.Context,
	req *connect.Request[flyteapp.GetRequest],
) (*connect.Response[flyteapp.GetResponse], error) {
	appID, ok := req.Msg.GetIdentifier().(*flyteapp.GetRequest_AppId)
	if ok && appID.AppId != nil && s.cache != nil {
		if app, hit := s.cache.Get(cacheKey(appID.AppId)); hit {
			return connect.NewResponse(&flyteapp.GetResponse{App: app}), nil
		}
	}

	resp, err := s.internalClient.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	if ok && appID.AppId != nil && s.cache != nil {
		s.cache.Add(cacheKey(appID.AppId), resp.Msg.GetApp())
	}
	return resp, nil
}

// Update forwards to InternalAppService and invalidates the cache entry.
func (s *AppService) Update(
	ctx context.Context,
	req *connect.Request[flyteapp.UpdateRequest],
) (*connect.Response[flyteapp.UpdateResponse], error) {
	resp, err := s.internalClient.Update(ctx, req)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		s.cache.Remove(cacheKey(req.Msg.GetApp().GetMetadata().GetId()))
	}
	return resp, nil
}

// Delete forwards to InternalAppService and invalidates the cache entry.
func (s *AppService) Delete(
	ctx context.Context,
	req *connect.Request[flyteapp.DeleteRequest],
) (*connect.Response[flyteapp.DeleteResponse], error) {
	resp, err := s.internalClient.Delete(ctx, req)
	if err != nil {
		return nil, err
	}
	if s.cache != nil {
		s.cache.Remove(cacheKey(req.Msg.GetAppId()))
	}
	return resp, nil
}

// List always forwards to InternalAppService — results vary by filter/pagination.
func (s *AppService) List(
	ctx context.Context,
	req *connect.Request[flyteapp.ListRequest],
) (*connect.Response[flyteapp.ListResponse], error) {
	return s.internalClient.List(ctx, req)
}

// Watch proxies the server-streaming Watch RPC to InternalAppService.
func (s *AppService) Watch(
	ctx context.Context,
	req *connect.Request[flyteapp.WatchRequest],
	stream *connect.ServerStream[flyteapp.WatchResponse],
) error {
	clientStream, err := s.internalClient.Watch(ctx, req)
	if err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}
	defer clientStream.Close()
	for clientStream.Receive() {
		if err := stream.Send(clientStream.Msg()); err != nil {
			return err
		}
	}
	return clientStream.Err()
}

// cacheKey returns a stable string key for an app identifier.
func cacheKey(id *flyteapp.Identifier) string {
	if id == nil {
		return ""
	}
	return fmt.Sprintf("%s/%s/%s", id.GetProject(), id.GetDomain(), id.GetName())
}
