package agent

import (
	"context"
	"testing"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	agentMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/webapi/agent/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func getMockMetadataServiceClient() *agentMocks.AgentMetadataServiceClient {
	mockMetadataServiceClient := new(agentMocks.AgentMetadataServiceClient)
	mockRequest := &admin.ListAgentsRequest{}
	mockResponse := &admin.ListAgentsResponse{
		Agents: []*admin.Agent{
			{
				Name:               "test-agent",
				SupportedTaskTypes: []string{"task1", "task2", "task3"},
			},
		},
	}

	mockMetadataServiceClient.On("ListAgents", mock.Anything, mockRequest).Return(mockResponse, nil)
	return mockMetadataServiceClient
}

// TODO, USE CREATE, GET DELETE FUNCTION TO MOCK THE OUTPUT
func getMockServiceClient() *agentMocks.AsyncAgentServiceClient {
	mockServiceClient := new(agentMocks.AsyncAgentServiceClient)
	mockRequest := &admin.ListAgentsRequest{}
	mockResponse := &admin.ListAgentsResponse{
		Agents: []*admin.Agent{
			{
				Name:               "test-agent",
				SupportedTaskTypes: []string{"task1", "task2", "task3"},
			},
		},
	}

	mockServiceClient.On("ListAgents", mock.Anything, mockRequest).Return(mockResponse, nil)
	return mockServiceClient
}

func mockGetBadAsyncClientFunc() *agentMocks.AsyncAgentServiceClient {
	return nil
}

func TestInitializeClientFunc(t *testing.T) {
	cfg := defaultConfig
	ctx := context.Background()
	err := SetConfig(&cfg)
	assert.NoError(t, err)
	cs, err := initializeClients(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, cs)
}
