package aws

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyteorg/flyteadmin/pkg/async/schedule/aws/interfaces"
	scheduleInterfaces "github.com/flyteorg/flyteadmin/pkg/async/schedule/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/errors"
	appInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc/codes"
)

// To indicate that a schedule rule is enabled.
var enableState = "ENABLED"

// CloudWatch schedule expressions.
const (
	cronExpression = "cron(%s)"
	rateExpression = "rate(%v %s)"
)

const timePlaceholder = "time"

var timeValue = "$.time"

const scheduleNameInputsFormat = "%s:%s:%s"
const scheduleDescriptionFormat = "Schedule for Project:%s Domain:%s Name:%s launch plan"
const scheduleNameFormat = "%s_%d"

// Container for initialized metrics objects
type cloudWatchSchedulerMetrics struct {
	Scope             promutils.Scope
	InvalidSchedules  prometheus.Counter
	AddRuleFailures   prometheus.Counter
	AddTargetFailures prometheus.Counter
	SchedulesAdded    prometheus.Counter

	RemoveRuleFailures      prometheus.Counter
	RemoveRuleDoesntExist   prometheus.Counter
	RemoveTargetFailures    prometheus.Counter
	RemoveTargetDoesntExist prometheus.Counter
	RemovedSchedules        prometheus.Counter

	ActiveSchedules prometheus.Gauge
}

// An AWS CloudWatch implementation of the EventScheduler.
type cloudWatchScheduler struct {
	// The ARN of the IAM role associated with the scheduler.
	scheduleRoleArn string
	// The ARN of the SQS target used for registering schedule events.
	targetSqsArn string
	// AWS CloudWatchEvents service client.
	cloudWatchEventClient interfaces.CloudWatchEventClient
	// For emitting scheduler-related metrics
	metrics cloudWatchSchedulerMetrics
}

func getScheduleName(scheduleNamePrefix string, identifier core.Identifier) string {
	hashedIdentifier := hashIdentifier(identifier)
	if len(scheduleNamePrefix) > 0 {
		return fmt.Sprintf(scheduleNameFormat, scheduleNamePrefix, hashedIdentifier)
	}
	return fmt.Sprintf("%d", hashedIdentifier)
}

func getScheduleDescription(identifier core.Identifier) string {
	return fmt.Sprintf(scheduleDescriptionFormat,
		identifier.Project, identifier.Domain, identifier.Name)
}

func getScheduleExpression(schedule admin.Schedule) (string, error) {
	if schedule.GetCronExpression() != "" {
		return fmt.Sprintf(cronExpression, schedule.GetCronExpression()), nil
	}
	if schedule.GetRate() != nil {
		// AWS uses pluralization for units of values not equal to 1.
		// See https://docs.aws.amazon.com/lambda/latest/dg/tutorial-scheduled-events-schedule-expressions.html
		unit := strings.ToLower(schedule.GetRate().Unit.String())
		if schedule.GetRate().Value != 1 {
			unit = fmt.Sprintf("%ss", unit)
		}
		return fmt.Sprintf(rateExpression, schedule.GetRate().Value, unit), nil
	}
	logger.Debugf(context.Background(), "scheduler encountered invalid schedule expression: %s", schedule.String())
	return "", errors.NewFlyteAdminErrorf(codes.InvalidArgument, "unrecognized schedule expression")
}

func formatEventScheduleInputs(inputTemplate *string) cloudwatchevents.InputTransformer {
	inputsPathMap := map[string]*string{
		timePlaceholder: &timeValue,
	}
	return cloudwatchevents.InputTransformer{
		InputPathsMap: inputsPathMap,
		InputTemplate: inputTemplate,
	}
}

func (s *cloudWatchScheduler) AddSchedule(ctx context.Context, input scheduleInterfaces.AddScheduleInput) error {
	if input.Payload == nil {
		logger.Debugf(ctx, "AddSchedule called with empty input payload: %+v", input)
		return errors.NewFlyteAdminError(codes.InvalidArgument, "payload serialization function cannot be nil")
	}
	scheduleExpression, err := getScheduleExpression(input.ScheduleExpression)
	if err != nil {
		s.metrics.InvalidSchedules.Inc()
		return err
	}
	scheduleName := getScheduleName(input.ScheduleNamePrefix, input.Identifier)
	scheduleDescription := getScheduleDescription(input.Identifier)
	// First define a rule which gets triggered on a schedule.
	requestInput := cloudwatchevents.PutRuleInput{
		ScheduleExpression: &scheduleExpression,
		Name:               &scheduleName,
		Description:        &scheduleDescription,
		RoleArn:            &s.scheduleRoleArn,
		State:              &enableState,
	}
	putRuleOutput, err := s.cloudWatchEventClient.PutRule(&requestInput)
	if err != nil {
		logger.Infof(ctx, "Failed to add rule to cloudwatch for schedule [%+v] with name %s and expression %s with err: %v",
			input.Identifier, scheduleName, scheduleExpression, err)
		s.metrics.AddRuleFailures.Inc()
		return errors.NewFlyteAdminErrorf(codes.Internal, "failed to add rule to cloudwatch with err: %v", err)
	}
	eventInputTransformer := formatEventScheduleInputs(input.Payload)
	// Next, add a target which gets invoked when the above rule is triggered.
	putTargetOutput, err := s.cloudWatchEventClient.PutTargets(&cloudwatchevents.PutTargetsInput{
		Rule: &scheduleName,
		Targets: []*cloudwatchevents.Target{
			{
				Arn:              &s.targetSqsArn,
				Id:               &scheduleName,
				InputTransformer: &eventInputTransformer,
			},
		},
	})
	if err != nil {
		logger.Infof(ctx, "Failed to add target for event schedule [%+v] with name %s with err: %v",
			input.Identifier, scheduleName, err)
		s.metrics.AddTargetFailures.Inc()
		return errors.NewFlyteAdminErrorf(codes.Internal, "failed to add target for event schedule with err: %v", err)
	} else if putTargetOutput.FailedEntryCount != nil && *putTargetOutput.FailedEntryCount > 0 {
		logger.Infof(ctx, "Failed to add target for event schedule [%+v] with name %s with failed entries: %d",
			input.Identifier, scheduleName, *putTargetOutput.FailedEntryCount)
		s.metrics.AddTargetFailures.Inc()
		return errors.NewFlyteAdminErrorf(codes.Internal,
			"failed to add target for event schedule with %v errs", *putTargetOutput.FailedEntryCount)
	}
	var putRuleOutputName string
	if putRuleOutput != nil && putRuleOutput.RuleArn != nil {
		putRuleOutputName = *putRuleOutput.RuleArn
	}
	logger.Debugf(ctx, "Added schedule %s [%s] with arn: %s (%s)",
		scheduleName, scheduleExpression, putRuleOutputName, scheduleDescription)
	s.metrics.SchedulesAdded.Inc()
	s.metrics.ActiveSchedules.Inc()
	return nil
}

func (s *cloudWatchScheduler) CreateScheduleInput(ctx context.Context, appConfig *appInterfaces.SchedulerConfig,
	identifier core.Identifier, schedule *admin.Schedule) (scheduleInterfaces.AddScheduleInput, error) {

	payload, err := SerializeScheduleWorkflowPayload(
		schedule.GetKickoffTimeInputArg(),
		admin.NamedEntityIdentifier{
			Project: identifier.Project,
			Domain:  identifier.Domain,
			Name:    identifier.Name,
		})
	if err != nil {
		logger.Errorf(ctx, "failed to serialize schedule workflow payload for launch plan: %v with err: %v",
			identifier, err)
		return scheduleInterfaces.AddScheduleInput{}, err
	}

	// Backward compatible with old EvenSchedulerConfig structure
	scheduleNamePrefix := appConfig.EventSchedulerConfig.GetScheduleNamePrefix()
	if appConfig.EventSchedulerConfig.GetAWSSchedulerConfig() != nil {
		scheduleNamePrefix = appConfig.EventSchedulerConfig.GetAWSSchedulerConfig().GetScheduleNamePrefix()
	}

	addScheduleInput := scheduleInterfaces.AddScheduleInput{
		Identifier:         identifier,
		ScheduleExpression: *schedule,
		Payload:            payload,
		ScheduleNamePrefix: scheduleNamePrefix,
	}
	return addScheduleInput, nil
}

func isResourceNotFoundException(err error) bool {
	switch err := err.(type) {
	case awserr.Error:
		return err.Code() == cloudwatchevents.ErrCodeResourceNotFoundException
	}
	return false
}

func (s *cloudWatchScheduler) RemoveSchedule(ctx context.Context, input scheduleInterfaces.RemoveScheduleInput) error {
	name := getScheduleName(input.ScheduleNamePrefix, input.Identifier)
	// All outbound targets for a rule must be deleted before the rule itself can be deleted.
	output, err := s.cloudWatchEventClient.RemoveTargets(&cloudwatchevents.RemoveTargetsInput{
		Ids: []*string{
			&name,
		},
		Rule: &name,
	})
	if err != nil {
		if isResourceNotFoundException(err) {
			s.metrics.RemoveTargetDoesntExist.Inc()
			logger.Debugf(ctx, "Tried to remove cloudwatch target %s but it was not found", name)
		} else {
			s.metrics.RemoveTargetFailures.Inc()
			logger.Errorf(ctx, "failed to remove cloudwatch target %s with err: %v", name, err)
			return errors.NewFlyteAdminErrorf(codes.Internal, "failed to remove cloudwatch target %s with err: %v", name, err)
		}
	}
	if output != nil && output.FailedEntryCount != nil && *output.FailedEntryCount > 0 {
		s.metrics.RemoveTargetFailures.Inc()
		logger.Errorf(ctx, "failed to remove cloudwatch target %s with %v errs",
			name, *output.FailedEntryCount)
		return errors.NewFlyteAdminErrorf(codes.Internal, "failed to remove cloudwatch target %s with %v errs",
			name, *output.FailedEntryCount)
	}

	// Output from the call to DeleteRule is an empty struct.
	_, err = s.cloudWatchEventClient.DeleteRule(&cloudwatchevents.DeleteRuleInput{
		Name: &name,
	})
	if err != nil {
		if isResourceNotFoundException(err) {
			s.metrics.RemoveRuleDoesntExist.Inc()
			logger.Debugf(ctx, "Tried to remove cloudwatch rule %s but it was not found", name)
		} else {
			s.metrics.RemoveRuleFailures.Inc()
			logger.Errorf(ctx, "failed to remove cloudwatch rule %s with err: %v", name, err)
			return errors.NewFlyteAdminErrorf(codes.Internal,
				"failed to remove cloudwatch rule %s with err: %v", name, err)
		}
	}
	s.metrics.RemovedSchedules.Inc()
	s.metrics.ActiveSchedules.Dec()
	logger.Debugf(ctx, "Removed schedule %s for identifier [%+v]", name, input.Identifier)
	return nil
}

// Initializes a new set of metrics specific to the cloudwatch scheduler implementation.
func newCloudWatchSchedulerMetrics(scope promutils.Scope) cloudWatchSchedulerMetrics {
	return cloudWatchSchedulerMetrics{
		Scope:            scope,
		InvalidSchedules: scope.MustNewCounter("schedules_invalid", "count of invalid schedule expressions submitted"),
		AddRuleFailures: scope.MustNewCounter("add_rule_failures",
			"count of attempts to add a cloudwatch rule that have failed"),
		AddTargetFailures: scope.MustNewCounter("add_target_failures",
			"count of attempts to add a cloudwatch target that have failed"),
		SchedulesAdded: scope.MustNewCounter("schedules_added",
			"count of all schedules successfully added to cloudwatch"),
		RemoveRuleFailures: scope.MustNewCounter("delete_rule_failures",
			"count of attempts to remove a cloudwatch rule that have failed"),
		RemoveRuleDoesntExist: scope.MustNewCounter("delete_rule_no_rule",
			"count of attempts to remove a cloudwatch rule that doesn't exist"),
		RemoveTargetFailures: scope.MustNewCounter("delete_target_failures",
			"count of attempts to remove a cloudwatch target that have failed"),
		RemoveTargetDoesntExist: scope.MustNewCounter("delete_target_no_target",
			"count of attempts to remove a cloudwatch target that doesn't exist"),
		RemovedSchedules: scope.MustNewCounter("schedules_removed",
			"count of all schedules successfully removed from cloudwatch"),
		ActiveSchedules: scope.MustNewGauge("active_schedules",
			"count of all active schedules currently in cloudwatch"),
	}
}

func NewCloudWatchScheduler(
	scheduleRoleArn, targetSqsArn string, session *session.Session, config *aws.Config,
	scope promutils.Scope) scheduleInterfaces.EventScheduler {
	cloudwatchEventClient := cloudwatchevents.New(session, config)
	metrics := newCloudWatchSchedulerMetrics(scope)
	return &cloudWatchScheduler{
		scheduleRoleArn:       scheduleRoleArn,
		targetSqsArn:          targetSqsArn,
		cloudWatchEventClient: cloudwatchEventClient,
		metrics:               metrics,
	}
}
