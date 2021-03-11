package runtime

import (
	"context"

	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/logger"
)

const taskResourceKey = "task_resources"

var taskResourceConfig = config.MustRegisterSection(taskResourceKey, &TaskResourceSpec{})

type TaskResourceSpec struct {
	Defaults interfaces.TaskResourceSet `json:"defaults"`
	Limits   interfaces.TaskResourceSet `json:"limits"`
}

// Implementation of an interfaces.TaskResourceConfiguration
type TaskResourceProvider struct{}

func (p *TaskResourceProvider) GetDefaults() interfaces.TaskResourceSet {
	if taskResourceConfig != nil {
		return taskResourceConfig.GetConfig().(*TaskResourceSpec).Defaults
	}
	logger.Warning(context.Background(), "failed to find task resource values in config. Returning empty struct")
	return interfaces.TaskResourceSet{}
}

func (p *TaskResourceProvider) GetLimits() interfaces.TaskResourceSet {
	if taskResourceConfig != nil {
		return taskResourceConfig.GetConfig().(*TaskResourceSpec).Limits
	}
	logger.Warning(context.Background(), "failed to find task resource values in config. Returning empty struct")
	return interfaces.TaskResourceSet{}
}

func NewTaskResourceProvider() interfaces.TaskResourceConfiguration {
	return &TaskResourceProvider{}
}
