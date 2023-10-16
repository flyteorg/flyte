package flytek8s

import (
	pluginsCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
)

// Wraps a regular TaskExecutionMetadata and overrides the IsInterruptible method to always return false
// This is useful as the runner and the scheduler pods should never be interruptible
type NonInterruptibleTaskExecutionMetadata struct {
	pluginsCore.TaskExecutionMetadata
}

func (n NonInterruptibleTaskExecutionMetadata) IsInterruptible() bool {
	return false
}

// A wrapper around a regular TaskExecutionContext allowing to inject a custom TaskExecutionMetadata which is
// non-interruptible
type NonInterruptibleTaskExecutionContext struct {
	pluginsCore.TaskExecutionContext
	metadata NonInterruptibleTaskExecutionMetadata
}

func (n NonInterruptibleTaskExecutionContext) TaskExecutionMetadata() pluginsCore.TaskExecutionMetadata {
	return n.metadata
}

func NewNonInterruptibleTaskExecutionContext(ctx pluginsCore.TaskExecutionContext) NonInterruptibleTaskExecutionContext {
	return NonInterruptibleTaskExecutionContext{
		TaskExecutionContext: ctx,
		metadata: NonInterruptibleTaskExecutionMetadata{
			ctx.TaskExecutionMetadata(),
		},
	}
}
