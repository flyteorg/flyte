package executor

import (
	"context"
	"strings"
	"time"

	"github.com/flyteorg/flyteadmin/scheduler/identifier"
	"github.com/flyteorg/flyteadmin/scheduler/repositories/models"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
)

// executor allows to call the admin with scheduled execution
type executor struct {
	adminServiceClient service.AdminServiceClient
	metrics            executorMetrics
}

type executorMetrics struct {
	Scope                      promutils.Scope
	FailedExecutionCounter     prometheus.Counter
	SuccessfulExecutionCounter prometheus.Counter
}

func (w *executor) Execute(ctx context.Context, scheduledTime time.Time, s models.SchedulableEntity) error {

	literalsInputMap := map[string]*core.Literal{}
	// Only add kickoff time input arg for cron based schedules
	if len(s.CronExpression) > 0 && len(s.KickoffTimeInputArg) > 0 {
		literalsInputMap[s.KickoffTimeInputArg] = &core.Literal{
			Value: &core.Literal_Scalar{
				Scalar: &core.Scalar{
					Value: &core.Scalar_Primitive{
						Primitive: &core.Primitive{
							Value: &core.Primitive_Datetime{
								Datetime: timestamppb.New(scheduledTime),
							},
						},
					},
				},
			},
		}
	}

	// Making the identifier deterministic using the hash of the identifier and scheduled time
	executionIdentifier, err := identifier.GetExecutionIdentifier(ctx, core.Identifier{
		Project: s.Project,
		Domain:  s.Domain,
		Name:    s.Name,
		Version: s.Version,
	}, scheduledTime)

	if err != nil {
		logger.Error(ctx, "failed to generate execution identifier for schedule %+v due to %v", s, err)
		return err
	}

	executionRequest := &admin.ExecutionCreateRequest{
		Project: s.Project,
		Domain:  s.Domain,
		Name:    "f" + strings.ReplaceAll(executionIdentifier.String(), "-", "")[:19],
		Spec: &admin.ExecutionSpec{
			LaunchPlan: &core.Identifier{
				ResourceType: core.ResourceType_LAUNCH_PLAN,
				Project:      s.Project,
				Domain:       s.Domain,
				Name:         s.Name,
				Version:      s.Version,
			},
			Metadata: &admin.ExecutionMetadata{
				Mode:        admin.ExecutionMetadata_SCHEDULED,
				ScheduledAt: timestamppb.New(scheduledTime),
			},
			// No dynamic notifications are configured either.
		},
		// No additional inputs beyond the to-be-filled-out kickoff time arg are specified.
		Inputs: &core.LiteralMap{
			Literals: literalsInputMap,
		},
	}
	if !*s.Active {
		// no longer active
		logger.Debugf(ctx, "schedule %+v is no longer active", s)
		return nil
	}

	// Do maximum of 30 retries on failures with constant backoff factor
	opts := wait.Backoff{Duration: 3000, Factor: 2.0, Steps: 30}
	err = retry.OnError(opts,
		func(err error) bool {
			// For idempotent behavior ignore the AlreadyExists error which happens if we try to schedule a launchplan
			// for execution at the same time which is already available in admin.
			// This is possible since idempotency guarantees are using the schedule time and the identifier
			if grpcError := status.Code(err); grpcError == codes.AlreadyExists {
				logger.Debugf(ctx, "duplicate schedule %+v already exists for schedule", s)
				return false
			}
			w.metrics.FailedExecutionCounter.Inc()
			logger.Error(ctx, "failed to create execution create request %+v due to %v", executionRequest, err)
			// TODO: Handle the case when admin launch plan state is archived but the schedule is active.
			// After this bug is fixed in admin https://github.com/flyteorg/flyte/issues/1354
			return true
		},
		func() error {
			_, execErr := w.adminServiceClient.CreateExecution(context.Background(), executionRequest)
			return execErr
		},
	)
	if err != nil && status.Code(err) != codes.AlreadyExists {
		logger.Error(ctx, "failed to create execution create request %+v due to %v after all retries", executionRequest, err)
		return err
	}
	w.metrics.SuccessfulExecutionCounter.Inc()
	logger.Infof(ctx, "successfully fired the request for schedule %+v for time %v", s, scheduledTime)
	return nil
}

func New(scope promutils.Scope,
	adminServiceClient service.AdminServiceClient) Executor {

	return &executor{
		adminServiceClient: adminServiceClient,
		metrics:            getExecutorMetrics(scope),
	}
}

func getExecutorMetrics(scope promutils.Scope) executorMetrics {
	return executorMetrics{
		Scope: scope,
		FailedExecutionCounter: scope.MustNewCounter("failed_execution_counter",
			"count of unsuccessful attempts to fire execution for a schedules"),
		SuccessfulExecutionCounter: scope.MustNewCounter("successful_execution_counter",
			"count of successful attempts to fire execution for a schedules"),
	}
}
