package service

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

func TestConvertActionToEnrichedProto_AllFields(t *testing.T) {
	svc := &RunService{}
	now := time.Now().UTC().Truncate(time.Second)
	endTime := now.Add(5 * time.Second)
	parent := "parent-action"

	action := &models.Action{
		Org:              "my-org",
		Project:          "my-project",
		Domain:           "production",
		Name:             "child-action",
		ParentActionName: &parent,
		Phase:            int32(common.ActionPhase_ACTION_PHASE_SUCCEEDED),
		ActionSpec: datatypes.JSON(mustJSON(t, map[string]interface{}{
			"action_id": map[string]interface{}{
				"run":  map[string]interface{}{"org": "my-org", "project": "my-project", "domain": "production", "name": "my-run"},
				"name": "child-action",
			},
		})),
		ActionDetails: datatypes.JSON([]byte(`{}`)),
		CreatedAt:     now,
		UpdatedAt:     endTime,
		EndedAt:       sql.NullTime{Time: endTime, Valid: true},
	}

	ea := svc.convertActionToEnrichedProto(action)
	require.NotNil(t, ea)
	require.NotNil(t, ea.Action)
	require.NotNil(t, ea.Action.Id)
	require.NotNil(t, ea.Action.Id.Run)

	// run.name should be populated from ActionSpec
	assert.Equal(t, "my-run", ea.Action.Id.Run.Name)
	assert.Equal(t, "my-org", ea.Action.Id.Run.Org)
	assert.Equal(t, "my-project", ea.Action.Id.Run.Project)
	assert.Equal(t, "production", ea.Action.Id.Run.Domain)
	assert.Equal(t, "child-action", ea.Action.Id.Name)

	// status fields
	require.NotNil(t, ea.Action.Status)
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_SUCCEEDED, ea.Action.Status.Phase)
	assert.Equal(t, now.Unix(), ea.Action.Status.StartTime.AsTime().Unix())
	require.NotNil(t, ea.Action.Status.EndTime)
	assert.Equal(t, endTime.Unix(), ea.Action.Status.EndTime.AsTime().Unix())
	assert.Equal(t, uint32(1), ea.Action.Status.Attempts)
	require.NotNil(t, ea.Action.Status.DurationMs)
	assert.Equal(t, uint64(5000), *ea.Action.Status.DurationMs)

	// metadata
	require.NotNil(t, ea.Action.Metadata)
	assert.Equal(t, "parent-action", ea.Action.Metadata.Parent)

	assert.True(t, ea.MeetsFilter)
}

func TestConvertActionToEnrichedProto_NoEndTime(t *testing.T) {
	svc := &RunService{}
	now := time.Now().UTC().Truncate(time.Second)

	action := &models.Action{
		Org:           "org",
		Project:       "proj",
		Domain:        "dev",
		Name:          "root-action",
		Phase:         int32(common.ActionPhase_ACTION_PHASE_RUNNING),
		ActionSpec:    datatypes.JSON([]byte(`{}`)),
		ActionDetails: datatypes.JSON([]byte(`{}`)),
		CreatedAt:     now,
		EndedAt:       sql.NullTime{Valid: false},
	}

	ea := svc.convertActionToEnrichedProto(action)
	require.NotNil(t, ea)
	require.NotNil(t, ea.Action.Status)

	assert.Equal(t, common.ActionPhase_ACTION_PHASE_RUNNING, ea.Action.Status.Phase)
	assert.NotNil(t, ea.Action.Status.StartTime)
	assert.Nil(t, ea.Action.Status.EndTime)
	assert.Nil(t, ea.Action.Status.DurationMs)
	assert.Equal(t, uint32(1), ea.Action.Status.Attempts)
}

func TestConvertActionToEnrichedProto_RootAction_RunName(t *testing.T) {
	svc := &RunService{}

	action := &models.Action{
		Org:           "org",
		Project:       "proj",
		Domain:        "dev",
		Name:          "run-123",
		Phase:         int32(common.ActionPhase_ACTION_PHASE_QUEUED),
		ActionSpec:    datatypes.JSON([]byte(`{}`)),
		ActionDetails: datatypes.JSON([]byte(`{}`)),
		CreatedAt:     time.Now(),
		EndedAt:       sql.NullTime{Valid: false},
	}

	ea := svc.convertActionToEnrichedProto(action)
	// Root action: run name = action name
	assert.Equal(t, "run-123", ea.Action.Id.Run.Name)
}

func TestExtractCacheStatus(t *testing.T) {
	t.Run("empty details", func(t *testing.T) {
		assert.Equal(t, core.CatalogCacheStatus_CACHE_DISABLED, extractCacheStatus(nil))
		assert.Equal(t, core.CatalogCacheStatus_CACHE_DISABLED, extractCacheStatus([]byte{}))
	})

	t.Run("with cache status", func(t *testing.T) {
		details := mustJSON(t, map[string]interface{}{
			"cache_status": int32(core.CatalogCacheStatus_CACHE_HIT),
		})
		assert.Equal(t, core.CatalogCacheStatus_CACHE_HIT, extractCacheStatus(details))
	})
}

func mustJSON(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}
