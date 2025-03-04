// Code generated by mockery v2.53.0. DO NOT EDIT.

package mocks

import (
	io "io"
	smtp "net/smtp"

	mock "github.com/stretchr/testify/mock"

	tls "crypto/tls"
)

// SMTPClient is an autogenerated mock type for the SMTPClient type
type SMTPClient struct {
	mock.Mock
}

type SMTPClient_Expecter struct {
	mock *mock.Mock
}

func (_m *SMTPClient) EXPECT() *SMTPClient_Expecter {
	return &SMTPClient_Expecter{mock: &_m.Mock}
}

// Auth provides a mock function with given fields: a
func (_m *SMTPClient) Auth(a smtp.Auth) error {
	ret := _m.Called(a)

	if len(ret) == 0 {
		panic("no return value specified for Auth")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(smtp.Auth) error); ok {
		r0 = rf(a)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SMTPClient_Auth_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Auth'
type SMTPClient_Auth_Call struct {
	*mock.Call
}

// Auth is a helper method to define mock.On call
//   - a smtp.Auth
func (_e *SMTPClient_Expecter) Auth(a interface{}) *SMTPClient_Auth_Call {
	return &SMTPClient_Auth_Call{Call: _e.mock.On("Auth", a)}
}

func (_c *SMTPClient_Auth_Call) Run(run func(a smtp.Auth)) *SMTPClient_Auth_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(smtp.Auth))
	})
	return _c
}

func (_c *SMTPClient_Auth_Call) Return(_a0 error) *SMTPClient_Auth_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SMTPClient_Auth_Call) RunAndReturn(run func(smtp.Auth) error) *SMTPClient_Auth_Call {
	_c.Call.Return(run)
	return _c
}

// Close provides a mock function with no fields
func (_m *SMTPClient) Close() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Close")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SMTPClient_Close_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Close'
type SMTPClient_Close_Call struct {
	*mock.Call
}

// Close is a helper method to define mock.On call
func (_e *SMTPClient_Expecter) Close() *SMTPClient_Close_Call {
	return &SMTPClient_Close_Call{Call: _e.mock.On("Close")}
}

func (_c *SMTPClient_Close_Call) Run(run func()) *SMTPClient_Close_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *SMTPClient_Close_Call) Return(_a0 error) *SMTPClient_Close_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SMTPClient_Close_Call) RunAndReturn(run func() error) *SMTPClient_Close_Call {
	_c.Call.Return(run)
	return _c
}

// Data provides a mock function with no fields
func (_m *SMTPClient) Data() (io.WriteCloser, error) {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Data")
	}

	var r0 io.WriteCloser
	var r1 error
	if rf, ok := ret.Get(0).(func() (io.WriteCloser, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() io.WriteCloser); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.WriteCloser)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SMTPClient_Data_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Data'
type SMTPClient_Data_Call struct {
	*mock.Call
}

// Data is a helper method to define mock.On call
func (_e *SMTPClient_Expecter) Data() *SMTPClient_Data_Call {
	return &SMTPClient_Data_Call{Call: _e.mock.On("Data")}
}

func (_c *SMTPClient_Data_Call) Run(run func()) *SMTPClient_Data_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *SMTPClient_Data_Call) Return(_a0 io.WriteCloser, _a1 error) *SMTPClient_Data_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *SMTPClient_Data_Call) RunAndReturn(run func() (io.WriteCloser, error)) *SMTPClient_Data_Call {
	_c.Call.Return(run)
	return _c
}

// Extension provides a mock function with given fields: ext
func (_m *SMTPClient) Extension(ext string) (bool, string) {
	ret := _m.Called(ext)

	if len(ret) == 0 {
		panic("no return value specified for Extension")
	}

	var r0 bool
	var r1 string
	if rf, ok := ret.Get(0).(func(string) (bool, string)); ok {
		return rf(ext)
	}
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(ext)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(string) string); ok {
		r1 = rf(ext)
	} else {
		r1 = ret.Get(1).(string)
	}

	return r0, r1
}

// SMTPClient_Extension_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Extension'
type SMTPClient_Extension_Call struct {
	*mock.Call
}

// Extension is a helper method to define mock.On call
//   - ext string
func (_e *SMTPClient_Expecter) Extension(ext interface{}) *SMTPClient_Extension_Call {
	return &SMTPClient_Extension_Call{Call: _e.mock.On("Extension", ext)}
}

func (_c *SMTPClient_Extension_Call) Run(run func(ext string)) *SMTPClient_Extension_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *SMTPClient_Extension_Call) Return(_a0 bool, _a1 string) *SMTPClient_Extension_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *SMTPClient_Extension_Call) RunAndReturn(run func(string) (bool, string)) *SMTPClient_Extension_Call {
	_c.Call.Return(run)
	return _c
}

// Hello provides a mock function with given fields: localName
func (_m *SMTPClient) Hello(localName string) error {
	ret := _m.Called(localName)

	if len(ret) == 0 {
		panic("no return value specified for Hello")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(localName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SMTPClient_Hello_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Hello'
type SMTPClient_Hello_Call struct {
	*mock.Call
}

// Hello is a helper method to define mock.On call
//   - localName string
func (_e *SMTPClient_Expecter) Hello(localName interface{}) *SMTPClient_Hello_Call {
	return &SMTPClient_Hello_Call{Call: _e.mock.On("Hello", localName)}
}

func (_c *SMTPClient_Hello_Call) Run(run func(localName string)) *SMTPClient_Hello_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *SMTPClient_Hello_Call) Return(_a0 error) *SMTPClient_Hello_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SMTPClient_Hello_Call) RunAndReturn(run func(string) error) *SMTPClient_Hello_Call {
	_c.Call.Return(run)
	return _c
}

// Mail provides a mock function with given fields: from
func (_m *SMTPClient) Mail(from string) error {
	ret := _m.Called(from)

	if len(ret) == 0 {
		panic("no return value specified for Mail")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(from)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SMTPClient_Mail_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Mail'
type SMTPClient_Mail_Call struct {
	*mock.Call
}

// Mail is a helper method to define mock.On call
//   - from string
func (_e *SMTPClient_Expecter) Mail(from interface{}) *SMTPClient_Mail_Call {
	return &SMTPClient_Mail_Call{Call: _e.mock.On("Mail", from)}
}

func (_c *SMTPClient_Mail_Call) Run(run func(from string)) *SMTPClient_Mail_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *SMTPClient_Mail_Call) Return(_a0 error) *SMTPClient_Mail_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SMTPClient_Mail_Call) RunAndReturn(run func(string) error) *SMTPClient_Mail_Call {
	_c.Call.Return(run)
	return _c
}

// Noop provides a mock function with no fields
func (_m *SMTPClient) Noop() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Noop")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SMTPClient_Noop_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Noop'
type SMTPClient_Noop_Call struct {
	*mock.Call
}

// Noop is a helper method to define mock.On call
func (_e *SMTPClient_Expecter) Noop() *SMTPClient_Noop_Call {
	return &SMTPClient_Noop_Call{Call: _e.mock.On("Noop")}
}

func (_c *SMTPClient_Noop_Call) Run(run func()) *SMTPClient_Noop_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *SMTPClient_Noop_Call) Return(_a0 error) *SMTPClient_Noop_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SMTPClient_Noop_Call) RunAndReturn(run func() error) *SMTPClient_Noop_Call {
	_c.Call.Return(run)
	return _c
}

// Rcpt provides a mock function with given fields: to
func (_m *SMTPClient) Rcpt(to string) error {
	ret := _m.Called(to)

	if len(ret) == 0 {
		panic("no return value specified for Rcpt")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(to)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SMTPClient_Rcpt_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Rcpt'
type SMTPClient_Rcpt_Call struct {
	*mock.Call
}

// Rcpt is a helper method to define mock.On call
//   - to string
func (_e *SMTPClient_Expecter) Rcpt(to interface{}) *SMTPClient_Rcpt_Call {
	return &SMTPClient_Rcpt_Call{Call: _e.mock.On("Rcpt", to)}
}

func (_c *SMTPClient_Rcpt_Call) Run(run func(to string)) *SMTPClient_Rcpt_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *SMTPClient_Rcpt_Call) Return(_a0 error) *SMTPClient_Rcpt_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SMTPClient_Rcpt_Call) RunAndReturn(run func(string) error) *SMTPClient_Rcpt_Call {
	_c.Call.Return(run)
	return _c
}

// StartTLS provides a mock function with given fields: config
func (_m *SMTPClient) StartTLS(config *tls.Config) error {
	ret := _m.Called(config)

	if len(ret) == 0 {
		panic("no return value specified for StartTLS")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(*tls.Config) error); ok {
		r0 = rf(config)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SMTPClient_StartTLS_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'StartTLS'
type SMTPClient_StartTLS_Call struct {
	*mock.Call
}

// StartTLS is a helper method to define mock.On call
//   - config *tls.Config
func (_e *SMTPClient_Expecter) StartTLS(config interface{}) *SMTPClient_StartTLS_Call {
	return &SMTPClient_StartTLS_Call{Call: _e.mock.On("StartTLS", config)}
}

func (_c *SMTPClient_StartTLS_Call) Run(run func(config *tls.Config)) *SMTPClient_StartTLS_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*tls.Config))
	})
	return _c
}

func (_c *SMTPClient_StartTLS_Call) Return(_a0 error) *SMTPClient_StartTLS_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SMTPClient_StartTLS_Call) RunAndReturn(run func(*tls.Config) error) *SMTPClient_StartTLS_Call {
	_c.Call.Return(run)
	return _c
}

// NewSMTPClient creates a new instance of SMTPClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSMTPClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *SMTPClient {
	mock := &SMTPClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
