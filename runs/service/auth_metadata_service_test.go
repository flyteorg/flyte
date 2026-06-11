package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/auth"
)

func TestGetOAuth2Metadata_NotConfigured(t *testing.T) {
	svc := NewAuthMetadataService("example.com", ExternalAuthServerConfig{})
	_, err := svc.GetOAuth2Metadata(context.Background(), connect.NewRequest(&auth.GetOAuth2MetadataRequest{}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeUnimplemented, connect.CodeOf(err))
}

func TestGetOAuth2Metadata_External(t *testing.T) {
	// Fake external IdP returning RFC 8414 snake_case metadata with extra fields
	// (as Okta does) to exercise DiscardUnknown.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/oauth2/default/.well-known/oauth-authorization-server", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"issuer": "https://idp.example.com/oauth2/default",
			"authorization_endpoint": "https://idp.example.com/oauth2/default/v1/authorize",
			"token_endpoint": "https://idp.example.com/oauth2/default/v1/token",
			"jwks_uri": "https://idp.example.com/oauth2/default/v1/keys",
			"introspection_endpoint": "https://idp.example.com/oauth2/default/v1/introspect",
			"response_modes_supported": ["query", "fragment", "form_post"],
			"claims_supported": ["iss", "sub", "aud"]
		}`))
	}))
	defer srv.Close()

	svc := NewAuthMetadataService("example.com", ExternalAuthServerConfig{
		BaseURL: srv.URL + "/oauth2/default",
	})
	resp, err := svc.GetOAuth2Metadata(context.Background(), connect.NewRequest(&auth.GetOAuth2MetadataRequest{}))
	require.NoError(t, err)
	assert.Equal(t, "https://idp.example.com/oauth2/default", resp.Msg.Issuer)
	assert.Equal(t, "https://idp.example.com/oauth2/default/v1/authorize", resp.Msg.AuthorizationEndpoint)
	assert.Equal(t, "https://idp.example.com/oauth2/default/v1/token", resp.Msg.TokenEndpoint)
	assert.Equal(t, "https://idp.example.com/oauth2/default/v1/keys", resp.Msg.JwksUri)
}

func TestGetOAuth2Metadata_ExternalCustomMetadataURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/custom/metadata", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"issuer":"https://idp.example.com"}`))
	}))
	defer srv.Close()

	svc := NewAuthMetadataService("example.com", ExternalAuthServerConfig{
		BaseURL:     srv.URL,
		MetadataURL: "custom/metadata",
	})
	resp, err := svc.GetOAuth2Metadata(context.Background(), connect.NewRequest(&auth.GetOAuth2MetadataRequest{}))
	require.NoError(t, err)
	assert.Equal(t, "https://idp.example.com", resp.Msg.Issuer)
}

func TestGetOAuth2Metadata_ExternalUnavailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	svc := NewAuthMetadataService("example.com", ExternalAuthServerConfig{
		BaseURL:       srv.URL,
		RetryAttempts: 1,
		RetryDelay:    time.Millisecond,
	})
	_, err := svc.GetOAuth2Metadata(context.Background(), connect.NewRequest(&auth.GetOAuth2MetadataRequest{}))
	require.Error(t, err)
	assert.Equal(t, connect.CodeUnavailable, connect.CodeOf(err))
}

func TestOAuth2MetadataHTTPHandler(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"issuer":"https://idp.example.com/oauth2/default","token_endpoint":"https://idp.example.com/oauth2/default/v1/token"}`))
	}))
	defer srv.Close()

	svc := NewAuthMetadataService("example.com", ExternalAuthServerConfig{BaseURL: srv.URL + "/oauth2/default"})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/.well-known/oauth-authorization-server", nil)
	OAuth2MetadataHTTPHandler(svc).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")

	var doc map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &doc))
	// proto3 JSON marshals to camelCase
	assert.Equal(t, "https://idp.example.com/oauth2/default", doc["issuer"])
	assert.Equal(t, "https://idp.example.com/oauth2/default/v1/token", doc["tokenEndpoint"])
}

func TestOAuth2MetadataHTTPHandler_NotConfigured(t *testing.T) {
	svc := NewAuthMetadataService("example.com", ExternalAuthServerConfig{})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/.well-known/oauth-authorization-server", nil)
	OAuth2MetadataHTTPHandler(svc).ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotImplemented, rec.Code)
}
