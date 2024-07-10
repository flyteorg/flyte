// Code generated by mockery v1.0.1. DO NOT EDIT.

package mocks

import (
	context "context"

	core "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	launchplan "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/subworkflow/launchplan"

	mock "github.com/stretchr/testify/mock"

	v1alpha1 "github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

// Executor is an autogenerated mock type for the Executor type
type Executor struct {
	mock.Mock
}

type Executor_GetStatus struct {
	*mock.Call
}

func (_m Executor_GetStatus) Return(_a0 launchplan.ExecutionStatus, _a1 error) *Executor_GetStatus {
	return &Executor_GetStatus{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *Executor) OnGetStatus(ctx context.Context, executionID *core.WorkflowExecutionIdentifier, launchPlan v1alpha1.ExecutableLaunchPlan, parentWorkflowID string) *Executor_GetStatus {
	c_call := _m.On("GetStatus", ctx, executionID, launchPlan, parentWorkflowID)
	return &Executor_GetStatus{Call: c_call}
}

func (_m *Executor) OnGetStatusMatch(matchers ...interface{}) *Executor_GetStatus {
	c_call := _m.On("GetStatus", matchers...)
	return &Executor_GetStatus{Call: c_call}
}

// GetStatus provides a mock function with given fields: ctx, executionID, launchPlan, parentWorkflowID
func (_m *Executor) GetStatus(ctx context.Context, executionID *core.WorkflowExecutionIdentifier, launchPlan v1alpha1.ExecutableLaunchPlan, parentWorkflowID string) (launchplan.ExecutionStatus, error) {
	ret := _m.Called(ctx, executionID, launchPlan, parentWorkflowID)

	var r0 launchplan.ExecutionStatus
	if rf, ok := ret.Get(0).(func(context.Context, *core.WorkflowExecutionIdentifier, v1alpha1.ExecutableLaunchPlan, string) launchplan.ExecutionStatus); ok {
		r0 = rf(ctx, executionID, launchPlan, parentWorkflowID)
	} else {
		r0 = ret.Get(0).(launchplan.ExecutionStatus)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *core.WorkflowExecutionIdentifier, v1alpha1.ExecutableLaunchPlan, string) error); ok {
		r1 = rf(ctx, executionID, launchPlan, parentWorkflowID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type Executor_Initialize struct {
	*mock.Call
}

func (_m Executor_Initialize) Return(_a0 error) *Executor_Initialize {
	return &Executor_Initialize{Call: _m.Call.Return(_a0)}
}

func (_m *Executor) OnInitialize(ctx context.Context) *Executor_Initialize {
	c_call := _m.On("Initialize", ctx)
	return &Executor_Initialize{Call: c_call}
}

func (_m *Executor) OnInitializeMatch(matchers ...interface{}) *Executor_Initialize {
	c_call := _m.On("Initialize", matchers...)
	return &Executor_Initialize{Call: c_call}
}

// Initialize provides a mock function with given fields: ctx
func (_m *Executor) Initialize(ctx context.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type Executor_Kill struct {
	*mock.Call
}

func (_m Executor_Kill) Return(_a0 error) *Executor_Kill {
	return &Executor_Kill{Call: _m.Call.Return(_a0)}
}

func (_m *Executor) OnKill(ctx context.Context, executionID *core.WorkflowExecutionIdentifier, reason string) *Executor_Kill {
	c_call := _m.On("Kill", ctx, executionID, reason)
	return &Executor_Kill{Call: c_call}
}

func (_m *Executor) OnKillMatch(matchers ...interface{}) *Executor_Kill {
	c_call := _m.On("Kill", matchers...)
	return &Executor_Kill{Call: c_call}
}

// Kill provides a mock function with given fields: ctx, executionID, reason
func (_m *Executor) Kill(ctx context.Context, executionID *core.WorkflowExecutionIdentifier, reason string) error {
	ret := _m.Called(ctx, executionID, reason)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *core.WorkflowExecutionIdentifier, string) error); ok {
		r0 = rf(ctx, executionID, reason)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type Executor_Launch struct {
	*mock.Call
}

func (_m Executor_Launch) Return(_a0 error) *Executor_Launch {
	return &Executor_Launch{Call: _m.Call.Return(_a0)}
}

func (_m *Executor) OnLaunch(ctx context.Context, launchCtx launchplan.LaunchContext, executionID *core.WorkflowExecutionIdentifier, launchPlan v1alpha1.ExecutableLaunchPlan, inputs *core.LiteralMap, parentWorkflowID string) *Executor_Launch {
	c_call := _m.On("Launch", ctx, launchCtx, executionID, launchPlan, inputs, parentWorkflowID)
	return &Executor_Launch{Call: c_call}
}

func (_m *Executor) OnLaunchMatch(matchers ...interface{}) *Executor_Launch {
	c_call := _m.On("Launch", matchers...)
	return &Executor_Launch{Call: c_call}
}

// Launch provides a mock function with given fields: ctx, launchCtx, executionID, launchPlan, inputs, parentWorkflowID
func (_m *Executor) Launch(ctx context.Context, launchCtx launchplan.LaunchContext, executionID *core.WorkflowExecutionIdentifier, launchPlan v1alpha1.ExecutableLaunchPlan, inputs *core.LiteralMap, parentWorkflowID string) error {
	ret := _m.Called(ctx, launchCtx, executionID, launchPlan, inputs, parentWorkflowID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, launchplan.LaunchContext, *core.WorkflowExecutionIdentifier, v1alpha1.ExecutableLaunchPlan, *core.LiteralMap, string) error); ok {
		r0 = rf(ctx, launchCtx, executionID, launchPlan, inputs, parentWorkflowID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
