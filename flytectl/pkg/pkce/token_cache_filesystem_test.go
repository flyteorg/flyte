package pkce

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestSaveAndGetTokenFilesystem(t *testing.T) {
	setup := func(t *testing.T) *tokenCacheFilesystemProvider {
		t.Helper()
		// Everything inside the directory is automatically cleaned up by the test runner.
		dir := t.TempDir()
		tokenCacheProvider := &tokenCacheFilesystemProvider{
			ServiceUser:     "testServiceUser",
			credentialsFile: filepath.Join(dir, "credentials.json"),
		}
		return tokenCacheProvider
	}

	t.Run("Valid Save/Get Token", func(t *testing.T) {
		tokenCacheProvider := setup(t)

		plan, err := os.ReadFile("testdata/token.json")
		require.NoError(t, err)

		var tokenData oauth2.Token
		err = json.Unmarshal(plan, &tokenData)
		require.NoError(t, err)

		err = tokenCacheProvider.SaveToken(&tokenData)
		require.NoError(t, err)

		var savedToken *oauth2.Token
		savedToken, err = tokenCacheProvider.GetToken()
		require.NoError(t, err)

		assert.NotNil(t, savedToken)
		assert.Equal(t, tokenData.AccessToken, savedToken.AccessToken)
		assert.Equal(t, tokenData.TokenType, savedToken.TokenType)
		assert.Equal(t, tokenData.Expiry, savedToken.Expiry)
	})

	t.Run("Empty access token Save", func(t *testing.T) {
		tokenCacheProvider := setup(t)

		plan, err := os.ReadFile("testdata/empty_access_token.json")
		require.NoError(t, err)

		var tokenData oauth2.Token
		err = json.Unmarshal(plan, &tokenData)
		require.NoError(t, err)

		err = tokenCacheProvider.SaveToken(&tokenData)
		assert.Error(t, err)
	})

	t.Run("Different service name", func(t *testing.T) {
		tokenCacheProvider := setup(t)

		plan, err := os.ReadFile("testdata/token.json")
		require.NoError(t, err)

		var tokenData oauth2.Token
		err = json.Unmarshal(plan, &tokenData)
		require.NoError(t, err)

		err = tokenCacheProvider.SaveToken(&tokenData)
		require.NoError(t, err)

		tokenCacheProvider2 := setup(t)

		var savedToken *oauth2.Token
		savedToken, err = tokenCacheProvider2.GetToken()
		assert.Error(t, err)
		assert.Nil(t, savedToken)

		err = tokenCacheProvider2.SaveToken(&tokenData)
		require.NoError(t, err)

		// new token exists
		savedToken, err = tokenCacheProvider2.GetToken()
		require.NoError(t, err)
		assert.NotNil(t, savedToken)

		// token for different service name still exists
		savedToken, err = tokenCacheProvider.GetToken()
		require.NoError(t, err)
		assert.NotNil(t, savedToken)
	})
}
