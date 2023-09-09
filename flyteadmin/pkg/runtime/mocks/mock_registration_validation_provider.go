package mocks

import "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"

type MockRegistrationValidationProvider struct {
	WorkflowNodeLimit    int
	MaxLabelEntries      int
	MaxAnnotationEntries int
	WorkflowSizeLimit    string
}

func (c *MockRegistrationValidationProvider) GetWorkflowNodeLimit() int {
	return c.WorkflowNodeLimit
}

func (c *MockRegistrationValidationProvider) GetMaxLabelEntries() int {
	return c.MaxLabelEntries
}

func (c *MockRegistrationValidationProvider) GetMaxAnnotationEntries() int {
	return c.MaxAnnotationEntries
}

func (c *MockRegistrationValidationProvider) GetWorkflowSizeLimit() string {
	return c.WorkflowSizeLimit
}

func NewMockRegistrationValidationProvider() interfaces.RegistrationValidationConfiguration {
	return &MockRegistrationValidationProvider{}
}
