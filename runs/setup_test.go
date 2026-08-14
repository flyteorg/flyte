package runs

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/flytestdlib/sentryutils"
	taskpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task/taskconnect"
)

type stubTaskService struct {
	taskconnect.UnimplementedTaskServiceHandler
	err error
}

func (s *stubTaskService) DeployTask(context.Context, *connect.Request[taskpb.DeployTaskRequest]) (*connect.Response[taskpb.DeployTaskResponse], error) {
	if s.err != nil {
		return nil, s.err
	}
	return connect.NewResponse(&taskpb.DeployTaskResponse{}), nil
}

// newTestClient serves svc behind the same interceptor chain Setup installs.
func newTestClient(t *testing.T, svc taskconnect.TaskServiceHandler) taskconnect.TaskServiceClient {
	t.Helper()
	mux := http.NewServeMux()
	path, handler := taskconnect.NewTaskServiceHandler(svc, connect.WithInterceptors(
		sentryutils.Interceptor(sentryOperations),
		deployTriggerInterceptor(),
	))
	mux.Handle(path, handler)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return taskconnect.NewTaskServiceClient(srv.Client(), srv.URL)
}

// bindMockTransport swaps in a Sentry client whose events the test can read.
func bindMockTransport(t *testing.T) *sentry.MockTransport {
	t.Helper()
	transport := &sentry.MockTransport{}
	client, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:       "https://key@example.invalid/1",
		Transport: transport,
	})
	require.NoError(t, err)
	hub := sentry.CurrentHub()
	previous := hub.Client()
	hub.BindClient(client)
	t.Cleanup(func() { hub.BindClient(previous) })
	return transport
}

// The interceptors must never change what the caller sees.
func TestDeployTaskInterceptorsArePassThrough(t *testing.T) {
	bindMockTransport(t)
	client := newTestClient(t, &stubTaskService{})

	resp, err := client.DeployTask(context.Background(), connect.NewRequest(&taskpb.DeployTaskRequest{
		Triggers: []*taskpb.TaskTrigger{{Name: "nightly"}, {Name: "on-merge"}},
	}))
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

// A server-side failure is worth an exception; a client-caused one is not.
func TestDeployTaskReportsOnlyServerSideFailures(t *testing.T) {
	for name, tc := range map[string]struct {
		code       connect.Code
		wantEvents int
	}{
		"internal":         {connect.CodeInternal, 1},
		"unknown":          {connect.CodeUnknown, 1},
		"invalid argument": {connect.CodeInvalidArgument, 0},
		"not found":        {connect.CodeNotFound, 0},
	} {
		t.Run(name, func(t *testing.T) {
			transport := bindMockTransport(t)
			client := newTestClient(t, &stubTaskService{
				err: connect.NewError(tc.code, errors.New("boom")),
			})

			_, err := client.DeployTask(context.Background(), connect.NewRequest(&taskpb.DeployTaskRequest{}))
			require.Error(t, err)
			assert.Len(t, transport.Events(), tc.wantEvents)
		})
	}
}

// The operation names have to match the ones flyte-sdk emits, or an OSS count
// cannot be subtracted from an SDK count.
func TestSentryOperationNamesMatchSDK(t *testing.T) {
	assert.Equal(t, map[string]string{
		"/flyteidl2.workflow.RunService/CreateRun":        "create_run",
		"/flyteidl2.task.TaskService/DeployTask":          "deploy_task",
		"/flyteidl2.trigger.TriggerService/DeployTrigger": "deploy_trigger",
	}, sentryOperations)
}
