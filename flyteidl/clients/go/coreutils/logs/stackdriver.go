package logs

import (
	"fmt"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

// TL;DR Log links in Stackdriver for configured GCP project and log Resource
//
// This is a simple stackdriver log plugin that creates a preformatted log link for a given project and logResource
// using resource.labels.pod_name as advancedFilter
type stackdriverLogPlugin struct {
	// the name of the project in GCP that the logs are being published under
	gcpProject string
	// The Log resource name for which the logs are published under
	logResource string
}

func (s *stackdriverLogPlugin) GetTaskLog(podName, namespace, containerName, containerID, logName string) (core.TaskLog, error) {
	return core.TaskLog{
		Uri: fmt.Sprintf(
			"https://console.cloud.google.com/logs/viewer?project=%s&angularJsUrl=%%2Flogs%%2Fviewer%%3Fproject%%3D%s&resource=%s&advancedFilter=resource.labels.pod_name%%3D%s",
			s.gcpProject,
			s.gcpProject,
			s.logResource,
			podName,
		),
		Name:          logName,
		MessageFormat: core.TaskLog_JSON,
	}, nil
}

// Deprecated: Please use NewTemplateLogPlugin from github.com/lyft/flyteplugins/go/tasks/pluginmachinery/tasklog instead.
func NewStackdriverLogPlugin(gcpProject, logResource string) LogPlugin {
	return &stackdriverLogPlugin{
		gcpProject:  gcpProject,
		logResource: logResource,
	}
}
