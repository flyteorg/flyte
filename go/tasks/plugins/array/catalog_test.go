package array

import (
	"context"
	"errors"
	"testing"

	stdErrors "github.com/lyft/flytestdlib/errors"

	pluginErrors "github.com/lyft/flyteplugins/go/tasks/errors"

	"github.com/lyft/flytestdlib/bitarray"

	core2 "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core"

	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/catalog"

	"github.com/lyft/flytestdlib/promutils"
	"github.com/lyft/flytestdlib/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	catalogMocks "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/catalog/mocks"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/io"
	ioMocks "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/io/mocks"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"

	pluginMocks "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core/mocks"

	arrayCore "github.com/lyft/flyteplugins/go/tasks/plugins/array/core"

	"github.com/go-test/deep"
)

func TestNewLiteralScalarOfInteger(t *testing.T) {
	l := NewLiteralScalarOfInteger(int64(65))
	assert.Equal(t, int64(65), l.Value.(*core.Literal_Scalar).Scalar.Value.(*core.Scalar_Primitive).
		Primitive.Value.(*core.Primitive_Integer).Integer)
}

func TestCatalogBitsetToLiteralCollection(t *testing.T) {
	ba := bitarray.NewBitSet(3)
	ba.Set(1)
	lc := CatalogBitsetToLiteralCollection(ba, 3)
	assert.Equal(t, 2, len(lc.Literals))
	assert.Equal(t, int64(0), lc.Literals[0].Value.(*core.Literal_Scalar).Scalar.Value.(*core.Scalar_Primitive).
		Primitive.Value.(*core.Primitive_Integer).Integer)
	assert.Equal(t, int64(2), lc.Literals[1].Value.(*core.Literal_Scalar).Scalar.Value.(*core.Scalar_Primitive).
		Primitive.Value.(*core.Primitive_Integer).Integer)
}

func runDetermineDiscoverabilityTest(t testing.TB, taskTemplate *core.TaskTemplate, future catalog.DownloadFuture,
	expectedState *arrayCore.State, expectedError error) {

	ctx := context.Background()

	tr := &pluginMocks.TaskReader{}
	tr.OnRead(ctx).Return(taskTemplate, nil)

	ds, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)

	cat := &catalogMocks.AsyncClient{}
	cat.OnDownloadMatch(mock.Anything, mock.Anything).Return(future, nil)

	ir := &ioMocks.InputReader{}
	ir.OnGetInputPrefixPath().Return("/prefix/")

	ow := &ioMocks.OutputWriter{}
	ow.OnGetOutputPrefixPath().Return("/prefix/")
	ow.OnGetRawOutputPrefix().Return("/sandbox/")
	ow.OnGetOutputPath().Return("/prefix/outputs.pb")
	ow.On("Put", mock.Anything, mock.Anything).Return(func(ctx context.Context, or io.OutputReader) error {
		m, ee, err := or.Read(ctx)
		assert.NoError(t, err)
		assert.Nil(t, ee)
		assert.NotNil(t, m)

		assert.NoError(t, ds.WriteProtobuf(ctx, "/prefix/outputs.pb", storage.Options{}, m))

		return err
	})

	tCtx := &pluginMocks.TaskExecutionContext{}
	tCtx.OnTaskReader().Return(tr)
	tCtx.OnInputReader().Return(ir)
	tCtx.OnDataStore().Return(ds)
	tCtx.OnCatalog().Return(cat)
	tCtx.OnOutputWriter().Return(ow)
	tCtx.OnTaskRefreshIndicator().Return(func(ctx context.Context) {
		t.Log("Refresh called")
	})

	state := &arrayCore.State{
		CurrentPhase: arrayCore.PhaseStart,
	}

	got, err := DetermineDiscoverability(ctx, tCtx, state)
	if expectedError != nil {
		assert.Error(t, err)
		assert.True(t, errors.Is(err, expectedError))
	} else {
		assert.NoError(t, err)
		if diff := deep.Equal(expectedState, got); diff != nil {
			t.Error(diff)
		}
	}
}

func TestDetermineDiscoverability(t *testing.T) {
	var template *core.TaskTemplate

	download := &catalogMocks.DownloadResponse{}
	download.OnGetCachedCount().Return(0)
	download.OnGetResultsSize().Return(1)

	f := &catalogMocks.DownloadFuture{}
	f.OnGetResponseStatus().Return(catalog.ResponseStatusReady)
	f.OnGetResponseError().Return(nil)
	f.OnGetResponse().Return(download, nil)

	t.Run("Bad Task Spec", func(t *testing.T) {
		runDetermineDiscoverabilityTest(t, template, f, nil, stdErrors.Errorf(pluginErrors.BadTaskSpecification, ""))
	})

	template = &core.TaskTemplate{
		Id: &core.Identifier{
			ResourceType: core.ResourceType_TASK,
			Project:      "p",
			Domain:       "d",
			Name:         "n",
			Version:      "1",
		},
		Interface: &core.TypedInterface{
			Inputs:  &core.VariableMap{Variables: map[string]*core.Variable{}},
			Outputs: &core.VariableMap{Variables: map[string]*core.Variable{}},
		},
		Target: &core.TaskTemplate_Container{
			Container: &core.Container{
				Command: []string{"cmd"},
				Args:    []string{"{{$inputPrefix}}"},
				Image:   "img1",
			},
		},
	}

	t.Run("Not discoverable", func(t *testing.T) {
		toCache := arrayCore.InvertBitSet(bitarray.NewBitSet(1), 1)

		runDetermineDiscoverabilityTest(t, template, f, &arrayCore.State{
			CurrentPhase:         arrayCore.PhasePreLaunch,
			PhaseVersion:         core2.DefaultPhaseVersion,
			ExecutionArraySize:   1,
			OriginalArraySize:    1,
			OriginalMinSuccesses: 1,
			IndexesToCache:       toCache,
			Reason:               "Task is not discoverable.",
		}, nil)
	})

	template.Metadata = &core.TaskMetadata{
		Discoverable:     true,
		DiscoveryVersion: "1",
	}

	t.Run("Discoverable but not cached", func(t *testing.T) {
		download.OnGetCachedResults().Return(bitarray.NewBitSet(1)).Once()
		toCache := bitarray.NewBitSet(1)
		toCache.Set(0)

		runDetermineDiscoverabilityTest(t, template, f, &arrayCore.State{
			CurrentPhase:         arrayCore.PhasePreLaunch,
			PhaseVersion:         core2.DefaultPhaseVersion,
			ExecutionArraySize:   1,
			OriginalArraySize:    1,
			OriginalMinSuccesses: 1,
			IndexesToCache:       toCache,
			Reason:               "Finished cache lookup.",
		}, nil)
	})

	t.Run("Discoverable and cached", func(t *testing.T) {
		cachedResults := bitarray.NewBitSet(1)
		cachedResults.Set(0)

		download.OnGetCachedResults().Return(cachedResults).Once()
		toCache := bitarray.NewBitSet(1)

		runDetermineDiscoverabilityTest(t, template, f, &arrayCore.State{
			CurrentPhase:         arrayCore.PhasePreLaunch,
			PhaseVersion:         core2.DefaultPhaseVersion,
			ExecutionArraySize:   1,
			OriginalArraySize:    1,
			OriginalMinSuccesses: 1,
			IndexesToCache:       toCache,
			Reason:               "Finished cache lookup.",
		}, nil)
	})
}
