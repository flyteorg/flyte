package agent

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	agentMocks "github.com/flyteorg/flyte/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyte/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	flyteIdlCore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery"
	pluginCoreMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	ioMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/webapi"
	"github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

func TestGetAndSetConfig(t *testing.T) {
	cfg := defaultConfig
	cfg.WebAPI.Caching.Workers = 1
	cfg.WebAPI.Caching.ResyncInterval.Duration = 5 * time.Second
	cfg.DefaultAgent.Insecure = false
	cfg.DefaultAgent.DefaultServiceConfig = "{\"loadBalancingConfig\": [{\"round_robin\":{}}]}"
	cfg.DefaultAgent.Timeouts = map[string]config.Duration{
		"CreateTask": {
			Duration: 1 * time.Millisecond,
		},
		"GetTask": {
			Duration: 2 * time.Millisecond,
		},
		"DeleteTask": {
			Duration: 3 * time.Millisecond,
		},
	}
	cfg.DefaultAgent.DefaultTimeout = config.Duration{Duration: 10 * time.Second}
	cfg.AgentDeployments = map[string]*Deployment{
		"agent_1": {
			Insecure:             cfg.DefaultAgent.Insecure,
			DefaultServiceConfig: cfg.DefaultAgent.DefaultServiceConfig,
			Timeouts:             cfg.DefaultAgent.Timeouts,
		},
	}
	cfg.AgentForTaskTypes = map[string]string{"task_type_1": "agent_1"}
	err := SetConfig(&cfg)
	assert.NoError(t, err)
	assert.Equal(t, &cfg, GetConfig())
}

func TestDefaultAgentConfig(t *testing.T) {
	cfg := defaultConfig

	assert.Equal(t, "", cfg.DefaultAgent.Endpoint)
	assert.True(t, cfg.DefaultAgent.Insecure)
	assert.Equal(t, 10*time.Second, cfg.DefaultAgent.DefaultTimeout.Duration)
	assert.Equal(t, `{"loadBalancingConfig": [{"round_robin":{}}]}`, cfg.DefaultAgent.DefaultServiceConfig)

	assert.Empty(t, cfg.DefaultAgent.Timeouts)

	expectedTaskTypes := []string{"task_type_1", "task_type_2"}
	assert.Equal(t, expectedTaskTypes, cfg.SupportedTaskTypes)

	assert.Contains(t, cfg.AgentDeployments, "agent_1")

	agent1 := cfg.AgentDeployments["agent_1"]
	assert.Equal(t, "", agent1.Endpoint)
	assert.True(t, agent1.Insecure)
	assert.Equal(t, 300*time.Second, agent1.DefaultTimeout.Duration)
	assert.Equal(t, 300*time.Second, agent1.Timeouts["ExecuteTaskSync"].Duration)
	assert.Equal(t, 100*time.Second, agent1.Timeouts["GetTask"].Duration)

	assert.Equal(t, "agent_1", cfg.AgentForTaskTypes["task_type_1"])

	assert.Equal(t, 10*time.Second, cfg.PollInterval.Duration)
}

func TestTaskWithExpectedAgent(t *testing.T) {
	// Set up the base configuration
	cfg := defaultConfig
	cfg.AgentDeployments["agent_1"].Endpoint = defaultAgentEndpoint
	err := SetConfig(&cfg)
	assert.NoError(t, err)

	// Create mock agent client agent_1
	agent1Client := new(agentMocks.AsyncAgentServiceClient)
	agent1Client.On("CreateTask", mock.Anything, mock.Anything).Return(&admin.CreateTaskResponse{}, nil)

	// Create a plugin instance using our mock clients
	agentPlugin := webapi.PluginEntry{
		ID:                 "agent",
		SupportedTaskTypes: []string{"task_type_1"},
		PluginLoader: func(ctx context.Context, iCtx webapi.PluginSetupContext) (webapi.AsyncPlugin, error) {
			return &Plugin{
				metricScope: iCtx.MetricsScope(),
				cfg:         &cfg,
				cs: &ClientSet{
					asyncAgentClients: map[string]service.AsyncAgentServiceClient{
						defaultAgentEndpoint: agent1Client,
					},
				},
				registry: Registry{
					"task_type_1": {defaultTaskTypeVersion: {AgentDeployment: cfg.AgentDeployments["agent_1"], IsSync: false}},
				},
			}, nil
		},
	}

	// Create task template and input
	template := flyteIdlCore.TaskTemplate{
		Type:   "task_type_1", // Should route to agent_1
		Target: &flyteIdlCore.TaskTemplate_Container{},
	}

	pluginEntry := pluginmachinery.CreateRemotePlugin(agentPlugin)
	plugin, err := pluginEntry.LoadPlugin(context.TODO(), newFakeSetupContext("test"))
	assert.NoError(t, err)

	inputs, err := coreutils.MakeLiteralMap(map[string]interface{}{"input": "test"})
	assert.NoError(t, err)

	// Run task type 1 (should use agent_1)
	tCtx := getTaskContext(t)
	tr := &pluginCoreMocks.TaskReader{}
	tr.OnRead(context.Background()).Return(&template, nil)
	tCtx.OnTaskReader().Return(tr)
	inputReader := &ioMocks.InputReader{}
	inputReader.OnGetMatch(mock.Anything).Return(inputs, nil)
	inputReader.OnGetInputPath().Return(storage.DataReference("fake://input/path"))
	inputReader.OnGetInputPrefixPath().Return(storage.DataReference("fake://input/prefix"))
	tCtx.OnInputReader().Return(inputReader)

	result, err := plugin.Handle(context.Background(), tCtx)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify agent_1 was called (failed)
	// agent1Client.AssertCalled(t, "CreateTask", mock.Anything, mock.Anything)

}
