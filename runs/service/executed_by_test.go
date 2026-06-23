package service

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

func TestActionMetadataFromModel_ExecutedBy(t *testing.T) {
	t.Run("subject resolved to subject-only identity", func(t *testing.T) {
		m := &models.Action{CreatedBy: sql.NullString{String: "00u1", Valid: true}}
		eb := actionMetadataFromModel(m).GetExecutedBy().GetUser()
		assert.Equal(t, "00u1", eb.GetId().GetSubject())
		// No user directory in the standalone service, so no profile is populated.
		assert.Nil(t, eb.GetSpec())
	})

	t.Run("null created_by yields nil executed_by", func(t *testing.T) {
		assert.Nil(t, actionMetadataFromModel(&models.Action{}).GetExecutedBy())
	})
}
