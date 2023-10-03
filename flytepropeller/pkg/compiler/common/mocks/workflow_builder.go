// Code generated by mockery v1.0.1. DO NOT EDIT.

package mocks

import (
	common "github.com/flyteorg/flyte/flytepropeller/pkg/compiler/common"
	core "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"

	errors "github.com/flyteorg/flyte/flytepropeller/pkg/compiler/errors"

	mock "github.com/stretchr/testify/mock"
)

// WorkflowBuilder is an autogenerated mock type for the WorkflowBuilder type
type WorkflowBuilder struct {
	mock.Mock
}

// AddDownstreamEdge provides a mock function with given fields: nodeProvider, nodeDependent
func (_m *WorkflowBuilder) AddDownstreamEdge(nodeProvider string, nodeDependent string) {
	_m.Called(nodeProvider, nodeDependent)
}

type WorkflowBuilder_AddEdges struct {
	*mock.Call
}

func (_m WorkflowBuilder_AddEdges) Return(ok bool) *WorkflowBuilder_AddEdges {
	return &WorkflowBuilder_AddEdges{Call: _m.Call.Return(ok)}
}

func (_m *WorkflowBuilder) OnAddEdges(n common.NodeBuilder, edgeDirection common.EdgeDirection, errs errors.CompileErrors) *WorkflowBuilder_AddEdges {
	c_call := _m.On("AddEdges", n, edgeDirection, errs)
	return &WorkflowBuilder_AddEdges{Call: c_call}
}

func (_m *WorkflowBuilder) OnAddEdgesMatch(matchers ...interface{}) *WorkflowBuilder_AddEdges {
	c_call := _m.On("AddEdges", matchers...)
	return &WorkflowBuilder_AddEdges{Call: c_call}
}

// AddEdges provides a mock function with given fields: n, edgeDirection, errs
func (_m *WorkflowBuilder) AddEdges(n common.NodeBuilder, edgeDirection common.EdgeDirection, errs errors.CompileErrors) bool {
	ret := _m.Called(n, edgeDirection, errs)

	var r0 bool
	if rf, ok := ret.Get(0).(func(common.NodeBuilder, common.EdgeDirection, errors.CompileErrors) bool); ok {
		r0 = rf(n, edgeDirection, errs)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// AddExecutionEdge provides a mock function with given fields: nodeFrom, nodeTo
func (_m *WorkflowBuilder) AddExecutionEdge(nodeFrom string, nodeTo string) {
	_m.Called(nodeFrom, nodeTo)
}

type WorkflowBuilder_AddNode struct {
	*mock.Call
}

func (_m WorkflowBuilder_AddNode) Return(node common.NodeBuilder, ok bool) *WorkflowBuilder_AddNode {
	return &WorkflowBuilder_AddNode{Call: _m.Call.Return(node, ok)}
}

func (_m *WorkflowBuilder) OnAddNode(n common.NodeBuilder, errs errors.CompileErrors) *WorkflowBuilder_AddNode {
	c_call := _m.On("AddNode", n, errs)
	return &WorkflowBuilder_AddNode{Call: c_call}
}

func (_m *WorkflowBuilder) OnAddNodeMatch(matchers ...interface{}) *WorkflowBuilder_AddNode {
	c_call := _m.On("AddNode", matchers...)
	return &WorkflowBuilder_AddNode{Call: c_call}
}

// AddNode provides a mock function with given fields: n, errs
func (_m *WorkflowBuilder) AddNode(n common.NodeBuilder, errs errors.CompileErrors) (common.NodeBuilder, bool) {
	ret := _m.Called(n, errs)

	var r0 common.NodeBuilder
	if rf, ok := ret.Get(0).(func(common.NodeBuilder, errors.CompileErrors) common.NodeBuilder); ok {
		r0 = rf(n, errs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(common.NodeBuilder)
		}
	}

	var r1 bool
	if rf, ok := ret.Get(1).(func(common.NodeBuilder, errors.CompileErrors) bool); ok {
		r1 = rf(n, errs)
	} else {
		r1 = ret.Get(1).(bool)
	}

	return r0, r1
}

// AddUpstreamEdge provides a mock function with given fields: nodeProvider, nodeDependent
func (_m *WorkflowBuilder) AddUpstreamEdge(nodeProvider string, nodeDependent string) {
	_m.Called(nodeProvider, nodeDependent)
}

type WorkflowBuilder_GetCompiledSubWorkflow struct {
	*mock.Call
}

func (_m WorkflowBuilder_GetCompiledSubWorkflow) Return(wf *core.CompiledWorkflow, found bool) *WorkflowBuilder_GetCompiledSubWorkflow {
	return &WorkflowBuilder_GetCompiledSubWorkflow{Call: _m.Call.Return(wf, found)}
}

func (_m *WorkflowBuilder) OnGetCompiledSubWorkflow(id core.Identifier) *WorkflowBuilder_GetCompiledSubWorkflow {
	c_call := _m.On("GetCompiledSubWorkflow", id)
	return &WorkflowBuilder_GetCompiledSubWorkflow{Call: c_call}
}

func (_m *WorkflowBuilder) OnGetCompiledSubWorkflowMatch(matchers ...interface{}) *WorkflowBuilder_GetCompiledSubWorkflow {
	c_call := _m.On("GetCompiledSubWorkflow", matchers...)
	return &WorkflowBuilder_GetCompiledSubWorkflow{Call: c_call}
}

// GetCompiledSubWorkflow provides a mock function with given fields: id
func (_m *WorkflowBuilder) GetCompiledSubWorkflow(id core.Identifier) (*core.CompiledWorkflow, bool) {
	ret := _m.Called(id)

	var r0 *core.CompiledWorkflow
	if rf, ok := ret.Get(0).(func(core.Identifier) *core.CompiledWorkflow); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*core.CompiledWorkflow)
		}
	}

	var r1 bool
	if rf, ok := ret.Get(1).(func(core.Identifier) bool); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Get(1).(bool)
	}

	return r0, r1
}

type WorkflowBuilder_GetCoreWorkflow struct {
	*mock.Call
}

func (_m WorkflowBuilder_GetCoreWorkflow) Return(_a0 *core.CompiledWorkflow) *WorkflowBuilder_GetCoreWorkflow {
	return &WorkflowBuilder_GetCoreWorkflow{Call: _m.Call.Return(_a0)}
}

func (_m *WorkflowBuilder) OnGetCoreWorkflow() *WorkflowBuilder_GetCoreWorkflow {
	c_call := _m.On("GetCoreWorkflow")
	return &WorkflowBuilder_GetCoreWorkflow{Call: c_call}
}

func (_m *WorkflowBuilder) OnGetCoreWorkflowMatch(matchers ...interface{}) *WorkflowBuilder_GetCoreWorkflow {
	c_call := _m.On("GetCoreWorkflow", matchers...)
	return &WorkflowBuilder_GetCoreWorkflow{Call: c_call}
}

// GetCoreWorkflow provides a mock function with given fields:
func (_m *WorkflowBuilder) GetCoreWorkflow() *core.CompiledWorkflow {
	ret := _m.Called()

	var r0 *core.CompiledWorkflow
	if rf, ok := ret.Get(0).(func() *core.CompiledWorkflow); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*core.CompiledWorkflow)
		}
	}

	return r0
}

type WorkflowBuilder_GetDownstreamNodes struct {
	*mock.Call
}

func (_m WorkflowBuilder_GetDownstreamNodes) Return(_a0 common.StringAdjacencyList) *WorkflowBuilder_GetDownstreamNodes {
	return &WorkflowBuilder_GetDownstreamNodes{Call: _m.Call.Return(_a0)}
}

func (_m *WorkflowBuilder) OnGetDownstreamNodes() *WorkflowBuilder_GetDownstreamNodes {
	c_call := _m.On("GetDownstreamNodes")
	return &WorkflowBuilder_GetDownstreamNodes{Call: c_call}
}

func (_m *WorkflowBuilder) OnGetDownstreamNodesMatch(matchers ...interface{}) *WorkflowBuilder_GetDownstreamNodes {
	c_call := _m.On("GetDownstreamNodes", matchers...)
	return &WorkflowBuilder_GetDownstreamNodes{Call: c_call}
}

// GetDownstreamNodes provides a mock function with given fields:
func (_m *WorkflowBuilder) GetDownstreamNodes() common.StringAdjacencyList {
	ret := _m.Called()

	var r0 common.StringAdjacencyList
	if rf, ok := ret.Get(0).(func() common.StringAdjacencyList); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(common.StringAdjacencyList)
		}
	}

	return r0
}

type WorkflowBuilder_GetFailureNode struct {
	*mock.Call
}

func (_m WorkflowBuilder_GetFailureNode) Return(_a0 common.Node) *WorkflowBuilder_GetFailureNode {
	return &WorkflowBuilder_GetFailureNode{Call: _m.Call.Return(_a0)}
}

func (_m *WorkflowBuilder) OnGetFailureNode() *WorkflowBuilder_GetFailureNode {
	c_call := _m.On("GetFailureNode")
	return &WorkflowBuilder_GetFailureNode{Call: c_call}
}

func (_m *WorkflowBuilder) OnGetFailureNodeMatch(matchers ...interface{}) *WorkflowBuilder_GetFailureNode {
	c_call := _m.On("GetFailureNode", matchers...)
	return &WorkflowBuilder_GetFailureNode{Call: c_call}
}

// GetFailureNode provides a mock function with given fields:
func (_m *WorkflowBuilder) GetFailureNode() common.Node {
	ret := _m.Called()

	var r0 common.Node
	if rf, ok := ret.Get(0).(func() common.Node); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(common.Node)
		}
	}

	return r0
}

type WorkflowBuilder_GetLaunchPlan struct {
	*mock.Call
}

func (_m WorkflowBuilder_GetLaunchPlan) Return(wf common.InterfaceProvider, found bool) *WorkflowBuilder_GetLaunchPlan {
	return &WorkflowBuilder_GetLaunchPlan{Call: _m.Call.Return(wf, found)}
}

func (_m *WorkflowBuilder) OnGetLaunchPlan(id core.Identifier) *WorkflowBuilder_GetLaunchPlan {
	c_call := _m.On("GetLaunchPlan", id)
	return &WorkflowBuilder_GetLaunchPlan{Call: c_call}
}

func (_m *WorkflowBuilder) OnGetLaunchPlanMatch(matchers ...interface{}) *WorkflowBuilder_GetLaunchPlan {
	c_call := _m.On("GetLaunchPlan", matchers...)
	return &WorkflowBuilder_GetLaunchPlan{Call: c_call}
}

// GetLaunchPlan provides a mock function with given fields: id
func (_m *WorkflowBuilder) GetLaunchPlan(id core.Identifier) (common.InterfaceProvider, bool) {
	ret := _m.Called(id)

	var r0 common.InterfaceProvider
	if rf, ok := ret.Get(0).(func(core.Identifier) common.InterfaceProvider); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(common.InterfaceProvider)
		}
	}

	var r1 bool
	if rf, ok := ret.Get(1).(func(core.Identifier) bool); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Get(1).(bool)
	}

	return r0, r1
}

type WorkflowBuilder_GetNode struct {
	*mock.Call
}

func (_m WorkflowBuilder_GetNode) Return(node common.NodeBuilder, found bool) *WorkflowBuilder_GetNode {
	return &WorkflowBuilder_GetNode{Call: _m.Call.Return(node, found)}
}

func (_m *WorkflowBuilder) OnGetNode(id string) *WorkflowBuilder_GetNode {
	c_call := _m.On("GetNode", id)
	return &WorkflowBuilder_GetNode{Call: c_call}
}

func (_m *WorkflowBuilder) OnGetNodeMatch(matchers ...interface{}) *WorkflowBuilder_GetNode {
	c_call := _m.On("GetNode", matchers...)
	return &WorkflowBuilder_GetNode{Call: c_call}
}

// GetNode provides a mock function with given fields: id
func (_m *WorkflowBuilder) GetNode(id string) (common.NodeBuilder, bool) {
	ret := _m.Called(id)

	var r0 common.NodeBuilder
	if rf, ok := ret.Get(0).(func(string) common.NodeBuilder); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(common.NodeBuilder)
		}
	}

	var r1 bool
	if rf, ok := ret.Get(1).(func(string) bool); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Get(1).(bool)
	}

	return r0, r1
}

type WorkflowBuilder_GetNodes struct {
	*mock.Call
}

func (_m WorkflowBuilder_GetNodes) Return(_a0 common.NodeIndex) *WorkflowBuilder_GetNodes {
	return &WorkflowBuilder_GetNodes{Call: _m.Call.Return(_a0)}
}

func (_m *WorkflowBuilder) OnGetNodes() *WorkflowBuilder_GetNodes {
	c_call := _m.On("GetNodes")
	return &WorkflowBuilder_GetNodes{Call: c_call}
}

func (_m *WorkflowBuilder) OnGetNodesMatch(matchers ...interface{}) *WorkflowBuilder_GetNodes {
	c_call := _m.On("GetNodes", matchers...)
	return &WorkflowBuilder_GetNodes{Call: c_call}
}

// GetNodes provides a mock function with given fields:
func (_m *WorkflowBuilder) GetNodes() common.NodeIndex {
	ret := _m.Called()

	var r0 common.NodeIndex
	if rf, ok := ret.Get(0).(func() common.NodeIndex); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(common.NodeIndex)
		}
	}

	return r0
}

type WorkflowBuilder_GetOrCreateNodeBuilder struct {
	*mock.Call
}

func (_m WorkflowBuilder_GetOrCreateNodeBuilder) Return(_a0 common.NodeBuilder) *WorkflowBuilder_GetOrCreateNodeBuilder {
	return &WorkflowBuilder_GetOrCreateNodeBuilder{Call: _m.Call.Return(_a0)}
}

func (_m *WorkflowBuilder) OnGetOrCreateNodeBuilder(n *core.Node) *WorkflowBuilder_GetOrCreateNodeBuilder {
	c_call := _m.On("GetOrCreateNodeBuilder", n)
	return &WorkflowBuilder_GetOrCreateNodeBuilder{Call: c_call}
}

func (_m *WorkflowBuilder) OnGetOrCreateNodeBuilderMatch(matchers ...interface{}) *WorkflowBuilder_GetOrCreateNodeBuilder {
	c_call := _m.On("GetOrCreateNodeBuilder", matchers...)
	return &WorkflowBuilder_GetOrCreateNodeBuilder{Call: c_call}
}

// GetOrCreateNodeBuilder provides a mock function with given fields: n
func (_m *WorkflowBuilder) GetOrCreateNodeBuilder(n *core.Node) common.NodeBuilder {
	ret := _m.Called(n)

	var r0 common.NodeBuilder
	if rf, ok := ret.Get(0).(func(*core.Node) common.NodeBuilder); ok {
		r0 = rf(n)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(common.NodeBuilder)
		}
	}

	return r0
}

type WorkflowBuilder_GetSubWorkflow struct {
	*mock.Call
}

func (_m WorkflowBuilder_GetSubWorkflow) Return(wf *core.CompiledWorkflow, found bool) *WorkflowBuilder_GetSubWorkflow {
	return &WorkflowBuilder_GetSubWorkflow{Call: _m.Call.Return(wf, found)}
}

func (_m *WorkflowBuilder) OnGetSubWorkflow(id core.Identifier) *WorkflowBuilder_GetSubWorkflow {
	c_call := _m.On("GetSubWorkflow", id)
	return &WorkflowBuilder_GetSubWorkflow{Call: c_call}
}

func (_m *WorkflowBuilder) OnGetSubWorkflowMatch(matchers ...interface{}) *WorkflowBuilder_GetSubWorkflow {
	c_call := _m.On("GetSubWorkflow", matchers...)
	return &WorkflowBuilder_GetSubWorkflow{Call: c_call}
}

// GetSubWorkflow provides a mock function with given fields: id
func (_m *WorkflowBuilder) GetSubWorkflow(id core.Identifier) (*core.CompiledWorkflow, bool) {
	ret := _m.Called(id)

	var r0 *core.CompiledWorkflow
	if rf, ok := ret.Get(0).(func(core.Identifier) *core.CompiledWorkflow); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*core.CompiledWorkflow)
		}
	}

	var r1 bool
	if rf, ok := ret.Get(1).(func(core.Identifier) bool); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Get(1).(bool)
	}

	return r0, r1
}

type WorkflowBuilder_GetTask struct {
	*mock.Call
}

func (_m WorkflowBuilder_GetTask) Return(task common.Task, found bool) *WorkflowBuilder_GetTask {
	return &WorkflowBuilder_GetTask{Call: _m.Call.Return(task, found)}
}

func (_m *WorkflowBuilder) OnGetTask(id core.Identifier) *WorkflowBuilder_GetTask {
	c_call := _m.On("GetTask", id)
	return &WorkflowBuilder_GetTask{Call: c_call}
}

func (_m *WorkflowBuilder) OnGetTaskMatch(matchers ...interface{}) *WorkflowBuilder_GetTask {
	c_call := _m.On("GetTask", matchers...)
	return &WorkflowBuilder_GetTask{Call: c_call}
}

// GetTask provides a mock function with given fields: id
func (_m *WorkflowBuilder) GetTask(id core.Identifier) (common.Task, bool) {
	ret := _m.Called(id)

	var r0 common.Task
	if rf, ok := ret.Get(0).(func(core.Identifier) common.Task); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(common.Task)
		}
	}

	var r1 bool
	if rf, ok := ret.Get(1).(func(core.Identifier) bool); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Get(1).(bool)
	}

	return r0, r1
}

type WorkflowBuilder_GetTasks struct {
	*mock.Call
}

func (_m WorkflowBuilder_GetTasks) Return(_a0 common.TaskIndex) *WorkflowBuilder_GetTasks {
	return &WorkflowBuilder_GetTasks{Call: _m.Call.Return(_a0)}
}

func (_m *WorkflowBuilder) OnGetTasks() *WorkflowBuilder_GetTasks {
	c_call := _m.On("GetTasks")
	return &WorkflowBuilder_GetTasks{Call: c_call}
}

func (_m *WorkflowBuilder) OnGetTasksMatch(matchers ...interface{}) *WorkflowBuilder_GetTasks {
	c_call := _m.On("GetTasks", matchers...)
	return &WorkflowBuilder_GetTasks{Call: c_call}
}

// GetTasks provides a mock function with given fields:
func (_m *WorkflowBuilder) GetTasks() common.TaskIndex {
	ret := _m.Called()

	var r0 common.TaskIndex
	if rf, ok := ret.Get(0).(func() common.TaskIndex); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(common.TaskIndex)
		}
	}

	return r0
}

type WorkflowBuilder_GetUpstreamNodes struct {
	*mock.Call
}

func (_m WorkflowBuilder_GetUpstreamNodes) Return(_a0 common.StringAdjacencyList) *WorkflowBuilder_GetUpstreamNodes {
	return &WorkflowBuilder_GetUpstreamNodes{Call: _m.Call.Return(_a0)}
}

func (_m *WorkflowBuilder) OnGetUpstreamNodes() *WorkflowBuilder_GetUpstreamNodes {
	c_call := _m.On("GetUpstreamNodes")
	return &WorkflowBuilder_GetUpstreamNodes{Call: c_call}
}

func (_m *WorkflowBuilder) OnGetUpstreamNodesMatch(matchers ...interface{}) *WorkflowBuilder_GetUpstreamNodes {
	c_call := _m.On("GetUpstreamNodes", matchers...)
	return &WorkflowBuilder_GetUpstreamNodes{Call: c_call}
}

// GetUpstreamNodes provides a mock function with given fields:
func (_m *WorkflowBuilder) GetUpstreamNodes() common.StringAdjacencyList {
	ret := _m.Called()

	var r0 common.StringAdjacencyList
	if rf, ok := ret.Get(0).(func() common.StringAdjacencyList); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(common.StringAdjacencyList)
		}
	}

	return r0
}

// StoreCompiledSubWorkflow provides a mock function with given fields: id, compiledWorkflow
func (_m *WorkflowBuilder) StoreCompiledSubWorkflow(id core.Identifier, compiledWorkflow *core.CompiledWorkflow) {
	_m.Called(id, compiledWorkflow)
}

type WorkflowBuilder_ValidateWorkflow struct {
	*mock.Call
}

func (_m WorkflowBuilder_ValidateWorkflow) Return(_a0 common.Workflow, _a1 bool) *WorkflowBuilder_ValidateWorkflow {
	return &WorkflowBuilder_ValidateWorkflow{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *WorkflowBuilder) OnValidateWorkflow(fg *core.CompiledWorkflow, errs errors.CompileErrors) *WorkflowBuilder_ValidateWorkflow {
	c_call := _m.On("ValidateWorkflow", fg, errs)
	return &WorkflowBuilder_ValidateWorkflow{Call: c_call}
}

func (_m *WorkflowBuilder) OnValidateWorkflowMatch(matchers ...interface{}) *WorkflowBuilder_ValidateWorkflow {
	c_call := _m.On("ValidateWorkflow", matchers...)
	return &WorkflowBuilder_ValidateWorkflow{Call: c_call}
}

// ValidateWorkflow provides a mock function with given fields: fg, errs
func (_m *WorkflowBuilder) ValidateWorkflow(fg *core.CompiledWorkflow, errs errors.CompileErrors) (common.Workflow, bool) {
	ret := _m.Called(fg, errs)

	var r0 common.Workflow
	if rf, ok := ret.Get(0).(func(*core.CompiledWorkflow, errors.CompileErrors) common.Workflow); ok {
		r0 = rf(fg, errs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(common.Workflow)
		}
	}

	var r1 bool
	if rf, ok := ret.Get(1).(func(*core.CompiledWorkflow, errors.CompileErrors) bool); ok {
		r1 = rf(fg, errs)
	} else {
		r1 = ret.Get(1).(bool)
	}

	return r0, r1
}
