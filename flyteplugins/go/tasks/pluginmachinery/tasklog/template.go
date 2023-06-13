package tasklog

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

// A simple log plugin that supports templates in urls to build the final log link. Supported templates are:
// {{ .podName }}: Gets the pod name as it shows in k8s dashboard,
// {{ .podUID }}: Gets the pod UID,
// {{ .namespace }}: K8s namespace where the pod runs,
// {{ .containerName }}: The container name that generated the log,
// {{ .containerId }}: The container id docker/crio generated at run time,
// {{ .logName }}: A deployment specific name where to expect the logs to be.
// {{ .hostname }}: The hostname where the pod is running and where logs reside.
// {{ .PodRFC3339StartTime }}: The pod creation time in RFC3339 format
// {{ .PodRFC3339FinishTime }}: Don't have a good mechanism for this yet, but approximating with time.Now for now
// {{ .podUnixStartTime }}: The pod creation time (in unix seconds, not millis)
// {{ .podUnixFinishTime }}: Don't have a good mechanism for this yet, but approximating with time.Now for now
type TemplateLogPlugin struct {
	templateUris  []string
	messageFormat core.TaskLog_MessageFormat
}

type regexValPair struct {
	regex *regexp.Regexp
	val   string
}

type templateRegexes struct {
	PodName              *regexp.Regexp
	PodUID               *regexp.Regexp
	Namespace            *regexp.Regexp
	ContainerName        *regexp.Regexp
	ContainerID          *regexp.Regexp
	LogName              *regexp.Regexp
	Hostname             *regexp.Regexp
	PodRFC3339StartTime  *regexp.Regexp
	PodRFC3339FinishTime *regexp.Regexp
	PodUnixStartTime     *regexp.Regexp
	PodUnixFinishTime    *regexp.Regexp
}

func mustInitTemplateRegexes() templateRegexes {
	return templateRegexes{
		PodName:              mustCreateRegex("podName"),
		PodUID:               mustCreateRegex("podUID"),
		Namespace:            mustCreateRegex("namespace"),
		ContainerName:        mustCreateRegex("containerName"),
		ContainerID:          mustCreateRegex("containerID"),
		LogName:              mustCreateRegex("logName"),
		Hostname:             mustCreateRegex("hostname"),
		PodRFC3339StartTime:  mustCreateRegex("podRFC3339StartTime"),
		PodRFC3339FinishTime: mustCreateRegex("podRFC3339FinishTime"),
		PodUnixStartTime:     mustCreateRegex("podUnixStartTime"),
		PodUnixFinishTime:    mustCreateRegex("podUnixFinishTime"),
	}
}

var regexes = mustInitTemplateRegexes()

func mustCreateRegex(varName string) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf(`(?i){{\s*[\.$]%s\s*}}`, varName))
}

func replaceAll(template string, values []regexValPair) string {
	for _, v := range values {
		template = v.regex.ReplaceAllString(template, v.val)
	}

	return template
}

func (s TemplateLogPlugin) GetTaskLog(podName, podUID, namespace, containerName, containerID, logName string, podRFC3339StartTime string, podRFC3339FinishTime string, podUnixStartTime, podUnixFinishTime int64) (core.TaskLog, error) {
	o, err := s.GetTaskLogs(Input{
		LogName:              logName,
		Namespace:            namespace,
		PodName:              podName,
		PodUID:               podUID,
		ContainerName:        containerName,
		ContainerID:          containerID,
		PodRFC3339StartTime:  podRFC3339StartTime,
		PodRFC3339FinishTime: podRFC3339FinishTime,
		PodUnixStartTime:     podUnixStartTime,
		PodUnixFinishTime:    podUnixFinishTime,
	})

	if err != nil || len(o.TaskLogs) == 0 {
		return core.TaskLog{}, err
	}

	return *o.TaskLogs[0], nil
}

func (s TemplateLogPlugin) GetTaskLogs(input Input) (Output, error) {
	// Container IDs are prefixed with docker://, cri-o://, etc. which is stripped by fluentd before pushing to a log
	// stream. Therefore, we must also strip the prefix.
	containerID := input.ContainerID
	stripDelimiter := "://"
	if split := strings.Split(input.ContainerID, stripDelimiter); len(split) > 1 {
		containerID = split[1]
	}

	taskLogs := make([]*core.TaskLog, 0, len(s.templateUris))
	for _, templateURI := range s.templateUris {
		taskLogs = append(taskLogs, &core.TaskLog{
			Uri: replaceAll(
				templateURI,
				[]regexValPair{
					{
						regex: regexes.PodName,
						val:   input.PodName,
					},
					{
						regex: regexes.PodUID,
						val:   input.PodUID,
					},
					{
						regex: regexes.Namespace,
						val:   input.Namespace,
					},
					{
						regex: regexes.ContainerName,
						val:   input.ContainerName,
					},
					{
						regex: regexes.ContainerID,
						val:   containerID,
					},
					{
						regex: regexes.LogName,
						val:   input.LogName,
					},
					{
						regex: regexes.Hostname,
						val:   input.HostName,
					},
					{
						regex: regexes.PodRFC3339StartTime,
						val:   input.PodRFC3339StartTime,
					},
					{
						regex: regexes.PodRFC3339FinishTime,
						val:   input.PodRFC3339FinishTime,
					},
					{
						regex: regexes.PodUnixStartTime,
						val:   strconv.FormatInt(input.PodUnixStartTime, 10),
					},
					{
						regex: regexes.PodUnixFinishTime,
						val:   strconv.FormatInt(input.PodUnixFinishTime, 10),
					},
				},
			),
			Name:          input.LogName,
			MessageFormat: s.messageFormat,
		})
	}

	return Output{
		TaskLogs: taskLogs,
	}, nil
}

// NewTemplateLogPlugin creates a template-based log plugin with the provided template Uri and message format. Supported
// templates are:
// {{ .podName }}: Gets the pod name as it shows in k8s dashboard,
// {{ .podUID }}: Gets the pod UID,
// {{ .namespace }}: K8s namespace where the pod runs,
// {{ .containerName }}: The container name that generated the log,
// {{ .containerId }}: The container id docker/crio generated at run time,
// {{ .logName }}: A deployment specific name where to expect the logs to be.
// {{ .hostname }}: The hostname where the pod is running and where logs reside.
// {{ .PodRFC3339StartTime }}: The pod creation time in RFC3339 format
// {{ .PodRFC3339FinishTime }}: Don't have a good mechanism for this yet, but approximating with time.Now for now
// {{ .podUnixStartTime }}: The pod creation time (in unix seconds, not millis)
// {{ .podUnixFinishTime }}: Don't have a good mechanism for this yet, but approximating with time.Now for now
func NewTemplateLogPlugin(templateUris []string, messageFormat core.TaskLog_MessageFormat) TemplateLogPlugin {
	return TemplateLogPlugin{
		templateUris:  templateUris,
		messageFormat: messageFormat,
	}
}
