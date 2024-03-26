package authzserver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/sets"
)

func Test_verifyClaims(t *testing.T) {
	t.Run("Empty claims, fail", func(t *testing.T) {
		_, err := verifyClaims(sets.NewString("https://myserver"), map[string]interface{}{})
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
		})

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
			})

		assert.NoError(t, err)
		assert.Equal(t, "https://myserver", identityCtx.Audience())
	})

	t.Run("No matching audience", func(t *testing.T) {
		_, err := verifyClaims(sets.NewString("https://myserver", "https://myserver2"),
			map[string]interface{}{
				"aud": []string{"https://myserver3"},
			})

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
			})

		assert.NoError(t, err)
		assert.Equal(t, "https://myserver", identityCtx.Audience())
	})

	t.Run("String scope", func(t *testing.T) {
		identityCtx, err := verifyClaims(sets.NewString("https://myserver", "https://myserver2"),
			map[string]interface{}{
				"aud": []string{"https://myserver"},
				"scp": "all",
			})

		assert.NoError(t, err)
		assert.Equal(t, "https://myserver", identityCtx.Audience())
		assert.Equal(t, sets.NewString("all"), identityCtx.Scopes())
	})
	t.Run("unknown scope", func(t *testing.T) {
		identityCtx, err := verifyClaims(sets.NewString("https://myserver", "https://myserver2"),
			map[string]interface{}{
				"aud": []string{"https://myserver"},
				"scp": 1,
			})

		assert.Error(t, err)
		assert.Nil(t, identityCtx)
		assert.Equal(t, "failed getting scope claims due to  unknown type int with value 1", err.Error())
	})
}
