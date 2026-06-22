package service

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
)

const (
	// albDataHeader is the signed JWT of user claims set by ALB authenticate-oidc
	// (browser/cookie path). Its payload carries sub, email, given_name, family_name.
	albDataHeader = "X-Amzn-Oidc-Data"
	// albIdentityHeader is also set by ALB authenticate-oidc and carries the OIDC
	// subject (`sub`) directly — used as a fallback when the data header is absent.
	albIdentityHeader = "X-Amzn-Oidc-Identity"
	// authorizationHeader carries the Bearer token on the JWT-validation path
	// (SDK/CLI). The load balancer validates it and forwards it unchanged.
	authorizationHeader = "Authorization"
	bearerPrefix        = "Bearer "
)

// oidcClaims is the subset of OIDC claims we surface as the executing identity.
type oidcClaims struct {
	Sub        string `json:"sub"`
	Email      string `json:"email"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
}

// identityFromHeaders builds the EnrichedIdentity of the caller from the auth headers
// the load balancer forwards. Auth is enforced upstream (e.g. ALB OIDC / JWT
// validation), so the claims are trusted and only decoded here — not re-verified.
// Returns nil when no authenticated identity is present.
func identityFromHeaders(h http.Header) *common.EnrichedIdentity {
	// authenticate-oidc (browser/cookie) path: full claims in the signed data JWT.
	if id := identityFromJWT(h.Get(albDataHeader)); id != nil {
		return id
	}
	// Same path, subject only — when the data header is unavailable.
	if sub := strings.TrimSpace(h.Get(albIdentityHeader)); sub != "" {
		return subjectOnlyIdentity(sub)
	}
	// JWT (SDK/CLI) path: decode the forwarded Bearer token's claims.
	if token := bearerToken(h); token != "" {
		return identityFromJWT(token)
	}
	return nil
}

// bearerToken returns the value of an Authorization: Bearer <token> header, or "".
func bearerToken(h http.Header) string {
	authz := h.Get(authorizationHeader)
	if len(authz) > len(bearerPrefix) && strings.EqualFold(authz[:len(bearerPrefix)], bearerPrefix) {
		return strings.TrimSpace(authz[len(bearerPrefix):])
	}
	return ""
}

// identityFromJWT decodes a JWT's claims payload (without verifying the signature —
// the load balancer already validated it) into an EnrichedIdentity. Returns nil on
// any malformed input or when no subject claim is present.
func identityFromJWT(token string) *common.EnrichedIdentity {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil
	}
	payload, err := decodeJWTSegment(parts[1])
	if err != nil {
		return nil
	}
	var c oidcClaims
	if err := json.Unmarshal(payload, &c); err != nil || c.Sub == "" {
		return nil
	}
	return mergeClaims(subjectOnlyIdentity(c.Sub), &c)
}

// decodeJWTSegment base64url-decodes a JWT segment, tolerating both the unpadded
// form (per the JWT spec) and the padded form some issuers — notably AWS ALB's
// x-amzn-oidc-data — emit. Without this, a payload whose length isn't a multiple of
// 4 fails strict RawURLEncoding and the claims (email, name) are silently dropped.
func decodeJWTSegment(seg string) ([]byte, error) {
	if b, err := base64.RawURLEncoding.DecodeString(seg); err == nil {
		return b, nil
	}
	if pad := len(seg) % 4; pad != 0 {
		seg += strings.Repeat("=", 4-pad)
	}
	return base64.URLEncoding.DecodeString(seg)
}

// subjectOnlyIdentity builds a minimal EnrichedIdentity carrying just the subject.
// Mirrors the cloud transformer fallback; used when only the subject is available.
// Callers pass a non-empty subject.
func subjectOnlyIdentity(subject string) *common.EnrichedIdentity {
	return &common.EnrichedIdentity{
		Principal: &common.EnrichedIdentity_User{
			User: &common.User{
				Id: &common.UserIdentifier{Subject: subject},
			},
		},
	}
}
