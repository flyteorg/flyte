package interceptors

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyte/flyteadmin/auth"
	"github.com/flyteorg/flyte/flyteadmin/auth/config"
	"github.com/flyteorg/flyte/flyteadmin/auth/interceptors/interceptorstest"
	"github.com/flyteorg/flyte/flyteadmin/auth/interfaces/mocks"
	"github.com/flyteorg/flyte/flyteadmin/auth/isolation"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

func TestGetAuthorizationInterceptor(t *testing.T) {

	t.Run("policy validation fails", func(t *testing.T) {

		cfg := &config.Config{
			Rbac: config.Rbac{
				Policies: []config.AuthorizationPolicy{
					{
						Role: "admin",
						Rules: []config.Rule{
							{
								Name:          "example",
								MethodPattern: ".*",
								Project:       "",
								Domain:        "development",
							},
						},
					},
				},
			},
		}
		authCtx := &mocks.AuthenticationContext{}
		authCtx.OnOptions().Return(cfg)
		_, err := GetAuthorizationInterceptor(authCtx)
		require.ErrorContains(t, err, "authorization policy rule example has invalid resource scope")
	})
}

func TestAuthorizationInterceptor(t *testing.T) {

	err := logger.SetConfig(&logger.Config{Level: logger.DebugLevel})
	require.NoError(t, err)
	ctx := context.Background()

	info := &grpc.UnaryServerInfo{
		FullMethod: "ExampleMethod",
	}

	adminAuthPolicy := config.AuthorizationPolicy{
		Role: "admin",
		Rules: []config.Rule{
			{
				Name:          "example",
				MethodPattern: ".*",
				Project:       "flytesnacks",
				Domain:        "development",
			},
		},
	}

	t.Run("bypass method pattern wildcard match", func(t *testing.T) {

		cfg := &config.Config{
			Rbac: config.Rbac{
				BypassMethodPatterns: []string{".*"},
			},
		}
		authCtx := &mocks.AuthenticationContext{}
		authCtx.OnOptions().Return(cfg)

		interceptor, err := GetAuthorizationInterceptor(authCtx)
		require.NoError(t, err)

		handler := &interceptorstest.TestUnaryHandler{}

		_, err = interceptor(ctx, nil, info, handler.Handle)
		require.NoError(t, err)
		require.Equal(t, 1, handler.GetHandleCallCount())

		isolationCtx := isolation.IsolationContextFromContext(handler.GetCapturedCtx())
		require.Empty(t, isolationCtx.GetResourceScopes())
	})

	t.Run("bypass method pattern exact match", func(t *testing.T) {

		cfg := &config.Config{
			Rbac: config.Rbac{
				BypassMethodPatterns: []string{"ExampleMethod"},
			},
		}
		authCtx := &mocks.AuthenticationContext{}
		authCtx.OnOptions().Return(cfg)

		interceptor, err := GetAuthorizationInterceptor(authCtx)
		require.NoError(t, err)

		handler := &interceptorstest.TestUnaryHandler{}

		_, err = interceptor(ctx, nil, info, handler.Handle)
		require.NoError(t, err)
		require.Equal(t, 1, handler.GetHandleCallCount())

		isolationCtx := isolation.IsolationContextFromContext(handler.GetCapturedCtx())
		require.Empty(t, isolationCtx.GetResourceScopes())
	})

	t.Run("bypass method pattern no match", func(t *testing.T) {

		cfg := &config.Config{
			Rbac: config.Rbac{
				BypassMethodPatterns: []string{"NoMethod"},
			},
		}
		authCtx := &mocks.AuthenticationContext{}
		authCtx.OnOptions().Return(cfg)

		interceptor, err := GetAuthorizationInterceptor(authCtx)
		require.NoError(t, err)

		handler := &interceptorstest.TestUnaryHandler{}

		_, err = interceptor(ctx, nil, info, handler.Handle)
		require.ErrorIs(t, err, status.Errorf(codes.PermissionDenied, ""))
		require.Equal(t, 0, handler.GetHandleCallCount())
	})

	t.Run("authorization fails due to no roles", func(t *testing.T) {

		cfg := &config.Config{
			Rbac: config.Rbac{
				Policies: []config.AuthorizationPolicy{
					adminAuthPolicy,
				},
			},
		}
		authCtx := &mocks.AuthenticationContext{}
		authCtx.OnOptions().Return(cfg)

		interceptor, err := GetAuthorizationInterceptor(authCtx)
		require.NoError(t, err)

		handler := &interceptorstest.TestUnaryHandler{}

		_, err = interceptor(ctx, nil, info, handler.Handle)
		require.ErrorIs(t, err, status.Errorf(codes.PermissionDenied, ""))
		require.Equal(t, 0, handler.GetHandleCallCount())
	})

	t.Run("authorization success with scope based roles resolution", func(t *testing.T) {

		cfg := &config.Config{
			Rbac: config.Rbac{
				Policies: []config.AuthorizationPolicy{
					adminAuthPolicy,
				},
				TokenScopeRoleResolver: config.TokenScopeRoleResolver{
					Enabled: true,
				},
			},
		}
		authCtx := &mocks.AuthenticationContext{}
		authCtx.OnOptions().Return(cfg)

		interceptor, err := GetAuthorizationInterceptor(authCtx)
		require.NoError(t, err)

		handler := &interceptorstest.TestUnaryHandler{}

		scopes := sets.NewString("admin")
		tokenIdentityContext, err := auth.NewIdentityContext("", "", "", time.Now(), scopes, nil, nil)
		ctxWithIdentity := tokenIdentityContext.WithContext(ctx)
		require.NoError(t, err)

		_, err = interceptor(ctxWithIdentity, nil, info, handler.Handle)
		require.NoError(t, err)
		require.Equal(t, 1, handler.GetHandleCallCount())

		isolationCtx := isolation.IsolationContextFromContext(handler.GetCapturedCtx())
		require.Len(t, isolationCtx.GetResourceScopes(), 1)

		resourceScope := isolationCtx.GetResourceScopes()[0]
		expectedResourceScope := isolation.ResourceScope{
			Project: "flytesnacks",
			Domain:  "development",
		}
		require.Equal(t, expectedResourceScope, resourceScope)
	})

	t.Run("authorization fails with scope based roles resolution", func(t *testing.T) {

		cfg := &config.Config{
			Rbac: config.Rbac{
				Policies: []config.AuthorizationPolicy{
					adminAuthPolicy,
				},
				TokenScopeRoleResolver: config.TokenScopeRoleResolver{
					Enabled: true,
				},
			},
		}
		authCtx := &mocks.AuthenticationContext{}
		authCtx.OnOptions().Return(cfg)

		interceptor, err := GetAuthorizationInterceptor(authCtx)
		require.NoError(t, err)

		handler := &interceptorstest.TestUnaryHandler{}

		scopes := sets.NewString("notadmin")
		tokenIdentityContext, err := auth.NewIdentityContext("", "", "", time.Now(), scopes, nil, nil)
		ctxWithIdentity := tokenIdentityContext.WithContext(ctx)
		require.NoError(t, err)

		_, err = interceptor(ctxWithIdentity, nil, info, handler.Handle)
		require.ErrorIs(t, err, status.Errorf(codes.PermissionDenied, ""))
		require.Equal(t, 0, handler.GetHandleCallCount())
	})

	t.Run("authorization success with string claim based roles resolution", func(t *testing.T) {

		cfg := &config.Config{
			Rbac: config.Rbac{
				Policies: []config.AuthorizationPolicy{
					adminAuthPolicy,
				},
				TokenClaimRoleResolver: config.TokenClaimRoleResolver{
					Enabled: true,
					TokenClaims: []config.TokenClaim{
						{
							Name: "group",
						},
					},
				},
			},
		}
		authCtx := &mocks.AuthenticationContext{}
		authCtx.OnOptions().Return(cfg)

		interceptor, err := GetAuthorizationInterceptor(authCtx)
		require.NoError(t, err)

		handler := &interceptorstest.TestUnaryHandler{}

		claims := map[string]interface{}{
			"group": "admin",
		}
		tokenIdentityContext, err := auth.NewIdentityContext("", "", "", time.Now(), nil, nil, claims)
		ctxWithIdentity := tokenIdentityContext.WithContext(ctx)
		require.NoError(t, err)

		_, err = interceptor(ctxWithIdentity, nil, info, handler.Handle)
		require.NoError(t, err)
		require.Equal(t, 1, handler.GetHandleCallCount())

		isolationCtx := isolation.IsolationContextFromContext(handler.GetCapturedCtx())
		require.Len(t, isolationCtx.GetResourceScopes(), 1)

		resourceScope := isolationCtx.GetResourceScopes()[0]
		expectedResourceScope := isolation.ResourceScope{
			Project: "flytesnacks",
			Domain:  "development",
		}
		require.Equal(t, expectedResourceScope, resourceScope)
	})

	t.Run("authorization fails with string claim based roles resolution", func(t *testing.T) {

		cfg := &config.Config{
			Rbac: config.Rbac{
				Policies: []config.AuthorizationPolicy{
					adminAuthPolicy,
				},
				TokenClaimRoleResolver: config.TokenClaimRoleResolver{
					Enabled: true,
					TokenClaims: []config.TokenClaim{
						{
							Name: "group",
						},
					},
				},
			},
		}
		authCtx := &mocks.AuthenticationContext{}
		authCtx.OnOptions().Return(cfg)

		interceptor, err := GetAuthorizationInterceptor(authCtx)
		require.NoError(t, err)

		handler := &interceptorstest.TestUnaryHandler{}

		claims := map[string]interface{}{
			"group": "notadmin",
		}
		tokenIdentityContext, err := auth.NewIdentityContext("", "", "", time.Now(), nil, nil, claims)
		ctxWithIdentity := tokenIdentityContext.WithContext(ctx)
		require.NoError(t, err)

		_, err = interceptor(ctxWithIdentity, nil, info, handler.Handle)
		require.ErrorIs(t, err, status.Errorf(codes.PermissionDenied, ""))
		require.Equal(t, 0, handler.GetHandleCallCount())
	})

	t.Run("authorization success with string list claim based roles resolution", func(t *testing.T) {

		cfg := &config.Config{
			Rbac: config.Rbac{
				Policies: []config.AuthorizationPolicy{
					adminAuthPolicy,
				},
				TokenClaimRoleResolver: config.TokenClaimRoleResolver{
					Enabled: true,
					TokenClaims: []config.TokenClaim{
						{
							Name: "groups",
						},
					},
				},
			},
		}
		authCtx := &mocks.AuthenticationContext{}
		authCtx.OnOptions().Return(cfg)

		interceptor, err := GetAuthorizationInterceptor(authCtx)
		require.NoError(t, err)

		handler := &interceptorstest.TestUnaryHandler{}

		claims := map[string]interface{}{
			"groups": []interface{}{"admin", "notadmin"},
		}
		tokenIdentityContext, err := auth.NewIdentityContext("", "", "", time.Now(), nil, nil, claims)
		ctxWithIdentity := tokenIdentityContext.WithContext(ctx)
		require.NoError(t, err)

		_, err = interceptor(ctxWithIdentity, nil, info, handler.Handle)
		require.NoError(t, err)
		require.Equal(t, 1, handler.GetHandleCallCount())

		isolationCtx := isolation.IsolationContextFromContext(handler.GetCapturedCtx())
		require.Len(t, isolationCtx.GetResourceScopes(), 1)

		resourceScope := isolationCtx.GetResourceScopes()[0]
		expectedResourceScope := isolation.ResourceScope{
			Project: "flytesnacks",
			Domain:  "development",
		}
		require.Equal(t, expectedResourceScope, resourceScope)
	})

	t.Run("authorization fails with string claim based roles resolution", func(t *testing.T) {

		cfg := &config.Config{
			Rbac: config.Rbac{
				Policies: []config.AuthorizationPolicy{
					adminAuthPolicy,
				},
				TokenClaimRoleResolver: config.TokenClaimRoleResolver{
					Enabled: true,
					TokenClaims: []config.TokenClaim{
						{
							Name: "groups",
						},
					},
				},
			},
		}
		authCtx := &mocks.AuthenticationContext{}
		authCtx.OnOptions().Return(cfg)

		interceptor, err := GetAuthorizationInterceptor(authCtx)
		require.NoError(t, err)

		handler := &interceptorstest.TestUnaryHandler{}

		claims := map[string]interface{}{
			"groups": []interface{}{"notadmin"},
		}
		tokenIdentityContext, err := auth.NewIdentityContext("", "", "", time.Now(), nil, nil, claims)
		ctxWithIdentity := tokenIdentityContext.WithContext(ctx)
		require.NoError(t, err)
		_, err = interceptor(ctxWithIdentity, nil, info, handler.Handle)
		require.ErrorIs(t, err, status.Errorf(codes.PermissionDenied, ""))
		require.Equal(t, 0, handler.GetHandleCallCount())
	})
}
