package k8s

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	executorv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	runmocks "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect/mocks"
)

func newGateTestClient(t *testing.T) (*runmocks.InternalRunServiceClient, *ActionsClient) {
	runClient := runmocks.NewInternalRunServiceClient(t)
	return runClient, &ActionsClient{
		runClient:       runClient,
		recoveryMetrics: newRecoveryMetrics(promutils.NewTestScope()),
	}
}

func recoveryContextFor(sourceRun string, forceRerun ...string) *executorv1.RecoveryContext {
	relation, err := proto.Marshal(&common.Relation{
		RelatedTo: &common.RunIdentifier{
			Org: "org1", Project: "proj", Domain: "dev", Name: sourceRun,
		},
		RelationType: common.RelationType_RELATION_TYPE_RECOVER,
	})
	if err != nil {
		panic(err)
	}
	return &executorv1.RecoveryContext{Relation: relation, ForceRerunActions: forceRerun}
}

func taskActionWith(recoveryContext *executorv1.RecoveryContext) *executorv1.TaskAction {
	return &executorv1.TaskAction{
		Spec: executorv1.TaskActionSpec{RecoveryContext: recoveryContext},
	}
}

func childAction(name string) *actions.Action {
	parent := "a0"
	return &actions.Action{
		ActionId: &common.ActionIdentifier{
			Run:  &common.RunIdentifier{Org: "org1", Project: "proj", Domain: "dev", Name: "run2"},
			Name: name,
		},
		ParentActionName: &parent,
	}
}

func lookupResponse(msg *workflow.LookupActionResponse) *connect.Response[workflow.LookupActionResponse] {
	return connect.NewResponse(msg)
}

// Gate 1: the root decides which children to enqueue, so it always runs fresh.
func TestResolveRecoveredFrom_RootAlwaysRunsFresh(t *testing.T) {
	_, c := newGateTestClient(t)

	got := c.resolveRecoveredFrom(context.Background(),
		taskActionWith(recoveryContextFor("r1")), childAction("a0"), true)
	assert.Nil(t, got)
}

// Gate 2: an ordinary run carries no recovery context at all.
func TestResolveRecoveredFrom_NoRecoveryContext(t *testing.T) {
	_, c := newGateTestClient(t)

	got := c.resolveRecoveredFrom(context.Background(), taskActionWith(nil), childAction("a1"), false)
	assert.Nil(t, got)
}

// Gate 3: the escape hatch is evaluated before any lookup — the mock would fail the test
// if a lookup were attempted.
func TestResolveRecoveredFrom_ForcedActionSkipsLookup(t *testing.T) {
	_, c := newGateTestClient(t)

	got := c.resolveRecoveredFrom(context.Background(),
		taskActionWith(recoveryContextFor("r1", "a1", "a7")), childAction("a1"), false)
	assert.Nil(t, got)
}

// Gate 4: the lookup is keyed by run identity, so a cross-scope relation would read
// another tenant's rows.
func TestResolveRecoveredFrom_CrossScopeSourceSkipsLookup(t *testing.T) {
	_, c := newGateTestClient(t)

	relation, err := proto.Marshal(&common.Relation{
		RelatedTo: &common.RunIdentifier{
			Org: "org1", Project: "other", Domain: "dev", Name: "r1",
		},
		RelationType: common.RelationType_RELATION_TYPE_RECOVER,
	})
	require.NoError(t, err)

	got := c.resolveRecoveredFrom(context.Background(),
		taskActionWith(&executorv1.RecoveryContext{Relation: relation}), childAction("a1"), false)
	assert.Nil(t, got)
}

func TestResolveRecoveredFrom_HitStampsSourceResult(t *testing.T) {
	runClient, c := newGateTestClient(t)
	runClient.EXPECT().LookupAction(mock.Anything, mock.Anything).
		Return(lookupResponse(&workflow.LookupActionResponse{
			Found:       true,
			Phase:       common.ActionPhase_ACTION_PHASE_SUCCEEDED,
			Attempts:    2,
			CacheStatus: core.CatalogCacheStatus_CACHE_HIT,
			OutputUri:   "s3://bucket/r1/a1/1/outputs.pb",
		}), nil).Once()

	got := c.resolveRecoveredFrom(context.Background(),
		taskActionWith(recoveryContextFor("r1")), childAction("a1"), false)

	require.NotNil(t, got)
	assert.Equal(t, "r1", got.SourceRunName)
	assert.Equal(t, "s3://bucket/r1/a1/1/outputs.pb", got.OutputUri)
	assert.Equal(t, uint32(2), got.Attempts)
	assert.Equal(t, int32(core.CatalogCacheStatus_CACHE_HIT), got.CacheStatus)
}

// A signalled condition settles with an inline Literal and no outputs file. Gating on the
// URI alone re-ran it, and a fresh condition pauses for a signal the source run was already
// given — so the run being recovered hung instead of finishing.
func TestResolveRecoveredFrom_SignalledConditionIsAHit(t *testing.T) {
	runClient, c := newGateTestClient(t)
	signal := &core.Literal{
		Value: &core.Literal_Scalar{Scalar: &core.Scalar{
			Value: &core.Scalar_Primitive{Primitive: &core.Primitive{
				Value: &core.Primitive_Boolean{Boolean: true},
			}},
		}},
	}
	runClient.EXPECT().LookupAction(mock.Anything, mock.Anything).
		Return(lookupResponse(&workflow.LookupActionResponse{
			Found:  true,
			Phase:  common.ActionPhase_ACTION_PHASE_SUCCEEDED,
			Output: signal,
		}), nil).Once()

	got := c.resolveRecoveredFrom(context.Background(),
		taskActionWith(recoveryContextFor("r1")), childAction("a1"), false)

	require.NotNil(t, got, "a condition with an inline result must recover")
	assert.Empty(t, got.OutputUri)

	roundTripped := &core.Literal{}
	require.NoError(t, proto.Unmarshal(got.Output, roundTripped))
	assert.True(t, proto.Equal(signal, roundTripped))
}

// Chained recovery is the ordinary path: recovering a recovery run lands on RECOVERED rows
// whose URI already points further back.
func TestResolveRecoveredFrom_RecoveredSourceIsAHit(t *testing.T) {
	runClient, c := newGateTestClient(t)
	runClient.EXPECT().LookupAction(mock.Anything, mock.Anything).
		Return(lookupResponse(&workflow.LookupActionResponse{
			Found:     true,
			Phase:     common.ActionPhase_ACTION_PHASE_RECOVERED,
			OutputUri: "s3://bucket/r0/a1/1/outputs.pb",
		}), nil).Once()

	got := c.resolveRecoveredFrom(context.Background(),
		taskActionWith(recoveryContextFor("r1")), childAction("a1"), false)

	require.NotNil(t, got)
	assert.Equal(t, "s3://bucket/r0/a1/1/outputs.pb", got.OutputUri)
}

func TestResolveRecoveredFrom_MissAndUnusableSourcesRunFresh(t *testing.T) {
	for _, tc := range []struct {
		name string
		resp *workflow.LookupActionResponse
	}{
		{"missing", &workflow.LookupActionResponse{Found: false}},
		{"failed", &workflow.LookupActionResponse{
			Found: true, Phase: common.ActionPhase_ACTION_PHASE_FAILED, OutputUri: "s3://x",
		}},
		{"aborted", &workflow.LookupActionResponse{
			Found: true, Phase: common.ActionPhase_ACTION_PHASE_ABORTED, OutputUri: "s3://x",
		}},
		{"succeeded without outputs", &workflow.LookupActionResponse{
			Found: true, Phase: common.ActionPhase_ACTION_PHASE_SUCCEEDED,
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runClient, c := newGateTestClient(t)
			runClient.EXPECT().LookupAction(mock.Anything, mock.Anything).
				Return(lookupResponse(tc.resp), nil).Once()

			got := c.resolveRecoveredFrom(context.Background(),
				taskActionWith(recoveryContextFor("r1")), childAction("a1"), false)
			assert.Nil(t, got)
		})
	}
}

// Recovery is an optimisation: a lookup that fails degrades into a fresh run rather than
// failing the enqueue.
func TestResolveRecoveredFrom_LookupFailureRunsFresh(t *testing.T) {
	runClient, c := newGateTestClient(t)
	runClient.EXPECT().LookupAction(mock.Anything, mock.Anything).
		Return(nil, connect.NewError(connect.CodeUnavailable, assert.AnError)).Once()

	got := c.resolveRecoveredFrom(context.Background(),
		taskActionWith(recoveryContextFor("r1")), childAction("a1"), false)
	assert.Nil(t, got)
}

func TestResolveRecoveredFrom_UnparseableRelationRunsFresh(t *testing.T) {
	_, c := newGateTestClient(t)

	got := c.resolveRecoveredFrom(context.Background(),
		taskActionWith(&executorv1.RecoveryContext{Relation: []byte("not-a-proto")}),
		childAction("a1"), false)
	assert.Nil(t, got)
}

// The watch stream is what a running task consumes, so a recovered action must advertise the
// source run's outputs there too — computing a path under this run's base points at a
// location nothing ever wrote.
func TestBuildOutputUri_RecoveredActionUsesSourceOutputs(t *testing.T) {
	ta := &executorv1.TaskAction{
		Spec: executorv1.TaskActionSpec{
			RunOutputBase: "s3://bucket/org/proj/dev/recovery-run",
			ActionName:    "a1",
			RecoveredFrom: &executorv1.RecoveredFrom{
				SourceRunName: "source-run",
				OutputUri:     "s3://bucket/org/proj/dev/source-run/a1/1/outputs.pb",
			},
		},
	}

	// The SDK joins "outputs.pb" onto whatever the watch stream carries, so this must be the
	// directory: returning the file itself produced .../outputs.pb/outputs.pb.
	assert.Equal(t, "s3://bucket/org/proj/dev/source-run/a1/1", BuildOutputUri(context.Background(), ta))
}

func TestBuildOutputUri_OrdinaryActionUsesItsOwnRun(t *testing.T) {
	ta := &executorv1.TaskAction{
		Spec: executorv1.TaskActionSpec{
			RunOutputBase: "s3://bucket/org/proj/dev/run1",
			ActionName:    "a1",
		},
	}

	got := BuildOutputUri(context.Background(), ta)
	assert.Contains(t, got, "run1")
	assert.NotEmpty(t, got)
}

func TestBuildOutputUri_RecoveredWithoutOutputsFileSuffix(t *testing.T) {
	ta := &executorv1.TaskAction{
		Spec: executorv1.TaskActionSpec{
			RunOutputBase: "s3://bucket/org/proj/dev/recovery-run",
			ActionName:    "a1",
			RecoveredFrom: &executorv1.RecoveredFrom{
				SourceRunName: "source-run",
				OutputUri:     "s3://bucket/org/proj/dev/source-run/trace1",
			},
		},
	}

	assert.Equal(t, "s3://bucket/org/proj/dev/source-run/trace1", BuildOutputUri(context.Background(), ta))
}
