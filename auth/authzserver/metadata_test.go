package authzserver

import (
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flyteorg/flyteadmin/auth/interfaces/mocks"
	"github.com/lestrrat-go/jwx/jwk"

	"github.com/flyteorg/flyteadmin/auth"

	"github.com/stretchr/testify/assert"
)

func Test_newJSONWebKeySet(t *testing.T) {
	publicKeys := make([]rsa.PublicKey, 0, 5)
	for i := 0; i < cap(publicKeys); i++ {
		secrets, err := auth.NewSecrets()
		assert.NoError(t, err)
		publicKeys = append(publicKeys, secrets.TokenSigningRSAPrivateKey.PublicKey)
	}

	keySet, err := newJSONWebKeySet(publicKeys)
	assert.NoError(t, err)
	for i := 0; i < cap(publicKeys); i++ {
		k, found := keySet.Get(i)
		assert.True(t, found)

		actualPublicKey := &rsa.PublicKey{}
		err = k.Raw(actualPublicKey)
		assert.NoError(t, err)
		assert.Equal(t, &publicKeys[i], actualPublicKey)
	}

	_, found := keySet.Get(cap(publicKeys))
	assert.False(t, found)
}

func TestGetJSONWebKeysEndpoint(t *testing.T) {
	t.Run("Empty keyset", func(t *testing.T) {
		authCtx := &mocks.AuthenticationContext{}
		oauth2Provider := &mocks.OAuth2Provider{}
		authCtx.OnOAuth2Provider().Return(oauth2Provider)
		oauth2Provider.OnKeySet().Return(jwk.NewSet())

		handler := GetJSONWebKeysEndpoint(authCtx)
		responseRecorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		handler.ServeHTTP(responseRecorder, request)
		assert.Equal(t, "{\"keys\":[]}", responseRecorder.Body.String())
	})

	t.Run("2 keys", func(t *testing.T) {
		authCtx := &mocks.AuthenticationContext{}
		oauth2Provider := &mocks.OAuth2Provider{}
		authCtx.OnOAuth2Provider().Return(oauth2Provider)

		publicKeys := make([]rsa.PublicKey, 0, 5)
		for i := 0; i < cap(publicKeys); i++ {
			secrets, err := auth.NewSecrets()
			assert.NoError(t, err)
			publicKeys = append(publicKeys, secrets.TokenSigningRSAPrivateKey.PublicKey)
		}

		keySet, err := newJSONWebKeySet(publicKeys)
		assert.NoError(t, err)

		oauth2Provider.OnKeySet().Return(keySet)

		handler := GetJSONWebKeysEndpoint(authCtx)
		responseRecorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		handler.ServeHTTP(responseRecorder, request)

		actualKeySet := jwk.NewSet()
		assert.NoError(t, json.Unmarshal(responseRecorder.Body.Bytes(), &actualKeySet))

		for i := 0; i < keySet.Len(); i++ {
			expectedKeyRaw, _ := keySet.Get(i)
			expectedKey := &rsa.PublicKey{}
			assert.NoError(t, expectedKeyRaw.Raw(expectedKey))
		}

		assert.Equal(t, keySet, actualKeySet)
	})
}
