package ioutils

import (
	"context"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
)

type precomputedRawOutputPaths struct {
	path storage.DataReference
}

func (r precomputedRawOutputPaths) GetRawOutputPrefix() storage.DataReference {
	return r.path
}

// A simple Output sandbox at a given path
func NewRawOutputPaths(_ context.Context, rawOutputPrefix storage.DataReference) io.RawOutputPaths {
	return precomputedRawOutputPaths{path: rawOutputPrefix}
}
