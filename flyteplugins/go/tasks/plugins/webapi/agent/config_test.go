package agent

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
	cfg.AgentApps = map[string]*Deployment{
		"connector_app": {
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

	assert.Empty(t, cfg.AgentDeployments)

	assert.Empty(t, cfg.AgentForTaskTypes)

	assert.Equal(t, 10*time.Second, cfg.PollInterval.Duration)
}
