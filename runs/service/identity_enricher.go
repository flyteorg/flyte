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
	albAccessTokenHdr   = "X-Amzn-Oidc-Accesstoken"
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
	cache       map[string]cachedIdentity
}

type cachedIdentity struct {
	id      *common.EnrichedIdentity
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
		cache:             map[string]cachedIdentity{},
	}
}

// enrich augments base with profile claims fetched from userinfo when base lacks
// them and an access token is available. base is returned unchanged on any miss,
// cache hit without profile, or error — enrichment never blocks or fails run creation.
func (e *identityEnricher) enrich(ctx context.Context, accessToken string, base *common.EnrichedIdentity) *common.EnrichedIdentity {
	if e == nil || base.GetUser() == nil || hasProfile(base) || accessToken == "" {
		return base
	}
	subject := base.GetUser().GetId().GetSubject()
	if cached := e.cachedFor(subject); cached != nil {
		return cached
	}
	claims, err := e.fetchUserinfo(ctx, accessToken)
	if err != nil {
		logger.Warnf(ctx, "identity enrichment: userinfo fetch failed for subject %q: %v", subject, err)
		return base
	}
	enriched := mergeClaims(base, claims)
	e.store(subject, enriched)
	return enriched
}

func (e *identityEnricher) cachedFor(subject string) *common.EnrichedIdentity {
	e.mu.Lock()
	defer e.mu.Unlock()
	if c, ok := e.cache[subject]; ok && time.Now().Before(c.expires) {
		return c.id
	}
	return nil
}

func (e *identityEnricher) store(subject string, id *common.EnrichedIdentity) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.cache[subject] = cachedIdentity{id: id, expires: time.Now().Add(identityCacheTTL)}
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

// hasProfile reports whether the identity already carries any profile field.
func hasProfile(id *common.EnrichedIdentity) bool {
	s := id.GetUser().GetSpec()
	return s.GetEmail() != "" || s.GetFirstName() != "" || s.GetLastName() != ""
}

// mergeClaims sets base's user spec from the fetched claims when any are present.
func mergeClaims(base *common.EnrichedIdentity, c *oidcClaims) *common.EnrichedIdentity {
	if c == nil || base.GetUser() == nil {
		return base
	}
	if c.Email != "" || c.GivenName != "" || c.FamilyName != "" {
		base.GetUser().Spec = &common.UserSpec{
			FirstName: c.GivenName,
			LastName:  c.FamilyName,
			Email:     c.Email,
		}
	}
	return base
}

// accessTokenFromHeaders returns the caller's access token: the forwarded Bearer
// token (SDK/JWT path) or the ALB-provided access token (cookie path).
func accessTokenFromHeaders(h http.Header) string {
	if authz := h.Get(authorizationHeader); len(authz) > len(bearerPrefix) &&
		strings.EqualFold(authz[:len(bearerPrefix)], bearerPrefix) {
		return strings.TrimSpace(authz[len(bearerPrefix):])
	}
	return strings.TrimSpace(h.Get(albAccessTokenHdr))
}
