package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

func TestRunStateManagerUpsertActions(t *testing.T) {
	t.Run("stores initial batch", func(t *testing.T) {
		rsm := newRunStateManager()

		actions := []*models.Action{
			{Name: "a", Phase: int32(common.ActionPhase_ACTION_PHASE_QUEUED)},
			{Name: "b", Phase: int32(common.ActionPhase_ACTION_PHASE_RUNNING)},
		}

		changed := rsm.upsertActions(actions)

		require.Len(t, changed, 2)
		require.Equal(t, int32(common.ActionPhase_ACTION_PHASE_QUEUED), rsm.actions["a"].Phase)
		require.Equal(t, int32(common.ActionPhase_ACTION_PHASE_RUNNING), rsm.actions["b"].Phase)
	})

	t.Run("dedupes unchanged updates", func(t *testing.T) {
		rsm := newRunStateManager()
		action := &models.Action{Name: "a", Phase: int32(common.ActionPhase_ACTION_PHASE_QUEUED)}

		require.Len(t, rsm.upsertActions([]*models.Action{action}), 1)
		require.Empty(t, rsm.upsertActions([]*models.Action{action}))
	})

	t.Run("keeps only latest update for the same action in one batch", func(t *testing.T) {
		rsm := newRunStateManager()

		changed := rsm.upsertActions([]*models.Action{
			{Name: "a", Phase: int32(common.ActionPhase_ACTION_PHASE_QUEUED)},
			{Name: "a", Phase: int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED)},
		})

		require.Len(t, changed, 1)
		require.Equal(t, int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED), changed[0].Phase)
		require.Equal(t, int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED), rsm.actions["a"].Phase)
	})
}
