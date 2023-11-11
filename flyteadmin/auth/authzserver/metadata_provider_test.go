package authzserver

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteadmin/auth/config"
	authConfig "github.com/flyteorg/flyte/flyteadmin/auth/config"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	config2 "github.com/flyteorg/flyte/flytestdlib/config"
)

var oauthMetadataFailureErrorMessage = "Failed to get oauth metadata."

func TestOAuth2MetadataProvider_FlyteClient(t *testing.T) {
	provider := NewService(&authConfig.Config{
		AppAuth: authConfig.OAuth2Options{
			ThirdParty: authConfig.ThirdPartyConfigOptions{
				FlyteClientConfig: authConfig.FlyteClientConfig{
					ClientID:    "my-client",
					RedirectURI: "client/",
					Scopes:      []string{"all"},
					Audience:    "http://dummyServer",
				},
			},
		},
	})

	ctx := context.Background()
	resp, err := provider.GetPublicClientConfig(ctx, &service.PublicClientAuthConfigRequest{})
	assert.NoError(t, err)
	assert.Equal(t, "my-client", resp.ClientId)
	assert.Equal(t, "client/", resp.RedirectUri)
	assert.Equal(t, []string{"all"}, resp.Scopes)
	assert.Equal(t, "http://dummyServer", resp.Audience)
}

func TestOAuth2MetadataProvider_OAuth2Metadata(t *testing.T) {
	t.Run("Self AuthServer", func(t *testing.T) {
		provider := NewService(&authConfig.Config{
			AuthorizedURIs: []config2.URL{{URL: *config.MustParseURL("https://issuer/")}},
		})

		ctx := context.Background()
		resp, err := provider.GetOAuth2Metadata(ctx, &service.OAuth2MetadataRequest{})
		assert.NoError(t, err)
		assert.Equal(t, "https://issuer/", resp.Issuer)
	})

	var issuer string
	hf := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/.well-known/oauth-authorization-server" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(w, strings.ReplaceAll(`{
				"issuer": "https://dev-14186422.okta.com",
				"authorization_endpoint": "https://example.com/auth",
				"token_endpoint": "https://example.com/token",
				"jwks_uri": "https://example.com/keys",
				"id_token_signing_alg_values_supported": ["RS256"]
			}`, "ISSUER", issuer))
		if !assert.NoError(t, err) {
			t.FailNow()
		}
	}

	s := httptest.NewServer(http.HandlerFunc(hf))
	defer s.Close()

	http.DefaultClient = s.Client()

	t.Run("External AuthServer", func(t *testing.T) {
		provider := NewService(&authConfig.Config{
			AuthorizedURIs: []config2.URL{{URL: *config.MustParseURL("https://issuer/")}},
			AppAuth: authConfig.OAuth2Options{
				AuthServerType: authConfig.AuthorizationServerTypeExternal,
				ExternalAuthServer: authConfig.ExternalAuthorizationServer{
					BaseURL: config2.URL{URL: *config.MustParseURL(s.URL)},
				},
			},
		})

		ctx := context.Background()
		resp, err := provider.GetOAuth2Metadata(ctx, &service.OAuth2MetadataRequest{})
		assert.NoError(t, err)
		assert.Equal(t, "https://dev-14186422.okta.com", resp.Issuer)
	})

	t.Run("External AuthServer fallback url", func(t *testing.T) {
		provider := NewService(&authConfig.Config{
			AuthorizedURIs: []config2.URL{{URL: *config.MustParseURL("https://issuer/")}},
			AppAuth: authConfig.OAuth2Options{
				AuthServerType: authConfig.AuthorizationServerTypeExternal,
			},
			UserAuth: authConfig.UserAuthConfig{
				OpenID: authConfig.OpenIDOptions{
					BaseURL: config2.URL{URL: *config.MustParseURL(s.URL)},
				},
			},
		})

		ctx := context.Background()
		resp, err := provider.GetOAuth2Metadata(ctx, &service.OAuth2MetadataRequest{})
		assert.NoError(t, err)
		assert.Equal(t, "https://dev-14186422.okta.com", resp.Issuer)
	})
}

func TestSendAndRetryHttpRequest(t *testing.T) {
	t.Run("Retry into failure", func(t *testing.T) {
		requestAttempts := 0
		hf := func(w http.ResponseWriter, r *http.Request) {
			switch strings.TrimSpace(r.URL.Path) {
			case "/":
				w.WriteHeader(500)
				requestAttempts++
			default:
				http.NotFoundHandler().ServeHTTP(w, r)
			}
		}

		server := httptest.NewServer(http.HandlerFunc(hf))
		defer server.Close()
		http.DefaultClient = server.Client()
		retryAttempts := 5
		totalAttempts := retryAttempts + 1 // 1 for the initial try

		resp, err := sendAndRetryHTTPRequest(context.Background(), server.Client(), server.URL, retryAttempts, 0 /* for testing */)
		assert.Error(t, err)
		assert.Equal(t, oauthMetadataFailureErrorMessage, err.Error())
		assert.Nil(t, resp)
		assert.Equal(t, totalAttempts, requestAttempts)
	})

	t.Run("Retry into success", func(t *testing.T) {
		requestAttempts := 0
		hf := func(w http.ResponseWriter, r *http.Request) {
			switch strings.TrimSpace(r.URL.Path) {
			case "/":
				if requestAttempts > 2 {
					w.WriteHeader(200)
				} else {
					requestAttempts++
					w.WriteHeader(500)
				}
			default:
				http.NotFoundHandler().ServeHTTP(w, r)
			}
		}

		server := httptest.NewServer(http.HandlerFunc(hf))
		defer server.Close()
		http.DefaultClient = server.Client()
		retryAttempts := 5
		expectedRequestAttempts := 3

		resp, err := sendAndRetryHTTPRequest(context.Background(), server.Client(), server.URL, retryAttempts, 0 /* for testing */)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, expectedRequestAttempts, requestAttempts)
	})

	t.Run("Success", func(t *testing.T) {
		requestAttempts := 0
		hf := func(w http.ResponseWriter, r *http.Request) {
			switch strings.TrimSpace(r.URL.Path) {
			case "/":
				w.WriteHeader(200)
			default:
				http.NotFoundHandler().ServeHTTP(w, r)
			}
		}

		server := httptest.NewServer(http.HandlerFunc(hf))
		defer server.Close()
		http.DefaultClient = server.Client()
		retryAttempts := 5
		expectedRequestAttempts := 0

		resp, err := sendAndRetryHTTPRequest(context.Background(), server.Client(), server.URL, retryAttempts, 0 /* for testing */)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 200, resp.StatusCode)
		assert.Equal(t, expectedRequestAttempts, requestAttempts)
	})

}
