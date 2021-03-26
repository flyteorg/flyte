package ioutils

import (
	"context"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flytestdlib/storage"
)

const (
	InputsSuffix       = "inputs.pb"
	TaskTemplateSuffix = "task.pb"
	FuturesSuffix      = "futures.pb"
	OutputsSuffix      = "outputs.pb"
	ErrorsSuffix       = "error.pb"
	IndexLookupSuffix  = "indexlookup.pb"
)

func constructPath(store storage.ReferenceConstructor, base storage.DataReference, suffix string) storage.DataReference {
	res, err := store.ConstructReference(context.Background(), base, suffix)
	if err != nil {
		logger.Error(context.Background(), "Failed to construct path. Base[%v] Error: %v", base, err)
	}

	return res
}

// GetTaskTemplatePath returns a protobuf file path where TaskTemplate is stored
func GetTaskTemplatePath(ctx context.Context, store storage.ReferenceConstructor, base storage.DataReference) (storage.DataReference, error) {
	return store.ConstructReference(ctx, base, TaskTemplateSuffix)
}

// GetIndexLookupPath returns the indexpath suffixed to IndexLookupSuffix
func GetIndexLookupPath(ctx context.Context, store storage.ReferenceConstructor, prefix storage.DataReference) (res storage.DataReference, err error) {
	return store.ConstructReference(ctx, prefix, IndexLookupSuffix)
}
