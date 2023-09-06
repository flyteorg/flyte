package agent

import (
	"testing"
	"time"

	"github.com/flyteorg/flytestdlib/config"

	"github.com/stretchr/testify/assert"
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
	cfg.Agents = map[string]*Agent{
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
