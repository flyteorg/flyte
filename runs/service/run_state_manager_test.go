package service

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

func TestRunStateManagerTracksChildPhaseCounts(t *testing.T) {
	rsm, err := newRunStateManager(nil)
	require.NoError(t, err)

	updates, err := rsm.upsertActions(context.Background(), []*models.Action{
		testAction("parent", nil, common.ActionPhase_ACTION_PHASE_RUNNING, 1),
		testAction("child", stringPtr("parent"), common.ActionPhase_ACTION_PHASE_QUEUED, 2),
	})
	require.NoError(t, err)
	require.Len(t, updates, 2)

	parent := rsm.GetActionTreeNodeByName("parent")
	require.NotNil(t, parent)
	require.Equal(t, 1, parent.ChildPhaseCounts[common.ActionPhase_ACTION_PHASE_QUEUED])

	updates, err = rsm.upsertActions(context.Background(), []*models.Action{
		testAction("child", stringPtr("parent"), common.ActionPhase_ACTION_PHASE_SUCCEEDED, 2),
	})
	require.NoError(t, err)

	parent = rsm.GetActionTreeNodeByName("parent")
	require.Equal(t, 0, parent.ChildPhaseCounts[common.ActionPhase_ACTION_PHASE_QUEUED])
	require.Equal(t, 1, parent.ChildPhaseCounts[common.ActionPhase_ACTION_PHASE_SUCCEEDED])
}

func TestRunStateManagerTracksVisibilityFromFilters(t *testing.T) {
	rsm, err := newRunStateManager([]*common.Filter{
		{
			Field:    "PHASE",
			Function: common.Filter_VALUE_IN,
			Values:   []string{fmt.Sprintf("%d", common.ActionPhase_ACTION_PHASE_RUNNING)},
		},
	})
	require.NoError(t, err)

	updates, err := rsm.upsertActions(context.Background(), []*models.Action{
		testAction("parent", nil, common.ActionPhase_ACTION_PHASE_QUEUED, 1),
		testAction("child", stringPtr("parent"), common.ActionPhase_ACTION_PHASE_RUNNING, 2),
	})
	require.NoError(t, err)

	requireNodeUpdate(t, updates, "parent", true)
	requireNodeUpdate(t, updates, "child", true)

	updates, err = rsm.upsertActions(context.Background(), []*models.Action{
		testAction("child", stringPtr("parent"), common.ActionPhase_ACTION_PHASE_SUCCEEDED, 2),
	})
	require.NoError(t, err)

	requireNodeUpdate(t, updates, "child", false)
	requireNodeUpdate(t, updates, "parent", false)
}

func TestRunStateManagerSupportsNameFilter(t *testing.T) {
	rsm, err := newRunStateManager([]*common.Filter{
		{
			Field:    "NAME",
			Function: common.Filter_CONTAINS_CASE_INSENSITIVE,
			Values:   []string{"important"},
		},
	})
	require.NoError(t, err)

	updates, err := rsm.upsertActions(context.Background(), []*models.Action{
		testActionWithTask("a", nil, common.ActionPhase_ACTION_PHASE_QUEUED, 1, "pkg.important_task"),
		testActionWithTask("b", nil, common.ActionPhase_ACTION_PHASE_QUEUED, 2, "pkg.other_task"),
	})
	require.NoError(t, err)

	requireNodeUpdate(t, updates, "a", true)
	requireNoNodeUpdate(t, updates, "b")
}

func TestRunStateManagerErrorsWhenParentMissing(t *testing.T) {
	rsm, err := newRunStateManager(nil)
	require.NoError(t, err)

	_, err = rsm.upsertActions(context.Background(), []*models.Action{
		testAction("child", stringPtr("parent"), common.ActionPhase_ACTION_PHASE_QUEUED, 1),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "parent node [parent] not found")
	require.Nil(t, rsm.GetActionTreeNodeByName("child"))
}

// TestRunStateManagerSeedThenRestreamIsCountNeutral covers the invariant the
// aggregate-count fix relies on: child phase counts are seeded up front from a
// lightweight query (so the count is correct from the first streamed page), and
// then the full snapshot re-streams the same rows. Re-upserting a node with the
// same phase must be count-neutral, so the seeded total neither doubles nor
// regresses as the streaming loop replays it.
func TestRunStateManagerSeedThenRestreamIsCountNeutral(t *testing.T) {
	rsm, err := newRunStateManager(nil)
	require.NoError(t, err)
	ctx := context.Background()

	// root -> mapTask -> 150 QUEUED children, mirroring a large map task.
	const children = 150
	seed := []*models.Action{
		testAction("root", nil, common.ActionPhase_ACTION_PHASE_RUNNING, 1),
		testAction("mapTask", stringPtr("root"), common.ActionPhase_ACTION_PHASE_RUNNING, 2),
	}
	for i := 0; i < children; i++ {
		seed = append(seed, testAction(fmt.Sprintf("c%03d", i), stringPtr("mapTask"),
			common.ActionPhase_ACTION_PHASE_QUEUED, int64(10+i)))
	}

	// Seed (listAndSendAllActions discards these updates).
	_, err = rsm.upsertActions(ctx, seed)
	require.NoError(t, err)

	queued := common.ActionPhase_ACTION_PHASE_QUEUED
	require.Equal(t, children, rsm.GetActionTreeNodeByName("mapTask").ChildPhaseCounts[queued])
	// Transitive: root counts the mapTask node (RUNNING) plus all 150 grandchildren.
	require.Equal(t, children, rsm.GetActionTreeNodeByName("root").ChildPhaseCounts[queued])
	require.Equal(t, 1, rsm.GetActionTreeNodeByName("root").ChildPhaseCounts[common.ActionPhase_ACTION_PHASE_RUNNING])

	// Re-stream the same rows in pages (same phases) -> counts must not move.
	const pageSize = 100
	for off := 0; off < len(seed); off += pageSize {
		end := off + pageSize
		if end > len(seed) {
			end = len(seed)
		}
		_, err = rsm.upsertActions(ctx, seed[off:end])
		require.NoError(t, err)
		require.Equal(t, children, rsm.GetActionTreeNodeByName("mapTask").ChildPhaseCounts[queued],
			"re-streaming with the same phase must be count-neutral")
	}

	// A genuine live phase change after the snapshot still adjusts the count.
	_, err = rsm.upsertActions(ctx, []*models.Action{
		testAction("c000", stringPtr("mapTask"), common.ActionPhase_ACTION_PHASE_SUCCEEDED, 10),
	})
	require.NoError(t, err)
	require.Equal(t, children-1, rsm.GetActionTreeNodeByName("mapTask").ChildPhaseCounts[queued])
	require.Equal(t, 1, rsm.GetActionTreeNodeByName("mapTask").ChildPhaseCounts[common.ActionPhase_ACTION_PHASE_SUCCEEDED])
}

func testAction(name string, parent *string, phase common.ActionPhase, createdAtSec int64) *models.Action {
	return testActionWithTask(name, parent, phase, createdAtSec, "")
}

func testActionWithTask(name string, parent *string, phase common.ActionPhase, createdAtSec int64, taskName string) *models.Action {
	var parentNullStr sql.NullString
	if parent != nil {
		parentNullStr = sql.NullString{String: *parent, Valid: true}
	}
	action := &models.Action{
		Project:          "p",
		Domain:           "d",
		Name:             name,
		ParentActionName: parentNullStr,
		Phase:            int32(phase),
		CreatedAt:        time.Unix(createdAtSec, 0),
	}
	if taskName != "" {
		spec := &workflow.ActionSpec{
			Spec: &workflow.ActionSpec_Task{
				Task: &workflow.TaskAction{
					Id: &task.TaskIdentifier{Name: taskName},
				},
			},
		}
		action.ActionSpec, _ = proto.Marshal(spec)
	}
	return action
}

func stringPtr(s string) *string { return &s }

func requireNodeUpdate(t *testing.T, updates []*nodeUpdate, name string, meetsFilter bool) {
	t.Helper()
	for _, update := range updates {
		if update != nil && update.Node != nil && update.Node.Action != nil && update.Node.Action.Name == name {
			require.Equal(t, meetsFilter, update.MeetsFilter)
			return
		}
	}
	t.Fatalf("expected node update for %s", name)
}

func requireNoNodeUpdate(t *testing.T, updates []*nodeUpdate, name string) {
	t.Helper()
	for _, update := range updates {
		if update != nil && update.Node != nil && update.Node.Action != nil && update.Node.Action.Name == name {
			t.Fatalf("unexpected node update for %s", name)
		}
	}
}
