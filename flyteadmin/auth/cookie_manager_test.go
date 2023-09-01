package auth

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/flyteorg/flyteadmin/auth/config"
)

func TestCookieManager_SetTokenCookies(t *testing.T) {
	ctx := context.Background()
	// These were generated for unit testing only.
	hashKeyEncoded := "wG4pE1ccdw/pHZ2ml8wrD5VJkOtLPmBpWbKHmezWXktGaFbRoAhXidWs8OpbA3y7N8vyZhz1B1E37+tShWC7gA" //nolint:goconst
	blockKeyEncoded := "afyABVgGOvWJFxVyOvCWCupoTn6BkNl4SOHmahho16Q"                                           //nolint:goconst
	cookieSetting := config.CookieSettings{
		SameSitePolicy: config.SameSiteDefaultMode,
		Domain:         "default",
	}
	manager, err := NewCookieManager(ctx, hashKeyEncoded, blockKeyEncoded, cookieSetting)
	assert.NoError(t, err)
	token := &oauth2.Token{
		AccessToken:  "access",
		RefreshToken: "refresh",
	}
	token = token.WithExtra(map[string]interface{}{
		"id_token": "id token",
	})

	t.Run("invalid_hash_key", func(t *testing.T) {
		_, err := NewCookieManager(ctx, "wrong", blockKeyEncoded, cookieSetting)

		assert.EqualError(t, err, "[BINARY_DECODING_FAILED] Error decoding hash key bytes, caused by: illegal base64 data at input byte 4")
	})

	t.Run("invalid_block_key", func(t *testing.T) {
		_, err := NewCookieManager(ctx, hashKeyEncoded, "wrong", cookieSetting)

		assert.EqualError(t, err, "[BINARY_DECODING_FAILED] Error decoding block key bytes, caused by: illegal base64 data at input byte 4")
	})

	t.Run("set_token_cookies", func(t *testing.T) {
		w := httptest.NewRecorder()

		err = manager.SetTokenCookies(ctx, w, token)

		assert.NoError(t, err)
		fmt.Println(w.Header().Get("Set-Cookie"))
		c := w.Result().Cookies()
		assert.Equal(t, "flyte_at", c[0].Name)
		assert.Equal(t, "flyte_idt", c[1].Name)
		assert.Equal(t, "flyte_rt", c[2].Name)
	})

	t.Run("set_token_cookies_wrong_key", func(t *testing.T) {
		wrongKey := base64.RawStdEncoding.EncodeToString(bytes.Repeat([]byte("X"), 75))
		wrongManager, err := NewCookieManager(ctx, wrongKey, wrongKey, cookieSetting)
		require.NoError(t, err)
		w := httptest.NewRecorder()

		err = wrongManager.SetTokenCookies(ctx, w, token)

		assert.EqualError(t, err, "[SECURE_COOKIE_ERROR] Error creating secure cookie, caused by: securecookie: error - caused by: crypto/aes: invalid key size 75")
	})

	t.Run("retrieve_token_values", func(t *testing.T) {
		w := httptest.NewRecorder()

		err = manager.SetTokenCookies(ctx, w, token)
		assert.NoError(t, err)

		cookies := w.Result().Cookies()
		req, err := http.NewRequest("GET", "/api/v1/projects", nil)
		assert.NoError(t, err)
		for _, c := range cookies {
			req.AddCookie(c)
		}

		idToken, access, refresh, err := manager.RetrieveTokenValues(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, "id token", idToken)
		assert.Equal(t, "access", access)
		assert.Equal(t, "refresh", refresh)
	})

	t.Run("retrieve_token_values_wrong_key", func(t *testing.T) {
		wrongKey := base64.RawStdEncoding.EncodeToString(bytes.Repeat([]byte("X"), 75))
		wrongManager, err := NewCookieManager(ctx, wrongKey, wrongKey, cookieSetting)
		require.NoError(t, err)

		w := httptest.NewRecorder()

		err = manager.SetTokenCookies(ctx, w, token)
		assert.NoError(t, err)

		cookies := w.Result().Cookies()
		req, err := http.NewRequest("GET", "/api/v1/projects", nil)
		assert.NoError(t, err)
		for _, c := range cookies {
			req.AddCookie(c)
		}

		_, _, _, err = wrongManager.RetrieveTokenValues(ctx, req)

		assert.EqualError(t, err, "[EMPTY_OAUTH_TOKEN] Error reading existing secure cookie [flyte_idt]. Error: [SECURE_COOKIE_ERROR] Error reading secure cookie flyte_idt, caused by: securecookie: error - caused by: crypto/aes: invalid key size 75")
	})

	t.Run("logout_access_cookie", func(t *testing.T) {
		cookie := manager.getLogoutAccessCookie()

		assert.True(t, time.Now().After(cookie.Expires))
		assert.Equal(t, cookieSetting.Domain, cookie.Domain)
	})

	t.Run("logout_refresh_cookie", func(t *testing.T) {
		cookie := manager.getLogoutRefreshCookie()

		assert.True(t, time.Now().After(cookie.Expires))
		assert.Equal(t, cookieSetting.Domain, cookie.Domain)
	})

	t.Run("delete_cookies", func(t *testing.T) {
		w := httptest.NewRecorder()

		manager.DeleteCookies(ctx, w)

		cookies := w.Result().Cookies()
		require.Equal(t, 2, len(cookies))
		assert.True(t, time.Now().After(cookies[0].Expires))
		assert.Equal(t, cookieSetting.Domain, cookies[0].Domain)
		assert.True(t, time.Now().After(cookies[1].Expires))
		assert.Equal(t, cookieSetting.Domain, cookies[1].Domain)
	})

	t.Run("get_http_same_site_policy", func(t *testing.T) {
		manager.sameSitePolicy = config.SameSiteLaxMode
		assert.Equal(t, http.SameSiteLaxMode, manager.getHTTPSameSitePolicy())

		manager.sameSitePolicy = config.SameSiteStrictMode
		assert.Equal(t, http.SameSiteStrictMode, manager.getHTTPSameSitePolicy())

		manager.sameSitePolicy = config.SameSiteNoneMode
		assert.Equal(t, http.SameSiteNoneMode, manager.getHTTPSameSitePolicy())
	})
}
