package ioutils

import (
	"context"
	"fmt"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyte/flytestdlib/errors"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

const (
	ErrFailedRead string = "READ_FAILED"
)

var (
	// Ensure we get an early build break if interface changes and these classes do not conform.
	_ io.InputFilePaths = &SimpleInputFilePath{}
	_ io.InputReader    = RemoteFileInputReader{}
)

type RemoteFileInputReader struct {
	io.InputFilePaths
	store storage.ProtobufStore
}

func (r RemoteFileInputReader) Get(ctx context.Context) (*core.InputData, error) {
	d := &core.InputData{}
	oldFormat := &core.LiteralMap{}
	if msgIndex, err := r.store.ReadProtobufAny(ctx, r.InputFilePaths.GetInputDataPath(), d, oldFormat); err != nil {
		// TODO change flytestdlib to return protobuf unmarshal errors separately. As this can indicate malformed output and we should catch that
		return nil, errors.Wrapf(ErrFailedRead, err, "failed to read data from dataDir [%v].", r.InputFilePaths.GetInputDataPath())
	} else if msgIndex == 1 {
		d.Inputs = oldFormat
	}

	return d, nil

}

func NewRemoteFileInputReader(_ context.Context, store storage.ProtobufStore, inputPaths io.InputFilePaths) RemoteFileInputReader {
	return RemoteFileInputReader{
		InputFilePaths: inputPaths,
		store:          store,
	}
}

type SimpleInputFilePath struct {
	pathPrefix storage.DataReference
	store      *storage.DataStore
}

func (s *SimpleInputFilePath) GetInputPrefixPath() storage.DataReference {
	return s.pathPrefix
}

func (s *SimpleInputFilePath) GetInputPath(ctx context.Context) (storage.DataReference, error) {
	oldInputPath := constructPath(s.store, s.pathPrefix, InputsSuffix)
	newInput := &core.InputData{}
	err := s.store.ReadProtobuf(ctx, s.GetInputDataPath(), newInput)
	if err != nil {
		return "", fmt.Errorf("failed to read existing inputs from [%v]. Error: %w", s.GetInputDataPath(), err)
	}

	err = s.store.WriteProtobuf(ctx, oldInputPath, storage.Options{}, newInput.GetInputs())
	if err != nil {
		return "", fmt.Errorf("failed to write inputs to [%v]. Error: %w", oldInputPath, err)
	}

	return oldInputPath, nil
}

func (s *SimpleInputFilePath) GetInputDataPath() storage.DataReference {
	return constructPath(s.store, s.pathPrefix, InputDataSuffix)
}

func NewInputFilePaths(_ context.Context, store *storage.DataStore, inputPathPrefix storage.DataReference) *SimpleInputFilePath {
	return &SimpleInputFilePath{
		store:      store,
		pathPrefix: inputPathPrefix,
	}
}
