package runtime

import (
	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"

	"github.com/flyteorg/flytestdlib/config"
)

const queuesKey = "queues"

var executionQueuesConfig = config.MustRegisterSection(queuesKey, &interfaces.QueueConfig{
	ExecutionQueues: make([]interfaces.ExecutionQueue, 0),
	WorkflowConfigs: make([]interfaces.WorkflowConfig, 0),
})

// Implementation of an interfaces.QueueConfiguration
type QueueConfigurationProvider struct{}

func (p *QueueConfigurationProvider) GetExecutionQueues() []interfaces.ExecutionQueue {
	return executionQueuesConfig.GetConfig().(*interfaces.QueueConfig).ExecutionQueues
}

func (p *QueueConfigurationProvider) GetWorkflowConfigs() []interfaces.WorkflowConfig {
	return executionQueuesConfig.GetConfig().(*interfaces.QueueConfig).WorkflowConfigs
}

func NewQueueConfigurationProvider() interfaces.QueueConfiguration {
	return &QueueConfigurationProvider{}
}
