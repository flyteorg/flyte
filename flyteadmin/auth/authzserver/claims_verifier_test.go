package authzserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/sets"
)

func Test_verifyClaims(t *testing.T) {
	t.Run("Empty claims, fail", func(t *testing.T) {
		_, err := verifyClaims(sets.NewString("https://myserver"), map[string]interface{}{}, nil, nil)
		assert.Error(t, err)
	})

	t.Run("All filled", func(t *testing.T) {
		identityCtx, err := verifyClaims(sets.NewString("https://myserver"), map[string]interface{}{
			"aud": []string{"https://myserver"},
			"user_info": map[string]interface{}{
				"preferred_name": "John Doe",
			},
			"sub":       "123",
			"client_id": "my-client",
			"scp":       []interface{}{"all", "offline"},
			"email":     "byhsu@linkedin.com",
		}, nil, nil)

		assert.NoError(t, err)
		assert.Equal(t, sets.NewString("all", "offline"), identityCtx.Scopes())
		assert.Equal(t, "my-client", identityCtx.AppID())
		assert.Equal(t, "123", identityCtx.UserID())
		assert.Equal(t, "https://myserver", identityCtx.Audience())
		assert.Equal(t, "byhsu@linkedin.com", identityCtx.UserInfo().Email)
	})

	t.Run("Multiple audience", func(t *testing.T) {
		identityCtx, err := verifyClaims(sets.NewString("https://myserver", "https://myserver2"),
			map[string]interface{}{
				"aud": []string{"https://myserver"},
				"user_info": map[string]interface{}{
					"preferred_name": "John Doe",
				},
				"sub":       "123",
				"client_id": "my-client",
				"scp":       []interface{}{"all", "offline"},
			}, nil, nil)

		assert.NoError(t, err)
		assert.Equal(t, "https://myserver", identityCtx.Audience())
	})

	t.Run("No matching audience", func(t *testing.T) {
		_, err := verifyClaims(sets.NewString("https://myserver", "https://myserver2"),
			map[string]interface{}{
				"aud": []string{"https://myserver3"},
			}, nil, nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid audience")
	})

	t.Run("Use first matching audience", func(t *testing.T) {
		identityCtx, err := verifyClaims(sets.NewString("https://myserver", "https://myserver2", "https://myserver3"),
			map[string]interface{}{
				"aud": []string{"https://myserver", "https://myserver2"},
				"user_info": map[string]interface{}{
					"preferred_name": "John Doe",
				},
				"sub":       "123",
				"client_id": "my-client",
				"scp":       []interface{}{"all", "offline"},
			}, nil, nil)

		assert.NoError(t, err)
		assert.Equal(t, "https://myserver", identityCtx.Audience())
	})

	t.Run("String scope", func(t *testing.T) {
		identityCtx, err := verifyClaims(sets.NewString("https://myserver", "https://myserver2"),
			map[string]interface{}{
				"aud": []string{"https://myserver"},
				"scp": "all",
			}, nil, nil)

		assert.NoError(t, err)
		assert.Equal(t, "https://myserver", identityCtx.Audience())
		assert.Equal(t, sets.NewString("all"), identityCtx.Scopes())
	})

	t.Run("unknown scope", func(t *testing.T) {
		identityCtx, err := verifyClaims(sets.NewString("https://myserver", "https://myserver2"),
			map[string]interface{}{
				"aud": []string{"https://myserver"},
				"scp": 1,
			}, nil, nil)

		assert.Error(t, err)
		assert.Nil(t, identityCtx)
		assert.Equal(t, "failed getting scope claims due to  unknown type int with value 1", err.Error())
	})
}

func Test_verifyClaims_IdentityTypeClaimsForApps(t *testing.T) {
	entraConfig := map[string][]string{
		"idtyp":        {"app"},
		"identitytype": {"app"},
	}

	t.Run("No config, no normalization", func(t *testing.T) {
		// Without identityTypeClaimsForApps, existing identitytype claim passes through unchanged.
		identityCtx, err := verifyClaims(sets.NewString("https://myserver"), map[string]interface{}{
			"aud":          []string{"https://myserver"},
			"sub":          "user-123",
			"identitytype": "user",
		}, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, "user", identityCtx.Claims()["identitytype"])
	})

	t.Run("Configured: identitytype=app (Okta path)", func(t *testing.T) {
		identityCtx, err := verifyClaims(sets.NewString("https://myserver"), map[string]interface{}{
			"aud":          []string{"https://myserver"},
			"sub":          "user-123",
			"identitytype": "app",
		}, nil, entraConfig)
		assert.NoError(t, err)
		assert.Equal(t, "app", identityCtx.Claims()["identitytype"])
	})

	t.Run("Configured: idtyp=app, identitytype absent (Entra app token)", func(t *testing.T) {
		identityCtx, err := verifyClaims(sets.NewString("https://myserver"), map[string]interface{}{
			"aud":   []string{"https://myserver"},
			"sub":   "app-client-id",
			"idtyp": "app",
		}, nil, entraConfig)
		assert.NoError(t, err)
		assert.Equal(t, "app", identityCtx.Claims()["identitytype"])
	})

	t.Run("Configured: no identity type claims present (Entra user token — defaults to user)", func(t *testing.T) {
		identityCtx, err := verifyClaims(sets.NewString("https://myserver"), map[string]interface{}{
			"aud": []string{"https://myserver"},
			"sub": "user-abc",
		}, nil, entraConfig)
		assert.NoError(t, err)
		assert.Equal(t, "user", identityCtx.Claims()["identitytype"])
	})

	t.Run("Configured: idtyp present but value not in app set — normalized to user", func(t *testing.T) {
		identityCtx, err := verifyClaims(sets.NewString("https://myserver"), map[string]interface{}{
			"aud":   []string{"https://myserver"},
			"sub":   "user-123",
			"idtyp": "unexpected_value",
		}, nil, entraConfig)
		assert.NoError(t, err)
		assert.Equal(t, "user", identityCtx.Claims()["identitytype"])
	})

	t.Run("Configured: multiple app values, second value matches", func(t *testing.T) {
		identityCtx, err := verifyClaims(sets.NewString("https://myserver"), map[string]interface{}{
			"aud":        []string{"https://myserver"},
			"sub":        "svc-account",
			"token_type": "service_account",
		}, nil, map[string][]string{
			"token_type": {"app", "service_account", "daemon"},
		})
		assert.NoError(t, err)
		assert.Equal(t, "app", identityCtx.Claims()["identitytype"])
	})
}

func Test_verifyClaims_SubjectClaimNames(t *testing.T) {
	t.Run("No config, defaults to sub", func(t *testing.T) {
		identityCtx, err := verifyClaims(sets.NewString("https://myserver"), map[string]interface{}{
			"aud": []string{"https://myserver"},
			"sub": "user-123",
			"azp": "should-not-use",
		}, nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, "user-123", identityCtx.UserID())
	})

	t.Run("Configured: sub first, uses sub when present", func(t *testing.T) {
		identityCtx, err := verifyClaims(sets.NewString("https://myserver"), map[string]interface{}{
			"aud": []string{"https://myserver"},
			"sub": "user-123",
			"azp": "should-not-use",
		}, []string{"sub", "azp"}, nil)
		assert.NoError(t, err)
		assert.Equal(t, "user-123", identityCtx.UserID())
	})

	t.Run("Configured: sub empty, falls back to azp", func(t *testing.T) {
		claimsRaw := map[string]interface{}{
			"aud": []string{"https://myserver"},
			"sub": "",
			"azp": "app-client-id",
		}
		identityCtx, err := verifyClaims(sets.NewString("https://myserver"), claimsRaw, []string{"sub", "azp"}, nil)
		assert.NoError(t, err)
		assert.Equal(t, "app-client-id", identityCtx.UserID())
		assert.Equal(t, "app-client-id", identityCtx.Claims()["sub"])
		assert.Equal(t, "app-client-id", identityCtx.UserInfo().Subject)
	})

	t.Run("Configured: sub missing, falls back to client_id", func(t *testing.T) {
		claimsRaw := map[string]interface{}{
			"aud":       []string{"https://myserver"},
			"client_id": "my-client",
		}
		identityCtx, err := verifyClaims(sets.NewString("https://myserver"), claimsRaw, []string{"sub", "azp", "client_id"}, nil)
		assert.NoError(t, err)
		assert.Equal(t, "my-client", identityCtx.UserID())
		assert.Equal(t, "my-client", identityCtx.Claims()["sub"])
	})

	t.Run("Configured: sub null, falls back to azp", func(t *testing.T) {
		claimsRaw := map[string]interface{}{
			"aud": []string{"https://myserver"},
			"sub": nil,
			"azp": "app-client-id",
		}
		identityCtx, err := verifyClaims(sets.NewString("https://myserver"), claimsRaw, []string{"sub", "azp"}, nil)
		assert.NoError(t, err)
		assert.Equal(t, "app-client-id", identityCtx.UserID())
	})

	t.Run("Configured: order respected, azp before client_id", func(t *testing.T) {
		identityCtx, err := verifyClaims(sets.NewString("https://myserver"), map[string]interface{}{
			"aud":       []string{"https://myserver"},
			"sub":       "",
			"azp":       "azp-value",
			"client_id": "client-id-value",
		}, []string{"sub", "azp", "client_id"}, nil)
		assert.NoError(t, err)
		assert.Equal(t, "azp-value", identityCtx.UserID())
	})

	t.Run("Configured: no claims match, subject empty", func(t *testing.T) {
		identityCtx, err := verifyClaims(sets.NewString("https://myserver"), map[string]interface{}{
			"aud": []string{"https://myserver"},
		}, []string{"sub", "azp"}, nil)
		assert.NoError(t, err)
		assert.Equal(t, "", identityCtx.UserID())
	})

	t.Run("Configured: azp-only (no sub in list)", func(t *testing.T) {
		identityCtx, err := verifyClaims(sets.NewString("https://myserver"), map[string]interface{}{
			"aud": []string{"https://myserver"},
			"sub": "ignored-sub",
			"azp": "azp-value",
		}, []string{"azp"}, nil)
		assert.NoError(t, err)
		assert.Equal(t, "azp-value", identityCtx.UserID())
		assert.Equal(t, "azp-value", identityCtx.Claims()["sub"])
	})
}
