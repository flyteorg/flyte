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

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/securecookie"
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

func TestExtractAccessTokenFromCookies(t *testing.T) {
	hashKeyFromSecret, err := base64.StdEncoding.DecodeString("ODg0K2EybG1IaHJHRUNUcUNsRDB2U3BhRzJIbUREWU1FeXUyYU9maTZ0RnJnYy83bVEzRC9rdTU1ZzRrZ3E3MlFiQ2E1ZmduK2NtTkw1Y2wwaVBsL2c=")
	require.NoError(t, err)
	blockKeyFromSecret, err := base64.StdEncoding.DecodeString("dkZKaG9ZcUxQSkc3dCt5VUtSWHhGcFBTOWtoNEpBYjgvZk9NeDN1bFN4Zw==")
	require.NoError(t, err)

	hashKeyEncoded := string(hashKeyFromSecret)
	blockKeyEncoded := string(blockKeyFromSecret)

	t.Logf("Hash key encoded: %s", hashKeyEncoded)
	t.Logf("Block key encoded: %s", blockKeyEncoded)

	hashKey, err := base64.RawStdEncoding.DecodeString(hashKeyEncoded)
	require.NoError(t, err)
	blockKey, err := base64.RawStdEncoding.DecodeString(blockKeyEncoded)
	require.NoError(t, err)

	t.Logf("Hash key length: %d bytes", len(hashKey))
	t.Logf("Block key length: %d bytes", len(blockKey))

	cookieValue1 := "MTc2OTczMzU0MnxHd3k2YnUxeXRkTTlYa2dIenJKUFhjQ0IwX0tSb2swZE9nUDhSU3I3ZHQyTnNjdGxqM01rQ09SNi1WXzU3Nm9VdHVpMHRsa0RIeHpoVUE1WjhwU29KSHc0d0VVbFprZTNSRnlIa1Z4aGRpeDVsT3VCb05BZ1VyaHVzUFNoazNYQVlaTTdCZHZKakRNWFBrRnY5bjhRVkQ0MUt2b2MybV9tZTk5a2pfdkI3anNLRUNvSURQWWJKOVA5YUhUNFNSVDh4b1hyWUVydXVjZHpZcHhqVndvQjVweko4TlFCX3JaVWtKNU4tYmVCMDBBUzJ3cENFY1d6azFZQjJtWW4wSVZWV0dGa3NSWi12akM1NDVpTklWSU1VeHppMFZGNHc4MEpza05QZm1oN05lWVFFUjJMMXZHOGhBNEVzRUVQR3hZd3RLRG1USFlnTC1FWG5rQURXMkJsWFhtT3lVS3QzLVhSMEp4SEhrT0hUQXZIRy1ZTVg1ZlBKRlZ6clpfM25ZVmZidVJ4V2NLS3U0b21fTzhSTXdyNWZVX0hyOEwxVGtHdlBQWEpRZ1QtYng2Qk5FcUpEWk9YdzlSMFJjdFk5THZfUTJ0VDdEa0lkVU1rb1pWWGN6ZEhtQ1RJeTBhWlpDWmlaWjBoeWNzTTUtd2MyaHhjdmVybFl4LW0wSFozNGM0RTNoNy1US0ZKVlRuc1NKT0tMamVOOU9zb2ZZbXJuUy1veExuN1FqclZycGc0NEh4ZVJxczJvZmxoR1FSVlNVWjNIQUxpMXY2YTdOOUNCN0FLTFkyVmdURmgzdU9QX1JqRmplakZyeWlsWllDRGEyX3NTRzhyTHlQZnZMUWJmbWZWLWF4V1UwX3ZUYi1NUnVBRW03N1ZXRXRYLXB4akFQeWM5VDFzbmtiWlBxMU9SX2VRQWc9PXw1_-70RST-lk0CReGjLEdF6K3-a7Pq31LmoTvZW3nrPg=="
	cookieValue2 := "MTc2OTczMzU0Mnw3OGVhSXhvcGR1N09yT21wQkRMa2puckVzME5FVmJnZnY1aUJZTkVzZWl5WEduY0NueTFuamtvaDh5YTRmWml5c1JieGJ6WTVfM1dBZjQ0eHE5akFMaU4zRVNYdlBnanFUcVh1dnRoQzgtYVg0THU0c2NrcGtiMlFhR001YThqSW4zdXJBenUxc0tLTVlGSnhocWxnbVN3RnJkekF1VGljb09yRUFZZlo4MkpYNElqRHppbEIweE90aU50Qk1jMDhjQk9zSkpLenBNNW9femRuQkZKYzMxeWNtTmU4VzVSamdOT0NIaGJGX1F0UjdOaXRfVzZCM1RSV25OdTI2amY0eDBoeTFydDJOZHl3QVpGSUdjZ1pDNE9IOE1LcVA4MGV3X3VjdWk5NWlYaExnbW9mWEI5U2kwenFIWnZBNFZESnV4UjJxcHpMZ0gtNVZHUm9PY1RDelJCdzlpMC1CTjg0eXRjNlJhcnpKcnhYWTNJejBfQk9Zd25pOC14ZUZOR3BsLWZrT2xzRUJESldlSUp0LWYtdTlDR0hqREZQVHVubENCV3FLU09kTWJKc2h6WHo2Q3BKWlRhVUd1VFZEc2t5QUZ5Z2QwaTFxY09RNklTNWR6VXk4MV85YWRsLXctRnA4dW56bU1lTHhZeFlJMDhvU25UaFptMmVEUEVNMXhPVzRGNERURXRoc2o5LXUxSDY2cmxZT181UUdEWEM0WXNEcHVmb3R3ZWJWZVZKM2xPS0FpN1dQclRJUUE4bFdTQndxNDJnSnRhWUlaRy1sZ2ZqWlFTZXRhUEs2SkE3M0dYR2FIYjlzMjZHRm4zMGd6TlpOYVZ4UzlqeE5rblVONUhtM1pNZ0x1R3lfUU4wUTBoYnNQRi1DODhEWW5ieGZyUGM4ZXpCZEROQ2RZR0R1eEtCZEE9PXyZ83CfdFK3CRERbq2smHnCl-Y_Vj3DcnEObviBKousyA=="

	decoded, err := base64.StdEncoding.DecodeString(cookieValue1)
	if err == nil {
		t.Logf("Cookie 1 raw decoded (first 50 bytes): %s", string(decoded[:50]))
	}
	s := securecookie.New(hashKey, blockKey)
	var firstHalf string
	err = s.Decode("flyte_at_1", cookieValue1, &firstHalf)
	if err != nil {
		t.Logf("Error decoding cookie 1 with MaxAge disabled: %v", err)
		s2 := securecookie.New(hashKey, blockKey)
		err2 := s2.Decode("flyte_at_1", cookieValue1, &firstHalf)
		t.Logf("Error with validation enabled: %v", err2)
	}
	require.NoError(t, err)
	t.Logf("First half of access token: %s", firstHalf)

	var secondHalf string
	err = s.Decode("flyte_at_2", cookieValue2, &secondHalf)
	require.NoError(t, err)
	t.Logf("Second half of access token: %s", secondHalf)

	fullAccessToken := firstHalf + secondHalf
	t.Logf("Full access token: %s", fullAccessToken)
	assert.NotEmpty(t, fullAccessToken)
}
