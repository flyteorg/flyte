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
	EmailHeader:   "X-Auth-Request-Email",
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

func TestIdentityFromJWT_ToleratesPadding(t *testing.T) {
	id := identityFromJWT(jwtPadded(`{"sub":"00u1","given_name":"Kevin","family_name":"Su","email":"kevin@union.ai"}`))
	assert.Equal(t, "00u1", id.GetUser().GetId().GetSubject())
	assert.Equal(t, "Kevin", id.GetUser().GetSpec().GetFirstName())
	assert.Equal(t, "Su", id.GetUser().GetSpec().GetLastName())
	assert.Equal(t, "kevin@union.ai", id.GetUser().GetSpec().GetEmail())
}

func TestIdentityFromHeaders(t *testing.T) {
	tests := []struct {
		name                         string
		cfg                          config.IdentityHeadersConfig
		headers                      map[string]string
		wantNil                      bool
		wantSub, wantFirst, wantLast string
		wantEmail                    string
	}{
		{
			name:      "amzn oidc data: full claims (cookie path)",
			cfg:       albCfg,
			headers:   map[string]string{"X-Amzn-Oidc-Data": jwt(`{"sub":"00u123","given_name":"Carina","family_name":"Didilescu","email":"carina@union.ai"}`)},
			wantSub:   "00u123",
			wantFirst: "Carina",
			wantLast:  "Didilescu",
			wantEmail: "carina@union.ai",
		},
		{
			name:      "amzn oidc data: subject + email only (no profile scope)",
			cfg:       albCfg,
			headers:   map[string]string{"X-Amzn-Oidc-Data": jwt(`{"sub":"00u999","email":"a@b.com"}`)},
			wantSub:   "00u999",
			wantEmail: "a@b.com",
		},
		{
			name:    "amzn oidc identity header fallback (subject only)",
			cfg:     albCfg,
			headers: map[string]string{"X-Amzn-Oidc-Identity": "okta|sub-only"},
			wantSub: "okta|sub-only",
		},
		{
			name:      "oauth2-proxy/traefik plain headers (subject + email)",
			cfg:       proxyCfg,
			headers:   map[string]string{"X-Auth-Request-User": "user-42", "X-Auth-Request-Email": "u42@union.ai"},
			wantSub:   "user-42",
			wantEmail: "u42@union.ai",
		},
		{
			name:      "bearer token claims (SDK path)",
			cfg:       albCfg,
			headers:   map[string]string{authorizationHeader: "Bearer " + jwt(`{"sub":"sdk-user","given_name":"Dev","email":"dev@union.ai"}`)},
			wantSub:   "sdk-user",
			wantFirst: "Dev",
			wantEmail: "dev@union.ai",
		},
		{
			name: "data header takes precedence over bearer",
			cfg:  albCfg,
			headers: map[string]string{
				"X-Amzn-Oidc-Data":  jwt(`{"sub":"cookie-user"}`),
				authorizationHeader: "Bearer " + jwt(`{"sub":"bearer-user"}`),
			},
			wantSub: "cookie-user",
		},
		{name: "no auth headers", cfg: albCfg, headers: map[string]string{}, wantNil: true},
		{name: "non-bearer authorization", cfg: albCfg, headers: map[string]string{authorizationHeader: "Basic abc"}, wantNil: true},
		{name: "malformed jwt (two segments)", cfg: albCfg, headers: map[string]string{"X-Amzn-Oidc-Data": "a.b"}, wantNil: true},
		{name: "jwt without sub", cfg: albCfg, headers: map[string]string{"X-Amzn-Oidc-Data": jwt(`{"email":"a@b.com"}`)}, wantNil: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := http.Header{}
			for k, v := range tt.headers {
				h.Set(k, v)
			}
			id := identityFromHeaders(h, tt.cfg)
			if tt.wantNil {
				assert.Nil(t, id)
				return
			}
			assert.Equal(t, tt.wantSub, id.GetUser().GetId().GetSubject())
			assert.Equal(t, tt.wantFirst, id.GetUser().GetSpec().GetFirstName())
			assert.Equal(t, tt.wantLast, id.GetUser().GetSpec().GetLastName())
			assert.Equal(t, tt.wantEmail, id.GetUser().GetSpec().GetEmail())
		})
	}
}

func TestSubjectOnlyIdentity(t *testing.T) {
	id := subjectOnlyIdentity("user-123")
	assert.Equal(t, "user-123", id.GetUser().GetId().GetSubject())
	assert.Nil(t, id.GetUser().GetSpec())
}
