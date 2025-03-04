// Code generated by mockery v2.53.0. DO NOT EDIT.

package mocks

import (
	context "context"

	event "github.com/cloudevents/sdk-go/v2/event"

	mock "github.com/stretchr/testify/mock"
)

// Sender is an autogenerated mock type for the Sender type
type Sender struct {
	mock.Mock
}

type Sender_Expecter struct {
	mock *mock.Mock
}

func (_m *Sender) EXPECT() *Sender_Expecter {
	return &Sender_Expecter{mock: &_m.Mock}
}

// Send provides a mock function with given fields: ctx, notificationType, _a2
func (_m *Sender) Send(ctx context.Context, notificationType string, _a2 event.Event) error {
	ret := _m.Called(ctx, notificationType, _a2)

	if len(ret) == 0 {
		panic("no return value specified for Send")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, event.Event) error); ok {
		r0 = rf(ctx, notificationType, _a2)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Sender_Send_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Send'
type Sender_Send_Call struct {
	*mock.Call
}

// Send is a helper method to define mock.On call
//   - ctx context.Context
//   - notificationType string
//   - _a2 event.Event
func (_e *Sender_Expecter) Send(ctx interface{}, notificationType interface{}, _a2 interface{}) *Sender_Send_Call {
	return &Sender_Send_Call{Call: _e.mock.On("Send", ctx, notificationType, _a2)}
}

func (_c *Sender_Send_Call) Run(run func(ctx context.Context, notificationType string, _a2 event.Event)) *Sender_Send_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(event.Event))
	})
	return _c
}

func (_c *Sender_Send_Call) Return(_a0 error) *Sender_Send_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Sender_Send_Call) RunAndReturn(run func(context.Context, string, event.Event) error) *Sender_Send_Call {
	_c.Call.Return(run)
	return _c
}

// NewSender creates a new instance of Sender. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSender(t interface {
	mock.TestingT
	Cleanup(func())
}) *Sender {
	mock := &Sender{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
