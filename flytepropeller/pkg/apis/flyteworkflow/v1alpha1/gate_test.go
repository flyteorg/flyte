package v1alpha1

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/jsonpb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func mockMarshalPbToBytes(msg proto.Message) ([]byte, error) {
	var buf bytes.Buffer
	jMarshaller := jsonpb.Marshaler{}
	if err := jMarshaller.Marshal(&buf, msg); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func TestApproveConditionJSONMarshalling(t *testing.T) {
	approveCondition := ApproveCondition{
		&core.ApproveCondition{
			SignalId: "TestSignalId",
		},
	}

	expected, mockErr := mockMarshalPbToBytes(approveCondition.ApproveCondition)
	assert.Nil(t, mockErr)

	// MarshalJSON
	approveConditionBytes, mErr := approveCondition.MarshalJSON()
	assert.Nil(t, mErr)
	assert.Equal(t, expected, approveConditionBytes)

	// UnmarshalJSON
	approveConditionObj := &ApproveCondition{}
	uErr := approveConditionObj.UnmarshalJSON(approveConditionBytes)
	assert.Nil(t, uErr)
	assert.Equal(t, approveCondition.SignalId, approveConditionObj.SignalId)
}

func TestSignalConditionJSONMarshalling(t *testing.T) {
	signalCondition := SignalCondition{
		&core.SignalCondition{
			SignalId: "TestSignalId",
		},
	}

	expected, mockErr := mockMarshalPbToBytes(signalCondition.SignalCondition)
	assert.Nil(t, mockErr)

	// MarshalJSON
	signalConditionBytes, mErr := signalCondition.MarshalJSON()
	assert.Nil(t, mErr)
	assert.Equal(t, expected, signalConditionBytes)

	// UnmarshalJSON
	signalConditionObj := &SignalCondition{}
	uErr := signalConditionObj.UnmarshalJSON(signalConditionBytes)
	assert.Nil(t, uErr)
	assert.Equal(t, signalCondition.SignalId, signalConditionObj.SignalId)
}

func TestSleepConditionJSONMarshalling(t *testing.T) {
	sleepCondition := SleepCondition{
		&core.SleepCondition{
			Duration: &durationpb.Duration{
				Seconds: 10,
				Nanos:   10,
			},
		},
	}

	expected, mockErr := mockMarshalPbToBytes(sleepCondition.SleepCondition)
	assert.Nil(t, mockErr)

	// MarshalJSON
	sleepConditionBytes, mErr := sleepCondition.MarshalJSON()
	assert.Nil(t, mErr)
	assert.Equal(t, expected, sleepConditionBytes)

	// UnmarshalJSON
	sleepConditionObj := &SleepCondition{}
	uErr := sleepConditionObj.UnmarshalJSON(sleepConditionBytes)
	assert.Nil(t, uErr)
	assert.Equal(t, sleepCondition.Duration, sleepConditionObj.Duration)
}

func TestGateNodeSpec_GetKind(t *testing.T) {
	kind := ConditionKindApprove
	gateNodeSpec := GateNodeSpec{
		Kind: kind,
	}

	if gateNodeSpec.GetKind() != kind {
		t.Errorf("Expected %s, but got %s", kind, gateNodeSpec.GetKind())
	}
}

func TestGateNodeSpec_GetApprove(t *testing.T) {
	approveCondition := &ApproveCondition{
		&core.ApproveCondition{
			SignalId: "TestSignalId",
		},
	}
	gateNodeSpec := GateNodeSpec{
		Approve: approveCondition,
	}

	if gateNodeSpec.GetApprove() != approveCondition.ApproveCondition {
		t.Errorf("Expected approveCondition, but got a different value")
	}
}

func TestGateNodeSpec_GetSignal(t *testing.T) {
	signalCondition := &SignalCondition{
		&core.SignalCondition{
			SignalId: "TestSignalId",
		},
	}
	gateNodeSpec := GateNodeSpec{
		Signal: signalCondition,
	}

	if gateNodeSpec.GetSignal() != signalCondition.SignalCondition {
		t.Errorf("Expected signalCondition, but got a different value")
	}
}

func TestGateNodeSpec_GetSleep(t *testing.T) {
	sleepCondition := &SleepCondition{
		&core.SleepCondition{
			Duration: &durationpb.Duration{
				Seconds: 10,
				Nanos:   10,
			},
		},
	}
	gateNodeSpec := GateNodeSpec{
		Sleep: sleepCondition,
	}

	if gateNodeSpec.GetSleep() != sleepCondition.SleepCondition {
		t.Errorf("Expected sleepCondition, but got a different value")
	}
}
