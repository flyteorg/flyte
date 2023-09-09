package aws

import (
	"fmt"
	"testing"
	"time"

	"github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/golang/protobuf/proto"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
)

const testKickoffTimeArg = "kickoff time arg"

var testLaunchPlanIdentifier = admin.NamedEntityIdentifier{
	Name:    "name",
	Project: "project",
	Domain:  "domain",
}

func TestNewSerializeScheduleWorkflowPayloadFunc(t *testing.T) {
	payload, err := SerializeScheduleWorkflowPayload(testKickoffTimeArg, testLaunchPlanIdentifier)
	assert.Nil(t, err)
	expectedPayload := "{\"time\":<time>,\"kickoff_time_arg\":\"kickoff time arg\",\"payload\":" +
		"\"Cgdwcm9qZWN0EgZkb21haW4aBG5hbWU=\"}"
	assert.Equal(t, expectedPayload, *payload)
}

func TestDeserializeScheduleWorkflowPayload(t *testing.T) {
	payload := "{\"time\":\"2017-12-22T18:43:48Z\",\"kickoff_time_arg\":\"kickoff time arg\",\"payload\":" +
		"\"Cgdwcm9qZWN0EgZkb21haW4aBG5hbWU=\"}"
	scheduledWorkflowExecutionRequest, err := DeserializeScheduleWorkflowPayload([]byte(payload))
	assert.Nil(t, err)
	assert.Equal(t,
		time.Date(2017, 12, 22, 18, 43, 48, 0, time.UTC),
		scheduledWorkflowExecutionRequest.KickoffTime)
	assert.Equal(t, testKickoffTimeArg, scheduledWorkflowExecutionRequest.KickoffTimeArg)
	assert.True(t, proto.Equal(&testLaunchPlanIdentifier, &scheduledWorkflowExecutionRequest.LaunchPlanIdentifier),
		fmt.Sprintf("scheduledWorkflowExecutionRequest.LaunchPlanIdentifier %v", &scheduledWorkflowExecutionRequest.LaunchPlanIdentifier))
}

func TestDeserializeScheduleWorkflowPayload_MessageError(t *testing.T) {
	payload := "{wonky fake json}"
	_, err := DeserializeScheduleWorkflowPayload([]byte(payload))
	assert.Equal(t, codes.InvalidArgument, err.(errors.FlyteAdminError).Code())
}

func TestDeserializeScheduleWorkflowPayload_TimestampError(t *testing.T) {
	payload := "{\"time\":\"timestamp\",\"kickoff_time_arg\":\"kickoff time arg\",\"payload\"" +
		":\"\n\aproject\x12\x06domain\x1a\x04name\"}"
	_, err := DeserializeScheduleWorkflowPayload([]byte(payload))
	assert.Equal(t, codes.InvalidArgument, err.(errors.FlyteAdminError).Code())
}

func TestDeserializeScheduleWorkflowPayload_PayloadError(t *testing.T) {
	payload := "{\"time\":\"2017-12-22T18:43:48Z\",\"kickoff_time_arg\":\"kickoff time arg\",\"payload\":\"foobar\"}"
	_, err := DeserializeScheduleWorkflowPayload([]byte(payload))
	assert.Equal(t, codes.InvalidArgument, err.(errors.FlyteAdminError).Code())
}
