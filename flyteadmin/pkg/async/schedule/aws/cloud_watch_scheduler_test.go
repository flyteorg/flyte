package aws

import (
	"context"
	"fmt"
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/async/schedule/aws/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/async/schedule/aws/mocks"
	scheduleInterfaces "github.com/flyteorg/flyteadmin/pkg/async/schedule/interfaces"
	flyteAdminErrors "github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/promutils"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

const testScheduleName = "flyte_16301494360130577061"
const testScheduleDescription = "Schedule for Project:project Domain:domain Name:name launch plan"

var expectedError = flyteAdminErrors.NewFlyteAdminError(codes.Internal, "foo")

var testSerializedPayload = fmt.Sprintf("event triggered at '%s'", awsTimestampPlaceholder)

var testSchedulerIdentifier = core.Identifier{
	Project: "project",
	Domain:  "domain",
	Name:    "name",
	Version: "ignored",
}

var scope = promutils.NewScope("test_scheduler")

var testCloudWatchSchedulerMetrics = newCloudWatchSchedulerMetrics(scope)

const testScheduleNamePrefix = "flyte"

func TestGetScheduleName(t *testing.T) {
	scheduleName := getScheduleName(testScheduleNamePrefix, testSchedulerIdentifier)
	assert.Equal(t, "flyte_16301494360130577061", scheduleName)
}

func TestGetScheduleName_NoSystemPrefix(t *testing.T) {
	scheduleName := getScheduleName("", testSchedulerIdentifier)
	assert.Equal(t, "16301494360130577061", scheduleName)
}

func TestGetScheduleDescription(t *testing.T) {
	scheduleDescription := getScheduleDescription(testSchedulerIdentifier)
	assert.Equal(t, "Schedule for Project:project Domain:domain Name:name launch plan", scheduleDescription)
}

func TestGetScheduleExpression(t *testing.T) {
	expression, err := getScheduleExpression(admin.Schedule{
		ScheduleExpression: &admin.Schedule_CronExpression{
			CronExpression: "foo",
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "cron(foo)", expression)

	expression, err = getScheduleExpression(admin.Schedule{
		ScheduleExpression: &admin.Schedule_Rate{
			Rate: &admin.FixedRate{
				Value: 1,
				Unit:  admin.FixedRateUnit_DAY,
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "rate(1 day)", expression)

	expression, err = getScheduleExpression(admin.Schedule{
		ScheduleExpression: &admin.Schedule_Rate{
			Rate: &admin.FixedRate{
				Value: 2,
				Unit:  admin.FixedRateUnit_HOUR,
			},
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, "rate(2 hours)", expression)

	_, err = getScheduleExpression(admin.Schedule{})
	assert.Equal(t, codes.InvalidArgument, err.(flyteAdminErrors.FlyteAdminError).Code())
}

func TestFormatEventScheduleInputs(t *testing.T) {
	inputTransformer := formatEventScheduleInputs(&testSerializedPayload)
	assert.EqualValues(t, map[string]*string{
		"time": &timeValue,
	}, inputTransformer.InputPathsMap)
	assert.Equal(t, testSerializedPayload, *inputTransformer.InputTemplate)
}

func getCloudWatchSchedulerForTest(client interfaces.CloudWatchEventClient) scheduleInterfaces.EventScheduler {

	return &cloudWatchScheduler{
		scheduleRoleArn:       "ScheduleRole",
		targetSqsArn:          "TargetSqsArn",
		cloudWatchEventClient: client,
		metrics:               testCloudWatchSchedulerMetrics,
	}
}

func TestAddSchedule(t *testing.T) {
	mockCloudWatchEventClient := mocks.NewMockCloudWatchEventClient()
	mockCloudWatchEventClient.(*mocks.MockCloudWatchEventClient).SetPutRuleFunc(func(
		input *cloudwatchevents.PutRuleInput) (*cloudwatchevents.PutRuleOutput, error) {
		assert.Equal(t, "rate(1 minute)", *input.ScheduleExpression)
		assert.Equal(t, testScheduleName, *input.Name)
		assert.Equal(t, testScheduleDescription, *input.Description)
		assert.Equal(t, "ScheduleRole", *input.RoleArn)
		assert.Equal(t, enableState, *input.State)
		return &cloudwatchevents.PutRuleOutput{}, nil
	})

	mockCloudWatchEventClient.(*mocks.MockCloudWatchEventClient).SetPutTargetsFunc(func(
		input *cloudwatchevents.PutTargetsInput) (*cloudwatchevents.PutTargetsOutput, error) {
		assert.Equal(t, testScheduleName, *input.Rule)
		assert.Len(t, input.Targets, 1)
		assert.Equal(t, "TargetSqsArn", *input.Targets[0].Arn)
		assert.Equal(t, testScheduleName, *input.Targets[0].Id)
		assert.NotEmpty(t, *input.Targets[0].InputTransformer)
		return &cloudwatchevents.PutTargetsOutput{}, nil
	})

	scheduler := getCloudWatchSchedulerForTest(mockCloudWatchEventClient)
	assert.Nil(t, scheduler.AddSchedule(context.Background(),
		scheduleInterfaces.AddScheduleInput{
			Identifier: testSchedulerIdentifier,
			ScheduleExpression: admin.Schedule{
				ScheduleExpression: &admin.Schedule_Rate{
					Rate: &admin.FixedRate{
						Value: 1,
						Unit:  admin.FixedRateUnit_MINUTE,
					},
				},
			},
			Payload:            &testSerializedPayload,
			ScheduleNamePrefix: testScheduleNamePrefix,
		}))
}

func TestAddSchedule_InvalidScheduleExpression(t *testing.T) {
	mockCloudWatchEventClient := mocks.NewMockCloudWatchEventClient()
	scheduler := getCloudWatchSchedulerForTest(mockCloudWatchEventClient)
	err := scheduler.AddSchedule(context.Background(),
		scheduleInterfaces.AddScheduleInput{
			Identifier: testSchedulerIdentifier,
			Payload:    &testSerializedPayload,
		})
	assert.Equal(t, codes.InvalidArgument, err.(flyteAdminErrors.FlyteAdminError).Code())
}

func TestAddSchedule_PutRuleError(t *testing.T) {
	mockCloudWatchEventClient := mocks.NewMockCloudWatchEventClient()
	mockCloudWatchEventClient.(*mocks.MockCloudWatchEventClient).SetPutRuleFunc(func(
		input *cloudwatchevents.PutRuleInput) (*cloudwatchevents.PutRuleOutput, error) {
		return nil, expectedError
	})

	scheduler := getCloudWatchSchedulerForTest(mockCloudWatchEventClient)
	err := scheduler.AddSchedule(context.Background(),
		scheduleInterfaces.AddScheduleInput{
			Identifier: testSchedulerIdentifier,
			ScheduleExpression: admin.Schedule{
				ScheduleExpression: &admin.Schedule_Rate{
					Rate: &admin.FixedRate{
						Value: 1,
						Unit:  admin.FixedRateUnit_MINUTE,
					},
				},
			},
			Payload: &testSerializedPayload,
		})
	assert.Equal(t, codes.Internal, err.(flyteAdminErrors.FlyteAdminError).Code())
}

func TestAddSchedule_PutTargetsError(t *testing.T) {
	mockCloudWatchEventClient := mocks.NewMockCloudWatchEventClient()
	mockCloudWatchEventClient.(*mocks.MockCloudWatchEventClient).SetPutRuleFunc(func(
		input *cloudwatchevents.PutRuleInput) (*cloudwatchevents.PutRuleOutput, error) {
		return &cloudwatchevents.PutRuleOutput{}, nil
	})
	mockCloudWatchEventClient.(*mocks.MockCloudWatchEventClient).SetPutTargetsFunc(func(
		input *cloudwatchevents.PutTargetsInput) (*cloudwatchevents.PutTargetsOutput, error) {
		return nil, expectedError
	})
	scheduler := getCloudWatchSchedulerForTest(mockCloudWatchEventClient)
	err := scheduler.AddSchedule(context.Background(),
		scheduleInterfaces.AddScheduleInput{
			Identifier: testSchedulerIdentifier,
			ScheduleExpression: admin.Schedule{
				ScheduleExpression: &admin.Schedule_Rate{
					Rate: &admin.FixedRate{
						Value: 1,
						Unit:  admin.FixedRateUnit_MINUTE,
					},
				},
			},
			Payload: &testSerializedPayload,
		})
	assert.Equal(t, codes.Internal, err.(flyteAdminErrors.FlyteAdminError).Code())
}

func TestRemoveSchedule(t *testing.T) {
	mockCloudWatchEventClient := mocks.NewMockCloudWatchEventClient()
	mockCloudWatchEventClient.(*mocks.MockCloudWatchEventClient).SetRemoveTargetsFunc(func(
		input *cloudwatchevents.RemoveTargetsInput) (*cloudwatchevents.RemoveTargetsOutput, error) {
		assert.Len(t, input.Ids, 1)
		assert.Equal(t, testScheduleName, *input.Ids[0])
		assert.Equal(t, testScheduleName, *input.Rule)
		return &cloudwatchevents.RemoveTargetsOutput{}, nil
	})
	mockCloudWatchEventClient.(*mocks.MockCloudWatchEventClient).SetDeleteRuleFunc(func(
		input *cloudwatchevents.DeleteRuleInput) (*cloudwatchevents.DeleteRuleOutput, error) {
		assert.Equal(t, testScheduleName, *input.Name)
		return &cloudwatchevents.DeleteRuleOutput{}, nil
	})
	scheduler := getCloudWatchSchedulerForTest(mockCloudWatchEventClient)
	assert.Nil(t, scheduler.RemoveSchedule(context.Background(), scheduleInterfaces.RemoveScheduleInput{
		Identifier:         testSchedulerIdentifier,
		ScheduleNamePrefix: testScheduleNamePrefix,
	}))
}

func TestRemoveSchedule_RemoveTargetsError(t *testing.T) {
	mockCloudWatchEventClient := mocks.NewMockCloudWatchEventClient()
	mockCloudWatchEventClient.(*mocks.MockCloudWatchEventClient).SetRemoveTargetsFunc(func(
		input *cloudwatchevents.RemoveTargetsInput) (*cloudwatchevents.RemoveTargetsOutput, error) {
		return nil, expectedError
	})
	scheduler := getCloudWatchSchedulerForTest(mockCloudWatchEventClient)
	err := scheduler.RemoveSchedule(context.Background(), scheduleInterfaces.RemoveScheduleInput{
		Identifier:         testSchedulerIdentifier,
		ScheduleNamePrefix: testScheduleNamePrefix,
	})
	assert.Equal(t, codes.Internal, err.(flyteAdminErrors.FlyteAdminError).Code())
}

func TestRemoveSchedule_InvalidTarget(t *testing.T) {
	mockCloudWatchEventClient := mocks.NewMockCloudWatchEventClient()
	mockCloudWatchEventClient.(*mocks.MockCloudWatchEventClient).SetRemoveTargetsFunc(func(
		input *cloudwatchevents.RemoveTargetsInput) (*cloudwatchevents.RemoveTargetsOutput, error) {
		return nil, awserr.New(cloudwatchevents.ErrCodeResourceNotFoundException, "foo", expectedError)
	})
	scheduler := getCloudWatchSchedulerForTest(mockCloudWatchEventClient)
	err := scheduler.RemoveSchedule(context.Background(), scheduleInterfaces.RemoveScheduleInput{
		Identifier:         testSchedulerIdentifier,
		ScheduleNamePrefix: testScheduleNamePrefix,
	})
	assert.Nil(t, err)
}

func TestRemoveSchedule_DeleteRuleError(t *testing.T) {
	mockCloudWatchEventClient := mocks.NewMockCloudWatchEventClient()
	mockCloudWatchEventClient.(*mocks.MockCloudWatchEventClient).SetRemoveTargetsFunc(func(
		input *cloudwatchevents.RemoveTargetsInput) (*cloudwatchevents.RemoveTargetsOutput, error) {
		return &cloudwatchevents.RemoveTargetsOutput{}, nil
	})
	mockCloudWatchEventClient.(*mocks.MockCloudWatchEventClient).SetDeleteRuleFunc(func(
		input *cloudwatchevents.DeleteRuleInput) (*cloudwatchevents.DeleteRuleOutput, error) {
		return nil, expectedError
	})
	scheduler := getCloudWatchSchedulerForTest(mockCloudWatchEventClient)
	err := scheduler.RemoveSchedule(context.Background(), scheduleInterfaces.RemoveScheduleInput{
		Identifier:         testSchedulerIdentifier,
		ScheduleNamePrefix: testScheduleNamePrefix,
	})
	assert.Equal(t, codes.Internal, err.(flyteAdminErrors.FlyteAdminError).Code())
}

func TestRemoveSchedule_InvalidRule(t *testing.T) {
	mockCloudWatchEventClient := mocks.NewMockCloudWatchEventClient()
	mockCloudWatchEventClient.(*mocks.MockCloudWatchEventClient).SetRemoveTargetsFunc(func(
		input *cloudwatchevents.RemoveTargetsInput) (*cloudwatchevents.RemoveTargetsOutput, error) {
		return &cloudwatchevents.RemoveTargetsOutput{}, nil
	})
	mockCloudWatchEventClient.(*mocks.MockCloudWatchEventClient).SetDeleteRuleFunc(func(
		input *cloudwatchevents.DeleteRuleInput) (*cloudwatchevents.DeleteRuleOutput, error) {
		return nil, awserr.New(cloudwatchevents.ErrCodeResourceNotFoundException, "foo", expectedError)
	})
	scheduler := getCloudWatchSchedulerForTest(mockCloudWatchEventClient)
	err := scheduler.RemoveSchedule(context.Background(), scheduleInterfaces.RemoveScheduleInput{
		Identifier:         testSchedulerIdentifier,
		ScheduleNamePrefix: testScheduleNamePrefix,
	})
	assert.Nil(t, err)
}
