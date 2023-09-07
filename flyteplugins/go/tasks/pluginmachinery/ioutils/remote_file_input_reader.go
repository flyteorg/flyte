package ioutils

import (
	"context"

	"github.com/flyteorg/flytestdlib/errors"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/storage"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"
)

const (
	ErrFailedRead string = "READ_FAILED"
)

var (
	// Ensure we get an early build break if interface changes and these classes do not conform.
	_ io.InputFilePaths = SimpleInputFilePath{}
	_ io.InputReader    = RemoteFileInputReader{}
)

type RemoteFileInputReader struct {
	io.InputFilePaths
	store storage.ProtobufStore
}

func (r RemoteFileInputReader) Get(ctx context.Context) (*core.LiteralMap, error) {
	d := &core.LiteralMap{}
	if err := r.store.ReadProtobuf(ctx, r.InputFilePaths.GetInputPath(), d); err != nil {
		// TODO change flytestdlib to return protobuf unmarshal errors separately. As this can indicate malformed output and we should catch that
		return nil, errors.Wrapf(ErrFailedRead, err, "failed to read data from dataDir [%v].", r.InputFilePaths.GetInputPath())
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
	store      storage.ReferenceConstructor
}

func (s SimpleInputFilePath) GetInputPrefixPath() storage.DataReference {
	return s.pathPrefix
}

func (s SimpleInputFilePath) GetInputPath() storage.DataReference {
	return constructPath(s.store, s.pathPrefix, InputsSuffix)
}

func NewInputFilePaths(_ context.Context, store storage.ReferenceConstructor, inputPathPrefix storage.DataReference) SimpleInputFilePath {
	return SimpleInputFilePath{
		store:      store,
		pathPrefix: inputPathPrefix,
	}
}
