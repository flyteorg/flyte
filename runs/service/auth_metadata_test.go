package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/auth"
	"github.com/flyteorg/flyte/v2/runs/config"

	stdlibConfig "github.com/flyteorg/flyte/v2/flytestdlib/config"
)

func mustParseURL(rawURL string) stdlibConfig.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return stdlibConfig.URL{URL: *u}
}

func TestGetPublicClientConfig(t *testing.T) {
	cfg := config.AuthConfig{
		GrpcAuthorizationHeader: "flyte-authorization",
		AppAuth: config.OAuth2Options{
			ThirdParty: config.ThirdPartyConfigOptions{
				FlyteClientConfig: config.FlyteClientConfig{
					ClientID:    "flyte-client",
					RedirectURI: "http://localhost:12345/callback",
					Scopes:      []string{"openid", "offline"},
					Audience:    "https://flyte.example.com",
				},
			},
		},
	}

	svc := NewAuthMetadataService(cfg)
	resp, err := svc.GetPublicClientConfig(context.Background(), connect.NewRequest(&auth.GetPublicClientConfigRequest{}))
	require.NoError(t, err)

	msg := resp.Msg
	assert.Equal(t, "flyte-client", msg.ClientId)
	assert.Equal(t, "http://localhost:12345/callback", msg.RedirectUri)
	assert.Equal(t, []string{"openid", "offline"}, msg.Scopes)
	assert.Equal(t, "flyte-authorization", msg.AuthorizationMetadataKey)
	assert.Equal(t, "https://flyte.example.com", msg.Audience)
}

func TestGetOAuth2Metadata_SelfAuthServer(t *testing.T) {
	cfg := config.AuthConfig{
		AuthorizedURIs: []stdlibConfig.URL{
			mustParseURL("https://flyte.example.com"),
		},
		AppAuth: config.OAuth2Options{
			AuthServerType: config.AuthorizationServerTypeSelf,
		},
	}

	svc := NewAuthMetadataService(cfg)
	resp, err := svc.GetOAuth2Metadata(context.Background(), connect.NewRequest(&auth.GetOAuth2MetadataRequest{}))
	require.NoError(t, err)

	msg := resp.Msg
	assert.Equal(t, "https://flyte.example.com", msg.Issuer)
	assert.Equal(t, "https://flyte.example.com/oauth2/authorize", msg.AuthorizationEndpoint)
	assert.Equal(t, "https://flyte.example.com/oauth2/token", msg.TokenEndpoint)
	assert.Equal(t, "https://flyte.example.com/oauth2/jwks", msg.JwksUri)
	assert.Equal(t, []string{"S256"}, msg.CodeChallengeMethodsSupported)
	assert.Equal(t, []string{"code", "token", "code token"}, msg.ResponseTypesSupported)
	assert.Equal(t, []string{"client_credentials", "refresh_token", "authorization_code"}, msg.GrantTypesSupported)
	assert.Equal(t, []string{"all"}, msg.ScopesSupported)
	assert.Equal(t, []string{"client_secret_basic"}, msg.TokenEndpointAuthMethodsSupported)
}

func TestGetOAuth2Metadata_SelfAuthServerWithCustomIssuer(t *testing.T) {
	cfg := config.AuthConfig{
		AuthorizedURIs: []stdlibConfig.URL{
			mustParseURL("https://flyte.example.com"),
		},
		AppAuth: config.OAuth2Options{
			AuthServerType: config.AuthorizationServerTypeSelf,
			SelfAuthServer: config.AuthorizationServer{
				Issuer: "https://custom-issuer.example.com",
			},
		},
	}

	svc := NewAuthMetadataService(cfg)
	resp, err := svc.GetOAuth2Metadata(context.Background(), connect.NewRequest(&auth.GetOAuth2MetadataRequest{}))
	require.NoError(t, err)

	msg := resp.Msg
	assert.Equal(t, "https://custom-issuer.example.com", msg.Issuer)
	assert.Equal(t, "https://flyte.example.com/oauth2/authorize", msg.AuthorizationEndpoint)
}

func TestGetOAuth2Metadata_SelfAuthServerNoAuthorizedURIs(t *testing.T) {
	cfg := config.AuthConfig{
		AppAuth: config.OAuth2Options{
			AuthServerType: config.AuthorizationServerTypeSelf,
		},
	}

	svc := NewAuthMetadataService(cfg)
	resp, err := svc.GetOAuth2Metadata(context.Background(), connect.NewRequest(&auth.GetOAuth2MetadataRequest{}))
	require.NoError(t, err)

	msg := resp.Msg
	assert.Equal(t, "http://localhost:8090", msg.Issuer)
	assert.Equal(t, "http://localhost:8090/oauth2/token", msg.TokenEndpoint)
}

func TestGetOAuth2Metadata_ExternalAuthServer(t *testing.T) {
	expectedMetadata := &auth.GetOAuth2MetadataResponse{
		Issuer:                "https://external-idp.example.com",
		AuthorizationEndpoint: "https://external-idp.example.com/authorize",
		TokenEndpoint:         "https://external-idp.example.com/token",
		JwksUri:               "https://external-idp.example.com/.well-known/jwks.json",
	}

	metadataJSON, err := json.Marshal(expectedMetadata)
	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/"+oauth2MetadataEndpoint, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.Write(metadataJSON)
	}))
	defer ts.Close()

	cfg := config.AuthConfig{
		AppAuth: config.OAuth2Options{
			AuthServerType: config.AuthorizationServerTypeExternal,
			ExternalAuthServer: config.ExternalAuthorizationServer{
				BaseURL:       mustParseURL(ts.URL),
				RetryAttempts: 1,
				RetryDelay:    stdlibConfig.Duration{Duration: 100 * time.Millisecond},
			},
		},
	}

	svc := NewAuthMetadataService(cfg)
	resp, err := svc.GetOAuth2Metadata(context.Background(), connect.NewRequest(&auth.GetOAuth2MetadataRequest{}))
	require.NoError(t, err)

	msg := resp.Msg
	assert.Equal(t, "https://external-idp.example.com", msg.Issuer)
	assert.Equal(t, "https://external-idp.example.com/authorize", msg.AuthorizationEndpoint)
	assert.Equal(t, "https://external-idp.example.com/token", msg.TokenEndpoint)
	assert.Equal(t, "https://external-idp.example.com/.well-known/jwks.json", msg.JwksUri)
}

func TestGetOAuth2Metadata_ExternalWithCustomMetadataURL(t *testing.T) {
	expectedMetadata := &auth.GetOAuth2MetadataResponse{
		Issuer:        "https://external-idp.example.com",
		TokenEndpoint: "https://external-idp.example.com/token",
	}

	metadataJSON, err := json.Marshal(expectedMetadata)
	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/custom/metadata", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.Write(metadataJSON)
	}))
	defer ts.Close()

	cfg := config.AuthConfig{
		AppAuth: config.OAuth2Options{
			AuthServerType: config.AuthorizationServerTypeExternal,
			ExternalAuthServer: config.ExternalAuthorizationServer{
				BaseURL:             mustParseURL(ts.URL),
				MetadataEndpointURL: mustParseURL(ts.URL + "/custom/metadata"),
				RetryAttempts:       1,
				RetryDelay:          stdlibConfig.Duration{Duration: 100 * time.Millisecond},
			},
		},
	}

	svc := NewAuthMetadataService(cfg)
	resp, err := svc.GetOAuth2Metadata(context.Background(), connect.NewRequest(&auth.GetOAuth2MetadataRequest{}))
	require.NoError(t, err)

	assert.Equal(t, "https://external-idp.example.com", resp.Msg.Issuer)
	assert.Equal(t, "https://external-idp.example.com/token", resp.Msg.TokenEndpoint)
}

func TestGetOAuth2Metadata_ExternalWithTokenProxy(t *testing.T) {
	expectedMetadata := &auth.GetOAuth2MetadataResponse{
		Issuer:                "https://external-idp.example.com",
		AuthorizationEndpoint: "https://external-idp.example.com/authorize",
		TokenEndpoint:         "https://external-idp.example.com/oauth/token",
	}

	metadataJSON, err := json.Marshal(expectedMetadata)
	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(metadataJSON)
	}))
	defer ts.Close()

	cfg := config.AuthConfig{
		AuthorizedURIs: []stdlibConfig.URL{
			mustParseURL("https://flyte.example.com"),
		},
		AppAuth: config.OAuth2Options{
			AuthServerType: config.AuthorizationServerTypeExternal,
			ExternalAuthServer: config.ExternalAuthorizationServer{
				BaseURL:       mustParseURL(ts.URL),
				RetryAttempts: 1,
				RetryDelay:    stdlibConfig.Duration{Duration: 100 * time.Millisecond},
			},
		},
		TokenEndpointProxyConfig: config.TokenEndpointProxyConfig{
			Enabled: true,
		},
	}

	svc := NewAuthMetadataService(cfg)
	resp, err := svc.GetOAuth2Metadata(context.Background(), connect.NewRequest(&auth.GetOAuth2MetadataRequest{}))
	require.NoError(t, err)

	msg := resp.Msg
	assert.Equal(t, "https://external-idp.example.com", msg.Issuer)
	assert.Equal(t, "https://external-idp.example.com/authorize", msg.AuthorizationEndpoint)
	// Token endpoint should be rewritten to the public URL
	assert.Equal(t, "https://flyte.example.com/oauth/token", msg.TokenEndpoint)
}

func TestGetOAuth2Metadata_ExternalWithTokenProxyAndPathPrefix(t *testing.T) {
	expectedMetadata := &auth.GetOAuth2MetadataResponse{
		TokenEndpoint: "https://external-idp.example.com/oauth/token",
	}

	metadataJSON, err := json.Marshal(expectedMetadata)
	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(metadataJSON)
	}))
	defer ts.Close()

	cfg := config.AuthConfig{
		AppAuth: config.OAuth2Options{
			AuthServerType: config.AuthorizationServerTypeExternal,
			ExternalAuthServer: config.ExternalAuthorizationServer{
				BaseURL:       mustParseURL(ts.URL),
				RetryAttempts: 1,
				RetryDelay:    stdlibConfig.Duration{Duration: 100 * time.Millisecond},
			},
		},
		TokenEndpointProxyConfig: config.TokenEndpointProxyConfig{
			Enabled:    true,
			PublicURL:  mustParseURL("https://proxy.example.com"),
			PathPrefix: "api/v1",
		},
	}

	svc := NewAuthMetadataService(cfg)
	resp, err := svc.GetOAuth2Metadata(context.Background(), connect.NewRequest(&auth.GetOAuth2MetadataRequest{}))
	require.NoError(t, err)

	assert.Equal(t, "https://proxy.example.com/api/v1/oauth/token", resp.Msg.TokenEndpoint)
}

func TestGetOAuth2Metadata_ExternalNoBaseURL(t *testing.T) {
	cfg := config.AuthConfig{
		AppAuth: config.OAuth2Options{
			AuthServerType: config.AuthorizationServerTypeExternal,
		},
	}

	svc := NewAuthMetadataService(cfg)
	_, err := svc.GetOAuth2Metadata(context.Background(), connect.NewRequest(&auth.GetOAuth2MetadataRequest{}))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "external auth server base URL is not configured")
}

func TestSendAndRetryHTTPRequest_ImmediateSuccess(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	body, err := sendAndRetryHTTPRequest(context.Background(), http.DefaultClient, ts.URL, 3, 10*time.Millisecond)
	require.NoError(t, err)
	assert.Equal(t, `{"status":"ok"}`, string(body))
}

func TestSendAndRetryHTTPRequest_RetryIntoSuccess(t *testing.T) {
	attempt := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		if attempt < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer ts.Close()

	body, err := sendAndRetryHTTPRequest(context.Background(), http.DefaultClient, ts.URL, 5, 10*time.Millisecond)
	require.NoError(t, err)
	assert.Equal(t, `{"status":"ok"}`, string(body))
	assert.Equal(t, 3, attempt)
}

func TestSendAndRetryHTTPRequest_AllRetrysFail(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	_, err := sendAndRetryHTTPRequest(context.Background(), http.DefaultClient, ts.URL, 3, 10*time.Millisecond)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "all 3 attempts failed")
}

func TestSendAndRetryHTTPRequest_ContextCancelled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := sendAndRetryHTTPRequest(ctx, http.DefaultClient, ts.URL, 5, 10*time.Millisecond)
	require.Error(t, err)
}
