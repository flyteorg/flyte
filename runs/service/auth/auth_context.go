// Package auth contains types needed to start up a standalone OAuth2 Authorization Server or delegate
// authentication to an external provider. It supports OpenID Connect for user authentication.
package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/runs/service/auth/config"
)

const (
	IdpConnectionTimeout = 10 * time.Second
)

// AuthenticationContext holds all the utilities necessary to run authentication.
//
// The auth package supports two request flows, both producing an IdentityContext:
//
//	Browser (HTTP)                          API (gRPC / HTTP Bearer)
//	──────────────                          ───────────────────────
//	handlers.go                             interceptor.go
//	  /login  -> IdP redirect                 ▼
//	  /callback -> exchange code          token.go / resource_server.go
//	  /logout -> clear cookies                ▼
//	      │                               claims_verifier.go
//	      ▼                                   │
//	cookie_manager.go                         │
//	  (read/write encrypted cookies)          │
//	      │                                   │
//	      ▼                                   │
//	  cookie.go                               │
//	  (CSRF, secure cookie helpers)           │
//	      │                                   │
//	      ▼                                   ▼
//	  token.go ──────────────────────> identity_context.go
//	  (parse/validate JWT)              (UserID, AppID, Scopes, Claims)
//
type AuthenticationContext struct {
	oauth2Config      *oauth2.Config
	cookieManager     CookieManager
	oidcProvider      *oidc.Provider
	resourceServer    OAuth2ResourceServer
	cfg               config.Config
	httpClient        *http.Client
	oauth2MetadataURL *url.URL
	oidcMetadataURL   *url.URL
}

func (c *AuthenticationContext) OAuth2Config() *oauth2.Config         { return c.oauth2Config }
func (c *AuthenticationContext) CookieManager() CookieManager         { return c.cookieManager }
func (c *AuthenticationContext) OIDCProvider() *oidc.Provider         { return c.oidcProvider }
func (c *AuthenticationContext) ResourceServer() OAuth2ResourceServer { return c.resourceServer }
func (c *AuthenticationContext) Config() config.Config                { return c.cfg }
func (c *AuthenticationContext) HTTPClient() *http.Client             { return c.httpClient }
func (c *AuthenticationContext) OAuth2MetadataURL() *url.URL          { return c.oauth2MetadataURL }
func (c *AuthenticationContext) OIDCMetadataURL() *url.URL            { return c.oidcMetadataURL }

// NewAuthContext creates a new AuthContext with all the components needed for authentication.
func NewAuthContext(ctx context.Context, cfg config.Config, resourceServer OAuth2ResourceServer,
	hashKeyBase64, blockKeyBase64 string) (*AuthenticationContext, error) {

	cookieManager, err := NewCookieManager(ctx, hashKeyBase64, blockKeyBase64, cfg.UserAuth.CookieSetting)
	if err != nil {
		logger.Errorf(ctx, "Error creating cookie manager %s", err)
		return nil, fmt.Errorf("error creating cookie manager: %w", err)
	}

	httpClient := &http.Client{
		Timeout: IdpConnectionTimeout,
	}

	if len(cfg.UserAuth.HTTPProxyURL.String()) > 0 {
		logger.Infof(ctx, "HTTPProxy URL for OAuth2 is: %s", cfg.UserAuth.HTTPProxyURL.String())
		httpClient.Transport = &http.Transport{Proxy: http.ProxyURL(&cfg.UserAuth.HTTPProxyURL.URL)}
	}

	oidcCtx := oidc.ClientContext(ctx, httpClient)
	baseURL := cfg.UserAuth.OpenID.BaseURL.String()
	provider, err := oidc.NewProvider(oidcCtx, baseURL)
	if err != nil {
		return nil, fmt.Errorf("error creating oidc provider with issuer [%v]: %w", baseURL, err)
	}

	oauth2Config := &oauth2.Config{
		RedirectURL: "callback",
		ClientID:    cfg.UserAuth.OpenID.ClientID,
		Scopes:      cfg.UserAuth.OpenID.Scopes,
		Endpoint:    provider.Endpoint(),
	}

	oauth2MetadataURL, err := url.Parse(OAuth2MetadataEndpoint)
	if err != nil {
		return nil, fmt.Errorf("error parsing oauth2 metadata URL: %w", err)
	}

	oidcMetadataURL, err := url.Parse(OIdCMetadataEndpoint)
	if err != nil {
		return nil, fmt.Errorf("error parsing oidc metadata URL: %w", err)
	}

	return &AuthenticationContext{
		oauth2Config:      oauth2Config,
		cookieManager:     cookieManager,
		oidcProvider:      provider,
		resourceServer:    resourceServer,
		cfg:               cfg,
		httpClient:        httpClient,
		oauth2MetadataURL: oauth2MetadataURL,
		oidcMetadataURL:   oidcMetadataURL,
	}, nil
}

// HandlerConfig returns an AuthHandlerConfig suitable for use with RegisterHandlers.
func (c *AuthenticationContext) HandlerConfig() *AuthHandlerConfig {
	return &AuthHandlerConfig{
		CookieManager:  c.cookieManager,
		OAuth2Config:   c.oauth2Config,
		OIDCProvider:   c.oidcProvider,
		ResourceServer: c.resourceServer,
		AuthConfig:     c.cfg,
		HTTPClient:     c.httpClient,
	}
}
