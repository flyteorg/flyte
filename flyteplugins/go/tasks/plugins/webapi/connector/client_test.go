package connector

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

func TestAgentForTaskTypesAlwaysOverwrite(t *testing.T) {
	deploymentX := Deployment{Endpoint: "x"}
	deploymentY := Deployment{Endpoint: "y"}
	deploymentZ := Deployment{Endpoint: "z"}
	cfg := defaultConfig
	cfg.ConnectorDeployments = map[string]*Deployment{
		"x": &deploymentX,
		"y": &deploymentY,
		"z": &deploymentZ,
	}
	cfg.ConnectorForTaskTypes = map[string]string{
		"task1": "x", // we expect the "task1" task type should always route to deploymentX
	}
	ctx := context.Background()
	err := SetConfig(&cfg)
	assert.NoError(t, err)
	cs := getConnectorClientSets(ctx)

	// let's mock the "ListAgent" behaviour for 3 deployments
	// they both have SupportedTaskTypes "task1"
	mockClientForDeploymentX := mocks.NewAgentMetadataServiceClient(t)
	mockClientForDeploymentY := mocks.NewAgentMetadataServiceClient(t)
	mockClientForDeploymentZ := mocks.NewAgentMetadataServiceClient(t)
	mockClientForDeploymentX.On("ListAgents", mock.Anything, mock.Anything).Return(&admin.ListAgentsResponse{
		Agents: []*admin.Agent{
			{
				Name:               "connector1",
				SupportedTaskTypes: []string{"task1"},
			},
		},
	}, nil)
	mockClientForDeploymentY.On("ListAgents", mock.Anything, mock.Anything).Return(&admin.ListAgentsResponse{
		Agents: []*admin.Agent{
			{
				Name:               "connector2",
				SupportedTaskTypes: []string{"task1"},
			},
		},
	}, nil)
	mockClientForDeploymentZ.On("ListAgents", mock.Anything, mock.Anything).Return(&admin.ListAgentsResponse{
		Agents: []*admin.Agent{
			{
				Name:               "connector3",
				SupportedTaskTypes: []string{"task1"},
			},
		},
	}, nil)
	cs.connectorMetadataClients[deploymentX.Endpoint] = mockClientForDeploymentX
	cs.connectorMetadataClients[deploymentY.Endpoint] = mockClientForDeploymentY
	cs.connectorMetadataClients[deploymentZ.Endpoint] = mockClientForDeploymentZ
	// while auto-discovery execute in getAgentRegistry function, the deployment of task1 will be amended to deploymentZ
	// but the always-overwrite policy will overwrite deployment of task1 back to deploymentX according to cfg.AgentForTaskTypes
	registry := getConnectorRegistry(ctx, cs)
	finalDeployment := registry["task1"][defaultTaskTypeVersion].ConnectorDeployment
	expectedDeployment := &deploymentX
	assert.Equal(t, finalDeployment, expectedDeployment)
}
