package validation

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/pkg/errors"
	repoInterfaces "github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
)

// ValidateClusterForExecutionID validates that the execution denoted by executionId is recorded as executing on `cluster`.
func ValidateClusterForExecutionID(ctx context.Context, db repoInterfaces.Repository, executionID *core.WorkflowExecutionIdentifier, clusterInEvent string) error {
	workflowExecution, err := db.ExecutionRepo().Get(ctx, repoInterfaces.Identifier{
		Project: executionID.Project,
		Domain:  executionID.Domain,
		Name:    executionID.Name,
	})
	if err != nil {
		logger.Debugf(ctx, "Failed to find existing execution with id [%+v] with err: %v", executionID, err)
		return err
	}
	return ValidateCluster(ctx, workflowExecution.Cluster, clusterInEvent)
}

// ValidateCluster validates that the execution is recorded as executing on `cluster`.
// clusterInEvent represents the cluster name, or historically, producerID sent in an execution event.
// clusterInDB represents the cluster recorded as running the execution in the database.
func ValidateCluster(ctx context.Context, clusterInDB, clusterInEvent string) error {
	// DefaultProducerID is used in older versions of propeller which hard code this producer id.
	// See https://github.com/flyteorg/flytepropeller/blob/eaf084934de5d630cd4c11aae15ecae780cc787e/pkg/controller/nodes/task/transformer.go#L114
	if len(clusterInEvent) == 0 || clusterInEvent == common.DefaultProducerID {
		return nil
	}
	if clusterInEvent != clusterInDB {
		errorMsg := fmt.Sprintf("Cluster/producer from event [%s] does not match existing workflow execution clusterInDB: [%s]",
			clusterInEvent, clusterInDB)
		return errors.NewIncompatibleClusterError(ctx, errorMsg, clusterInEvent)
	}
	return nil
}
