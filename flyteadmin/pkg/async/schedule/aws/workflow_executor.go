// AWS-specific implementation of a schedule.WorkflowExecutor.
package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/async"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"

	"github.com/flyteorg/flytestdlib/contextutils"

	"github.com/golang/protobuf/ptypes/timestamp"

	"github.com/flyteorg/flyteadmin/pkg/common"

	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"
	"github.com/prometheus/client_golang/prometheus"

	scheduleInterfaces "github.com/flyteorg/flyteadmin/pkg/async/schedule/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/golang/protobuf/ptypes"

	"github.com/NYTimes/gizmo/pubsub"
	"github.com/NYTimes/gizmo/pubsub/aws"
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/manager/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flytestdlib/logger"
	"google.golang.org/grpc/codes"
)

type workflowExecutorMetrics struct {
	Scope                               promutils.Scope
	NoActiveLaunchPlanVersionsFound     prometheus.Counter
	GreaterThan1LaunchPlanVersionsFound prometheus.Counter
	FailedPayloadDeserialization        prometheus.Counter
	FailedMarkMessageAsDone             prometheus.Counter
	FailedResolveKickoffTimeArg         prometheus.Counter
	FailedKickoffExecution              prometheus.Counter
	ScheduledEventsProcessed            prometheus.Counter
	ScheduledExecutionSystemDelay       labeled.StopWatch
	MessageReceivedDelay                labeled.StopWatch
	ScheduledEventProcessingDelay       labeled.StopWatch
	CreateExecutionDuration             labeled.StopWatch
	ChannelClosedError                  prometheus.Counter
	StopSubscriberFailed                prometheus.Counter
}

type workflowExecutor struct {
	subscriber        pubsub.Subscriber
	launchPlanManager interfaces.LaunchPlanInterface
	executionManager  interfaces.ExecutionInterface
	metrics           workflowExecutorMetrics
}

const workflowIdentifierFmt = "%s_%s_%s"

var activeLaunchPlanFilter = fmt.Sprintf("eq(state,%d)", int32(admin.LaunchPlanState_ACTIVE))

var timeout = int64(5)
var doNotconsumeBase64 = false

// The kickoff time argument isn't required for scheduled workflows. However, if it exists we substitute the kick-off
// time value for the input argument.
func (e *workflowExecutor) resolveKickoffTimeArg(
	request ScheduledWorkflowExecutionRequest, launchPlan admin.LaunchPlan,
	executionRequest *admin.ExecutionCreateRequest) error {
	if request.KickoffTimeArg == "" || launchPlan.Closure.ExpectedInputs == nil {
		logger.Debugf(context.Background(), "No kickoff time to resolve for scheduled workflow execution: [%s/%s/%s]",
			executionRequest.Project, executionRequest.Domain, executionRequest.Name)
		return nil
	}
	for name := range launchPlan.Closure.ExpectedInputs.Parameters {
		if name == request.KickoffTimeArg {
			ts, err := ptypes.TimestampProto(request.KickoffTime)
			if err != nil {
				logger.Warningf(context.Background(),
					"failed to serialize kickoff time %+v to timestamp proto for scheduled workflow execution with "+
						"launchPlan [%+v]", request.KickoffTime, launchPlan.Id)
				return errors.NewFlyteAdminErrorf(
					codes.Internal, "could not serialize kickoff time %+v to timestamp proto", request.KickoffTime)
			}
			executionRequest.Inputs.Literals[name] = &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Primitive{
							Primitive: &core.Primitive{
								Value: &core.Primitive_Datetime{
									Datetime: ts,
								},
							},
						},
					},
				},
			}
			return nil
		}
	}
	logger.Warningf(context.Background(),
		"expected kickoff time arg with launch plan [%+v] but did not find any matching expected input to resolve",
		launchPlan.Id)
	return nil
}

func (e *workflowExecutor) getActiveLaunchPlanVersion(launchPlanIdentifier *admin.NamedEntityIdentifier) (admin.LaunchPlan, error) {
	launchPlans, err := e.launchPlanManager.ListLaunchPlans(context.Background(), admin.ResourceListRequest{
		Id:      launchPlanIdentifier,
		Filters: activeLaunchPlanFilter,
		Limit:   1,
	})
	if err != nil {
		logger.Warningf(context.Background(), "failed to find active launch plan with identifier [%+v]",
			launchPlanIdentifier)
		e.metrics.NoActiveLaunchPlanVersionsFound.Inc()
		return admin.LaunchPlan{}, err
	}
	if len(launchPlans.LaunchPlans) != 1 {
		e.metrics.GreaterThan1LaunchPlanVersionsFound.Inc()
		logger.Warningf(context.Background(), "failed to get exactly one active launch plan for identifier: %+v",
			launchPlanIdentifier)
		return admin.LaunchPlan{}, errors.NewFlyteAdminErrorf(codes.Internal,
			"failed to get exactly one active launch plan for identifier: %+v", launchPlanIdentifier)
	}
	return *launchPlans.LaunchPlans[0], nil
}

func generateExecutionName(launchPlan admin.LaunchPlan, kickoffTime time.Time) string {
	hashedIdentifier := hashIdentifier(core.Identifier{
		Project: launchPlan.Id.Project,
		Domain:  launchPlan.Id.Domain,
		Name:    launchPlan.Id.Name,
	})
	randomSeed := kickoffTime.UnixNano() + int64(hashedIdentifier)
	return common.GetExecutionName(randomSeed)
}

func (e *workflowExecutor) formulateExecutionCreateRequest(
	launchPlan admin.LaunchPlan, kickoffTime time.Time) admin.ExecutionCreateRequest {
	// Deterministically assign a name based on the schedule kickoff time/launch plan definition.
	name := generateExecutionName(launchPlan, kickoffTime)
	logger.Debugf(context.Background(), "generated name [%s] for scheduled execution with launch plan [%+v]",
		name, launchPlan.Id)
	kickoffTimeProto, err := ptypes.TimestampProto(kickoffTime)
	if err != nil {
		// We expected that kickoff times are valid (in order for a scheduled event to fire).
		// If, for whatever reason we fail to record the kickoff time in the metadata, that's fine,
		// we don't fail the execution simply use an empty value instead.
		kickoffTimeProto = &timestamp.Timestamp{}
		logger.Warningf(context.Background(), "failed to serialize kickoff time [%v] to proto with err: %v",
			kickoffTime, err)
	}
	executionRequest := admin.ExecutionCreateRequest{
		Project: launchPlan.Id.Project,
		Domain:  launchPlan.Id.Domain,
		Name:    name,
		Spec: &admin.ExecutionSpec{
			LaunchPlan: launchPlan.Id,
			Metadata: &admin.ExecutionMetadata{
				Mode:        admin.ExecutionMetadata_SCHEDULED,
				ScheduledAt: kickoffTimeProto,
			},
			// No dynamic notifications are configured either.
		},
		// No additional inputs beyond the to-be-filled-out kickoff time arg are specified.
		Inputs: &core.LiteralMap{
			Literals: map[string]*core.Literal{},
		},
	}
	return executionRequest
}

func (e *workflowExecutor) Run() {
	for {
		logger.Warningf(context.Background(), "Starting workflow executor")
		err := e.run()
		logger.Errorf(context.Background(), "error with workflow executor err: [%v] ", err)
		time.Sleep(async.RetryDelay)
	}
}

func (e *workflowExecutor) run() error {
	for message := range e.subscriber.Start() {
		scheduledWorkflowExecutionRequest, err := DeserializeScheduleWorkflowPayload(message.Message())
		ctx := context.Background()
		observedMessageTriggeredTime := time.Now()
		if err != nil {
			e.metrics.FailedPayloadDeserialization.Inc()
			logger.Error(context.Background(), err.Error())
			continue
		}

		logger.Debugf(context.Background(), "Processing scheduled workflow execution event: %+v",
			scheduledWorkflowExecutionRequest)
		launchPlan, err := e.getActiveLaunchPlanVersion(&scheduledWorkflowExecutionRequest.LaunchPlanIdentifier)
		if err != nil {
			// In the rare case that a scheduled event fires right before a user disables the currently active launch
			// plan version (and triggers deleting the schedule rule) there may be no active launch plans. This is fine,
			// remove the message and move on.
			logger.Infof(context.Background(),
				"failed to get an active launch plan for scheduled workflow with launch plan identifier %v "+
					"removing scheduled event message without triggering execution",
				scheduledWorkflowExecutionRequest.LaunchPlanIdentifier)
			err = message.Done()
			if err != nil {
				e.metrics.FailedMarkMessageAsDone.Inc()
				panic(fmt.Sprintf(
					"failed to delete successfully created scheduled workflow execution from the queue with err: %v", err))
			}
			continue
		}
		executionRequest := e.formulateExecutionCreateRequest(launchPlan, scheduledWorkflowExecutionRequest.KickoffTime)

		ctx = contextutils.WithWorkflowID(ctx, fmt.Sprintf(workflowIdentifierFmt, executionRequest.Project,
			executionRequest.Domain, executionRequest.Name))
		err = e.resolveKickoffTimeArg(scheduledWorkflowExecutionRequest, launchPlan, &executionRequest)
		if err != nil {
			e.metrics.FailedResolveKickoffTimeArg.Inc()
			logger.Error(context.Background(), err.Error())
			continue
		}
		e.metrics.ScheduledEventProcessingDelay.Observe(ctx, scheduledWorkflowExecutionRequest.KickoffTime, time.Now())
		var response *admin.ExecutionCreateResponse
		e.metrics.CreateExecutionDuration.Time(ctx, func() {
			response, err = e.executionManager.CreateExecution(
				context.Background(), executionRequest, scheduledWorkflowExecutionRequest.KickoffTime)
		})

		if err != nil {
			ec, ok := err.(errors.FlyteAdminError)
			if ok && ec.Code() != codes.AlreadyExists {
				e.metrics.FailedKickoffExecution.Inc()
				logger.Errorf(context.Background(), "failed to execute scheduled workflow [%s:%s:%s] with err: %v",
					executionRequest.Project, executionRequest.Domain, executionRequest.Name, err)
				continue
			}
		} else {
			logger.Debugf(context.Background(), "created scheduled workflow execution %+v with kickoff time %+v",
				response.Id, scheduledWorkflowExecutionRequest.KickoffTime)
		}
		executionLaunchTime := time.Now()

		// Delete successfully scheduled executions from the queue.
		err = message.Done()
		if err != nil {
			e.metrics.FailedMarkMessageAsDone.Inc()
			logger.Warningf(context.Background(), fmt.Sprintf(
				"failed to delete successfully created scheduled workflow execution from the queue with err: %v",
				err))
		}
		e.metrics.ScheduledEventsProcessed.Inc()
		e.metrics.ScheduledExecutionSystemDelay.Observe(ctx, scheduledWorkflowExecutionRequest.KickoffTime,
			executionLaunchTime)
		e.metrics.MessageReceivedDelay.Observe(ctx, scheduledWorkflowExecutionRequest.KickoffTime,
			observedMessageTriggeredTime)
	}
	err := e.subscriber.Err()
	if err != nil {
		logger.Errorf(context.TODO(), "Gizmo subscriber closed channel with err: [%+v]", err)
		e.metrics.ChannelClosedError.Inc()
	}
	return err
}

func (e *workflowExecutor) Stop() error {
	err := e.subscriber.Stop()
	if err != nil {
		logger.Warningf(context.Background(), "failed to stop workflow executor with err %v", err)
		e.metrics.StopSubscriberFailed.Inc()
		return errors.NewFlyteAdminErrorf(codes.Internal, "failed to stop workflow executor with err %v", err)
	}
	return nil
}

func newWorkflowExecutorMetrics(scope promutils.Scope) workflowExecutorMetrics {
	return workflowExecutorMetrics{
		Scope: scope,
		NoActiveLaunchPlanVersionsFound: scope.MustNewCounter("no_active_lps",
			"no active launch plans found at scheduled workflow execution time"),
		GreaterThan1LaunchPlanVersionsFound: scope.MustNewCounter("too_many_active_lps",
			"> 1 active launch plans found at scheduled workflow execution time"),
		FailedPayloadDeserialization: scope.MustNewCounter("payload_deserialization_failures",
			"count of failures for deserialization of SQS message payload"),
		FailedMarkMessageAsDone: scope.MustNewCounter("messaged_marked_done_failures",
			"count of failures marking a message as done"),
		FailedResolveKickoffTimeArg: scope.MustNewCounter("kickoff_time_resolution_failures",
			"count of failures resolving the kickoff time argument"),
		FailedKickoffExecution: scope.MustNewCounter("workflow_execution_kickoff_failures",
			"count of failures kicking-off workflow execution"),
		ScheduledEventsProcessed: scope.MustNewCounter("scheduled_events_processed",
			"total number of schedule events successfully processed"),
		ScheduledExecutionSystemDelay: labeled.NewStopWatch("schedule_execution_delay",
			"observed time between when an execution was scheduled to be launched and when it was actually launched",
			time.Second, scope, labeled.EmitUnlabeledMetric),
		MessageReceivedDelay: labeled.NewStopWatch("message_received_delay",
			"observed time between receiving a scheduled event and desired trigger time",
			time.Second, scope, labeled.EmitUnlabeledMetric),
		ScheduledEventProcessingDelay: labeled.NewStopWatch("scheduled_event_delay",
			"time spent processing a triggered event before firing a workflow execution",
			time.Second, scope, labeled.EmitUnlabeledMetric),
		CreateExecutionDuration: labeled.NewStopWatch("create_execution_duration",
			"time spent waiting on the call to CreateExecution to return",
			time.Second, scope, labeled.EmitUnlabeledMetric),
		ChannelClosedError:   scope.MustNewCounter("channel_closed_error", "count of channel closing errors"),
		StopSubscriberFailed: scope.MustNewCounter("stop_subscriber_failed", "failures stopping the event subscriber"),
	}
}

func NewWorkflowExecutor(
	config aws.SQSConfig, schedulerConfig runtimeInterfaces.SchedulerConfig, executionManager interfaces.ExecutionInterface,
	launchPlanManager interfaces.LaunchPlanInterface, scope promutils.Scope) scheduleInterfaces.WorkflowExecutor {

	config.TimeoutSeconds = &timeout
	// By default gizmo tries to base64 decode messages. Since we don't use the gizmo publisher interface to publish
	// messages these are not encoded in base64 by default. Disable this behavior.
	config.ConsumeBase64 = &doNotconsumeBase64

	maxReconnectAttempts := schedulerConfig.ReconnectAttempts
	reconnectDelay := time.Duration(schedulerConfig.ReconnectDelaySeconds) * time.Second
	var subscriber pubsub.Subscriber
	var err error
	err = async.Retry(maxReconnectAttempts, reconnectDelay, func() error {
		subscriber, err = aws.NewSubscriber(config)
		if err != nil {
			logger.Warnf(context.TODO(), "Failed to initialize new gizmo aws subscriber with config: [%+v] and err: %v", config, err)
		}
		return err
	})

	if err != nil {
		scope.MustNewCounter(
			"initialize_executor_failed", "failures initializing scheduled workflow executor").Inc()
		panic(fmt.Sprintf("Failed to initialize scheduled workflow executor SQS subscriber with err: %v", err))

	}
	metrics := newWorkflowExecutorMetrics(scope)
	return &workflowExecutor{
		subscriber:        subscriber,
		executionManager:  executionManager,
		launchPlanManager: launchPlanManager,
		metrics:           metrics,
	}
}
