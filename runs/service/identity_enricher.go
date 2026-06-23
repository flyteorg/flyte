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
	identityCacheTTL    = 24 * time.Hour
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

// enrich resolves the caller's full identity (subject, email, first/last name) from the
// OIDC userinfo endpoint on the Bearer path. base already carries a complete profile on
// the cookie path (claims are in x-amzn-oidc-data) or has no access token to query with —
// in those cases it is returned unchanged. Otherwise userinfo is queried (cached by the
// token's subject; a synchronous call bounded by userinfoHTTPTimeout that adds latency to
// run creation). On any error or timeout it returns base unchanged — enrichment is
// best-effort and never fails run creation.
//
// userinfo is the authoritative statement of who the access token belongs to, so its
// subject and profile are adopted wholesale: an OAuth access token's own `sub` is not
// guaranteed to be the user (e.g. Okta sets it to the client/app id), and trusting it
// would both lose the profile and attribute the run to a non-user principal — diverging
// from the cookie path, which records the real user. We therefore replace base's
// token-derived subject with userinfo's. (There is no token-confusion risk: userinfo is a
// synchronous call with the caller's own LB-validated access token.)
func (e *identityEnricher) enrich(ctx context.Context, accessToken string, base *common.EnrichedIdentity) *common.EnrichedIdentity {
	if e == nil || base.GetUser() == nil {
		return base
	}
	tokenSubject := base.GetUser().GetId().GetSubject()
	if tokenSubject == "" {
		return base
	}
	if isCompleteProfile(base) || accessToken == "" {
		return base
	}
	// Cached by the token subject — the same token always resolves to the same user.
	claims := e.cachedFor(tokenSubject)
	if claims == nil {
		fetched, err := e.fetchUserinfo(ctx, accessToken)
		if err != nil {
			logger.Warnf(ctx, "identity enrichment: userinfo fetch failed for subject %q: %v", tokenSubject, err)
			return base
		}
		if fetched.Sub == "" {
			logger.Warnf(ctx, "identity enrichment: userinfo returned no subject for caller %q; keeping subject-only", tokenSubject)
			return base
		}
		e.store(tokenSubject, fetched)
		claims = fetched
	}
	// Adopt userinfo's subject (authoritative) plus its profile fields.
	return mergeClaims(subjectOnlyIdentity(claims.Sub), claims)
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
	// Map is keyed by distinct user subject, so it's bounded by user count; cachedFor
	// evicts each stale entry on read. Add size-capped eviction if that ever fails.
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
