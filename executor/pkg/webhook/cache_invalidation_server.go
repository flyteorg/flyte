package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

// InvalidateSecretPath is the HTTP path for the cache invalidation endpoint.
const InvalidateSecretPath = "/invalidate-secret"

// InvalidateRequest is the body of a POST to InvalidateSecretPath.
type InvalidateRequest struct {
	Org     string `json:"org"`
	Domain  string `json:"domain"`
	Project string `json:"project"`
	Name    string `json:"name"`
}

// StartCacheInvalidationServer starts a plain HTTP server that listens for cache invalidation
// requests. It exposes POST /invalidate-secret so other services (the secret service) can drop
// secret values cached in the webhook process instead of waiting out the cache TTL.
//
// The server is plain HTTP and unauthenticated, so it must stay cluster-internal: it is bound to
// its own port and never routed through an ingress. Blocks until ctx is cancelled.
func StartCacheInvalidationServer(ctx context.Context, port int, mutator *secret.SecretsPodMutator) error {
	mux := http.NewServeMux()
	mux.HandleFunc(InvalidateSecretPath, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req InvalidateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
			return
		}

		if req.Name == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}

		mutator.InvalidateCache(r.Context(), req.Org, req.Domain, req.Project, req.Name)
		w.WriteHeader(http.StatusOK)
	})

	addr := fmt.Sprintf(":%d", port)
	server := &http.Server{ // #nosec G112 -- internal-only server, not exposed outside the cluster
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		if err := server.Shutdown(context.Background()); err != nil {
			logger.Errorf(ctx, "Failed to shutdown cache invalidation server: %v", err)
		}
	}()

	logger.Infof(ctx, "Starting cache invalidation server on %s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("cache invalidation server failed: %w", err)
	}

	return nil
}
