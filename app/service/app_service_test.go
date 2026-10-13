package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"
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

func (m *mockInternalClient) Create(ctx context.Context, req *flyteapp.CreateRequest) (*flyteapp.CreateResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*flyteapp.CreateResponse), args.Error(1)
}

func (m *mockInternalClient) Get(ctx context.Context, req *flyteapp.GetRequest) (*flyteapp.GetResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*flyteapp.GetResponse), args.Error(1)
}

func (m *mockInternalClient) Update(ctx context.Context, req *flyteapp.UpdateRequest) (*flyteapp.UpdateResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*flyteapp.UpdateResponse), args.Error(1)
}

func (m *mockInternalClient) UpdateStatus(ctx context.Context, req *flyteapp.UpdateStatusRequest) (*flyteapp.UpdateStatusResponse, error) {
	return nil, connect.NewError(connect.CodeUnimplemented, "")
}

func (m *mockInternalClient) Delete(ctx context.Context, req *flyteapp.DeleteRequest) (*flyteapp.DeleteResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*flyteapp.DeleteResponse), args.Error(1)
}

func (m *mockInternalClient) List(ctx context.Context, req *flyteapp.ListRequest) (*flyteapp.ListResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*flyteapp.ListResponse), args.Error(1)
}

func (m *mockInternalClient) Watch(ctx context.Context, req *flyteapp.WatchRequest) (appconnect.AppServiceWatchClientStream, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return appconnect.AppServiceWatchClientStream{}, args.Error(1)
	}
	return args.Get(0).(appconnect.AppServiceWatchClientStream), args.Error(1)
}

func (m *mockInternalClient) Lease(ctx context.Context, req *flyteapp.LeaseRequest) (appconnect.AppServiceLeaseClientStream, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return appconnect.AppServiceLeaseClientStream{}, args.Error(1)
	}
	return args.Get(0).(appconnect.AppServiceLeaseClientStream), args.Error(1)
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
		&flyteapp.GetResponse{App: app}, nil,
	)

	resp, err := svc.Get(context.Background(), &flyteapp.GetRequest{
		Identifier: &flyteapp.GetRequest_AppId{AppId: appID},
	})
	require.NoError(t, err)
	assert.Equal(t, flyteapp.Status_DEPLOYMENT_STATUS_ACTIVE, resp.App.Status.Conditions[0].DeploymentStatus)
	internal.AssertExpectations(t)
}

func TestGet_CacheHit_SkipsInternal(t *testing.T) {
	internal := &mockInternalClient{}
	svc := NewAppService(internal, 30*time.Second)

	// Pre-populate cache.
	appID := testAppID()
	svc.cache.Add(cacheKey(appID), testApp())

	// Internal should NOT be called.
	resp, err := svc.Get(context.Background(), &flyteapp.GetRequest{
		Identifier: &flyteapp.GetRequest_AppId{AppId: appID},
	})
	require.NoError(t, err)
	assert.Equal(t, flyteapp.Status_DEPLOYMENT_STATUS_ACTIVE, resp.App.Status.Conditions[0].DeploymentStatus)
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
		&flyteapp.GetResponse{App: app}, nil,
	)

	_, err := svc.Get(context.Background(), &flyteapp.GetRequest{
		Identifier: &flyteapp.GetRequest_AppId{AppId: appID},
	})
	require.NoError(t, err)
	internal.AssertExpectations(t)
}

func TestGet_TransitionalState_SkipsCache(t *testing.T) {
	// Simulate the Start-App bug: after clicking Start, K8s hasn't brought the
	// pod up yet so InternalAppService still returns STOPPED. The cache must NOT
	// be populated in this state, otherwise every poll for the full TTL window
	// returns stale "stopped" and the UI appears stuck.
	internal := &mockInternalClient{}
	svc := NewAppService(internal, 30*time.Second)

	appID := testAppID()
	// App desired=ACTIVE but status=STOPPED (transitional: just started).
	stoppedApp := &flyteapp.App{
		Metadata: &flyteapp.Meta{Id: appID},
		Spec: &flyteapp.Spec{
			AppPayload: &flyteapp.Spec_Container{
				Container: &flytecoreapp.Container{Image: "nginx:latest"},
			},
			DesiredState: flyteapp.Spec_DESIRED_STATE_ACTIVE,
		},
		Status: &flyteapp.Status{
			Conditions: []*flyteapp.Condition{
				{DeploymentStatus: flyteapp.Status_DEPLOYMENT_STATUS_STOPPED},
			},
		},
	}

	// Internal is called twice because the response is never cached.
	internal.On("Get", mock.Anything, mock.Anything).Return(
		&flyteapp.GetResponse{App: stoppedApp}, nil,
	).Times(2)

	req := &flyteapp.GetRequest{
		Identifier: &flyteapp.GetRequest_AppId{AppId: appID},
	}
	_, err := svc.Get(context.Background(), req)
	require.NoError(t, err)
	_, err = svc.Get(context.Background(), req)
	require.NoError(t, err)

	_, hit := svc.cache.Get(cacheKey(appID))
	assert.False(t, hit, "transitional state must not be cached")
	internal.AssertExpectations(t)
}

func TestGet_StoppedWithDesiredStopped_Caches(t *testing.T) {
	// Stable stopped state (user explicitly stopped it) should be cached normally.
	internal := &mockInternalClient{}
	svc := NewAppService(internal, 30*time.Second)

	appID := testAppID()
	stoppedApp := &flyteapp.App{
		Metadata: &flyteapp.Meta{Id: appID},
		Spec: &flyteapp.Spec{
			AppPayload: &flyteapp.Spec_Container{
				Container: &flytecoreapp.Container{Image: "nginx:latest"},
			},
			DesiredState: flyteapp.Spec_DESIRED_STATE_STOPPED,
		},
		Status: &flyteapp.Status{
			Conditions: []*flyteapp.Condition{
				{DeploymentStatus: flyteapp.Status_DEPLOYMENT_STATUS_STOPPED},
			},
		},
	}

	// Internal called once; second call served from cache.
	internal.On("Get", mock.Anything, mock.Anything).Return(
		&flyteapp.GetResponse{App: stoppedApp}, nil,
	).Once()

	req := &flyteapp.GetRequest{
		Identifier: &flyteapp.GetRequest_AppId{AppId: appID},
	}
	_, err := svc.Get(context.Background(), req)
	require.NoError(t, err)
	_, err = svc.Get(context.Background(), req)
	require.NoError(t, err)

	_, hit := svc.cache.Get(cacheKey(appID))
	assert.True(t, hit, "stable stopped state should be cached")
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
		&flyteapp.CreateResponse{App: app}, nil,
	)

	_, err := svc.Create(context.Background(), &flyteapp.CreateRequest{App: app})
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
		&flyteapp.UpdateResponse{App: app}, nil,
	)

	_, err := svc.Update(context.Background(), &flyteapp.UpdateRequest{App: app})
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
		&flyteapp.DeleteResponse{}, nil,
	)

	_, err := svc.Delete(context.Background(), &flyteapp.DeleteRequest{AppId: appID})
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
		&flyteapp.ListResponse{Apps: []*flyteapp.App{testApp()}}, nil,
	)

	resp, err := svc.List(context.Background(), &flyteapp.ListRequest{})
	require.NoError(t, err)
	assert.Len(t, resp.Apps, 1)
	internal.AssertExpectations(t)
}

// --- Watch streams through ---

func TestWatch_ProxiesStream(t *testing.T) {
	// Use a real httptest server to exercise the streaming path.
	internal := &mockInternalClient{}
	svc := NewAppService(internal, 30*time.Second)

	handlerServer := connect.NewServer()
	appconnect.RegisterAppServiceHandler(handlerServer, svc)
	path := "/" + appconnect.AppServiceName + "/"
	handler := http.NewServeMux()
	connecthttp.Mount(handler, handlerServer, connecthttp.WithReadMaxBytes(0))
	mux := http.NewServeMux()
	mux.Handle(path, handler)
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	// Mount InternalAppService on the same test server at /internal so the
	// proxy can route to it.
	internalSvcHandlerServer := connect.NewServer()
	appconnect.RegisterAppServiceHandler(internalSvcHandlerServer, &echoWatchService{app: testApp()})
	internalSvcPath := "/" + appconnect.AppServiceName + "/"
	internalSvcHandler := http.NewServeMux()
	connecthttp.Mount(internalSvcHandler, internalSvcHandlerServer, connecthttp.WithReadMaxBytes(0))
	mux.Handle("/internal"+internalSvcPath, http.StripPrefix("/internal", internalSvcHandler))

	// Point AppService proxy at the internal path on the same server.
	svc.internalClient = appconnect.NewAppServiceClient(connect.NewClient(connecthttp.NewTransport(http.DefaultClient, server.URL+"/internal", connecthttp.WithReadMaxBytes(0))))

	client := appconnect.NewAppServiceClient(connect.NewClient(connecthttp.NewTransport(http.DefaultClient, server.URL, connecthttp.WithReadMaxBytes(0))))
	stream, err := client.Watch(context.Background(), &flyteapp.WatchRequest{})
	require.NoError(t, err)

	streamMsg1, streamErr1 := stream.Receive()
	require.NoError(t, streamErr1)
	assert.Equal(t, "myapp", streamMsg1.GetCreateEvent().GetApp().GetMetadata().GetId().GetName())
	stream.Close()
}

// echoWatchService sends one CreateEvent then closes the stream.
type echoWatchService struct {
	appconnect.UnimplementedAppServiceHandler
	app *flyteapp.App
}

func (e *echoWatchService) Watch(
	_ context.Context,
	_ *flyteapp.WatchRequest,
	stream appconnect.AppServiceWatchServerStream,
) error {
	return stream.Send(&flyteapp.WatchResponse{
		Event: &flyteapp.WatchResponse_CreateEvent{
			CreateEvent: &flyteapp.CreateEvent{App: e.app},
		},
	})
}
