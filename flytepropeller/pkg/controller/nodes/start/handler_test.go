package start

import (
	"context"
	"testing"

	"github.com/lyft/flytestdlib/contextutils"
	"github.com/lyft/flytestdlib/promutils/labeled"

	"github.com/stretchr/testify/assert"

	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
)

func init() {
	labeled.SetMetricKeys(contextutils.NodeIDKey)
}

func TestStartNodeHandler_Initialize(t *testing.T) {
	h := startHandler{}
	// Do nothing
	assert.NoError(t, h.Setup(context.TODO(), nil))
}

func TestStartNodeHandler_StartNode(t *testing.T) {
	ctx := context.Background()
	h := New()
	t.Run("Any", func(t *testing.T) {
		s, err := h.Handle(ctx, nil)
		assert.NoError(t, err)
		assert.Equal(t, handler.EPhaseSuccess, s.Info().GetPhase())
	})
}
