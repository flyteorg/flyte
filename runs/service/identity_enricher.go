package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
)

const (
	userinfoHTTPTimeout = 3 * time.Second
	identityCacheTTL    = 10 * time.Minute
	oidcDiscoveryPath   = "/.well-known/openid-configuration"
)

// identityEnricher fills in a caller's profile (email, first/last name) by calling
// the OIDC userinfo endpoint with their access token. It is needed on the Bearer
// path, where the access token carries only the subject — the profile claims live
// in userinfo, not the token. Results are cached by subject. Every failure mode is
// best-effort: the caller's unenriched (subject-only) identity is returned instead.
type identityEnricher struct {
	authServerBaseURL string
	httpClient        *http.Client

	mu          sync.Mutex
	userinfoURL string // resolved lazily from OIDC discovery, then cached
	cache       map[string]cachedClaims
}

type cachedClaims struct {
	claims  *oidcClaims
	expires time.Time
}

// newIdentityEnricher returns an enricher for the given OAuth2 authorization-server
// base URL (e.g. https://signin.example.com/oauth2/default), or nil when unset —
// in which case enrich is a no-op and identities stay subject-only.
func newIdentityEnricher(authServerBaseURL string) *identityEnricher {
	if authServerBaseURL == "" {
		return nil
	}
	return &identityEnricher{
		authServerBaseURL: strings.TrimRight(authServerBaseURL, "/"),
		httpClient:        &http.Client{Timeout: userinfoHTTPTimeout},
		cache:             map[string]cachedClaims{},
	}
}

// enrich fills any profile fields (email, first/last name) missing from base with
// userinfo claims fetched using the access token. Fields already present on base
// (e.g. from x-amzn-oidc-data) are authoritative and kept. userinfo is queried only
// when the profile is incomplete and not cached. On a cache miss it makes a
// synchronous userinfo call (bounded by userinfoHTTPTimeout), which adds latency to
// run creation; on any error or timeout it returns base unchanged — enrichment is
// best-effort and never fails run creation.
func (e *identityEnricher) enrich(ctx context.Context, accessToken string, base *common.EnrichedIdentity) *common.EnrichedIdentity {
	if e == nil || base.GetUser() == nil {
		return base
	}
	subject := base.GetUser().GetId().GetSubject()
	if subject == "" {
		return base
	}
	// Serve a previously fetched set of claims without another userinfo round-trip.
	if cached := e.cachedFor(subject); cached != nil {
		return mergeClaims(base, cached)
	}
	if isCompleteProfile(base) || accessToken == "" {
		return base
	}
	claims, err := e.fetchUserinfo(ctx, accessToken)
	if err != nil {
		logger.Warnf(ctx, "identity enrichment: userinfo fetch failed for subject %q: %v", subject, err)
		return base
	}
	// Guard against token confusion / IdP misconfiguration: never associate a profile
	// fetched for a different subject with this run.
	if claims.Sub != "" && claims.Sub != subject {
		logger.Warnf(ctx, "identity enrichment: userinfo subject %q does not match caller %q; ignoring", claims.Sub, subject)
		return base
	}
	e.store(subject, claims)
	return mergeClaims(base, claims)
}

func (e *identityEnricher) cachedFor(subject string) *oidcClaims {
	e.mu.Lock()
	defer e.mu.Unlock()
	c, ok := e.cache[subject]
	if !ok {
		return nil
	}
	if time.Now().After(c.expires) {
		// Drop the stale entry so the map does not accumulate dead keys.
		delete(e.cache, subject)
		return nil
	}
	return c.claims
}

func (e *identityEnricher) store(subject string, claims *oidcClaims) {
	e.mu.Lock()
	defer e.mu.Unlock()
	// Opportunistically evict any other expired entries to bound the map size.
	now := time.Now()
	for k, c := range e.cache {
		if now.After(c.expires) {
			delete(e.cache, k)
		}
	}
	e.cache[subject] = cachedClaims{claims: claims, expires: time.Now().Add(identityCacheTTL)}
}

func (e *identityEnricher) fetchUserinfo(ctx context.Context, accessToken string) (*oidcClaims, error) {
	url, err := e.resolveUserinfoURL(ctx)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(authorizationHeader, bearerPrefix+accessToken)
	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo returned status %d", resp.StatusCode)
	}
	var c oidcClaims
	if err := json.NewDecoder(resp.Body).Decode(&c); err != nil {
		return nil, fmt.Errorf("decode userinfo: %w", err)
	}
	return &c, nil
}

// resolveUserinfoURL reads userinfo_endpoint from the OIDC discovery document once,
// then caches it for the life of the process.
func (e *identityEnricher) resolveUserinfoURL(ctx context.Context) (string, error) {
	e.mu.Lock()
	cached := e.userinfoURL
	e.mu.Unlock()
	if cached != "" {
		return cached, nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, e.authServerBaseURL+oidcDiscoveryPath, nil)
	if err != nil {
		return "", err
	}
	resp, err := e.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("oidc discovery returned status %d", resp.StatusCode)
	}
	var doc struct {
		UserinfoEndpoint string `json:"userinfo_endpoint"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return "", fmt.Errorf("decode oidc discovery: %w", err)
	}
	if doc.UserinfoEndpoint == "" {
		return "", fmt.Errorf("oidc discovery has no userinfo_endpoint")
	}
	e.mu.Lock()
	e.userinfoURL = doc.UserinfoEndpoint
	e.mu.Unlock()
	return doc.UserinfoEndpoint, nil
}

// isCompleteProfile reports whether the identity already carries every profile
// field, in which case no userinfo lookup is needed.
func isCompleteProfile(id *common.EnrichedIdentity) bool {
	s := id.GetUser().GetSpec()
	return s.GetFirstName() != "" && s.GetLastName() != "" && s.GetEmail() != ""
}

// mergeClaims fills only the profile fields missing from base with the fetched
// claims; fields already set on base are kept (header-provided values win).
func mergeClaims(base *common.EnrichedIdentity, c *oidcClaims) *common.EnrichedIdentity {
	if c == nil || base.GetUser() == nil {
		return base
	}
	s := base.GetUser().GetSpec()
	first, last, email := s.GetFirstName(), s.GetLastName(), s.GetEmail()
	if first == "" {
		first = c.GivenName
	}
	if last == "" {
		last = c.FamilyName
	}
	if email == "" {
		email = c.Email
	}
	if first != "" || last != "" || email != "" {
		base.GetUser().Spec = &common.UserSpec{FirstName: first, LastName: last, Email: email}
	}
	return base
}

// accessTokenFromHeaders returns the caller's Bearer access token (SDK/JWT path),
// which the load balancer has validated. The cookie path is not enriched this way:
// its forwarded access token is short-lived and already expired by request time, so
// it relies on the claims the proxy injects into x-amzn-oidc-data instead.
func accessTokenFromHeaders(h http.Header) string {
	if authz := h.Get(authorizationHeader); len(authz) > len(bearerPrefix) &&
		strings.EqualFold(authz[:len(bearerPrefix)], bearerPrefix) {
		return strings.TrimSpace(authz[len(bearerPrefix):])
	}
	return ""
}
