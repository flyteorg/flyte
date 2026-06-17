package service

import (
	"encoding/base64"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// bearerToken builds a syntactically valid (unsigned) JWT carrying the given sub claim.
func bearerToken(payloadJSON string) string {
	enc := func(s string) string { return base64.RawURLEncoding.EncodeToString([]byte(s)) }
	return "Bearer " + enc(`{"alg":"RS256"}`) + "." + enc(payloadJSON) + ".sig"
}

func TestSubjectFromHeaders(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
		want    string
	}{
		{
			name:    "amzn oidc identity header (cookie path)",
			headers: map[string]string{albIdentityHeader: "okta|user-123"},
			want:    "okta|user-123",
		},
		{
			name:    "amzn oidc identity header is trimmed",
			headers: map[string]string{albIdentityHeader: "  user-456  "},
			want:    "user-456",
		},
		{
			name:    "bearer token sub claim (jwt path)",
			headers: map[string]string{authorizationHeader: bearerToken(`{"sub":"sdk-user-789","email":"a@b.com"}`)},
			want:    "sdk-user-789",
		},
		{
			name: "amzn header takes precedence over bearer",
			headers: map[string]string{
				albIdentityHeader:   "cookie-user",
				authorizationHeader: bearerToken(`{"sub":"bearer-user"}`),
			},
			want: "cookie-user",
		},
		{name: "no auth headers", headers: map[string]string{}, want: ""},
		{name: "non-bearer authorization", headers: map[string]string{authorizationHeader: "Basic abc"}, want: ""},
		{name: "malformed bearer (two segments)", headers: map[string]string{authorizationHeader: "Bearer a.b"}, want: ""},
		{name: "bearer without sub", headers: map[string]string{authorizationHeader: bearerToken(`{"email":"a@b.com"}`)}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := http.Header{}
			for k, v := range tt.headers {
				h.Set(k, v)
			}
			assert.Equal(t, tt.want, subjectFromHeaders(h))
		})
	}
}

func TestSubjectOnlyIdentity(t *testing.T) {
	assert.Nil(t, subjectOnlyIdentity(""))

	id := subjectOnlyIdentity("user-123")
	assert.Equal(t, "user-123", id.GetUser().GetId().GetSubject())
}
