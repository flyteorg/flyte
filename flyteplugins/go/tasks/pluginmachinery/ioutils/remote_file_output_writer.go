package ioutils

import (
	"context"
	"fmt"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/lyft/flytestdlib/storage"
)

type RemoteFileOutputWriter struct {
	io.OutputFilePaths
	store storage.ProtobufStore
}

type RemoteFileOutputPaths struct {
	outputPrefix storage.DataReference
	store        storage.ReferenceConstructor
}

var (
	_ io.OutputWriter    = RemoteFileOutputWriter{}
	_ io.OutputFilePaths = RemoteFileOutputPaths{}
)

func (w RemoteFileOutputPaths) GetOutputPrefixPath() storage.DataReference {
	return w.outputPrefix
}

func (w RemoteFileOutputPaths) GetOutputPath() storage.DataReference {
	return constructPath(w.store, w.outputPrefix, OutputsSuffix)
}

func (w RemoteFileOutputPaths) GetErrorPath() storage.DataReference {
	return constructPath(w.store, w.outputPrefix, ErrorsSuffix)
}

func (w RemoteFileOutputPaths) GetFuturesPath() storage.DataReference {
	return constructPath(w.store, w.outputPrefix, FuturesSuffix)
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

func NewRemoteFileOutputPaths(_ context.Context, store storage.ReferenceConstructor, outputPrefix storage.DataReference) RemoteFileOutputPaths {
	return RemoteFileOutputPaths{
		store:        store,
		outputPrefix: outputPrefix,
	}
}

func NewRemoteFileOutputWriter(_ context.Context, store storage.ProtobufStore, outputFilePaths io.OutputFilePaths) RemoteFileOutputWriter {
	return RemoteFileOutputWriter{
		OutputFilePaths: outputFilePaths,
		store:           store,
	}
}
