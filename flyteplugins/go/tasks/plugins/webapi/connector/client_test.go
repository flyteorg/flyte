package connector

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
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
