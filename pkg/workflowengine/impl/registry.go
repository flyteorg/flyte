package impl

import (
	"context"
	"sync"

	interfaces2 "github.com/flyteorg/flyteadmin/pkg/workflowengine/interfaces"

	"github.com/flyteorg/flyteadmin/pkg/workflowengine/interfaces"
	"github.com/flyteorg/flytestdlib/logger"
)

// Implements interfaces.WorkflowExecutorRegistry.
type workflowExecutorRegistry struct {
	// m is a read/write lock used for fetching and updating the K8sWorkflowExecutors.
	m               sync.RWMutex
	executor        interfaces.WorkflowExecutor
	defaultExecutor interfaces.WorkflowExecutor
}

func (r *workflowExecutorRegistry) Register(executor interfaces.WorkflowExecutor) {
	r.m.Lock()
	defer r.m.Unlock()
	if r.executor == nil {
		logger.Debugf(context.TODO(), "setting flyte k8s workflow executor [%s]", executor.ID())
	} else {
		logger.Debugf(context.TODO(), "updating flyte k8s workflow executor [%s]", executor.ID())
	}
	r.executor = executor
}

func (r *workflowExecutorRegistry) RegisterDefault(executor interfaces.WorkflowExecutor) {
	r.m.Lock()
	defer r.m.Unlock()
	if r.defaultExecutor == nil {
		logger.Debugf(context.TODO(), "setting default flyte k8s workflow executor [%s]", executor.ID())
	} else {
		logger.Debugf(context.TODO(), "updating default flyte k8s workflow executor [%s]", executor.ID())
	}
	r.defaultExecutor = executor
}

func (r *workflowExecutorRegistry) GetExecutor() interfaces.WorkflowExecutor {
	r.m.RLock()
	defer r.m.RUnlock()
	if r.executor == nil {
		return r.defaultExecutor
	}
	return r.executor
}

func NewRegistry() interfaces2.WorkflowExecutorRegistry {
	return &workflowExecutorRegistry{}
}
