package agent

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/types/known/structpb"

	agentMocks "github.com/flyteorg/flyte/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	flyteIdl "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	flyteIdlCore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	pluginsCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	pluginCoreMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	webapiPlugin "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/webapi/mocks"
	"github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

const defaultAgentEndpoint = "localhost:8000"

func TestPlugin(t *testing.T) {
	fakeSetupContext := pluginCoreMocks.SetupContext{}
	fakeSetupContext.OnMetricsScope().Return(promutils.NewScope("test"))

	cfg := defaultConfig
	cfg.WebAPI.Caching.Workers = 1
	cfg.WebAPI.Caching.ResyncInterval.Duration = 5 * time.Second
	cfg.DefaultAgent = Deployment{Endpoint: "test-agent.flyte.svc.cluster.local:80"}
	cfg.AgentDeployments = map[string]*Deployment{"spark_agent": {Endpoint: "localhost:80"}}
	cfg.AgentForTaskTypes = map[string]string{"spark": "spark_agent", "bar": "bar_agent"}

	agent := &Agent{AgentDeployment: &Deployment{Endpoint: "localhost:80"}}
	agentRegistry := Registry{"spark": {defaultTaskTypeVersion: agent}}
	plugin := Plugin{
		metricScope: fakeSetupContext.MetricsScope(),
		cfg:         GetConfig(),
		registry:    agentRegistry,
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
		p := newMockAsyncAgentPlugin()
		assert.NotNil(t, p)
		assert.Equal(t, "agent-service", p.ID)
		assert.NotNil(t, p.PluginLoader)
	})

	t.Run("test getFinalAgent", func(t *testing.T) {
		spark := &admin.TaskCategory{Name: "spark", Version: defaultTaskTypeVersion}
		foo := &admin.TaskCategory{Name: "foo", Version: defaultTaskTypeVersion}
		bar := &admin.TaskCategory{Name: "bar", Version: defaultTaskTypeVersion}
		agentDeployment, _ := plugin.getFinalAgent(spark, &cfg)
		assert.Equal(t, agentDeployment.Endpoint, "localhost:80")
		agentDeployment, _ = plugin.getFinalAgent(foo, &cfg)
		assert.Equal(t, agentDeployment.Endpoint, cfg.DefaultAgent.Endpoint)
		agentDeployment, _ = plugin.getFinalAgent(bar, &cfg)
		assert.Equal(t, agentDeployment.Endpoint, cfg.DefaultAgent.Endpoint)
	})

	t.Run("test getFinalTimeout", func(t *testing.T) {
		timeout := getFinalTimeout("CreateTask", &Deployment{Endpoint: "localhost:8080", Timeouts: map[string]config.Duration{"CreateTask": {Duration: 1 * time.Millisecond}}})
		assert.Equal(t, 1*time.Millisecond, timeout.Duration)
		timeout = getFinalTimeout("GetTask", &Deployment{Endpoint: "localhost:8080", Timeouts: map[string]config.Duration{"GetTask": {Duration: 1 * time.Millisecond}}})
		assert.Equal(t, 1*time.Millisecond, timeout.Duration)
		timeout = getFinalTimeout("DeleteTask", &Deployment{Endpoint: "localhost:8080", DefaultTimeout: config.Duration{Duration: 10 * time.Second}})
		assert.Equal(t, 10*time.Second, timeout.Duration)
		timeout = getFinalTimeout("ExecuteTaskSync", &Deployment{Endpoint: "localhost:8080", Timeouts: map[string]config.Duration{"ExecuteTaskSync": {Duration: 1 * time.Millisecond}}})
		assert.Equal(t, 1*time.Millisecond, timeout.Duration)
	})

	t.Run("test getFinalContext", func(t *testing.T) {

		ctx, _ := getFinalContext(context.TODO(), "CreateTask", &Deployment{Endpoint: "localhost:8080", Timeouts: map[string]config.Duration{"CreateTask": {Duration: 1 * time.Millisecond}}})
		assert.NotEqual(t, context.TODO(), ctx)

		ctx, _ = getFinalContext(context.TODO(), "GetTask", &Deployment{Endpoint: "localhost:8080", Timeouts: map[string]config.Duration{"GetTask": {Duration: 1 * time.Millisecond}}})
		assert.NotEqual(t, context.TODO(), ctx)

		ctx, _ = getFinalContext(context.TODO(), "DeleteTask", &Deployment{})
		assert.Equal(t, context.TODO(), ctx)

		ctx, _ = getFinalContext(context.TODO(), "ExecuteTaskSync", &Deployment{Endpoint: "localhost:8080", Timeouts: map[string]config.Duration{"ExecuteTaskSync": {Duration: 10 * time.Second}}})
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
		simpleStruct := structpb.Struct{
			Fields: map[string]*structpb.Value{
				"foo": {Kind: &structpb.Value_StringValue{StringValue: "foo"}},
			},
		}
		taskContext := new(webapiPlugin.StatusContext)
		taskContext.On("Resource").Return(ResourceWrapper{
			State:      admin.State_RUNNING,
			Outputs:    nil,
			Message:    "Job is running",
			LogLinks:   []*flyteIdlCore.TaskLog{{Uri: "http://localhost:3000/log", Name: "Log Link"}},
			CustomInfo: &simpleStruct,
		})

		phase, err := plugin.Status(context.Background(), taskContext)
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.PhaseRunning, phase.Phase())
		assert.Equal(t, &simpleStruct, phase.Info().CustomInfo)
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

func getMockMetadataServiceClient() *agentMocks.AgentMetadataServiceClient {
	mockMetadataServiceClient := new(agentMocks.AgentMetadataServiceClient)
	mockRequest := &admin.ListAgentsRequest{}
	supportedTaskCategories := make([]*admin.TaskCategory, 3)
	supportedTaskCategories[0] = &admin.TaskCategory{Name: "task1", Version: defaultTaskTypeVersion}
	supportedTaskCategories[1] = &admin.TaskCategory{Name: "task2", Version: defaultTaskTypeVersion}
	supportedTaskCategories[2] = &admin.TaskCategory{Name: "task3", Version: defaultTaskTypeVersion}
	mockResponse := &admin.ListAgentsResponse{
		Agents: []*admin.Agent{
			{
				Name:                    "test-agent",
				SupportedTaskCategories: supportedTaskCategories,
			},
		},
	}

	mockMetadataServiceClient.On("ListAgents", mock.Anything, mockRequest).Return(mockResponse, nil)
	return mockMetadataServiceClient
}

func TestInitializeAgentRegistry(t *testing.T) {
	agentClients := make(map[string]service.AsyncAgentServiceClient)
	agentMetadataClients := make(map[string]service.AgentMetadataServiceClient)
	agentClients[defaultAgentEndpoint] = &agentMocks.AsyncAgentServiceClient{}
	agentMetadataClients[defaultAgentEndpoint] = getMockMetadataServiceClient()

	cs := &ClientSet{
		asyncAgentClients:    agentClients,
		agentMetadataClients: agentMetadataClients,
	}

	cfg := defaultConfig
	cfg.AgentDeployments = map[string]*Deployment{"custom_agent": {Endpoint: defaultAgentEndpoint}}
	cfg.AgentForTaskTypes = map[string]string{"task1": "agent-deployment-1", "task2": "agent-deployment-2"}
	err := SetConfig(&cfg)
	assert.NoError(t, err)

	agentRegistry := getAgentRegistry(context.Background(), cs)
	agentRegistryKeys := maps.Keys(agentRegistry)
	expectedKeys := []string{"task1", "task2", "task3", "task_type_1", "task_type_2"}

	for _, key := range expectedKeys {
		assert.Contains(t, agentRegistryKeys, key)
	}
}
