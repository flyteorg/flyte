// Code generated by mockery v2.40.3. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// UnsafeAdminServiceServer is an autogenerated mock type for the UnsafeAdminServiceServer type
type UnsafeAdminServiceServer struct {
	mock.Mock
}

type UnsafeAdminServiceServer_Expecter struct {
	mock *mock.Mock
}

func (_m *UnsafeAdminServiceServer) EXPECT() *UnsafeAdminServiceServer_Expecter {
	return &UnsafeAdminServiceServer_Expecter{mock: &_m.Mock}
}

// mustEmbedUnimplementedAdminServiceServer provides a mock function with given fields:
func (_m *UnsafeAdminServiceServer) mustEmbedUnimplementedAdminServiceServer() {
	_m.Called()
}

// UnsafeAdminServiceServer_mustEmbedUnimplementedAdminServiceServer_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'mustEmbedUnimplementedAdminServiceServer'
type UnsafeAdminServiceServer_mustEmbedUnimplementedAdminServiceServer_Call struct {
	*mock.Call
}

// mustEmbedUnimplementedAdminServiceServer is a helper method to define mock.On call
func (_e *UnsafeAdminServiceServer_Expecter) mustEmbedUnimplementedAdminServiceServer() *UnsafeAdminServiceServer_mustEmbedUnimplementedAdminServiceServer_Call {
	return &UnsafeAdminServiceServer_mustEmbedUnimplementedAdminServiceServer_Call{Call: _e.mock.On("mustEmbedUnimplementedAdminServiceServer")}
}

func (_c *UnsafeAdminServiceServer_mustEmbedUnimplementedAdminServiceServer_Call) Run(run func()) *UnsafeAdminServiceServer_mustEmbedUnimplementedAdminServiceServer_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *UnsafeAdminServiceServer_mustEmbedUnimplementedAdminServiceServer_Call) Return() *UnsafeAdminServiceServer_mustEmbedUnimplementedAdminServiceServer_Call {
	_c.Call.Return()
	return _c
}

func (_c *UnsafeAdminServiceServer_mustEmbedUnimplementedAdminServiceServer_Call) RunAndReturn(run func()) *UnsafeAdminServiceServer_mustEmbedUnimplementedAdminServiceServer_Call {
	_c.Call.Return(run)
	return _c
}

// NewUnsafeAdminServiceServer creates a new instance of UnsafeAdminServiceServer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewUnsafeAdminServiceServer(t interface {
	mock.TestingT
	Cleanup(func())
}) *UnsafeAdminServiceServer {
	mock := &UnsafeAdminServiceServer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
