package end

import (
	"context"
	"testing"

	"github.com/lyft/flytestdlib/contextutils"
	"github.com/lyft/flytestdlib/promutils/labeled"

	"github.com/golang/protobuf/proto"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/errors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
	"github.com/lyft/flytepropeller/pkg/utils"
	flyteassert "github.com/lyft/flytepropeller/pkg/utils/assert"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/storage"
	regErrors "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var testScope = promutils.NewScope("end_test")

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

type TestProtoDataStore struct {
	ReadProtobufCb  func(ctx context.Context, reference storage.DataReference, msg proto.Message) error
	WriteProtobufCb func(ctx context.Context, reference storage.DataReference, opts storage.Options, msg proto.Message) error
}

func (t TestProtoDataStore) ReadProtobuf(ctx context.Context, reference storage.DataReference, msg proto.Message) error {
	return t.ReadProtobufCb(ctx, reference, msg)
}

func (t TestProtoDataStore) WriteProtobuf(ctx context.Context, reference storage.DataReference, opts storage.Options, msg proto.Message) error {
	return t.WriteProtobufCb(ctx, reference, opts, msg)
}

func TestEndHandler_CheckNodeStatus(t *testing.T) {
	e := endHandler{}
	s, err := e.CheckNodeStatus(context.TODO(), nil, nil, nil)
	assert.NoError(t, err)
	assert.Equal(t, handler.StatusSuccess, s)
}

func TestEndHandler_HandleFailingNode(t *testing.T) {
	e := endHandler{}
	node := &v1alpha1.NodeSpec{
		ID: v1alpha1.EndNodeID,
	}
	w := &v1alpha1.FlyteWorkflow{
		WorkflowSpec: &v1alpha1.WorkflowSpec{
			ID: v1alpha1.WorkflowID("w1"),
		},
	}
	s, err := e.HandleFailingNode(context.TODO(), w, node)
	assert.NoError(t, err)
	assert.Equal(t, errors.IllegalStateError, s.Err.(*errors.NodeError).Code)
}

func TestEndHandler_Initialize(t *testing.T) {
	e := endHandler{}
	assert.NoError(t, e.Initialize(context.TODO()))
}

func TestEndHandler_StartNode(t *testing.T) {
	inMem := createInmemoryDataStore(t, testScope.NewSubScope("x"))
	e := New(inMem)
	ctx := context.Background()

	inputs := &core.LiteralMap{
		Literals: map[string]*core.Literal{
			"x": utils.MustMakePrimitiveLiteral("hello"),
			"y": utils.MustMakePrimitiveLiteral("blah"),
		},
	}

	outputRef := v1alpha1.DataReference("testRef")

	node := &v1alpha1.NodeSpec{
		ID: v1alpha1.EndNodeID,
	}
	w := &v1alpha1.FlyteWorkflow{
		WorkflowSpec: &v1alpha1.WorkflowSpec{
			ID: v1alpha1.WorkflowID("w1"),
		},
	}
	w.Status.NodeStatus = map[v1alpha1.NodeID]*v1alpha1.NodeStatus{
		v1alpha1.EndNodeID: {
			DataDir: outputRef,
		},
	}

	t.Run("NoInputs", func(t *testing.T) {
		s, err := e.StartNode(ctx, w, node, nil)
		assert.NoError(t, err)
		assert.Equal(t, handler.StatusSuccess, s)
	})

	outputLoc := v1alpha1.GetOutputsFile(outputRef)
	t.Run("WithInputs", func(t *testing.T) {
		s, err := e.StartNode(ctx, w, node, inputs)
		assert.NoError(t, err)
		assert.Equal(t, handler.StatusSuccess, s)
		actual := &core.LiteralMap{}
		if assert.NoError(t, inMem.ReadProtobuf(ctx, outputLoc, actual)) {
			flyteassert.EqualLiteralMap(t, inputs, actual)
		}
	})

	t.Run("StoreFailure", func(t *testing.T) {
		store := &TestProtoDataStore{
			WriteProtobufCb: func(ctx context.Context, reference v1alpha1.DataReference, opts storage.Options, msg proto.Message) error {
				return regErrors.Errorf("Fail")
			},
		}
		e := New(store)
		s, err := e.StartNode(ctx, w, node, inputs)
		assert.Error(t, err)
		assert.True(t, errors.Matches(err, errors.CausedByError))
		assert.Equal(t, handler.StatusUndefined, s)
	})
}
