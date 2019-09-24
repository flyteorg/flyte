package mocks

import "github.com/lyft/flyteadmin/pkg/runtime/interfaces"

type MockWhitelistConfiguration struct {
	TaskTypeWhitelist interfaces.TaskTypeWhitelist
}

func (c *MockWhitelistConfiguration) GetTaskTypeWhitelist() interfaces.TaskTypeWhitelist {
	return c.TaskTypeWhitelist
}

func NewMockWhitelistConfiguration() interfaces.WhitelistConfiguration {
	return &MockWhitelistConfiguration{}
}
