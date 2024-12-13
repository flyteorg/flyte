package workflowstore

import (
	"context"
	"fmt"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/labels"

	lister "github.com/flyteorg/flyte/flytepropeller/pkg/client/listers/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

const (
	resourceLevelExecutionStatsSyncDuration = 10 * time.Second
)

type ExecutionStatsMonitor struct {
	Scope promutils.Scope

	// Cached workflow list from the Informer
	lister lister.FlyteWorkflowLister

	// Provides stats for the currently active executions
	activeExecutions *ExecutionStatsHolder

	// These are currently aggregated values across all active workflows
	ActiveNodeExecutions     prometheus.Gauge
	ActiveTaskExecutions     prometheus.Gauge
	ActiveWorkflowExecutions prometheus.Gauge
}

func (e *ExecutionStatsMonitor) updateExecutionStats(ctx context.Context) {
	// Convert the list of workflows to a map for faster lookups.
	// Note that lister will exclude any workflows which are already marked terminated as the
	// controller creates the informer with a label selector that excludes them - IgnoreCompletedWorkflowsLabelSelector()
	workflows, err := e.lister.List(labels.Everything())
	if err != nil {
		logger.Errorf(ctx, "Failed to list workflows while removing terminated executions, %v", err)
		return
	}
	workflowSet := make(map[string]bool)
	for _, wf := range workflows {
		workflowSet[wf.GetExecutionID().String()] = true
	}

	err = e.activeExecutions.RemoveTerminatedExecutions(ctx, workflowSet)
	if err != nil {
		logger.Errorf(ctx, "Error while removing terminated executions from stats: %v ", err)
	}
}

func (e *ExecutionStatsMonitor) emitExecutionStats(ctx context.Context) {
	executions, nodes, tasks, err := e.activeExecutions.AggregateActiveValues()
	if err != nil {
		logger.Errorf(ctx, "Error aggregating active execution stats: %v", err)
	}
	logger.Debugf(ctx, "Execution stats: ActiveExecutions: %d ActiveNodes: %d, ActiveTasks: %d", executions, nodes, tasks)
	e.ActiveNodeExecutions.Set(float64(nodes))
	e.ActiveTaskExecutions.Set(float64(tasks))
	e.ActiveWorkflowExecutions.Set(float64(executions))
}

func (e *ExecutionStatsMonitor) RunStatsMonitor(ctx context.Context) {
	ticker := time.NewTicker(resourceLevelExecutionStatsSyncDuration)
	execStatsCtx := contextutils.WithGoroutineLabel(ctx, "execution-stats-monitor")

	go func() {
		pprof.SetGoroutineLabels(execStatsCtx)
		for {
			select {
			case <-execStatsCtx.Done():
				return
			case <-ticker.C:
				e.updateExecutionStats(ctx)
				e.emitExecutionStats(ctx)
			}
		}
	}()
}

func NewExecutionStatsMonitor(scope promutils.Scope, lister lister.FlyteWorkflowLister, activeExecutions *ExecutionStatsHolder) *ExecutionStatsMonitor {
	return &ExecutionStatsMonitor{
		Scope:                    scope,
		lister:                   lister,
		activeExecutions:         activeExecutions,
		ActiveNodeExecutions:     scope.MustNewGauge("active_node_executions", "active node executions for propeller"),
		ActiveTaskExecutions:     scope.MustNewGauge("active_task_executions", "active task executions for propeller"),
		ActiveWorkflowExecutions: scope.MustNewGauge("active_workflow_executions", "active workflow executions for propeller"),
	}
}

// SingleExecutionStats holds stats about a single workflow execution, such as active node and task counts.
type SingleExecutionStats struct {
	ActiveNodeCount uint32
	ActiveTaskCount uint32
}

// ExecutionStatsHolder manages a map of execution IDs to their ExecutionStats.
type ExecutionStatsHolder struct {
	mu         sync.Mutex // Guards access to the map
	executions map[string]SingleExecutionStats
}

// NewExecutionStatsHolder creates a new ExecutionStatsHolder instance with an initialized map.
func NewExecutionStatsHolder() (*ExecutionStatsHolder, error) {
	return &ExecutionStatsHolder{
		executions: make(map[string]SingleExecutionStats),
	}, nil
}

// AddOrUpdateEntry adds or updates an entry in the executions map.
func (esh *ExecutionStatsHolder) AddOrUpdateEntry(executionID string, executionStats SingleExecutionStats) error {
	if esh == nil || esh.executions == nil {
		return fmt.Errorf("ExecutionStatsHolder is not initialized")
	}

	esh.mu.Lock()
	defer esh.mu.Unlock()
	esh.executions[executionID] = executionStats

	return nil
}

// Returns the aggregate of all active node and task counts in the map.
func (esh *ExecutionStatsHolder) AggregateActiveValues() (int, uint32, uint32, error) {
	if esh == nil || esh.executions == nil {
		return 0, 0, 0, fmt.Errorf("ActiveExecutions is not initialized")
	}

	esh.mu.Lock()
	defer esh.mu.Unlock()

	var sumNodes, sumTasks uint32 = 0, 0
	for _, stats := range esh.executions {
		sumNodes += stats.ActiveNodeCount
		sumTasks += stats.ActiveTaskCount
	}
	return len(esh.executions), sumNodes, sumTasks, nil
}

func (esh *ExecutionStatsHolder) LogAllActiveExecutions(ctx context.Context) {
	esh.mu.Lock()
	defer esh.mu.Unlock()

	logger.Debugf(ctx, "Current Active Executions:")
	for execID, stats := range esh.executions {
		logger.Debugf(ctx, "ExecutionID: %s, ActiveNodeCount: %d, ActiveTaskCount: %d\n", execID, stats.ActiveNodeCount, stats.ActiveTaskCount)
	}
}

// RemoveTerminatedExecutions removes all terminated or deleted workflows from the executions map.
// This expects a set of strings for simplified lookup in the critical section.
func (esh *ExecutionStatsHolder) RemoveTerminatedExecutions(ctx context.Context, workflows map[string]bool) error {
	// Acquire the mutex and remove all terminated or deleted workflows.
	esh.mu.Lock()
	defer esh.mu.Unlock()
	for execID := range esh.executions {
		if !workflows[execID] {
			logger.Debugf(ctx, "Deleting active execution entry for execId: %s", execID)
			delete(esh.executions, execID)
		}
	}
	return nil
}
