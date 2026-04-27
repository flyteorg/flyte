package authzserver

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	jwtgo "github.com/golang-jwt/jwt/v5"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/auth"
	"github.com/flyteorg/flyte/v2/runs/service/auth/config"
	authpkg "github.com/flyteorg/flyte/v2/runs/service/auth"
)

// ResourceServer authorizes access requests issued by an external Authorization Server.
type ResourceServer struct {
	signatureVerifier oidc.KeySet
	allowedAudience   []string
}

// NewOAuth2ResourceServer initializes a new OAuth2ResourceServer.
func NewOAuth2ResourceServer(ctx context.Context, cfg config.ExternalAuthorizationServer, fallbackBaseURL config.URL) (*ResourceServer, error) {
	u := cfg.BaseURL
	if len(u.String()) == 0 {
		u = fallbackBaseURL
	}

	verifier, err := getJwksForIssuer(ctx, u.URL, cfg)
	if err != nil {
		return nil, err
	}

	return &ResourceServer{
		signatureVerifier: verifier,
		allowedAudience:   cfg.AllowedAudience,
	}, nil
}

// ValidateAccessToken verifies the token signature, validates claims, and returns the identity context.
func (r *ResourceServer) ValidateAccessToken(ctx context.Context, expectedAudience, tokenStr string) (*authpkg.IdentityContext, error) {
	_, err := r.signatureVerifier.VerifySignature(ctx, tokenStr)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token signature: %w", err)
	}

	claims := jwtgo.MapClaims{}
	parser := jwtgo.NewParser()
	if _, _, err = parser.ParseUnverified(tokenStr, claims); err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	allowed := make(map[string]bool, len(r.allowedAudience)+1)
	for _, a := range r.allowedAudience {
		allowed[a] = true
	}
	allowed[expectedAudience] = true

	return verifyClaims(allowed, claims)
}

// getJwksForIssuer fetches the OAuth2 metadata from the external auth server and returns the remote JWKS key set.
func getJwksForIssuer(ctx context.Context, issuerBaseURL url.URL, cfg config.ExternalAuthorizationServer) (oidc.KeySet, error) {
	issuerBaseURL.Path = strings.TrimSuffix(issuerBaseURL.Path, "/") + "/"

	var wellKnown *url.URL
	if len(cfg.MetadataEndpointURL.String()) > 0 {
		wellKnown = issuerBaseURL.ResolveReference(&cfg.MetadataEndpointURL.URL)
	} else {
		wellKnown = issuerBaseURL.ResolveReference(oauth2MetadataRelURL)
	}

	httpClient := &http.Client{}
	if len(cfg.HTTPProxyURL.String()) > 0 {
		httpClient.Transport = &http.Transport{
			Proxy: http.ProxyURL(&cfg.HTTPProxyURL.URL),
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, wellKnown.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", resp.Status, body)
	}

	p := &auth.GetOAuth2MetadataResponse{}
	if err = unmarshalResp(resp, body, p); err != nil {
		return nil, fmt.Errorf("failed to decode provider discovery object: %w", err)
	}

	return oidc.NewRemoteKeySet(oidc.ClientContext(ctx, httpClient), p.JwksUri), nil
}
