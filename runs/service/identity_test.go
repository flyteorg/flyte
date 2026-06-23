package service

import (
	"encoding/base64"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/v2/runs/config"
)

// albCfg is the default (AWS ALB) header configuration.
var albCfg = config.IdentityHeadersConfig{
	ClaimsJWTHeader: "X-Amzn-Oidc-Data",
	SubjectHeader:   "X-Amzn-Oidc-Identity",
}

// proxyCfg is a plain-value proxy (oauth2-proxy / Traefik forward-auth).
var proxyCfg = config.IdentityHeadersConfig{
	SubjectHeader: "X-Auth-Request-User",
}

// jwt builds a syntactically valid (unsigned) JWT carrying the given claims payload.
func jwt(payloadJSON string) string {
	enc := func(s string) string { return base64.RawURLEncoding.EncodeToString([]byte(s)) }
	return enc(`{"alg":"RS256"}`) + "." + enc(payloadJSON) + ".sig"
}

// jwtPadded builds a JWT whose payload uses padded base64url, as AWS ALB's
// x-amzn-oidc-data does. The decoder must tolerate the padding.
func jwtPadded(payloadJSON string) string {
	enc := func(s string) string { return base64.RawURLEncoding.EncodeToString([]byte(s)) }
	return enc(`{"alg":"ES256"}`) + "." + base64.URLEncoding.EncodeToString([]byte(payloadJSON)) + ".sig"
}

func TestSubjectFromJWT_ToleratesPadding(t *testing.T) {
	assert.Equal(t, "00u1", subjectFromJWT(jwtPadded(`{"sub":"00u1","email":"kevin@union.ai"}`)))
}

func TestSubjectFromHeaders(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.IdentityHeadersConfig
		headers map[string]string
		want    string
	}{
		{
			name:    "amzn oidc data JWT (cookie path)",
			cfg:     albCfg,
			headers: map[string]string{"X-Amzn-Oidc-Data": jwt(`{"sub":"00u123","email":"carina@union.ai"}`)},
			want:    "00u123",
		},
		{
			name:    "amzn oidc identity header fallback (subject only)",
			cfg:     albCfg,
			headers: map[string]string{"X-Amzn-Oidc-Identity": "okta|sub-only"},
			want:    "okta|sub-only",
		},
		{
			name:    "oauth2-proxy/traefik plain subject header",
			cfg:     proxyCfg,
			headers: map[string]string{"X-Auth-Request-User": "user-42"},
			want:    "user-42",
		},
		{
			name:    "bearer token subject (SDK path)",
			cfg:     albCfg,
			headers: map[string]string{authorizationHeader: "Bearer " + jwt(`{"sub":"sdk-user"}`)},
			want:    "sdk-user",
		},
		{
			name: "data header takes precedence over bearer",
			cfg:  albCfg,
			headers: map[string]string{
				"X-Amzn-Oidc-Data":  jwt(`{"sub":"cookie-user"}`),
				authorizationHeader: "Bearer " + jwt(`{"sub":"bearer-user"}`),
			},
			want: "cookie-user",
		},
		{name: "no auth headers", cfg: albCfg, headers: map[string]string{}, want: ""},
		{name: "non-bearer authorization", cfg: albCfg, headers: map[string]string{authorizationHeader: "Basic abc"}, want: ""},
		{name: "malformed jwt (two segments)", cfg: albCfg, headers: map[string]string{"X-Amzn-Oidc-Data": "a.b"}, want: ""},
		{name: "jwt without sub", cfg: albCfg, headers: map[string]string{"X-Amzn-Oidc-Data": jwt(`{"email":"a@b.com"}`)}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := http.Header{}
			for k, v := range tt.headers {
				h.Set(k, v)
			}
			assert.Equal(t, tt.want, subjectFromHeaders(h, tt.cfg))
		})
	}
}

func TestSubjectOnlyIdentity(t *testing.T) {
	id := subjectOnlyIdentity("user-123")
	assert.Equal(t, "user-123", id.GetUser().GetId().GetSubject())
	assert.Nil(t, id.GetUser().GetSpec())
}
