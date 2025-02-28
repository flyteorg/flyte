// Code generated by mockery v2.52.1. DO NOT EDIT.

package mocks

import (
	executors "github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors"
	interfaces "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces"

	io "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io"

	ioutils "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/ioutils"

	mock "github.com/stretchr/testify/mock"

	storage "github.com/flyteorg/flyte/flytestdlib/storage"

	v1alpha1 "github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

// NodeExecutionContext is an autogenerated mock type for the NodeExecutionContext type
type NodeExecutionContext struct {
	mock.Mock
}

type NodeExecutionContext_Expecter struct {
	mock *mock.Mock
}

func (_m *NodeExecutionContext) EXPECT() *NodeExecutionContext_Expecter {
	return &NodeExecutionContext_Expecter{mock: &_m.Mock}
}

// ContextualNodeLookup provides a mock function with no fields
func (_m *NodeExecutionContext) ContextualNodeLookup() executors.NodeLookup {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for ContextualNodeLookup")
	}

	var r0 executors.NodeLookup
	if rf, ok := ret.Get(0).(func() executors.NodeLookup); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(executors.NodeLookup)
		}
	}

	return r0
}

// NodeExecutionContext_ContextualNodeLookup_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ContextualNodeLookup'
type NodeExecutionContext_ContextualNodeLookup_Call struct {
	*mock.Call
}

// ContextualNodeLookup is a helper method to define mock.On call
func (_e *NodeExecutionContext_Expecter) ContextualNodeLookup() *NodeExecutionContext_ContextualNodeLookup_Call {
	return &NodeExecutionContext_ContextualNodeLookup_Call{Call: _e.mock.On("ContextualNodeLookup")}
}

func (_c *NodeExecutionContext_ContextualNodeLookup_Call) Run(run func()) *NodeExecutionContext_ContextualNodeLookup_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *NodeExecutionContext_ContextualNodeLookup_Call) Return(_a0 executors.NodeLookup) *NodeExecutionContext_ContextualNodeLookup_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeExecutionContext_ContextualNodeLookup_Call) RunAndReturn(run func() executors.NodeLookup) *NodeExecutionContext_ContextualNodeLookup_Call {
	_c.Call.Return(run)
	return _c
}

// CurrentAttempt provides a mock function with no fields
func (_m *NodeExecutionContext) CurrentAttempt() uint32 {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for CurrentAttempt")
	}

	var r0 uint32
	if rf, ok := ret.Get(0).(func() uint32); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint32)
	}

	return r0
}

// NodeExecutionContext_CurrentAttempt_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CurrentAttempt'
type NodeExecutionContext_CurrentAttempt_Call struct {
	*mock.Call
}

// CurrentAttempt is a helper method to define mock.On call
func (_e *NodeExecutionContext_Expecter) CurrentAttempt() *NodeExecutionContext_CurrentAttempt_Call {
	return &NodeExecutionContext_CurrentAttempt_Call{Call: _e.mock.On("CurrentAttempt")}
}

func (_c *NodeExecutionContext_CurrentAttempt_Call) Run(run func()) *NodeExecutionContext_CurrentAttempt_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *NodeExecutionContext_CurrentAttempt_Call) Return(_a0 uint32) *NodeExecutionContext_CurrentAttempt_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeExecutionContext_CurrentAttempt_Call) RunAndReturn(run func() uint32) *NodeExecutionContext_CurrentAttempt_Call {
	_c.Call.Return(run)
	return _c
}

// DataStore provides a mock function with no fields
func (_m *NodeExecutionContext) DataStore() *storage.DataStore {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for DataStore")
	}

	var r0 *storage.DataStore
	if rf, ok := ret.Get(0).(func() *storage.DataStore); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*storage.DataStore)
		}
	}

	return r0
}

// NodeExecutionContext_DataStore_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DataStore'
type NodeExecutionContext_DataStore_Call struct {
	*mock.Call
}

// DataStore is a helper method to define mock.On call
func (_e *NodeExecutionContext_Expecter) DataStore() *NodeExecutionContext_DataStore_Call {
	return &NodeExecutionContext_DataStore_Call{Call: _e.mock.On("DataStore")}
}

func (_c *NodeExecutionContext_DataStore_Call) Run(run func()) *NodeExecutionContext_DataStore_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *NodeExecutionContext_DataStore_Call) Return(_a0 *storage.DataStore) *NodeExecutionContext_DataStore_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeExecutionContext_DataStore_Call) RunAndReturn(run func() *storage.DataStore) *NodeExecutionContext_DataStore_Call {
	_c.Call.Return(run)
	return _c
}

// EnqueueOwnerFunc provides a mock function with no fields
func (_m *NodeExecutionContext) EnqueueOwnerFunc() func() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for EnqueueOwnerFunc")
	}

	var r0 func() error
	if rf, ok := ret.Get(0).(func() func() error); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(func() error)
		}
	}

	return r0
}

// NodeExecutionContext_EnqueueOwnerFunc_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'EnqueueOwnerFunc'
type NodeExecutionContext_EnqueueOwnerFunc_Call struct {
	*mock.Call
}

// EnqueueOwnerFunc is a helper method to define mock.On call
func (_e *NodeExecutionContext_Expecter) EnqueueOwnerFunc() *NodeExecutionContext_EnqueueOwnerFunc_Call {
	return &NodeExecutionContext_EnqueueOwnerFunc_Call{Call: _e.mock.On("EnqueueOwnerFunc")}
}

func (_c *NodeExecutionContext_EnqueueOwnerFunc_Call) Run(run func()) *NodeExecutionContext_EnqueueOwnerFunc_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *NodeExecutionContext_EnqueueOwnerFunc_Call) Return(_a0 func() error) *NodeExecutionContext_EnqueueOwnerFunc_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeExecutionContext_EnqueueOwnerFunc_Call) RunAndReturn(run func() func() error) *NodeExecutionContext_EnqueueOwnerFunc_Call {
	_c.Call.Return(run)
	return _c
}

// EventsRecorder provides a mock function with no fields
func (_m *NodeExecutionContext) EventsRecorder() interfaces.EventRecorder {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for EventsRecorder")
	}

	var r0 interfaces.EventRecorder
	if rf, ok := ret.Get(0).(func() interfaces.EventRecorder); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interfaces.EventRecorder)
		}
	}

	return r0
}

// NodeExecutionContext_EventsRecorder_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'EventsRecorder'
type NodeExecutionContext_EventsRecorder_Call struct {
	*mock.Call
}

// EventsRecorder is a helper method to define mock.On call
func (_e *NodeExecutionContext_Expecter) EventsRecorder() *NodeExecutionContext_EventsRecorder_Call {
	return &NodeExecutionContext_EventsRecorder_Call{Call: _e.mock.On("EventsRecorder")}
}

func (_c *NodeExecutionContext_EventsRecorder_Call) Run(run func()) *NodeExecutionContext_EventsRecorder_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *NodeExecutionContext_EventsRecorder_Call) Return(_a0 interfaces.EventRecorder) *NodeExecutionContext_EventsRecorder_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeExecutionContext_EventsRecorder_Call) RunAndReturn(run func() interfaces.EventRecorder) *NodeExecutionContext_EventsRecorder_Call {
	_c.Call.Return(run)
	return _c
}

// ExecutionContext provides a mock function with no fields
func (_m *NodeExecutionContext) ExecutionContext() executors.ExecutionContext {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for ExecutionContext")
	}

	var r0 executors.ExecutionContext
	if rf, ok := ret.Get(0).(func() executors.ExecutionContext); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(executors.ExecutionContext)
		}
	}

	return r0
}

// NodeExecutionContext_ExecutionContext_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ExecutionContext'
type NodeExecutionContext_ExecutionContext_Call struct {
	*mock.Call
}

// ExecutionContext is a helper method to define mock.On call
func (_e *NodeExecutionContext_Expecter) ExecutionContext() *NodeExecutionContext_ExecutionContext_Call {
	return &NodeExecutionContext_ExecutionContext_Call{Call: _e.mock.On("ExecutionContext")}
}

func (_c *NodeExecutionContext_ExecutionContext_Call) Run(run func()) *NodeExecutionContext_ExecutionContext_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *NodeExecutionContext_ExecutionContext_Call) Return(_a0 executors.ExecutionContext) *NodeExecutionContext_ExecutionContext_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeExecutionContext_ExecutionContext_Call) RunAndReturn(run func() executors.ExecutionContext) *NodeExecutionContext_ExecutionContext_Call {
	_c.Call.Return(run)
	return _c
}

// InputReader provides a mock function with no fields
func (_m *NodeExecutionContext) InputReader() io.InputReader {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for InputReader")
	}

	var r0 io.InputReader
	if rf, ok := ret.Get(0).(func() io.InputReader); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.InputReader)
		}
	}

	return r0
}

// NodeExecutionContext_InputReader_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'InputReader'
type NodeExecutionContext_InputReader_Call struct {
	*mock.Call
}

// InputReader is a helper method to define mock.On call
func (_e *NodeExecutionContext_Expecter) InputReader() *NodeExecutionContext_InputReader_Call {
	return &NodeExecutionContext_InputReader_Call{Call: _e.mock.On("InputReader")}
}

func (_c *NodeExecutionContext_InputReader_Call) Run(run func()) *NodeExecutionContext_InputReader_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *NodeExecutionContext_InputReader_Call) Return(_a0 io.InputReader) *NodeExecutionContext_InputReader_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeExecutionContext_InputReader_Call) RunAndReturn(run func() io.InputReader) *NodeExecutionContext_InputReader_Call {
	_c.Call.Return(run)
	return _c
}

// Node provides a mock function with no fields
func (_m *NodeExecutionContext) Node() v1alpha1.ExecutableNode {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Node")
	}

	var r0 v1alpha1.ExecutableNode
	if rf, ok := ret.Get(0).(func() v1alpha1.ExecutableNode); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(v1alpha1.ExecutableNode)
		}
	}

	return r0
}

// NodeExecutionContext_Node_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Node'
type NodeExecutionContext_Node_Call struct {
	*mock.Call
}

// Node is a helper method to define mock.On call
func (_e *NodeExecutionContext_Expecter) Node() *NodeExecutionContext_Node_Call {
	return &NodeExecutionContext_Node_Call{Call: _e.mock.On("Node")}
}

func (_c *NodeExecutionContext_Node_Call) Run(run func()) *NodeExecutionContext_Node_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *NodeExecutionContext_Node_Call) Return(_a0 v1alpha1.ExecutableNode) *NodeExecutionContext_Node_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeExecutionContext_Node_Call) RunAndReturn(run func() v1alpha1.ExecutableNode) *NodeExecutionContext_Node_Call {
	_c.Call.Return(run)
	return _c
}

// NodeExecutionMetadata provides a mock function with no fields
func (_m *NodeExecutionContext) NodeExecutionMetadata() interfaces.NodeExecutionMetadata {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for NodeExecutionMetadata")
	}

	var r0 interfaces.NodeExecutionMetadata
	if rf, ok := ret.Get(0).(func() interfaces.NodeExecutionMetadata); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interfaces.NodeExecutionMetadata)
		}
	}

	return r0
}

// NodeExecutionContext_NodeExecutionMetadata_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'NodeExecutionMetadata'
type NodeExecutionContext_NodeExecutionMetadata_Call struct {
	*mock.Call
}

// NodeExecutionMetadata is a helper method to define mock.On call
func (_e *NodeExecutionContext_Expecter) NodeExecutionMetadata() *NodeExecutionContext_NodeExecutionMetadata_Call {
	return &NodeExecutionContext_NodeExecutionMetadata_Call{Call: _e.mock.On("NodeExecutionMetadata")}
}

func (_c *NodeExecutionContext_NodeExecutionMetadata_Call) Run(run func()) *NodeExecutionContext_NodeExecutionMetadata_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *NodeExecutionContext_NodeExecutionMetadata_Call) Return(_a0 interfaces.NodeExecutionMetadata) *NodeExecutionContext_NodeExecutionMetadata_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeExecutionContext_NodeExecutionMetadata_Call) RunAndReturn(run func() interfaces.NodeExecutionMetadata) *NodeExecutionContext_NodeExecutionMetadata_Call {
	_c.Call.Return(run)
	return _c
}

// NodeID provides a mock function with no fields
func (_m *NodeExecutionContext) NodeID() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for NodeID")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// NodeExecutionContext_NodeID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'NodeID'
type NodeExecutionContext_NodeID_Call struct {
	*mock.Call
}

// NodeID is a helper method to define mock.On call
func (_e *NodeExecutionContext_Expecter) NodeID() *NodeExecutionContext_NodeID_Call {
	return &NodeExecutionContext_NodeID_Call{Call: _e.mock.On("NodeID")}
}

func (_c *NodeExecutionContext_NodeID_Call) Run(run func()) *NodeExecutionContext_NodeID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *NodeExecutionContext_NodeID_Call) Return(_a0 string) *NodeExecutionContext_NodeID_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeExecutionContext_NodeID_Call) RunAndReturn(run func() string) *NodeExecutionContext_NodeID_Call {
	_c.Call.Return(run)
	return _c
}

// NodeStateReader provides a mock function with no fields
func (_m *NodeExecutionContext) NodeStateReader() interfaces.NodeStateReader {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for NodeStateReader")
	}

	var r0 interfaces.NodeStateReader
	if rf, ok := ret.Get(0).(func() interfaces.NodeStateReader); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interfaces.NodeStateReader)
		}
	}

	return r0
}

// NodeExecutionContext_NodeStateReader_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'NodeStateReader'
type NodeExecutionContext_NodeStateReader_Call struct {
	*mock.Call
}

// NodeStateReader is a helper method to define mock.On call
func (_e *NodeExecutionContext_Expecter) NodeStateReader() *NodeExecutionContext_NodeStateReader_Call {
	return &NodeExecutionContext_NodeStateReader_Call{Call: _e.mock.On("NodeStateReader")}
}

func (_c *NodeExecutionContext_NodeStateReader_Call) Run(run func()) *NodeExecutionContext_NodeStateReader_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *NodeExecutionContext_NodeStateReader_Call) Return(_a0 interfaces.NodeStateReader) *NodeExecutionContext_NodeStateReader_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeExecutionContext_NodeStateReader_Call) RunAndReturn(run func() interfaces.NodeStateReader) *NodeExecutionContext_NodeStateReader_Call {
	_c.Call.Return(run)
	return _c
}

// NodeStateWriter provides a mock function with no fields
func (_m *NodeExecutionContext) NodeStateWriter() interfaces.NodeStateWriter {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for NodeStateWriter")
	}

	var r0 interfaces.NodeStateWriter
	if rf, ok := ret.Get(0).(func() interfaces.NodeStateWriter); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interfaces.NodeStateWriter)
		}
	}

	return r0
}

// NodeExecutionContext_NodeStateWriter_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'NodeStateWriter'
type NodeExecutionContext_NodeStateWriter_Call struct {
	*mock.Call
}

// NodeStateWriter is a helper method to define mock.On call
func (_e *NodeExecutionContext_Expecter) NodeStateWriter() *NodeExecutionContext_NodeStateWriter_Call {
	return &NodeExecutionContext_NodeStateWriter_Call{Call: _e.mock.On("NodeStateWriter")}
}

func (_c *NodeExecutionContext_NodeStateWriter_Call) Run(run func()) *NodeExecutionContext_NodeStateWriter_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *NodeExecutionContext_NodeStateWriter_Call) Return(_a0 interfaces.NodeStateWriter) *NodeExecutionContext_NodeStateWriter_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeExecutionContext_NodeStateWriter_Call) RunAndReturn(run func() interfaces.NodeStateWriter) *NodeExecutionContext_NodeStateWriter_Call {
	_c.Call.Return(run)
	return _c
}

// NodeStatus provides a mock function with no fields
func (_m *NodeExecutionContext) NodeStatus() v1alpha1.ExecutableNodeStatus {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for NodeStatus")
	}

	var r0 v1alpha1.ExecutableNodeStatus
	if rf, ok := ret.Get(0).(func() v1alpha1.ExecutableNodeStatus); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(v1alpha1.ExecutableNodeStatus)
		}
	}

	return r0
}

// NodeExecutionContext_NodeStatus_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'NodeStatus'
type NodeExecutionContext_NodeStatus_Call struct {
	*mock.Call
}

// NodeStatus is a helper method to define mock.On call
func (_e *NodeExecutionContext_Expecter) NodeStatus() *NodeExecutionContext_NodeStatus_Call {
	return &NodeExecutionContext_NodeStatus_Call{Call: _e.mock.On("NodeStatus")}
}

func (_c *NodeExecutionContext_NodeStatus_Call) Run(run func()) *NodeExecutionContext_NodeStatus_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *NodeExecutionContext_NodeStatus_Call) Return(_a0 v1alpha1.ExecutableNodeStatus) *NodeExecutionContext_NodeStatus_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeExecutionContext_NodeStatus_Call) RunAndReturn(run func() v1alpha1.ExecutableNodeStatus) *NodeExecutionContext_NodeStatus_Call {
	_c.Call.Return(run)
	return _c
}

// OutputShardSelector provides a mock function with no fields
func (_m *NodeExecutionContext) OutputShardSelector() ioutils.ShardSelector {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for OutputShardSelector")
	}

	var r0 ioutils.ShardSelector
	if rf, ok := ret.Get(0).(func() ioutils.ShardSelector); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(ioutils.ShardSelector)
		}
	}

	return r0
}

// NodeExecutionContext_OutputShardSelector_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OutputShardSelector'
type NodeExecutionContext_OutputShardSelector_Call struct {
	*mock.Call
}

// OutputShardSelector is a helper method to define mock.On call
func (_e *NodeExecutionContext_Expecter) OutputShardSelector() *NodeExecutionContext_OutputShardSelector_Call {
	return &NodeExecutionContext_OutputShardSelector_Call{Call: _e.mock.On("OutputShardSelector")}
}

func (_c *NodeExecutionContext_OutputShardSelector_Call) Run(run func()) *NodeExecutionContext_OutputShardSelector_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *NodeExecutionContext_OutputShardSelector_Call) Return(_a0 ioutils.ShardSelector) *NodeExecutionContext_OutputShardSelector_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeExecutionContext_OutputShardSelector_Call) RunAndReturn(run func() ioutils.ShardSelector) *NodeExecutionContext_OutputShardSelector_Call {
	_c.Call.Return(run)
	return _c
}

// RawOutputPrefix provides a mock function with no fields
func (_m *NodeExecutionContext) RawOutputPrefix() storage.DataReference {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for RawOutputPrefix")
	}

	var r0 storage.DataReference
	if rf, ok := ret.Get(0).(func() storage.DataReference); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(storage.DataReference)
	}

	return r0
}

// NodeExecutionContext_RawOutputPrefix_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RawOutputPrefix'
type NodeExecutionContext_RawOutputPrefix_Call struct {
	*mock.Call
}

// RawOutputPrefix is a helper method to define mock.On call
func (_e *NodeExecutionContext_Expecter) RawOutputPrefix() *NodeExecutionContext_RawOutputPrefix_Call {
	return &NodeExecutionContext_RawOutputPrefix_Call{Call: _e.mock.On("RawOutputPrefix")}
}

func (_c *NodeExecutionContext_RawOutputPrefix_Call) Run(run func()) *NodeExecutionContext_RawOutputPrefix_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *NodeExecutionContext_RawOutputPrefix_Call) Return(_a0 storage.DataReference) *NodeExecutionContext_RawOutputPrefix_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeExecutionContext_RawOutputPrefix_Call) RunAndReturn(run func() storage.DataReference) *NodeExecutionContext_RawOutputPrefix_Call {
	_c.Call.Return(run)
	return _c
}

// TaskReader provides a mock function with no fields
func (_m *NodeExecutionContext) TaskReader() interfaces.TaskReader {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for TaskReader")
	}

	var r0 interfaces.TaskReader
	if rf, ok := ret.Get(0).(func() interfaces.TaskReader); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interfaces.TaskReader)
		}
	}

	return r0
}

// NodeExecutionContext_TaskReader_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'TaskReader'
type NodeExecutionContext_TaskReader_Call struct {
	*mock.Call
}

// TaskReader is a helper method to define mock.On call
func (_e *NodeExecutionContext_Expecter) TaskReader() *NodeExecutionContext_TaskReader_Call {
	return &NodeExecutionContext_TaskReader_Call{Call: _e.mock.On("TaskReader")}
}

func (_c *NodeExecutionContext_TaskReader_Call) Run(run func()) *NodeExecutionContext_TaskReader_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *NodeExecutionContext_TaskReader_Call) Return(_a0 interfaces.TaskReader) *NodeExecutionContext_TaskReader_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *NodeExecutionContext_TaskReader_Call) RunAndReturn(run func() interfaces.TaskReader) *NodeExecutionContext_TaskReader_Call {
	_c.Call.Return(run)
	return _c
}

// NewNodeExecutionContext creates a new instance of NodeExecutionContext. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewNodeExecutionContext(t interface {
	mock.TestingT
	Cleanup(func())
}) *NodeExecutionContext {
	mock := &NodeExecutionContext{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
