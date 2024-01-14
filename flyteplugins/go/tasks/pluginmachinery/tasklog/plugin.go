package tasklog

import (
	"regexp"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	pluginsCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
)

//go:generate enumer --type=TemplateScheme --trimprefix=TemplateScheme -json -yaml

type TemplateScheme int

const (
	TemplateSchemePod TemplateScheme = iota
	TemplateSchemeTaskExecution
	TemplateSchemeFlyin
)

// TemplateURI is a URI that accepts templates. See: go/tasks/pluginmachinery/tasklog/template.go for available templates.
type TemplateURI = string

type TemplateVar struct {
	Regex *regexp.Regexp
	Value string
}

type TemplateVars []TemplateVar

type TemplateVarsByScheme struct {
	Common        TemplateVars
	Pod           TemplateVars
	TaskExecution TemplateVars
	Flyin         TemplateVars
}

// Input contains all available information about task's execution that a log plugin can use to construct task's
// log links.
type Input struct {
	HostName                  string
	PodName                   string
	Namespace                 string
	ContainerName             string
	ContainerID               string
	LogName                   string
	PodRFC3339StartTime       string
	PodRFC3339FinishTime      string
	PodUnixStartTime          int64
	PodUnixFinishTime         int64
	PodUID                    string
	TaskExecutionID           pluginsCore.TaskExecutionID
	ExtraTemplateVarsByScheme *TemplateVarsByScheme
	TaskTemplate              *core.TaskTemplate
}

// Output contains all task logs a plugin generates for a given Input.
type Output struct {
	TaskLogs []*core.TaskLog
}

// Plugin represents an interface for task log plugins to implement to plug generated task log links into task events.
type Plugin interface {
	// Generates a TaskLog object given necessary computation information
	GetTaskLogs(i Input) (logs Output, err error)
}

type TemplateLogPlugin struct {
	DisplayName      string                     `json:"displayName" pflag:",Display name for the generated log when displayed in the console."`
	TemplateURIs     []TemplateURI              `json:"templateUris" pflag:",URI Templates for generating task log links."`
	MessageFormat    core.TaskLog_MessageFormat `json:"messageFormat" pflag:",Log Message Format."`
	Scheme           TemplateScheme             `json:"scheme" pflag:",Templating scheme to use. Supported values are Pod and TaskExecution."`
	ShowWhilePending bool                       `json:"showWhilePending" pflag:",If true, the log link will be shown even if the task is in a pending state."`
}
