package array

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/go-test/deep"

	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array/arraystatus"

	arrayCore "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/core"

	mocks3 "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"

	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"

	"github.com/flyteorg/flyteidl/clients/go/coreutils"

	"github.com/flyteorg/flytestdlib/bitarray"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"
	mocks2 "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io/mocks"

	"github.com/flyteorg/flytestdlib/storage"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/workqueue/mocks"
	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	pluginCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/workqueue"
	"github.com/flyteorg/flytestdlib/promutils"
)

func TestOutputAssembler_Queue(t *testing.T) {
	type args struct {
		id   workqueue.WorkItemID
		item *outputAssembleItem
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"nil item", args{"id", nil}, false},
		{"valid", args{"id", &outputAssembleItem{}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &mocks.IndexedWorkQueue{}
			q.OnQueueMatch(mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

			o := OutputAssembler{
				IndexedWorkQueue: q,
			}
			if err := o.Queue(context.TODO(), tt.args.id, tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("OutputAssembler.Queue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func init() {
	labeled.SetMetricKeys(contextutils.NamespaceKey)
}

func Test_assembleOutputsWorker_Process(t *testing.T) {
	ctx := context.Background()

	t.Run("EmptyInputs", func(t *testing.T) {
		memStore, err := storage.NewDataStore(&storage.Config{
			Type: storage.TypeMemory,
		}, promutils.NewTestScope())
		assert.NoError(t, err)

		// Setup the expected data to be written to outputWriter.
		ow := &mocks2.OutputWriter{}
		ow.OnGetOutputPrefixPath().Return("/bucket/prefix")
		ow.OnGetOutputPath().Return("/bucket/prefix/outputs.pb")
		ow.OnGetRawOutputPrefix().Return("/bucket/sandbox/")

		// Setup the input phases that inform outputs worker about which tasks failed/succeeded.
		phases := arrayCore.NewPhasesCompactArray(0)

		item := &outputAssembleItem{
			outputPaths:    ow,
			varNames:       []string{"var1", "var2"},
			finalPhases:    phases,
			dataStore:      memStore,
			isAwsSingleJob: false,
		}

		w := assembleOutputsWorker{}
		actual, err := w.Process(ctx, item)
		assert.NoError(t, err)
		assert.Equal(t, workqueue.WorkStatusSucceeded, actual)

		actualOutputs := &core.LiteralMap{}
		assert.NoError(t, memStore.ReadProtobuf(ctx, "/bucket/prefix/outputs.pb", actualOutputs))
		expected := coreutils.MustMakeLiteral(map[string]interface{}{
			"var1": []interface{}{},
			"var2": []interface{}{},
		}).GetMap()

		expectedBytes, err := json.Marshal(expected)
		assert.NoError(t, err)

		actualBytes, err := json.Marshal(actualOutputs)
		assert.NoError(t, err)

		if diff := deep.Equal(string(actualBytes), string(expectedBytes)); diff != nil {
			assert.FailNow(t, "Should be equal.", "Diff: %v", diff)
		}
	})

	t.Run("MissingTasks", func(t *testing.T) {
		memStore, err := storage.NewDataStore(&storage.Config{
			Type: storage.TypeMemory,
		}, promutils.NewTestScope())
		assert.NoError(t, err)

		// Write data to 1st and 3rd tasks only. Simulate a failed 2nd and 4th tasks.
		l := coreutils.MustMakeLiteral(map[string]interface{}{
			"var1": 5,
			"var2": "hello world",
		})
		assert.NoError(t, memStore.WriteProtobuf(ctx, "/bucket/prefix/0/outputs.pb", storage.Options{}, l.GetMap()))
		assert.NoError(t, memStore.WriteProtobuf(ctx, "/bucket/prefix/2/outputs.pb", storage.Options{}, l.GetMap()))

		// Setup the expected data to be written to outputWriter.
		ow := &mocks2.OutputWriter{}
		ow.OnGetOutputPrefixPath().Return("/bucket/prefix")
		ow.OnGetOutputPath().Return("/bucket/prefix/outputs.pb")
		ow.OnGetRawOutputPrefix().Return("/bucket/sandbox/")

		// Setup the input phases that inform outputs worker about which tasks failed/succeeded.
		phases := arrayCore.NewPhasesCompactArray(4)
		phases.SetItem(0, bitarray.Item(pluginCore.PhaseSuccess))
		phases.SetItem(1, bitarray.Item(pluginCore.PhasePermanentFailure))
		phases.SetItem(2, bitarray.Item(pluginCore.PhaseSuccess))
		phases.SetItem(3, bitarray.Item(pluginCore.PhasePermanentFailure))

		item := &outputAssembleItem{
			outputPaths:    ow,
			varNames:       []string{"var1", "var2"},
			finalPhases:    phases,
			dataStore:      memStore,
			isAwsSingleJob: false,
		}

		w := assembleOutputsWorker{}
		actual, err := w.Process(ctx, item)
		assert.NoError(t, err)
		assert.Equal(t, workqueue.WorkStatusSucceeded, actual)

		actualOutputs := &core.LiteralMap{}
		assert.NoError(t, memStore.ReadProtobuf(ctx, "/bucket/prefix/outputs.pb", actualOutputs))
		// Since 2nd and 4th tasks failed, there should be nil literals in their expected places.
		expected := coreutils.MustMakeLiteral(map[string]interface{}{
			"var1": []interface{}{5, nil, 5, nil},
			"var2": []interface{}{"hello world", nil, "hello world", nil},
		}).GetMap()

		expectedBytes, err := json.Marshal(expected)
		assert.NoError(t, err)

		actualBytes, err := json.Marshal(actualOutputs)
		assert.NoError(t, err)

		if diff := deep.Equal(string(actualBytes), string(expectedBytes)); diff != nil {
			assert.FailNow(t, "Should be equal.", "Diff: %v", diff)
		}
	})
}

func Test_appendSubTaskOutput(t *testing.T) {
	nativeMap := map[string]interface{}{
		"var1": 5,
		"var2": "hello",
	}
	validOutputs := coreutils.MustMakeLiteral(nativeMap).GetMap()

	t.Run("append to empty", func(t *testing.T) {
		expected := map[string]interface{}{
			"var1": []interface{}{coreutils.MustMakeLiteral(5)},
			"var2": []interface{}{coreutils.MustMakeLiteral("hello")},
		}

		actual := &core.LiteralMap{
			Literals: map[string]*core.Literal{},
		}
		appendSubTaskOutput(actual, validOutputs, 1)
		assert.Equal(t, actual, coreutils.MustMakeLiteral(expected).GetMap())
	})

	t.Run("append to existing", func(t *testing.T) {
		expected := coreutils.MustMakeLiteral(map[string]interface{}{
			"var1": []interface{}{nilLiteral, coreutils.MustMakeLiteral(5)},
			"var2": []interface{}{nilLiteral, coreutils.MustMakeLiteral("hello")},
		}).GetMap()

		actual := coreutils.MustMakeLiteral(map[string]interface{}{
			"var1": []interface{}{nilLiteral},
			"var2": []interface{}{nilLiteral},
		}).GetMap()

		appendSubTaskOutput(actual, validOutputs, 1)
		assert.Equal(t, actual, expected)
	})
}

func TestAssembleFinalOutputs(t *testing.T) {
	ctx := context.Background()
	t.Run("Found succeeded", func(t *testing.T) {
		q := &mocks.IndexedWorkQueue{}
		called := false
		q.On("Queue", mock.Anything, mock.Anything, mock.Anything).Return(
			func(ctx context.Context, id workqueue.WorkItemID, workItem workqueue.WorkItem) error {
				i, casted := workItem.(*outputAssembleItem)
				assert.True(t, casted)
				assert.Equal(t, []string{"var1"}, i.varNames)
				called = true
				return nil
			})
		assemblyQueue := OutputAssembler{
			IndexedWorkQueue: q,
		}

		info := &mocks.WorkItemInfo{}
		info.OnStatus().Return(workqueue.WorkStatusSucceeded)
		q.OnGet("found").Return(info, true, nil)

		s := &arrayCore.State{}

		tID := &mocks3.TaskExecutionID{}
		tID.OnGetGeneratedName().Return("found")

		tMeta := &mocks3.TaskExecutionMetadata{}
		tMeta.OnGetTaskExecutionID().Return(tID)

		ow := &mocks2.OutputWriter{}
		ow.OnPutMatch(mock.Anything, mock.Anything).Return(nil)
		ow.OnGetOutputPath().Return("/location/prefix/outputs.pb")
		ow.OnGetErrorPath().Return("/location/prefix/error.pb")

		d, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
		assert.NoError(t, err)

		tCtx := &mocks3.TaskExecutionContext{}
		tCtx.OnTaskExecutionMetadata().Return(tMeta)
		tCtx.OnOutputWriter().Return(ow)
		tCtx.OnMaxDatasetSizeBytes().Return(10000)
		tCtx.OnDataStore().Return(d)

		_, err = AssembleFinalOutputs(ctx, assemblyQueue, tCtx, arrayCore.PhaseSuccess, 1, s)
		assert.NoError(t, err)
		assert.Equal(t, arrayCore.PhaseSuccess, s.CurrentPhase)
		assert.Equal(t, uint32(1), s.PhaseVersion)
		assert.False(t, called)
	})

	t.Run("Found failed", func(t *testing.T) {
		q := &mocks.IndexedWorkQueue{}
		called := false
		q.On("Queue", mock.Anything, mock.Anything, mock.Anything).Return(
			func(ctx context.Context, id workqueue.WorkItemID, workItem workqueue.WorkItem) error {
				i, casted := workItem.(*outputAssembleItem)
				assert.True(t, casted)
				assert.Equal(t, []string{"var1"}, i.varNames)
				called = true
				return nil
			})
		assemblyQueue := OutputAssembler{
			IndexedWorkQueue: q,
		}

		info := &mocks.WorkItemInfo{}
		info.OnStatus().Return(workqueue.WorkStatusFailed)
		info.OnError().Return(fmt.Errorf("expected error"))

		q.OnGet("found_failed").Return(info, true, nil)

		s := &arrayCore.State{}

		tID := &mocks3.TaskExecutionID{}
		tID.OnGetGeneratedName().Return("found_failed")

		tMeta := &mocks3.TaskExecutionMetadata{}
		tMeta.OnGetTaskExecutionID().Return(tID)

		tCtx := &mocks3.TaskExecutionContext{}
		tCtx.OnTaskExecutionMetadata().Return(tMeta)

		_, err := AssembleFinalOutputs(ctx, assemblyQueue, tCtx, arrayCore.PhaseSuccess, 1, s)
		assert.NoError(t, err)
		assert.Equal(t, arrayCore.PhaseRetryableFailure, s.CurrentPhase)
		assert.Equal(t, uint32(0), s.PhaseVersion)
		assert.False(t, called)
	})

	t.Run("Not Found Queued then Succeeded", func(t *testing.T) {
		q := &mocks.IndexedWorkQueue{}
		called := false
		q.On("Queue", mock.Anything, mock.Anything, mock.Anything).Return(
			func(ctx context.Context, id workqueue.WorkItemID, workItem workqueue.WorkItem) error {
				i, casted := workItem.(*outputAssembleItem)
				assert.True(t, casted)
				assert.Equal(t, []string{"var1"}, i.varNames)
				called = true
				return nil
			})
		assemblyQueue := OutputAssembler{
			IndexedWorkQueue: q,
		}

		info := &mocks.WorkItemInfo{}
		info.OnStatus().Return(workqueue.WorkStatusSucceeded)
		q.OnGet("notfound").Return(nil, false, nil).Once()
		q.OnGet("notfound").Return(info, true, nil).Once()

		detailedStatus := arrayCore.NewPhasesCompactArray(2)
		detailedStatus.SetItem(0, bitarray.Item(pluginCore.PhaseSuccess))
		detailedStatus.SetItem(1, bitarray.Item(pluginCore.PhaseSuccess))

		s := &arrayCore.State{
			ArrayStatus: arraystatus.ArrayStatus{
				Detailed: detailedStatus,
			},
			IndexesToCache:       arrayCore.InvertBitSet(bitarray.NewBitSet(2), 2),
			OriginalArraySize:    2,
			ExecutionArraySize:   2,
			OriginalMinSuccesses: 2,
		}

		tID := &mocks3.TaskExecutionID{}
		tID.OnGetGeneratedName().Return("notfound")

		tMeta := &mocks3.TaskExecutionMetadata{}
		tMeta.OnGetTaskExecutionID().Return(tID)

		tReader := &mocks3.TaskReader{}
		tReader.OnReadMatch(mock.Anything).Return(&core.TaskTemplate{
			Interface: &core.TypedInterface{
				Outputs: &core.VariableMap{
					Variables: map[string]*core.Variable{"var1": {Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}}}},
				},
			},
		}, nil)

		ds, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
		assert.NoError(t, err)

		taskOutput := coreutils.MustMakeLiteral(map[string]interface{}{
			"var1": []interface{}{5, 2},
		})

		assert.NoError(t, ds.WriteProtobuf(ctx, "/prefix/outputs.pb", storage.Options{}, taskOutput.GetMap()))

		ow := &mocks2.OutputWriter{}
		ow.OnGetOutputPrefixPath().Return("/prefix/")
		ow.OnGetOutputPath().Return("/prefix/outputs.pb")
		ow.On("Put", mock.Anything, mock.Anything).Return(func(ctx context.Context, or io.OutputReader) error {
			m, ee, err := or.Read(ctx)
			assert.NoError(t, err)
			assert.Nil(t, ee)
			assert.NotNil(t, m)

			return err
		})

		tCtx := &mocks3.TaskExecutionContext{}
		tCtx.OnTaskExecutionMetadata().Return(tMeta)
		tCtx.OnTaskReader().Return(tReader)
		tCtx.OnOutputWriter().Return(ow)
		tCtx.OnDataStore().Return(ds)
		tCtx.OnMaxDatasetSizeBytes().Return(10000)

		_, err = AssembleFinalOutputs(ctx, assemblyQueue, tCtx, arrayCore.PhaseSuccess, 1, s)
		assert.NoError(t, err)
		assert.Equal(t, arrayCore.PhaseSuccess, s.CurrentPhase)
		assert.Equal(t, uint32(1), s.PhaseVersion)
		assert.True(t, called)
	})
}

func Test_assembleErrorsWorker_Process(t *testing.T) {
	ctx := context.Background()

	memStore, err := storage.NewDataStore(&storage.Config{
		Type: storage.TypeMemory,
	}, promutils.NewTestScope())
	assert.NoError(t, err)

	// Write data to 1st and 3rd tasks only. Simulate a failed 2nd and 4th tasks.
	l := coreutils.MustMakeLiteral(map[string]interface{}{
		"var1": 5,
		"var2": "hello world",
	})

	ee := &core.ErrorDocument{
		Error: &core.ContainerError{
			Message: "Expected error",
		},
	}

	assert.NoError(t, memStore.WriteProtobuf(ctx, "/bucket/prefix/0/outputs.pb", storage.Options{}, l.GetMap()))
	assert.NoError(t, memStore.WriteProtobuf(ctx, "/bucket/prefix/1/error.pb", storage.Options{}, ee))
	assert.NoError(t, memStore.WriteProtobuf(ctx, "/bucket/prefix/2/outputs.pb", storage.Options{}, l.GetMap()))
	assert.NoError(t, memStore.WriteProtobuf(ctx, "/bucket/prefix/3/error.pb", storage.Options{}, ee))

	// Setup the expected data to be written to outputWriter.
	ow := &mocks2.OutputWriter{}
	ow.OnGetRawOutputPrefix().Return("/bucket/sandbox/")
	ow.OnGetOutputPrefixPath().Return("/bucket/prefix")
	ow.OnGetErrorPath().Return("/bucket/prefix/error.pb")
	ow.On("Put", mock.Anything, mock.Anything).Return(func(ctx context.Context, reader io.OutputReader) error {
		// Since 2nd and 4th tasks failed, there should be nil literals in their expected places.

		final, ee, err := reader.Read(ctx)
		assert.NoError(t, err)
		assert.Nil(t, final)
		assert.NotNil(t, ee)
		assert.Equal(t, `[1][3]: message:"Expected error" 
`, ee.Message)

		return nil
	}).Once()

	// Setup the input phases that inform outputs worker about which tasks failed/succeeded.
	phases := arrayCore.NewPhasesCompactArray(4)
	phases.SetItem(0, bitarray.Item(pluginCore.PhaseSuccess))
	phases.SetItem(1, bitarray.Item(pluginCore.PhasePermanentFailure))
	phases.SetItem(2, bitarray.Item(pluginCore.PhaseSuccess))
	phases.SetItem(3, bitarray.Item(pluginCore.PhasePermanentFailure))

	item := &outputAssembleItem{
		varNames:       []string{"var1", "var2"},
		finalPhases:    phases,
		outputPaths:    ow,
		dataStore:      memStore,
		isAwsSingleJob: false,
	}

	w := assembleErrorsWorker{
		maxErrorMessageLength: 1000,
	}
	actual, err := w.Process(ctx, item)
	assert.NoError(t, err)
	assert.Equal(t, workqueue.WorkStatusSucceeded, actual)

	item.isAwsSingleJob = true
	actual, err = w.Process(ctx, item)
	assert.NoError(t, err)
	assert.Equal(t, workqueue.WorkStatusSucceeded, actual)
}

func TestNewOutputAssembler(t *testing.T) {
	t.Run("Invalid Config", func(t *testing.T) {
		_, err := NewOutputAssembler(workqueue.Config{
			Workers: 1,
		}, promutils.NewTestScope())
		assert.Error(t, err)
	})

	t.Run("Valid Config", func(t *testing.T) {
		o, err := NewOutputAssembler(workqueue.Config{
			Workers:            1,
			IndexCacheMaxItems: 10,
		}, promutils.NewTestScope())
		assert.NoError(t, err)
		assert.NotNil(t, o)
	})
}

func TestNewErrorAssembler(t *testing.T) {
	t.Run("Invalid Config", func(t *testing.T) {
		_, err := NewErrorAssembler(0, workqueue.Config{
			Workers: 1,
		}, promutils.NewTestScope())
		assert.Error(t, err)
	})

	t.Run("Valid Config", func(t *testing.T) {
		o, err := NewErrorAssembler(0, workqueue.Config{
			Workers:            1,
			IndexCacheMaxItems: 10,
		}, promutils.NewTestScope())
		assert.NoError(t, err)
		assert.NotNil(t, o)
	})
}

func Test_buildFinalPhases(t *testing.T) {
	executedTasks := arrayCore.NewPhasesCompactArray(2)
	executedTasks.SetItem(0, bitarray.Item(pluginCore.PhaseSuccess))
	executedTasks.SetItem(1, bitarray.Item(pluginCore.PhaseRetryableFailure))

	indexes := bitarray.NewBitSet(5)
	indexes.Set(1)
	indexes.Set(2)

	expected := arrayCore.NewPhasesCompactArray(5)
	expected.SetItem(0, bitarray.Item(pluginCore.PhaseSuccess))
	expected.SetItem(1, bitarray.Item(pluginCore.PhaseSuccess))
	expected.SetItem(2, bitarray.Item(pluginCore.PhaseRetryableFailure))
	expected.SetItem(3, bitarray.Item(pluginCore.PhaseSuccess))
	expected.SetItem(4, bitarray.Item(pluginCore.PhaseSuccess))

	actual := buildFinalPhases(executedTasks, indexes, 5)

	if diff := deep.Equal(expected, actual); diff != nil {
		t.Errorf("expected != actual. Diff: %v", diff)
		for i, expectedPhaseIdx := range expected.GetItems() {
			actualPhaseIdx := actual.GetItem(i)
			if expectedPhaseIdx != actualPhaseIdx {
				t.Errorf("[%v] expected = %v, actual = %v", i, pluginCore.Phases[expectedPhaseIdx], pluginCore.Phases[actualPhaseIdx])
			}
		}
	}
}
