package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
)

// newTestIdP returns an httptest server exposing OIDC discovery + userinfo, plus a
// counter of how many times userinfo is hit.
func newTestIdP(t *testing.T, userinfo string, status int) (*httptest.Server, *int32) {
	t.Helper()
	var userinfoHits int32
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"userinfo_endpoint":"` + srv.URL + `/v1/userinfo"}`))
	})
	mux.HandleFunc("/v1/userinfo", func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&userinfoHits, 1)
		assert.Equal(t, "Bearer access-tok", r.Header.Get("Authorization"))
		w.WriteHeader(status)
		_, _ = w.Write([]byte(userinfo))
	})
	t.Cleanup(srv.Close)
	return srv, &userinfoHits
}

func TestEnrich_FillsProfileFromUserinfo(t *testing.T) {
	srv, userinfoHits := newTestIdP(t, `{"sub":"00u1","given_name":"Carina","family_name":"Didilescu","email":"carina@union.ai"}`, http.StatusOK)
	e := newIdentityEnricher(srv.URL)

	got := e.enrich(context.Background(), "access-tok", subjectOnlyIdentity("00u1"))
	spec := got.GetUser().GetSpec()
	assert.Equal(t, "Carina", spec.GetFirstName())
	assert.Equal(t, "Didilescu", spec.GetLastName())
	assert.Equal(t, "carina@union.ai", spec.GetEmail())
	assert.Equal(t, "00u1", got.GetUser().GetId().GetSubject())

	// Second call for the same subject is served from cache — no extra userinfo hit.
	e.enrich(context.Background(), "access-tok", subjectOnlyIdentity("00u1"))
	assert.Equal(t, int32(1), atomic.LoadInt32(userinfoHits))
}

func TestEnrich_NoOpCases(t *testing.T) {
	srv, userinfoHits := newTestIdP(t, `{"sub":"x","email":"e@x.com"}`, http.StatusOK)
	e := newIdentityEnricher(srv.URL)

	// nil enricher returns base unchanged.
	var nilE *identityEnricher
	base := subjectOnlyIdentity("s")
	assert.Equal(t, base, nilE.enrich(context.Background(), "tok", base))

	// no access token: can't fetch, skip enrichment.
	e.enrich(context.Background(), "", subjectOnlyIdentity("s"))

	// already a complete profile: nothing missing, skip enrichment.
	complete := subjectOnlyIdentity("s")
	complete.GetUser().Spec = &common.UserSpec{FirstName: "A", LastName: "B", Email: "a@b.com"}
	e.enrich(context.Background(), "tok", complete)

	assert.Equal(t, int32(0), atomic.LoadInt32(userinfoHits))
}

func TestEnrich_FillsOnlyMissingFields(t *testing.T) {
	// userinfo returns names (no email); base already has an email from the cookie header.
	srv, _ := newTestIdP(t, `{"sub":"00u3","given_name":"Kevin","family_name":"Su"}`, http.StatusOK)
	e := newIdentityEnricher(srv.URL)

	base := subjectOnlyIdentity("00u3")
	base.GetUser().Spec = &common.UserSpec{Email: "kevin@union.ai"}

	got := e.enrich(context.Background(), "access-tok", base)
	spec := got.GetUser().GetSpec()
	assert.Equal(t, "Kevin", spec.GetFirstName())
	assert.Equal(t, "Su", spec.GetLastName())
	assert.Equal(t, "kevin@union.ai", spec.GetEmail()) // header email preserved
}

func TestEnrich_UserinfoErrorFallsBackToBase(t *testing.T) {
	srv, _ := newTestIdP(t, `nope`, http.StatusUnauthorized)
	e := newIdentityEnricher(srv.URL)

	got := e.enrich(context.Background(), "access-tok", subjectOnlyIdentity("00u2"))
	assert.Nil(t, got.GetUser().GetSpec())
	assert.Equal(t, "00u2", got.GetUser().GetId().GetSubject())
}

func TestNewIdentityEnricher_EmptyURL(t *testing.T) {
	assert.Nil(t, newIdentityEnricher(""))
}

func TestAccessTokenFromHeaders(t *testing.T) {
	h := http.Header{}
	h.Set("Authorization", "Bearer abc")
	assert.Equal(t, "abc", accessTokenFromHeaders(h))

	h = http.Header{}
	h.Set(albAccessTokenHdr, "alb-tok")
	assert.Equal(t, "alb-tok", accessTokenFromHeaders(h))

	assert.Equal(t, "", accessTokenFromHeaders(http.Header{}))
}
