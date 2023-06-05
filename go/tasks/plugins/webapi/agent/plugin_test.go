package agent

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"

	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	pluginCoreMocks "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
)

func TestPlugin(t *testing.T) {
	fakeSetupContext := pluginCoreMocks.SetupContext{}
	fakeSetupContext.OnMetricsScope().Return(promutils.NewScope("test"))

	plugin := Plugin{
		metricScope: fakeSetupContext.MetricsScope(),
		cfg:         GetConfig(),
	}
	t.Run("get config", func(t *testing.T) {
		cfg := defaultConfig
		cfg.WebAPI.Caching.Workers = 1
		cfg.WebAPI.Caching.ResyncInterval.Duration = 5 * time.Second
		cfg.DefaultGrpcEndpoint = "test-agent.flyte.svc.cluster.local:80"
		cfg.EndpointForTaskTypes = map[string]string{"spark": "localhost:80"}
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
		assert.Equal(t, p.ID, "agent-service")
		assert.NotNil(t, p.PluginLoader)
	})

	t.Run("test getFinalEndpoint", func(t *testing.T) {
		endpoint := getFinalEndpoint("spark", "localhost:8080", map[string]string{"spark": "localhost:80"})
		assert.Equal(t, endpoint, "localhost:80")
		endpoint = getFinalEndpoint("spark", "localhost:8080", map[string]string{})
		assert.Equal(t, endpoint, "localhost:8080")
	})

	t.Run("test getClientFunc", func(t *testing.T) {
		client, err := getClientFunc(context.Background(), "localhost:80", map[string]*grpc.ClientConn{})
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})
}
