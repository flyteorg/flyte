package serviceclient

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHTTPClientDisabledReturnsBaseClient(t *testing.T) {
	base := &http.Client{Timeout: time.Second}
	actual, err := NewHTTPClient(context.Background(), base, ServiceConfig{URL: "http://service"})
	require.NoError(t, err)
	assert.Same(t, base, actual)
}

// Concurrent requests through the default base client must reuse their connections.
func TestNewHTTPClientDefaultReusesConnections(t *testing.T) {
	var opened atomic.Int32
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(10 * time.Millisecond)
	}))
	server.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		if state == http.StateNew {
			opened.Add(1)
		}
	}
	server.Start()
	t.Cleanup(server.Close)

	client, err := NewHTTPClient(context.Background(), nil, ServiceConfig{URL: server.URL})
	require.NoError(t, err)

	burst := func() {
		var wg sync.WaitGroup
		for range 50 {
			wg.Go(func() {
				resp, err := client.Get(server.URL)
				if assert.NoError(t, err) {
					_, _ = io.Copy(io.Discard, resp.Body)
					_ = resp.Body.Close()
				}
			})
		}
		wg.Wait()
	}
	burst()
	first := opened.Load()
	burst()
	assert.Equal(t, first, opened.Load(), "second burst opened new connections")
}

func TestNewHTTPClientAuthenticatesAndReusesToken(t *testing.T) {
	var tokenRequests atomic.Int32
	var discoveryRequests atomic.Int32
	mux := http.NewServeMux()
	server := httptest.NewTestServer(t, mux)
	baseClient := server.Client()

	mux.HandleFunc("/oauth2/default"+oauthAuthorizationServerMetadataPath, func(w http.ResponseWriter, _ *http.Request) {
		discoveryRequests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"token_endpoint":"`+server.URL+`/token"}`)
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		tokenRequests.Add(1)
		require.NoError(t, r.ParseForm())
		clientID, clientSecret, ok := r.BasicAuth()
		if !ok {
			clientID, clientSecret = r.Form.Get("client_id"), r.Form.Get("client_secret")
		}
		assert.Equal(t, "executor", clientID)
		assert.Equal(t, "secret", clientSecret)
		assert.Equal(t, "client_credentials", r.Form.Get("grant_type"))
		assert.Equal(t, "events cache", r.Form.Get("scope"))
		assert.Equal(t, "flyte-services", r.Form.Get("audience"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"access_token":"service-token","token_type":"Bearer","expires_in":3600}`)
	})
	mux.HandleFunc("/service", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer service-token", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusNoContent)
	})

	cfg := ServiceConfig{
		URL: server.URL,
		Auth: AuthConfig{
			Type:         AuthTypeOAuth2ClientCredentials,
			IssuerURL:    server.URL + "/oauth2/default",
			ClientID:     "executor",
			ClientSecret: "secret",
			Scopes:       []string{"events", "cache"},
			Audience:     "flyte-services",
		},
	}
	client, err := NewHTTPClient(context.Background(), baseClient, cfg)
	require.NoError(t, err)

	for range 2 {
		resp, requestErr := client.Get(cfg.URL + "/service")
		require.NoError(t, requestErr)
		require.NoError(t, resp.Body.Close())
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	}
	assert.Equal(t, int32(1), tokenRequests.Load())
	assert.Equal(t, int32(1), discoveryRequests.Load())
}

func TestNewHTTPClientReadsClientSecretFile(t *testing.T) {
	secretFile, err := os.CreateTemp(t.TempDir(), "oauth-secret-")
	require.NoError(t, err)
	_, err = secretFile.WriteString("file-secret\n")
	require.NoError(t, err)
	require.NoError(t, secretFile.Close())

	mux := http.NewServeMux()
	server := httptest.NewTestServer(t, mux)
	baseClient := server.Client()
	mux.HandleFunc(oauthAuthorizationServerMetadataPath, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"token_endpoint":"`+server.URL+`/token"}`)
	})
	mux.HandleFunc("/service", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			assert.Equal(t, "Bearer token", r.Header.Get("Authorization"))
			w.WriteHeader(http.StatusNoContent)
		}
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		clientID, clientSecret, ok := r.BasicAuth()
		if !ok {
			clientID, clientSecret = r.Form.Get("client_id"), r.Form.Get("client_secret")
		}
		assert.Equal(t, "executor", clientID)
		assert.Equal(t, "file-secret", clientSecret)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"access_token":"token","token_type":"Bearer","expires_in":3600}`)
	})

	client, err := NewHTTPClient(context.Background(), baseClient, ServiceConfig{
		URL: server.URL,
		Auth: AuthConfig{
			Type:             AuthTypeOAuth2ClientCredentials,
			IssuerURL:        server.URL,
			ClientID:         "executor",
			ClientSecretFile: secretFile.Name(),
		},
	})
	require.NoError(t, err)
	resp, err := client.Get(server.URL + "/service")
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())
}

func TestNewHTTPClientValidatesConfiguration(t *testing.T) {
	tests := []struct {
		name string
		cfg  ServiceConfig
		want string
	}{
		{name: "service URL", cfg: ServiceConfig{}, want: "url is required"},
		{name: "missing auth type", cfg: ServiceConfig{URL: "https://service", Auth: AuthConfig{ClientID: "id"}}, want: "type is required"},
		{name: "unsupported auth type", cfg: ServiceConfig{URL: "https://service", Auth: AuthConfig{Type: "Unknown"}}, want: "unsupported type"},
		{name: "issuer URL", cfg: ServiceConfig{URL: "https://service", Auth: AuthConfig{Type: AuthTypeOAuth2ClientCredentials, ClientID: "id"}}, want: "issuerUrl is required"},
		{name: "client ID", cfg: ServiceConfig{URL: "https://service", Auth: AuthConfig{Type: AuthTypeOAuth2ClientCredentials, IssuerURL: "https://idp"}}, want: "clientId is required"},
		{name: "secret", cfg: ServiceConfig{URL: "https://service", Auth: AuthConfig{Type: AuthTypeOAuth2ClientCredentials, IssuerURL: "https://idp", ClientID: "id"}}, want: "clientSecret or clientSecretFile is required"},
		{name: "two secrets", cfg: ServiceConfig{URL: "https://service", Auth: AuthConfig{Type: AuthTypeOAuth2ClientCredentials, IssuerURL: "https://idp", ClientID: "id", ClientSecret: "a", ClientSecretFile: "b"}}, want: "mutually exclusive"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewHTTPClient(context.Background(), http.DefaultClient, tt.cfg)
			require.ErrorContains(t, err, tt.want)
		})
	}
}
