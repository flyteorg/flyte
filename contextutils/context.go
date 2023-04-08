// Package contextutils contains common flyte context utils.
package contextutils

import (
	"context"
	"fmt"
	"runtime/pprof"

	"google.golang.org/grpc/metadata"
)

type Key string

const (
	AppNameKey         Key = "app_name"
	NamespaceKey       Key = "ns"
	TaskTypeKey        Key = "tasktype"
	ProjectKey         Key = "project"
	DomainKey          Key = "domain"
	WorkflowIDKey      Key = "wf"
	NodeIDKey          Key = "node"
	TaskIDKey          Key = "task"
	ExecIDKey          Key = "exec_id"
	JobIDKey           Key = "job_id"
	PhaseKey           Key = "phase"
	RoutineLabelKey    Key = "routine"
	LaunchPlanIDKey    Key = "lp"
	ResourceVersionKey Key = "res_ver"
	SignalIDKey        Key = "signal"
	RequestIDKey       Key = "x-request-id"
)

func (k Key) String() string {
	return string(k)
}

var logKeys = []Key{
	AppNameKey,
	JobIDKey,
	NamespaceKey,
	ExecIDKey,
	NodeIDKey,
	WorkflowIDKey,
	TaskTypeKey,
	PhaseKey,
	RoutineLabelKey,
	LaunchPlanIDKey,
	ResourceVersionKey,
	RequestIDKey,
}

// MetricKeysFromStrings is a convenience method to convert a slice of strings into a slice of Keys
func MetricKeysFromStrings(keys []string) []Key {
	res := make([]Key, 0, len(keys))

	for _, k := range keys {
		res = append(res, Key(k))
	}

	return res
}

// WithResourceVersion gets a new context with the resource version set.
func WithResourceVersion(ctx context.Context, resourceVersion string) context.Context {
	return context.WithValue(ctx, ResourceVersionKey, resourceVersion)
}

// WithNamespace gets a new context with namespace set.
func WithNamespace(ctx context.Context, namespace string) context.Context {
	return context.WithValue(ctx, NamespaceKey, namespace)
}

// WithJobID gets a new context with JobId set. If the existing context already has a job id, the new context will have
// <old_jobID>/<new_jobID> set as the job id.
func WithJobID(ctx context.Context, jobID string) context.Context {
	existingJobID := ctx.Value(JobIDKey)
	if existingJobID != nil {
		jobID = fmt.Sprintf("%v/%v", existingJobID, jobID)
	}

	return context.WithValue(ctx, JobIDKey, jobID)
}

// WithAppName gets a new context with AppName set.
func WithAppName(ctx context.Context, appName string) context.Context {
	return context.WithValue(ctx, AppNameKey, appName)
}

// WithPhase gets a new context with Phase set.
func WithPhase(ctx context.Context, phase string) context.Context {
	return context.WithValue(ctx, PhaseKey, phase)
}

// WithExecutionID gets a new context with ExecutionID set.
func WithExecutionID(ctx context.Context, execID string) context.Context {
	return context.WithValue(ctx, ExecIDKey, execID)
}

// WithNodeID gets a new context with NodeID (nested) set.
func WithNodeID(ctx context.Context, nodeID string) context.Context {
	existingNodeID := ctx.Value(NodeIDKey)
	if existingNodeID != nil {
		nodeID = fmt.Sprintf("%v/%v", existingNodeID, nodeID)
	}
	return context.WithValue(ctx, NodeIDKey, nodeID)
}

// WithWorkflowID gets a new context with WorkflowName set.
func WithWorkflowID(ctx context.Context, workflow string) context.Context {
	return context.WithValue(ctx, WorkflowIDKey, workflow)
}

// WithLaunchPlanID gets a new context with a launch plan ID set.
func WithLaunchPlanID(ctx context.Context, launchPlan string) context.Context {
	return context.WithValue(ctx, LaunchPlanIDKey, launchPlan)
}

// WithProjectDomain gets new context with Project and Domain values set
func WithProjectDomain(ctx context.Context, project, domain string) context.Context {
	c := context.WithValue(ctx, ProjectKey, project)
	return context.WithValue(c, DomainKey, domain)
}

// WithTaskID gets a new context with WorkflowName set.
func WithTaskID(ctx context.Context, taskID string) context.Context {
	return context.WithValue(ctx, TaskIDKey, taskID)
}

// WithTaskType gets a new context with TaskType set.
func WithTaskType(ctx context.Context, taskType string) context.Context {
	return context.WithValue(ctx, TaskTypeKey, taskType)
}

// WithSignalID gets a new context with SignalID set.
func WithSignalID(ctx context.Context, signalID string) context.Context {
	return context.WithValue(ctx, SignalIDKey, signalID)
}

// WithRequestID gets a new context with RequestID set.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return metadata.AppendToOutgoingContext(context.WithValue(ctx, RequestIDKey, requestID), RequestIDKey.String(), requestID)
}

// WithGoroutineLabel gets a new context with Go Routine label key set and a label assigned to the context using
// pprof.Labels.
// You can then call pprof.SetGoroutineLabels(ctx) to annotate the current go-routine and have that show up in
// pprof analysis.
func WithGoroutineLabel(ctx context.Context, routineLabel string) context.Context {
	ctx = pprof.WithLabels(ctx, pprof.Labels(RoutineLabelKey.String(), routineLabel))
	return context.WithValue(ctx, RoutineLabelKey, routineLabel)
}

func addFieldIfNotNil(ctx context.Context, m map[string]interface{}, fieldKey Key) {
	val := ctx.Value(fieldKey)
	if val != nil {
		m[fieldKey.String()] = val
	}
}

func addStringFieldWithDefaults(ctx context.Context, m map[string]string, fieldKey Key) {
	val := ctx.Value(fieldKey)
	if val == nil {
		val = ""
	}

	m[fieldKey.String()] = val.(string)
}

// GetLogFields gets a map of all known logKeys set on the context. logKeys are special and should be used incase,
// context fields are to be added to the log lines.
func GetLogFields(ctx context.Context) map[string]interface{} {
	res := map[string]interface{}{}
	for _, k := range logKeys {
		addFieldIfNotNil(ctx, res, k)
	}

	return res
}

func Value(ctx context.Context, key Key) string {
	val := ctx.Value(key)
	if val != nil {
		return val.(string)
	}

	return ""
}

func Values(ctx context.Context, keys ...Key) map[string]string {
	res := map[string]string{}
	for _, k := range keys {
		addStringFieldWithDefaults(ctx, res, k)
	}

	return res
}
