package flyte

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"

	auth "github.com/flyteorg/flyte/v2/flytestdlib/authorization"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy/dataproxyconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
)

func interceptor(t *testing.T, policy string) *auth.Interceptor {
	t.Helper()
	path := filepath.Join(t.TempDir(), "policy.cedar")
	require.NoError(t, os.WriteFile(path, []byte(policy), 0600))
	i, err := auth.New(auth.Config{Enabled: true, TrustForwardedClaims: true, ClaimsHeader: "Authorization", Issuer: "issuer", PolicyFile: path}, Resolve)
	require.NoError(t, err)
	return i
}

func claims(h http.Header) {
	h.Set("Authorization", "Bearer e30."+base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"alice","iss":"issuer"}`))+".sig")
}

func TestUnaryAndStreamingEnforceBeforeHandler(t *testing.T) {
	i := interceptor(t, `permit(principal == Flyte::Principal::"alice", action == Flyte::Action::"RunRead", resource == Flyte::Scope::"payments/prod");`)
	var reads, watches atomic.Int32
	mux := http.NewServeMux()
	mux.Handle(workflowconnect.RunServiceGetRunDetailsProcedure, connect.NewUnaryHandler(workflowconnect.RunServiceGetRunDetailsProcedure,
		func(context.Context, *connect.Request[workflow.GetRunDetailsRequest]) (*connect.Response[workflow.GetRunDetailsResponse], error) {
			reads.Add(1)
			return connect.NewResponse(&workflow.GetRunDetailsResponse{}), nil
		}, connect.WithInterceptors(i)))
	mux.Handle(workflowconnect.RunServiceWatchRunDetailsProcedure, connect.NewServerStreamHandler(workflowconnect.RunServiceWatchRunDetailsProcedure,
		func(_ context.Context, _ *connect.Request[workflow.WatchRunDetailsRequest], stream *connect.ServerStream[workflow.WatchRunDetailsResponse]) error {
			watches.Add(1)
			return stream.Send(&workflow.WatchRunDetailsResponse{})
		}, connect.WithInterceptors(i)))
	srv := httptest.NewServer(mux)
	defer srv.Close()
	client := workflowconnect.NewRunServiceClient(srv.Client(), srv.URL)
	for _, tc := range []struct {
		project                string
		authenticated, allowed bool
	}{{"payments", true, true}, {"other", true, false}, {"payments", false, false}} {
		rid := &common.RunIdentifier{Project: tc.project, Domain: "prod", Name: "run"}
		req := connect.NewRequest(&workflow.GetRunDetailsRequest{RunId: rid})
		if tc.authenticated {
			claims(req.Header())
		}
		_, err := client.GetRunDetails(context.Background(), req)
		if tc.allowed {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}
		watch := connect.NewRequest(&workflow.WatchRunDetailsRequest{RunId: rid})
		if tc.authenticated {
			claims(watch.Header())
		}
		stream, err := client.WatchRunDetails(context.Background(), watch)
		if err == nil {
			for stream.Receive() {
			}
			err = stream.Err()
			require.NoError(t, stream.Close())
		}
		if tc.allowed {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}
	}
	require.EqualValues(t, 1, reads.Load())
	require.EqualValues(t, 1, watches.Load())
}

func TestDownstreamReauthorizesForwardedCaller(t *testing.T) {
	// DataProxy may read data, but that does not grant RunRead downstream.
	i := interceptor(t, `permit(principal, action == Flyte::Action::"DataRead", resource == Flyte::Scope::"payments/prod");`)
	var reached atomic.Bool
	backend := httptest.NewServer(connect.NewUnaryHandler(workflowconnect.RunServiceGetRunDetailsProcedure,
		func(context.Context, *connect.Request[workflow.GetRunDetailsRequest]) (*connect.Response[workflow.GetRunDetailsResponse], error) {
			reached.Store(true)
			return connect.NewResponse(&workflow.GetRunDetailsResponse{}), nil
		}, connect.WithInterceptors(i)))
	defer backend.Close()
	runs := workflowconnect.NewRunServiceClient(backend.Client(), backend.URL, connect.WithInterceptors(auth.ForwardClaims()))
	proxy := httptest.NewServer(connect.NewUnaryHandler(dataproxyconnect.DataProxyServiceGetActionDataProcedure,
		func(ctx context.Context, req *connect.Request[dataproxy.GetActionDataRequest]) (*connect.Response[dataproxy.GetActionDataResponse], error) {
			_, err := runs.GetRunDetails(ctx, connect.NewRequest(&workflow.GetRunDetailsRequest{RunId: req.Msg.GetActionId().GetRun()}))
			return nil, err
		}, connect.WithInterceptors(i)))
	defer proxy.Close()
	client := dataproxyconnect.NewDataProxyServiceClient(proxy.Client(), proxy.URL)
	req := connect.NewRequest(&dataproxy.GetActionDataRequest{ActionId: &common.ActionIdentifier{Run: &common.RunIdentifier{Project: "payments", Domain: "prod", Name: "run"}}})
	claims(req.Header())
	_, err := client.GetActionData(context.Background(), req)
	require.Equal(t, connect.CodePermissionDenied, connect.CodeOf(err))
	require.False(t, reached.Load())
}

func TestCompoundOperationDeniedBeforeMutation(t *testing.T) {
	i := interceptor(t, `permit(principal, action in [Flyte::Action::"RunCreate", Flyte::Action::"TaskExecute"], resource == Flyte::Scope::"payments/prod");`)
	var mutations atomic.Int32
	srv := httptest.NewServer(connect.NewUnaryHandler(workflowconnect.RunServiceCreateRunProcedure,
		func(context.Context, *connect.Request[workflow.CreateRunRequest]) (*connect.Response[workflow.CreateRunResponse], error) {
			mutations.Add(1)
			return connect.NewResponse(&workflow.CreateRunResponse{}), nil
		}, connect.WithInterceptors(i)))
	defer srv.Close()
	client := workflowconnect.NewRunServiceClient(srv.Client(), srv.URL)
	for _, project := range []string{"payments", "other"} {
		req := connect.NewRequest(&workflow.CreateRunRequest{
			Id:   &workflow.CreateRunRequest_ProjectId{ProjectId: &common.ProjectIdentifier{Name: "payments", Domain: "prod"}},
			Task: &workflow.CreateRunRequest_TaskId{TaskId: &task.TaskIdentifier{Project: project, Domain: "prod", Name: "task", Version: "v1"}},
		})
		claims(req.Header())
		_, err := client.CreateRun(context.Background(), req)
		if project == "payments" {
			require.NoError(t, err)
		} else {
			require.Equal(t, connect.CodePermissionDenied, connect.CodeOf(err))
		}
	}
	require.EqualValues(t, 1, mutations.Load())
}

func TestScopedReaderCannotBroadenList(t *testing.T) {
	i := interceptor(t, `permit(principal, action == Flyte::Action::"RunRead", resource == Flyte::Scope::"payments/prod");`)
	var queries atomic.Int32
	srv := httptest.NewServer(connect.NewUnaryHandler(workflowconnect.RunServiceListRunsProcedure,
		func(context.Context, *connect.Request[workflow.ListRunsRequest]) (*connect.Response[workflow.ListRunsResponse], error) {
			queries.Add(1)
			return connect.NewResponse(&workflow.ListRunsResponse{}), nil
		}, connect.WithInterceptors(i)))
	defer srv.Close()
	client := workflowconnect.NewRunServiceClient(srv.Client(), srv.URL)
	for _, msg := range []*workflow.ListRunsRequest{
		{ScopeBy: &workflow.ListRunsRequest_Org{Org: "org"}},
		{ScopeBy: &workflow.ListRunsRequest_TaskName{TaskName: &task.TaskName{Project: "payments", Domain: "prod", Name: "task"}}},
		{},
	} {
		req := connect.NewRequest(msg)
		claims(req.Header())
		_, err := client.ListRuns(context.Background(), req)
		require.Equal(t, connect.CodePermissionDenied, connect.CodeOf(err))
	}
	require.Zero(t, queries.Load())
}

func TestInternalStreamRequiresRuntimePermission(t *testing.T) {
	i := interceptor(t, `permit(principal, action == Flyte::Action::"RunRead", resource);`)
	var handled atomic.Int32
	srv := httptest.NewUnstartedServer(connect.NewBidiStreamHandler(workflowconnect.InternalRunServiceRecordActionStreamProcedure,
		func(_ context.Context, stream *connect.BidiStream[workflow.RecordActionStreamRequest, workflow.RecordActionStreamResponse]) error {
			for {
				_, err := stream.Receive()
				if err != nil {
					return err
				}
				handled.Add(1)
				if err := stream.Send(&workflow.RecordActionStreamResponse{}); err != nil {
					return err
				}
			}
		}, connect.WithInterceptors(i)))
	srv.EnableHTTP2 = true
	srv.StartTLS()
	defer srv.Close()
	client := workflowconnect.NewInternalRunServiceClient(srv.Client(), srv.URL)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream := client.RecordActionStream(ctx)
	claims(stream.RequestHeader())
	_ = stream.Send(&workflow.RecordActionStreamRequest{Request: &workflow.RecordActionRequest{}})
	_, err := stream.Receive()
	require.Equal(t, connect.CodePermissionDenied, connect.CodeOf(err))
	require.Zero(t, handled.Load())
	_ = stream.CloseRequest()
	_ = stream.CloseResponse()
}
