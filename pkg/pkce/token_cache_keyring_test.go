package pkce

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"
)

func TestSaveAndGetToken(t *testing.T) {
	keyring.MockInit()
	tokenCacheProvider := TokenCacheKeyringProvider{
		ServiceUser: "testServiceUser",
		ServiceName: "testServiceName",
	}

	t.Run("Valid Save/Get Token", func(t *testing.T) {
		plan, _ := ioutil.ReadFile("testdata/token.json")
		var tokenData oauth2.Token
		err := json.Unmarshal(plan, &tokenData)
		assert.NoError(t, err)
		err = tokenCacheProvider.SaveToken(&tokenData)
		assert.NoError(t, err)
		var savedToken *oauth2.Token
		savedToken, err = tokenCacheProvider.GetToken()
		assert.NoError(t, err)
		assert.NotNil(t, savedToken)
		assert.Equal(t, tokenData.AccessToken, savedToken.AccessToken)
		assert.Equal(t, tokenData.TokenType, savedToken.TokenType)
		assert.Equal(t, tokenData.Expiry, savedToken.Expiry)
	})

	t.Run("Empty access token Save", func(t *testing.T) {
		plan, _ := ioutil.ReadFile("testdata/empty_access_token.json")
		var tokenData oauth2.Token
		var err error
		err = json.Unmarshal(plan, &tokenData)
		assert.NoError(t, err)

		err = tokenCacheProvider.SaveToken(&tokenData)
		assert.Error(t, err)
	})

	t.Run("Different service name", func(t *testing.T) {
		plan, _ := ioutil.ReadFile("testdata/token.json")
		var tokenData oauth2.Token
		err := json.Unmarshal(plan, &tokenData)
		assert.NoError(t, err)
		err = tokenCacheProvider.SaveToken(&tokenData)
		assert.NoError(t, err)
		tokenCacheProvider2 := TokenCacheKeyringProvider{
			ServiceUser: "testServiceUser2",
			ServiceName: "testServiceName2",
		}

		var savedToken *oauth2.Token
		savedToken, err = tokenCacheProvider2.GetToken()
		assert.Error(t, err)
		assert.Nil(t, savedToken)
	})
}
