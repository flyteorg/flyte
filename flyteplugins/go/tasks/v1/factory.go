package v1

import (
	"context"
	"fmt"
	"github.com/lyft/flyteplugins/go/tasks/v1/config"
	"github.com/lyft/flytestdlib/logger"
	"sync"
	"time"

	"github.com/lyft/flyteplugins/go/tasks/v1/flytek8s"
	"github.com/lyft/flyteplugins/go/tasks/v1/types"
	"github.com/lyft/flyteplugins/go/tasks/v1/utils"
	"github.com/lyft/flytestdlib/sets"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

type taskFactory struct {
	registeredTasksForTypes map[types.TaskType]types.Executor
	registeredDefault       types.Executor
	registrationLock        sync.RWMutex
}

var taskFactorySingleton = &taskFactory{
	registeredTasksForTypes: make(map[types.TaskType]types.Executor),
}

func GetTaskExecutor(taskType types.TaskType) (types.Executor, error) {
	taskFactorySingleton.registrationLock.RLock()
	defer taskFactorySingleton.registrationLock.RUnlock()

	e, ok := taskFactorySingleton.registeredTasksForTypes[taskType]
	if !ok {
		exec := taskFactorySingleton.registeredDefault
		if exec != nil {
			logger.Debugf(context.TODO(), "All registered plugins [%v]", taskFactorySingleton.registeredTasksForTypes)
			logger.Debugf(context.TODO(), "Using default plugin for [%s]", taskType)
			return exec, nil
		}

		return nil, errors.Errorf("No Executor defined for TaskType [%v] and no default executor is configured.", taskType)
	}
	logger.Debugf(context.TODO(), "Using plugin [%s] for task type [%s]", e.GetID(), taskType)

	return e, nil
}

func ListAllTaskExecutors() []types.Executor {
	taskFactorySingleton.registrationLock.RLock()
	defer taskFactorySingleton.registrationLock.RUnlock()

	taskExecutors := sets.NewGeneric()
	if taskFactorySingleton.registeredDefault != nil {
		taskExecutors.Insert(taskFactorySingleton.registeredDefault)
	}

	for _, v := range taskFactorySingleton.registeredTasksForTypes {
		taskExecutors.Insert(v)
	}
	setList := taskExecutors.UnsortedList()
	tList := make([]types.Executor, 0, len(setList))
	for _, t := range setList {
		tList = append(tList, t.(types.Executor))
	}
	return tList
}

func GetAllTaskTypeExecutors() map[types.TaskType]types.Executor {
	taskFactorySingleton.registrationLock.RLock()
	defer taskFactorySingleton.registrationLock.RUnlock()
	m := make(map[types.TaskType]types.Executor, len(taskFactorySingleton.registeredTasksForTypes))
	for k, v := range taskFactorySingleton.registeredTasksForTypes {
		m[k] = v
	}

	return m
}

func isEnabled(enabledPlugins []string, pluginToCheck string) bool {
	return enabledPlugins != nil && len(enabledPlugins) >= 1 &&
		(enabledPlugins[0] == "*" || utils.Contains(enabledPlugins, pluginToCheck))
}

func RegisterAsDefault(executor types.Executor) error {
	enabledPlugins := config.GetConfig().EnabledPlugins
	if isEnabled(enabledPlugins, executor.GetID()) {
		taskFactorySingleton.registrationLock.Lock()
		defer taskFactorySingleton.registrationLock.Unlock()
		if existingDefault := taskFactorySingleton.registeredDefault; existingDefault != nil {
			return fmt.Errorf("a default Executor already exists. Existing Plugin Id [%v], Proposed Plugin [%v]",
				existingDefault.GetID(), executor.GetID())
		}

		taskFactorySingleton.registeredDefault = executor
	}

	return nil
}

func RegisterForTaskTypes(executor types.Executor, taskTypes ...types.TaskType) error {
	logger.InfofNoCtx("Request to register executor [%v] for types [%+v]", executor.GetID(), taskTypes)
	enabledPlugins := config.GetConfig().EnabledPlugins
	if isEnabled(enabledPlugins, executor.GetID()) {
		logger.InfofNoCtx("Executor [%v] is enabled, attempting to register it.", executor.GetID())
		taskFactorySingleton.registrationLock.Lock()
		defer taskFactorySingleton.registrationLock.Unlock()
		for _, t := range taskTypes {
			x, ok := taskFactorySingleton.registeredTasksForTypes[t]
			if ok {
				return fmt.Errorf("an Executor already exists for TaskType [%v]. Existing Plugin Id [%v], Proposed Plugin [%v]", t, x.GetID(), executor.GetID())
			}

			logger.InfofNoCtx("Registering type [%s] with executor [%s]", t, executor.GetID())
			taskFactorySingleton.registeredTasksForTypes[t] = executor
		}
	} else {
		logger.InfofNoCtx("Executor [%v] is not enabled, not registering it. Enabled Plugins: [%+v]", executor.GetID(), enabledPlugins)
	}

	return nil
}

func K8sRegisterAsDefault(id string, resourceToWatch runtime.Object, resyncPeriod time.Duration, handler flytek8s.K8sResourceHandler) error {
	exec := flytek8s.NewK8sTaskExecutorForResource(id, resourceToWatch, handler, resyncPeriod)
	return RegisterAsDefault(exec)
}

func K8sRegisterForTaskTypes(id string, resourceToWatch runtime.Object, resyncPeriod time.Duration, handler flytek8s.K8sResourceHandler, taskTypes ...types.TaskType) error {
	exec := flytek8s.NewK8sTaskExecutorForResource(id, resourceToWatch, handler, resyncPeriod)
	return RegisterForTaskTypes(exec, taskTypes...)
}

// Clears all registered plugins. This should only be called after all packages/plugins have been safely loaded. Otherwise,
// init-based registrations will sneak through.
func ClearRegistry(_ context.Context) error {
	taskFactorySingleton.registrationLock.Lock()
	defer taskFactorySingleton.registrationLock.Unlock()

	taskFactorySingleton.registeredTasksForTypes = map[types.TaskType]types.Executor{}
	taskFactorySingleton.registeredDefault = nil
	return nil
}
