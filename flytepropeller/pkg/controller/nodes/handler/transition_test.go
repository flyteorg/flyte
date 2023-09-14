package handler

import (
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flytestdlib/storage"

	"github.com/stretchr/testify/assert"
)

func AsPointer[T any](val T) *T {
	return &val
}

func TestDoTransition(t *testing.T) {
	t.Run("ephemeral", func(t *testing.T) {
		tr := DoTransition(TransitionTypeEphemeral, PhaseInfoQueued("queued", &core.LiteralMap{}))
		assert.Equal(t, TransitionTypeEphemeral, tr.Type())
		assert.Equal(t, EPhaseQueued, tr.Info().p)
	})

	t.Run("barrier", func(t *testing.T) {
		tr := DoTransition(TransitionTypeBarrier, PhaseInfoSuccess(&ExecutionInfo{
			OutputInfo: &OutputInfo{OutputURI: "uri", DeckURI: AsPointer(storage.DataReference("deck"))},
		}))
		assert.Equal(t, TransitionTypeBarrier, tr.Type())
		assert.Equal(t, EPhaseSuccess, tr.Info().p)
		assert.Equal(t, storage.DataReference("uri"), tr.Info().GetInfo().OutputInfo.OutputURI)
		assert.Equal(t, AsPointer(storage.DataReference("deck")), tr.Info().GetInfo().OutputInfo.DeckURI)
	})
}

func TestTransition_WithInfo(t *testing.T) {
	tr := DoTransition(TransitionTypeEphemeral, PhaseInfoQueued("queued", &core.LiteralMap{}))
	assert.Equal(t, EPhaseQueued, tr.info.p)
	tr = tr.WithInfo(PhaseInfoSuccess(&ExecutionInfo{}))
	assert.Equal(t, EPhaseSuccess, tr.info.p)
}
