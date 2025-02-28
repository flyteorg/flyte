// Code generated by mockery v2.52.1. DO NOT EDIT.

package mocks

import (
	context "context"

	fosite "github.com/ory/fosite"

	http "net/http"

	interfaces "github.com/flyteorg/flyte/flyteadmin/auth/interfaces"

	jwk "github.com/lestrrat-go/jwx/jwk"

	mock "github.com/stretchr/testify/mock"

	oauth2 "github.com/ory/fosite/handler/oauth2"

	service "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
)

// OAuth2Provider is an autogenerated mock type for the OAuth2Provider type
type OAuth2Provider struct {
	mock.Mock
}

type OAuth2Provider_Expecter struct {
	mock *mock.Mock
}

func (_m *OAuth2Provider) EXPECT() *OAuth2Provider_Expecter {
	return &OAuth2Provider_Expecter{mock: &_m.Mock}
}

// IntrospectToken provides a mock function with given fields: ctx, token, tokenUse, session, scope
func (_m *OAuth2Provider) IntrospectToken(ctx context.Context, token string, tokenUse fosite.TokenType, session fosite.Session, scope ...string) (fosite.TokenType, fosite.AccessRequester, error) {
	_va := make([]interface{}, len(scope))
	for _i := range scope {
		_va[_i] = scope[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, token, tokenUse, session)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for IntrospectToken")
	}

	var r0 fosite.TokenType
	var r1 fosite.AccessRequester
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, string, fosite.TokenType, fosite.Session, ...string) (fosite.TokenType, fosite.AccessRequester, error)); ok {
		return rf(ctx, token, tokenUse, session, scope...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, fosite.TokenType, fosite.Session, ...string) fosite.TokenType); ok {
		r0 = rf(ctx, token, tokenUse, session, scope...)
	} else {
		r0 = ret.Get(0).(fosite.TokenType)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, fosite.TokenType, fosite.Session, ...string) fosite.AccessRequester); ok {
		r1 = rf(ctx, token, tokenUse, session, scope...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(fosite.AccessRequester)
		}
	}

	if rf, ok := ret.Get(2).(func(context.Context, string, fosite.TokenType, fosite.Session, ...string) error); ok {
		r2 = rf(ctx, token, tokenUse, session, scope...)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// OAuth2Provider_IntrospectToken_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IntrospectToken'
type OAuth2Provider_IntrospectToken_Call struct {
	*mock.Call
}

// IntrospectToken is a helper method to define mock.On call
//   - ctx context.Context
//   - token string
//   - tokenUse fosite.TokenType
//   - session fosite.Session
//   - scope ...string
func (_e *OAuth2Provider_Expecter) IntrospectToken(ctx interface{}, token interface{}, tokenUse interface{}, session interface{}, scope ...interface{}) *OAuth2Provider_IntrospectToken_Call {
	return &OAuth2Provider_IntrospectToken_Call{Call: _e.mock.On("IntrospectToken",
		append([]interface{}{ctx, token, tokenUse, session}, scope...)...)}
}

func (_c *OAuth2Provider_IntrospectToken_Call) Run(run func(ctx context.Context, token string, tokenUse fosite.TokenType, session fosite.Session, scope ...string)) *OAuth2Provider_IntrospectToken_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]string, len(args)-4)
		for i, a := range args[4:] {
			if a != nil {
				variadicArgs[i] = a.(string)
			}
		}
		run(args[0].(context.Context), args[1].(string), args[2].(fosite.TokenType), args[3].(fosite.Session), variadicArgs...)
	})
	return _c
}

func (_c *OAuth2Provider_IntrospectToken_Call) Return(_a0 fosite.TokenType, _a1 fosite.AccessRequester, _a2 error) *OAuth2Provider_IntrospectToken_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *OAuth2Provider_IntrospectToken_Call) RunAndReturn(run func(context.Context, string, fosite.TokenType, fosite.Session, ...string) (fosite.TokenType, fosite.AccessRequester, error)) *OAuth2Provider_IntrospectToken_Call {
	_c.Call.Return(run)
	return _c
}

// KeySet provides a mock function with no fields
func (_m *OAuth2Provider) KeySet() jwk.Set {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for KeySet")
	}

	var r0 jwk.Set
	if rf, ok := ret.Get(0).(func() jwk.Set); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(jwk.Set)
		}
	}

	return r0
}

// OAuth2Provider_KeySet_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'KeySet'
type OAuth2Provider_KeySet_Call struct {
	*mock.Call
}

// KeySet is a helper method to define mock.On call
func (_e *OAuth2Provider_Expecter) KeySet() *OAuth2Provider_KeySet_Call {
	return &OAuth2Provider_KeySet_Call{Call: _e.mock.On("KeySet")}
}

func (_c *OAuth2Provider_KeySet_Call) Run(run func()) *OAuth2Provider_KeySet_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *OAuth2Provider_KeySet_Call) Return(_a0 jwk.Set) *OAuth2Provider_KeySet_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *OAuth2Provider_KeySet_Call) RunAndReturn(run func() jwk.Set) *OAuth2Provider_KeySet_Call {
	_c.Call.Return(run)
	return _c
}

// NewAccessRequest provides a mock function with given fields: ctx, req, session
func (_m *OAuth2Provider) NewAccessRequest(ctx context.Context, req *http.Request, session fosite.Session) (fosite.AccessRequester, error) {
	ret := _m.Called(ctx, req, session)

	if len(ret) == 0 {
		panic("no return value specified for NewAccessRequest")
	}

	var r0 fosite.AccessRequester
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *http.Request, fosite.Session) (fosite.AccessRequester, error)); ok {
		return rf(ctx, req, session)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *http.Request, fosite.Session) fosite.AccessRequester); ok {
		r0 = rf(ctx, req, session)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(fosite.AccessRequester)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *http.Request, fosite.Session) error); ok {
		r1 = rf(ctx, req, session)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OAuth2Provider_NewAccessRequest_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'NewAccessRequest'
type OAuth2Provider_NewAccessRequest_Call struct {
	*mock.Call
}

// NewAccessRequest is a helper method to define mock.On call
//   - ctx context.Context
//   - req *http.Request
//   - session fosite.Session
func (_e *OAuth2Provider_Expecter) NewAccessRequest(ctx interface{}, req interface{}, session interface{}) *OAuth2Provider_NewAccessRequest_Call {
	return &OAuth2Provider_NewAccessRequest_Call{Call: _e.mock.On("NewAccessRequest", ctx, req, session)}
}

func (_c *OAuth2Provider_NewAccessRequest_Call) Run(run func(ctx context.Context, req *http.Request, session fosite.Session)) *OAuth2Provider_NewAccessRequest_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*http.Request), args[2].(fosite.Session))
	})
	return _c
}

func (_c *OAuth2Provider_NewAccessRequest_Call) Return(_a0 fosite.AccessRequester, _a1 error) *OAuth2Provider_NewAccessRequest_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *OAuth2Provider_NewAccessRequest_Call) RunAndReturn(run func(context.Context, *http.Request, fosite.Session) (fosite.AccessRequester, error)) *OAuth2Provider_NewAccessRequest_Call {
	_c.Call.Return(run)
	return _c
}

// NewAccessResponse provides a mock function with given fields: ctx, requester
func (_m *OAuth2Provider) NewAccessResponse(ctx context.Context, requester fosite.AccessRequester) (fosite.AccessResponder, error) {
	ret := _m.Called(ctx, requester)

	if len(ret) == 0 {
		panic("no return value specified for NewAccessResponse")
	}

	var r0 fosite.AccessResponder
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, fosite.AccessRequester) (fosite.AccessResponder, error)); ok {
		return rf(ctx, requester)
	}
	if rf, ok := ret.Get(0).(func(context.Context, fosite.AccessRequester) fosite.AccessResponder); ok {
		r0 = rf(ctx, requester)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(fosite.AccessResponder)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, fosite.AccessRequester) error); ok {
		r1 = rf(ctx, requester)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OAuth2Provider_NewAccessResponse_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'NewAccessResponse'
type OAuth2Provider_NewAccessResponse_Call struct {
	*mock.Call
}

// NewAccessResponse is a helper method to define mock.On call
//   - ctx context.Context
//   - requester fosite.AccessRequester
func (_e *OAuth2Provider_Expecter) NewAccessResponse(ctx interface{}, requester interface{}) *OAuth2Provider_NewAccessResponse_Call {
	return &OAuth2Provider_NewAccessResponse_Call{Call: _e.mock.On("NewAccessResponse", ctx, requester)}
}

func (_c *OAuth2Provider_NewAccessResponse_Call) Run(run func(ctx context.Context, requester fosite.AccessRequester)) *OAuth2Provider_NewAccessResponse_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(fosite.AccessRequester))
	})
	return _c
}

func (_c *OAuth2Provider_NewAccessResponse_Call) Return(_a0 fosite.AccessResponder, _a1 error) *OAuth2Provider_NewAccessResponse_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *OAuth2Provider_NewAccessResponse_Call) RunAndReturn(run func(context.Context, fosite.AccessRequester) (fosite.AccessResponder, error)) *OAuth2Provider_NewAccessResponse_Call {
	_c.Call.Return(run)
	return _c
}

// NewAuthorizeRequest provides a mock function with given fields: ctx, req
func (_m *OAuth2Provider) NewAuthorizeRequest(ctx context.Context, req *http.Request) (fosite.AuthorizeRequester, error) {
	ret := _m.Called(ctx, req)

	if len(ret) == 0 {
		panic("no return value specified for NewAuthorizeRequest")
	}

	var r0 fosite.AuthorizeRequester
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *http.Request) (fosite.AuthorizeRequester, error)); ok {
		return rf(ctx, req)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *http.Request) fosite.AuthorizeRequester); ok {
		r0 = rf(ctx, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(fosite.AuthorizeRequester)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *http.Request) error); ok {
		r1 = rf(ctx, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OAuth2Provider_NewAuthorizeRequest_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'NewAuthorizeRequest'
type OAuth2Provider_NewAuthorizeRequest_Call struct {
	*mock.Call
}

// NewAuthorizeRequest is a helper method to define mock.On call
//   - ctx context.Context
//   - req *http.Request
func (_e *OAuth2Provider_Expecter) NewAuthorizeRequest(ctx interface{}, req interface{}) *OAuth2Provider_NewAuthorizeRequest_Call {
	return &OAuth2Provider_NewAuthorizeRequest_Call{Call: _e.mock.On("NewAuthorizeRequest", ctx, req)}
}

func (_c *OAuth2Provider_NewAuthorizeRequest_Call) Run(run func(ctx context.Context, req *http.Request)) *OAuth2Provider_NewAuthorizeRequest_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*http.Request))
	})
	return _c
}

func (_c *OAuth2Provider_NewAuthorizeRequest_Call) Return(_a0 fosite.AuthorizeRequester, _a1 error) *OAuth2Provider_NewAuthorizeRequest_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *OAuth2Provider_NewAuthorizeRequest_Call) RunAndReturn(run func(context.Context, *http.Request) (fosite.AuthorizeRequester, error)) *OAuth2Provider_NewAuthorizeRequest_Call {
	_c.Call.Return(run)
	return _c
}

// NewAuthorizeResponse provides a mock function with given fields: ctx, requester, session
func (_m *OAuth2Provider) NewAuthorizeResponse(ctx context.Context, requester fosite.AuthorizeRequester, session fosite.Session) (fosite.AuthorizeResponder, error) {
	ret := _m.Called(ctx, requester, session)

	if len(ret) == 0 {
		panic("no return value specified for NewAuthorizeResponse")
	}

	var r0 fosite.AuthorizeResponder
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, fosite.AuthorizeRequester, fosite.Session) (fosite.AuthorizeResponder, error)); ok {
		return rf(ctx, requester, session)
	}
	if rf, ok := ret.Get(0).(func(context.Context, fosite.AuthorizeRequester, fosite.Session) fosite.AuthorizeResponder); ok {
		r0 = rf(ctx, requester, session)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(fosite.AuthorizeResponder)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, fosite.AuthorizeRequester, fosite.Session) error); ok {
		r1 = rf(ctx, requester, session)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OAuth2Provider_NewAuthorizeResponse_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'NewAuthorizeResponse'
type OAuth2Provider_NewAuthorizeResponse_Call struct {
	*mock.Call
}

// NewAuthorizeResponse is a helper method to define mock.On call
//   - ctx context.Context
//   - requester fosite.AuthorizeRequester
//   - session fosite.Session
func (_e *OAuth2Provider_Expecter) NewAuthorizeResponse(ctx interface{}, requester interface{}, session interface{}) *OAuth2Provider_NewAuthorizeResponse_Call {
	return &OAuth2Provider_NewAuthorizeResponse_Call{Call: _e.mock.On("NewAuthorizeResponse", ctx, requester, session)}
}

func (_c *OAuth2Provider_NewAuthorizeResponse_Call) Run(run func(ctx context.Context, requester fosite.AuthorizeRequester, session fosite.Session)) *OAuth2Provider_NewAuthorizeResponse_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(fosite.AuthorizeRequester), args[2].(fosite.Session))
	})
	return _c
}

func (_c *OAuth2Provider_NewAuthorizeResponse_Call) Return(_a0 fosite.AuthorizeResponder, _a1 error) *OAuth2Provider_NewAuthorizeResponse_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *OAuth2Provider_NewAuthorizeResponse_Call) RunAndReturn(run func(context.Context, fosite.AuthorizeRequester, fosite.Session) (fosite.AuthorizeResponder, error)) *OAuth2Provider_NewAuthorizeResponse_Call {
	_c.Call.Return(run)
	return _c
}

// NewIntrospectionRequest provides a mock function with given fields: ctx, r, session
func (_m *OAuth2Provider) NewIntrospectionRequest(ctx context.Context, r *http.Request, session fosite.Session) (fosite.IntrospectionResponder, error) {
	ret := _m.Called(ctx, r, session)

	if len(ret) == 0 {
		panic("no return value specified for NewIntrospectionRequest")
	}

	var r0 fosite.IntrospectionResponder
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *http.Request, fosite.Session) (fosite.IntrospectionResponder, error)); ok {
		return rf(ctx, r, session)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *http.Request, fosite.Session) fosite.IntrospectionResponder); ok {
		r0 = rf(ctx, r, session)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(fosite.IntrospectionResponder)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *http.Request, fosite.Session) error); ok {
		r1 = rf(ctx, r, session)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OAuth2Provider_NewIntrospectionRequest_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'NewIntrospectionRequest'
type OAuth2Provider_NewIntrospectionRequest_Call struct {
	*mock.Call
}

// NewIntrospectionRequest is a helper method to define mock.On call
//   - ctx context.Context
//   - r *http.Request
//   - session fosite.Session
func (_e *OAuth2Provider_Expecter) NewIntrospectionRequest(ctx interface{}, r interface{}, session interface{}) *OAuth2Provider_NewIntrospectionRequest_Call {
	return &OAuth2Provider_NewIntrospectionRequest_Call{Call: _e.mock.On("NewIntrospectionRequest", ctx, r, session)}
}

func (_c *OAuth2Provider_NewIntrospectionRequest_Call) Run(run func(ctx context.Context, r *http.Request, session fosite.Session)) *OAuth2Provider_NewIntrospectionRequest_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*http.Request), args[2].(fosite.Session))
	})
	return _c
}

func (_c *OAuth2Provider_NewIntrospectionRequest_Call) Return(_a0 fosite.IntrospectionResponder, _a1 error) *OAuth2Provider_NewIntrospectionRequest_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *OAuth2Provider_NewIntrospectionRequest_Call) RunAndReturn(run func(context.Context, *http.Request, fosite.Session) (fosite.IntrospectionResponder, error)) *OAuth2Provider_NewIntrospectionRequest_Call {
	_c.Call.Return(run)
	return _c
}

// NewJWTSessionToken provides a mock function with given fields: subject, appID, issuer, audience, userInfoClaims
func (_m *OAuth2Provider) NewJWTSessionToken(subject string, appID string, issuer string, audience string, userInfoClaims *service.UserInfoResponse) *oauth2.JWTSession {
	ret := _m.Called(subject, appID, issuer, audience, userInfoClaims)

	if len(ret) == 0 {
		panic("no return value specified for NewJWTSessionToken")
	}

	var r0 *oauth2.JWTSession
	if rf, ok := ret.Get(0).(func(string, string, string, string, *service.UserInfoResponse) *oauth2.JWTSession); ok {
		r0 = rf(subject, appID, issuer, audience, userInfoClaims)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*oauth2.JWTSession)
		}
	}

	return r0
}

// OAuth2Provider_NewJWTSessionToken_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'NewJWTSessionToken'
type OAuth2Provider_NewJWTSessionToken_Call struct {
	*mock.Call
}

// NewJWTSessionToken is a helper method to define mock.On call
//   - subject string
//   - appID string
//   - issuer string
//   - audience string
//   - userInfoClaims *service.UserInfoResponse
func (_e *OAuth2Provider_Expecter) NewJWTSessionToken(subject interface{}, appID interface{}, issuer interface{}, audience interface{}, userInfoClaims interface{}) *OAuth2Provider_NewJWTSessionToken_Call {
	return &OAuth2Provider_NewJWTSessionToken_Call{Call: _e.mock.On("NewJWTSessionToken", subject, appID, issuer, audience, userInfoClaims)}
}

func (_c *OAuth2Provider_NewJWTSessionToken_Call) Run(run func(subject string, appID string, issuer string, audience string, userInfoClaims *service.UserInfoResponse)) *OAuth2Provider_NewJWTSessionToken_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string), args[2].(string), args[3].(string), args[4].(*service.UserInfoResponse))
	})
	return _c
}

func (_c *OAuth2Provider_NewJWTSessionToken_Call) Return(_a0 *oauth2.JWTSession) *OAuth2Provider_NewJWTSessionToken_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *OAuth2Provider_NewJWTSessionToken_Call) RunAndReturn(run func(string, string, string, string, *service.UserInfoResponse) *oauth2.JWTSession) *OAuth2Provider_NewJWTSessionToken_Call {
	_c.Call.Return(run)
	return _c
}

// NewRevocationRequest provides a mock function with given fields: ctx, r
func (_m *OAuth2Provider) NewRevocationRequest(ctx context.Context, r *http.Request) error {
	ret := _m.Called(ctx, r)

	if len(ret) == 0 {
		panic("no return value specified for NewRevocationRequest")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *http.Request) error); ok {
		r0 = rf(ctx, r)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// OAuth2Provider_NewRevocationRequest_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'NewRevocationRequest'
type OAuth2Provider_NewRevocationRequest_Call struct {
	*mock.Call
}

// NewRevocationRequest is a helper method to define mock.On call
//   - ctx context.Context
//   - r *http.Request
func (_e *OAuth2Provider_Expecter) NewRevocationRequest(ctx interface{}, r interface{}) *OAuth2Provider_NewRevocationRequest_Call {
	return &OAuth2Provider_NewRevocationRequest_Call{Call: _e.mock.On("NewRevocationRequest", ctx, r)}
}

func (_c *OAuth2Provider_NewRevocationRequest_Call) Run(run func(ctx context.Context, r *http.Request)) *OAuth2Provider_NewRevocationRequest_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*http.Request))
	})
	return _c
}

func (_c *OAuth2Provider_NewRevocationRequest_Call) Return(_a0 error) *OAuth2Provider_NewRevocationRequest_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *OAuth2Provider_NewRevocationRequest_Call) RunAndReturn(run func(context.Context, *http.Request) error) *OAuth2Provider_NewRevocationRequest_Call {
	_c.Call.Return(run)
	return _c
}

// ValidateAccessToken provides a mock function with given fields: ctx, expectedAudience, tokenStr
func (_m *OAuth2Provider) ValidateAccessToken(ctx context.Context, expectedAudience string, tokenStr string) (interfaces.IdentityContext, error) {
	ret := _m.Called(ctx, expectedAudience, tokenStr)

	if len(ret) == 0 {
		panic("no return value specified for ValidateAccessToken")
	}

	var r0 interfaces.IdentityContext
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (interfaces.IdentityContext, error)); ok {
		return rf(ctx, expectedAudience, tokenStr)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) interfaces.IdentityContext); ok {
		r0 = rf(ctx, expectedAudience, tokenStr)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interfaces.IdentityContext)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, expectedAudience, tokenStr)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OAuth2Provider_ValidateAccessToken_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ValidateAccessToken'
type OAuth2Provider_ValidateAccessToken_Call struct {
	*mock.Call
}

// ValidateAccessToken is a helper method to define mock.On call
//   - ctx context.Context
//   - expectedAudience string
//   - tokenStr string
func (_e *OAuth2Provider_Expecter) ValidateAccessToken(ctx interface{}, expectedAudience interface{}, tokenStr interface{}) *OAuth2Provider_ValidateAccessToken_Call {
	return &OAuth2Provider_ValidateAccessToken_Call{Call: _e.mock.On("ValidateAccessToken", ctx, expectedAudience, tokenStr)}
}

func (_c *OAuth2Provider_ValidateAccessToken_Call) Run(run func(ctx context.Context, expectedAudience string, tokenStr string)) *OAuth2Provider_ValidateAccessToken_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(string))
	})
	return _c
}

func (_c *OAuth2Provider_ValidateAccessToken_Call) Return(_a0 interfaces.IdentityContext, _a1 error) *OAuth2Provider_ValidateAccessToken_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *OAuth2Provider_ValidateAccessToken_Call) RunAndReturn(run func(context.Context, string, string) (interfaces.IdentityContext, error)) *OAuth2Provider_ValidateAccessToken_Call {
	_c.Call.Return(run)
	return _c
}

// WriteAccessError provides a mock function with given fields: rw, requester, err
func (_m *OAuth2Provider) WriteAccessError(rw http.ResponseWriter, requester fosite.AccessRequester, err error) {
	_m.Called(rw, requester, err)
}

// OAuth2Provider_WriteAccessError_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WriteAccessError'
type OAuth2Provider_WriteAccessError_Call struct {
	*mock.Call
}

// WriteAccessError is a helper method to define mock.On call
//   - rw http.ResponseWriter
//   - requester fosite.AccessRequester
//   - err error
func (_e *OAuth2Provider_Expecter) WriteAccessError(rw interface{}, requester interface{}, err interface{}) *OAuth2Provider_WriteAccessError_Call {
	return &OAuth2Provider_WriteAccessError_Call{Call: _e.mock.On("WriteAccessError", rw, requester, err)}
}

func (_c *OAuth2Provider_WriteAccessError_Call) Run(run func(rw http.ResponseWriter, requester fosite.AccessRequester, err error)) *OAuth2Provider_WriteAccessError_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(http.ResponseWriter), args[1].(fosite.AccessRequester), args[2].(error))
	})
	return _c
}

func (_c *OAuth2Provider_WriteAccessError_Call) Return() *OAuth2Provider_WriteAccessError_Call {
	_c.Call.Return()
	return _c
}

func (_c *OAuth2Provider_WriteAccessError_Call) RunAndReturn(run func(http.ResponseWriter, fosite.AccessRequester, error)) *OAuth2Provider_WriteAccessError_Call {
	_c.Run(run)
	return _c
}

// WriteAccessResponse provides a mock function with given fields: rw, requester, responder
func (_m *OAuth2Provider) WriteAccessResponse(rw http.ResponseWriter, requester fosite.AccessRequester, responder fosite.AccessResponder) {
	_m.Called(rw, requester, responder)
}

// OAuth2Provider_WriteAccessResponse_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WriteAccessResponse'
type OAuth2Provider_WriteAccessResponse_Call struct {
	*mock.Call
}

// WriteAccessResponse is a helper method to define mock.On call
//   - rw http.ResponseWriter
//   - requester fosite.AccessRequester
//   - responder fosite.AccessResponder
func (_e *OAuth2Provider_Expecter) WriteAccessResponse(rw interface{}, requester interface{}, responder interface{}) *OAuth2Provider_WriteAccessResponse_Call {
	return &OAuth2Provider_WriteAccessResponse_Call{Call: _e.mock.On("WriteAccessResponse", rw, requester, responder)}
}

func (_c *OAuth2Provider_WriteAccessResponse_Call) Run(run func(rw http.ResponseWriter, requester fosite.AccessRequester, responder fosite.AccessResponder)) *OAuth2Provider_WriteAccessResponse_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(http.ResponseWriter), args[1].(fosite.AccessRequester), args[2].(fosite.AccessResponder))
	})
	return _c
}

func (_c *OAuth2Provider_WriteAccessResponse_Call) Return() *OAuth2Provider_WriteAccessResponse_Call {
	_c.Call.Return()
	return _c
}

func (_c *OAuth2Provider_WriteAccessResponse_Call) RunAndReturn(run func(http.ResponseWriter, fosite.AccessRequester, fosite.AccessResponder)) *OAuth2Provider_WriteAccessResponse_Call {
	_c.Run(run)
	return _c
}

// WriteAuthorizeError provides a mock function with given fields: rw, requester, err
func (_m *OAuth2Provider) WriteAuthorizeError(rw http.ResponseWriter, requester fosite.AuthorizeRequester, err error) {
	_m.Called(rw, requester, err)
}

// OAuth2Provider_WriteAuthorizeError_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WriteAuthorizeError'
type OAuth2Provider_WriteAuthorizeError_Call struct {
	*mock.Call
}

// WriteAuthorizeError is a helper method to define mock.On call
//   - rw http.ResponseWriter
//   - requester fosite.AuthorizeRequester
//   - err error
func (_e *OAuth2Provider_Expecter) WriteAuthorizeError(rw interface{}, requester interface{}, err interface{}) *OAuth2Provider_WriteAuthorizeError_Call {
	return &OAuth2Provider_WriteAuthorizeError_Call{Call: _e.mock.On("WriteAuthorizeError", rw, requester, err)}
}

func (_c *OAuth2Provider_WriteAuthorizeError_Call) Run(run func(rw http.ResponseWriter, requester fosite.AuthorizeRequester, err error)) *OAuth2Provider_WriteAuthorizeError_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(http.ResponseWriter), args[1].(fosite.AuthorizeRequester), args[2].(error))
	})
	return _c
}

func (_c *OAuth2Provider_WriteAuthorizeError_Call) Return() *OAuth2Provider_WriteAuthorizeError_Call {
	_c.Call.Return()
	return _c
}

func (_c *OAuth2Provider_WriteAuthorizeError_Call) RunAndReturn(run func(http.ResponseWriter, fosite.AuthorizeRequester, error)) *OAuth2Provider_WriteAuthorizeError_Call {
	_c.Run(run)
	return _c
}

// WriteAuthorizeResponse provides a mock function with given fields: rw, requester, responder
func (_m *OAuth2Provider) WriteAuthorizeResponse(rw http.ResponseWriter, requester fosite.AuthorizeRequester, responder fosite.AuthorizeResponder) {
	_m.Called(rw, requester, responder)
}

// OAuth2Provider_WriteAuthorizeResponse_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WriteAuthorizeResponse'
type OAuth2Provider_WriteAuthorizeResponse_Call struct {
	*mock.Call
}

// WriteAuthorizeResponse is a helper method to define mock.On call
//   - rw http.ResponseWriter
//   - requester fosite.AuthorizeRequester
//   - responder fosite.AuthorizeResponder
func (_e *OAuth2Provider_Expecter) WriteAuthorizeResponse(rw interface{}, requester interface{}, responder interface{}) *OAuth2Provider_WriteAuthorizeResponse_Call {
	return &OAuth2Provider_WriteAuthorizeResponse_Call{Call: _e.mock.On("WriteAuthorizeResponse", rw, requester, responder)}
}

func (_c *OAuth2Provider_WriteAuthorizeResponse_Call) Run(run func(rw http.ResponseWriter, requester fosite.AuthorizeRequester, responder fosite.AuthorizeResponder)) *OAuth2Provider_WriteAuthorizeResponse_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(http.ResponseWriter), args[1].(fosite.AuthorizeRequester), args[2].(fosite.AuthorizeResponder))
	})
	return _c
}

func (_c *OAuth2Provider_WriteAuthorizeResponse_Call) Return() *OAuth2Provider_WriteAuthorizeResponse_Call {
	_c.Call.Return()
	return _c
}

func (_c *OAuth2Provider_WriteAuthorizeResponse_Call) RunAndReturn(run func(http.ResponseWriter, fosite.AuthorizeRequester, fosite.AuthorizeResponder)) *OAuth2Provider_WriteAuthorizeResponse_Call {
	_c.Run(run)
	return _c
}

// WriteIntrospectionError provides a mock function with given fields: rw, err
func (_m *OAuth2Provider) WriteIntrospectionError(rw http.ResponseWriter, err error) {
	_m.Called(rw, err)
}

// OAuth2Provider_WriteIntrospectionError_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WriteIntrospectionError'
type OAuth2Provider_WriteIntrospectionError_Call struct {
	*mock.Call
}

// WriteIntrospectionError is a helper method to define mock.On call
//   - rw http.ResponseWriter
//   - err error
func (_e *OAuth2Provider_Expecter) WriteIntrospectionError(rw interface{}, err interface{}) *OAuth2Provider_WriteIntrospectionError_Call {
	return &OAuth2Provider_WriteIntrospectionError_Call{Call: _e.mock.On("WriteIntrospectionError", rw, err)}
}

func (_c *OAuth2Provider_WriteIntrospectionError_Call) Run(run func(rw http.ResponseWriter, err error)) *OAuth2Provider_WriteIntrospectionError_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(http.ResponseWriter), args[1].(error))
	})
	return _c
}

func (_c *OAuth2Provider_WriteIntrospectionError_Call) Return() *OAuth2Provider_WriteIntrospectionError_Call {
	_c.Call.Return()
	return _c
}

func (_c *OAuth2Provider_WriteIntrospectionError_Call) RunAndReturn(run func(http.ResponseWriter, error)) *OAuth2Provider_WriteIntrospectionError_Call {
	_c.Run(run)
	return _c
}

// WriteIntrospectionResponse provides a mock function with given fields: rw, r
func (_m *OAuth2Provider) WriteIntrospectionResponse(rw http.ResponseWriter, r fosite.IntrospectionResponder) {
	_m.Called(rw, r)
}

// OAuth2Provider_WriteIntrospectionResponse_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WriteIntrospectionResponse'
type OAuth2Provider_WriteIntrospectionResponse_Call struct {
	*mock.Call
}

// WriteIntrospectionResponse is a helper method to define mock.On call
//   - rw http.ResponseWriter
//   - r fosite.IntrospectionResponder
func (_e *OAuth2Provider_Expecter) WriteIntrospectionResponse(rw interface{}, r interface{}) *OAuth2Provider_WriteIntrospectionResponse_Call {
	return &OAuth2Provider_WriteIntrospectionResponse_Call{Call: _e.mock.On("WriteIntrospectionResponse", rw, r)}
}

func (_c *OAuth2Provider_WriteIntrospectionResponse_Call) Run(run func(rw http.ResponseWriter, r fosite.IntrospectionResponder)) *OAuth2Provider_WriteIntrospectionResponse_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(http.ResponseWriter), args[1].(fosite.IntrospectionResponder))
	})
	return _c
}

func (_c *OAuth2Provider_WriteIntrospectionResponse_Call) Return() *OAuth2Provider_WriteIntrospectionResponse_Call {
	_c.Call.Return()
	return _c
}

func (_c *OAuth2Provider_WriteIntrospectionResponse_Call) RunAndReturn(run func(http.ResponseWriter, fosite.IntrospectionResponder)) *OAuth2Provider_WriteIntrospectionResponse_Call {
	_c.Run(run)
	return _c
}

// WriteRevocationResponse provides a mock function with given fields: rw, err
func (_m *OAuth2Provider) WriteRevocationResponse(rw http.ResponseWriter, err error) {
	_m.Called(rw, err)
}

// OAuth2Provider_WriteRevocationResponse_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'WriteRevocationResponse'
type OAuth2Provider_WriteRevocationResponse_Call struct {
	*mock.Call
}

// WriteRevocationResponse is a helper method to define mock.On call
//   - rw http.ResponseWriter
//   - err error
func (_e *OAuth2Provider_Expecter) WriteRevocationResponse(rw interface{}, err interface{}) *OAuth2Provider_WriteRevocationResponse_Call {
	return &OAuth2Provider_WriteRevocationResponse_Call{Call: _e.mock.On("WriteRevocationResponse", rw, err)}
}

func (_c *OAuth2Provider_WriteRevocationResponse_Call) Run(run func(rw http.ResponseWriter, err error)) *OAuth2Provider_WriteRevocationResponse_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(http.ResponseWriter), args[1].(error))
	})
	return _c
}

func (_c *OAuth2Provider_WriteRevocationResponse_Call) Return() *OAuth2Provider_WriteRevocationResponse_Call {
	_c.Call.Return()
	return _c
}

func (_c *OAuth2Provider_WriteRevocationResponse_Call) RunAndReturn(run func(http.ResponseWriter, error)) *OAuth2Provider_WriteRevocationResponse_Call {
	_c.Run(run)
	return _c
}

// NewOAuth2Provider creates a new instance of OAuth2Provider. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewOAuth2Provider(t interface {
	mock.TestingT
	Cleanup(func())
}) *OAuth2Provider {
	mock := &OAuth2Provider{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
