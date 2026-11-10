package otelutils

import (
	"context"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/baggage"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestConnectTracePropagation(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	interceptor, err := NewConnectInterceptor(otelconnect.WithTracerProvider(provider))
	require.NoError(t, err)
	const procedure = "/test.Service/Ping"
	option := connect.WithInterceptors(interceptor)
	handler := connect.NewUnaryHandler(procedure, func(ctx context.Context, _ *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error) {
		response := connect.NewResponse(&emptypb.Empty{})
		response.Header().Set("X-Observed-Baggage", baggage.FromContext(ctx).Member("project").Value())
		return response, nil
	}, option)
	server := httptest.NewServer(handler)
	defer server.Close()
	client := connect.NewClient[emptypb.Empty, emptypb.Empty](server.Client(), server.URL+procedure, option)

	bag, err := baggage.Parse("project=flyte")
	require.NoError(t, err)
	ctx, root := provider.Tracer("test").Start(baggage.ContextWithBaggage(t.Context(), bag), "root")
	defer root.End()
	response, err := client.CallUnary(ctx, connect.NewRequest(&emptypb.Empty{}))
	require.NoError(t, err)
	require.Equal(t, "flyte", response.Header().Get("X-Observed-Baggage"))

	spans := recorder.Ended()
	require.Len(t, spans, 2)
	serverSpan, clientSpan := spans[0], spans[1]
	require.Equal(t, trace.SpanKindServer, serverSpan.SpanKind())
	require.Equal(t, trace.SpanKindClient, clientSpan.SpanKind())
	require.Equal(t, root.SpanContext().SpanID(), clientSpan.Parent().SpanID())
	require.Equal(t, root.SpanContext().TraceID(), serverSpan.SpanContext().TraceID())
	require.Equal(t, clientSpan.SpanContext().SpanID(), serverSpan.Parent().SpanID())
	require.True(t, serverSpan.Parent().IsRemote())
}
