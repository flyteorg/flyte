package ioutils

import (
	"context"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flytestdlib/storage"
)

const (
	// InputsSuffix specifies the name of the file that contains the task inputs in the form core.LiteralMap
	InputsSuffix = "inputs.pb"
	// TaskTemplateSuffix In case a task requests for a task template, it is passed into the task using this filename.
	// The format is of type core.TaskTemplate
	TaskTemplateSuffix = "task.pb"
	// FuturesSuffix specifies that for dynamic workflows, the futures files is written with this filename/suffix.
	// The format is core.DynamicJobSpec
	FuturesSuffix = "futures.pb"
	// OutputsSuffix specifies that outputs are assumed to be written to this "file"/"suffix" under the given prefix
	// The outputs file has a format of core.LiteralMap
	OutputsSuffix = "outputs.pb"
	// deckSuffix specifies that deck file are assumed to be written to this "file"/"suffix" under the given prefix
	// The deck file has a format of HTML
	deckSuffix = "deck.html"
	// ErrorsSuffix specifies that the errors are written to this prefix/file under the given prefix. The Error File
	// has a format of core.ErrorDocument
	ErrorsSuffix      = "error.pb"
	IndexLookupSuffix = "indexlookup.pb"
	// CheckpointPrefix specifies the common prefix that can be used as a directory where all the checkpoint information
	// will be stored. This directory is under the raw output-prefix path
	CheckpointPrefix = "_flytecheckpoints"
)

// ConstructCheckpointPath returns a checkpoint path under the given `base` / rawOutputPrefix, following the conventions of
// the store
func ConstructCheckpointPath(store storage.ReferenceConstructor, rawOutputPrefix storage.DataReference) storage.DataReference {
	if len(rawOutputPrefix) == 0 {
		return ""
	}
	return constructPath(store, rawOutputPrefix, CheckpointPrefix)
}

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
