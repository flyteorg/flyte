package array

import (
	"context"
	"fmt"
	"math"
	"strconv"

	idlCore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	idlPlugins "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/plugins"

	arrayCore "github.com/flyteorg/flyteplugins/go/tasks/plugins/array/core"

	"github.com/flyteorg/flytestdlib/bitarray"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/storage"

	"github.com/flyteorg/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/ioutils"
)

const AwsBatchTaskType = "aws-batch"

// DetermineDiscoverability checks if there are any previously cached tasks. If there are we will only submit an
// ArrayJob for the non-cached tasks. The ArrayJob is now a different size, and each task will get a new index location
// which is different than their original location. To find the original index we construct an indexLookup array.
// The subtask can find it's original index value in indexLookup[JOB_ARRAY_INDEX] where JOB_ARRAY_INDEX is an
// environment variable in the pod
func DetermineDiscoverability(ctx context.Context, tCtx core.TaskExecutionContext, maxArrayJobSize int64, state *arrayCore.State) (
	*arrayCore.State, error) {

	// Check that the taskTemplate is valid
	taskTemplate, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		return state, err
	} else if taskTemplate == nil {
		return state, errors.Errorf(errors.BadTaskSpecification, "Required value not set, taskTemplate is nil")
	}

	// Extract the custom plugin pb
	var arrayJob *idlPlugins.ArrayJob
	if taskTemplate.Type == AwsBatchTaskType {
		arrayJob = &idlPlugins.ArrayJob{
			Parallelism: 1,
			Size:        1,
			SuccessCriteria: &idlPlugins.ArrayJob_MinSuccesses{
				MinSuccesses: 1,
			},
		}
	} else {
		arrayJob, err = arrayCore.ToArrayJob(taskTemplate.GetCustom(), taskTemplate.TaskTypeVersion)
	}
	if err != nil {
		return state, err
	}

	var arrayJobSize int64
	var inputReaders []io.InputReader

	// Save this in the state
	if taskTemplate.TaskTypeVersion == 0 {
		state = state.SetOriginalArraySize(arrayJob.Size)
		arrayJobSize = arrayJob.Size
		state = state.SetOriginalMinSuccesses(arrayJob.GetMinSuccesses())

		// build input readers
		inputReaders, err = ConstructRemoteFileInputReaders(ctx, tCtx.DataStore(), tCtx.InputReader().GetInputPrefixPath(), int(arrayJobSize))
		if err != nil {
			return state, err
		}
	} else {
		inputs, err := tCtx.InputReader().Get(ctx)
		if err != nil {
			return state, errors.Errorf(errors.MetadataAccessFailed, "Could not read inputs and therefore failed to determine array job size")
		}

		size := -1
		var literalCollection *idlCore.LiteralCollection
		literals := make([][]*idlCore.Literal, 0)
		discoveredInputNames := make([]string, 0)
		for inputName, literal := range inputs.Literals {
			if literalCollection = literal.GetCollection(); literalCollection != nil {
				// validate length of input list
				if size != -1 && size != len(literalCollection.Literals) {
					state = state.SetPhase(arrayCore.PhasePermanentFailure, 0).SetReason("all maptask input lists must be the same length")
					return state, nil
				}

				literals = append(literals, literalCollection.Literals)
				discoveredInputNames = append(discoveredInputNames, inputName)

				size = len(literalCollection.Literals)
			}
		}

		if size < 0 {
			// Something is wrong, we should have inferred the array size when it is not specified by the size of the
			// input collection (for any input value). Non-collection type inputs are not currently supported for
			// taskTypeVersion > 0.
			return state, errors.Errorf(errors.BadTaskSpecification, "Unable to determine array size from inputs")
		}

		minSuccesses := math.Ceil(float64(arrayJob.GetMinSuccessRatio()) * float64(size))

		logger.Debugf(ctx, "Computed state: size [%d] and minSuccesses [%d]", int64(size), int64(minSuccesses))
		state = state.SetOriginalArraySize(int64(size))
		// We can cast the min successes because we already computed the ceiling value from the ratio
		state = state.SetOriginalMinSuccesses(int64(minSuccesses))

		arrayJobSize = int64(size)

		// build input readers
		inputReaders = ConstructStaticInputReaders(tCtx.InputReader(), literals, discoveredInputNames)
	}

	if arrayJobSize > maxArrayJobSize {
		ee := fmt.Errorf("array size > max allowed. requested [%v]. allowed [%v]", arrayJobSize, maxArrayJobSize)
		logger.Info(ctx, ee)
		state = state.SetPhase(arrayCore.PhasePermanentFailure, 0).SetReason(ee.Error())
		return state, nil
	}

	// If the task is not discoverable, then skip data catalog work and move directly to launch
	if taskTemplate.Metadata == nil || !taskTemplate.Metadata.Discoverable {
		logger.Infof(ctx, "Task is not discoverable, moving to launch phase...")
		// Set an all set indexes to cache. This task won't try to write to catalog anyway.
		state = state.SetIndexesToCache(arrayCore.InvertBitSet(bitarray.NewBitSet(uint(arrayJobSize)), uint(arrayJobSize)))
		state = state.SetPhase(arrayCore.PhasePreLaunch, core.DefaultPhaseVersion).SetReason("Task is not discoverable.")

		state.SetExecutionArraySize(int(arrayJobSize))
		return state, nil
	}

	// Otherwise, run the data catalog steps - create and submit work items to the catalog processor,

	// check that the number of items in the cache index LRU cache is greater than the number of
	// jobs in the array. if not we will never complete the cache lookup and be stuck in an
	// infinite loop.
	cfg := catalog.GetConfig()
	if int(arrayJobSize) > cfg.WriterWorkqueueConfig.IndexCacheMaxItems || int(arrayJobSize) > cfg.ReaderWorkqueueConfig.IndexCacheMaxItems {
		ee := fmt.Errorf("array size > max allowed for cache lookup. requested [%v]. writer allowed [%v] reader allowed [%v]",
			arrayJobSize, cfg.WriterWorkqueueConfig.IndexCacheMaxItems, cfg.ReaderWorkqueueConfig.IndexCacheMaxItems)
		logger.Error(ctx, ee)
		state = state.SetPhase(arrayCore.PhasePermanentFailure, 0).SetReason(ee.Error())
		return state, nil
	}

	// build output writers
	outputWriters, err := ConstructOutputWriters(ctx, tCtx.DataStore(), tCtx.OutputWriter().GetOutputPrefixPath(), tCtx.OutputWriter().GetRawOutputPrefix(), int(arrayJobSize))
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

			state = state.SetIndexesToCache(arrayCore.InvertBitSet(bitarray.NewBitSet(uint(arrayJobSize)), uint(arrayJobSize)))
			state = state.SetExecutionArraySize(int(arrayJobSize))
			state = state.SetPhase(arrayCore.PhasePreLaunch, core.DefaultPhaseVersion).SetReason(fmt.Sprintf("Skipping cache check due to err [%v]", err))
			return state, nil
		}

		logger.Debug(ctx, "Catalog download response is ready.")
		resp, err := future.GetResponse()
		if err != nil {
			return state, err
		}

		cachedResults := resp.GetCachedResults()
		state = state.SetIndexesToCache(arrayCore.InvertBitSet(cachedResults, uint(arrayJobSize)))
		state = state.SetExecutionArraySize(int(arrayJobSize) - resp.GetCachedCount())

		// If all the sub-tasks are actually done, then we can just move on.
		if resp.GetCachedCount() == int(arrayJobSize) {
			state.SetPhase(arrayCore.PhaseAssembleFinalOutput, core.DefaultPhaseVersion).SetReason("All subtasks are cached. assembling final outputs.")
			return state, nil
		}

		indexLookup := CatalogBitsetToLiteralCollection(cachedResults, resp.GetResultsSize())
		// TODO: Is the right thing to use?  Haytham please take a look
		indexLookupPath, err := ioutils.GetIndexLookupPath(ctx, tCtx.DataStore(), tCtx.InputReader().GetInputPrefixPath())
		if err != nil {
			return state, err
		}

		logger.Infof(ctx, "Writing indexlookup file to [%s], cached count [%d/%d], ",
			indexLookupPath, resp.GetCachedCount(), arrayJobSize)
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

func WriteToDiscovery(ctx context.Context, tCtx core.TaskExecutionContext, state *arrayCore.State, phaseOnSuccess arrayCore.Phase, versionOnSuccess uint32) (*arrayCore.State, []*core.ExternalResource, error) {
	var externalResources []*core.ExternalResource

	// Check that the taskTemplate is valid
	taskTemplate, err := tCtx.TaskReader().Read(ctx)
	if err != nil {
		return state, externalResources, err
	} else if taskTemplate == nil {
		return state, externalResources, errors.Errorf(errors.BadTaskSpecification, "Required value not set, taskTemplate is nil")
	}

	if tMeta := taskTemplate.Metadata; tMeta == nil || !tMeta.Discoverable {
		logger.Debugf(ctx, "Task is not marked as discoverable. Moving to [%v] phase.", phaseOnSuccess)
		return state.SetPhase(phaseOnSuccess, versionOnSuccess).SetReason("Task is not discoverable."), externalResources, nil
	}

	var inputReaders []io.InputReader
	arrayJobSize := int(state.GetOriginalArraySize())
	if taskTemplate.TaskTypeVersion == 0 {
		// input readers
		inputReaders, err = ConstructRemoteFileInputReaders(ctx, tCtx.DataStore(), tCtx.InputReader().GetInputPrefixPath(), arrayJobSize)
		if err != nil {
			return nil, externalResources, err
		}
	} else {
		inputs, err := tCtx.InputReader().Get(ctx)
		if err != nil {
			return state, externalResources, errors.Errorf(errors.MetadataAccessFailed, "Could not read inputs and therefore failed to determine array job size")
		}

		var literalCollection *idlCore.LiteralCollection
		literals := make([][]*idlCore.Literal, 0)
		discoveredInputNames := make([]string, 0)
		for inputName, literal := range inputs.Literals {
			if literalCollection = literal.GetCollection(); literalCollection != nil {
				literals = append(literals, literalCollection.Literals)
				discoveredInputNames = append(discoveredInputNames, inputName)
			}
		}

		// build input readers
		inputReaders = ConstructStaticInputReaders(tCtx.InputReader(), literals, discoveredInputNames)
	}

	// output reader
	outputReaders, err := ConstructOutputReaders(ctx, tCtx.DataStore(), tCtx.OutputWriter().GetOutputPrefixPath(), tCtx.OutputWriter().GetRawOutputPrefix(), arrayJobSize)
	if err != nil {
		return nil, externalResources, err
	}

	iface := *taskTemplate.Interface
	iface.Outputs = makeSingularTaskInterface(iface.Outputs)

	// Do not cache failed tasks. Retrieve the final phase from array status and unset the non-successful ones.

	tasksToCache := state.GetIndexesToCache().DeepCopy()
	for idx, phaseIdx := range state.ArrayStatus.Detailed.GetItems() {
		phase := core.Phases[phaseIdx]
		if !phase.IsSuccess() {
			// tasksToCache is built on the originalArraySize and ArrayStatus.Detailed is the executionArraySize
			originalIdx := arrayCore.CalculateOriginalIndex(idx, state.GetIndexesToCache())
			tasksToCache.Clear(uint(originalIdx))
		}
	}

	// Create catalog put items, but only put the ones that were not originally cached (as read from the catalog results bitset)
	catalogWriterItems, err := ConstructCatalogUploadRequests(*tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetID().TaskId,
		tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetID(), taskTemplate.Metadata.DiscoveryVersion,
		iface, &tasksToCache, inputReaders, outputReaders)

	if err != nil {
		return nil, externalResources, err
	}

	if len(catalogWriterItems) == 0 {
		state.SetPhase(phaseOnSuccess, versionOnSuccess).SetReason("No outputs need to be cached.")
		return state, externalResources, nil
	}

	allWritten, err := WriteToCatalog(ctx, tCtx.TaskRefreshIndicator(), tCtx.Catalog(), catalogWriterItems)
	if err != nil {
		return nil, externalResources, err
	}

	if allWritten {
		state.SetPhase(phaseOnSuccess, versionOnSuccess).SetReason("Finished writing catalog cache.")

		// set CACHE_POPULATED CacheStatus on all cached subtasks
		externalResources = make([]*core.ExternalResource, 0)
		for idx, phaseIdx := range state.ArrayStatus.Detailed.GetItems() {
			originalIdx := arrayCore.CalculateOriginalIndex(idx, state.GetIndexesToCache())
			if !tasksToCache.IsSet(uint(originalIdx)) {
				continue
			}

			externalResources = append(externalResources,
				&core.ExternalResource{
					CacheStatus:  idlCore.CatalogCacheStatus_CACHE_POPULATED,
					Index:        uint32(originalIdx),
					RetryAttempt: uint32(state.RetryAttempts.GetItem(idx)),
					Phase:        core.Phases[phaseIdx],
				},
			)
		}
	}

	return state, externalResources, nil
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

func ConstructCatalogUploadRequests(keyID idlCore.Identifier, taskExecID idlCore.TaskExecutionIdentifier,
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
				Identifier:     keyID,
				InputReader:    input,
				CacheVersion:   cacheVersion,
				TypedInterface: taskInterface,
			},
			ArtifactData: outputReaders[idx],
			ArtifactMetadata: catalog.Metadata{
				TaskExecutionIdentifier: &taskExecID,
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
		// TODO: Check if Identifier or Interface are empty and return err
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

// ConstructStaticInputReaders constructs input readers that comply with the io.InputReader interface but have their
// inputs already populated.
func ConstructStaticInputReaders(inputPaths io.InputFilePaths, inputs [][]*idlCore.Literal, inputNames []string) []io.InputReader {
	inputReaders := make([]io.InputReader, 0, len(inputs))
	if len(inputs) == 0 {
		return inputReaders
	}

	for i := 0; i < len(inputs[0]); i++ {
		literals := make(map[string]*idlCore.Literal)
		for j := 0; j < len(inputNames); j++ {
			literals[inputNames[j]] = inputs[j][i]
		}

		inputReaders = append(inputReaders, NewStaticInputReader(inputPaths, &idlCore.LiteralMap{Literals: literals}))
	}

	return inputReaders
}

func ConstructRemoteFileInputReaders(ctx context.Context, dataStore *storage.DataStore, inputPrefix storage.DataReference,
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

func ConstructOutputWriters(ctx context.Context, dataStore *storage.DataStore, outputPrefix, baseOutputSandbox storage.DataReference,
	size int) ([]io.OutputWriter, error) {

	outputWriters := make([]io.OutputWriter, 0, size)

	for i := 0; i < size; i++ {
		outputSandbox, err := dataStore.ConstructReference(ctx, baseOutputSandbox, strconv.Itoa(i))
		if err != nil {
			return nil, err
		}
		ow, err := ConstructOutputWriter(ctx, dataStore, outputPrefix, outputSandbox, i)
		if err != nil {
			return outputWriters, err
		}

		outputWriters = append(outputWriters, ow)
	}

	return outputWriters, nil
}

func ConstructOutputWriter(ctx context.Context, dataStore *storage.DataStore, outputPrefix, outputSandbox storage.DataReference,
	index int) (io.OutputWriter, error) {
	dataReference, err := dataStore.ConstructReference(ctx, outputPrefix, strconv.Itoa(index))
	if err != nil {
		return nil, err
	}

	// checkpoint paths are not computed here because this function is only called when writing
	// existing cached outputs. if this functionality changes this will need to be revisited.
	p := ioutils.NewCheckpointRemoteFilePaths(ctx, dataStore, dataReference, ioutils.NewRawOutputPaths(ctx, outputSandbox), "")
	return ioutils.NewRemoteFileOutputWriter(ctx, dataStore, p), nil
}

func ConstructOutputReaders(ctx context.Context, dataStore *storage.DataStore, outputPrefix, baseOutputSandbox storage.DataReference,
	size int) ([]io.OutputReader, error) {

	outputReaders := make([]io.OutputReader, 0, size)

	for i := 0; i < size; i++ {
		reader, err := ConstructOutputReader(ctx, dataStore, outputPrefix, baseOutputSandbox, i)
		if err != nil {
			return nil, err
		}

		outputReaders = append(outputReaders, reader)
	}

	return outputReaders, nil
}

func ConstructOutputReader(ctx context.Context, dataStore *storage.DataStore, outputPrefix, baseOutputSandbox storage.DataReference,
	index int) (io.OutputReader, error) {
	strIndex := strconv.Itoa(index)
	dataReference, err := dataStore.ConstructReference(ctx, outputPrefix, strIndex)
	if err != nil {
		return nil, err
	}

	outputSandbox, err := dataStore.ConstructReference(ctx, baseOutputSandbox, strIndex)
	if err != nil {
		return nil, err
	}

	// checkpoint paths are not computed here because this function is only called when writing
	// existing cached outputs. if this functionality changes this will need to be revisited.
	outputPath := ioutils.NewCheckpointRemoteFilePaths(ctx, dataStore, dataReference, ioutils.NewRawOutputPaths(ctx, outputSandbox), "")
	return ioutils.NewRemoteFileOutputReader(ctx, dataStore, outputPath, int64(999999999)), nil
}
