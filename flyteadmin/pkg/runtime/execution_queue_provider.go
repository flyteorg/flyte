package runtime

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flytestdlib/config"
)

const queuesKey = "queues"

var executionQueuesConfig = config.MustRegisterSection(queuesKey, &interfaces.QueueConfig{})

// Implementation of an interfaces.QueueConfiguration
type QueueConfigurationProvider struct{}

func (p *QueueConfigurationProvider) GetExecutionQueues() []interfaces.ExecutionQueue {
	if executionQueuesConfig != nil {
		return executionQueuesConfig.GetConfig().(*interfaces.QueueConfig).ExecutionQueues
	}
	logger.Warningf(context.Background(), "Failed to find execution queues in config. Returning an empty slice")
	return make([]interfaces.ExecutionQueue, 0)
}

func (p *QueueConfigurationProvider) GetWorkflowConfigs() []interfaces.WorkflowConfig {
	if executionQueuesConfig != nil {
		return executionQueuesConfig.GetConfig().(*interfaces.QueueConfig).WorkflowConfigs
	}
	logger.Warningf(context.Background(), "Failed to find workflows with attributes in config. Returning an empty slice")
	return make([]interfaces.WorkflowConfig, 0)
}

func NewQueueConfigurationProvider() interfaces.QueueConfiguration {
	return &QueueConfigurationProvider{}
}
