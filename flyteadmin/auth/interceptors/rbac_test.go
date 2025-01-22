package interceptors

import (
	"context"
	"github.com/flyteorg/flyte/flyteadmin/auth"
	"github.com/flyteorg/flyte/flyteadmin/auth/config"
	"github.com/flyteorg/flyte/flyteadmin/auth/interfaces/mocks"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/util/sets"
	"testing"
	"time"
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

	logger.SetConfig(&logger.Config{Level: logger.DebugLevel})
	ctx := context.Background()

	info := &grpc.UnaryServerInfo{
		FullMethod: "ExampleMethod",
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

		handler := func(ctx context.Context, req any) (any, error) {
			return nil, nil
		}

		_, err = interceptor(ctx, nil, info, handler)
		require.NoError(t, err)
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

		handler := func(ctx context.Context, req any) (any, error) {
			return nil, nil
		}

		_, err = interceptor(ctx, nil, info, handler)
		require.NoError(t, err)
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

		// FIXME: use mocks and validate they aren't called
		handler := func(ctx context.Context, req any) (any, error) {
			return nil, nil
		}

		_, err = interceptor(ctx, nil, info, handler)
		require.ErrorIs(t, err, status.Errorf(codes.PermissionDenied, ""))
	})

	t.Run("authorization fails due to no roles", func(t *testing.T) {

		cfg := &config.Config{
			Rbac: config.Rbac{
				Policies: []config.AuthorizationPolicy{
					{
						Role: "admin",
						Rules: []config.Rule{
							{
								Name:          "example",
								MethodPattern: ".*",
								Project:       "flytesnacks",
								Domain:        "development",
							},
						},
					},
				},
			},
		}
		authCtx := &mocks.AuthenticationContext{}
		authCtx.OnOptions().Return(cfg)

		interceptor, err := GetAuthorizationInterceptor(authCtx)
		require.NoError(t, err)

		handler := func(ctx context.Context, req any) (any, error) {
			return nil, nil
		}

		_, err = interceptor(ctx, nil, info, handler)
		require.ErrorIs(t, err, status.Errorf(codes.PermissionDenied, ""))
	})

	t.Run("authorization success with scope based roles resolution", func(t *testing.T) {

		cfg := &config.Config{
			Rbac: config.Rbac{
				Policies: []config.AuthorizationPolicy{
					{
						Role: "admin",
						Rules: []config.Rule{
							{
								Name:          "example",
								MethodPattern: ".*",
								Project:       "flytesnacks",
								Domain:        "development",
							},
						},
					},
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

		handler := func(ctx context.Context, req any) (any, error) {
			return nil, nil
		}

		scopes := sets.NewString("admin")
		tokenIdentityContext, err := auth.NewIdentityContext("", "", "", time.Now(), scopes, nil, nil)
		ctxWithIdentity := tokenIdentityContext.WithContext(ctx)
		require.NoError(t, err)
		_, err = interceptor(ctxWithIdentity, nil, info, handler)
		require.NoError(t, err)
	})
}
