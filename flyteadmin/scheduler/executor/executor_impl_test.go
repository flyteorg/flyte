package executor

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/scheduler/repositories/models"
	adminMocks "github.com/flyteorg/flyte/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

var (
	mockAdminClient *adminMocks.AdminServiceClient
)

func setupExecutor(scope string) Executor {
	mockAdminClient = new(adminMocks.AdminServiceClient)
	return New(promutils.NewScope(scope), mockAdminClient)
}

func TestExecutor(t *testing.T) {
	executor := setupExecutor("testExecutor1")
	active := true
	mockAdminClient.OnCreateExecutionMatch(context.Background(), mock.Anything).Return(&admin.ExecutionCreateResponse{}, nil)
	t.Run("kickoff_time_arg", func(t *testing.T) {
		schedule := models.SchedulableEntity{
			SchedulableEntityKey: models.SchedulableEntityKey{
				Project: "project",
				Domain:  "domain",
				Name:    "cron_schedule",
				Version: "v1",
			},
			CronExpression:      "*/1 * * * *",
			KickoffTimeInputArg: "kickoff_time",
			Active:              &active,
		}
		err := executor.Execute(context.Background(), time.Now(), schedule)
		assert.Nil(t, err)
	})
	t.Run("without kickoff_time_arg", func(t *testing.T) {
		schedule := models.SchedulableEntity{
			SchedulableEntityKey: models.SchedulableEntityKey{
				Project: "project",
				Domain:  "domain",
				Name:    "cron_schedule",
				Version: "v1",
			},
			CronExpression: "*/1 * * * *",
			Active:         &active,
		}
		err := executor.Execute(context.Background(), time.Now(), schedule)
		assert.Nil(t, err)
	})
}

func TestExecutorAlreadyExists(t *testing.T) {
	executor := setupExecutor("testExecutor2")
	active := true
	schedule := models.SchedulableEntity{
		SchedulableEntityKey: models.SchedulableEntityKey{
			Project: "project",
			Domain:  "domain",
			Name:    "cron_schedule",
			Version: "v1",
		},
		CronExpression:      "*/1 * * * *",
		KickoffTimeInputArg: "kickoff_time",
		Active:              &active,
	}
	mockAdminClient.OnCreateExecutionMatch(mock.Anything, mock.Anything).Return(nil,
		errors.NewFlyteAdminErrorf(codes.AlreadyExists, "Already exists"))
	err := executor.Execute(context.Background(), time.Now(), schedule)
	assert.Nil(t, err)
}

func TestExecutorInactiveSchedule(t *testing.T) {
	executor := setupExecutor("testExecutor3")
	active := false
	schedule := models.SchedulableEntity{
		SchedulableEntityKey: models.SchedulableEntityKey{
			Project: "project",
			Domain:  "domain",
			Name:    "cron_schedule",
			Version: "v1",
		},
		CronExpression:      "*/1 * * * *",
		KickoffTimeInputArg: "kickoff_time",
		Active:              &active,
	}
	mockAdminClient.OnCreateExecutionMatch(context.Background(), mock.Anything).Return(&admin.ExecutionCreateResponse{}, nil)
	err := executor.Execute(context.Background(), time.Now(), schedule)
	assert.Nil(t, err)
}

func TestIsInactiveProjectError(t *testing.T) {
	statusErr := status.New(codes.InvalidArgument, "foo")
	var transformationErr error
	statusErr, transformationErr = statusErr.WithDetails(&admin.InactiveProject{
		Id: "project",
	})
	assert.NoError(t, transformationErr)

	assert.True(t, isInactiveProjectError(statusErr.Err()))
}
