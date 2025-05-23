package agent

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
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

func TestAgentForTaskTypesAlwaysOverwrite(t *testing.T) {
	deploymentX := Deployment{Endpoint: "x"}
	deploymentY := Deployment{Endpoint: "y"}
	deploymentZ := Deployment{Endpoint: "z"}
	cfg := defaultConfig
	cfg.AgentDeployments = map[string]*Deployment{
		"x": &deploymentX,
		"y": &deploymentY,
		"z": &deploymentZ,
	}
	cfg.AgentForTaskTypes = map[string]string{
		"task1": "x", // we expect the "task1" task type should always route to deploymentX
	}
	ctx := context.Background()
	err := SetConfig(&cfg)
	assert.NoError(t, err)
	cs := getAgentClientSets(ctx)

	// let's mock the "ListAgent" behaviour for 3 deployments
	// they both have SupportedTaskTypes "task1"
	mockClientForDeploymentX := mocks.NewAgentMetadataServiceClient(t)
	mockClientForDeploymentY := mocks.NewAgentMetadataServiceClient(t)
	mockClientForDeploymentZ := mocks.NewAgentMetadataServiceClient(t)
	mockClientForDeploymentX.On("ListAgents", mock.Anything, mock.Anything).Return(&admin.ListAgentsResponse{
		Agents: []*admin.Agent{
			{
				Name:               "agent1",
				SupportedTaskTypes: []string{"task1"},
			},
		},
	}, nil)
	mockClientForDeploymentY.On("ListAgents", mock.Anything, mock.Anything).Return(&admin.ListAgentsResponse{
		Agents: []*admin.Agent{
			{
				Name:               "agent2",
				SupportedTaskTypes: []string{"task1"},
			},
		},
	}, nil)
	mockClientForDeploymentZ.On("ListAgents", mock.Anything, mock.Anything).Return(&admin.ListAgentsResponse{
		Agents: []*admin.Agent{
			{
				Name:               "agent3",
				SupportedTaskTypes: []string{"task1"},
			},
		},
	}, nil)
	cs.agentMetadataClients[deploymentX.Endpoint] = mockClientForDeploymentX
	cs.agentMetadataClients[deploymentY.Endpoint] = mockClientForDeploymentY
	cs.agentMetadataClients[deploymentZ.Endpoint] = mockClientForDeploymentZ
	// while auto-discovery execute in getAgentRegistry function, the deployment of task1 will be amended to deploymentZ
	// but the always-overwrite policy will overwrite deployment of task1 back to deploymentX according to cfg.AgentForTaskTypes
	registry := getAgentRegistry(ctx, cs)
	finalDeployment := registry["task1"][defaultTaskTypeVersion].AgentDeployment
	expectedDeployment := &deploymentX
	assert.Equal(t, finalDeployment, expectedDeployment)
}
