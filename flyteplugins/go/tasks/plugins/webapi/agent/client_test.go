package agent

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
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
	cs := &ClientSet{
		asyncAgentClients:    make(map[string]service.AsyncAgentServiceClient),
		syncAgentClients:     make(map[string]service.SyncAgentServiceClient),
		agentMetadataClients: make(map[string]service.AgentMetadataServiceClient),
	}
	updateAgentClientSets(ctx, cs)
	_, ok := cs.syncAgentClients["y"]
	assert.True(t, ok)
	_, ok = cs.asyncAgentClients["x"]
	assert.True(t, ok)
}
