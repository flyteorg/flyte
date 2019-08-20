package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTaskStatus_CopyConstructor(t *testing.T) {
	t.Run("WithOccurredAt", func(t *testing.T) {
		input := TaskStatusSucceeded
		assert.True(t, input.OccurredAt.IsZero())
		actual := input.WithOccurredAt(time.Now())
		assert.True(t, input.OccurredAt.IsZero())
		assert.False(t, actual.OccurredAt.IsZero())
	})

	t.Run("WithPhaseVersion", func(t *testing.T) {
		input := TaskStatusSucceeded
		assert.Zero(t, input.PhaseVersion)
		actual := input.WithPhaseVersion(4)
		assert.Zero(t, input.PhaseVersion)
		assert.NotZero(t, actual.PhaseVersion)
	})

	t.Run("WithState", func(t *testing.T) {
		input := TaskStatusSucceeded
		assert.Nil(t, input.State)
		actual := input.WithState(map[string]interface{}{"hello": "world"})
		assert.Nil(t, input.State)
		assert.NotNil(t, actual.State)
	})
}
