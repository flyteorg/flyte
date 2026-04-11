package auth

import (
	"net"
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

// isLoopbackRequest returns true when the request originated from the local
// loopback interface. The unified Flyte binary makes intra-process connect-rpc
// calls to its own HTTP mux via http://localhost:<port> (e.g. RunService ->
// ActionsService). Those calls have no Authorization header and must not be
// forced through the external auth gate, or every run creation will fail with
// 401. External traffic (ALB, port-forward from outside the pod) never has a
// loopback RemoteAddr.
func isLoopbackRequest(req *http.Request) bool {
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		host = req.RemoteAddr
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	return ip.IsLoopback()
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

			if isLoopbackRequest(req) {
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
