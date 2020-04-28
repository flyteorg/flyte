package executors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type immExecContext struct {
	ImmutableExecutionContext
}

type tdGetter struct {
	TaskDetailsGetter
}

type subWfGetter struct {
	SubWorkflowGetter
}

func TestExecutionContext(t *testing.T) {
	eCtx := immExecContext{}
	taskGetter := tdGetter{}
	subWfGetter := subWfGetter{}

	ec := NewExecutionContext(eCtx, taskGetter, subWfGetter)
	assert.NotNil(t, ec)
	typed := ec.(execContext)
	assert.Equal(t, typed.ImmutableExecutionContext, eCtx)
	assert.Equal(t, typed.SubWorkflowGetter, subWfGetter)
	assert.Equal(t, typed.TaskDetailsGetter, taskGetter)

	taskGetter2 := tdGetter{}
	NewExecutionContextWithTasksGetter(ec, taskGetter2)
	assert.NotNil(t, ec)
	typed = ec.(execContext)
	assert.Equal(t, typed.ImmutableExecutionContext, eCtx)
	assert.Equal(t, typed.SubWorkflowGetter, subWfGetter)
	assert.Equal(t, typed.TaskDetailsGetter, taskGetter2)

	subWfGetter2 := subWfGetter
	NewExecutionContextWithWorkflowGetter(ec, subWfGetter2)
	assert.NotNil(t, ec)
	typed = ec.(execContext)
	assert.Equal(t, typed.ImmutableExecutionContext, eCtx)
	assert.Equal(t, typed.SubWorkflowGetter, subWfGetter2)
	assert.Equal(t, typed.TaskDetailsGetter, taskGetter)

}
