// Contains interface definitions and helper functions for log plugins

package logs

import (
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

// Deprecated: Please use Plugin interface from github.com/lyft/flyteplugins/go/tasks/pluginmachinery/tasklog instead.
type LogPlugin interface {
	// Generates a TaskLog object given necessary computation information
	GetTaskLog(podName, namespace, containerName, containerID, logName string) (core.TaskLog, error)
}
