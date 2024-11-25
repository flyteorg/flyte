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

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/flyteorg/flyte/flyteadmin/auth/config"
	serverConfig "github.com/flyteorg/flyte/flyteadmin/pkg/config"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
)

func TestCookieManager(t *testing.T) {
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
		assert.Equal(t, "flyte_at_1", c[0].Name)
		assert.Equal(t, "flyte_at_2", c[1].Name)
		assert.Equal(t, "flyte_idt", c[2].Name)
		assert.Equal(t, "flyte_rt", c[3].Name)
	})

	t.Run("set_token_nil", func(t *testing.T) {
		w := httptest.NewRecorder()

		err = manager.SetTokenCookies(ctx, w, nil)

		assert.EqualError(t, err, "[EMPTY_OAUTH_TOKEN] Attempting to set cookies with nil token")
	})

	t.Run("set_long_token_cookies", func(t *testing.T) {
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
		longString := "asdfkjashdfkljqwpeuqwoieuiposdasdfasdfsdfcuzxcvjzvjlasfuo9qwuoiqwueoqoieulkszcjvhlkshcvlkasdhflkashfqoiwaskldfjhasdfljk" +
			"aklsdjflkasdfhlkjasdhfklhfjkasdhfkasdhfjasdfhasldkfjhaskldfhaklsdfhlasjdfhalksjdfhlskdqoiweuqioweuqioweuqoiew" +
			"aklsdjfqwoieuioqwerupweiruoqpweurpqweuropqweurpqweurpoqwuetopqweuropqwuerpoqwuerpoqweuropqweurpoqweyitoqpwety"
		// These were generated for unit testing only.
		tokenData := jwt.MapClaims{
			"iss":                "https://login.microsoftonline.com/72f988bf-86f1-41af-91ab-2d7cd011db47/v2.0",
			"aio":                "AXQAi/8UAAAAvfR1B135YmjOZGtNGTM/fgvUY0ugwwk2NWCjmNglWR9v+b5sI3cCSGXOp1Zw96qUNb1dm0jqBHHYDKQtc4UplZAtFULbitt3x2KYdigeS5tXl0yNIeMiYsA/1Dpd43xg9sXAtLU3+iZetXiDasdfkpsaldg==",
			"sub":                "subject",
			"aud":                "audience",
			"exp":                1677642301,
			"nbf":                1677641001,
			"iat":                1677641001,
			"jti":                "a-unique-identifier",
			"user_id":            "123456",
			"username":           "john_doe",
			"preferred_username": "john_doe",
			"given_name":         "John",
			"family_name":        "Doe",
			"email":              "john.doe@example.com",
			"scope":              "read write",
			"role":               "user",
			"is_active":          true,
			"is_verified":        true,
			"client_id":          "client123",
			"custom_field1":      longString,
			"custom_field2":      longString,
			"custom_field3":      longString,
			"custom_field4":      longString,
			"custom_field5":      longString,
			"custom_field6":      []string{longString, longString},
			"additional_field1":  "extra_value1",
			"additional_field2":  "extra_value2",
		}
		secretKey := []byte("tJL6Wr2JUsxLyNezPQh1J6zn6wSoDAhgRYSDkaMuEHy75VikiB8wg25WuR96gdMpookdlRvh7SnRvtjQN9b5m4zJCMpSRcJ5DuXl4mcd7Cg3Zp1C5")
		rawToken := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenData)
		tokenString, err := rawToken.SignedString(secretKey)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		token := &oauth2.Token{
			AccessToken:  tokenString,
			RefreshToken: "refresh",
		}

		token = token.WithExtra(map[string]interface{}{
			"id_token": "id token",
		})

		w := httptest.NewRecorder()
		_, err = http.NewRequest("GET", "/api/v1/projects", nil)
		assert.NoError(t, err)
		err = manager.SetTokenCookies(ctx, w, token)
		assert.NoError(t, err)
		fmt.Println(w.Header().Get("Set-Cookie"))
		c := w.Result().Cookies()
		assert.Equal(t, "flyte_at_1", c[0].Name)
		assert.Equal(t, "flyte_at_2", c[1].Name)
		assert.Equal(t, "flyte_idt", c[2].Name)
		assert.Equal(t, "flyte_rt", c[3].Name)
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

	tests := []struct {
		name                 string
		insecureCookieHeader bool
		expectedSecure       bool
	}{
		{
			name:                 "secure_cookies",
			insecureCookieHeader: false,
			expectedSecure:       true,
		},
		{
			name:                 "insecure_cookies",
			insecureCookieHeader: true,
			expectedSecure:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			serverConfig.SetConfig(&serverConfig.ServerConfig{
				Security: serverConfig.ServerSecurityOptions{
					InsecureCookieHeader: tt.insecureCookieHeader,
				},
			})

			manager.DeleteCookies(ctx, w)

			cookies := w.Result().Cookies()
			require.Equal(t, 5, len(cookies))

			// Check secure flag for each cookie
			for _, cookie := range cookies {
				assert.Equal(t, tt.expectedSecure, cookie.Secure)
				assert.True(t, time.Now().After(cookie.Expires))
				assert.Equal(t, cookieSetting.Domain, cookie.Domain)
			}

			// Check cookie names
			assert.Equal(t, accessTokenCookieName, cookies[0].Name)
			assert.Equal(t, accessTokenCookieNameSplitFirst, cookies[1].Name)
			assert.Equal(t, accessTokenCookieNameSplitSecond, cookies[2].Name)
			assert.Equal(t, refreshTokenCookieName, cookies[3].Name)
			assert.Equal(t, idTokenCookieName, cookies[4].Name)
		})
	}

	t.Run("get_http_same_site_policy", func(t *testing.T) {
		manager.sameSitePolicy = config.SameSiteLaxMode
		assert.Equal(t, http.SameSiteLaxMode, manager.getHTTPSameSitePolicy())

		manager.sameSitePolicy = config.SameSiteStrictMode
		assert.Equal(t, http.SameSiteStrictMode, manager.getHTTPSameSitePolicy())

		manager.sameSitePolicy = config.SameSiteNoneMode
		assert.Equal(t, http.SameSiteNoneMode, manager.getHTTPSameSitePolicy())
	})

	t.Run("set_user_info", func(t *testing.T) {
		w := httptest.NewRecorder()
		info := &service.UserInfoResponse{
			Subject: "sub",
			Name:    "foo",
		}

		err := manager.SetUserInfoCookie(ctx, w, info)

		assert.NoError(t, err)
		cookies := w.Result().Cookies()
		require.Len(t, cookies, 1)
		assert.Equal(t, "flyte_user_info", cookies[0].Name)
	})

	t.Run("set_auth_code", func(t *testing.T) {
		w := httptest.NewRecorder()

		err := manager.SetAuthCodeCookie(ctx, w, "foo.com")

		assert.NoError(t, err)
		cookies := w.Result().Cookies()
		require.Len(t, cookies, 1)
		assert.Equal(t, "flyte_auth_code", cookies[0].Name)
	})
}
