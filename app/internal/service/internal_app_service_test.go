package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	appconfig "github.com/flyteorg/flyte/v2/app/config"
	flyteapp "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/app"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/app/appconnect"
	flytecoreapp "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
)

// mockAppK8sClient is a testify mock for AppK8sClientInterface.
type mockAppK8sClient struct {
	mock.Mock
}

func (m *mockAppK8sClient) Deploy(ctx context.Context, app *flyteapp.App) error {
	return m.Called(ctx, app).Error(0)
}

func (m *mockAppK8sClient) Stop(ctx context.Context, appID *flyteapp.Identifier) error {
	return m.Called(ctx, appID).Error(0)
}

func (m *mockAppK8sClient) Delete(ctx context.Context, appID *flyteapp.Identifier) error {
	return m.Called(ctx, appID).Error(0)
}

func (m *mockAppK8sClient) GetStatus(ctx context.Context, appID *flyteapp.Identifier) (*flyteapp.Status, error) {
	args := m.Called(ctx, appID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*flyteapp.Status), args.Error(1)
}

func (m *mockAppK8sClient) List(ctx context.Context, project, domain, appName string, limit uint32, token string) ([]*flyteapp.App, string, error) {
	args := m.Called(ctx, project, domain, appName, limit, token)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).([]*flyteapp.App), args.String(1), args.Error(2)
}

func (m *mockAppK8sClient) GetReplicas(ctx context.Context, appID *flyteapp.Identifier) ([]*flyteapp.Replica, error) {
	args := m.Called(ctx, appID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*flyteapp.Replica), args.Error(1)
}

func (m *mockAppK8sClient) DeleteReplica(ctx context.Context, replicaID *flyteapp.ReplicaIdentifier) error {
	return m.Called(ctx, replicaID).Error(0)
}

func (m *mockAppK8sClient) Watch(ctx context.Context, project, domain, appName string) (<-chan *flyteapp.WatchResponse, error) {
	args := m.Called(ctx, project, domain, appName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(<-chan *flyteapp.WatchResponse), args.Error(1)
}

// --- helpers ---

func testCfg() *appconfig.AppConfig {
	return &appconfig.AppConfig{
		Enabled:               true,
		BaseDomain:            "apps.example.com",
		DefaultRequestTimeout: 5 * time.Minute,
		MaxRequestTimeout:     time.Hour,
	}
}

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
	}
}

func testStatus(phase flyteapp.Status_DeploymentStatus) *flyteapp.Status {
	return &flyteapp.Status{
		Conditions: []*flyteapp.Condition{
			{DeploymentStatus: phase},
		},
	}
}

func newTestClient(t *testing.T, k8s *mockAppK8sClient) appconnect.AppServiceClient {
	svc := NewInternalAppService(k8s, testCfg())
	path, handler := appconnect.NewAppServiceHandler(svc)
	mux := http.NewServeMux()
	mux.Handle("/internal"+path, http.StripPrefix("/internal", handler))
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return appconnect.NewAppServiceClient(http.DefaultClient, server.URL+"/internal")
}

// --- Create ---

func TestCreate_Success(t *testing.T) {
	k8s := &mockAppK8sClient{}
	svc := NewInternalAppService(k8s, testCfg())

	app := testApp()
	k8s.On("Deploy", mock.Anything, app).Return(nil)

	resp, err := svc.Create(context.Background(), connect.NewRequest(&flyteapp.CreateRequest{App: app}))
	require.NoError(t, err)
	assert.Equal(t, flyteapp.Status_DEPLOYMENT_STATUS_PENDING, resp.Msg.App.Status.Conditions[0].DeploymentStatus)
	assert.Equal(t, "https://myapp-proj-dev.apps.example.com", resp.Msg.App.Status.Ingress.PublicUrl)
	k8s.AssertExpectations(t)
}

func TestCreate_MissingID(t *testing.T) {
	svc := NewInternalAppService(&mockAppK8sClient{}, testCfg())

	_, err := svc.Create(context.Background(), connect.NewRequest(&flyteapp.CreateRequest{
		App: &flyteapp.App{Spec: testApp().Spec},
	}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}

func TestCreate_MissingSpec(t *testing.T) {
	svc := NewInternalAppService(&mockAppK8sClient{}, testCfg())

	_, err := svc.Create(context.Background(), connect.NewRequest(&flyteapp.CreateRequest{
		App: &flyteapp.App{Metadata: &flyteapp.Meta{Id: testAppID()}},
	}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}

func TestCreate_MissingPayload(t *testing.T) {
	svc := NewInternalAppService(&mockAppK8sClient{}, testCfg())

	_, err := svc.Create(context.Background(), connect.NewRequest(&flyteapp.CreateRequest{
		App: &flyteapp.App{
			Metadata: &flyteapp.Meta{Id: testAppID()},
			Spec:     &flyteapp.Spec{},
		},
	}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}

func TestCreate_NoBaseDomain_NoIngress(t *testing.T) {
	k8s := &mockAppK8sClient{}
	cfg := testCfg()
	cfg.BaseDomain = ""
	svc := NewInternalAppService(k8s, cfg)

	app := testApp()
	k8s.On("Deploy", mock.Anything, app).Return(nil)

	resp, err := svc.Create(context.Background(), connect.NewRequest(&flyteapp.CreateRequest{App: app}))
	require.NoError(t, err)
	assert.Nil(t, resp.Msg.App.Status.Ingress)
	k8s.AssertExpectations(t)
}

// --- Get ---

func TestGet_Success(t *testing.T) {
	k8s := &mockAppK8sClient{}
	svc := NewInternalAppService(k8s, testCfg())

	appID := testAppID()
	k8s.On("GetStatus", mock.Anything, appID).Return(testStatus(flyteapp.Status_DEPLOYMENT_STATUS_ACTIVE), nil)

	resp, err := svc.Get(context.Background(), connect.NewRequest(&flyteapp.GetRequest{
		Identifier: &flyteapp.GetRequest_AppId{AppId: appID},
	}))
	require.NoError(t, err)
	assert.Equal(t, flyteapp.Status_DEPLOYMENT_STATUS_ACTIVE, resp.Msg.App.Status.Conditions[0].DeploymentStatus)
	k8s.AssertExpectations(t)
}

func TestGet_NotFound(t *testing.T) {
	k8s := &mockAppK8sClient{}
	svc := NewInternalAppService(k8s, testCfg())

	appID := testAppID()
	notFoundErr := fmt.Errorf("KService myapp not found: %w", kerrors.NewNotFound(schema.GroupResource{}, "myapp"))
	k8s.On("GetStatus", mock.Anything, appID).Return(nil, notFoundErr)

	_, err := svc.Get(context.Background(), connect.NewRequest(&flyteapp.GetRequest{
		Identifier: &flyteapp.GetRequest_AppId{AppId: appID},
	}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
	k8s.AssertExpectations(t)
}

func TestGet_MissingAppID(t *testing.T) {
	svc := NewInternalAppService(&mockAppK8sClient{}, testCfg())

	_, err := svc.Get(context.Background(), connect.NewRequest(&flyteapp.GetRequest{}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}

// --- Update ---

func TestUpdate_Deploy(t *testing.T) {
	k8s := &mockAppK8sClient{}
	svc := NewInternalAppService(k8s, testCfg())

	app := testApp()
	k8s.On("Deploy", mock.Anything, app).Return(nil)
	k8s.On("GetStatus", mock.Anything, app.Metadata.Id).Return(testStatus(flyteapp.Status_DEPLOYMENT_STATUS_DEPLOYING), nil)

	resp, err := svc.Update(context.Background(), connect.NewRequest(&flyteapp.UpdateRequest{App: app}))
	require.NoError(t, err)
	assert.Equal(t, flyteapp.Status_DEPLOYMENT_STATUS_DEPLOYING, resp.Msg.App.Status.Conditions[0].DeploymentStatus)
	k8s.AssertExpectations(t)
}

func TestUpdate_Stop(t *testing.T) {
	k8s := &mockAppK8sClient{}
	svc := NewInternalAppService(k8s, testCfg())

	app := testApp()
	app.Spec.DesiredState = flyteapp.Spec_DESIRED_STATE_STOPPED
	k8s.On("Stop", mock.Anything, app.Metadata.Id).Return(nil)
	k8s.On("GetStatus", mock.Anything, app.Metadata.Id).Return(testStatus(flyteapp.Status_DEPLOYMENT_STATUS_STOPPED), nil)

	resp, err := svc.Update(context.Background(), connect.NewRequest(&flyteapp.UpdateRequest{App: app}))
	require.NoError(t, err)
	assert.Equal(t, flyteapp.Status_DEPLOYMENT_STATUS_STOPPED, resp.Msg.App.Status.Conditions[0].DeploymentStatus)
	k8s.AssertExpectations(t)
}

func TestUpdate_MissingID(t *testing.T) {
	svc := NewInternalAppService(&mockAppK8sClient{}, testCfg())

	_, err := svc.Update(context.Background(), connect.NewRequest(&flyteapp.UpdateRequest{
		App: &flyteapp.App{},
	}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}

// --- Delete ---

func TestDelete_Success(t *testing.T) {
	k8s := &mockAppK8sClient{}
	svc := NewInternalAppService(k8s, testCfg())

	appID := testAppID()
	k8s.On("Delete", mock.Anything, appID).Return(nil)

	_, err := svc.Delete(context.Background(), connect.NewRequest(&flyteapp.DeleteRequest{AppId: appID}))
	require.NoError(t, err)
	k8s.AssertExpectations(t)
}

func TestDelete_MissingID(t *testing.T) {
	svc := NewInternalAppService(&mockAppK8sClient{}, testCfg())

	_, err := svc.Delete(context.Background(), connect.NewRequest(&flyteapp.DeleteRequest{}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
}

// --- List ---

func TestList_ByProject(t *testing.T) {
	k8s := &mockAppK8sClient{}
	svc := NewInternalAppService(k8s, testCfg())

	apps := []*flyteapp.App{testApp()}
	k8s.On("List", mock.Anything, "proj", "dev", "", uint32(10), "tok").Return(apps, "nexttok", nil)

	resp, err := svc.List(context.Background(), connect.NewRequest(&flyteapp.ListRequest{
		FilterBy: &flyteapp.ListRequest_Project{
			Project: &common.ProjectIdentifier{Name: "proj", Domain: "dev"},
		},
		Request: &common.ListRequest{Limit: 10, Token: "tok"},
	}))
	require.NoError(t, err)
	assert.Len(t, resp.Msg.Apps, 1)
	assert.Equal(t, "nexttok", resp.Msg.Token)
	k8s.AssertExpectations(t)
}

func TestList_NoFilter(t *testing.T) {
	k8s := &mockAppK8sClient{}
	svc := NewInternalAppService(k8s, testCfg())

	k8s.On("List", mock.Anything, "", "", "", uint32(0), "").Return([]*flyteapp.App{}, "", nil)

	resp, err := svc.List(context.Background(), connect.NewRequest(&flyteapp.ListRequest{}))
	require.NoError(t, err)
	assert.Empty(t, resp.Msg.Apps)
	k8s.AssertExpectations(t)
}

// --- Watch ---

func TestWatch_InitialSnapshot(t *testing.T) {
	k8s := &mockAppK8sClient{}

	apps := []*flyteapp.App{testApp()}
	ch := make(chan *flyteapp.WatchResponse)
	close(ch)

	k8s.On("Watch", mock.Anything, "proj", "dev", "").Return((<-chan *flyteapp.WatchResponse)(ch), nil)
	k8s.On("List", mock.Anything, "proj", "dev", "", uint32(0), "").Return(apps, "", nil)

	client := newTestClient(t, k8s)
	stream, err := client.Watch(context.Background(), connect.NewRequest(&flyteapp.WatchRequest{
		Target: &flyteapp.WatchRequest_Project{
			Project: &common.ProjectIdentifier{Name: "proj", Domain: "dev"},
		},
	}))
	require.NoError(t, err)

	// Expect one CreateEvent from the initial snapshot.
	require.True(t, stream.Receive())
	resp := stream.Msg()
	ce, ok := resp.Event.(*flyteapp.WatchResponse_CreateEvent)
	require.True(t, ok)
	assert.Equal(t, "myapp", ce.CreateEvent.App.Metadata.Id.Name)

	// Channel is closed — stream should end.
	assert.False(t, stream.Receive())
	k8s.AssertExpectations(t)
}

func TestWatch_AppIDTarget(t *testing.T) {
	k8s := &mockAppK8sClient{}

	ch := make(chan *flyteapp.WatchResponse)
	close(ch)

	k8s.On("Watch", mock.Anything, "proj", "dev", "myapp").Return((<-chan *flyteapp.WatchResponse)(ch), nil)
	k8s.On("List", mock.Anything, "proj", "dev", "myapp", uint32(0), "").Return([]*flyteapp.App{}, "", nil)

	client := newTestClient(t, k8s)
	stream, err := client.Watch(context.Background(), connect.NewRequest(&flyteapp.WatchRequest{
		Target: &flyteapp.WatchRequest_AppId{AppId: testAppID()},
	}))
	require.NoError(t, err)

	// No snapshot apps, channel closed — stream ends immediately.
	assert.False(t, stream.Receive())
	k8s.AssertExpectations(t)
}
