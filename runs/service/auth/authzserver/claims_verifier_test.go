package authzserver

import (
	"testing"
	"time"

	jwtgo "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifyClaims_MatchingAudience(t *testing.T) {
	claims := jwtgo.MapClaims{
		"aud": []interface{}{"https://flyte.example.com"},
		"sub": "user123",
		"iat": float64(time.Now().Unix()),
	}

	allowed := map[string]bool{"https://flyte.example.com": true}
	identity, err := verifyClaims(allowed, claims)
	require.NoError(t, err)
	assert.Equal(t, "https://flyte.example.com", identity.Audience())
	assert.Equal(t, "user123", identity.UserID())
}

func TestVerifyClaims_NoMatchingAudience(t *testing.T) {
	claims := jwtgo.MapClaims{
		"aud": []interface{}{"https://other.example.com"},
		"sub": "user123",
	}

	allowed := map[string]bool{"https://flyte.example.com": true}
	_, err := verifyClaims(allowed, claims)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid audience")
}

func TestVerifyClaims_WithClientID(t *testing.T) {
	claims := jwtgo.MapClaims{
		"aud":       "https://flyte.example.com",
		"sub":       "user123",
		"client_id": "my-client",
		"scp":       []interface{}{"read", "write"},
	}

	allowed := map[string]bool{"https://flyte.example.com": true}
	identity, err := verifyClaims(allowed, claims)
	require.NoError(t, err)
	assert.Equal(t, "my-client", identity.AppID())
	assert.Equal(t, []string{"read", "write"}, identity.Scopes())
}

func TestVerifyClaims_UserOnlyDefaultsToScopeAll(t *testing.T) {
	claims := jwtgo.MapClaims{
		"aud": "https://flyte.example.com",
		"sub": "user123",
	}

	allowed := map[string]bool{"https://flyte.example.com": true}
	identity, err := verifyClaims(allowed, claims)
	require.NoError(t, err)
	assert.Equal(t, []string{"all"}, identity.Scopes())
	assert.Equal(t, "", identity.AppID())
}

func TestVerifyClaims_WithEmail(t *testing.T) {
	claims := jwtgo.MapClaims{
		"aud":   "https://flyte.example.com",
		"sub":   "user123",
		"email": "user@example.com",
	}

	allowed := map[string]bool{"https://flyte.example.com": true}
	identity, err := verifyClaims(allowed, claims)
	require.NoError(t, err)
	assert.Equal(t, "user@example.com", identity.UserInfo().Email)
}

func TestVerifyClaims_WithUserInfoClaim(t *testing.T) {
	claims := jwtgo.MapClaims{
		"aud": "https://flyte.example.com",
		"sub": "user123",
		"user_info": map[string]interface{}{
			"name":  "Test User",
			"email": "test@example.com",
		},
	}

	allowed := map[string]bool{"https://flyte.example.com": true}
	identity, err := verifyClaims(allowed, claims)
	require.NoError(t, err)
	assert.Equal(t, "Test User", identity.UserInfo().Name)
	assert.Equal(t, "test@example.com", identity.UserInfo().Email)
}

func TestVerifyClaims_ScopeAsString(t *testing.T) {
	claims := jwtgo.MapClaims{
		"aud":       "https://flyte.example.com",
		"sub":       "user123",
		"client_id": "my-client",
		"scp":       "read",
	}

	allowed := map[string]bool{"https://flyte.example.com": true}
	identity, err := verifyClaims(allowed, claims)
	require.NoError(t, err)
	assert.Equal(t, []string{"read"}, identity.Scopes())
}

func TestInterfaceSliceToStringSlice(t *testing.T) {
	input := []interface{}{"a", "b", "c"}
	result := interfaceSliceToStringSlice(input)
	assert.Equal(t, []string{"a", "b", "c"}, result)
}
