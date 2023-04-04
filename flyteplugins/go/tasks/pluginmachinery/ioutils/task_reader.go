package ioutils

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flytestdlib/atomic"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/pkg/errors"
)

var (
	_ pluginsCore.TaskReader = &lazyUploadingTaskReader{}
)

// SimpleTaskReader provides only the TaskReader interface. This is created to conveniently use the uploading taskreader
// interface
type SimpleTaskReader interface {
	Read(ctx context.Context) (*core.TaskTemplate, error)
}

// lazyUploadingTaskReader provides a lazy interface that uploads the core.TaskTemplate to a configured location,
// only if the location is accessed. This reduces the potential overhead of writing the template
type lazyUploadingTaskReader struct {
	SimpleTaskReader
	uploaded   atomic.Bool
	store      storage.ProtobufStore
	remotePath storage.DataReference
}

func (r *lazyUploadingTaskReader) Path(ctx context.Context) (storage.DataReference, error) {
	// We are using atomic because it is ok to re-upload in some cases. We know that most of the plugins are
	// executed in a single go-routine, so chances of a race condition are minimal.
	if !r.uploaded.Load() {
		t, err := r.SimpleTaskReader.Read(ctx)
		if err != nil {
			return "", err
		}
		err = r.store.WriteProtobuf(ctx, r.remotePath, storage.Options{}, t)
		if err != nil {
			return "", errors.Wrapf(err, "failed to store task template to remote path [%s]", r.remotePath)
		}
		r.uploaded.Store(true)
	}
	return r.remotePath, nil
}

// NewLazyUploadingTaskReader decorates an existing TaskReader and adds a functionality to allow lazily uploading the task template to
// a remote location, only when the location information is accessed
func NewLazyUploadingTaskReader(baseTaskReader SimpleTaskReader, remotePath storage.DataReference, store storage.ProtobufStore) pluginsCore.TaskReader {
	return &lazyUploadingTaskReader{
		SimpleTaskReader: baseTaskReader,
		uploaded:         atomic.NewBool(false),
		store:            store,
		remotePath:       remotePath,
	}
}
