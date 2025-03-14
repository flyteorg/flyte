package connector

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flytestdlib/config"
)

func TestGetAndSetConfig(t *testing.T) {
	cfg := defaultConfig
	cfg.WebAPI.Caching.Workers = 1
	cfg.WebAPI.Caching.ResyncInterval.Duration = 5 * time.Second
	cfg.DefaultConnector.Insecure = false
	cfg.DefaultConnector.DefaultServiceConfig = "{\"loadBalancingConfig\": [{\"round_robin\":{}}]}"
	cfg.DefaultConnector.Timeouts = map[string]config.Duration{
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
	cfg.DefaultConnector.DefaultTimeout = config.Duration{Duration: 10 * time.Second}
	cfg.ConnectorDeployments = map[string]*Deployment{
		"connector_1": {
			Insecure:             cfg.DefaultConnector.Insecure,
			DefaultServiceConfig: cfg.DefaultConnector.DefaultServiceConfig,
			Timeouts:             cfg.DefaultConnector.Timeouts,
		},
	}
	cfg.ConnectorForTaskTypes = map[string]string{"task_type_1": "agent_1"}
	err := SetConfig(&cfg)
	assert.NoError(t, err)
	assert.Equal(t, &cfg, GetConfig())
}

func TestDefaultAgentConfig(t *testing.T) {
	cfg := defaultConfig

	assert.Equal(t, "", cfg.DefaultConnector.Endpoint)
	assert.True(t, cfg.DefaultConnector.Insecure)
	assert.Equal(t, 10*time.Second, cfg.DefaultConnector.DefaultTimeout.Duration)
	assert.Equal(t, `{"loadBalancingConfig": [{"round_robin":{}}]}`, cfg.DefaultConnector.DefaultServiceConfig)

	assert.Empty(t, cfg.DefaultConnector.Timeouts)

	expectedTaskTypes := []string{"task_type_1", "task_type_2"}
	assert.Equal(t, expectedTaskTypes, cfg.SupportedTaskTypes)

	assert.Empty(t, cfg.ConnectorDeployments)

	assert.Empty(t, cfg.ConnectorDeployments)

	assert.Equal(t, 10*time.Second, cfg.PollInterval.Duration)
}
