package service

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/runs/config"
)

const (
	// authorizationHeader carries the Bearer token on the JWT-validation path
	// (SDK/CLI). The load balancer validates it and forwards it unchanged. This path
	// is proxy-agnostic, so its header is fixed rather than configurable.
	authorizationHeader = "Authorization"
	bearerPrefix        = "Bearer "
)

// subjectFromHeaders returns the OIDC subject of the caller from the auth headers the
// proxy forwards, using the configured header names (defaults match AWS ALB; works for
// oauth2-proxy/Traefik when configured). Auth is enforced upstream, so the forwarded
// values are trusted and only decoded here — not re-verified. We capture just the
// subject; the display identity (name/email) is resolved from it at read time. Returns
// "" when no authenticated identity is present.
func subjectFromHeaders(h http.Header, cfg config.IdentityHeadersConfig) string {
	// Proxy that forwards the subject inside a claims JWT (ALB authenticate-oidc data header).
	if cfg.ClaimsJWTHeader != "" {
		if sub := subjectFromJWT(h.Get(cfg.ClaimsJWTHeader)); sub != "" {
			return sub
		}
	}
	// Proxy that forwards the subject as a plain header value
	// (ALB identity header; oauth2-proxy X-Auth-Request-User).
	if cfg.SubjectHeader != "" {
		if sub := strings.TrimSpace(h.Get(cfg.SubjectHeader)); sub != "" {
			return sub
		}
	}
	// JWT (SDK/CLI) path: decode the forwarded Bearer token's subject. Proxy-agnostic.
	if token := bearerToken(h); token != "" {
		return subjectFromJWT(token)
	}
	return ""
}

// bearerToken returns the value of an Authorization: Bearer <token> header, or "".
func bearerToken(h http.Header) string {
	authz := h.Get(authorizationHeader)
	if len(authz) > len(bearerPrefix) && strings.EqualFold(authz[:len(bearerPrefix)], bearerPrefix) {
		return strings.TrimSpace(authz[len(bearerPrefix):])
	}
	return ""
}

// subjectFromJWT decodes a JWT's claims payload (without verifying the signature — the
// load balancer already validated it) and returns the `sub` claim. Returns "" on any
// malformed input or when no subject claim is present.
func subjectFromJWT(token string) string {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return ""
	}
	payload, err := decodeJWTSegment(parts[1])
	if err != nil {
		return ""
	}
	var c struct {
		Sub string `json:"sub"`
	}
	if err := json.Unmarshal(payload, &c); err != nil {
		return ""
	}
	return c.Sub
}

// decodeJWTSegment base64url-decodes a JWT segment, tolerating both the unpadded
// form (per the JWT spec) and the padded form some issuers — notably AWS ALB's
// x-amzn-oidc-data — emit. Without this, a payload whose length isn't a multiple of
// 4 fails strict RawURLEncoding and the subject claim is silently dropped.
func decodeJWTSegment(seg string) ([]byte, error) {
	if b, err := base64.RawURLEncoding.DecodeString(seg); err == nil {
		return b, nil
	}
	if pad := len(seg) % 4; pad != 0 {
		seg += strings.Repeat("=", 4-pad)
	}
	return base64.URLEncoding.DecodeString(seg)
}

// subjectOnlyIdentity builds a minimal EnrichedIdentity carrying just the subject. Used
// at read time when no user directory is available to resolve the subject's profile.
func subjectOnlyIdentity(subject string) *common.EnrichedIdentity {
	return &common.EnrichedIdentity{
		Principal: &common.EnrichedIdentity_User{
			User: &common.User{
				Id: &common.UserIdentifier{Subject: subject},
			},
		},
	}
}
