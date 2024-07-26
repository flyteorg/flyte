// Code generated by mockery v1.0.1. DO NOT EDIT.

package mocks

import (
	context "context"

	array "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/array"

	executors "github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors"

	handler "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/handler"

	interfaces "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces"

	mock "github.com/stretchr/testify/mock"

	v1alpha1 "github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

// arrayNodeStateStore is an autogenerated mock type for the arrayNodeStateStore type
type arrayNodeStateStore struct {
	mock.Mock
}

type arrayNodeStateStore_buildArrayNodeContext struct {
	*mock.Call
}

func (_m arrayNodeStateStore_buildArrayNodeContext) Return(_a0 interfaces.Node, _a1 executors.ExecutionContext, _a2 executors.DAGStructure, _a3 executors.NodeLookup, _a4 *v1alpha1.NodeSpec, _a5 *v1alpha1.NodeStatus, _a6 error) *arrayNodeStateStore_buildArrayNodeContext {
	return &arrayNodeStateStore_buildArrayNodeContext{Call: _m.Call.Return(_a0, _a1, _a2, _a3, _a4, _a5, _a6)}
}

func (_m *arrayNodeStateStore) OnbuildArrayNodeContext(ctx context.Context, nCtx interfaces.NodeExecutionContext, arrayNode v1alpha1.ExecutableArrayNode, subNodeIndex int, eventRecorder array.ArrayEventRecorder) *arrayNodeStateStore_buildArrayNodeContext {
	c_call := _m.On("buildArrayNodeContext", ctx, nCtx, arrayNode, subNodeIndex, eventRecorder)
	return &arrayNodeStateStore_buildArrayNodeContext{Call: c_call}
}

func (_m *arrayNodeStateStore) OnbuildArrayNodeContextMatch(matchers ...interface{}) *arrayNodeStateStore_buildArrayNodeContext {
	c_call := _m.On("buildArrayNodeContext", matchers...)
	return &arrayNodeStateStore_buildArrayNodeContext{Call: c_call}
}

// buildArrayNodeContext provides a mock function with given fields: ctx, nCtx, arrayNode, subNodeIndex, eventRecorder
func (_m *arrayNodeStateStore) buildArrayNodeContext(ctx context.Context, nCtx interfaces.NodeExecutionContext, arrayNode v1alpha1.ExecutableArrayNode, subNodeIndex int, eventRecorder array.ArrayEventRecorder) (interfaces.Node, executors.ExecutionContext, executors.DAGStructure, executors.NodeLookup, *v1alpha1.NodeSpec, *v1alpha1.NodeStatus, error) {
	ret := _m.Called(ctx, nCtx, arrayNode, subNodeIndex, eventRecorder)

	var r0 interfaces.Node
	if rf, ok := ret.Get(0).(func(context.Context, interfaces.NodeExecutionContext, v1alpha1.ExecutableArrayNode, int, array.ArrayEventRecorder) interfaces.Node); ok {
		r0 = rf(ctx, nCtx, arrayNode, subNodeIndex, eventRecorder)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interfaces.Node)
		}
	}

	var r1 executors.ExecutionContext
	if rf, ok := ret.Get(1).(func(context.Context, interfaces.NodeExecutionContext, v1alpha1.ExecutableArrayNode, int, array.ArrayEventRecorder) executors.ExecutionContext); ok {
		r1 = rf(ctx, nCtx, arrayNode, subNodeIndex, eventRecorder)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(executors.ExecutionContext)
		}
	}

	var r2 executors.DAGStructure
	if rf, ok := ret.Get(2).(func(context.Context, interfaces.NodeExecutionContext, v1alpha1.ExecutableArrayNode, int, array.ArrayEventRecorder) executors.DAGStructure); ok {
		r2 = rf(ctx, nCtx, arrayNode, subNodeIndex, eventRecorder)
	} else {
		if ret.Get(2) != nil {
			r2 = ret.Get(2).(executors.DAGStructure)
		}
	}

	var r3 executors.NodeLookup
	if rf, ok := ret.Get(3).(func(context.Context, interfaces.NodeExecutionContext, v1alpha1.ExecutableArrayNode, int, array.ArrayEventRecorder) executors.NodeLookup); ok {
		r3 = rf(ctx, nCtx, arrayNode, subNodeIndex, eventRecorder)
	} else {
		if ret.Get(3) != nil {
			r3 = ret.Get(3).(executors.NodeLookup)
		}
	}

	var r4 *v1alpha1.NodeSpec
	if rf, ok := ret.Get(4).(func(context.Context, interfaces.NodeExecutionContext, v1alpha1.ExecutableArrayNode, int, array.ArrayEventRecorder) *v1alpha1.NodeSpec); ok {
		r4 = rf(ctx, nCtx, arrayNode, subNodeIndex, eventRecorder)
	} else {
		if ret.Get(4) != nil {
			r4 = ret.Get(4).(*v1alpha1.NodeSpec)
		}
	}

	var r5 *v1alpha1.NodeStatus
	if rf, ok := ret.Get(5).(func(context.Context, interfaces.NodeExecutionContext, v1alpha1.ExecutableArrayNode, int, array.ArrayEventRecorder) *v1alpha1.NodeStatus); ok {
		r5 = rf(ctx, nCtx, arrayNode, subNodeIndex, eventRecorder)
	} else {
		if ret.Get(5) != nil {
			r5 = ret.Get(5).(*v1alpha1.NodeStatus)
		}
	}

	var r6 error
	if rf, ok := ret.Get(6).(func(context.Context, interfaces.NodeExecutionContext, v1alpha1.ExecutableArrayNode, int, array.ArrayEventRecorder) error); ok {
		r6 = rf(ctx, nCtx, arrayNode, subNodeIndex, eventRecorder)
	} else {
		r6 = ret.Error(6)
	}

	return r0, r1, r2, r3, r4, r5, r6
}

type arrayNodeStateStore_getAttempts struct {
	*mock.Call
}

func (_m arrayNodeStateStore_getAttempts) Return(_a0 uint32) *arrayNodeStateStore_getAttempts {
	return &arrayNodeStateStore_getAttempts{Call: _m.Call.Return(_a0)}
}

func (_m *arrayNodeStateStore) OngetAttempts(ctx context.Context, nCtx interfaces.NodeExecutionContext, index int) *arrayNodeStateStore_getAttempts {
	c_call := _m.On("getAttempts", ctx, nCtx, index)
	return &arrayNodeStateStore_getAttempts{Call: c_call}
}

func (_m *arrayNodeStateStore) OngetAttemptsMatch(matchers ...interface{}) *arrayNodeStateStore_getAttempts {
	c_call := _m.On("getAttempts", matchers...)
	return &arrayNodeStateStore_getAttempts{Call: c_call}
}

// getAttempts provides a mock function with given fields: ctx, nCtx, index
func (_m *arrayNodeStateStore) getAttempts(ctx context.Context, nCtx interfaces.NodeExecutionContext, index int) uint32 {
	ret := _m.Called(ctx, nCtx, index)

	var r0 uint32
	if rf, ok := ret.Get(0).(func(context.Context, interfaces.NodeExecutionContext, int) uint32); ok {
		r0 = rf(ctx, nCtx, index)
	} else {
		r0 = ret.Get(0).(uint32)
	}

	return r0
}

type arrayNodeStateStore_getTaskPhase struct {
	*mock.Call
}

func (_m arrayNodeStateStore_getTaskPhase) Return(_a0 int) *arrayNodeStateStore_getTaskPhase {
	return &arrayNodeStateStore_getTaskPhase{Call: _m.Call.Return(_a0)}
}

func (_m *arrayNodeStateStore) OngetTaskPhase(ctx context.Context, nCtx interfaces.NodeExecutionContext, index int) *arrayNodeStateStore_getTaskPhase {
	c_call := _m.On("getTaskPhase", ctx, nCtx, index)
	return &arrayNodeStateStore_getTaskPhase{Call: c_call}
}

func (_m *arrayNodeStateStore) OngetTaskPhaseMatch(matchers ...interface{}) *arrayNodeStateStore_getTaskPhase {
	c_call := _m.On("getTaskPhase", matchers...)
	return &arrayNodeStateStore_getTaskPhase{Call: c_call}
}

// getTaskPhase provides a mock function with given fields: ctx, nCtx, index
func (_m *arrayNodeStateStore) getTaskPhase(ctx context.Context, nCtx interfaces.NodeExecutionContext, index int) int {
	ret := _m.Called(ctx, nCtx, index)

	var r0 int
	if rf, ok := ret.Get(0).(func(context.Context, interfaces.NodeExecutionContext, int) int); ok {
		r0 = rf(ctx, nCtx, index)
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

type arrayNodeStateStore_initArrayNodeState struct {
	*mock.Call
}

func (_m arrayNodeStateStore_initArrayNodeState) Return(_a0 error) *arrayNodeStateStore_initArrayNodeState {
	return &arrayNodeStateStore_initArrayNodeState{Call: _m.Call.Return(_a0)}
}

func (_m *arrayNodeStateStore) OninitArrayNodeState(arrayNodeState *handler.ArrayNodeState, maxAttemptsValue int, maxSystemFailuresValue int, size int) *arrayNodeStateStore_initArrayNodeState {
	c_call := _m.On("initArrayNodeState", arrayNodeState, maxAttemptsValue, maxSystemFailuresValue, size)
	return &arrayNodeStateStore_initArrayNodeState{Call: c_call}
}

func (_m *arrayNodeStateStore) OninitArrayNodeStateMatch(matchers ...interface{}) *arrayNodeStateStore_initArrayNodeState {
	c_call := _m.On("initArrayNodeState", matchers...)
	return &arrayNodeStateStore_initArrayNodeState{Call: c_call}
}

// initArrayNodeState provides a mock function with given fields: arrayNodeState, maxAttemptsValue, maxSystemFailuresValue, size
func (_m *arrayNodeStateStore) initArrayNodeState(arrayNodeState *handler.ArrayNodeState, maxAttemptsValue int, maxSystemFailuresValue int, size int) error {
	ret := _m.Called(arrayNodeState, maxAttemptsValue, maxSystemFailuresValue, size)

	var r0 error
	if rf, ok := ret.Get(0).(func(*handler.ArrayNodeState, int, int, int) error); ok {
		r0 = rf(arrayNodeState, maxAttemptsValue, maxSystemFailuresValue, size)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// persistArraySubNodeState provides a mock function with given fields: ctx, nCtx, subNodeStatus, index
func (_m *arrayNodeStateStore) persistArraySubNodeState(ctx context.Context, nCtx interfaces.NodeExecutionContext, subNodeStatus *v1alpha1.NodeStatus, index int) {
	_m.Called(ctx, nCtx, subNodeStatus, index)
}
