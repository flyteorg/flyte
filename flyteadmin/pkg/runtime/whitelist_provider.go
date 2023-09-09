package runtime

import (
	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/config"
)

const whitelistKey = "task_type_whitelist"

var whiteListProviderDefault = make(map[string][]interfaces.WhitelistScope)

var whitelistConfig = config.MustRegisterSection(whitelistKey, &whiteListProviderDefault)

// Implementation of an interfaces.QueueConfiguration
type WhitelistConfigurationProvider struct{}

func (p *WhitelistConfigurationProvider) GetTaskTypeWhitelist() interfaces.TaskTypeWhitelist {
	whitelists := whitelistConfig.GetConfig().(*interfaces.TaskTypeWhitelist)
	return *whitelists
}

func NewWhitelistConfigurationProvider() interfaces.WhitelistConfiguration {
	return &WhitelistConfigurationProvider{}
}
