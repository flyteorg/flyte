package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	flyteapp "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/app"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/app/appconnect"
	flytecoreapp "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

// mockInternalClient is a testify mock for appconnect.AppServiceClient.
type mockInternalClient struct {
	mock.Mock
}

func (m *mockInternalClient) Create(ctx context.Context, req *connect.Request[flyteapp.CreateRequest]) (*connect.Response[flyteapp.CreateResponse], error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*connect.Response[flyteapp.CreateResponse]), args.Error(1)
}

func (m *mockInternalClient) Get(ctx context.Context, req *connect.Request[flyteapp.GetRequest]) (*connect.Response[flyteapp.GetResponse], error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*connect.Response[flyteapp.GetResponse]), args.Error(1)
}

func (m *mockInternalClient) Update(ctx context.Context, req *connect.Request[flyteapp.UpdateRequest]) (*connect.Response[flyteapp.UpdateResponse], error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*connect.Response[flyteapp.UpdateResponse]), args.Error(1)
}

func (m *mockInternalClient) UpdateStatus(ctx context.Context, req *connect.Request[flyteapp.UpdateStatusRequest]) (*connect.Response[flyteapp.UpdateStatusResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (m *mockInternalClient) Delete(ctx context.Context, req *connect.Request[flyteapp.DeleteRequest]) (*connect.Response[flyteapp.DeleteResponse], error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*connect.Response[flyteapp.DeleteResponse]), args.Error(1)
}

func (m *mockInternalClient) List(ctx context.Context, req *connect.Request[flyteapp.ListRequest]) (*connect.Response[flyteapp.ListResponse], error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*connect.Response[flyteapp.ListResponse]), args.Error(1)
}

func (m *mockInternalClient) Watch(ctx context.Context, req *connect.Request[flyteapp.WatchRequest]) (*connect.ServerStreamForClient[flyteapp.WatchResponse], error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*connect.ServerStreamForClient[flyteapp.WatchResponse]), args.Error(1)
}

func (m *mockInternalClient) Lease(ctx context.Context, req *connect.Request[flyteapp.LeaseRequest]) (*connect.ServerStreamForClient[flyteapp.LeaseResponse], error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*connect.ServerStreamForClient[flyteapp.LeaseResponse]), args.Error(1)
}

// --- helpers ---

func testAppID() *flyteapp.Identifier {
	return &flyteapp.Identifier{Project: "proj", Domain: "dev", Name: "myapp"}
}

func testApp() *flyteapp.App {
	return &flyteapp.App{
		Metadata: &flyteapp.Meta{Id: testAppID()},
		Spec: &flyteapp.Spec{
			AppPayload: &flyteapp.Spec_Container{
				Container: &flytecoreapp.Container{Image: "nginx:latest"},
			},
		},
		Status: &flyteapp.Status{
			Conditions: []*flyteapp.Condition{
				{DeploymentStatus: flyteapp.Status_DEPLOYMENT_STATUS_ACTIVE},
			},
		},
	}
}

// --- Get with cache ---

func TestGet_CacheMiss_CallsInternal(t *testing.T) {
	internal := &mockInternalClient{}
	svc := NewAppService(internal, 30*time.Second)

	appID := testAppID()
	app := testApp()
	internal.On("Get", mock.Anything, mock.Anything).Return(
		connect.NewResponse(&flyteapp.GetResponse{App: app}), nil,
	)

	resp, err := svc.Get(context.Background(), connect.NewRequest(&flyteapp.GetRequest{
		Identifier: &flyteapp.GetRequest_AppId{AppId: appID},
	}))
	require.NoError(t, err)
	assert.Equal(t, flyteapp.Status_DEPLOYMENT_STATUS_ACTIVE, resp.Msg.App.Status.Conditions[0].DeploymentStatus)
	internal.AssertExpectations(t)
}

func TestGet_CacheHit_SkipsInternal(t *testing.T) {
	internal := &mockInternalClient{}
	svc := NewAppService(internal, 30*time.Second)

	// Pre-populate cache.
	appID := testAppID()
	svc.cache.Add(cacheKey(appID), testApp())

	// Internal should NOT be called.
	resp, err := svc.Get(context.Background(), connect.NewRequest(&flyteapp.GetRequest{
		Identifier: &flyteapp.GetRequest_AppId{AppId: appID},
	}))
	require.NoError(t, err)
	assert.Equal(t, flyteapp.Status_DEPLOYMENT_STATUS_ACTIVE, resp.Msg.App.Status.Conditions[0].DeploymentStatus)
	internal.AssertNotCalled(t, "Get")
}

func TestGet_CacheExpired_CallsInternal(t *testing.T) {
	internal := &mockInternalClient{}
	svc := NewAppService(internal, 1*time.Millisecond)

	appID := testAppID()
	svc.cache.Add(cacheKey(appID), testApp())
	time.Sleep(5 * time.Millisecond) // let TTL expire

	app := testApp()
	internal.On("Get", mock.Anything, mock.Anything).Return(
		connect.NewResponse(&flyteapp.GetResponse{App: app}), nil,
	)

	_, err := svc.Get(context.Background(), connect.NewRequest(&flyteapp.GetRequest{
		Identifier: &flyteapp.GetRequest_AppId{AppId: appID},
	}))
	require.NoError(t, err)
	internal.AssertExpectations(t)
}

// --- Create / Update / Delete invalidate cache ---

func TestCreate_InvalidatesCache(t *testing.T) {
	internal := &mockInternalClient{}
	svc := NewAppService(internal, 30*time.Second)

	app := testApp()
	// Pre-populate cache so we can confirm it's cleared.
	svc.cache.Add(cacheKey(app.Metadata.Id), app)

	internal.On("Create", mock.Anything, mock.Anything).Return(
		connect.NewResponse(&flyteapp.CreateResponse{App: app}), nil,
	)

	_, err := svc.Create(context.Background(), connect.NewRequest(&flyteapp.CreateRequest{App: app}))
	require.NoError(t, err)

	_, hit := svc.cache.Get(cacheKey(app.Metadata.Id))
	assert.False(t, hit, "cache should be invalidated after Create")
	internal.AssertExpectations(t)
}

func TestUpdate_InvalidatesCache(t *testing.T) {
	internal := &mockInternalClient{}
	svc := NewAppService(internal, 30*time.Second)

	app := testApp()
	svc.cache.Add(cacheKey(app.Metadata.Id), app)

	internal.On("Update", mock.Anything, mock.Anything).Return(
		connect.NewResponse(&flyteapp.UpdateResponse{App: app}), nil,
	)

	_, err := svc.Update(context.Background(), connect.NewRequest(&flyteapp.UpdateRequest{App: app}))
	require.NoError(t, err)

	_, hit := svc.cache.Get(cacheKey(app.Metadata.Id))
	assert.False(t, hit, "cache should be invalidated after Update")
	internal.AssertExpectations(t)
}

func TestDelete_InvalidatesCache(t *testing.T) {
	internal := &mockInternalClient{}
	svc := NewAppService(internal, 30*time.Second)

	appID := testAppID()
	svc.cache.Add(cacheKey(appID), testApp())

	internal.On("Delete", mock.Anything, mock.Anything).Return(
		connect.NewResponse(&flyteapp.DeleteResponse{}), nil,
	)

	_, err := svc.Delete(context.Background(), connect.NewRequest(&flyteapp.DeleteRequest{AppId: appID}))
	require.NoError(t, err)

	_, hit := svc.cache.Get(cacheKey(appID))
	assert.False(t, hit, "cache should be invalidated after Delete")
	internal.AssertExpectations(t)
}

// --- List always forwards ---

func TestList_AlwaysCallsInternal(t *testing.T) {
	internal := &mockInternalClient{}
	svc := NewAppService(internal, 30*time.Second)

	internal.On("List", mock.Anything, mock.Anything).Return(
		connect.NewResponse(&flyteapp.ListResponse{Apps: []*flyteapp.App{testApp()}}), nil,
	)

	resp, err := svc.List(context.Background(), connect.NewRequest(&flyteapp.ListRequest{}))
	require.NoError(t, err)
	assert.Len(t, resp.Msg.Apps, 1)
	internal.AssertExpectations(t)
}

// --- Watch streams through ---

func TestWatch_ProxiesStream(t *testing.T) {
	// Use a real httptest server to exercise the streaming path.
	internal := &mockInternalClient{}
	svc := NewAppService(internal, 30*time.Second)

	path, handler := appconnect.NewAppServiceHandler(svc)
	mux := http.NewServeMux()
	mux.Handle(path, handler)
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	// Mount InternalAppService on the same test server at /internal so the
	// proxy can route to it.
	internalSvcPath, internalSvcHandler := appconnect.NewAppServiceHandler(
		&echoWatchService{app: testApp()},
	)
	mux.Handle("/internal"+internalSvcPath, http.StripPrefix("/internal", internalSvcHandler))

	// Point AppService proxy at the internal path on the same server.
	svc.internalClient = appconnect.NewAppServiceClient(http.DefaultClient, server.URL+"/internal")

	client := appconnect.NewAppServiceClient(http.DefaultClient, server.URL)
	stream, err := client.Watch(context.Background(), connect.NewRequest(&flyteapp.WatchRequest{}))
	require.NoError(t, err)

	require.True(t, stream.Receive())
	assert.Equal(t, "myapp", stream.Msg().GetCreateEvent().GetApp().GetMetadata().GetId().GetName())
	stream.Close()
}

// echoWatchService sends one CreateEvent then closes the stream.
type echoWatchService struct {
	appconnect.UnimplementedAppServiceHandler
	app *flyteapp.App
}

func (e *echoWatchService) Watch(
	_ context.Context,
	_ *connect.Request[flyteapp.WatchRequest],
	stream *connect.ServerStream[flyteapp.WatchResponse],
) error {
	return stream.Send(&flyteapp.WatchResponse{
		Event: &flyteapp.WatchResponse_CreateEvent{
			CreateEvent: &flyteapp.CreateEvent{App: e.app},
		},
	})
}
