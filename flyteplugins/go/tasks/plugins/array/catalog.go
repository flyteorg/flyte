package array

import (
	"context"
	"fmt"
	"strconv"

	arrayCore "github.com/lyft/flyteplugins/go/tasks/plugins/array/core"

	"github.com/lyft/flyteplugins/go/tasks/errors"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/lyft/flytestdlib/bitarray"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/storage"

	idlCore "github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
)

// Check if there are any previously cached tasks. If there are we will only submit an ArrayJob for the
// non-cached tasks. The ArrayJob is now a different size, and each task will get a new index location
// which is different than their original location. To find the original index we construct an indexLookup array.
// The subtask can find it's original index value in indexLookup[JOB_ARRAY_INDEX] where JOB_ARRAY_INDEX is an
// environment variable in the pod
func DetermineDiscoverability(ctx context.Context, tCtx core.TaskExecutionContext, state *arrayCore.State) (
	*arrayCore.State, error) {

	// Check that the taskTemplate is valid
	taskTemplate, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		return state, err
	} else if taskTemplate == nil {
		return state, errors.Errorf(errors.BadTaskSpecification, "Required value not set, taskTemplate is nil")
	}

	// Extract the custom plugin pb
	arrayJob, err := arrayCore.ToArrayJob(taskTemplate.GetCustom())
	if err != nil {
		return state, err
	}

	// Save this in the state
	state = state.SetOriginalArraySize(arrayJob.Size)
	state = state.SetOriginalMinSuccesses(arrayJob.MinSuccesses)

	// If the task is not discoverable, then skip data catalog work and move directly to launch
	if taskTemplate.Metadata == nil || !taskTemplate.Metadata.Discoverable {
		logger.Infof(ctx, "Task is not discoverable, moving to launch phase...")
		// Set an all set indexes to cache. This task won't try to write to catalog anyway.
		state = state.SetIndexesToCache(arrayCore.InvertBitSet(bitarray.NewBitSet(uint(arrayJob.Size)), uint(arrayJob.Size)))
		state = state.SetExecutionArraySize(int(arrayJob.Size))
		state = state.SetPhase(arrayCore.PhasePreLaunch, core.DefaultPhaseVersion).SetReason("Task is not discoverable.")
		return state, nil
	}

	// Otherwise, run the data catalog steps - create and submit work items to the catalog processor,
	// build input readers
	inputReaders, err := ConstructInputReaders(ctx, tCtx.DataStore(), tCtx.InputReader().GetInputPrefixPath(), int(arrayJob.Size))
	if err != nil {
		return state, err
	}

	// build output writers
	outputWriters, err := ConstructOutputWriters(ctx, tCtx.DataStore(), tCtx.OutputWriter().GetOutputPrefixPath(), int(arrayJob.Size))
	if err != nil {
		return state, err
	}

	// build work items from inputs and outputs
	workItems, err := ConstructCatalogReaderWorkItems(ctx, tCtx.TaskReader(), inputReaders, outputWriters)
	if err != nil {
		return state, err
	}

	// Check catalog, and if we have responses from catalog for everything, then move to writing the mapping file.
	future, err := tCtx.Catalog().Download(ctx, workItems...)
	if err != nil {
		return state, err
	}

	switch future.GetResponseStatus() {
	case catalog.ResponseStatusReady:
		if err = future.GetResponseError(); err != nil {
			// TODO: maybe add a config option to decide the behavior on catalog failure.
			logger.Warnf(ctx, "Failing to lookup catalog. Will move on to launching the task. Error: %v", err)

			state = state.SetIndexesToCache(arrayCore.InvertBitSet(bitarray.NewBitSet(uint(arrayJob.Size)), uint(arrayJob.Size)))
			state = state.SetExecutionArraySize(int(arrayJob.Size))
			state = state.SetPhase(arrayCore.PhasePreLaunch, core.DefaultPhaseVersion).SetReason(fmt.Sprintf("Skipping cache check due to err [%v]", err))
			return state, nil
		}

		logger.Debug(ctx, "Catalog download response is ready.")
		resp, err := future.GetResponse()
		if err != nil {
			return state, err
		}

		cachedResults := resp.GetCachedResults()
		state = state.SetIndexesToCache(arrayCore.InvertBitSet(cachedResults, uint(arrayJob.Size)))
		state = state.SetExecutionArraySize(int(arrayJob.Size) - resp.GetCachedCount())

		// If all the sub-tasks are actually done, then we can just move on.
		if resp.GetCachedCount() == int(arrayJob.Size) {
			state.SetPhase(arrayCore.PhaseAssembleFinalOutput, core.DefaultPhaseVersion).SetReason("All subtasks are cached. assembling final outputs.")
			return state, nil
		}

		indexLookup := CatalogBitsetToLiteralCollection(cachedResults, resp.GetResultsSize())
		// TODO: Is the right thing to use?  Haytham please take a look
		indexLookupPath, err := ioutils.GetIndexLookupPath(ctx, tCtx.DataStore(), tCtx.OutputWriter().GetOutputPrefixPath())
		if err != nil {
			return state, err
		}

		logger.Infof(ctx, "Writing indexlookup file to [%s], cached count [%d/%d], ",
			indexLookupPath, resp.GetCachedCount(), arrayJob.Size)
		err = tCtx.DataStore().WriteProtobuf(ctx, indexLookupPath, storage.Options{}, indexLookup)
		if err != nil {
			return state, err
		}

		state = state.SetPhase(arrayCore.PhasePreLaunch, core.DefaultPhaseVersion).SetReason("Finished cache lookup.")
	case catalog.ResponseStatusNotReady:
		ownerSignal := tCtx.TaskRefreshIndicator()
		future.OnReady(func(ctx context.Context, _ catalog.Future) {
			ownerSignal(ctx)
		})
	}

	return state, nil
}

func WriteToDiscovery(ctx context.Context, tCtx core.TaskExecutionContext, state *arrayCore.State, phaseOnSuccess arrayCore.Phase) (*arrayCore.State, error) {

	// Check that the taskTemplate is valid
	taskTemplate, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		return state, err
	} else if taskTemplate == nil {
		return state, errors.Errorf(errors.BadTaskSpecification, "Required value not set, taskTemplate is nil")
	}

	if tMeta := taskTemplate.Metadata; tMeta == nil || !tMeta.Discoverable {
		logger.Debugf(ctx, "Task is not marked as discoverable. Moving to [%v] phase.", phaseOnSuccess)
		return state.SetPhase(phaseOnSuccess, core.DefaultPhaseVersion).SetReason("Task is not discoverable."), nil
	}

	// Extract the custom plugin pb
	arrayJob, err := arrayCore.ToArrayJob(taskTemplate.GetCustom())
	if err != nil {
		return state, err
	} else if arrayJob == nil {
		return state, errors.Errorf(errors.BadTaskSpecification, "Could not extract custom array job")
	}

	// input readers
	inputReaders, err := ConstructInputReaders(ctx, tCtx.DataStore(), tCtx.InputReader().GetInputPrefixPath(), int(arrayJob.Size))
	if err != nil {
		return nil, err
	}

	// output reader
	outputReaders, err := ConstructOutputReaders(ctx, tCtx.DataStore(), tCtx.OutputWriter().GetOutputPrefixPath(), int(arrayJob.Size))
	if err != nil {
		return nil, err
	}

	iface := *taskTemplate.Interface
	iface.Outputs = makeSingularTaskInterface(iface.Outputs)

	// Do not cache failed tasks. Retrieve the final phase from array status and unset the non-successful ones.
	tasksToCache := state.GetIndexesToCache().DeepCopy()
	for idx, phaseIdx := range state.ArrayStatus.Detailed.GetItems() {
		phase := core.Phases[phaseIdx]
		if !phase.IsSuccess() {
			tasksToCache.Clear(uint(idx))
		}
	}

	// Create catalog put items, but only put the ones that were not originally cached (as read from the catalog results bitset)
	catalogWriterItems, err := ConstructCatalogUploadRequests(*tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetID().TaskId,
		tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetID(), taskTemplate.Metadata.DiscoveryVersion,
		iface, &tasksToCache, inputReaders, outputReaders)

	if err != nil {
		return nil, err
	}

	if len(catalogWriterItems) == 0 {
		state.SetPhase(phaseOnSuccess, core.DefaultPhaseVersion).SetReason("No outputs need to be cached.")
		return state, nil
	}

	allWritten, err := WriteToCatalog(ctx, tCtx.TaskRefreshIndicator(), tCtx.Catalog(), catalogWriterItems)
	if err != nil {
		return nil, err
	}

	if allWritten {
		state.SetPhase(phaseOnSuccess, core.DefaultPhaseVersion).SetReason("Finished writing catalog cache.")
	}

	return state, nil
}

func WriteToCatalog(ctx context.Context, ownerSignal core.SignalAsync, catalogClient catalog.AsyncClient,
	workItems []catalog.UploadRequest) (bool, error) {

	// Enqueue work items
	future, err := catalogClient.Upload(ctx, workItems...)
	if err != nil {
		return false, errors.Wrapf(arrayCore.ErrorWorkQueue, err,
			"Error enqueuing work items")
	}

	// Immediately read back from the work queue, and see if it's done.
	if future.GetResponseStatus() == catalog.ResponseStatusReady {
		if err = future.GetResponseError(); err != nil {
			// TODO: Add a config option to determine the behavior of catalog write failure.
			logger.Warnf(ctx, "Catalog write failed. Will be ignored. Error: %v", err)
		}

		return true, nil
	}

	future.OnReady(func(ctx context.Context, _ catalog.Future) {
		ownerSignal(ctx)
	})

	return false, nil
}

func ConstructCatalogUploadRequests(keyId idlCore.Identifier, taskExecId idlCore.TaskExecutionIdentifier,
	cacheVersion string, taskInterface idlCore.TypedInterface, whichTasksToCache *bitarray.BitSet,
	inputReaders []io.InputReader, outputReaders []io.OutputReader) ([]catalog.UploadRequest, error) {

	writerWorkItems := make([]catalog.UploadRequest, 0, len(inputReaders))

	if len(inputReaders) != len(outputReaders) {
		return nil, errors.Errorf(arrayCore.ErrorInternalMismatch, "Length different building catalog writer items %d %d",
			len(inputReaders), len(outputReaders))
	}

	for idx, input := range inputReaders {
		if !whichTasksToCache.IsSet(uint(idx)) {
			continue
		}

		wi := catalog.UploadRequest{
			Key: catalog.Key{
				Identifier:     keyId,
				InputReader:    input,
				CacheVersion:   cacheVersion,
				TypedInterface: taskInterface,
			},
			ArtifactData: outputReaders[idx],
			ArtifactMetadata: catalog.Metadata{
				TaskExecutionIdentifier: &taskExecId,
			},
		}

		writerWorkItems = append(writerWorkItems, wi)
	}

	return writerWorkItems, nil
}

func NewLiteralScalarOfInteger(number int64) *idlCore.Literal {
	return &idlCore.Literal{
		Value: &idlCore.Literal_Scalar{
			Scalar: &idlCore.Scalar{
				Value: &idlCore.Scalar_Primitive{
					Primitive: &idlCore.Primitive{
						Value: &idlCore.Primitive_Integer{
							Integer: number,
						},
					},
				},
			},
		},
	}
}

// When an AWS Batch array job kicks off, it is given the index of the array job in an environment variable.
// The SDK will use this index to look up the real index of the job using the output of this function. That is,
// if there are five subtasks originally, but 0-2 are cached in Catalog, then an array job with two jobs will kick off.
// The first job will have an AWS supplied index of 0, which will resolve to 3 from this function, and the second
// will have an index of 1, which will resolve to 4.
// The size argument to this function is needed because the BitSet may create more bits (has a capacity) higher than
// the original requested amount. If you make a BitSet with 10 bits, it may create 64 in the background, so you need
// to keep track of how many were actually requested.
func CatalogBitsetToLiteralCollection(catalogResults *bitarray.BitSet, size int) *idlCore.LiteralCollection {
	literals := make([]*idlCore.Literal, 0, size)
	for i := 0; i < size; i++ {
		if !catalogResults.IsSet(uint(i)) {
			literals = append(literals, NewLiteralScalarOfInteger(int64(i)))
		}
	}
	return &idlCore.LiteralCollection{
		Literals: literals,
	}
}

func makeSingularTaskInterface(varMap *idlCore.VariableMap) *idlCore.VariableMap {
	if varMap == nil || len(varMap.Variables) == 0 {
		return varMap
	}

	res := &idlCore.VariableMap{
		Variables: make(map[string]*idlCore.Variable, len(varMap.Variables)),
	}

	for key, val := range varMap.Variables {
		if val.GetType().GetCollectionType() != nil {
			res.Variables[key] = &idlCore.Variable{Type: val.GetType().GetCollectionType()}
		} else {
			res.Variables[key] = val
		}
	}

	return res

}

func ConstructCatalogReaderWorkItems(ctx context.Context, taskReader core.TaskReader, inputs []io.InputReader,
	outputs []io.OutputWriter) ([]catalog.DownloadRequest, error) {

	t, err := taskReader.Read(ctx)
	if err != nil {
		return nil, err
	}

	workItems := make([]catalog.DownloadRequest, 0, len(inputs))

	iface := *t.Interface
	iface.Outputs = makeSingularTaskInterface(iface.Outputs)

	for idx, inputReader := range inputs {
		// TODO: Check if Id or Interface are empty and return err
		item := catalog.DownloadRequest{
			Key: catalog.Key{
				Identifier:     *t.Id,
				CacheVersion:   t.GetMetadata().DiscoveryVersion,
				InputReader:    inputReader,
				TypedInterface: iface,
			},
			Target: outputs[idx],
		}
		workItems = append(workItems, item)
	}

	return workItems, nil
}

func ConstructInputReaders(ctx context.Context, dataStore *storage.DataStore, inputPrefix storage.DataReference,
	size int) ([]io.InputReader, error) {

	inputReaders := make([]io.InputReader, 0, size)
	for i := 0; i < size; i++ {
		indexedInputLocation, err := dataStore.ConstructReference(ctx, inputPrefix, strconv.Itoa(i))
		if err != nil {
			return inputReaders, err
		}

		inputReader := ioutils.NewRemoteFileInputReader(ctx, dataStore, ioutils.NewInputFilePaths(ctx, dataStore, indexedInputLocation))
		inputReaders = append(inputReaders, inputReader)
	}

	return inputReaders, nil
}

func ConstructOutputWriters(ctx context.Context, dataStore *storage.DataStore, outputPrefix storage.DataReference,
	size int) ([]io.OutputWriter, error) {

	outputWriters := make([]io.OutputWriter, 0, size)

	for i := 0; i < size; i++ {
		ow, err := ConstructOutputWriter(ctx, dataStore, outputPrefix, i)
		if err != nil {
			return outputWriters, err
		}

		outputWriters = append(outputWriters, ow)
	}

	return outputWriters, nil
}

func ConstructOutputWriter(ctx context.Context, dataStore *storage.DataStore, outputPrefix storage.DataReference,
	index int) (io.OutputWriter, error) {
	dataReference, err := dataStore.ConstructReference(ctx, outputPrefix, strconv.Itoa(index))
	if err != nil {
		return nil, err
	}

	return ioutils.NewRemoteFileOutputWriter(ctx, dataStore, ioutils.NewRemoteFileOutputPaths(ctx, dataStore, dataReference)), nil
}

func ConstructOutputReaders(ctx context.Context, dataStore *storage.DataStore, outputPrefix storage.DataReference,
	size int) ([]io.OutputReader, error) {

	outputReaders := make([]io.OutputReader, 0, size)

	for i := 0; i < size; i++ {
		reader, err := ConstructOutputReader(ctx, dataStore, outputPrefix, i)
		if err != nil {
			return nil, err
		}

		outputReaders = append(outputReaders, reader)
	}

	return outputReaders, nil
}

func ConstructOutputReader(ctx context.Context, dataStore *storage.DataStore, outputPrefix storage.DataReference,
	index int) (io.OutputReader, error) {
	dataReference, err := dataStore.ConstructReference(ctx, outputPrefix, strconv.Itoa(index))
	if err != nil {
		return nil, err
	}

	outputPath := ioutils.NewRemoteFileOutputPaths(ctx, dataStore, dataReference)
	return ioutils.NewRemoteFileOutputReader(ctx, dataStore, outputPath, int64(999999999)), nil
}
