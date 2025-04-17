// Code generated by mockery v2.52.1. DO NOT EDIT.

package mocks

import (
	gographviz "github.com/awalterschulze/gographviz"
	mock "github.com/stretchr/testify/mock"
)

// Graphvizer is an autogenerated mock type for the Graphvizer type
type Graphvizer struct {
	mock.Mock
}

type Graphvizer_Expecter struct {
	mock *mock.Mock
}

func (_m *Graphvizer) EXPECT() *Graphvizer_Expecter {
	return &Graphvizer_Expecter{mock: &_m.Mock}
}

// AddAttr provides a mock function with given fields: parentGraph, field, value
func (_m *Graphvizer) AddAttr(parentGraph string, field string, value string) error {
	ret := _m.Called(parentGraph, field, value)

	if len(ret) == 0 {
		panic("no return value specified for AddAttr")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string) error); ok {
		r0 = rf(parentGraph, field, value)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Graphvizer_AddAttr_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddAttr'
type Graphvizer_AddAttr_Call struct {
	*mock.Call
}

// AddAttr is a helper method to define mock.On call
//   - parentGraph string
//   - field string
//   - value string
func (_e *Graphvizer_Expecter) AddAttr(parentGraph interface{}, field interface{}, value interface{}) *Graphvizer_AddAttr_Call {
	return &Graphvizer_AddAttr_Call{Call: _e.mock.On("AddAttr", parentGraph, field, value)}
}

func (_c *Graphvizer_AddAttr_Call) Run(run func(parentGraph string, field string, value string)) *Graphvizer_AddAttr_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *Graphvizer_AddAttr_Call) Return(_a0 error) *Graphvizer_AddAttr_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Graphvizer_AddAttr_Call) RunAndReturn(run func(string, string, string) error) *Graphvizer_AddAttr_Call {
	_c.Call.Return(run)
	return _c
}

// AddEdge provides a mock function with given fields: src, dst, directed, attrs
func (_m *Graphvizer) AddEdge(src string, dst string, directed bool, attrs map[string]string) error {
	ret := _m.Called(src, dst, directed, attrs)

	if len(ret) == 0 {
		panic("no return value specified for AddEdge")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, bool, map[string]string) error); ok {
		r0 = rf(src, dst, directed, attrs)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Graphvizer_AddEdge_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddEdge'
type Graphvizer_AddEdge_Call struct {
	*mock.Call
}

// AddEdge is a helper method to define mock.On call
//   - src string
//   - dst string
//   - directed bool
//   - attrs map[string]string
func (_e *Graphvizer_Expecter) AddEdge(src interface{}, dst interface{}, directed interface{}, attrs interface{}) *Graphvizer_AddEdge_Call {
	return &Graphvizer_AddEdge_Call{Call: _e.mock.On("AddEdge", src, dst, directed, attrs)}
}

func (_c *Graphvizer_AddEdge_Call) Run(run func(src string, dst string, directed bool, attrs map[string]string)) *Graphvizer_AddEdge_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string), args[2].(bool), args[3].(map[string]string))
	})
	return _c
}

func (_c *Graphvizer_AddEdge_Call) Return(_a0 error) *Graphvizer_AddEdge_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Graphvizer_AddEdge_Call) RunAndReturn(run func(string, string, bool, map[string]string) error) *Graphvizer_AddEdge_Call {
	_c.Call.Return(run)
	return _c
}

// AddNode provides a mock function with given fields: parentGraph, name, attrs
func (_m *Graphvizer) AddNode(parentGraph string, name string, attrs map[string]string) error {
	ret := _m.Called(parentGraph, name, attrs)

	if len(ret) == 0 {
		panic("no return value specified for AddNode")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, map[string]string) error); ok {
		r0 = rf(parentGraph, name, attrs)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Graphvizer_AddNode_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddNode'
type Graphvizer_AddNode_Call struct {
	*mock.Call
}

// AddNode is a helper method to define mock.On call
//   - parentGraph string
//   - name string
//   - attrs map[string]string
func (_e *Graphvizer_Expecter) AddNode(parentGraph interface{}, name interface{}, attrs interface{}) *Graphvizer_AddNode_Call {
	return &Graphvizer_AddNode_Call{Call: _e.mock.On("AddNode", parentGraph, name, attrs)}
}

func (_c *Graphvizer_AddNode_Call) Run(run func(parentGraph string, name string, attrs map[string]string)) *Graphvizer_AddNode_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string), args[2].(map[string]string))
	})
	return _c
}

func (_c *Graphvizer_AddNode_Call) Return(_a0 error) *Graphvizer_AddNode_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Graphvizer_AddNode_Call) RunAndReturn(run func(string, string, map[string]string) error) *Graphvizer_AddNode_Call {
	_c.Call.Return(run)
	return _c
}

// AddSubGraph provides a mock function with given fields: parentGraph, name, attrs
func (_m *Graphvizer) AddSubGraph(parentGraph string, name string, attrs map[string]string) error {
	ret := _m.Called(parentGraph, name, attrs)

	if len(ret) == 0 {
		panic("no return value specified for AddSubGraph")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, map[string]string) error); ok {
		r0 = rf(parentGraph, name, attrs)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Graphvizer_AddSubGraph_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'AddSubGraph'
type Graphvizer_AddSubGraph_Call struct {
	*mock.Call
}

// AddSubGraph is a helper method to define mock.On call
//   - parentGraph string
//   - name string
//   - attrs map[string]string
func (_e *Graphvizer_Expecter) AddSubGraph(parentGraph interface{}, name interface{}, attrs interface{}) *Graphvizer_AddSubGraph_Call {
	return &Graphvizer_AddSubGraph_Call{Call: _e.mock.On("AddSubGraph", parentGraph, name, attrs)}
}

func (_c *Graphvizer_AddSubGraph_Call) Run(run func(parentGraph string, name string, attrs map[string]string)) *Graphvizer_AddSubGraph_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string), args[2].(map[string]string))
	})
	return _c
}

func (_c *Graphvizer_AddSubGraph_Call) Return(_a0 error) *Graphvizer_AddSubGraph_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Graphvizer_AddSubGraph_Call) RunAndReturn(run func(string, string, map[string]string) error) *Graphvizer_AddSubGraph_Call {
	_c.Call.Return(run)
	return _c
}

// DoesEdgeExist provides a mock function with given fields: src, dest
func (_m *Graphvizer) DoesEdgeExist(src string, dest string) bool {
	ret := _m.Called(src, dest)

	if len(ret) == 0 {
		panic("no return value specified for DoesEdgeExist")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func(string, string) bool); ok {
		r0 = rf(src, dest)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Graphvizer_DoesEdgeExist_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DoesEdgeExist'
type Graphvizer_DoesEdgeExist_Call struct {
	*mock.Call
}

// DoesEdgeExist is a helper method to define mock.On call
//   - src string
//   - dest string
func (_e *Graphvizer_Expecter) DoesEdgeExist(src interface{}, dest interface{}) *Graphvizer_DoesEdgeExist_Call {
	return &Graphvizer_DoesEdgeExist_Call{Call: _e.mock.On("DoesEdgeExist", src, dest)}
}

func (_c *Graphvizer_DoesEdgeExist_Call) Run(run func(src string, dest string)) *Graphvizer_DoesEdgeExist_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *Graphvizer_DoesEdgeExist_Call) Return(_a0 bool) *Graphvizer_DoesEdgeExist_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Graphvizer_DoesEdgeExist_Call) RunAndReturn(run func(string, string) bool) *Graphvizer_DoesEdgeExist_Call {
	_c.Call.Return(run)
	return _c
}

// GetEdge provides a mock function with given fields: src, dest
func (_m *Graphvizer) GetEdge(src string, dest string) *gographviz.Edge {
	ret := _m.Called(src, dest)

	if len(ret) == 0 {
		panic("no return value specified for GetEdge")
	}

	var r0 *gographviz.Edge
	if rf, ok := ret.Get(0).(func(string, string) *gographviz.Edge); ok {
		r0 = rf(src, dest)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gographviz.Edge)
		}
	}

	return r0
}

// Graphvizer_GetEdge_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetEdge'
type Graphvizer_GetEdge_Call struct {
	*mock.Call
}

// GetEdge is a helper method to define mock.On call
//   - src string
//   - dest string
func (_e *Graphvizer_Expecter) GetEdge(src interface{}, dest interface{}) *Graphvizer_GetEdge_Call {
	return &Graphvizer_GetEdge_Call{Call: _e.mock.On("GetEdge", src, dest)}
}

func (_c *Graphvizer_GetEdge_Call) Run(run func(src string, dest string)) *Graphvizer_GetEdge_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *Graphvizer_GetEdge_Call) Return(_a0 *gographviz.Edge) *Graphvizer_GetEdge_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Graphvizer_GetEdge_Call) RunAndReturn(run func(string, string) *gographviz.Edge) *Graphvizer_GetEdge_Call {
	_c.Call.Return(run)
	return _c
}

// GetNode provides a mock function with given fields: key
func (_m *Graphvizer) GetNode(key string) *gographviz.Node {
	ret := _m.Called(key)

	if len(ret) == 0 {
		panic("no return value specified for GetNode")
	}

	var r0 *gographviz.Node
	if rf, ok := ret.Get(0).(func(string) *gographviz.Node); ok {
		r0 = rf(key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gographviz.Node)
		}
	}

	return r0
}

// Graphvizer_GetNode_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetNode'
type Graphvizer_GetNode_Call struct {
	*mock.Call
}

// GetNode is a helper method to define mock.On call
//   - key string
func (_e *Graphvizer_Expecter) GetNode(key interface{}) *Graphvizer_GetNode_Call {
	return &Graphvizer_GetNode_Call{Call: _e.mock.On("GetNode", key)}
}

func (_c *Graphvizer_GetNode_Call) Run(run func(key string)) *Graphvizer_GetNode_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Graphvizer_GetNode_Call) Return(_a0 *gographviz.Node) *Graphvizer_GetNode_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Graphvizer_GetNode_Call) RunAndReturn(run func(string) *gographviz.Node) *Graphvizer_GetNode_Call {
	_c.Call.Return(run)
	return _c
}

// SetName provides a mock function with given fields: name
func (_m *Graphvizer) SetName(name string) error {
	ret := _m.Called(name)

	if len(ret) == 0 {
		panic("no return value specified for SetName")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(name)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Graphvizer_SetName_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetName'
type Graphvizer_SetName_Call struct {
	*mock.Call
}

// SetName is a helper method to define mock.On call
//   - name string
func (_e *Graphvizer_Expecter) SetName(name interface{}) *Graphvizer_SetName_Call {
	return &Graphvizer_SetName_Call{Call: _e.mock.On("SetName", name)}
}

func (_c *Graphvizer_SetName_Call) Run(run func(name string)) *Graphvizer_SetName_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Graphvizer_SetName_Call) Return(_a0 error) *Graphvizer_SetName_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Graphvizer_SetName_Call) RunAndReturn(run func(string) error) *Graphvizer_SetName_Call {
	_c.Call.Return(run)
	return _c
}

// NewGraphvizer creates a new instance of Graphvizer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewGraphvizer(t interface {
	mock.TestingT
	Cleanup(func())
}) *Graphvizer {
	mock := &Graphvizer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
