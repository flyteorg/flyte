package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/auth"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/runs/config"
)

// userInfoReq builds a UserInfoRequest carrying the given headers.
func userInfoReq(headers map[string]string) *connect.Request[auth.UserInfoRequest] {
	req := connect.NewRequest(&auth.UserInfoRequest{})
	for k, v := range headers {
		req.Header().Set(k, v)
	}
	return req
}

func TestIdentityService_UserInfo(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.IdentityHeadersConfig
		headers map[string]string

		wantCode connect.Code // non-zero means an error is expected
		wantSub  string
		wantName string
		want     *auth.UserInfoResponse
	}{
		{
			name: "alb claims JWT: full profile (cookie path)",
			cfg:  albCfg,
			headers: map[string]string{
				"X-Amzn-Oidc-Data": jwt(`{"sub":"00u3ipl8rr92YyW525d7","given_name":"Kevin","family_name":"Su","email":"kevin@union.ai"}`),
			},
			want: &auth.UserInfoResponse{
				Subject:    "00u3ipl8rr92YyW525d7",
				Name:       "Kevin Su",
				GivenName:  "Kevin",
				FamilyName: "Su",
				Email:      "kevin@union.ai",
			},
		},
		{
			name: "alb identity header: subject only",
			cfg:  albCfg,
			headers: map[string]string{
				"X-Amzn-Oidc-Identity": "0uav5j5qn0g1YptkJ5d7",
			},
			want: &auth.UserInfoResponse{Subject: "0uav5j5qn0g1YptkJ5d7"},
		},
		{
			name: "oauth2-proxy: subject and email as plain values",
			cfg:  proxyCfg,
			headers: map[string]string{
				"X-Auth-Request-User":  "00u456",
				"X-Auth-Request-Email": "carina@union.ai",
			},
			want: &auth.UserInfoResponse{Subject: "00u456", Email: "carina@union.ai"},
		},
		{
			name: "bearer token: claims decoded from the JWT (SDK/CLI path)",
			cfg:  albCfg,
			headers: map[string]string{
				"Authorization": "Bearer " + jwt(`{"sub":"00u789","given_name":"Ada","family_name":"Lovelace","email":"ada@union.ai"}`),
			},
			want: &auth.UserInfoResponse{
				Subject:    "00u789",
				Name:       "Ada Lovelace",
				GivenName:  "Ada",
				FamilyName: "Lovelace",
				Email:      "ada@union.ai",
			},
		},
		{
			name:     "no identity headers at all",
			cfg:      albCfg,
			headers:  nil,
			wantCode: connect.CodeUnauthenticated,
		},
		{
			name:     "malformed bearer token",
			cfg:      albCfg,
			headers:  map[string]string{"Authorization": "Bearer not-a-jwt"},
			wantCode: connect.CodeUnauthenticated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// No auth-server URL: the enricher is nil, so no userinfo round-trip.
			svc := NewIdentityService("", true, tt.cfg)

			resp, err := svc.UserInfo(context.Background(), userInfoReq(tt.headers))

			if tt.wantCode != 0 {
				require.Error(t, err)
				assert.Equal(t, tt.wantCode, connect.CodeOf(err))
				assert.Nil(t, resp)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tt.want.GetSubject(), resp.Msg.GetSubject())
			assert.Equal(t, tt.want.GetName(), resp.Msg.GetName())
			assert.Equal(t, tt.want.GetGivenName(), resp.Msg.GetGivenName())
			assert.Equal(t, tt.want.GetFamilyName(), resp.Msg.GetFamilyName())
			assert.Equal(t, tt.want.GetEmail(), resp.Msg.GetEmail())
		})
	}
}

// Identity must not be derived from headers the deployment says not to trust —
// otherwise any client could claim to be anyone by setting them itself.
func TestIdentityService_UserInfo_UntrustedHeadersRejected(t *testing.T) {
	svc := NewIdentityService("", false, albCfg)

	resp, err := svc.UserInfo(context.Background(), userInfoReq(map[string]string{
		"X-Amzn-Oidc-Data": jwt(`{"sub":"00u123","given_name":"Spoofed","family_name":"User","email":"attacker@example.com"}`),
	}))

	require.Error(t, err)
	assert.Equal(t, connect.CodeUnauthenticated, connect.CodeOf(err))
	assert.Nil(t, resp)
}

// The Bearer path carries only a subject; name/email come from the OIDC userinfo
// endpoint, and userinfo's subject is authoritative over the token's.
func TestIdentityService_UserInfo_EnrichesBearerViaUserinfo(t *testing.T) {
	srv := newUserinfoTestServer(t, `{"sub":"00u-real","given_name":"Kevin","family_name":"Su","email":"kevin@union.ai"}`)
	defer srv.Close()

	svc := NewIdentityService(srv.URL, true, albCfg)

	resp, err := svc.UserInfo(context.Background(), userInfoReq(map[string]string{
		"Authorization": "Bearer " + jwt(`{"sub":"0oa-client-id"}`),
	}))

	require.NoError(t, err)
	assert.Equal(t, "00u-real", resp.Msg.GetSubject(), "userinfo subject must win over the token's")
	assert.Equal(t, "Kevin Su", resp.Msg.GetName())
	assert.Equal(t, "kevin@union.ai", resp.Msg.GetEmail())
}

// A userinfo outage must not fail whoami — the subject-only identity still answers
// "who am I", so degrade rather than error.
func TestIdentityService_UserInfo_UserinfoFailureFallsBackToSubject(t *testing.T) {
	srv := newFailingUserinfoTestServer(t)
	defer srv.Close()

	svc := NewIdentityService(srv.URL, true, albCfg)

	resp, err := svc.UserInfo(context.Background(), userInfoReq(map[string]string{
		"Authorization": "Bearer " + jwt(`{"sub":"00u-token"}`),
	}))

	require.NoError(t, err)
	assert.Equal(t, "00u-token", resp.Msg.GetSubject())
	assert.Empty(t, resp.Msg.GetEmail())
}

func TestUserInfoFromUser_NameFromFirstLast(t *testing.T) {
	tests := []struct {
		name        string
		first, last string
		wantName    string
	}{
		{name: "both names", first: "Kevin", last: "Su", wantName: "Kevin Su"},
		{name: "first only", first: "Kevin", wantName: "Kevin"},
		{name: "last only", last: "Su", wantName: "Su"},
		{name: "neither", wantName: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := subjectOnlyIdentity("00u1")
			id.GetUser().Spec = &common.UserSpec{FirstName: tt.first, LastName: tt.last}
			got := userInfoFromUser(id.GetUser())
			assert.Equal(t, tt.wantName, got.GetName())
			assert.Equal(t, "00u1", got.GetSubject())
		})
	}
}

// Fields with no OIDC-claim source of their own still map through from UserSpec.
func TestUserInfoFromUser_MapsHandleAndPhoto(t *testing.T) {
	id := subjectOnlyIdentity("00u1")
	id.GetUser().Spec = &common.UserSpec{
		FirstName:  "Kevin",
		LastName:   "Su",
		Email:      "kevin@union.ai",
		UserHandle: "ksu",
		PhotoUrl:   "https://example.com/k.png",
	}

	got := userInfoFromUser(id.GetUser())

	assert.Equal(t, "ksu", got.GetPreferredUsername())
	assert.Equal(t, "https://example.com/k.png", got.GetPicture())
	assert.Equal(t, "kevin@union.ai", got.GetEmail())
}

// newUserinfoTestServer serves an OIDC discovery document pointing at a userinfo
// endpoint that returns the given claims JSON.
func newUserinfoTestServer(t *testing.T, claimsJSON string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	var base string
	mux.HandleFunc(oidcDiscoveryPath, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"userinfo_endpoint":"` + base + `/userinfo"}`))
	})
	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(claimsJSON))
	})
	srv := httptest.NewServer(mux)
	base = srv.URL
	return srv
}

// newFailingUserinfoTestServer serves discovery but errors on userinfo.
func newFailingUserinfoTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	var base string
	mux.HandleFunc(oidcDiscoveryPath, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"userinfo_endpoint":"` + base + `/userinfo"}`))
	})
	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	srv := httptest.NewServer(mux)
	base = srv.URL
	return srv
}
