package authzserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/auth"
	"github.com/flyteorg/flyte/v2/runs/service/auth/config"
)

func TestGetJwksForIssuer_Success(t *testing.T) {
	metadata := &auth.GetOAuth2MetadataResponse{
		JwksUri: "https://example.com/.well-known/jwks.json",
	}
	metadataJSON, err := json.Marshal(metadata)
	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(metadataJSON)
	}))
	defer ts.Close()

	cfg := config.ExternalAuthorizationServer{
		BaseURL: mustParseTestURL(ts.URL),
	}

	keySet, err := getJwksForIssuer(context.Background(), cfg.BaseURL.URL, cfg)
	require.NoError(t, err)
	assert.NotNil(t, keySet)
}

func TestGetJwksForIssuer_CustomMetadataURL(t *testing.T) {
	metadata := &auth.GetOAuth2MetadataResponse{
		JwksUri: "https://example.com/.well-known/jwks.json",
	}
	metadataJSON, err := json.Marshal(metadata)
	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/custom/metadata", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.Write(metadataJSON)
	}))
	defer ts.Close()

	cfg := config.ExternalAuthorizationServer{
		BaseURL:             mustParseTestURL(ts.URL),
		MetadataEndpointURL: mustParseTestURL(ts.URL + "/custom/metadata"),
	}

	keySet, err := getJwksForIssuer(context.Background(), cfg.BaseURL.URL, cfg)
	require.NoError(t, err)
	assert.NotNil(t, keySet)
}

func TestGetJwksForIssuer_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer ts.Close()

	cfg := config.ExternalAuthorizationServer{
		BaseURL: mustParseTestURL(ts.URL),
	}

	_, err := getJwksForIssuer(context.Background(), cfg.BaseURL.URL, cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestNewOAuth2ResourceServer_FallbackBaseURL(t *testing.T) {
	metadata := &auth.GetOAuth2MetadataResponse{
		JwksUri: "https://example.com/.well-known/jwks.json",
	}
	metadataJSON, err := json.Marshal(metadata)
	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(metadataJSON)
	}))
	defer ts.Close()

	cfg := config.ExternalAuthorizationServer{
		AllowedAudience: []string{"https://flyte.example.com"},
	}
	fallback := mustParseTestURL(ts.URL)

	rs, err := NewOAuth2ResourceServer(context.Background(), cfg, config.URL(fallback))
	require.NoError(t, err)
	assert.NotNil(t, rs)
	assert.Equal(t, []string{"https://flyte.example.com"}, rs.allowedAudience)
}
