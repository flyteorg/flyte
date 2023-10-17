package validation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"

	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	repositoryMocks "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/mocks"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

var testCluster = "C1"

var testExecID = &core.WorkflowExecutionIdentifier{
	Project: "p",
	Domain:  "d",
	Name:    "n",
}

func TestValidateClusterForExecutionID(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
		return models.Execution{
			Cluster: testCluster,
		}, nil
	})
	assert.NoError(t, ValidateClusterForExecutionID(context.TODO(), repository, testExecID, testCluster))
	assert.NoError(t, ValidateClusterForExecutionID(context.TODO(), repository, testExecID, common.DefaultProducerID))
}

func TestValidateCluster_Nonmatching(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
		return models.Execution{
			Cluster: "C2",
		}, nil
	})
	err := ValidateClusterForExecutionID(context.TODO(), repository, testExecID, testCluster)
	assert.Equal(t, codes.FailedPrecondition, err.(errors.FlyteAdminError).Code())
}

func TestValidateCluster_NoExecution(t *testing.T) {
	repository := repositoryMocks.NewMockRepository()
	expectedErr := errors.NewFlyteAdminError(codes.Internal, "foo")
	repository.ExecutionRepo().(*repositoryMocks.MockExecutionRepo).SetGetCallback(func(ctx context.Context, input interfaces.Identifier) (models.Execution, error) {
		return models.Execution{}, expectedErr
	})
	err := ValidateClusterForExecutionID(context.TODO(), repository, testExecID, testCluster)
	assert.Equal(t, expectedErr, err)
}
