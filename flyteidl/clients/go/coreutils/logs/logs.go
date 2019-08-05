// Contains interface definitions and helper functions for log plugins

package logs

import (
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
)

type LogPlugin interface {
	// Generates a TaskLog object given necessary computation information
	GetTaskLog(podName, namespace, containerName, containerID, logName string) (core.TaskLog, error)
}
