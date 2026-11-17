package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

var testCondition = &workflow.ConditionAction{
	Name:   "approve",
	Type:   &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_BOOLEAN}},
	Prompt: "approve?",
}

func testBoolLiteral(v bool) *core.Literal {
	return &core.Literal{
		Value: &core.Literal_Scalar{Scalar: &core.Scalar{
			Value: &core.Scalar_Primitive{Primitive: &core.Primitive{
				Value: &core.Primitive_Boolean{Boolean: v},
			}},
		}},
	}
}

func TestPayloadToLiteral(t *testing.T) {
	assert.True(t, proto.Equal(testBoolLiteral(true),
		payloadToLiteral(&workflow.EventPayload{Value: &workflow.EventPayload_BoolValue{BoolValue: true}})))
	assert.EqualValues(t, 42,
		payloadToLiteral(&workflow.EventPayload{Value: &workflow.EventPayload_IntValue{IntValue: 42}}).GetScalar().GetPrimitive().GetInteger())
	assert.EqualValues(t, 1.5,
		payloadToLiteral(&workflow.EventPayload{Value: &workflow.EventPayload_FloatValue{FloatValue: 1.5}}).GetScalar().GetPrimitive().GetFloatValue())
	assert.Equal(t, "yes",
		payloadToLiteral(&workflow.EventPayload{Value: &workflow.EventPayload_StringValue{StringValue: "yes"}}).GetScalar().GetPrimitive().GetStringValue())
	assert.Nil(t, payloadToLiteral(nil))
}

func TestGetActionDetails_ConditionWithSignalInfo(t *testing.T) {
	actionRepo, _, svc := newTestServiceWithTaskRepo(t)

	principal := &common.EnrichedIdentity{
		Principal: &common.EnrichedIdentity_User{
			User: &common.User{Id: &common.UserIdentifier{Subject: "user@example.com"}},
		},
	}
	detailedInfo, err := proto.Marshal(&workflow.RunInfo{
		Condition: testCondition,
		Output:    testBoolLiteral(true),
		Principal: principal,
	})
	require.NoError(t, err)
	actionSpec, err := proto.Marshal(&workflow.ActionSpec{
		ActionId: testActionID,
		Spec:     &workflow.ActionSpec_Condition{Condition: testCondition},
	})
	require.NoError(t, err)

	actionRepo.On("GetAction", mock.Anything, testActionID).Return(&models.Action{
		Project:      testActionID.Run.Project,
		Domain:       testActionID.Run.Domain,
		RunName:      testActionID.Run.Name,
		Name:         testActionID.Name,
		ActionType:   int32(workflow.ActionType_ACTION_TYPE_CONDITION),
		FunctionName: "approve",
		Phase:        int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED),
		DetailedInfo: detailedInfo,
		ActionSpec:   actionSpec,
	}, nil)
	actionRepo.On("ListEvents", mock.Anything, testActionID, 500).Return([]*models.ActionEvent{}, nil)

	resp, err := svc.GetActionDetails(context.Background(), &workflow.GetActionDetailsRequest{
		ActionId: testActionID,
	})
	require.NoError(t, err)

	details := resp.GetDetails()
	assert.Equal(t, workflow.ActionType_ACTION_TYPE_CONDITION, details.GetMetadata().GetActionType())
	assert.Equal(t, "approve", details.GetMetadata().GetCondition().GetName())
	assert.True(t, proto.Equal(testCondition, details.GetCondition()))
	require.NotNil(t, details.GetSignalInfo())
	assert.True(t, proto.Equal(testBoolLiteral(true), details.GetSignalInfo().GetOutput()))
	assert.True(t, proto.Equal(principal, details.GetSignalInfo().GetSignalledBy()))
}
