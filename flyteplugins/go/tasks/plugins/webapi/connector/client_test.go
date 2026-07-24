package connector

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

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
	cs := getConnectorClientSets(ctx)
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

func TestGetConnectorMetadataClientReusesConnection(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// A pre-existing cached client for endpoint "ep".
	conn, err := grpc.NewClient("passthrough:///ep", grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })
	existing := connectorpb.NewConnectorMetadataServiceClient(conn)

	p := &Plugin{cs: &ClientSet{
		asyncConnectorClients:    map[string]connectorpb.AsyncConnectorServiceClient{},
		connectorMetadataClients: map[string]connectorpb.ConnectorMetadataServiceClient{"ep": existing},
	}}

	// Already cached -> must reuse the same client, not re-dial/replace it.
	assert.NoError(t, p.getConnectorMetadataClient(ctx, &Deployment{Endpoint: "ep"}))
	assert.True(t, existing == p.cs.connectorMetadataClients["ep"], "cached connection should be reused, not re-dialed")

	// New endpoint -> dialed and cached.
	assert.NoError(t, p.getConnectorMetadataClient(ctx, &Deployment{Endpoint: "passthrough:///new", Insecure: true}))
	_, ok := p.cs.connectorMetadataClients["passthrough:///new"]
	assert.True(t, ok, "new endpoint should be dialed and cached")
}
