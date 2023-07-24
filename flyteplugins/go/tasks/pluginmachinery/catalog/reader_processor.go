package catalog

import (
	"context"
	"fmt"
	"reflect"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/flyteorg/flyteplugins/go/tasks/errors"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/workqueue"
)

type ReaderWorkItem struct {
	// ReaderWorkItem outputs:
	cached bool

	// ReaderWorkItem Inputs:
	outputsWriter io.OutputWriter
	// Inputs to query data catalog
	key Key
}

func (item ReaderWorkItem) IsCached() bool {
	return item.cached
}

func NewReaderWorkItem(key Key, outputsWriter io.OutputWriter) *ReaderWorkItem {
	return &ReaderWorkItem{
		key:           key,
		outputsWriter: outputsWriter,
	}
}

type ReaderProcessor struct {
	catalogClient Client
}

func (p ReaderProcessor) Process(ctx context.Context, workItem workqueue.WorkItem) (workqueue.WorkStatus, error) {
	wi, casted := workItem.(*ReaderWorkItem)
	if !casted {
		return workqueue.WorkStatusNotDone, fmt.Errorf("wrong work item type. Received: %v", reflect.TypeOf(workItem))
	}

	op, err := p.catalogClient.Get(ctx, wi.key)
	if err != nil {
		if IsNotFound(err) {
			logger.Infof(ctx, "Artifact not found in Catalog. Key: %v", wi.key)
			wi.cached = false
			return workqueue.WorkStatusSucceeded, nil
		}

		err = errors.Wrapf("CausedBy", err, "Failed to call catalog for Key: %v.", wi.key)
		logger.Warnf(ctx, "Cache call failed: %v", err)
		return workqueue.WorkStatusFailed, err
	}

	if op.status.GetCacheStatus() == core.CatalogCacheStatus_CACHE_LOOKUP_FAILURE {
		return workqueue.WorkStatusFailed, errors.Errorf(errors.DownstreamSystemError, "failed to lookup cache")
	}

	if op.status.GetCacheStatus() == core.CatalogCacheStatus_CACHE_MISS || op.GetOutputs() == nil {
		wi.cached = false
		return workqueue.WorkStatusSucceeded, nil
	}

	// TODO: Check task interface, if it has outputs but literalmap is empty (or not matching output), error.
	logger.Debugf(ctx, "Persisting output to %v", wi.outputsWriter.GetOutputPath())
	err = wi.outputsWriter.Put(ctx, op.GetOutputs())
	if err != nil {
		err = errors.Wrapf("CausedBy", err, "Failed to persist cached output for Key: %v.", wi.key)
		logger.Warnf(ctx, "Cache write to output writer failed: %v", err)
		return workqueue.WorkStatusFailed, err
	}

	wi.cached = true

	logger.Debugf(ctx, "Successfully read from catalog. Key [%v]", wi.key)
	return workqueue.WorkStatusSucceeded, nil
}

func NewReaderProcessor(catalogClient Client) ReaderProcessor {
	return ReaderProcessor{
		catalogClient: catalogClient,
	}
}
