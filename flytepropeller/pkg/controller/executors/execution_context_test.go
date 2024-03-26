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

type immutableParentInfo struct {
	ImmutableParentInfo
}

func TestExecutionContext(t *testing.T) {
	eCtx := immExecContext{}
	taskGetter := tdGetter{}
	subWfGetter := subWfGetter{}
	immutableParentInfo := immutableParentInfo{}

	ec := NewExecutionContext(eCtx, taskGetter, subWfGetter, immutableParentInfo, InitializeControlFlow())
	assert.NotNil(t, ec)
	typed := ec.(execContext)
	assert.Equal(t, typed.ImmutableExecutionContext, eCtx)
	assert.Equal(t, typed.SubWorkflowGetter, subWfGetter)
	assert.Equal(t, typed.TaskDetailsGetter, taskGetter)
	assert.Equal(t, typed.GetParentInfo(), immutableParentInfo)

	taskGetter2 := tdGetter{}
	NewExecutionContextWithTasksGetter(ec, taskGetter2)
	assert.NotNil(t, ec)
	typed = ec.(execContext)
	assert.Equal(t, typed.ImmutableExecutionContext, eCtx)
	assert.Equal(t, typed.SubWorkflowGetter, subWfGetter)
	assert.Equal(t, typed.TaskDetailsGetter, taskGetter2)
	assert.Equal(t, typed.GetParentInfo(), immutableParentInfo)

	subWfGetter2 := subWfGetter
	NewExecutionContextWithWorkflowGetter(ec, subWfGetter2)
	assert.NotNil(t, ec)
	typed = ec.(execContext)
	assert.Equal(t, typed.ImmutableExecutionContext, eCtx)
	assert.Equal(t, typed.SubWorkflowGetter, subWfGetter2)
	assert.Equal(t, typed.TaskDetailsGetter, taskGetter)
	assert.Equal(t, typed.GetParentInfo(), immutableParentInfo)

	immutableParentInfo2 := immutableParentInfo
	NewExecutionContextWithParentInfo(ec, immutableParentInfo2)
	assert.NotNil(t, ec)
	typed = ec.(execContext)
	assert.Equal(t, typed.ImmutableExecutionContext, eCtx)
	assert.Equal(t, typed.SubWorkflowGetter, subWfGetter2)
	assert.Equal(t, typed.TaskDetailsGetter, taskGetter)
	assert.Equal(t, typed.GetParentInfo(), immutableParentInfo2)
}

func TestParentExecutionInfo_GetUniqueID(t *testing.T) {
	expectedID := "testID"
	parentInfo := NewParentInfo(expectedID, 1)
	assert.Equal(t, expectedID, parentInfo.GetUniqueID())
}

func TestParentExecutionInfo_CurrentAttempt(t *testing.T) {
	expectedAttempt := uint32(123465)
	parentInfo := NewParentInfo("testID", expectedAttempt)
	assert.Equal(t, expectedAttempt, parentInfo.CurrentAttempt())
}

func TestControlFlow_ControlFlowParallelism(t *testing.T) {
	cFlow := InitializeControlFlow().(*controlFlow)
	assert.Equal(t, uint32(0), cFlow.CurrentParallelism())
	cFlow.IncrementParallelism()
	assert.Equal(t, uint32(1), cFlow.CurrentParallelism())
	cFlow.IncrementParallelism()
	assert.Equal(t, uint32(2), cFlow.CurrentParallelism())
}

func TestNewParentInfo(t *testing.T) {
	expectedID := "testID"
	expectedAttempt := uint32(123465)
	parentInfo := NewParentInfo(expectedID, expectedAttempt).(*parentExecutionInfo)
	assert.Equal(t, expectedID, parentInfo.uniqueID)
	assert.Equal(t, expectedAttempt, parentInfo.currentAttempts)
}
