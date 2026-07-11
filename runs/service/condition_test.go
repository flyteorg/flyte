package service

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions"
	actionsconnectmocks "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions/actionsconnect/mocks"
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

func TestRecordAction_Condition(t *testing.T) {
	actionRepo, _, svc := newTestServiceWithTaskRepo(t)

	var created *models.Action
	actionRepo.On("CreateAction", mock.Anything, mock.Anything, false).
		Run(func(args mock.Arguments) { created = args.Get(1).(*models.Action) }).
		Return(&models.Action{}, nil)

	resp, err := svc.RecordAction(context.Background(), connect.NewRequest(&workflow.RecordActionRequest{
		ActionId: testActionID,
		Parent:   "a0",
		Spec:     &workflow.RecordActionRequest_Condition{Condition: testCondition},
	}))
	require.NoError(t, err)
	assert.EqualValues(t, 0, resp.Msg.GetStatus().GetCode())

	require.NotNil(t, created)
	assert.Equal(t, int32(workflow.ActionType_ACTION_TYPE_CONDITION), created.ActionType)
	assert.Equal(t, "approve", created.FunctionName)
	assert.Equal(t, "approve", created.TaskName.String)

	spec := &workflow.ActionSpec{}
	require.NoError(t, proto.Unmarshal(created.ActionSpec, spec))
	assert.True(t, proto.Equal(testCondition, spec.GetCondition()))

	info := &workflow.RunInfo{}
	require.NoError(t, proto.Unmarshal(created.DetailedInfo, info))
	assert.True(t, proto.Equal(testCondition, info.GetCondition()))
}

func TestUpdateActionStatus_PersistsOutput(t *testing.T) {
	actionRepo, _, svc := newTestServiceWithTaskRepo(t)

	existingInfo, err := proto.Marshal(&workflow.RunInfo{InputsUri: "s3://in", Condition: testCondition})
	require.NoError(t, err)

	principal := &common.EnrichedIdentity{
		Principal: &common.EnrichedIdentity_User{
			User: &common.User{Id: &common.UserIdentifier{Subject: "user@example.com"}},
		},
	}

	actionRepo.On("UpdateActionPhase", mock.Anything, testActionID,
		common.ActionPhase_ACTION_PHASE_SUCCEEDED, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	actionRepo.On("GetAction", mock.Anything, testActionID).
		Return(&models.Action{DetailedInfo: existingInfo}, nil)

	var savedInfo []byte
	actionRepo.On("UpdateActionDetailedInfo", mock.Anything, testActionID, mock.Anything).
		Run(func(args mock.Arguments) { savedInfo = args.Get(2).([]byte) }).
		Return(nil)

	resp, err := svc.UpdateActionStatus(context.Background(), connect.NewRequest(&workflow.UpdateActionStatusRequest{
		ActionId:  testActionID,
		Status:    &workflow.ActionStatus{Phase: common.ActionPhase_ACTION_PHASE_SUCCEEDED},
		Output:    testBoolLiteral(true),
		Principal: principal,
	}))
	require.NoError(t, err)
	assert.EqualValues(t, 0, resp.Msg.GetStatus().GetCode())

	info := &workflow.RunInfo{}
	require.NoError(t, proto.Unmarshal(savedInfo, info))
	assert.True(t, proto.Equal(testBoolLiteral(true), info.GetOutput()))
	assert.True(t, proto.Equal(principal, info.GetPrincipal()))
	// Existing RunInfo fields survive the merge.
	assert.Equal(t, "s3://in", info.GetInputsUri())
	assert.True(t, proto.Equal(testCondition, info.GetCondition()))
}

func TestUpdateActionStatus_NoOutputSkipsRunInfoWrite(t *testing.T) {
	actionRepo, _, svc := newTestServiceWithTaskRepo(t)

	actionRepo.On("UpdateActionPhase", mock.Anything, testActionID,
		common.ActionPhase_ACTION_PHASE_RUNNING, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	// No GetAction/UpdateActionDetailedInfo expectations: AssertExpectations
	// fails if the RunInfo write path runs.

	_, err := svc.UpdateActionStatus(context.Background(), connect.NewRequest(&workflow.UpdateActionStatusRequest{
		ActionId: testActionID,
		Status:   &workflow.ActionStatus{Phase: common.ActionPhase_ACTION_PHASE_RUNNING},
	}))
	require.NoError(t, err)
}

func TestSignalEvent(t *testing.T) {
	newSvc := func(t *testing.T) (*actionsconnectmocks.ActionsServiceClient, *RunService) {
		actionsClient := actionsconnectmocks.NewActionsServiceClient(t)
		return actionsClient, &RunService{actionsClient: actionsClient}
	}
	req := &workflow.SignalEventRequest{
		ActionId:         testActionID,
		ParentActionName: "a0",
		Payload:          &workflow.EventPayload{Value: &workflow.EventPayload_BoolValue{BoolValue: true}},
	}

	t.Run("forwards converted literal to actions service", func(t *testing.T) {
		actionsClient, svc := newSvc(t)
		actionsClient.EXPECT().Signal(mock.Anything, mock.MatchedBy(func(r *connect.Request[actions.SignalRequest]) bool {
			return proto.Equal(r.Msg.GetValue(), testBoolLiteral(true)) &&
				r.Msg.GetActionId().GetName() == testActionID.Name &&
				r.Msg.GetParentActionName() == "a0"
		})).Return(connect.NewResponse(&actions.SignalResponse{}), nil)

		_, err := svc.SignalEvent(context.Background(), connect.NewRequest(req))
		assert.NoError(t, err)
	})

	t.Run("actions service error passes through", func(t *testing.T) {
		actionsClient, svc := newSvc(t)
		actionsClient.EXPECT().Signal(mock.Anything, mock.Anything).
			Return(nil, connect.NewError(connect.CodeFailedPrecondition, errors.New("already completed")))

		_, err := svc.SignalEvent(context.Background(), connect.NewRequest(req))
		assert.Equal(t, connect.CodeFailedPrecondition, connect.CodeOf(err))
	})
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

	resp, err := svc.GetActionDetails(context.Background(), connect.NewRequest(&workflow.GetActionDetailsRequest{
		ActionId: testActionID,
	}))
	require.NoError(t, err)

	details := resp.Msg.GetDetails()
	assert.Equal(t, workflow.ActionType_ACTION_TYPE_CONDITION, details.GetMetadata().GetActionType())
	assert.Equal(t, "approve", details.GetMetadata().GetCondition().GetName())
	assert.True(t, proto.Equal(testCondition, details.GetCondition()))
	require.NotNil(t, details.GetSignalInfo())
	assert.True(t, proto.Equal(testBoolLiteral(true), details.GetSignalInfo().GetOutput()))
	assert.True(t, proto.Equal(principal, details.GetSignalInfo().GetSignalledBy()))
}
