// Code generated by mockery v1.0.1. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	pb "github.com/unionai/flyte/fasttask/plugin/pb"
)

// FastTaskServer is an autogenerated mock type for the FastTaskServer type
type FastTaskServer struct {
	mock.Mock
}

type FastTaskServer_Heartbeat struct {
	*mock.Call
}

func (_m FastTaskServer_Heartbeat) Return(_a0 error) *FastTaskServer_Heartbeat {
	return &FastTaskServer_Heartbeat{Call: _m.Call.Return(_a0)}
}

func (_m *FastTaskServer) OnHeartbeat(_a0 pb.FastTask_HeartbeatServer) *FastTaskServer_Heartbeat {
	c_call := _m.On("Heartbeat", _a0)
	return &FastTaskServer_Heartbeat{Call: c_call}
}

func (_m *FastTaskServer) OnHeartbeatMatch(matchers ...interface{}) *FastTaskServer_Heartbeat {
	c_call := _m.On("Heartbeat", matchers...)
	return &FastTaskServer_Heartbeat{Call: c_call}
}

// Heartbeat provides a mock function with given fields: _a0
func (_m *FastTaskServer) Heartbeat(_a0 pb.FastTask_HeartbeatServer) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(pb.FastTask_HeartbeatServer) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mustEmbedUnimplementedFastTaskServer provides a mock function with given fields:
func (_m *FastTaskServer) mustEmbedUnimplementedFastTaskServer() {
	_m.Called()
}
