package runtime

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/logger"
)

const whitelistKey = "task_type_whitelist"

var whitelistConfig = config.MustRegisterSection(whitelistKey, &interfaces.TaskTypeWhitelist{})

// Implementation of an interfaces.QueueConfiguration
type WhitelistConfigurationProvider struct{}

func (p *WhitelistConfigurationProvider) GetTaskTypeWhitelist() interfaces.TaskTypeWhitelist {
	if whitelistConfig != nil && whitelistConfig.GetConfig() != nil {
		whitelists := whitelistConfig.GetConfig().(*interfaces.TaskTypeWhitelist)
		return *whitelists
	}
	logger.Warningf(context.Background(), "Failed to find task type whitelist in config. Returning an empty slice")
	return interfaces.TaskTypeWhitelist{}
}

func NewWhitelistConfigurationProvider() interfaces.WhitelistConfiguration {
	return &WhitelistConfigurationProvider{}
}
