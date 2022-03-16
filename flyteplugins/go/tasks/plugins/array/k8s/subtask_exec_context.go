package k8s

import (
	"context"
	"fmt"
	"strconv"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/array"
	podPlugin "github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/pod"
)

// SubTaskExecutionContext wraps the core TaskExecutionContext so that the k8s array task context
// can be used within the pod plugin
type SubTaskExecutionContext struct {
	pluginsCore.TaskExecutionContext
	arrayInputReader io.InputReader
	metadataOverride pluginsCore.TaskExecutionMetadata
	originalIndex    int
	subtaskReader    SubTaskReader
}

// InputReader overrides the base TaskExecutionContext to return a custom InputReader
func (s SubTaskExecutionContext) InputReader() io.InputReader {
	return s.arrayInputReader
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

// newSubtaskExecutionContext constructs a SubTaskExecutionContext using the provided parameters
func newSubTaskExecutionContext(tCtx pluginsCore.TaskExecutionContext, taskTemplate *core.TaskTemplate,
	executionIndex, originalIndex int, retryAttempt uint64) SubTaskExecutionContext {

	arrayInputReader := array.GetInputReader(tCtx, taskTemplate)
	taskExecutionMetadata := tCtx.TaskExecutionMetadata()
	taskExecutionID := taskExecutionMetadata.GetTaskExecutionID()
	metadataOverride := SubTaskExecutionMetadata{
		taskExecutionMetadata,
		SubTaskExecutionID{
			taskExecutionID,
			executionIndex,
			taskExecutionID.GetGeneratedName(),
			retryAttempt,
			taskExecutionID.GetID().RetryAttempt,
		},
	}

	subtaskTemplate := &core.TaskTemplate{}
	*subtaskTemplate = *taskTemplate

	if subtaskTemplate != nil {
		subtaskTemplate.TaskTypeVersion = 2
		if subtaskTemplate.GetContainer() != nil {
			subtaskTemplate.Type = podPlugin.ContainerTaskType
		} else if taskTemplate.GetK8SPod() != nil {
			subtaskTemplate.Type = podPlugin.SidecarTaskType
		}
	}

	subtaskReader := SubTaskReader{tCtx.TaskReader(), subtaskTemplate}

	return SubTaskExecutionContext{
		TaskExecutionContext: tCtx,
		arrayInputReader:     arrayInputReader,
		metadataOverride:     metadataOverride,
		originalIndex:        originalIndex,
		subtaskReader:        subtaskReader,
	}
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

// SubTaskExecutionMetadata wraps the core TaskExecutionMetadata to customize the TaskExecutionID
type SubTaskExecutionMetadata struct {
	pluginsCore.TaskExecutionMetadata
	subtaskExecutionID SubTaskExecutionID
}

// GetTaskExecutionID overrides the base TaskExecutionMetadata to return a custom TaskExecutionID
func (s SubTaskExecutionMetadata) GetTaskExecutionID() pluginsCore.TaskExecutionID {
	return s.subtaskExecutionID
}
