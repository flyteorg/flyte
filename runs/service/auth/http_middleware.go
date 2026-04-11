package auth

import (
	"net/http"
	"strings"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

// publicPathPrefixes lists request paths that never require authentication.
// These must cover health probes, the browser OAuth2/OIDC flow, metadata
// discovery, and the AuthMetadataService (which clients call *before* they
// have a token).
var publicPathPrefixes = []string{
	"/healthz",
	"/readyz",
	"/healthcheck",
	"/login",
	"/callback",
	"/logout",
	"/.well-known/",
	"/flyteidl2.auth.AuthMetadataService/",
}

// IsPublicPath reports whether an HTTP request path bypasses authentication.
func IsPublicPath(path string) bool {
	for _, p := range publicPathPrefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// GetAuthenticationHTTPInterceptor returns middleware that validates a bearer
// token or auth cookies on incoming HTTP requests and injects the resulting
// IdentityContext into the request context. Public paths (see IsPublicPath)
// pass through without validation. When DisableForHTTP is set on the config,
// every request passes through unchanged.
func GetAuthenticationHTTPInterceptor(h *AuthHandlerConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if IsPublicPath(req.URL.Path) {
				next.ServeHTTP(w, req)
				return
			}

			if h.AuthConfig.DisableForHTTP {
				next.ServeHTTP(w, req)
				return
			}

			ctx := req.Context()
			identity, err := IdentityContextFromRequest(ctx, req, h)
			if err != nil {
				logger.Infof(ctx, "unauthenticated request to %s: %v", req.URL.Path, err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, req.WithContext(identity.WithContext(ctx)))
		})
	}
}
