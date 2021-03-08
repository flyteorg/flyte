package catalog

import (
	"context"
	"fmt"
	"reflect"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/workqueue"
)

type WriterWorkItem struct {
	// WriterWorkItem Inputs
	key      Key
	data     io.OutputReader
	metadata Metadata
}

func NewWriterWorkItem(key Key, data io.OutputReader, metadata Metadata) *WriterWorkItem {
	return &WriterWorkItem{
		key:      key,
		data:     data,
		metadata: metadata,
	}
}

type writerProcessor struct {
	catalogClient Client
}

func (p writerProcessor) Process(ctx context.Context, workItem workqueue.WorkItem) (workqueue.WorkStatus, error) {
	wi, casted := workItem.(*WriterWorkItem)
	if !casted {
		return workqueue.WorkStatusNotDone, fmt.Errorf("wrong work item type. Received: %v", reflect.TypeOf(workItem))
	}

	status, err := p.catalogClient.Put(ctx, wi.key, wi.data, wi.metadata)
	if err != nil {
		logger.Errorf(ctx, "Error putting to catalog [%s]", err)
		return workqueue.WorkStatusNotDone, errors.Wrapf(errors.DownstreamSystemError, err,
			"Error writing to catalog, key id [%v] cache version [%v]",
			wi.key.Identifier, wi.key.CacheVersion)
	}

	if status.GetCacheStatus() == core.CatalogCacheStatus_CACHE_PUT_FAILURE {
		return workqueue.WorkStatusNotDone, errors.Errorf(errors.DownstreamSystemError,
			"Error writing to catalog, key id [%v] cache version [%v]",
			wi.key.Identifier, wi.key.CacheVersion)
	}

	logger.Debugf(ctx, "Successfully wrote to catalog. Key [%v]", wi.key)
	return workqueue.WorkStatusSucceeded, nil
}

func NewWriterProcessor(catalogClient Client) workqueue.Processor {
	return writerProcessor{
		catalogClient: catalogClient,
	}
}
