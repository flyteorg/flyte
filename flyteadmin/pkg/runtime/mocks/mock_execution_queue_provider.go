package mocks

import "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"

type MockQueueConfigurationProvider struct {
	executionQueues []interfaces.ExecutionQueue
	workflowConfigs []interfaces.WorkflowConfig
}

func (p *MockQueueConfigurationProvider) GetExecutionQueues() []interfaces.ExecutionQueue {
	return p.executionQueues
}

func (p *MockQueueConfigurationProvider) GetWorkflowConfigs() []interfaces.WorkflowConfig {
	return p.workflowConfigs
}

func NewMockQueueConfigurationProvider(
	executionQueues []interfaces.ExecutionQueue,
	workflowConfigs []interfaces.WorkflowConfig) interfaces.QueueConfiguration {
	return &MockQueueConfigurationProvider{
		executionQueues: executionQueues,
		workflowConfigs: workflowConfigs,
	}
}
