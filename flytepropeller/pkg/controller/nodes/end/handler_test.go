package end

import (
	"context"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	mocks2 "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/lyft/flytestdlib/contextutils"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/promutils/labeled"
	"github.com/lyft/flytestdlib/storage"
	regErrors "github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	mocks3 "github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/errors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler/mocks"
	"github.com/lyft/flytepropeller/pkg/utils"
	flyteassert "github.com/lyft/flytepropeller/pkg/utils/assert"
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
	storage.ComposedProtobufStore
	ReadProtobufCb  func(ctx context.Context, reference storage.DataReference, msg proto.Message) error
	WriteProtobufCb func(ctx context.Context, reference storage.DataReference, opts storage.Options, msg proto.Message) error
}

func (t TestProtoDataStore) ReadProtobuf(ctx context.Context, reference storage.DataReference, msg proto.Message) error {
	return t.ReadProtobufCb(ctx, reference, msg)
}

func (t TestProtoDataStore) WriteProtobuf(ctx context.Context, reference storage.DataReference, opts storage.Options, msg proto.Message) error {
	return t.WriteProtobufCb(ctx, reference, opts, msg)
}

func TestEndHandler_Setup(t *testing.T) {
	e := endHandler{}
	assert.NoError(t, e.Setup(context.TODO(), nil))
}

func TestEndHandler_Handle(t *testing.T) {
	inMem := createInmemoryDataStore(t, testScope.NewSubScope("x"))
	e := New()
	ctx := context.Background()

	inputs := &core.LiteralMap{
		Literals: map[string]*core.Literal{
			"x": utils.MustMakePrimitiveLiteral("hello"),
			"y": utils.MustMakePrimitiveLiteral("blah"),
		},
	}

	outputRef := v1alpha1.DataReference("testRef")

	createNodeCtx := func(inputs *core.LiteralMap, store *storage.DataStore) *mocks.NodeExecutionContext {
		ir := &mocks2.InputReader{}
		ir.On("Get", mock.Anything).Return(inputs, nil)
		nCtx := &mocks.NodeExecutionContext{}
		nCtx.On("InputReader").Return(ir)
		nCtx.On("DataStore").Return(store)
		ns := &mocks3.ExecutableNodeStatus{}
		ns.On("GetDataDir").Return(outputRef)
		ns.On("GetOutputDir").Return(outputRef)
		nCtx.On("NodeStatus").Return(ns)
		nCtx.On("NodeID").Return("end-node")
		return nCtx
	}

	t.Run("NoInputs", func(t *testing.T) {
		nCtx := createNodeCtx(nil, nil)
		s, err := e.Handle(ctx, nCtx)
		assert.NoError(t, err)
		assert.Equal(t, handler.EPhaseSuccess, s.Info().GetPhase())
	})

	outputLoc := v1alpha1.GetOutputsFile(outputRef)

	t.Run("WithInputs", func(t *testing.T) {
		nCtx := createNodeCtx(inputs, inMem)
		s, err := e.Handle(ctx, nCtx)
		assert.NoError(t, err)
		assert.Equal(t, handler.EPhaseSuccess, s.Info().GetPhase())
		actual := &core.LiteralMap{}
		if assert.NoError(t, inMem.ReadProtobuf(ctx, outputLoc, actual)) {
			flyteassert.EqualLiteralMap(t, inputs, actual)
		}
	})

	t.Run("StoreFailure", func(t *testing.T) {
		store := &storage.DataStore{
			ComposedProtobufStore: &TestProtoDataStore{
				WriteProtobufCb: func(ctx context.Context, reference v1alpha1.DataReference, opts storage.Options, msg proto.Message) error {
					return regErrors.Errorf("Fail")
				},
			},
		}
		nCtx := createNodeCtx(inputs, store)
		e := New()
		s, err := e.Handle(ctx, nCtx)
		assert.Error(t, err)
		assert.True(t, errors.Matches(err, errors.CausedByError))
		assert.Equal(t, handler.UnknownTransition, s)
	})
}
