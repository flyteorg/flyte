package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/util/sets"
)

func TestBlanketAuthorization(t *testing.T) {
	t.Run("authenticated and authorized", func(t *testing.T) {
		allScopes := sets.NewString(ScopeAll)
		identityCtx := IdentityContext{
			audience: "aud",
			userID:   "uid",
			appID:    "appid",
			scopes:   &allScopes,
		}
		handlerCalled := false
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			handlerCalled = true
			return nil, nil
		}
		ctx := context.WithValue(context.TODO(), ContextKeyIdentityContext, identityCtx)
		_, err := BlanketAuthorization(ctx, nil, nil, handler)
		assert.NoError(t, err)
		assert.True(t, handlerCalled)
	})
	t.Run("unauthenticated", func(t *testing.T) {
		handlerCalled := false
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			handlerCalled = true
			return nil, nil
		}
		ctx := context.TODO()
		_, err := BlanketAuthorization(ctx, nil, nil, handler)
		assert.NoError(t, err)
		assert.True(t, handlerCalled)
	})
	t.Run("authenticated and not authorized", func(t *testing.T) {
		identityCtx := IdentityContext{
			audience: "aud",
			userID:   "uid",
			appID:    "appid",
		}
		handlerCalled := false
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			handlerCalled = true
			return nil, nil
		}
		ctx := context.WithValue(context.TODO(), ContextKeyIdentityContext, identityCtx)
		_, err := BlanketAuthorization(ctx, nil, nil, handler)
		asStatus, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, asStatus.Code(), codes.Unauthenticated)
		assert.False(t, handlerCalled)
	})
}

func TestGetUserIdentityFromContext(t *testing.T) {
	identityContext := IdentityContext{
		userID: "yeee",
	}

	ctx := identityContext.WithContext(context.Background())

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		identityContext := IdentityContextFromContext(ctx)
		euid := identityContext.ExecutionIdentity()
		assert.Equal(t, euid, "yeee")
		return nil, nil
	}

	_, err := ExecutionUserIdentifierInterceptor(ctx, nil, nil, handler)
	assert.NoError(t, err)
}
