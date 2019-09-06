package start

import (
	"context"
	"testing"

	"github.com/lyft/flytestdlib/contextutils"
	"github.com/lyft/flytestdlib/promutils/labeled"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
	"github.com/lyft/flytepropeller/pkg/utils"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/storage"
	"github.com/stretchr/testify/assert"
)

var testScope = promutils.NewScope("start_test")

func createInmemoryDataStore(t testing.TB, scope promutils.Scope) *storage.DataStore {
	cfg := storage.Config{
		Type: storage.TypeMemory,
	}
	d, err := storage.NewDataStore(&cfg, scope)
	assert.NoError(t, err)
	return d
}

func init() {
	labeled.SetMetricKeys(contextutils.NodeIDKey)
}

func TestStartNodeHandler_Initialize(t *testing.T) {
	h := startHandler{}
	// Do nothing
	assert.NoError(t, h.Initialize(context.TODO()))
}

func TestStartNodeHandler_StartNode(t *testing.T) {
	ctx := context.Background()
	mockStorage := createInmemoryDataStore(t, testScope.NewSubScope("z"))
	h := New(mockStorage)
	node := &v1alpha1.NodeSpec{
		ID: v1alpha1.EndNodeID,
	}
	w := &v1alpha1.FlyteWorkflow{
		WorkflowSpec: &v1alpha1.WorkflowSpec{
			ID: v1alpha1.WorkflowID("w1"),
		},
	}
	t.Run("NoInputs", func(t *testing.T) {
		s, err := h.StartNode(ctx, w, node, nil)
		assert.NoError(t, err)
		assert.Equal(t, handler.StatusSuccess, s)
	})
	t.Run("WithInputs", func(t *testing.T) {
		node := &v1alpha1.NodeSpec{
			ID: v1alpha1.NodeID("n1"),
			InputBindings: []*v1alpha1.Binding{
				{
					Binding: utils.MakeBinding("x", utils.MustMakePrimitiveBindingData("hello")),
				},
			},
		}
		s, err := h.StartNode(ctx, w, node, nil)
		assert.NoError(t, err)
		assert.Equal(t, handler.StatusSuccess, s)
	})
}

func TestStartNodeHandler_HandleNode(t *testing.T) {
	ctx := context.Background()
	h := startHandler{}
	s, err := h.CheckNodeStatus(ctx, nil, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, handler.StatusSuccess, s)
}
