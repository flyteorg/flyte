package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/flytepropeller/pkg/secret"
	"github.com/flyteorg/flyte/flytepropeller/pkg/secret/config"
	"github.com/flyteorg/flyte/flytepropeller/pkg/secret/mocks"
)

func newTestMutatorWithMockInjector(t *testing.T, mockInjector *mocks.SecretsInjector) *secret.SecretsPodMutator {
	t.Helper()
	mockInjector.OnType().Return(config.SecretManagerTypeEmbedded)
	return secret.NewSecretsMutatorFromInjectors(
		[]config.SecretManagerType{config.SecretManagerTypeEmbedded},
		map[config.SecretManagerType]secret.SecretsInjector{
			config.SecretManagerTypeEmbedded: mockInjector,
		},
	)
}

func TestCacheInvalidationHandler_Success(t *testing.T) {
	mockInjector := &mocks.SecretsInjector{}
	mockInjector.On("InvalidateCache", context.Background(), "org1", "domain1", "project1", "secret1").Return()
	mockInjector.OnType().Return(config.SecretManagerTypeEmbedded)
	mutator := secret.NewSecretsMutatorFromInjectors(
		[]config.SecretManagerType{config.SecretManagerTypeEmbedded},
		map[config.SecretManagerType]secret.SecretsInjector{
			config.SecretManagerTypeEmbedded: mockInjector,
		},
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/invalidate-secret", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req invalidateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.Name == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}

		mutator.InvalidateCache(r.Context(), req.Org, req.Domain, req.Project, req.Name)
		w.WriteHeader(http.StatusOK)
	})

	body, _ := json.Marshal(invalidateRequest{
		Org:     "org1",
		Domain:  "domain1",
		Project: "project1",
		Name:    "secret1",
	})

	req := httptest.NewRequest(http.MethodPost, "/invalidate-secret", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	mockInjector.AssertCalled(t, "InvalidateCache", context.Background(), "org1", "domain1", "project1", "secret1")
}

func TestCacheInvalidationHandler_MethodNotAllowed(t *testing.T) {
	mockInjector := &mocks.SecretsInjector{}
	mutator := newTestMutatorWithMockInjector(t, mockInjector)

	mux := http.NewServeMux()
	mux.HandleFunc("/invalidate-secret", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		mutator.InvalidateCache(r.Context(), "", "", "", "")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/invalidate-secret", nil)
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestCacheInvalidationHandler_MissingName(t *testing.T) {
	mockInjector := &mocks.SecretsInjector{}
	mutator := newTestMutatorWithMockInjector(t, mockInjector)

	mux := http.NewServeMux()
	mux.HandleFunc("/invalidate-secret", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req invalidateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.Name == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}

		mutator.InvalidateCache(r.Context(), req.Org, req.Domain, req.Project, req.Name)
		w.WriteHeader(http.StatusOK)
	})

	body, _ := json.Marshal(invalidateRequest{
		Org:    "org1",
		Domain: "domain1",
	})

	req := httptest.NewRequest(http.MethodPost, "/invalidate-secret", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCacheInvalidationHandler_InvalidJSON(t *testing.T) {
	mockInjector := &mocks.SecretsInjector{}
	mutator := newTestMutatorWithMockInjector(t, mockInjector)

	mux := http.NewServeMux()
	mux.HandleFunc("/invalidate-secret", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req invalidateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.Name == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}

		mutator.InvalidateCache(r.Context(), req.Org, req.Domain, req.Project, req.Name)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/invalidate-secret", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	mux.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}
