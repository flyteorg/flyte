package plugin

import (
	"context"
	"strings"

	"google.golang.org/protobuf/proto"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

var _ pluginsCore.TaskExecutionContext = &taskExecutionContext{}

type taskExecutionContext struct {
	resourceManager      pluginsCore.ResourceManager
	secretManager        pluginsCore.SecretManager
	taskRefreshIndicator pluginsCore.SignalAsync
	dataStore            *storage.DataStore
	pluginStateReader    pluginsCore.PluginStateReader
	pluginStateWriter    pluginsCore.PluginStateWriter
	taskReader           pluginsCore.TaskReader
	inputReader          io.InputReader
	taskExecMetadata     pluginsCore.TaskExecutionMetadata
	outputWriter         io.OutputWriter
	catalogClient        catalog.AsyncClient
}

func (t *taskExecutionContext) ResourceManager() pluginsCore.ResourceManager             { return t.resourceManager }
func (t *taskExecutionContext) SecretManager() pluginsCore.SecretManager                 { return t.secretManager }
func (t *taskExecutionContext) TaskRefreshIndicator() pluginsCore.SignalAsync             { return t.taskRefreshIndicator }
func (t *taskExecutionContext) DataStore() *storage.DataStore                            { return t.dataStore }
func (t *taskExecutionContext) PluginStateReader() pluginsCore.PluginStateReader          { return t.pluginStateReader }
func (t *taskExecutionContext) PluginStateWriter() pluginsCore.PluginStateWriter          { return t.pluginStateWriter }
func (t *taskExecutionContext) TaskReader() pluginsCore.TaskReader                        { return t.taskReader }
func (t *taskExecutionContext) InputReader() io.InputReader                               { return t.inputReader }
func (t *taskExecutionContext) TaskExecutionMetadata() pluginsCore.TaskExecutionMetadata  { return t.taskExecMetadata }
func (t *taskExecutionContext) OutputWriter() io.OutputWriter                             { return t.outputWriter }
func (t *taskExecutionContext) Catalog() catalog.AsyncClient                              { return t.catalogClient }

// inlineTaskReader reads a TaskTemplate from bytes stored inline in the CRD.
type inlineTaskReader struct {
	data []byte
}

func (r *inlineTaskReader) Path(_ context.Context) (storage.DataReference, error) {
	return "inline://taskTemplate", nil
}

func (r *inlineTaskReader) Read(_ context.Context) (*core.TaskTemplate, error) {
	t := &core.TaskTemplate{}
	if err := proto.Unmarshal(r.data, t); err != nil {
		return nil, err
	}
	return t, nil
}

// NewTaskExecutionContext creates a TaskExecutionContext from a TaskAction and supporting dependencies.
func NewTaskExecutionContext(
	taskAction *flyteorgv1.TaskAction,
	dataStore *storage.DataStore,
	pluginStateMgr *PluginStateManager,
	secretManager pluginsCore.SecretManager,
	resourceManager pluginsCore.ResourceManager,
	catalogClient catalog.AsyncClient,
) (*taskExecutionContext, error) {
	ctx := context.Background()

	// Task reader (inline from CRD)
	taskReader := &inlineTaskReader{
		data: taskAction.Spec.TaskTemplate,
	}

	// Input reader — InputURI may be a full path (ending in inputs.pb) or just
	// the prefix directory.  NewInputFilePaths always appends "inputs.pb", so
	// strip a trailing suffix to avoid the doubled "inputs.pb/inputs.pb" path.
	inputPathPrefix := storage.DataReference(
		strings.TrimSuffix(taskAction.Spec.InputURI, "/"+ioutils.InputsSuffix),
	)
	inputPaths := ioutils.NewInputFilePaths(ctx, dataStore, inputPathPrefix)
	inputReader := ioutils.NewRemoteFileInputReader(ctx, dataStore, inputPaths)

	// Output writer — scope outputs per action so actions don't overwrite each other.
	// RunOutputBase is run-level (e.g. s3://bucket/org/proj/domain/run/),
	// append the action name to make it action-specific.
	outputPrefix := storage.DataReference(
		strings.TrimRight(taskAction.Spec.RunOutputBase, "/") + "/" + taskAction.Spec.ActionName,
	)
	rawOutputPaths := ioutils.NewRawOutputPaths(ctx, outputPrefix)
	outputFilePaths := ioutils.NewCheckpointRemoteFilePaths(ctx, dataStore, outputPrefix, rawOutputPaths, "")
	outputWriter := ioutils.NewRemoteFileOutputWriter(ctx, dataStore, outputFilePaths)

	// Task execution metadata
	taskExecMeta := NewTaskExecutionMetadata(taskAction)

	return &taskExecutionContext{
		resourceManager:      resourceManager,
		secretManager:        secretManager,
		taskRefreshIndicator: func(_ context.Context) {},
		dataStore:            dataStore,
		pluginStateReader:    pluginStateMgr,
		pluginStateWriter:    pluginStateMgr,
		taskReader:           taskReader,
		inputReader:          inputReader,
		taskExecMetadata:     taskExecMeta,
		outputWriter:         outputWriter,
		catalogClient:        catalogClient,
	}, nil
}
