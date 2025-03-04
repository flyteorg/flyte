// Code generated by mockery v2.53.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// WorkItem is an autogenerated mock type for the WorkItem type
type WorkItem struct {
	mock.Mock
}

type WorkItem_Expecter struct {
	mock *mock.Mock
}

func (_m *WorkItem) EXPECT() *WorkItem_Expecter {
	return &WorkItem_Expecter{mock: &_m.Mock}
}

// NewWorkItem creates a new instance of WorkItem. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewWorkItem(t interface {
	mock.TestingT
	Cleanup(func())
}) *WorkItem {
	mock := &WorkItem{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
