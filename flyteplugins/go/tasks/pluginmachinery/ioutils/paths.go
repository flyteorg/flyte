package ioutils

import (
	"context"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flytestdlib/storage"
)

const (
	InputsSuffix      = "inputs.pb"
	FuturesSuffix     = "futures.pb"
	OutputsSuffix     = "outputs.pb"
	ErrorsSuffix      = "error.pb"
	IndexLookupSuffix = "indexlookup.pb"
)

func constructPath(store storage.ReferenceConstructor, base storage.DataReference, suffix string) storage.DataReference {
	res, err := store.ConstructReference(context.Background(), base, suffix)
	if err != nil {
		logger.Error(context.Background(), "Failed to construct path. Base[%v] Error: %v", base, err)
	}

	return res
}

func GetPath(ctx context.Context, store storage.ReferenceConstructor, root storage.DataReference, subNames ...string) (res storage.DataReference, err error) {
	return store.ConstructReference(ctx, root, subNames...)
}

func GetMasterOutputsPath(ctx context.Context, store storage.ReferenceConstructor, output storage.DataReference) (res storage.DataReference, err error) {
	return store.ConstructReference(ctx, output, OutputsSuffix)
}

func GetInputsPath(ctx context.Context, store storage.ReferenceConstructor, prefix storage.DataReference) (res storage.DataReference, err error) {
	return store.ConstructReference(ctx, prefix, InputsSuffix)
}

func GetOutputsPath(ctx context.Context, store storage.ReferenceConstructor, prefix storage.DataReference) (res storage.DataReference, err error) {
	return store.ConstructReference(ctx, prefix, OutputsSuffix)
}

func GetFuturesPath(ctx context.Context, store storage.ReferenceConstructor, prefix storage.DataReference) (res storage.DataReference, err error) {
	return store.ConstructReference(ctx, prefix, FuturesSuffix)
}

func GetErrorsPath(ctx context.Context, store storage.ReferenceConstructor, prefix storage.DataReference) (res storage.DataReference, err error) {
	return store.ConstructReference(ctx, prefix, ErrorsSuffix)
}

func GetIndexLookupPath(ctx context.Context, store storage.ReferenceConstructor, prefix storage.DataReference) (res storage.DataReference, err error) {
	return store.ConstructReference(ctx, prefix, IndexLookupSuffix)
}
