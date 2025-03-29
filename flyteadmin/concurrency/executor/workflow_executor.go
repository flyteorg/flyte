package executor

import (
	"context"
	"time"

	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/prometheus/client_golang/prometheus"
)

// WorkflowExecutor handles creating and managing workflow executions
type WorkflowExecutor struct {
	executionManager interfaces.ExecutionInterface
	scope            promutils.Scope
	metrics          *workflowExecutorMetrics
}

type workflowExecutorMetrics struct {
	successfulExecutions prometheus.Counter
	failedExecutions     prometheus.Counter
	abortedExecutions    prometheus.Counter
}

// CreateExecution creates a workflow execution in the system
func (w *WorkflowExecutor) CreateExecution(ctx context.Context, execution models.Execution) error {
	// Ensure we have the proper identifier
	launchPlanIdentifier := &core.Identifier{
		ResourceType: core.ResourceType_LAUNCH_PLAN,
		Project:      execution.Project,
		Domain:       execution.Domain,
		Name:         execution.Name,
	}

	// Create execution metadata
	metadata := &admin.ExecutionMetadata{
		Mode: admin.ExecutionMetadata_SCHEDULED, // Or whatever mode is appropriate
	}

	// Create the execution request
	executionRequest := &admin.ExecutionCreateRequest{
		Domain:  execution.Domain,
		Project: execution.Project,
		Name:    execution.Name,
		Spec: &admin.ExecutionSpec{
			LaunchPlan: launchPlanIdentifier,
			Metadata:   metadata,
		},
	}

	// Create the workflow execution
	_, err := w.executionManager.CreateExecution(ctx, executionRequest, time.Now())
	if err != nil {
		w.metrics.failedExecutions.Inc()
		logger.Errorf(ctx, "Failed to create execution: %v", err)
		return err
	}

	w.metrics.successfulExecutions.Inc()
	return nil
}

// AbortExecution aborts an existing workflow execution
func (w *WorkflowExecutor) AbortExecution(ctx context.Context, executionID core.WorkflowExecutionIdentifier, cause string) error {
	_, err := w.executionManager.TerminateExecution(ctx, &admin.ExecutionTerminateRequest{
		Id:    &executionID,
		Cause: cause,
	})

	if err != nil {
		logger.Errorf(ctx, "Failed to abort execution %v: %v", executionID, err)
		return err
	}

	w.metrics.abortedExecutions.Inc()
	return nil
}
