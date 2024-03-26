package start

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/handler"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
)

func init() {
	labeled.SetMetricKeys(contextutils.NodeIDKey)
}

func TestStartNodeHandler_Initialize(t *testing.T) {
	h := startHandler{}
	// Do nothing
	assert.NoError(t, h.Setup(context.TODO(), nil))
}

func TestStartNodeHandler_Handle(t *testing.T) {
	ctx := context.Background()
	h := New()
	t.Run("Any", func(t *testing.T) {
		s, err := h.Handle(ctx, nil)
		assert.NoError(t, err)
		assert.Equal(t, handler.EPhaseSuccess, s.Info().GetPhase())
	})
}

func TestEndHandler_Abort(t *testing.T) {
	e := New()
	assert.NoError(t, e.Abort(context.TODO(), nil, ""))
}

func TestEndHandler_Finalize(t *testing.T) {
	e := New()
	assert.NoError(t, e.Finalize(context.TODO(), nil))
}

func TestEndHandler_FinalizeRequired(t *testing.T) {
	e := New()
	assert.False(t, e.FinalizeRequired())
}
