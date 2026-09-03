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

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
	repoMocks "github.com/flyteorg/flyte/v2/runs/repository/mocks"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

// newRecoveryTestService mirrors newTestService but leaves ActionRepo() optional: most
// validation failures are decided before any repository call, so asserting the accessor was
// reached would make the tests assert the opposite of what they are checking.
func newRecoveryTestService(t *testing.T) (*repoMocks.ActionRepo, *RunService) {
	actionRepo := &repoMocks.ActionRepo{}
	repo := &repoMocks.Repository{}
	repo.On("ActionRepo").Maybe().Return(actionRepo)

	t.Cleanup(func() {
		repo.AssertExpectations(t)
		actionRepo.AssertExpectations(t)
	})

	return actionRepo, &RunService{repo: repo}
}

func recoveryTestRunID() *common.RunIdentifier {
	return &common.RunIdentifier{Org: "org1", Project: "proj1", Domain: "dev", Name: "new-run"}
}

func recoverRelation(source *common.RunIdentifier) *task.RunSpec {
	return &task.RunSpec{
		Relation: &common.Relation{
			RelatedTo:    source,
			RelationType: common.RelationType_RELATION_TYPE_RECOVER,
		},
	}
}

func sourceRunID(name string) *common.RunIdentifier {
	return &common.RunIdentifier{Org: "org1", Project: "proj1", Domain: "dev", Name: name}
}

// The action lookup is keyed by run identity alone, so a cross-scope relation would read
// another tenant's rows and hand back their output URIs.
func TestValidateRecovery_RejectsCrossScopeSource(t *testing.T) {
	_, svc := newRecoveryTestService(t)

	for _, tc := range []struct {
		name   string
		source *common.RunIdentifier
	}{
		{"other project", &common.RunIdentifier{Org: "org1", Project: "other", Domain: "dev", Name: "r1"}},
		{"other domain", &common.RunIdentifier{Org: "org1", Project: "proj1", Domain: "prod", Name: "r1"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := svc.validateRecovery(context.Background(), recoveryTestRunID(), recoverRelation(tc.source))
			require.Error(t, err)
			assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
		})
	}
}

func TestValidateRecovery_RejectsMissingSourceRun(t *testing.T) {
	actionRepo, svc := newRecoveryTestService(t)
	source := sourceRunID("gone")

	actionRepo.On("GetAction", mock.Anything,
		matchActionID(&common.ActionIdentifier{Run: source, Name: RootActionName})).
		Return(nil, errors.New("action not found")).Once()

	err := svc.validateRecovery(context.Background(), recoveryTestRunID(), recoverRelation(source))
	require.Error(t, err)
	assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
}

// A source still in flight has actions changing phase underneath the lookup, so every
// recovery decision taken against it would be a race.
func TestValidateRecovery_RejectsNonTerminalSourceRun(t *testing.T) {
	actionRepo, svc := newRecoveryTestService(t)
	source := sourceRunID("still-going")

	actionRepo.On("GetAction", mock.Anything,
		matchActionID(&common.ActionIdentifier{Run: source, Name: RootActionName})).
		Return(&models.Action{Phase: int32(common.ActionPhase_ACTION_PHASE_RUNNING)}, nil).Once()

	err := svc.validateRecovery(context.Background(), recoveryTestRunID(), recoverRelation(source))
	require.Error(t, err)
	assert.Equal(t, connect.CodeFailedPrecondition, connect.CodeOf(err))
}

func TestValidateRecovery_AcceptsTerminalSourceRun(t *testing.T) {
	actionRepo, svc := newRecoveryTestService(t)
	source := sourceRunID("r1")

	for _, phase := range []common.ActionPhase{
		common.ActionPhase_ACTION_PHASE_SUCCEEDED,
		common.ActionPhase_ACTION_PHASE_FAILED,
		common.ActionPhase_ACTION_PHASE_RECOVERED,
	} {
		actionRepo.On("GetAction", mock.Anything,
			matchActionID(&common.ActionIdentifier{Run: source, Name: RootActionName})).
			Return(&models.Action{Phase: int32(phase)}, nil).Once()

		require.NoError(t, svc.validateRecovery(context.Background(), recoveryTestRunID(), recoverRelation(source)),
			"phase %s", phase)
	}
}

func lookupActionID(run, name string) *common.ActionIdentifier {
	return &common.ActionIdentifier{Run: sourceRunID(run), Name: name}
}

func TestLookupAction_MissingActionIsNotAnError(t *testing.T) {
	actionRepo, svc := newRecoveryTestService(t)
	actionID := lookupActionID("r1", "a5")

	actionRepo.On("GetAction", mock.Anything, matchActionID(actionID)).
		Return(nil, interfaces.ErrActionNotFound).Once()

	resp, err := svc.LookupAction(context.Background(),
		connect.NewRequest(&workflow.LookupActionRequest{ActionId: actionID}))
	require.NoError(t, err)
	assert.False(t, resp.Msg.GetFound())
}

// A failing lookup must stay distinguishable from a miss: the caller counts them separately.
func TestLookupAction_RepositoryFailureIsAnError(t *testing.T) {
	actionRepo, svc := newRecoveryTestService(t)
	actionID := lookupActionID("r1", "a5")

	actionRepo.On("GetAction", mock.Anything, matchActionID(actionID)).
		Return(nil, errors.New("connection refused")).Once()

	_, err := svc.LookupAction(context.Background(),
		connect.NewRequest(&workflow.LookupActionRequest{ActionId: actionID}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeInternal, connect.CodeOf(err))
}

func TestLookupAction_TaskActionReadsLastAttemptOutputs(t *testing.T) {
	actionRepo, svc := newRecoveryTestService(t)
	actionID := lookupActionID("r1", "a5")

	actionRepo.On("GetAction", mock.Anything, matchActionID(actionID)).
		Return(&models.Action{
			Phase:       int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED),
			ActionType:  int32(workflow.ActionType_ACTION_TYPE_TASK),
			Attempts:    2,
			CacheStatus: core.CatalogCacheStatus_CACHE_HIT,
		}, nil).Once()
	actionRepo.On("ListEvents", mock.Anything, matchActionID(actionID), 500).
		Return([]*models.ActionEvent{
			newOutputEvent(t, actionID, 0, "s3://bucket/r1/a5/0/outputs.pb"),
			newOutputEvent(t, actionID, 1, "s3://bucket/r1/a5/1/outputs.pb"),
		}, nil).Once()

	resp, err := svc.LookupAction(context.Background(),
		connect.NewRequest(&workflow.LookupActionRequest{ActionId: actionID}))
	require.NoError(t, err)
	assert.True(t, resp.Msg.GetFound())
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_SUCCEEDED, resp.Msg.GetPhase())
	assert.Equal(t, uint32(2), resp.Msg.GetAttempts())
	assert.Equal(t, core.CatalogCacheStatus_CACHE_HIT, resp.Msg.GetCacheStatus())
	assert.Equal(t, "s3://bucket/r1/a5/1/outputs.pb", resp.Msg.GetOutputUri())
}

// A trace never emits attempt outputs; its URI lives on the stored RunInfo.
func TestLookupAction_TraceActionReadsDetailedInfo(t *testing.T) {
	actionRepo, svc := newRecoveryTestService(t)
	actionID := lookupActionID("r1", "trace1")

	info, err := proto.Marshal(&workflow.RunInfo{OutputsUri: "s3://bucket/r1/trace1/outputs.pb"})
	require.NoError(t, err)
	actionRepo.On("GetAction", mock.Anything, matchActionID(actionID)).
		Return(&models.Action{
			Phase:        int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED),
			ActionType:   int32(workflow.ActionType_ACTION_TYPE_TRACE),
			DetailedInfo: info,
		}, nil).Once()

	resp, err := svc.LookupAction(context.Background(),
		connect.NewRequest(&workflow.LookupActionRequest{ActionId: actionID}))
	require.NoError(t, err)
	assert.Equal(t, "s3://bucket/r1/trace1/outputs.pb", resp.Msg.GetOutputUri())
}

// A signalled condition has no outputs file: its result is a Literal on the RunInfo, and it
// is the only thing a recovery of that condition can hand downstream.
func TestLookupAction_ConditionReturnsInlineSignalValue(t *testing.T) {
	actionRepo, svc := newRecoveryTestService(t)
	actionID := lookupActionID("r1", "cond1")

	signal := &core.Literal{
		Value: &core.Literal_Scalar{Scalar: &core.Scalar{
			Value: &core.Scalar_Primitive{Primitive: &core.Primitive{
				Value: &core.Primitive_Boolean{Boolean: true},
			}},
		}},
	}
	info, err := proto.Marshal(&workflow.RunInfo{Output: signal})
	require.NoError(t, err)
	actionRepo.On("GetAction", mock.Anything, matchActionID(actionID)).
		Return(&models.Action{
			Phase:        int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED),
			ActionType:   int32(workflow.ActionType_ACTION_TYPE_CONDITION),
			DetailedInfo: info,
		}, nil).Once()
	actionRepo.On("ListEvents", mock.Anything, matchActionID(actionID), 500).
		Return([]*models.ActionEvent{}, nil).Once()

	resp, err := svc.LookupAction(context.Background(),
		connect.NewRequest(&workflow.LookupActionRequest{ActionId: actionID}))
	require.NoError(t, err)
	assert.Empty(t, resp.Msg.GetOutputUri(), "a condition writes no outputs file")
	require.NotNil(t, resp.Msg.GetOutput())
	assert.True(t, proto.Equal(signal, resp.Msg.GetOutput()))
}

// An action that succeeded without producing outputs is found, with no URI to hand back.
func TestLookupAction_NoOutputsYieldsEmptyURI(t *testing.T) {
	actionRepo, svc := newRecoveryTestService(t)
	actionID := lookupActionID("r1", "a5")

	actionRepo.On("GetAction", mock.Anything, matchActionID(actionID)).
		Return(&models.Action{
			Phase:      int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED),
			ActionType: int32(workflow.ActionType_ACTION_TYPE_TASK),
		}, nil).Once()
	actionRepo.On("ListEvents", mock.Anything, matchActionID(actionID), 500).
		Return([]*models.ActionEvent{}, nil).Once()

	resp, err := svc.LookupAction(context.Background(),
		connect.NewRequest(&workflow.LookupActionRequest{ActionId: actionID}))
	require.NoError(t, err)
	assert.True(t, resp.Msg.GetFound())
	assert.Empty(t, resp.Msg.GetOutputUri())
}

func newOutputEvent(t *testing.T, actionID *common.ActionIdentifier, attempt uint32, outputURI string) *models.ActionEvent {
	t.Helper()
	info, err := proto.Marshal(&workflow.ActionEvent{
		Id:      actionID,
		Attempt: attempt,
		Phase:   common.ActionPhase_ACTION_PHASE_SUCCEEDED,
		Outputs: &task.OutputReferences{OutputUri: outputURI},
	})
	require.NoError(t, err)
	return &models.ActionEvent{
		Project: actionID.GetRun().GetProject(),
		Domain:  actionID.GetRun().GetDomain(),
		RunName: actionID.GetRun().GetName(),
		Name:    actionID.GetName(),
		Attempt: attempt,
		Phase:   int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED),
		Info:    info,
	}
}
