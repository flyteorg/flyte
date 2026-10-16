package service

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/flyteorg/flyte/v2/dataproxy/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/authorization"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy/dataproxyconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/project"
	projectMocks "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/project/projectconnect/mocks"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func authAction(project string) *common.ActionIdentifier {
	return &common.ActionIdentifier{Run: &common.RunIdentifier{Org: "org", Domain: "prod", Project: project, Name: "rtest12345"}, Name: "a0"}
}
func inputRequest(project string) *dataproxy.UploadInputsRequest {
	return &dataproxy.UploadInputsRequest{Id: &dataproxy.UploadInputsRequest_RunId{RunId: authAction(project).Run}, Task: &dataproxy.UploadInputsRequest_TaskSpec{TaskSpec: &task.TaskSpec{TaskTemplate: &core.TaskTemplate{Id: &core.Identifier{Name: "task"}, Metadata: &core.TaskMetadata{}}}}, Inputs: &task.Inputs{}}
}
func TestAuthorizationDeniesBeforeSideEffects(t *testing.T) {
	svc := NewService(config.DataProxyConfig{Authorization: authorization.Config{Enabled: true}}, nil, nil, nil, nil, nil, nil)
	ctx := authorization.WithScopes(context.Background(), []authorization.Scope{{Domain: "prod", Project: "payments"}})
	// Every dependency is nil: an accidental storage, downstream, or log call panics.
	operations := map[string]func(context.Context) error{
		"upload location": func(ctx context.Context) error {
			_, err := svc.CreateUploadLocation(ctx, connect.NewRequest(&dataproxy.CreateUploadLocationRequest{Domain: "prod", Project: "other"}))
			return err
		},
		"inputs": func(ctx context.Context) error {
			_, err := svc.UploadInputs(ctx, connect.NewRequest(inputRequest("other")))
			return err
		},
		"download": func(ctx context.Context) error {
			_, err := svc.CreateDownloadLink(ctx, connect.NewRequest(&dataproxy.CreateDownloadLinkRequest{Source: &dataproxy.CreateDownloadLinkRequest_ActionAttemptId{ActionAttemptId: &common.ActionAttemptIdentifier{ActionId: authAction("other"), Attempt: 1}}}))
			return err
		},
		"data": func(ctx context.Context) error {
			_, err := svc.GetActionData(ctx, connect.NewRequest(&dataproxy.GetActionDataRequest{ActionId: authAction("other")}))
			return err
		},
		"logs": func(ctx context.Context) error {
			return svc.TailLogs(ctx, connect.NewRequest(&dataproxy.TailLogsRequest{ActionId: authAction("other"), Attempt: 1}), nil)
		},
	}
	for name, op := range operations {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, connect.CodePermissionDenied, connect.CodeOf(op(ctx)))
			require.Equal(t, connect.CodePermissionDenied, connect.CodeOf(op(context.Background())))
		})
	}
	// Inline target scope does not authorize a task or trigger in a different scope.
	for _, req := range []*dataproxy.UploadInputsRequest{
		{Task: &dataproxy.UploadInputsRequest_TaskId{TaskId: &task.TaskIdentifier{Domain: "prod", Project: "other"}}},
		{Task: &dataproxy.UploadInputsRequest_TriggerName{TriggerName: &common.TriggerName{Domain: "prod", Project: "other"}}},
	} {
		_, err := svc.resolveTaskTemplate(ctx, req)
		require.Equal(t, connect.CodePermissionDenied, connect.CodeOf(err))
	}
	req := inputRequest("payments")
	req.BaseDir = "s3://another-project"
	_, err := svc.UploadInputs(ctx, connect.NewRequest(req))
	require.Equal(t, connect.CodePermissionDenied, connect.CodeOf(err))
	for _, path := range []string{"../../other", "%2e%2e/other", "/other", "dir/../other", "dir\\other"} {
		_, err := svc.CreateUploadLocation(ctx, connect.NewRequest(&dataproxy.CreateUploadLocationRequest{Domain: "prod", Project: "payments", Filename: "file", FilenameRoot: path}))
		require.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
	}
}

type scopedBackend struct {
	workflowconnect.UnimplementedRunServiceHandler
}

func (scopedBackend) GetActionDataURIs(ctx context.Context, req *connect.Request[workflow.GetActionDataURIsRequest]) (*connect.Response[workflow.GetActionDataURIsResponse], error) {
	id := req.Msg.GetActionId().GetRun()
	if err := authorization.Check(ctx, id.GetDomain(), id.GetProject()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&workflow.GetActionDataURIsResponse{}), nil
}
func (scopedBackend) GetActionLogContext(ctx context.Context, req *connect.Request[workflow.GetActionLogContextRequest]) (*connect.Response[workflow.GetActionLogContextResponse], error) {
	id := req.Msg.GetActionId().GetRun()
	if err := authorization.Check(ctx, id.GetDomain(), id.GetProject()); err != nil {
		return nil, err
	}
	return connect.NewResponse(&workflow.GetActionLogContextResponse{LogContext: &core.LogContext{PrimaryPodName: "allowed"}}), nil
}
func authToken() string {
	return "header." + base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"alice","groups":"reader"}`)) + ".signature"
}
func TestAuthorizationAcrossDataProxyAndRuns(t *testing.T) {
	// The data proxy accepts two projects; the downstream grants only payments.
	// The original claims must be re-authorized for the actual Runs procedure.
	policy := authorization.Config{Enabled: true, ClaimsHeader: "X-Verified-Claims", Roles: []authorization.Role{
		{Name: "proxy", Claim: "groups", Values: []string{"reader"}, Procedures: []string{"/flyteidl2.dataproxy.DataProxyService/GetActionData", "/flyteidl2.dataproxy.DataProxyService/TailLogs"}, Scopes: []string{"prod:payments", "prod:reports"}},
		{Name: "runs", Claim: "groups", Values: []string{"reader"}, Procedures: []string{"/flyteidl2.workflow.RunService/GetActionDataURIs", "/flyteidl2.workflow.RunService/GetActionLogContext"}, Scopes: []string{"prod:payments"}},
	}}
	auth, err := authorization.NewInterceptor(policy, true)
	require.NoError(t, err)
	path, handler := workflowconnect.NewRunServiceHandler(scopedBackend{}, connect.WithInterceptors(auth))
	mux := http.NewServeMux()
	mux.Handle(path, handler)
	backend := httptest.NewServer(mux)
	defer backend.Close()
	runClient := workflowconnect.NewRunServiceClient(backend.Client(), backend.URL, connect.WithInterceptors(authorization.ForwardClaims()))
	svc := NewService(config.DataProxyConfig{Authorization: policy}, nil, nil, nil, runClient, nil, scopeLogStreamer{})
	path, handler = dataproxyconnect.NewDataProxyServiceHandler(svc, connect.WithInterceptors(auth))
	mux = http.NewServeMux()
	mux.Handle(path, handler)
	proxy := httptest.NewServer(mux)
	defer proxy.Close()
	client := dataproxyconnect.NewDataProxyServiceClient(proxy.Client(), proxy.URL)
	for _, test := range []struct {
		project, token string
		code           connect.Code
	}{
		{"payments", authToken(), 0}, {"reports", authToken(), connect.CodePermissionDenied}, {"other", authToken(), connect.CodePermissionDenied}, {"payments", "", connect.CodeUnauthenticated},
	} {
		req := connect.NewRequest(&dataproxy.GetActionDataRequest{ActionId: authAction(test.project)})
		if test.token != "" {
			req.Header().Set("X-Verified-Claims", test.token)
		}
		_, err := client.GetActionData(context.Background(), req)
		if test.code == 0 {
			require.NoError(t, err)
		} else {
			require.Equal(t, test.code, connect.CodeOf(err))
		}
	}
	// Permission for reading data does not grant uploading.
	upload := connect.NewRequest(&dataproxy.CreateUploadLocationRequest{Project: "payments", Domain: "prod"})
	upload.Header().Set("X-Verified-Claims", authToken())
	_, err = client.CreateUploadLocation(context.Background(), upload)
	require.Equal(t, connect.CodePermissionDenied, connect.CodeOf(err))
	// An allowed stream carries the caller through the Runs lookup to the log source.
	allowedLogs := connect.NewRequest(&dataproxy.TailLogsRequest{ActionId: authAction("payments"), Attempt: 1})
	allowedLogs.Header().Set("X-Verified-Claims", authToken())
	allowedStream, err := client.TailLogs(context.Background(), allowedLogs)
	require.NoError(t, err)
	require.True(t, allowedStream.Receive())
	require.False(t, allowedStream.Receive())
	require.NoError(t, allowedStream.Err())
	require.NoError(t, allowedStream.Close())
	// Streaming also authorizes before opening the log source.
	logs := connect.NewRequest(&dataproxy.TailLogsRequest{ActionId: authAction("other"), Attempt: 1})
	logs.Header().Set("X-Verified-Claims", authToken())
	stream, err := client.TailLogs(context.Background(), logs)
	if err == nil {
		require.False(t, stream.Receive())
		err = stream.Err()
		_ = stream.Close()
	}
	require.Equal(t, connect.CodePermissionDenied, connect.CodeOf(err))
}

type scopeLogStreamer struct{}

func (scopeLogStreamer) TailLogs(ctx context.Context, _ *core.LogContext, stream *connect.ServerStream[dataproxy.TailLogsResponse]) error {
	if err := authorization.Check(ctx, "prod", "payments"); err != nil {
		return err
	}
	return stream.Send(&dataproxy.TailLogsResponse{})
}

func TestAuthorizationAllowsScopedUpload(t *testing.T) {
	projectClient := projectMocks.NewProjectServiceClient(t)
	projectClient.EXPECT().GetProject(mock.Anything, mock.Anything).Return(connect.NewResponse(&project.GetProjectResponse{}), nil)
	svc := NewService(config.DataProxyConfig{Authorization: authorization.Config{Enabled: true}, Upload: config.DataProxyUploadConfig{StoragePrefix: "uploads"}}, setupMockDataStore(t), nil, nil, nil, projectClient, nil)
	ctx := authorization.WithScopes(context.Background(), []authorization.Scope{{Domain: "prod", Project: "payments"}})
	resp, err := svc.CreateUploadLocation(ctx, connect.NewRequest(&dataproxy.CreateUploadLocationRequest{Project: "payments", Domain: "prod", FilenameRoot: "root", Filename: "file"}))
	require.NoError(t, err)
	require.Contains(t, resp.Msg.NativeUrl, "uploads/payments/prod/root/file")
	require.NotEmpty(t, resp.Msg.SignedUrl)
}
