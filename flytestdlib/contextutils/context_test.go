package contextutils

import (
	"context"
	"runtime/pprof"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithAppName(t *testing.T) {
	ctx := context.Background()
	assert.Nil(t, ctx.Value(AppNameKey))
	ctx = WithAppName(ctx, "application-name-123")
	assert.Equal(t, "application-name-123", ctx.Value(AppNameKey))

	ctx = WithAppName(ctx, "app-name-modified")
	assert.Equal(t, "app-name-modified", ctx.Value(AppNameKey))
}

func TestWithPhase(t *testing.T) {
	ctx := context.Background()
	assert.Nil(t, ctx.Value(PhaseKey))
	ctx = WithPhase(ctx, "Running")
	assert.Equal(t, "Running", ctx.Value(PhaseKey))

	ctx = WithPhase(ctx, "Failed")
	assert.Equal(t, "Failed", ctx.Value(PhaseKey))
}

func TestWithJobId(t *testing.T) {
	ctx := context.Background()
	assert.Nil(t, ctx.Value(JobIDKey))
	ctx = WithJobID(ctx, "job123")
	assert.Equal(t, "job123", ctx.Value(JobIDKey))

	ctx = WithJobID(ctx, "subtask")
	assert.Equal(t, "job123/subtask", ctx.Value(JobIDKey))
}

func TestWithNamespace(t *testing.T) {
	ctx := context.Background()
	assert.Nil(t, ctx.Value(NamespaceKey))
	ctx = WithNamespace(ctx, "flyte")
	assert.Equal(t, "flyte", ctx.Value(NamespaceKey))

	ctx = WithNamespace(ctx, "flyte2")
	assert.Equal(t, "flyte2", ctx.Value(NamespaceKey))
}

func TestWithExecutionID(t *testing.T) {
	ctx := context.Background()
	assert.Nil(t, ctx.Value(ExecIDKey))
	ctx = WithExecutionID(ctx, "job123")
	assert.Equal(t, "job123", ctx.Value(ExecIDKey))
}

func TestWithTaskType(t *testing.T) {
	ctx := context.Background()
	assert.Nil(t, ctx.Value(TaskTypeKey))
	ctx = WithTaskType(ctx, "flyte")
	assert.Equal(t, "flyte", ctx.Value(TaskTypeKey))
}

func TestWithWorkflowID(t *testing.T) {
	ctx := context.Background()
	assert.Nil(t, ctx.Value(WorkflowIDKey))
	ctx = WithWorkflowID(ctx, "flyte")
	assert.Equal(t, "flyte", ctx.Value(WorkflowIDKey))
}

func TestWithLaunchPlanID(t *testing.T) {
	ctx := context.Background()
	assert.Nil(t, ctx.Value(LaunchPlanIDKey))
	ctx = WithLaunchPlanID(ctx, "flytelp")
	assert.Equal(t, "flytelp", ctx.Value(LaunchPlanIDKey))
}

func TestWithNodeID(t *testing.T) {
	ctx := context.Background()
	assert.Nil(t, ctx.Value(NodeIDKey))
	ctx = WithNodeID(ctx, "n1")
	assert.Equal(t, "n1", ctx.Value(NodeIDKey))

	ctx = WithNodeID(ctx, "n2")
	assert.Equal(t, "n1/n2", ctx.Value(NodeIDKey))
}

func TestWithProjectDomain(t *testing.T) {
	ctx := context.Background()
	assert.Nil(t, ctx.Value(ProjectKey))
	assert.Nil(t, ctx.Value(DomainKey))
	ctx = WithProjectDomain(ctx, "proj", "domain")
	assert.Equal(t, "proj", ctx.Value(ProjectKey))
	assert.Equal(t, "domain", ctx.Value(DomainKey))
}

func TestWithTaskID(t *testing.T) {
	ctx := context.Background()
	assert.Nil(t, ctx.Value(TaskIDKey))
	ctx = WithTaskID(ctx, "task")
	assert.Equal(t, "task", ctx.Value(TaskIDKey))
}

func TestWithSignalID(t *testing.T) {
	ctx := context.Background()
	assert.Nil(t, ctx.Value(SignalIDKey))
	ctx = WithSignalID(ctx, "signal")
	assert.Equal(t, "signal", ctx.Value(SignalIDKey))
}

func TestGetFields(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(WithJobID(WithNamespace(ctx, "ns123"), "job123"), "req123")
	m := GetLogFields(ctx)
	assert.Equal(t, "ns123", m[NamespaceKey.String()])
	assert.Equal(t, "job123", m[JobIDKey.String()])
	assert.Equal(t, "req123", m[RequestIDKey.String()])
}

func TestValues(t *testing.T) {
	ctx := context.Background()
	ctx = WithWorkflowID(ctx, "flyte")
	m := Values(ctx, ProjectKey, WorkflowIDKey)
	assert.NotNil(t, m)
	assert.Equal(t, 2, len(m))
	assert.Equal(t, "flyte", m[WorkflowIDKey.String()])
	assert.Equal(t, "", m[ProjectKey.String()])
}

func TestWithGoroutineLabel(t *testing.T) {
	ctx := context.Background()
	ctx = WithGoroutineLabel(ctx, "my_routine_123")
	pprof.SetGoroutineLabels(ctx)
	m := Values(ctx, RoutineLabelKey)
	assert.Equal(t, 1, len(m))
	assert.Equal(t, "my_routine_123", m[RoutineLabelKey.String()])
}
