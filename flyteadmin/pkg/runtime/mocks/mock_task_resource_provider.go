package mocks

import "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"

type MockTaskResourceConfiguration struct {
	Defaults                        interfaces.TaskResourceSet
	Limits                          interfaces.TaskResourceSet
	AllowCPULimitToFloatFromRequest bool
}

func (c *MockTaskResourceConfiguration) GetDefaults() interfaces.TaskResourceSet {
	return c.Defaults
}
func (c *MockTaskResourceConfiguration) GetLimits() interfaces.TaskResourceSet {
	return c.Limits
}
func (c *MockTaskResourceConfiguration) GetAllowCPULimitToFloatFromRequest() bool {
	return c.AllowCPULimitToFloatFromRequest
}

func NewMockTaskResourceConfiguration(defaults, limits interfaces.TaskResourceSet, allowCPULimitToFloat bool) interfaces.TaskResourceConfiguration {
	return &MockTaskResourceConfiguration{
		Defaults:                        defaults,
		Limits:                          limits,
		AllowCPULimitToFloatFromRequest: allowCPULimitToFloat,
	}
}
