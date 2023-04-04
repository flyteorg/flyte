package tasklog

import "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

// Input contains all available information about task's execution that a log plugin can use to construct task's
// log links.
type Input struct {
	HostName          string `json:"hostname"`
	PodName           string `json:"podName"`
	Namespace         string `json:"namespace"`
	ContainerName     string `json:"containerName"`
	ContainerID       string `json:"containerId"`
	LogName           string `json:"logName"`
	PodUnixStartTime  int64  `json:"podUnixStartTime"`
	PodUnixFinishTime int64  `json:"podUnixFinishTime"`
	PodUID            string `json:"podUID"`
}

// Output contains all task logs a plugin generates for a given Input.
type Output struct {
	TaskLogs []*core.TaskLog `json:"taskLogs"`
}

// Plugin represents an interface for task log plugins to implement to plug generated task log links into task events.
type Plugin interface {
	// Generates a TaskLog object given necessary computation information
	GetTaskLogs(input Input) (logs Output, err error)
}
