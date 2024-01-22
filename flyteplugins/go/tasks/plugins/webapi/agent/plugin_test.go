package agent

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/exp/maps"
	"google.golang.org/grpc"

	"github.com/flyteorg/flyte/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	flyteIdl "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	flyteIdlCore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery"
	pluginsCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	pluginCoreMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	ioMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	webapiPlugin "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/webapi/mocks"
	agentMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/webapi/agent/mocks"
	"github.com/flyteorg/flyte/flyteplugins/tests"
	"github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

func TestSyncTask(t *testing.T) {
	tCtx := getTaskContext(t)
	taskReader := new(pluginCoreMocks.TaskReader)

	template := flyteIdlCore.TaskTemplate{
		Type: "api_task",
	}

	taskReader.On("Read", mock.Anything).Return(&template, nil)

	tCtx.OnTaskReader().Return(taskReader)

	agentPlugin := newMockSyncAgentPlugin()
	pluginEntry := pluginmachinery.CreateRemotePlugin(agentPlugin)
	plugin, err := pluginEntry.LoadPlugin(context.TODO(), newFakeSetupContext("create_task_sync_test"))
	assert.NoError(t, err)

	inputs, err := coreutils.MakeLiteralMap(map[string]interface{}{"x": 1})
	assert.NoError(t, err)
	basePrefix := storage.DataReference("fake://bucket/prefix/")
	inputReader := &ioMocks.InputReader{}
	inputReader.OnGetInputPrefixPath().Return(basePrefix)
	inputReader.OnGetInputPath().Return(basePrefix + "/inputs.pb")
	inputReader.OnGetMatch(mock.Anything).Return(inputs, nil)
	tCtx.OnInputReader().Return(inputReader)

	phase := tests.RunPluginEndToEndTest(t, plugin, &template, inputs, nil, nil, nil)
	assert.Equal(t, true, phase.Phase().IsSuccess())
}

func TestPlugin(t *testing.T) {
	fakeSetupContext := pluginCoreMocks.SetupContext{}
	fakeSetupContext.OnMetricsScope().Return(promutils.NewScope("test"))

	cfg := defaultConfig
	cfg.WebAPI.Caching.Workers = 1
	cfg.WebAPI.Caching.ResyncInterval.Duration = 5 * time.Second
	cfg.DefaultAgent = Agent{Endpoint: "test-agent.flyte.svc.cluster.local:80"}
	cfg.Agents = map[string]*Agent{"spark_agent": {Endpoint: "localhost:80"}}
	cfg.AgentForTaskTypes = map[string]string{"spark": "spark_agent", "bar": "bar_agent"}

	plugin := Plugin{
		metricScope: fakeSetupContext.MetricsScope(),
		cfg:         GetConfig(),
	}
	t.Run("get config", func(t *testing.T) {
		err := SetConfig(&cfg)
		assert.NoError(t, err)
		assert.Equal(t, cfg.WebAPI, plugin.GetConfig())
	})
	t.Run("get ResourceRequirements", func(t *testing.T) {
		namespace, constraints, err := plugin.ResourceRequirements(context.TODO(), nil)
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.ResourceNamespace("default"), namespace)
		assert.Equal(t, plugin.cfg.ResourceConstraints, constraints)
	})

	t.Run("test newAgentPlugin", func(t *testing.T) {
		p := newMockAgentPlugin()
		assert.NotNil(t, p)
		assert.Equal(t, "agent-service", p.ID)
		assert.NotNil(t, p.PluginLoader)
	})

	t.Run("test getFinalAgent", func(t *testing.T) {
		agentRegistry := map[string]*Agent{"spark": {Endpoint: "localhost:80"}}
		agent := getFinalAgent("spark", &cfg, agentRegistry)
		assert.Equal(t, agent.Endpoint, "localhost:80")
		agent = getFinalAgent("foo", &cfg, agentRegistry)
		assert.Equal(t, agent.Endpoint, cfg.DefaultAgent.Endpoint)
		agent = getFinalAgent("bar", &cfg, agentRegistry)
		assert.Equal(t, agent.Endpoint, cfg.DefaultAgent.Endpoint)
	})

	t.Run("test getAgentMetadataClientFunc", func(t *testing.T) {
		client, err := getAgentMetadataClientFunc(context.Background(), &Agent{Endpoint: "localhost:80"}, map[*Agent]*grpc.ClientConn{})
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("test getClientFunc", func(t *testing.T) {
		client, err := getClientFunc(context.Background(), &Agent{Endpoint: "localhost:80"}, map[*Agent]*grpc.ClientConn{})
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("test getClientFunc more config", func(t *testing.T) {
		client, err := getClientFunc(context.Background(), &Agent{Endpoint: "localhost:80", Insecure: true, DefaultServiceConfig: "{\"loadBalancingConfig\": [{\"round_robin\":{}}]}"}, map[*Agent]*grpc.ClientConn{})
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("test getClientFunc cache hit", func(t *testing.T) {
		connectionCache := make(map[*Agent]*grpc.ClientConn)
		agent := &Agent{Endpoint: "localhost:80", Insecure: true, DefaultServiceConfig: "{\"loadBalancingConfig\": [{\"round_robin\":{}}]}"}

		client, err := getClientFunc(context.Background(), agent, connectionCache)
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.NotNil(t, client, connectionCache[agent])

		cachedClient, err := getClientFunc(context.Background(), agent, connectionCache)
		assert.NoError(t, err)
		assert.NotNil(t, cachedClient)
		assert.Equal(t, client, cachedClient)
	})

	t.Run("test getFinalTimeout", func(t *testing.T) {
		timeout := getFinalTimeout("CreateTask", &Agent{Endpoint: "localhost:8080", Timeouts: map[string]config.Duration{"CreateTask": {Duration: 1 * time.Millisecond}}})
		assert.Equal(t, 1*time.Millisecond, timeout.Duration)
		timeout = getFinalTimeout("DeleteTask", &Agent{Endpoint: "localhost:8080", DefaultTimeout: config.Duration{Duration: 10 * time.Second}})
		assert.Equal(t, 10*time.Second, timeout.Duration)
	})

	t.Run("test getFinalContext", func(t *testing.T) {
		ctx, _ := getFinalContext(context.TODO(), "DeleteTask", &Agent{})
		assert.Equal(t, context.TODO(), ctx)

		ctx, _ = getFinalContext(context.TODO(), "CreateTask", &Agent{Endpoint: "localhost:8080", Timeouts: map[string]config.Duration{"CreateTask": {Duration: 1 * time.Millisecond}}})
		assert.NotEqual(t, context.TODO(), ctx)
	})

	t.Run("test PENDING Status", func(t *testing.T) {
		taskContext := new(webapiPlugin.StatusContext)
		taskContext.On("Resource").Return(ResourceWrapper{
			State:    admin.State_PENDING,
			Outputs:  nil,
			Message:  "Waiting for cluster",
			LogLinks: []*flyteIdlCore.TaskLog{{Uri: "http://localhost:3000/log", Name: "Log Link"}},
		})

		phase, err := plugin.Status(context.Background(), taskContext)
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.PhaseInitializing, phase.Phase())
		assert.Equal(t, "Waiting for cluster", phase.Reason())
	})

	t.Run("test RUNNING Status", func(t *testing.T) {
		taskContext := new(webapiPlugin.StatusContext)
		taskContext.On("Resource").Return(ResourceWrapper{
			State:    admin.State_RUNNING,
			Outputs:  nil,
			Message:  "Job is running",
			LogLinks: []*flyteIdlCore.TaskLog{{Uri: "http://localhost:3000/log", Name: "Log Link"}},
		})

		phase, err := plugin.Status(context.Background(), taskContext)
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.PhaseRunning, phase.Phase())
	})

	t.Run("test PERMANENT_FAILURE Status", func(t *testing.T) {
		taskContext := new(webapiPlugin.StatusContext)
		taskContext.On("Resource").Return(ResourceWrapper{
			State:    admin.State_PERMANENT_FAILURE,
			Outputs:  nil,
			Message:  "",
			LogLinks: []*flyteIdlCore.TaskLog{{Uri: "http://localhost:3000/log", Name: "Log Link"}},
		})

		phase, err := plugin.Status(context.Background(), taskContext)
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.PhasePermanentFailure, phase.Phase())
	})

	t.Run("test RETRYABLE_FAILURE Status", func(t *testing.T) {
		taskContext := new(webapiPlugin.StatusContext)
		taskContext.On("Resource").Return(ResourceWrapper{
			State:    admin.State_RETRYABLE_FAILURE,
			Outputs:  nil,
			Message:  "",
			LogLinks: []*flyteIdlCore.TaskLog{{Uri: "http://localhost:3000/log", Name: "Log Link"}},
		})

		phase, err := plugin.Status(context.Background(), taskContext)
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.PhaseRetryableFailure, phase.Phase())
	})

	t.Run("test UNDEFINED Status", func(t *testing.T) {
		taskContext := new(webapiPlugin.StatusContext)
		taskContext.On("Resource").Return(ResourceWrapper{
			State:    5,
			Outputs:  nil,
			Message:  "",
			LogLinks: []*flyteIdlCore.TaskLog{{Uri: "http://localhost:3000/log", Name: "Log Link"}},
		})

		phase, err := plugin.Status(context.Background(), taskContext)
		assert.Error(t, err)
		assert.Equal(t, pluginsCore.PhaseUndefined, phase.Phase())
	})

	t.Run("test TaskExecution_UNDEFINED Status", func(t *testing.T) {
		taskContext := new(webapiPlugin.StatusContext)
		taskContext.On("Resource").Return(ResourceWrapper{
			Phase:    flyteIdl.TaskExecution_UNDEFINED,
			Outputs:  nil,
			Message:  "",
			LogLinks: []*flyteIdlCore.TaskLog{{Uri: "http://localhost:3000/log", Name: "Log Link"}},
		})

		phase, err := plugin.Status(context.Background(), taskContext)
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.PhaseRetryableFailure, phase.Phase())
	})

	t.Run("test TaskExecution_QUEUED Status", func(t *testing.T) {
		taskContext := new(webapiPlugin.StatusContext)
		taskContext.On("Resource").Return(ResourceWrapper{
			Phase:    flyteIdl.TaskExecution_QUEUED,
			Outputs:  nil,
			Message:  "",
			LogLinks: []*flyteIdlCore.TaskLog{{Uri: "http://localhost:3000/log", Name: "Log Link"}},
		})

		phase, err := plugin.Status(context.Background(), taskContext)
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.PhaseQueued, phase.Phase())
	})

	t.Run("test TaskExecution_WAITING_FOR_RESOURCES Status", func(t *testing.T) {
		taskContext := new(webapiPlugin.StatusContext)
		taskContext.On("Resource").Return(ResourceWrapper{
			Phase:    flyteIdl.TaskExecution_WAITING_FOR_RESOURCES,
			Outputs:  nil,
			Message:  "",
			LogLinks: []*flyteIdlCore.TaskLog{{Uri: "http://localhost:3000/log", Name: "Log Link"}},
		})

		phase, err := plugin.Status(context.Background(), taskContext)
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.PhaseWaitingForResources, phase.Phase())
	})

	t.Run("test TaskExecution_INITIALIZING Status", func(t *testing.T) {
		taskContext := new(webapiPlugin.StatusContext)
		taskContext.On("Resource").Return(ResourceWrapper{
			Phase:    flyteIdl.TaskExecution_INITIALIZING,
			Outputs:  nil,
			Message:  "",
			LogLinks: []*flyteIdlCore.TaskLog{{Uri: "http://localhost:3000/log", Name: "Log Link"}},
		})

		phase, err := plugin.Status(context.Background(), taskContext)
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.PhaseInitializing, phase.Phase())
	})

	t.Run("test TaskExecution_RUNNING Status", func(t *testing.T) {
		taskContext := new(webapiPlugin.StatusContext)
		taskContext.On("Resource").Return(ResourceWrapper{
			Phase:    flyteIdl.TaskExecution_RUNNING,
			Outputs:  nil,
			Message:  "",
			LogLinks: []*flyteIdlCore.TaskLog{{Uri: "http://localhost:3000/log", Name: "Log Link"}},
		})

		phase, err := plugin.Status(context.Background(), taskContext)
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.PhaseRunning, phase.Phase())
	})

	t.Run("test TaskExecution_ABORTED Status", func(t *testing.T) {
		taskContext := new(webapiPlugin.StatusContext)
		taskContext.On("Resource").Return(ResourceWrapper{
			Phase:    flyteIdl.TaskExecution_ABORTED,
			Outputs:  nil,
			Message:  "",
			LogLinks: []*flyteIdlCore.TaskLog{{Uri: "http://localhost:3000/log", Name: "Log Link"}},
		})

		phase, err := plugin.Status(context.Background(), taskContext)
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.PhasePermanentFailure, phase.Phase())
	})

	t.Run("test TaskExecution_FAILED Status", func(t *testing.T) {
		taskContext := new(webapiPlugin.StatusContext)
		taskContext.On("Resource").Return(ResourceWrapper{
			Phase:    flyteIdl.TaskExecution_FAILED,
			Outputs:  nil,
			Message:  "",
			LogLinks: []*flyteIdlCore.TaskLog{{Uri: "http://localhost:3000/log", Name: "Log Link"}},
		})

		phase, err := plugin.Status(context.Background(), taskContext)
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.PhasePermanentFailure, phase.Phase())
	})

	t.Run("test UNDEFINED Phase", func(t *testing.T) {
		taskContext := new(webapiPlugin.StatusContext)
		taskContext.On("Resource").Return(ResourceWrapper{
			State:    8,
			Outputs:  nil,
			Message:  "",
			LogLinks: []*flyteIdlCore.TaskLog{{Uri: "http://localhost:3000/log", Name: "Log Link"}},
		})

		phase, err := plugin.Status(context.Background(), taskContext)
		assert.Error(t, err)
		assert.Equal(t, pluginsCore.PhaseUndefined, phase.Phase())
	})
}

func TestInitializeAgentRegistry(t *testing.T) {
	mockClient := new(agentMocks.AgentMetadataServiceClient)
	mockRequest := &admin.ListAgentsRequest{}
	mockResponse := &admin.ListAgentsResponse{
		Agents: []*admin.Agent{
			{
				Name:               "test-agent",
				SupportedTaskTypes: []string{"task1", "task2", "task3"},
			},
		},
	}

	mockClient.On("ListAgents", mock.Anything, mockRequest).Return(mockResponse, nil)
	getAgentMetadataClientFunc := func(ctx context.Context, agent *Agent, connCache map[*Agent]*grpc.ClientConn) (service.AgentMetadataServiceClient, error) {
		return mockClient, nil
	}

	cfg := defaultConfig
	cfg.Agents = map[string]*Agent{"custom_agent": {Endpoint: "localhost:80"}}
	cfg.AgentForTaskTypes = map[string]string{"task1": "agent-deployment-1", "task2": "agent-deployment-2"}
	connectionCache := make(map[*Agent]*grpc.ClientConn)
	agentRegistry, err := initializeAgentRegistry(&cfg, connectionCache, getAgentMetadataClientFunc)
	assert.NoError(t, err)

	// In golang, the order of keys in a map is random. So, we sort the keys before asserting.
	agentRegistryKeys := maps.Keys(agentRegistry)
	sort.Strings(agentRegistryKeys)

	assert.Equal(t, agentRegistryKeys, []string{"task1", "task2", "task3"})
}
