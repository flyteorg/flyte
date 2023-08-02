package agent

import (
	"context"
	"testing"
	"time"

	"github.com/flyteorg/flytestdlib/config"

	"google.golang.org/grpc"

	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	pluginCoreMocks "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
)

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

	t.Run("tet newAgentPlugin", func(t *testing.T) {
		p := newAgentPlugin()
		assert.NotNil(t, p)
		assert.Equal(t, "agent-service", p.ID)
		assert.NotNil(t, p.PluginLoader)
	})

	t.Run("test getFinalAgent", func(t *testing.T) {
		agent, _ := getFinalAgent("spark", &cfg)
		assert.Equal(t, cfg.Agents["spark_agent"].Endpoint, agent.Endpoint)
		agent, _ = getFinalAgent("foo", &cfg)
		assert.Equal(t, cfg.DefaultAgent.Endpoint, agent.Endpoint)
		_, err := getFinalAgent("bar", &cfg)
		assert.NotNil(t, err)
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
}
