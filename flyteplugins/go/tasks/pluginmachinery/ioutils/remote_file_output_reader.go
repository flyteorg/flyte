package ioutils

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

type baseErrorReader struct {
	store          storage.ComposedProtobufStore
	maxPayloadSize int64
}

type singleFileErrorReader struct {
	baseErrorReader
	errorFilePath storage.DataReference
}

type earliestFileErrorReader struct {
	baseErrorReader
	errorDirPath        storage.DataReference
	errorFilePathPrefix storage.DataReference
	errorFileExtension  string
}

/*
   We have a 'single file error reader' and 'earliest file error reader' as two
   different strategies for reading task error files.

   Single file error reader is used to check for a single error.pb file uploaded
   by a task, and is the default strategy. Earliest file error reader is used to check for
   multiple error-<uuid>.pb files and pick the one that has the earliest error timestamp.
   It is used when a distributed task requests earliest timestamp error aggregation
   strategy. To support backward compatibility, the earliest file error reader also handles
   cases when there is a single error.pb file uploaded by the task. The earliest file
   error reader is currently used for the PyTorch plugin.

   A few notes:

   - While the earliest file error reader handles the single error file scenario as well,
   it is not set as the default, because its implementation depends on doing a listing operation
   on remote storage. We do not want the listing overhead to be paid for the more common case of
   having a single error file.
   - Under the multiple error aggregation scenario, it is possible that the error aggregation
   is performed before all the errors are reported. For PyTorch plugin specifically, the
   the training operator will mark the job as 'done' when it detects one of the pods as failing.
   Once Propeller detects this, it will perform the error aggregation. There is a rare scenario
   where the pod that has the earliest error gets delayed in uploading its error file to
   remote storage, and the pod that has a later error ends up completing first. If the
   training operator's detection of job completion and Propeller's error aggregation happen so
   fast that the pod with the earliest error has not yet uploaded it's error to remote storage,
   we may end up reporting the wrong error. This is highly unlikely in practice. The implementation
   we have here is significantly better than the prior behavior of reporting the latest written
   error.pb file (as there was a race condition on overwriting error files), which is almost always
   not the earliest error.
   - The training operator does not have any error aggregation strategy implemented. PyTorch
   distributed itself aggregates errors from the trainers running under the same elastic connection,
   and reports the earliest error. The aggregation we perform here extends that to across pods.
*/

const errorFileNotFoundErrorCode = "ErrorFileNotFound"

var ErrRemoteFileExceedsMaxSize = errors.New("remote file exceeds max size")

func newSingleFileErrorReader(errorFilePath storage.DataReference, store storage.ComposedProtobufStore, maxPayloadSize int64) *singleFileErrorReader {
	return &singleFileErrorReader{
		baseErrorReader: baseErrorReader{
			store:          store,
			maxPayloadSize: maxPayloadSize,
		},
		errorFilePath: errorFilePath,
	}
}

func (b *baseErrorReader) validatePayloadSize(filePath storage.DataReference, metadata storage.Metadata) error {
	if metadata.Exists() {
		if metadata.Size() > b.maxPayloadSize {
			return errors.Wrapf(ErrRemoteFileExceedsMaxSize,
				"output file @[%s] is too large [%d] bytes, max allowed [%d] bytes",
				filePath, metadata.Size(), b.maxPayloadSize)
		}
	}
	return nil
}

func (s *singleFileErrorReader) IsError(ctx context.Context) (bool, error) {
	metadata, err := s.store.Head(ctx, s.errorFilePath)
	if err != nil {
		return false, errors.Wrapf(err, "failed to read error file @[%s]", s.errorFilePath)
	}
	err = s.validatePayloadSize(s.errorFilePath, metadata)
	if err != nil {
		return false, err
	}
	return metadata.Exists(), nil
}

func errorDoc2ExecutionError(errorDoc *core.ErrorDocument, errorFilePath storage.DataReference) io.ExecutionError {
	if errorDoc.GetError() == nil {
		return io.ExecutionError{
			IsRecoverable: true,
			ExecutionError: &core.ExecutionError{
				Code:    "ErrorFileBadFormat",
				Message: fmt.Sprintf("error not formatted correctly, nil error @path [%s]", errorFilePath),
				Kind:    core.ExecutionError_SYSTEM,
			},
		}
	}
	executionError := io.ExecutionError{
		ExecutionError: &core.ExecutionError{
			Code:      errorDoc.GetError().GetCode(),
			Message:   errorDoc.GetError().GetMessage(),
			Kind:      errorDoc.GetError().GetOrigin(),
			Timestamp: errorDoc.GetError().GetTimestamp(),
			Worker:    errorDoc.GetError().GetWorker(),
		},
	}

	if errorDoc.GetError().GetKind() == core.ContainerError_RECOVERABLE {
		executionError.IsRecoverable = true
	}

	return executionError
}

func (s *singleFileErrorReader) ReadError(ctx context.Context) (io.ExecutionError, error) {
	errorDoc := &core.ErrorDocument{}
	err := s.store.ReadProtobuf(ctx, s.errorFilePath, errorDoc)
	if err != nil {
		if storage.IsNotFound(err) {
			return io.ExecutionError{
				IsRecoverable: true,
				ExecutionError: &core.ExecutionError{
					Code:    errorFileNotFoundErrorCode,
					Message: err.Error(),
					Kind:    core.ExecutionError_SYSTEM,
				},
			}, nil
		}
		return io.ExecutionError{}, errors.Wrapf(err, "failed to read error data from task @[%s]", s.errorFilePath)
	}

	return errorDoc2ExecutionError(errorDoc, s.errorFilePath), nil
}

func (e *earliestFileErrorReader) IsError(ctx context.Context) (bool, error) {
	hasError := false
	const maxItems = 1000
	cursor := storage.NewCursorAtStart()
	for cursor != storage.NewCursorAtEnd() {
		var err error
		var errorFilePaths []storage.DataReference
		errorFilePaths, cursor, err = e.store.List(ctx, e.errorFilePathPrefix, maxItems, cursor)
		if err != nil {
			return false, errors.Wrapf(err, "failed to list error files @[%s]", e.errorDirPath)
		}
		for _, errorFilePath := range errorFilePaths {
			if strings.HasSuffix(errorFilePath.String(), e.errorFileExtension) {
				metadata, err := e.store.Head(ctx, errorFilePath)
				if err != nil {
					return false, errors.Wrapf(err, "failed to read error file @[%s]", errorFilePath)
				}
				err = e.validatePayloadSize(errorFilePath, metadata)
				if err != nil {
					return false, err
				}
				hasError = true
			}
		}
	}
	return hasError, nil
}

func (e *earliestFileErrorReader) ReadError(ctx context.Context) (io.ExecutionError, error) {
	var earliestTimestamp *time.Time = nil
	earliestExecutionError := io.ExecutionError{}
	const maxItems = 1000
	cursor := storage.NewCursorAtStart()
	for cursor != storage.NewCursorAtEnd() {
		var err error
		var errorFilePaths []storage.DataReference
		errorFilePaths, cursor, err = e.store.List(ctx, e.errorFilePathPrefix, maxItems, cursor)
		if err != nil {
			return io.ExecutionError{}, errors.Wrapf(err, "failed to list error files @[%s]", e.errorDirPath)
		}
		for _, errorFilePath := range errorFilePaths {
			if !strings.HasSuffix(errorFilePath.String(), e.errorFileExtension) {
				continue
			}
			errorDoc := &core.ErrorDocument{}
			err := e.store.ReadProtobuf(ctx, errorFilePath, errorDoc)
			if err != nil {
				return io.ExecutionError{}, errors.Wrapf(err, "failed to read error file @[%s]", errorFilePath.String())
			}
			timestamp := errorDoc.GetError().GetTimestamp().AsTime()
			if earliestTimestamp == nil || earliestTimestamp.After(timestamp) {
				earliestExecutionError = errorDoc2ExecutionError(errorDoc, errorFilePath)
				earliestTimestamp = &timestamp
			}
		}
	}
	return earliestExecutionError, nil
}

func newEarliestFileErrorReader(errorDirPath storage.DataReference, canonicalErrorFilename string, store storage.ComposedProtobufStore, maxPayloadSize int64) (*earliestFileErrorReader, error) {
	// If the canonical error file name is error.pb, we expect multiple error files
	// to have name error<suffix>.pb
	pieces := strings.Split(canonicalErrorFilename, ".")
	if len(pieces) != 2 {
		return nil, errors.Errorf("expected canonical error filename to have a single dot (.), got %d", len(pieces))
	}
	errorFilePrefix := pieces[0]
	scheme, container, key, _ := errorDirPath.Split()
	errorFilePathPrefix := storage.NewDataReference(scheme, container, filepath.Join(key, errorFilePrefix))
	errorFileExtension := fmt.Sprintf(".%s", pieces[1])

	return &earliestFileErrorReader{
		baseErrorReader: baseErrorReader{
			store:          store,
			maxPayloadSize: maxPayloadSize,
		},
		errorDirPath:        errorDirPath,
		errorFilePathPrefix: errorFilePathPrefix,
		errorFileExtension:  errorFileExtension,
	}, nil
}

func newErrorReader(errorAggregationStrategy k8s.ErrorAggregationStrategy, errorDirPath storage.DataReference, errorFilename string, store storage.ComposedProtobufStore, maxPayloadSize int64) (io.ErrorReader, error) {
	if errorAggregationStrategy == k8s.DefaultErrorAggregationStrategy {
		scheme, container, key, err := errorDirPath.Split()
		if err != nil {
			return nil, errors.Wrapf(err, "invalid error dir path %s", errorDirPath)
		}
		errorFilePath := storage.NewDataReference(scheme, container, filepath.Join(key, errorFilename))
		return newSingleFileErrorReader(errorFilePath, store, maxPayloadSize), nil
	}
	if errorAggregationStrategy == k8s.EarliestErrorAggregationStrategy {
		return newEarliestFileErrorReader(errorDirPath, errorFilename, store, maxPayloadSize)
	}
	return nil, errors.Errorf("unknown error aggregation strategy: %v", errorAggregationStrategy)
}

type RemoteFileOutputReader struct {
	outPath        io.OutputFilePaths
	store          storage.ComposedProtobufStore
	maxPayloadSize int64
	errorReader    io.ErrorReader
}

func (r RemoteFileOutputReader) IsError(ctx context.Context) (bool, error) {
	return r.errorReader.IsError(ctx)
}

func (r RemoteFileOutputReader) ReadError(ctx context.Context) (io.ExecutionError, error) {
	return r.errorReader.ReadError(ctx)
}

func (r RemoteFileOutputReader) Exists(ctx context.Context) (bool, error) {
	md, err := r.store.Head(ctx, r.outPath.GetOutputPath())
	if err != nil {
		return false, err
	}
	if md.Exists() {
		if md.Size() > r.maxPayloadSize {
			return false, errors.Wrapf(ErrRemoteFileExceedsMaxSize, "output file @[%s] is too large [%d] bytes, max allowed [%d] bytes", r.outPath.GetOutputPath(), md.Size(), r.maxPayloadSize)
		}
		return true, nil
	}
	return false, nil
}

func (r RemoteFileOutputReader) Read(ctx context.Context) (*core.LiteralMap, *io.ExecutionError, error) {

	d := &core.LiteralMap{}
	if err := r.store.ReadProtobuf(ctx, r.outPath.GetOutputPath(), d); err != nil {
		// TODO change flytestdlib to return protobuf unmarshal errors separately. As this can indicate malformed output and we should catch that
		return nil, nil, fmt.Errorf("failed to read data from dataDir [%v]. Error: %v", r.outPath.GetOutputPath(), err)
	}

	if d.Literals == nil {
		return nil, &io.ExecutionError{
			IsRecoverable: true,
			ExecutionError: &core.ExecutionError{
				Code:    "No outputs produced",
				Message: fmt.Sprintf("outputs not found at [%s]", r.outPath.GetOutputPath()),
			},
		}, nil
	}

	return d, nil, nil
}

func (r RemoteFileOutputReader) IsFile(ctx context.Context) bool {
	return true
}

func (r RemoteFileOutputReader) DeckExists(ctx context.Context) (bool, error) {
	md, err := r.store.Head(ctx, r.outPath.GetDeckPath())
	if err != nil {
		return false, err
	}

	return md.Exists(), nil
}

func getMaxPayloadSize(maxDatasetSize int64) int64 {
	// Note: even though the data store retrieval checks against GetLimitMegabytes, there might be external
	// storage implementations, so we keep this check here as well.
	maxPayloadSize := maxDatasetSize
	if maxPayloadSize == 0 {
		maxPayloadSize = storage.GetConfig().Limits.GetLimitMegabytes * 1024 * 1024
	}
	return maxPayloadSize
}

func NewRemoteFileOutputReader(context context.Context, store storage.ComposedProtobufStore, outPaths io.OutputFilePaths, maxDatasetSize int64) RemoteFileOutputReader {
	maxPayloadSize := getMaxPayloadSize(maxDatasetSize)
	errorReader := newSingleFileErrorReader(outPaths.GetErrorPath(), store, maxPayloadSize)
	return RemoteFileOutputReader{
		outPath:        outPaths,
		store:          store,
		maxPayloadSize: maxPayloadSize,
		errorReader:    errorReader,
	}
}

func NewRemoteFileOutputReaderWithErrorAggregationStrategy(_ context.Context, store storage.ComposedProtobufStore, outPaths io.OutputFilePaths, maxDatasetSize int64, errorAggregationStrategy k8s.ErrorAggregationStrategy) (*RemoteFileOutputReader, error) {
	maxPayloadSize := getMaxPayloadSize(maxDatasetSize)
	scheme, container, key, err := outPaths.GetErrorPath().Split()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse error path %s", outPaths.GetErrorPath())
	}
	errorFilename := filepath.Base(key)
	errorDirPath := storage.NewDataReference(scheme, container, filepath.Dir(key))
	errorReader, err := newErrorReader(errorAggregationStrategy, errorDirPath, errorFilename, store, maxPayloadSize)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create remote output reader with error aggregation strategy %v", errorAggregationStrategy)
	}
	return &RemoteFileOutputReader{
		outPath:        outPaths,
		store:          store,
		maxPayloadSize: maxPayloadSize,
		errorReader:    errorReader,
	}, nil
}
