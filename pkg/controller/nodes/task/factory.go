package task

import (
	"time"

	v1 "github.com/lyft/flyteplugins/go/tasks/v1"
	"github.com/lyft/flyteplugins/go/tasks/v1/types"
	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/pkg/errors"
)

var testModeEnabled = false
var testTaskFactory Factory

type Factory interface {
	GetTaskExecutor(taskType v1alpha1.TaskType) (types.Executor, error)
	ListAllTaskExecutors() []types.Executor
}

// We create a simple facade so that if required we could make a Readonly cache of the Factory without any mutexes
// TODO decide if we want to make this a cache
type sealedTaskFactory struct {
}

func (sealedTaskFactory) GetTaskExecutor(taskType v1alpha1.TaskType) (types.Executor, error) {
	return v1.GetTaskExecutor(taskType)
}

func (sealedTaskFactory) ListAllTaskExecutors() []types.Executor {
	return v1.ListAllTaskExecutors()
}

func NewFactory(revalPeriod time.Duration) Factory {
	if testModeEnabled {
		return testTaskFactory
	}

	return sealedTaskFactory{}
}

func SetTestFactory(tf Factory) {
	testModeEnabled = true
	testTaskFactory = tf
}

func IsTestModeEnabled() bool {
	return testModeEnabled
}

func DisableTestMode() {
	testTaskFactory = nil
	testModeEnabled = false
}

type FactoryFuncs struct {
	GetTaskExecutorCb      func(taskType v1alpha1.TaskType) (types.Executor, error)
	ListAllTaskExecutorsCb func() []types.Executor
}

func (t *FactoryFuncs) GetTaskExecutor(taskType v1alpha1.TaskType) (types.Executor, error) {
	if t.GetTaskExecutorCb != nil {
		return t.GetTaskExecutorCb(taskType)
	}
	return nil, errors.Errorf("No implementation provided")
}

func (t *FactoryFuncs) ListAllTaskExecutors() []types.Executor {
	if t.ListAllTaskExecutorsCb != nil {
		return t.ListAllTaskExecutorsCb()
	}
	return nil
}
