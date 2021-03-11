// Functions for serializing and deserializing scheduled events in AWS.
package aws

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
)

const awsTimestampPlaceholder = "<time>"

// The structured message that will be serialized and added to event schedules.
type ScheduleWorkflowPayload struct {
	// The timestamp plalceholder arg which will be populated by AWS CloudWatch.
	Time string `json:"time"`
	// The name of the kickoff time input argument in the workflow definition. This will be filled with kickoff time.
	KickoffTimeArg string `json:"kickoff_time_arg"`
	// Serialized launch plan admin.Identifier.
	Payload []byte `json:"payload"`
}

// Encapsulates the data necessary to trigger a scheduled workflow execution.
type ScheduledWorkflowExecutionRequest struct {
	// The time at which the schedule event was triggered.
	KickoffTime time.Time
	// The name of the kickoff time input argument in the workflow definition. This will be filled with kickoff time.
	KickoffTimeArg string
	// The desired launch plan identifier to trigger on schedule event firings.
	LaunchPlanIdentifier admin.NamedEntityIdentifier
}

// This produces a function that is used to serialize messages enqueued on the cloudwatch scheduler.
func SerializeScheduleWorkflowPayload(
	kickoffTimeArg string, launchPlanIdentifier admin.NamedEntityIdentifier) (*string, error) {
	payload, err := proto.Marshal(&launchPlanIdentifier)
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "failed to marshall launch plan with err: %v", err)
	}
	sEnc := base64.StdEncoding.EncodeToString(payload)
	// Why not initialize a ScheduleWorkflowPayload struct and marshal to JSON? SQS doesn't actually use valid JSON
	// and the <time> placeholder isn't a string - so we manually construct the input template as a string.
	inputTemplateJSONStr := fmt.Sprintf(
		`{"time":%s,"kickoff_time_arg":"%s","payload":"%s"}`, awsTimestampPlaceholder, kickoffTimeArg, sEnc)
	logger.Debugf(context.Background(), "serialized schedule workflow payload for launch plan [%+v]: %s",
		launchPlanIdentifier, inputTemplateJSONStr)
	return &inputTemplateJSONStr, nil
}

func DeserializeScheduleWorkflowPayload(payload []byte) (ScheduledWorkflowExecutionRequest, error) {
	var scheduleWorkflowPayload ScheduleWorkflowPayload
	err := json.Unmarshal(payload, &scheduleWorkflowPayload)
	if err != nil {
		return ScheduledWorkflowExecutionRequest{}, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"failed to deserialize ScheduleWorkflowPayload bytes %x with err: %v", payload, err)
	}
	// AWS CloudWatch will fill in the timestamp placeholder argument using RFC 3339 format:
	// https://docs.aws.amazon.com/AmazonCloudWatch/latest/events/CloudWatchEventsandEventPatterns.html
	kickoffTime, err := time.Parse(time.RFC3339, scheduleWorkflowPayload.Time)
	if err != nil {
		return ScheduledWorkflowExecutionRequest{}, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"failed to deserialize ScheduleWorkflowPayload timestamp %s with err: %v",
			scheduleWorkflowPayload.Time, err)
	}
	var launchPlanIdentifier admin.NamedEntityIdentifier
	err = proto.Unmarshal(scheduleWorkflowPayload.Payload, &launchPlanIdentifier)
	if err != nil {
		return ScheduledWorkflowExecutionRequest{}, errors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"failed to deserialize ScheduleWorkflowPayload ExecutionCreateRequest %s with err: %v",
			scheduleWorkflowPayload.Payload, err)
	}

	return ScheduledWorkflowExecutionRequest{
		KickoffTime:          kickoffTime,
		KickoffTimeArg:       scheduleWorkflowPayload.KickoffTimeArg,
		LaunchPlanIdentifier: launchPlanIdentifier,
	}, nil
}
