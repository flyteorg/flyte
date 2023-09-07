package mocks

import "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"

type MockTaskResourceConfiguration struct {
	Defaults interfaces.TaskResourceSet
	Limits   interfaces.TaskResourceSet
}

func (c *MockTaskResourceConfiguration) GetDefaults() interfaces.TaskResourceSet {
	return c.Defaults
}
func (c *MockTaskResourceConfiguration) GetLimits() interfaces.TaskResourceSet {
	return c.Limits
}

func NewMockTaskResourceConfiguration(defaults, limits interfaces.TaskResourceSet) interfaces.TaskResourceConfiguration {
	return &MockTaskResourceConfiguration{
		Defaults: defaults,
		Limits:   limits,
	}
}
