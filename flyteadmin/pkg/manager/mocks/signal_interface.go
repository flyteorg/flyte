// Code generated by mockery v1.0.1. DO NOT EDIT.

package mocks

import (
	context "context"

	admin "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"

	mock "github.com/stretchr/testify/mock"
)

// SignalInterface is an autogenerated mock type for the SignalInterface type
type SignalInterface struct {
	mock.Mock
}

type SignalInterface_GetOrCreateSignal struct {
	*mock.Call
}

func (_m SignalInterface_GetOrCreateSignal) Return(_a0 *admin.Signal, _a1 error) *SignalInterface_GetOrCreateSignal {
	return &SignalInterface_GetOrCreateSignal{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *SignalInterface) OnGetOrCreateSignal(ctx context.Context, request admin.SignalGetOrCreateRequest) *SignalInterface_GetOrCreateSignal {
	c_call := _m.On("GetOrCreateSignal", ctx, request)
	return &SignalInterface_GetOrCreateSignal{Call: c_call}
}

func (_m *SignalInterface) OnGetOrCreateSignalMatch(matchers ...interface{}) *SignalInterface_GetOrCreateSignal {
	c_call := _m.On("GetOrCreateSignal", matchers...)
	return &SignalInterface_GetOrCreateSignal{Call: c_call}
}

// GetOrCreateSignal provides a mock function with given fields: ctx, request
func (_m *SignalInterface) GetOrCreateSignal(ctx context.Context, request admin.SignalGetOrCreateRequest) (*admin.Signal, error) {
	ret := _m.Called(ctx, request)

	var r0 *admin.Signal
	if rf, ok := ret.Get(0).(func(context.Context, admin.SignalGetOrCreateRequest) *admin.Signal); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*admin.Signal)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, admin.SignalGetOrCreateRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type SignalInterface_ListSignals struct {
	*mock.Call
}

func (_m SignalInterface_ListSignals) Return(_a0 *admin.SignalList, _a1 error) *SignalInterface_ListSignals {
	return &SignalInterface_ListSignals{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *SignalInterface) OnListSignals(ctx context.Context, request admin.SignalListRequest) *SignalInterface_ListSignals {
	c_call := _m.On("ListSignals", ctx, request)
	return &SignalInterface_ListSignals{Call: c_call}
}

func (_m *SignalInterface) OnListSignalsMatch(matchers ...interface{}) *SignalInterface_ListSignals {
	c_call := _m.On("ListSignals", matchers...)
	return &SignalInterface_ListSignals{Call: c_call}
}

// ListSignals provides a mock function with given fields: ctx, request
func (_m *SignalInterface) ListSignals(ctx context.Context, request admin.SignalListRequest) (*admin.SignalList, error) {
	ret := _m.Called(ctx, request)

	var r0 *admin.SignalList
	if rf, ok := ret.Get(0).(func(context.Context, admin.SignalListRequest) *admin.SignalList); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*admin.SignalList)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, admin.SignalListRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type SignalInterface_SetSignal struct {
	*mock.Call
}

func (_m SignalInterface_SetSignal) Return(_a0 *admin.SignalSetResponse, _a1 error) *SignalInterface_SetSignal {
	return &SignalInterface_SetSignal{Call: _m.Call.Return(_a0, _a1)}
}

func (_m *SignalInterface) OnSetSignal(ctx context.Context, request admin.SignalSetRequest) *SignalInterface_SetSignal {
	c_call := _m.On("SetSignal", ctx, request)
	return &SignalInterface_SetSignal{Call: c_call}
}

func (_m *SignalInterface) OnSetSignalMatch(matchers ...interface{}) *SignalInterface_SetSignal {
	c_call := _m.On("SetSignal", matchers...)
	return &SignalInterface_SetSignal{Call: c_call}
}

// SetSignal provides a mock function with given fields: ctx, request
func (_m *SignalInterface) SetSignal(ctx context.Context, request admin.SignalSetRequest) (*admin.SignalSetResponse, error) {
	ret := _m.Called(ctx, request)

	var r0 *admin.SignalSetResponse
	if rf, ok := ret.Get(0).(func(context.Context, admin.SignalSetRequest) *admin.SignalSetResponse); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*admin.SignalSetResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, admin.SignalSetRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
