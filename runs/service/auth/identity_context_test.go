package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/auth"
)

func TestNewIdentityContext(t *testing.T) {
	userInfo := &authpb.UserInfoResponse{
		Name:  "Test User",
		Email: "test@example.com",
	}
	now := time.Now()

	ic := NewIdentityContext("aud", "user1", "app1", now, []string{"read"}, userInfo, nil)

	assert.Equal(t, "aud", ic.Audience())
	assert.Equal(t, "user1", ic.UserID())
	assert.Equal(t, "app1", ic.AppID())
	assert.Equal(t, now, ic.AuthenticatedAt())
	assert.Equal(t, []string{"read"}, ic.Scopes())
	assert.Equal(t, "Test User", ic.UserInfo().Name)
	// Subject should be filled from userID when empty
	assert.Equal(t, "user1", ic.UserInfo().Subject)
}

func TestNewIdentityContext_PreservesSubject(t *testing.T) {
	userInfo := &authpb.UserInfoResponse{
		Subject: "existing-sub",
	}

	ic := NewIdentityContext("aud", "user1", "", time.Time{}, nil, userInfo, nil)
	assert.Equal(t, "existing-sub", ic.UserInfo().Subject)
}

func TestNewIdentityContext_NilUserInfo(t *testing.T) {
	ic := NewIdentityContext("aud", "user1", "", time.Time{}, nil, nil, nil)
	require.NotNil(t, ic.UserInfo())
	assert.Equal(t, "user1", ic.UserInfo().Subject)
}

func TestIdentityContext_IsEmpty(t *testing.T) {
	assert.True(t, (*IdentityContext)(nil).IsEmpty())
	assert.True(t, (&IdentityContext{}).IsEmpty())
	assert.False(t, (&IdentityContext{userID: "u"}).IsEmpty())
}

func TestIdentityContext_WithContext(t *testing.T) {
	ic := NewIdentityContext("aud", "user1", "app1", time.Now(), nil, nil, nil)
	ctx := ic.WithContext(context.Background())

	retrieved := IdentityContextFromContext(ctx)
	require.NotNil(t, retrieved)
	assert.Equal(t, "user1", retrieved.UserID())
}

func TestIdentityContextFromContext_Empty(t *testing.T) {
	assert.Nil(t, IdentityContextFromContext(context.Background()))
}
