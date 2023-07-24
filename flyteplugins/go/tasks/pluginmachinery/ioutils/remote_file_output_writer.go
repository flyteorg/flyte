package ioutils

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flytestdlib/storage"
)

var (
	_ io.OutputWriter    = RemoteFileOutputWriter{}
	_ io.OutputFilePaths = RemoteCheckpointPaths{}
)

// RemoteFileOutputPaths records all metadata output paths / keys on a remote storage system, e.g. S3 / GCS or any other
// key-value store. Theoretically if the storage.DataReference can support BigTable, this will work with it.
type RemoteFileOutputPaths struct {
	// All returned paths for the metadata outputs (inputs.pb and outputs.pb) of an input are constructed to under
	// this prefix
	outputPrefix storage.DataReference
	// Implemented of the ReferenceConstructor, used to construct the actual paths
	store storage.ReferenceConstructor
	// Arbitrary supplied of the RawOutputPath
	io.RawOutputPaths
}

func (w RemoteFileOutputPaths) GetOutputPrefixPath() storage.DataReference {
	return w.outputPrefix
}

func (w RemoteFileOutputPaths) GetOutputPath() storage.DataReference {
	return constructPath(w.store, w.outputPrefix, OutputsSuffix)
}

func (w RemoteFileOutputPaths) GetDeckPath() storage.DataReference {
	return constructPath(w.store, w.outputPrefix, deckSuffix)
}

func (w RemoteFileOutputPaths) GetErrorPath() storage.DataReference {
	return constructPath(w.store, w.outputPrefix, ErrorsSuffix)
}

func (w RemoteFileOutputPaths) GetFuturesPath() storage.DataReference {
	return constructPath(w.store, w.outputPrefix, FuturesSuffix)
}

// RemoteFileOutputWriter adds storage Write APIs to output paths / keys. In retrospect, the `path` should be
// generally replaced with keys
type RemoteFileOutputWriter struct {
	io.OutputFilePaths
	store storage.ProtobufStore
}

func (w RemoteFileOutputWriter) Put(ctx context.Context, reader io.OutputReader) error {
	literals, executionErr, err := reader.Read(ctx)
	if err != nil {
		return err
	}

	if executionErr != nil {
		errorKind := core.ContainerError_RECOVERABLE
		if !executionErr.IsRecoverable {
			errorKind = core.ContainerError_NON_RECOVERABLE
		}

		errDoc := &core.ErrorDocument{
			Error: &core.ContainerError{
				Code:    executionErr.Code,
				Message: executionErr.Message,
				Kind:    errorKind,
			},
		}

		return w.store.WriteProtobuf(ctx, w.GetErrorPath(), storage.Options{}, errDoc)
	}

	if literals != nil {
		return w.store.WriteProtobuf(ctx, w.GetOutputPath(), storage.Options{}, literals)
	}

	return fmt.Errorf("no data found to write")
}

// RemoteCheckpointPaths implements the CheckpointPaths Interface and adds on top of the OutputFilePaths Interface
type RemoteCheckpointPaths struct {
	RemoteFileOutputPaths

	previousPath storage.DataReference

	store storage.ReferenceConstructor
}

// GetPreviousCheckpointsPrefix returns the Prefix path for checkpoints for the previous attempt, or "" if this is
// the first attempt
func (r RemoteCheckpointPaths) GetPreviousCheckpointsPrefix() storage.DataReference {
	return r.previousPath
}

// GetCheckpointPrefix returns a new checkpoint path under the raw output prefix.
func (r RemoteCheckpointPaths) GetCheckpointPrefix() storage.DataReference {
	return ConstructCheckpointPath(r.store, r.GetRawOutputPrefix())
}

// NewRemoteFileOutputPaths returns a RemoteFileOutputPaths object, where all the paths are configured using the given
// outputPrefix and constructed using the storage.ReferenceConstructor
func NewRemoteFileOutputPaths(_ context.Context, store storage.ReferenceConstructor, outputPrefix storage.DataReference, sandbox io.RawOutputPaths) RemoteFileOutputPaths {
	return RemoteFileOutputPaths{
		store:          store,
		outputPrefix:   outputPrefix,
		RawOutputPaths: sandbox,
	}
}

// NewCheckpointRemoteFilePaths returns a new object constructed with an optional previousCheckpointPath and derives a new checkpointPath from the outputPrefix
func NewCheckpointRemoteFilePaths(ctx context.Context, store storage.ReferenceConstructor, outputPrefix storage.DataReference, sandbox io.RawOutputPaths, previousCheckpointPath storage.DataReference) RemoteCheckpointPaths {
	return RemoteCheckpointPaths{
		previousPath:          previousCheckpointPath,
		store:                 store,
		RemoteFileOutputPaths: NewRemoteFileOutputPaths(ctx, store, outputPrefix, sandbox),
	}
}

// NewReadOnlyOutputFilePaths can be used when data is only to be read from an existing remote location
func NewReadOnlyOutputFilePaths(ctx context.Context, store storage.ReferenceConstructor, outputPrefix storage.DataReference) RemoteCheckpointPaths {
	return NewCheckpointRemoteFilePaths(ctx, store, outputPrefix, nil, "")
}

// NewRemoteFileOutputWriter returns a writer that records all outputs to remote files / objects. Given outputs,
// it will automatically write it to the outputFile / key that is configured.
func NewRemoteFileOutputWriter(_ context.Context, store storage.ProtobufStore, outputFilePaths io.OutputFilePaths) RemoteFileOutputWriter {
	return RemoteFileOutputWriter{
		OutputFilePaths: outputFilePaths,
		store:           store,
	}
}
