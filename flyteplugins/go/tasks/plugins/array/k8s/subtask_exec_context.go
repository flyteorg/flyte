package k8s

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/tasklog"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils/secrets"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array"
	podPlugin "github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/pod"

	"github.com/flyteorg/flytestdlib/storage"
)

// SubTaskExecutionContext wraps the core TaskExecutionContext so that the k8s array task context
// can be used within the pod plugin
type SubTaskExecutionContext struct {
	pluginsCore.TaskExecutionContext
	arrayInputReader io.InputReader
	metadataOverride pluginsCore.TaskExecutionMetadata
	originalIndex    int
	outputWriter     io.OutputWriter
	subtaskReader    SubTaskReader
}

// InputReader overrides the base TaskExecutionContext to return a custom InputReader
func (s SubTaskExecutionContext) InputReader() io.InputReader {
	return s.arrayInputReader
}

// OutputWriter overrides the base TaskExecutionContext to return a custom OutputWriter
func (s SubTaskExecutionContext) OutputWriter() io.OutputWriter {
	return s.outputWriter
}

// TaskExecutionMetadata overrides the base TaskExecutionContext to return custom
// TaskExecutionMetadata
func (s SubTaskExecutionContext) TaskExecutionMetadata() pluginsCore.TaskExecutionMetadata {
	return s.metadataOverride
}

// TaskReader overrides the base TaskExecutionContext to return a custom TaskReader
func (s SubTaskExecutionContext) TaskReader() pluginsCore.TaskReader {
	return s.subtaskReader
}

// pluginStateReader overrides the default PluginStateReader because the maptask does not persist
// state between task evaluations.
type pluginStateReader struct{}

func (p pluginStateReader) GetStateVersion() uint8 {
	return 0
}

func (p pluginStateReader) Get(t interface{}) (stateVersion uint8, err error) {
	return 0, nil
}

// PluginStateReader overrides the default behavior to return a custom pluginStateReader.
func (s SubTaskExecutionContext) PluginStateReader() pluginsCore.PluginStateReader {
	return pluginStateReader{}
}

// NewSubtaskExecutionContext constructs a SubTaskExecutionContext using the provided parameters
func NewSubTaskExecutionContext(ctx context.Context, tCtx pluginsCore.TaskExecutionContext, taskTemplate *core.TaskTemplate,
	executionIndex, originalIndex int, retryAttempt uint64, systemFailures uint64) (SubTaskExecutionContext, error) {

	subTaskExecutionMetadata, err := NewSubTaskExecutionMetadata(tCtx.TaskExecutionMetadata(), taskTemplate, executionIndex, retryAttempt, systemFailures)
	if err != nil {
		return SubTaskExecutionContext{}, err
	}

	// construct TaskTemplate
	subtaskTemplate := &core.TaskTemplate{}
	*subtaskTemplate = *taskTemplate

	subtaskTemplate.TaskTypeVersion = 2
	if subtaskTemplate.GetContainer() != nil {
		subtaskTemplate.Type = podPlugin.ContainerTaskType
	} else if taskTemplate.GetK8SPod() != nil {
		subtaskTemplate.Type = podPlugin.SidecarTaskType
	}

	arrayInputReader := array.GetInputReader(tCtx, taskTemplate)
	subtaskReader := SubTaskReader{tCtx.TaskReader(), subtaskTemplate}

	// construct OutputWriter
	dataStore := tCtx.DataStore()
	checkpointPrefix, err := dataStore.ConstructReference(ctx, tCtx.OutputWriter().GetRawOutputPrefix(), strconv.Itoa(originalIndex))
	if err != nil {
		return SubTaskExecutionContext{}, err
	}

	checkpoint, err := dataStore.ConstructReference(ctx, checkpointPrefix, strconv.FormatUint(retryAttempt, 10))
	if err != nil {
		return SubTaskExecutionContext{}, err
	}
	checkpointPath := ioutils.NewRawOutputPaths(ctx, checkpoint)

	var prevCheckpoint storage.DataReference
	if retryAttempt == 0 {
		prevCheckpoint = ""
	} else {
		prevCheckpoint, err = dataStore.ConstructReference(ctx, checkpointPrefix, strconv.FormatUint(retryAttempt-1, 10))
		if err != nil {
			return SubTaskExecutionContext{}, err
		}
	}
	prevCheckpointPath := ioutils.ConstructCheckpointPath(dataStore, prevCheckpoint)

	// note that we must not append the originalIndex to the original OutputPrefixPath because
	// flytekit is already doing this
	p := ioutils.NewCheckpointRemoteFilePaths(ctx, dataStore, tCtx.OutputWriter().GetOutputPrefixPath(), checkpointPath, prevCheckpointPath)
	outputWriter := ioutils.NewRemoteFileOutputWriter(ctx, dataStore, p)

	return SubTaskExecutionContext{
		TaskExecutionContext: tCtx,
		arrayInputReader:     arrayInputReader,
		metadataOverride:     subTaskExecutionMetadata,
		originalIndex:        originalIndex,
		outputWriter:         outputWriter,
		subtaskReader:        subtaskReader,
	}, nil
}

// SubTaskReader wraps the core TaskReader to customize the task template task type and version
type SubTaskReader struct {
	pluginsCore.TaskReader
	subtaskTemplate *core.TaskTemplate
}

// Read overrides the base TaskReader to return a custom TaskTemplate
func (s SubTaskReader) Read(ctx context.Context) (*core.TaskTemplate, error) {
	return s.subtaskTemplate, nil
}

// SubTaskExecutionID wraps the core TaskExecutionID to customize the generated pod name
type SubTaskExecutionID struct {
	pluginsCore.TaskExecutionID
	executionIndex      int
	parentName          string
	subtaskRetryAttempt uint64
	taskRetryAttempt    uint32
}

// GetGeneratedName overrides the base TaskExecutionID to append the subtask index and retryAttempt
func (s SubTaskExecutionID) GetGeneratedName() string {
	indexStr := strconv.Itoa(s.executionIndex)

	// If the retryAttempt is 0 we do not include it in the pod name. The gives us backwards
	// compatibility in the ability to dynamically transition running map tasks to use subtask retries.
	if s.subtaskRetryAttempt == 0 {
		return utils.ConvertToDNS1123SubdomainCompatibleString(fmt.Sprintf("%v-%v", s.parentName, indexStr))
	}

	retryAttemptStr := strconv.FormatUint(s.subtaskRetryAttempt, 10)
	return utils.ConvertToDNS1123SubdomainCompatibleString(fmt.Sprintf("%v-%v-%v", s.parentName, indexStr, retryAttemptStr))
}

// GetLogSuffix returns the suffix which should be appended to subtask log names
func (s SubTaskExecutionID) GetLogSuffix() string {
	// Append the retry attempt and executionIndex so that log names coincide with pod names per
	// https://github.com/flyteorg/flyteplugins/pull/186#discussion_r666569825. To maintain
	// backwards compatibility we append the subtaskRetryAttempt if it is not 0.
	if s.subtaskRetryAttempt == 0 {
		return fmt.Sprintf(" #%d-%d", s.taskRetryAttempt, s.executionIndex)
	}

	return fmt.Sprintf(" #%d-%d-%d", s.taskRetryAttempt, s.executionIndex, s.subtaskRetryAttempt)
}

var logTemplateRegexes = struct {
	ExecutionIndex     *regexp.Regexp
	ParentName         *regexp.Regexp
	RetryAttempt       *regexp.Regexp
	ParentRetryAttempt *regexp.Regexp
}{
	tasklog.MustCreateRegex("subtaskExecutionIndex"),
	tasklog.MustCreateRegex("subtaskParentName"),
	tasklog.MustCreateRegex("subtaskRetryAttempt"),
	tasklog.MustCreateRegex("subtaskParentRetryAttempt"),
}

func (s SubTaskExecutionID) TemplateVarsByScheme() *tasklog.TemplateVarsByScheme {
	return &tasklog.TemplateVarsByScheme{
		TaskExecution: tasklog.TemplateVars{
			{Regex: logTemplateRegexes.ParentName, Value: s.parentName},
			{
				Regex: logTemplateRegexes.ExecutionIndex,
				Value: strconv.FormatUint(uint64(s.executionIndex), 10),
			},
			{
				Regex: logTemplateRegexes.RetryAttempt,
				Value: strconv.FormatUint(s.subtaskRetryAttempt, 10),
			},
			{
				Regex: logTemplateRegexes.ParentRetryAttempt,
				Value: strconv.FormatUint(uint64(s.taskRetryAttempt), 10),
			},
		},
	}
}

// NewSubtaskExecutionID constructs a SubTaskExecutionID using the provided parameters
func NewSubTaskExecutionID(taskExecutionID pluginsCore.TaskExecutionID, executionIndex int, retryAttempt uint64) SubTaskExecutionID {
	return SubTaskExecutionID{
		taskExecutionID,
		executionIndex,
		taskExecutionID.GetGeneratedName(),
		retryAttempt,
		taskExecutionID.GetID().RetryAttempt,
	}
}

// SubTaskExecutionMetadata wraps the core TaskExecutionMetadata to customize the TaskExecutionID
type SubTaskExecutionMetadata struct {
	pluginsCore.TaskExecutionMetadata
	annotations        map[string]string
	labels             map[string]string
	interruptible      bool
	subtaskExecutionID SubTaskExecutionID
}

// GetAnnotations overrides the base TaskExecutionMetadata to return a custom map
func (s SubTaskExecutionMetadata) GetAnnotations() map[string]string {
	return s.annotations
}

// GetLabels overrides the base TaskExecutionMetadata to return a custom map
func (s SubTaskExecutionMetadata) GetLabels() map[string]string {
	return s.labels
}

// GetTaskExecutionID overrides the base TaskExecutionMetadata to return a custom TaskExecutionID
func (s SubTaskExecutionMetadata) GetTaskExecutionID() pluginsCore.TaskExecutionID {
	return s.subtaskExecutionID
}

// IsInterruptbile overrides the base NodeExecutionMetadata to return a subtask specific identifier
func (s SubTaskExecutionMetadata) IsInterruptible() bool {
	return s.interruptible
}

// NewSubtaskExecutionMetadata constructs a SubTaskExecutionMetadata using the provided parameters
func NewSubTaskExecutionMetadata(taskExecutionMetadata pluginsCore.TaskExecutionMetadata, taskTemplate *core.TaskTemplate,
	executionIndex int, retryAttempt uint64, systemFailures uint64) (SubTaskExecutionMetadata, error) {

	var err error
	secretsMap := make(map[string]string)
	injectSecretsLabel := make(map[string]string)
	if taskTemplate.SecurityContext != nil && len(taskTemplate.SecurityContext.Secrets) > 0 {
		secretsMap, err = secrets.MarshalSecretsToMapStrings(taskTemplate.SecurityContext.Secrets)
		if err != nil {
			return SubTaskExecutionMetadata{}, err
		}

		injectSecretsLabel = map[string]string{
			secrets.PodLabel: secrets.PodLabelValue,
		}
	}

	subTaskExecutionID := NewSubTaskExecutionID(taskExecutionMetadata.GetTaskExecutionID(), executionIndex, retryAttempt)
	interruptible := taskExecutionMetadata.IsInterruptible() && uint32(systemFailures) < taskExecutionMetadata.GetInterruptibleFailureThreshold()
	return SubTaskExecutionMetadata{
		taskExecutionMetadata,
		utils.UnionMaps(taskExecutionMetadata.GetAnnotations(), secretsMap),
		utils.UnionMaps(taskExecutionMetadata.GetLabels(), injectSecretsLabel),
		interruptible,
		subTaskExecutionID,
	}, nil
}
