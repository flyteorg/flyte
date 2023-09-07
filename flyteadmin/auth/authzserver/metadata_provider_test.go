package authzserver

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	config2 "github.com/flyteorg/flytestdlib/config"

	"github.com/flyteorg/flyteadmin/auth/config"
	authConfig "github.com/flyteorg/flyteadmin/auth/config"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/stretchr/testify/assert"
)

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
