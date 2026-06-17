package service

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
)

const (
	// albIdentityHeader is set by ALB authenticate-oidc (browser/cookie path) and
	// carries the OIDC subject (`sub`) directly.
	albIdentityHeader = "X-Amzn-Oidc-Identity"
	// authorizationHeader carries the Bearer token on the JWT-validation path
	// (SDK/CLI). The load balancer validates it and forwards it unchanged.
	authorizationHeader = "Authorization"
	bearerPrefix        = "Bearer "
)

// subjectFromHeaders extracts the authenticated subject (OIDC `sub`) from the auth
// headers the load balancer forwards. Auth is enforced upstream (e.g. ALB OIDC /
// JWT validation), so the claims are trusted and only decoded here — not
// re-verified. Returns "" when no authenticated identity is present.
func subjectFromHeaders(h http.Header) string {
	// authenticate-oidc (browser/cookie) path: subject is forwarded directly.
	if sub := strings.TrimSpace(h.Get(albIdentityHeader)); sub != "" {
		return sub
	}
	// JWT (SDK/CLI) path: read the `sub` claim from the forwarded Bearer token.
	return subjectFromBearer(h.Get(authorizationHeader))
}

// subjectFromBearer returns the `sub` claim of a Bearer JWT without verifying its
// signature (the load balancer already validated it). Returns "" on any malformed input.
func subjectFromBearer(authz string) string {
	if len(authz) <= len(bearerPrefix) || !strings.EqualFold(authz[:len(bearerPrefix)], bearerPrefix) {
		return ""
	}
	parts := strings.Split(strings.TrimSpace(authz[len(bearerPrefix):]), ".")
	if len(parts) != 3 {
		return ""
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}
	var claims struct {
		Sub string `json:"sub"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return ""
	}
	return claims.Sub
}

// subjectOnlyIdentity builds a minimal EnrichedIdentity carrying just the subject.
// Mirrors the cloud transformer fallback; the standalone runs service has no
// identity service to enrich the subject into full user details (email, name, groups).
func subjectOnlyIdentity(subject string) *common.EnrichedIdentity {
	if subject == "" {
		return nil
	}
	return &common.EnrichedIdentity{
		Principal: &common.EnrichedIdentity_User{
			User: &common.User{
				Id: &common.UserIdentifier{Subject: subject},
			},
		},
	}
}
