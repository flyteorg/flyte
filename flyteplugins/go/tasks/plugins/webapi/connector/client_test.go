package connector

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	connectorpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/connector"
)

func TestInitializeClients(t *testing.T) {
	cfg := defaultConfig
	cfg.ConnectorDeployments = map[string]*Deployment{
		"x": {
			Endpoint: "x",
		},
		"y": {
			Endpoint: "y",
		},
	}
	ctx := context.Background()
	err := SetConfig(&cfg)
	assert.NoError(t, err)
	cs := getConnectorClientSets(ctx, nil)
	_, ok := cs.asyncConnectorClients["y"]
	assert.True(t, ok)
	_, ok = cs.asyncConnectorClients["x"]
	assert.True(t, ok)
}

func TestAllConnectorDeployments(t *testing.T) {
	cfg := defaultConfig
	cfg.DefaultConnector = Deployment{Endpoint: "default"}
	cfg.ConnectorDeployments = map[string]*Deployment{"x": {Endpoint: "x"}}
	cfg.ConnectorApps = map[string]*Deployment{"app": {Endpoint: "app"}}

	endpoints := make([]string, 0)
	for _, d := range allConnectorDeployments(&cfg) {
		endpoints = append(endpoints, d.Endpoint)
	}
	assert.ElementsMatch(t, []string{"default", "x", "app"}, endpoints)
}

func TestGetOrDialMetadataClientReusesConnection(t *testing.T) {
	// Cancelable context so getGrpcConnection's close goroutine is cleaned up
	// when the test finishes.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// A pre-existing cached client for endpoint "ep".
	conn, err := grpc.NewClient("passthrough:///ep", grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	existing := connectorpb.NewConnectorMetadataServiceClient(conn)

	cs := &ClientSet{
		asyncConnectorClients:    map[string]connectorpb.AsyncConnectorServiceClient{},
		connectorMetadataClients: map[string]connectorpb.ConnectorMetadataServiceClient{"ep": existing},
	}

	// Already cached -> must reuse the same client, not re-dial/replace it.
	got, err := cs.getOrDialMetadataClient(ctx, &Deployment{Endpoint: "ep"})
	assert.NoError(t, err)
	assert.True(t, existing == got, "cached connection should be reused, not re-dialed")

	// New endpoint -> dialed and cached. The passthrough scheme avoids DNS.
	_, err = cs.getOrDialMetadataClient(ctx, &Deployment{Endpoint: "passthrough:///new", Insecure: true})
	assert.NoError(t, err)
	_, ok := cs.metadataClient("passthrough:///new")
	assert.True(t, ok, "new endpoint should be dialed and cached")
}

func TestConnectorRPCUsesServiceTracerProvider(t *testing.T) {
	traceparent := make(chan string, 1)
	server := grpc.NewServer(grpc.UnaryInterceptor(func(ctx context.Context, _ any, _ *grpc.UnaryServerInfo, _ grpc.UnaryHandler) (any, error) {
		headers := metadata.ValueFromIncomingContext(ctx, "traceparent")
		if len(headers) == 1 {
			traceparent <- headers[0]
		}
		return &connectorpb.ListConnectorsResponse{}, nil
	}))
	connectorpb.RegisterConnectorMetadataServiceServer(server, connectorpb.UnimplementedConnectorMetadataServiceServer{})
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	go server.Serve(listener) //nolint:errcheck
	defer server.Stop()

	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	cs := &ClientSet{tracerProvider: provider, connectorMetadataClients: make(map[string]connectorpb.ConnectorMetadataServiceClient)}
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	client, err := cs.getOrDialMetadataClient(ctx, &Deployment{Endpoint: listener.Addr().String(), Insecure: true})
	require.NoError(t, err)
	ctx, parent := provider.Tracer("test").Start(ctx, "executor")
	defer parent.End()
	_, err = client.ListConnectors(ctx, &connectorpb.ListConnectorsRequest{})
	require.NoError(t, err)

	spans := recorder.Ended()
	require.Len(t, spans, 1)
	clientSpan := spans[0]
	require.Equal(t, trace.SpanKindClient, clientSpan.SpanKind())
	require.Equal(t, parent.SpanContext().SpanID(), clientSpan.Parent().SpanID())
	require.Equal(t, parent.SpanContext().TraceID(), clientSpan.SpanContext().TraceID())
	require.Len(t, traceparent, 1)
	require.Equal(t, "00-"+parent.SpanContext().TraceID().String()+"-"+clientSpan.SpanContext().SpanID().String()+"-01", <-traceparent)
}
