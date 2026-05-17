package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy/dataproxyconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/logs/dataplane"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
)

// fakeDataProxyHandler is a minimal DataProxyServiceHandler for tests.
// Only TailLogs is overridden; all other methods return CodeUnimplemented via the embedded struct.
type fakeDataProxyHandler struct {
	dataproxyconnect.UnimplementedDataProxyServiceHandler
	tailLogsFn func(ctx context.Context, req *connect.Request[dataproxy.TailLogsRequest], stream *connect.ServerStream[dataproxy.TailLogsResponse]) error
}

func (f *fakeDataProxyHandler) TailLogs(ctx context.Context, req *connect.Request[dataproxy.TailLogsRequest], stream *connect.ServerStream[dataproxy.TailLogsResponse]) error {
	if f.tailLogsFn != nil {
		return f.tailLogsFn(ctx, req, stream)
	}
	return nil
}

var tailLogsActionID = &common.ActionIdentifier{
	Run: &common.RunIdentifier{
		Org:     "test-org",
		Project: "test-project",
		Domain:  "test-domain",
		Name:    "rtest12345",
	},
	Name: "action-1",
}

// newTailLogsTestClient wires a RunLogsService backed by a fake DataProxy server and
// returns a client connected to the RunLogsService.
func newTailLogsTestClient(t *testing.T, dpHandler dataproxyconnect.DataProxyServiceHandler) workflowconnect.RunLogsServiceClient {
	t.Helper()

	dpMux := http.NewServeMux()
	dpPath, dpHTTPHandler := dataproxyconnect.NewDataProxyServiceHandler(dpHandler)
	dpMux.Handle(dpPath, dpHTTPHandler)
	dpServer := httptest.NewServer(dpMux)
	t.Cleanup(dpServer.Close)

	dpClient := dataproxyconnect.NewDataProxyServiceClient(http.DefaultClient, dpServer.URL)
	svc := NewRunLogsService(dpClient)

	runLogsMux := http.NewServeMux()
	path, handler := workflowconnect.NewRunLogsServiceHandler(svc)
	runLogsMux.Handle(path, handler)
	server := httptest.NewServer(runLogsMux)
	t.Cleanup(server.Close)

	return workflowconnect.NewRunLogsServiceClient(http.DefaultClient, server.URL)
}

func TestTailLogs_HappyPath(t *testing.T) {
	dpHandler := &fakeDataProxyHandler{
		tailLogsFn: func(_ context.Context, _ *connect.Request[dataproxy.TailLogsRequest], stream *connect.ServerStream[dataproxy.TailLogsResponse]) error {
			return stream.Send(&dataproxy.TailLogsResponse{
				Logs: []*dataproxy.TailLogsResponse_Logs{
					{Lines: []*dataplane.LogLine{
						{Message: "hello world", Originator: dataplane.LogLineOriginator_USER},
					}},
				},
			})
		},
	}

	client := newTailLogsTestClient(t, dpHandler)

	stream, err := client.TailLogs(context.Background(), connect.NewRequest(&workflow.TailLogsRequest{
		ActionId: tailLogsActionID,
		Attempt:  1,
	}))
	assert.NoError(t, err)

	assert.True(t, stream.Receive())
	resp := stream.Msg()
	assert.Len(t, resp.Logs, 1)
	assert.Len(t, resp.Logs[0].Lines, 1)
	assert.Equal(t, "hello world", resp.Logs[0].Lines[0].Message)

	assert.False(t, stream.Receive())
	assert.NoError(t, stream.Err())
}

func TestTailLogs_MissingActionID(t *testing.T) {
	client := newTailLogsTestClient(t, &fakeDataProxyHandler{})

	stream, err := client.TailLogs(context.Background(), connect.NewRequest(&workflow.TailLogsRequest{
		Attempt: 1,
	}))
	assert.NoError(t, err)

	assert.False(t, stream.Receive())
	assert.Error(t, stream.Err())
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(stream.Err()))
}

func TestTailLogs_DataProxyError(t *testing.T) {
	dpHandler := &fakeDataProxyHandler{
		tailLogsFn: func(_ context.Context, _ *connect.Request[dataproxy.TailLogsRequest], _ *connect.ServerStream[dataproxy.TailLogsResponse]) error {
			return connect.NewError(connect.CodeNotFound, fmt.Errorf("action not found"))
		},
	}

	client := newTailLogsTestClient(t, dpHandler)

	stream, err := client.TailLogs(context.Background(), connect.NewRequest(&workflow.TailLogsRequest{
		ActionId: tailLogsActionID,
		Attempt:  1,
	}))
	assert.NoError(t, err)

	assert.False(t, stream.Receive())
	assert.Error(t, stream.Err())
}

func TestTailLogs_ConcurrencyLimit(t *testing.T) {
	dpMux := http.NewServeMux()
	dpPath, dpHTTPHandler := dataproxyconnect.NewDataProxyServiceHandler(&fakeDataProxyHandler{})
	dpMux.Handle(dpPath, dpHTTPHandler)
	dpServer := httptest.NewServer(dpMux)
	t.Cleanup(dpServer.Close)

	dpClient := dataproxyconnect.NewDataProxyServiceClient(http.DefaultClient, dpServer.URL)
	svc := NewRunLogsService(dpClient)
	// Exhaust the semaphore so the next request is rejected.
	svc.sem.Acquire(context.Background(), defaultMaxConcurrentStreams) //nolint:errcheck

	runLogsMux := http.NewServeMux()
	path, handler := workflowconnect.NewRunLogsServiceHandler(svc)
	runLogsMux.Handle(path, handler)
	server := httptest.NewServer(runLogsMux)
	t.Cleanup(server.Close)

	client := workflowconnect.NewRunLogsServiceClient(http.DefaultClient, server.URL)
	stream, err := client.TailLogs(context.Background(), connect.NewRequest(&workflow.TailLogsRequest{
		ActionId: tailLogsActionID,
		Attempt:  1,
	}))
	assert.NoError(t, err)

	assert.False(t, stream.Receive())
	assert.Error(t, stream.Err())
	assert.Equal(t, connect.CodeResourceExhausted, connect.CodeOf(stream.Err()))

	svc.sem.Release(defaultMaxConcurrentStreams)
}

func TestTailLogs_RequestForwardedToDataProxy(t *testing.T) {
	var capturedReq *dataproxy.TailLogsRequest
	dpHandler := &fakeDataProxyHandler{
		tailLogsFn: func(_ context.Context, req *connect.Request[dataproxy.TailLogsRequest], _ *connect.ServerStream[dataproxy.TailLogsResponse]) error {
			capturedReq = req.Msg
			return nil
		},
	}

	client := newTailLogsTestClient(t, dpHandler)

	stream, err := client.TailLogs(context.Background(), connect.NewRequest(&workflow.TailLogsRequest{
		ActionId: tailLogsActionID,
		Attempt:  3,
	}))
	assert.NoError(t, err)
	assert.False(t, stream.Receive())
	assert.NoError(t, stream.Err())

	assert.NotNil(t, capturedReq)
	assert.Equal(t, tailLogsActionID.GetName(), capturedReq.GetActionId().GetName())
	assert.Equal(t, uint32(3), capturedReq.GetAttempt())
}
