package runtime

import (
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flytestdlib/config"
)

const taskResourceKey = "task_resources"

var taskResourceConfig = config.MustRegisterSection(taskResourceKey, &TaskResourceSpec{
	Defaults: interfaces.TaskResourceSet{
		CPU:    resource.MustParse("2"),
		Memory: resource.MustParse("200Mi"),
	},
	Limits: interfaces.TaskResourceSet{
		CPU:    resource.MustParse("2"),
		Memory: resource.MustParse("1Gi"),
		GPU:    resource.MustParse("1"),
	},
})

type TaskResourceSpec struct {
	Defaults interfaces.TaskResourceSet `json:"defaults"`
	Limits   interfaces.TaskResourceSet `json:"limits"`
}

// Implementation of an interfaces.TaskResourceConfiguration
type TaskResourceProvider struct{}

func (p *TaskResourceProvider) GetDefaults() interfaces.TaskResourceSet {
	return taskResourceConfig.GetConfig().(*TaskResourceSpec).Defaults
}

func (p *TaskResourceProvider) GetLimits() interfaces.TaskResourceSet {
	return taskResourceConfig.GetConfig().(*TaskResourceSpec).Limits
}

func NewTaskResourceProvider() interfaces.TaskResourceConfiguration {
	return &TaskResourceProvider{}
}
