package agent

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitializeClients(t *testing.T) {
	cfg := defaultConfig
	cfg.AgentDeployments = map[string]*Deployment{
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
	cs := getAgentClientSets(ctx)
	_, ok := cs.syncAgentClients["y"]
	assert.True(t, ok)
	_, ok = cs.asyncAgentClients["x"]
	assert.True(t, ok)
}
