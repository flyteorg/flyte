package tasklog

import (
	"regexp"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
)

//go:generate enumer --type=TemplateScheme --trimprefix=TemplateScheme -json -yaml

type TemplateScheme int

const (
	TemplateSchemePod TemplateScheme = iota
	TemplateSchemeTaskExecution
)

type TemplateVar struct {
	Regex *regexp.Regexp
	Value string
}

type TemplateVars []TemplateVar

type TemplateVarsByScheme struct {
	Common        TemplateVars
	Pod           TemplateVars
	TaskExecution TemplateVars
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
	TaskExecutionIdentifier   *core.TaskExecutionIdentifier
	ExtraTemplateVarsByScheme *TemplateVarsByScheme
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
