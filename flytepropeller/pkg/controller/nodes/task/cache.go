package task

import (
	"context"

	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/ioutils"
	errors2 "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/errors"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
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
		Identifier:           *taskTemplate.Id, //nolint:protogetter
		CacheVersion:         taskTemplate.GetMetadata().GetDiscoveryVersion(),
		CacheIgnoreInputVars: taskTemplate.GetMetadata().GetCacheIgnoreInputVars(),
		TypedInterface:       *taskTemplate.GetInterface(),
		InputReader:          nCtx.InputReader(),
	}, nil
}

func (t *Handler) IsCacheable(ctx context.Context, nCtx interfaces.NodeExecutionContext) (bool, bool, error) {
	// check if plugin has caching disabled
	ttype := nCtx.TaskReader().GetTaskType()
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
	tVersion := taskTemplate.GetTaskTypeVersion()
	ctx = contextutils.WithTaskType(ctx, ttype)
	p, err := t.ResolvePlugin(ctx, ttype, nCtx.ExecutionContext().GetExecutionConfig(), tVersion)
	if err != nil {
		return false, false, errors2.Wrapf(errors2.UnsupportedTaskTypeError, nCtx.NodeID(), err, "unable to resolve plugin")
	}

	checkCatalog := !p.GetProperties().DisableNodeLevelCaching
	if !checkCatalog {
		logger.Infof(ctx, "Node level caching is disabled. Skipping catalog read.")
		return false, false, nil
	}

	return taskTemplate.GetMetadata().GetDiscoverable(), taskTemplate.GetMetadata().GetDiscoverable() && taskTemplate.GetMetadata().GetCacheSerializable(), nil
}
