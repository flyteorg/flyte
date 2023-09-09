package tasklog

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

func MustCreateRegex(varName string) *regexp.Regexp {
	return regexp.MustCompile(fmt.Sprintf(`(?i){{\s*[\.$]%s\s*}}`, varName))
}

type templateRegexes struct {
	LogName              *regexp.Regexp
	PodName              *regexp.Regexp
	PodUID               *regexp.Regexp
	Namespace            *regexp.Regexp
	ContainerName        *regexp.Regexp
	ContainerID          *regexp.Regexp
	Hostname             *regexp.Regexp
	PodRFC3339StartTime  *regexp.Regexp
	PodRFC3339FinishTime *regexp.Regexp
	PodUnixStartTime     *regexp.Regexp
	PodUnixFinishTime    *regexp.Regexp
	TaskID               *regexp.Regexp
	TaskVersion          *regexp.Regexp
	TaskProject          *regexp.Regexp
	TaskDomain           *regexp.Regexp
	TaskRetryAttempt     *regexp.Regexp
	NodeID               *regexp.Regexp
	ExecutionName        *regexp.Regexp
	ExecutionProject     *regexp.Regexp
	ExecutionDomain      *regexp.Regexp
}

func initDefaultRegexes() templateRegexes {
	return templateRegexes{
		MustCreateRegex("logName"),
		MustCreateRegex("podName"),
		MustCreateRegex("podUID"),
		MustCreateRegex("namespace"),
		MustCreateRegex("containerName"),
		MustCreateRegex("containerID"),
		MustCreateRegex("hostname"),
		MustCreateRegex("podRFC3339StartTime"),
		MustCreateRegex("podRFC3339FinishTime"),
		MustCreateRegex("podUnixStartTime"),
		MustCreateRegex("podUnixFinishTime"),
		MustCreateRegex("taskID"),
		MustCreateRegex("taskVersion"),
		MustCreateRegex("taskProject"),
		MustCreateRegex("taskDomain"),
		MustCreateRegex("taskRetryAttempt"),
		MustCreateRegex("nodeID"),
		MustCreateRegex("executionName"),
		MustCreateRegex("executionProject"),
		MustCreateRegex("executionDomain"),
	}
}

var defaultRegexes = initDefaultRegexes()

func replaceAll(template string, vars TemplateVars) string {
	for _, v := range vars {
		if len(v.Value) > 0 {
			template = v.Regex.ReplaceAllLiteralString(template, v.Value)
		}
	}
	return template
}

func (input Input) templateVarsForScheme(scheme TemplateScheme) TemplateVars {
	vars := TemplateVars{
		{defaultRegexes.LogName, input.LogName},
	}

	gotExtraTemplateVars := input.ExtraTemplateVarsByScheme != nil
	if gotExtraTemplateVars {
		vars = append(vars, input.ExtraTemplateVarsByScheme.Common...)
	}

	switch scheme {
	case TemplateSchemePod:
		// Container IDs are prefixed with docker://, cri-o://, etc. which is stripped by fluentd before pushing to a log
		// stream. Therefore, we must also strip the prefix.
		containerID := input.ContainerID
		stripDelimiter := "://"
		if split := strings.Split(input.ContainerID, stripDelimiter); len(split) > 1 {
			containerID = split[1]
		}
		vars = append(
			vars,
			TemplateVar{defaultRegexes.PodName, input.PodName},
			TemplateVar{defaultRegexes.PodUID, input.PodUID},
			TemplateVar{defaultRegexes.Namespace, input.Namespace},
			TemplateVar{defaultRegexes.ContainerName, input.ContainerName},
			TemplateVar{defaultRegexes.ContainerID, containerID},
			TemplateVar{defaultRegexes.Hostname, input.HostName},
			TemplateVar{defaultRegexes.PodRFC3339StartTime, input.PodRFC3339StartTime},
			TemplateVar{defaultRegexes.PodRFC3339FinishTime, input.PodRFC3339FinishTime},
			TemplateVar{
				defaultRegexes.PodUnixStartTime,
				strconv.FormatInt(input.PodUnixStartTime, 10),
			},
			TemplateVar{
				defaultRegexes.PodUnixFinishTime,
				strconv.FormatInt(input.PodUnixFinishTime, 10),
			},
		)
		if gotExtraTemplateVars {
			vars = append(vars, input.ExtraTemplateVarsByScheme.Pod...)
		}
	case TemplateSchemeTaskExecution:
		if input.TaskExecutionIdentifier != nil {
			vars = append(vars, TemplateVar{
				defaultRegexes.TaskRetryAttempt,
				strconv.FormatUint(uint64(input.TaskExecutionIdentifier.RetryAttempt), 10),
			})
			if input.TaskExecutionIdentifier.TaskId != nil {
				vars = append(
					vars,
					TemplateVar{
						defaultRegexes.TaskID,
						input.TaskExecutionIdentifier.TaskId.Name,
					},
					TemplateVar{
						defaultRegexes.TaskVersion,
						input.TaskExecutionIdentifier.TaskId.Version,
					},
					TemplateVar{
						defaultRegexes.TaskProject,
						input.TaskExecutionIdentifier.TaskId.Project,
					},
					TemplateVar{
						defaultRegexes.TaskDomain,
						input.TaskExecutionIdentifier.TaskId.Domain,
					},
				)
			}
			if input.TaskExecutionIdentifier.NodeExecutionId != nil {
				vars = append(vars, TemplateVar{
					defaultRegexes.NodeID,
					input.TaskExecutionIdentifier.NodeExecutionId.NodeId,
				})
				if input.TaskExecutionIdentifier.NodeExecutionId.ExecutionId != nil {
					vars = append(
						vars,
						TemplateVar{
							defaultRegexes.ExecutionName,
							input.TaskExecutionIdentifier.NodeExecutionId.ExecutionId.Name,
						},
						TemplateVar{
							defaultRegexes.ExecutionProject,
							input.TaskExecutionIdentifier.NodeExecutionId.ExecutionId.Project,
						},
						TemplateVar{
							defaultRegexes.ExecutionDomain,
							input.TaskExecutionIdentifier.NodeExecutionId.ExecutionId.Domain,
						},
					)
				}
			}
		}
		if gotExtraTemplateVars {
			vars = append(vars, input.ExtraTemplateVarsByScheme.TaskExecution...)
		}
	}

	return vars
}

// A simple log plugin that supports templates in urls to build the final log link.
// See `defaultRegexes` for supported templates.
type TemplateLogPlugin struct {
	scheme        TemplateScheme
	templateUris  []string
	messageFormat core.TaskLog_MessageFormat
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
	templateVars := input.templateVarsForScheme(s.scheme)
	taskLogs := make([]*core.TaskLog, 0, len(s.templateUris))
	for _, templateURI := range s.templateUris {
		taskLogs = append(taskLogs, &core.TaskLog{
			Uri:           replaceAll(templateURI, templateVars),
			Name:          input.LogName,
			MessageFormat: s.messageFormat,
		})
	}

	return Output{TaskLogs: taskLogs}, nil
}

// NewTemplateLogPlugin creates a template-based log plugin with the provided template Uri and message format.
// See `defaultRegexes` for supported templates.
func NewTemplateLogPlugin(scheme TemplateScheme, templateUris []string, messageFormat core.TaskLog_MessageFormat) TemplateLogPlugin {
	return TemplateLogPlugin{
		scheme:        scheme,
		templateUris:  templateUris,
		messageFormat: messageFormat,
	}
}
