package array

import (
	"context"
	"errors"
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/plugins"

	pluginErrors "github.com/flyteorg/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/catalog"
	catalogMocks "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/catalog/mocks"
	core2 "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	pluginMocks "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"
	ioMocks "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"
	arrayCore "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/core"

	"github.com/flyteorg/flytestdlib/bitarray"
	stdErrors "github.com/flyteorg/flytestdlib/errors"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/storage"

	"github.com/go-test/deep"

	structpb "github.com/golang/protobuf/ptypes/struct"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	dummyInputLiteral = &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Primitive{
					Primitive: &core.Primitive{
						Value: &core.Primitive_Integer{
							Integer: 3,
						},
					},
				},
			},
		},
	}
	singleInput = &core.LiteralMap{
		Literals: map[string]*core.Literal{
			"foo": {
				Value: &core.Literal_Collection{
					Collection: &core.LiteralCollection{
						Literals: []*core.Literal{
							dummyInputLiteral, dummyInputLiteral, dummyInputLiteral,
						},
					},
				},
			},
		},
	}
	multipleInputs = &core.LiteralMap{
		Literals: map[string]*core.Literal{
			"foo": {
				Value: &core.Literal_Collection{
					Collection: &core.LiteralCollection{
						Literals: []*core.Literal{
							dummyInputLiteral, dummyInputLiteral, dummyInputLiteral,
						},
					},
				},
			},
			"bar": {
				Value: &core.Literal_Collection{
					Collection: &core.LiteralCollection{
						Literals: []*core.Literal{
							dummyInputLiteral, dummyInputLiteral, dummyInputLiteral,
						},
					},
				},
			},
		},
	}
	multipleInputsInvalid = &core.LiteralMap{
		Literals: map[string]*core.Literal{
			"foo": {
				Value: &core.Literal_Collection{
					Collection: &core.LiteralCollection{
						Literals: []*core.Literal{
							dummyInputLiteral, dummyInputLiteral, dummyInputLiteral,
						},
					},
				},
			},
			"bar": {
				Value: &core.Literal_Collection{
					Collection: &core.LiteralCollection{
						Literals: []*core.Literal{
							dummyInputLiteral, dummyInputLiteral,
						},
					},
				},
			},
		},
	}
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
	inputs *core.LiteralMap, expectedState *arrayCore.State, maxArrayJobSize int64, expectedError error) {

	ctx := context.Background()

	tr := &pluginMocks.TaskReader{}
	tr.OnRead(ctx).Return(taskTemplate, nil)

	ds, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)

	cat := &catalogMocks.AsyncClient{}
	cat.OnDownloadMatch(mock.Anything, mock.Anything).Return(future, nil)

	ir := &ioMocks.InputReader{}
	ir.OnGetInputPrefixPath().Return("/prefix/")
	ir.On("Get", mock.Anything).Return(inputs, nil)

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

	got, err := DetermineDiscoverability(ctx, tCtx, maxArrayJobSize, state)
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
		runDetermineDiscoverabilityTest(t, template, f, singleInput, nil, 0, stdErrors.Errorf(pluginErrors.BadTaskSpecification, ""))
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

	t.Run("Run AWS Batch single job", func(t *testing.T) {
		toCache := arrayCore.InvertBitSet(bitarray.NewBitSet(1), 1)
		template.Type = AwsBatchTaskType
		runDetermineDiscoverabilityTest(t, template, f, singleInput, &arrayCore.State{
			CurrentPhase:         arrayCore.PhasePreLaunch,
			PhaseVersion:         core2.DefaultPhaseVersion,
			ExecutionArraySize:   1,
			OriginalArraySize:    1,
			OriginalMinSuccesses: 1,
			IndexesToCache:       toCache,
			Reason:               "Task is not discoverable.",
		}, 1, nil)
	})

	t.Run("Not discoverable", func(t *testing.T) {
		toCache := arrayCore.InvertBitSet(bitarray.NewBitSet(1), 1)

		runDetermineDiscoverabilityTest(t, template, f, singleInput, &arrayCore.State{
			CurrentPhase:         arrayCore.PhasePreLaunch,
			PhaseVersion:         core2.DefaultPhaseVersion,
			ExecutionArraySize:   1,
			OriginalArraySize:    1,
			OriginalMinSuccesses: 1,
			IndexesToCache:       toCache,
			Reason:               "Task is not discoverable.",
		}, 1, nil)
	})

	template.Metadata = &core.TaskMetadata{
		Discoverable:     true,
		DiscoveryVersion: "1",
	}

	t.Run("Discoverable but not cached", func(t *testing.T) {
		download.OnGetCachedResults().Return(bitarray.NewBitSet(1)).Once()
		toCache := bitarray.NewBitSet(1)
		toCache.Set(0)

		runDetermineDiscoverabilityTest(t, template, f, singleInput, &arrayCore.State{
			CurrentPhase:         arrayCore.PhasePreLaunch,
			PhaseVersion:         core2.DefaultPhaseVersion,
			ExecutionArraySize:   1,
			OriginalArraySize:    1,
			OriginalMinSuccesses: 1,
			IndexesToCache:       toCache,
			Reason:               "Finished cache lookup.",
		}, 1, nil)
	})

	t.Run("Discoverable and cached", func(t *testing.T) {
		cachedResults := bitarray.NewBitSet(1)
		cachedResults.Set(0)

		download.OnGetCachedResults().Return(cachedResults).Once()
		toCache := bitarray.NewBitSet(1)

		runDetermineDiscoverabilityTest(t, template, f, singleInput, &arrayCore.State{
			CurrentPhase:         arrayCore.PhasePreLaunch,
			PhaseVersion:         core2.DefaultPhaseVersion,
			ExecutionArraySize:   1,
			OriginalArraySize:    1,
			OriginalMinSuccesses: 1,
			IndexesToCache:       toCache,
			Reason:               "Finished cache lookup.",
		}, 1, nil)
	})

	t.Run("DiscoveryNotYetComplete ", func(t *testing.T) {
		future := &catalogMocks.DownloadFuture{}
		future.OnGetResponseStatus().Return(catalog.ResponseStatusNotReady)
		future.On("OnReady", mock.Anything).Return(func(_ context.Context, _ catalog.Future) {})

		runDetermineDiscoverabilityTest(t, template, future, singleInput, &arrayCore.State{
			CurrentPhase:         arrayCore.PhaseStart,
			PhaseVersion:         core2.DefaultPhaseVersion,
			OriginalArraySize:    1,
			OriginalMinSuccesses: 1,
		}, 1, nil)
	})

	t.Run("MaxArrayJobSizeFailure", func(t *testing.T) {
		runDetermineDiscoverabilityTest(t, template, f, singleInput, &arrayCore.State{
			CurrentPhase:         arrayCore.PhasePermanentFailure,
			PhaseVersion:         core2.DefaultPhaseVersion,
			OriginalArraySize:    1,
			OriginalMinSuccesses: 1,
			Reason:               "array size > max allowed. requested [1]. allowed [0]",
		}, 0, nil)
	})
}

func TestDiscoverabilityTaskType1(t *testing.T) {

	download := &catalogMocks.DownloadResponse{}
	download.OnGetCachedCount().Return(0)
	download.OnGetResultsSize().Return(1)

	f := &catalogMocks.DownloadFuture{}
	f.OnGetResponseStatus().Return(catalog.ResponseStatusReady)
	f.OnGetResponseError().Return(nil)
	f.OnGetResponse().Return(download, nil)

	arrayJob := &plugins.ArrayJob{
		SuccessCriteria: &plugins.ArrayJob_MinSuccessRatio{
			MinSuccessRatio: 0.5,
		},
	}
	var arrayJobCustom structpb.Struct
	err := utils.MarshalStruct(arrayJob, &arrayJobCustom)
	assert.NoError(t, err)
	templateType1 := &core.TaskTemplate{
		Id: &core.Identifier{
			ResourceType: core.ResourceType_TASK,
			Project:      "p",
			Domain:       "d",
			Name:         "n",
			Version:      "1",
		},
		Interface: &core.TypedInterface{
			Inputs: &core.VariableMap{Variables: map[string]*core.Variable{
				"foo": {
					Description: "foo",
				},
			}},
			Outputs: &core.VariableMap{Variables: map[string]*core.Variable{}},
		},
		Target: &core.TaskTemplate_Container{
			Container: &core.Container{
				Command: []string{"cmd"},
				Args:    []string{"{{$inputPrefix}}"},
				Image:   "img1",
			},
		},
		TaskTypeVersion: 1,
		Custom:          &arrayJobCustom,
	}

	t.Run("Not discoverable", func(t *testing.T) {
		download.OnGetCachedResults().Return(bitarray.NewBitSet(1)).Once()
		toCache := arrayCore.InvertBitSet(bitarray.NewBitSet(uint(3)), uint(3))

		runDetermineDiscoverabilityTest(t, templateType1, f, singleInput, &arrayCore.State{
			CurrentPhase:         arrayCore.PhasePreLaunch,
			PhaseVersion:         core2.DefaultPhaseVersion,
			ExecutionArraySize:   3,
			OriginalArraySize:    3,
			OriginalMinSuccesses: 2,
			IndexesToCache:       toCache,
			Reason:               "Task is not discoverable.",
		}, 3, nil)
	})

	t.Run("MultipleInputs", func(t *testing.T) {
		toCache := arrayCore.InvertBitSet(bitarray.NewBitSet(uint(3)), uint(3))

		runDetermineDiscoverabilityTest(t, templateType1, f, multipleInputs, &arrayCore.State{
			CurrentPhase:         arrayCore.PhasePreLaunch,
			PhaseVersion:         core2.DefaultPhaseVersion,
			ExecutionArraySize:   3,
			OriginalArraySize:    3,
			OriginalMinSuccesses: 2,
			IndexesToCache:       toCache,
			Reason:               "Task is not discoverable.",
		}, 3, nil)
	})

	t.Run("MultipleInputsInvalid", func(t *testing.T) {
		runDetermineDiscoverabilityTest(t, templateType1, f, multipleInputsInvalid, &arrayCore.State{
			CurrentPhase:         arrayCore.PhasePermanentFailure,
			PhaseVersion:         core2.DefaultPhaseVersion,
			ExecutionArraySize:   0,
			OriginalArraySize:    0,
			OriginalMinSuccesses: 0,
			IndexesToCache:       nil,
			Reason:               "all maptask input lists must be the same length",
		}, 3, nil)
	})
}
