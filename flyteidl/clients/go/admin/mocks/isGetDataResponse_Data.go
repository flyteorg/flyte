// Code generated by mockery v2.53.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// isGetDataResponse_Data is an autogenerated mock type for the isGetDataResponse_Data type
type isGetDataResponse_Data struct {
	mock.Mock
}

type isGetDataResponse_Data_Expecter struct {
	mock *mock.Mock
}

func (_m *isGetDataResponse_Data) EXPECT() *isGetDataResponse_Data_Expecter {
	return &isGetDataResponse_Data_Expecter{mock: &_m.Mock}
}

// isGetDataResponse_Data provides a mock function with no fields
func (_m *isGetDataResponse_Data) isGetDataResponse_Data() {
	_m.Called()
}

// isGetDataResponse_Data_isGetDataResponse_Data_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'isGetDataResponse_Data'
type isGetDataResponse_Data_isGetDataResponse_Data_Call struct {
	*mock.Call
}

// isGetDataResponse_Data is a helper method to define mock.On call
func (_e *isGetDataResponse_Data_Expecter) isGetDataResponse_Data() *isGetDataResponse_Data_isGetDataResponse_Data_Call {
	return &isGetDataResponse_Data_isGetDataResponse_Data_Call{Call: _e.mock.On("isGetDataResponse_Data")}
}

func (_c *isGetDataResponse_Data_isGetDataResponse_Data_Call) Run(run func()) *isGetDataResponse_Data_isGetDataResponse_Data_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *isGetDataResponse_Data_isGetDataResponse_Data_Call) Return() *isGetDataResponse_Data_isGetDataResponse_Data_Call {
	_c.Call.Return()
	return _c
}

func (_c *isGetDataResponse_Data_isGetDataResponse_Data_Call) RunAndReturn(run func()) *isGetDataResponse_Data_isGetDataResponse_Data_Call {
	_c.Run(run)
	return _c
}

// newIsGetDataResponse_Data creates a new instance of isGetDataResponse_Data. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newIsGetDataResponse_Data(t interface {
	mock.TestingT
	Cleanup(func())
}) *isGetDataResponse_Data {
	mock := &isGetDataResponse_Data{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
