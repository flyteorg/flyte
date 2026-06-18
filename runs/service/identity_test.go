package service

import (
	"encoding/base64"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// jwt builds a syntactically valid (unsigned) JWT carrying the given claims payload.
func jwt(payloadJSON string) string {
	enc := func(s string) string { return base64.RawURLEncoding.EncodeToString([]byte(s)) }
	return enc(`{"alg":"RS256"}`) + "." + enc(payloadJSON) + ".sig"
}

func TestIdentityFromHeaders(t *testing.T) {
	tests := []struct {
		name                         string
		headers                      map[string]string
		wantNil                      bool
		wantSub, wantFirst, wantLast string
		wantEmail                    string
	}{
		{
			name:      "amzn oidc data: full claims (cookie path)",
			headers:   map[string]string{albDataHeader: jwt(`{"sub":"00u123","given_name":"Carina","family_name":"Didilescu","email":"carina@union.ai"}`)},
			wantSub:   "00u123",
			wantFirst: "Carina",
			wantLast:  "Didilescu",
			wantEmail: "carina@union.ai",
		},
		{
			name:      "amzn oidc data: subject + email only (no profile scope)",
			headers:   map[string]string{albDataHeader: jwt(`{"sub":"00u999","email":"a@b.com"}`)},
			wantSub:   "00u999",
			wantEmail: "a@b.com",
		},
		{
			name:    "amzn oidc identity header fallback (subject only)",
			headers: map[string]string{albIdentityHeader: "okta|sub-only"},
			wantSub: "okta|sub-only",
		},
		{
			name:      "bearer token claims (SDK path)",
			headers:   map[string]string{authorizationHeader: "Bearer " + jwt(`{"sub":"sdk-user","given_name":"Dev","email":"dev@union.ai"}`)},
			wantSub:   "sdk-user",
			wantFirst: "Dev",
			wantEmail: "dev@union.ai",
		},
		{
			name: "data header takes precedence over bearer",
			headers: map[string]string{
				albDataHeader:       jwt(`{"sub":"cookie-user"}`),
				authorizationHeader: "Bearer " + jwt(`{"sub":"bearer-user"}`),
			},
			wantSub: "cookie-user",
		},
		{name: "no auth headers", headers: map[string]string{}, wantNil: true},
		{name: "non-bearer authorization", headers: map[string]string{authorizationHeader: "Basic abc"}, wantNil: true},
		{name: "malformed jwt (two segments)", headers: map[string]string{albDataHeader: "a.b"}, wantNil: true},
		{name: "jwt without sub", headers: map[string]string{albDataHeader: jwt(`{"email":"a@b.com"}`)}, wantNil: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := http.Header{}
			for k, v := range tt.headers {
				h.Set(k, v)
			}
			id := identityFromHeaders(h)
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
	assert.Nil(t, subjectOnlyIdentity(""))

	id := subjectOnlyIdentity("user-123")
	assert.Equal(t, "user-123", id.GetUser().GetId().GetSubject())
	assert.Nil(t, id.GetUser().GetSpec())
	assert.Equal(t, "user-123", identitySubject(id))
	assert.Equal(t, "", identitySubject(nil))
}
