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

func TestGetOrDialMetadataClientReusesConnection(t *testing.T) {
	// Cancelable context so getGrpcConnection's close goroutine is cleaned up
	// when the test finishes.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// A pre-existing cached client for endpoint "ep".
	conn, err := grpc.NewClient("passthrough:///ep", grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
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
