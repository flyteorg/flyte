package ioutils

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

type ErrorRetriever interface {
	HasError(ctx context.Context) (bool, error)
	GetError(ctx context.Context) (io.ExecutionError, error)
}

type ErrorRetrieverBase struct {
	store          storage.ComposedProtobufStore
	maxPayloadSize int64
}

type SingleFileErrorRetriever struct {
	ErrorRetrieverBase
	errorFilePath storage.DataReference
}

func NewSingleFileErrorRetriever(errorFilePath storage.DataReference, store storage.ComposedProtobufStore, maxPayloadSize int64) *SingleFileErrorRetriever {
	return &SingleFileErrorRetriever{
		ErrorRetrieverBase: ErrorRetrieverBase{
			store:          store,
			maxPayloadSize: maxPayloadSize,
		},
		errorFilePath: errorFilePath,
	}
}

func (s *SingleFileErrorRetriever) HasError(ctx context.Context) (bool, error) {
	metadata, err := s.store.Head(ctx, s.errorFilePath)
	if err != nil {
		return false, errors.Wrapf(err, "failed to read error file @[%s]", s.errorFilePath)
	}
	if metadata.Exists() {
		if metadata.Size() > s.maxPayloadSize {
			return false, errors.Wrapf(err, "error file @[%s] is too large [%d] bytes, max allowed [%d] bytes", s.errorFilePath, metadata.Size(), s.maxPayloadSize)
		}
		return true, nil
	}
	return false, nil
}

func errorDoc2ExecutionError(errorDoc *core.ErrorDocument, errorFilePath storage.DataReference) io.ExecutionError {
	if errorDoc.Error == nil {
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
			Code:    errorDoc.Error.Code,
			Message: errorDoc.Error.Message,
			Kind:    errorDoc.Error.Origin,
		},
	}

	if errorDoc.Error.Kind == core.ContainerError_RECOVERABLE {
		executionError.IsRecoverable = true
	}

	if errorDoc.Error.Kind == core.ContainerError_RECOVERABLE {
		executionError.IsRecoverable = true
	}

	return executionError
}

func (s *SingleFileErrorRetriever) GetError(ctx context.Context) (io.ExecutionError, error) {
	errorDoc := &core.ErrorDocument{}
	err := s.store.ReadProtobuf(ctx, storage.DataReference(s.errorFilePath), errorDoc)
	if err != nil {
		if storage.IsNotFound(err) {
			return io.ExecutionError{
				IsRecoverable: true,
				ExecutionError: &core.ExecutionError{
					Code:    "ErrorFileNotFound",
					Message: err.Error(),
					Kind:    core.ExecutionError_SYSTEM,
				},
			}, nil
		}
		return io.ExecutionError{}, errors.Wrapf(err, "failed to read error data from task @[%s]", s.errorFilePath)
	}

	return errorDoc2ExecutionError(errorDoc, s.errorFilePath), nil
}

type EarliestFileErrorRetriever struct {
	ErrorRetrieverBase
	errorDirPath           storage.DataReference
	canonicalErrorFilename string
}

func (e *EarliestFileErrorRetriever) parseErrorFilename() (errorFilePathPrefix storage.DataReference, errorFileExtension string, err error) {
	// If the canonical error file name is error.pb, we expect multiple error files
	// to have name error<suffix>.pb
	pieces := strings.Split(e.canonicalErrorFilename, ".")
	if len(pieces) != 2 {
		err = errors.Errorf("expected canoncal error filename to have a single ., got %d", len(pieces))
		return
	}
	errorFilePrefix := pieces[0]
	scheme, container, key, _ := e.errorDirPath.Split()
	errorFilePathPrefix = storage.NewDataReference(scheme, container, filepath.Join(key, errorFilePrefix))
	errorFileExtension = fmt.Sprintf(".%s", pieces[1])
	return
}

func (e *EarliestFileErrorRetriever) HasError(ctx context.Context) (bool, error) {
	errorFilePathPrefix, errorFileExtension, err := e.parseErrorFilename()
	if err != nil {
		return false, errors.Wrapf(err, "failed to parse canonical error filename @[%s]", e.canonicalErrorFilename)
	}
	const maxItems = 1000
	cursor := storage.NewCursorAtStart()
	for cursor != storage.NewCursorAtEnd() {
		var err error
		var errorFilePaths []storage.DataReference
		errorFilePaths, cursor, err = e.store.List(ctx, errorFilePathPrefix, maxItems, cursor)
		if err != nil {
			return false, errors.Wrapf(err, "failed to list error files @[%s]", e.errorDirPath)
		}
		for _, errorFilePath := range errorFilePaths {
			if strings.HasSuffix(errorFilePath.String(), errorFileExtension) {
				return true, nil
			}
		}
	}
	return false, nil
}

func (e *EarliestFileErrorRetriever) GetError(ctx context.Context) (io.ExecutionError, error) {
	errorFilePathPrefix, errorFileExtension, err := e.parseErrorFilename()
	if err != nil {
		return io.ExecutionError{}, errors.Wrapf(err, "failed to parse canonical error filename @[%s]", e.canonicalErrorFilename)
	}
	const maxItems = 1000
	cursor := storage.NewCursorAtStart()
	type ErrorFileAndDocument struct {
		errorFilePath storage.DataReference
		errorDoc      *core.ErrorDocument
	}
	var errorFileAndDocs []ErrorFileAndDocument
	for cursor != storage.NewCursorAtEnd() {
		var err error
		var errorFilePaths []storage.DataReference
		errorFilePaths, cursor, err = e.store.List(ctx, errorFilePathPrefix, maxItems, cursor)
		if err != nil {
			return io.ExecutionError{}, errors.Wrapf(err, "failed to list error files @[%s]", e.errorDirPath)
		}
		for _, errorFilePath := range errorFilePaths {
			if strings.HasSuffix(errorFilePath.String(), errorFileExtension) {
				errorDoc := &core.ErrorDocument{}
				err := e.store.ReadProtobuf(ctx, errorFilePath, errorDoc)
				if err != nil {
					return io.ExecutionError{}, errors.Wrapf(err, "failed to read error file @[%s]", errorFilePath.String())
				}
				errorFileAndDocs = append(errorFileAndDocs, ErrorFileAndDocument{errorFilePath: errorFilePath, errorDoc: errorDoc})
			}
		}
	}

	extractTimestampFromErrorDoc := func(errorDoc *core.ErrorDocument) int64 {
		// TODO: add optional timestamp to ErrorDocument
		if errorDoc == nil {
			panic("")
		}
		return 0
	}

	var earliestTimestamp int64 = math.MaxInt64
	earliestExecutionError := io.ExecutionError{}
	for _, errorFileAndDoc := range errorFileAndDocs {
		timestamp := extractTimestampFromErrorDoc(errorFileAndDoc.errorDoc)
		if earliestTimestamp >= timestamp {
			earliestExecutionError = errorDoc2ExecutionError(errorFileAndDoc.errorDoc, errorFileAndDoc.errorFilePath)
			earliestTimestamp = timestamp
		}
	}
	return earliestExecutionError, nil
}

func NewEarliestFileErrorRetriever(errorDirPath storage.DataReference, canonicalErrorFilename string, store storage.ComposedProtobufStore, maxPayloadSize int64) *EarliestFileErrorRetriever {
	return &EarliestFileErrorRetriever{
		ErrorRetrieverBase: ErrorRetrieverBase{
			store:          store,
			maxPayloadSize: maxPayloadSize,
		},
		errorDirPath:           errorDirPath,
		canonicalErrorFilename: canonicalErrorFilename,
	}
}

func NewErrorRetriever(errorAggregationStrategy k8s.ErrorAggregationStrategy, errorDirPath storage.DataReference, errorFilename string, store storage.ComposedProtobufStore, maxPayloadSize int64) ErrorRetriever {
	if errorAggregationStrategy == k8s.DefaultErrorAggregationStrategy {
		scheme, container, key, _ := errorDirPath.Split()
		errorFilePath := storage.NewDataReference(scheme, container, filepath.Join(key, errorFilename))
		return NewSingleFileErrorRetriever(errorFilePath, store, maxPayloadSize)
	}
	if errorAggregationStrategy == k8s.EarliestErrorAggregationStrategy {
		return NewEarliestFileErrorRetriever(errorDirPath, errorFilename, store, maxPayloadSize)
	}
	return nil
}

type RemoteFileOutputReader struct {
	outPath        io.OutputFilePaths
	store          storage.ComposedProtobufStore
	maxPayloadSize int64
	errorRetriever ErrorRetriever
}

func (r RemoteFileOutputReader) IsError(ctx context.Context) (bool, error) {
	return r.errorRetriever.HasError(ctx)
}

func (r RemoteFileOutputReader) ReadError(ctx context.Context) (io.ExecutionError, error) {
	return r.errorRetriever.GetError(ctx)
}

func (r RemoteFileOutputReader) Exists(ctx context.Context) (bool, error) {
	md, err := r.store.Head(ctx, r.outPath.GetOutputPath())
	if err != nil {
		return false, err
	}
	if md.Exists() {
		if md.Size() > r.maxPayloadSize {
			return false, errors.Errorf("output file @[%s] is too large [%d] bytes, max allowed [%d] bytes", r.outPath.GetOutputPath(), md.Size(), r.maxPayloadSize)
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

func NewRemoteFileOutputReader(context context.Context, store storage.ComposedProtobufStore, outPaths io.OutputFilePaths, maxDatasetSize int64) RemoteFileOutputReader {
	return NewRemoteFileOutputReaderWithErrorAggregationStrategy(context, store, outPaths, maxDatasetSize, k8s.DefaultErrorAggregationStrategy)
}

func NewRemoteFileOutputReaderWithErrorAggregationStrategy(_ context.Context, store storage.ComposedProtobufStore, outPaths io.OutputFilePaths, maxDatasetSize int64, errorAggregationStrategy k8s.ErrorAggregationStrategy) RemoteFileOutputReader {
	// Note: even though the data store retrieval checks against GetLimitMegabytes, there might be external
	// storage implementations, so we keep this check here as well.
	maxPayloadSize := maxDatasetSize
	if maxPayloadSize == 0 {
		maxPayloadSize = storage.GetConfig().Limits.GetLimitMegabytes * 1024 * 1024
	}
	scheme, container, key, _ := outPaths.GetErrorPath().Split()
	errorFilename := filepath.Base(key)
	errorDirPath := storage.NewDataReference(scheme, container, filepath.Dir(key))
	errorRetriever := NewErrorRetriever(errorAggregationStrategy, errorDirPath, errorFilename, store, maxPayloadSize)
	return RemoteFileOutputReader{
		outPath:        outPaths,
		store:          store,
		maxPayloadSize: maxPayloadSize,
		errorRetriever: errorRetriever,
	}
}
