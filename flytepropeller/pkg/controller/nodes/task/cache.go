package task

import (
	"context"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/ioutils"

	errors2 "github.com/flyteorg/flytepropeller/pkg/controller/nodes/errors"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/interfaces"

	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/logger"
)

func (t *Handler) GetCatalogKey(ctx context.Context, nCtx interfaces.NodeExecutionContext) (catalog.Key, error) {
	// read task template
	taskTemplatePath, err := ioutils.GetTaskTemplatePath(ctx, nCtx.DataStore(), nCtx.NodeStatus().GetDataDir())
	if err != nil {
		return catalog.Key{}, err
	}

	taskReader := ioutils.NewLazyUploadingTaskReader(nCtx.TaskReader(), taskTemplatePath, nCtx.DataStore())
	taskTemplate, err := taskReader.Read(ctx)
	if err != nil {
		logger.Errorf(ctx, "failed to read TaskTemplate, error :%s", err.Error())
		return catalog.Key{}, err
	}

	return catalog.Key{
		Identifier:     *taskTemplate.Id,
		CacheVersion:   taskTemplate.Metadata.DiscoveryVersion,
		TypedInterface: *taskTemplate.Interface,
		InputReader:    nCtx.InputReader(),
	}, nil
}

func (t *Handler) IsCacheable(ctx context.Context, nCtx interfaces.NodeExecutionContext) (bool, bool, error) {
	// check if plugin has caching disabled
	ttype := nCtx.TaskReader().GetTaskType()
	ctx = contextutils.WithTaskType(ctx, ttype)
	p, err := t.ResolvePlugin(ctx, ttype, nCtx.ExecutionContext().GetExecutionConfig())
	if err != nil {
		return false, false, errors2.Wrapf(errors2.UnsupportedTaskTypeError, nCtx.NodeID(), err, "unable to resolve plugin")
	}

	checkCatalog := !p.GetProperties().DisableNodeLevelCaching
	if !checkCatalog {
		logger.Infof(ctx, "Node level caching is disabled. Skipping catalog read.")
		return false, false, nil
	}

	// read task template
	taskTemplatePath, err := ioutils.GetTaskTemplatePath(ctx, nCtx.DataStore(), nCtx.NodeStatus().GetDataDir())
	if err != nil {
		return false, false, err
	}

	taskReader := ioutils.NewLazyUploadingTaskReader(nCtx.TaskReader(), taskTemplatePath, nCtx.DataStore())
	taskTemplate, err := taskReader.Read(ctx)
	if err != nil {
		logger.Errorf(ctx, "failed to read TaskTemplate, error :%s", err.Error())
		return false, false, err
	}

	return taskTemplate.Metadata.Discoverable, taskTemplate.Metadata.Discoverable && taskTemplate.Metadata.CacheSerializable, nil
}
