package plugin

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"google.golang.org/protobuf/proto"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/catalog"
	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
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

func (t *taskExecutionContext) ResourceManager() pluginsCore.ResourceManager {
	return t.resourceManager
}
func (t *taskExecutionContext) SecretManager() pluginsCore.SecretManager { return t.secretManager }
func (t *taskExecutionContext) TaskRefreshIndicator() pluginsCore.SignalAsync {
	return t.taskRefreshIndicator
}
func (t *taskExecutionContext) DataStore() *storage.DataStore { return t.dataStore }
func (t *taskExecutionContext) PluginStateReader() pluginsCore.PluginStateReader {
	return t.pluginStateReader
}
func (t *taskExecutionContext) PluginStateWriter() pluginsCore.PluginStateWriter {
	return t.pluginStateWriter
}
func (t *taskExecutionContext) TaskReader() pluginsCore.TaskReader { return t.taskReader }
func (t *taskExecutionContext) InputReader() io.InputReader        { return t.inputReader }
func (t *taskExecutionContext) TaskExecutionMetadata() pluginsCore.TaskExecutionMetadata {
	return t.taskExecMetadata
}
func (t *taskExecutionContext) OutputWriter() io.OutputWriter { return t.outputWriter }
func (t *taskExecutionContext) Catalog() catalog.AsyncClient  { return t.catalogClient }

// ComputeActionOutputPath constructs the full output directory for a task action:
//
//	<scheme>://<bucket>/<shard>/<rest-of-runOutputBase>/<actionName>/<attempt>
//
// The shard is a 2-char base-36 prefix derived deterministically from the
// TaskAction's namespace and name. It is inserted immediately after the bucket
// (AWS S3 hot-spot avoidance convention) so that storage traffic is spread
// across key prefixes from the root. The attempt segment isolates each retry.
func ComputeActionOutputPath(ctx context.Context, namespace, name, runOutputBase, actionName string, attempt uint32) (storage.DataReference, error) {
	sharder, err := ioutils.NewBase36PrefixShardSelector(ctx)
	if err != nil {
		return "", err
	}
	shard, err := sharder.GetShardPrefix(ctx, []byte(namespace+"/"+name))
	if err != nil {
		return "", err
	}

	u, err := url.Parse(runOutputBase)
	if err != nil {
		return "", fmt.Errorf("invalid RunOutputBase %q: %w", runOutputBase, err)
	}

	// Insert shard between bucket and the rest of the path:
	//   s3://bucket/org/proj/domain/run/ → s3://bucket/<shard>/org/proj/domain/run/<action>/<attempt>
	restPath := strings.Trim(u.Path, "/")
	segments := []string{shard}
	if restPath != "" {
		segments = append(segments, restPath)
	}
	segments = append(segments, actionName, strconv.Itoa(int(attempt)))
	u.Path = "/" + strings.Join(segments, "/")

	return storage.DataReference(u.String()), nil
}

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

	// Output writer — scope outputs per action and attempt so retries don't overwrite each other.
	// Path: <RunOutputBase>/<shard>/<ActionName>/<attempt>/
	attempt := taskAction.Status.Attempts
	if attempt == 0 { // if attempts is not set, default to 1
		attempt = 1
	}
	outputPrefix, err := ComputeActionOutputPath(ctx, taskAction.Namespace, taskAction.Name, taskAction.Spec.RunOutputBase, taskAction.Spec.ActionName, attempt)
	if err != nil {
		return nil, err
	}
	rawOutputPaths := ioutils.NewRawOutputPaths(ctx, outputPrefix)
	outputFilePaths := ioutils.NewCheckpointRemoteFilePaths(ctx, dataStore, outputPrefix, rawOutputPaths, "")
	outputWriter := ioutils.NewRemoteFileOutputWriter(ctx, dataStore, outputFilePaths)

	// Task execution metadata
	taskExecMeta, err := NewTaskExecutionMetadata(taskAction)
	if err != nil {
		return nil, err
	}

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
