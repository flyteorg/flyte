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
	parentInfo := NewParentInfo(expectedID, 1, false)
	assert.Equal(t, expectedID, parentInfo.GetUniqueID())
}

func TestParentExecutionInfo_CurrentAttempt(t *testing.T) {
	expectedAttempt := uint32(123465)
	parentInfo := NewParentInfo("testID", expectedAttempt, false)
	assert.Equal(t, expectedAttempt, parentInfo.CurrentAttempt())
}

func TestParentExecutionInfo_DynamicChain(t *testing.T) {
	expectedAttempt := uint32(123465)
	parentInfo := NewParentInfo("testID", expectedAttempt, true)
	assert.True(t, parentInfo.IsInDynamicChain())
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
	parentInfo := NewParentInfo(expectedID, expectedAttempt, false).(*parentExecutionInfo)
	assert.Equal(t, expectedID, parentInfo.uniqueID)
	assert.Equal(t, expectedAttempt, parentInfo.currentAttempts)
}

func TestNewVisitedNodes(t *testing.T) {
	vn := NewVisitedNodes()
	assert.NotNil(t, vn)
}

func TestVisitedNodes_Add(t *testing.T) {
	vn := NewVisitedNodes()
	nodeID := "node1"
	vn.Add(nodeID)
	assert.True(t, vn.Contains(nodeID))
}

func TestVisitedNodes_Contains(t *testing.T) {
	vn := NewVisitedNodes()
	nodeID := "node1"

	// Node should not be present initially
	assert.False(t, vn.Contains(nodeID))

	// Add node and verify it's present
	vn.Add(nodeID)
	assert.True(t, vn.Contains(nodeID))

	// Check that other nodes are not present
	assert.False(t, vn.Contains("node2"))
}

func TestVisitedNodes_MultipleNodes(t *testing.T) {
	vn := NewVisitedNodes()

	// Add multiple nodes
	vn.Add("node1")
	vn.Add("node2")
	vn.Add("node3")

	// Verify all nodes are present
	assert.True(t, vn.Contains("node1"))
	assert.True(t, vn.Contains("node2"))
	assert.True(t, vn.Contains("node3"))

	// Verify non-existent node is not present
	assert.False(t, vn.Contains("node4"))
}

func TestControlFlow_VisitedNodes(t *testing.T) {
	cFlow := InitializeControlFlow()
	visitedNodes := cFlow.VisitedNodes()

	// Should return non-nil visited nodes
	assert.NotNil(t, visitedNodes)

	// Should be able to add and retrieve nodes
	nodeID := "testNode"
	visitedNodes.Add(nodeID)
	assert.True(t, visitedNodes.Contains(nodeID))
}

func TestInitializeControlFlow_VisitedNodes(t *testing.T) {
	cFlow := InitializeControlFlow().(*controlFlow)

	// Visited nodes should be initialized
	assert.NotNil(t, cFlow.visitedNodes)

	// Should start empty
	assert.False(t, cFlow.visitedNodes.Contains("anyNode"))

	// Should be able to add nodes through VisitedNodes() method
	vn := cFlow.VisitedNodes()
	vn.Add("node1")
	assert.True(t, cFlow.visitedNodes.Contains("node1"))
}

func TestControlFlow_NodeExecutionCount(t *testing.T) {
	cFlow := InitializeControlFlow().(*controlFlow)
	assert.Equal(t, uint32(0), cFlow.CurrentNodeExecutionCount())
	cFlow.IncrementNodeExecutionCount()
	assert.Equal(t, uint32(1), cFlow.CurrentNodeExecutionCount())
	cFlow.IncrementNodeExecutionCount()
	assert.Equal(t, uint32(2), cFlow.CurrentNodeExecutionCount())
}

func TestControlFlow_TaskExecutionCount(t *testing.T) {
	cFlow := InitializeControlFlow().(*controlFlow)
	assert.Equal(t, uint32(0), cFlow.CurrentTaskExecutionCount())
	cFlow.IncrementTaskExecutionCount()
	assert.Equal(t, uint32(1), cFlow.CurrentTaskExecutionCount())
	cFlow.IncrementTaskExecutionCount()
	assert.Equal(t, uint32(2), cFlow.CurrentTaskExecutionCount())
}
