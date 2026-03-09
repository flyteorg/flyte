package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"connectrpc.com/connect"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/auth"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/auth/authconnect"
	"github.com/flyteorg/flyte/v2/runs/config"

	stdlibConfig "github.com/flyteorg/flyte/v2/flytestdlib/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

const (
	oauth2MetadataEndpoint = ".well-known/oauth-authorization-server"
	scopeAll               = "all"
)

type authMetadataService struct {
	authconnect.UnimplementedAuthMetadataServiceHandler
	cfg config.AuthConfig
}

func NewAuthMetadataService(cfg config.AuthConfig) authconnect.AuthMetadataServiceHandler {
	return &authMetadataService{cfg: cfg}
}

func (s *authMetadataService) GetOAuth2Metadata(
	ctx context.Context,
	_ *connect.Request[auth.GetOAuth2MetadataRequest],
) (*connect.Response[auth.GetOAuth2MetadataResponse], error) {
	switch s.cfg.AppAuth.AuthServerType {
	case config.AuthorizationServerTypeExternal:
		return s.getOAuth2MetadataExternal(ctx)
	default:
		return s.getOAuth2MetadataSelf()
	}
}

func (s *authMetadataService) getOAuth2MetadataSelf() (*connect.Response[auth.GetOAuth2MetadataResponse], error) {
	publicURL := getPublicURL(s.cfg.AuthorizedURIs)
	issuer := getIssuer(s.cfg, publicURL)

	base := strings.TrimRight(publicURL.String(), "/")

	resp := &auth.GetOAuth2MetadataResponse{
		Issuer:                            issuer,
		AuthorizationEndpoint:             base + "/oauth2/authorize",
		TokenEndpoint:                     base + "/oauth2/token",
		JwksUri:                           base + "/oauth2/jwks",
		CodeChallengeMethodsSupported:     []string{"S256"},
		ResponseTypesSupported:            []string{"code", "token", "code token"},
		GrantTypesSupported:               []string{"client_credentials", "refresh_token", "authorization_code"},
		ScopesSupported:                   []string{scopeAll},
		TokenEndpointAuthMethodsSupported: []string{"client_secret_basic"},
	}

	return connect.NewResponse(resp), nil
}

func (s *authMetadataService) getOAuth2MetadataExternal(ctx context.Context) (*connect.Response[auth.GetOAuth2MetadataResponse], error) {
	baseURL := s.cfg.AppAuth.ExternalAuthServer.BaseURL
	if baseURL.String() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("external auth server base URL is not configured"))
	}

	metadataURL := s.cfg.AppAuth.ExternalAuthServer.MetadataEndpointURL
	if metadataURL.String() == "" || metadataURL.String() == baseURL.String() {
		u := baseURL.URL
		u.Path = strings.TrimRight(u.Path, "/") + "/" + oauth2MetadataEndpoint
		metadataURL = stdlibConfig.URL{URL: u}
	}

	httpClient := &http.Client{}
	if s.cfg.HTTPProxyURL.String() != "" {
		httpClient.Transport = &http.Transport{
			Proxy: http.ProxyURL(&s.cfg.HTTPProxyURL.URL),
		}
	}

	retryAttempts := s.cfg.AppAuth.ExternalAuthServer.RetryAttempts
	if retryAttempts <= 0 {
		retryAttempts = 5
	}
	retryDelay := s.cfg.AppAuth.ExternalAuthServer.RetryDelay.Duration
	if retryDelay <= 0 {
		retryDelay = time.Second
	}

	body, err := sendAndRetryHTTPRequest(ctx, httpClient, metadataURL.String(), retryAttempts, retryDelay)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("failed to fetch OAuth2 metadata: %w", err))
	}

	resp := &auth.GetOAuth2MetadataResponse{}
	if err := json.Unmarshal(body, resp); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to unmarshal OAuth2 metadata: %w", err))
	}

	if s.cfg.TokenEndpointProxyConfig.Enabled && resp.TokenEndpoint != "" {
		proxyURL := s.cfg.TokenEndpointProxyConfig.PublicURL
		if proxyURL.String() == "" {
			proxyURL = stdlibConfig.URL{URL: *getPublicURL(s.cfg.AuthorizedURIs)}
		}
		rewritten := strings.TrimRight(proxyURL.String(), "/")
		if s.cfg.TokenEndpointProxyConfig.PathPrefix != "" {
			rewritten += "/" + strings.Trim(s.cfg.TokenEndpointProxyConfig.PathPrefix, "/")
		}

		// Preserve the original token endpoint path
		originalURL, parseErr := url.Parse(resp.TokenEndpoint)
		if parseErr == nil {
			rewritten += originalURL.Path
		}
		resp.TokenEndpoint = rewritten
	}

	return connect.NewResponse(resp), nil
}

func (s *authMetadataService) GetPublicClientConfig(
	_ context.Context,
	_ *connect.Request[auth.GetPublicClientConfigRequest],
) (*connect.Response[auth.GetPublicClientConfigResponse], error) {
	fc := s.cfg.AppAuth.ThirdParty.FlyteClientConfig
	return connect.NewResponse(&auth.GetPublicClientConfigResponse{
		ClientId:                 fc.ClientID,
		RedirectUri:              fc.RedirectURI,
		Scopes:                   fc.Scopes,
		AuthorizationMetadataKey: s.cfg.GrpcAuthorizationHeader,
		Audience:                 fc.Audience,
	}), nil
}

// getPublicURL returns the first AuthorizedURI as the public URL, or a default localhost URL.
func getPublicURL(authorizedURIs []stdlibConfig.URL) *url.URL {
	if len(authorizedURIs) > 0 {
		u := authorizedURIs[0].URL
		return &u
	}
	u, _ := url.Parse("http://localhost:8090")
	return u
}

// getIssuer returns the issuer from SelfAuthServer config, or falls back to public URL.
func getIssuer(cfg config.AuthConfig, publicURL *url.URL) string {
	if cfg.AppAuth.SelfAuthServer.Issuer != "" {
		return cfg.AppAuth.SelfAuthServer.Issuer
	}
	return strings.TrimRight(publicURL.String(), "/")
}

// sendAndRetryHTTPRequest fetches the given URL with retry logic.
func sendAndRetryHTTPRequest(ctx context.Context, client *http.Client, targetURL string, retryAttempts int, retryDelay time.Duration) ([]byte, error) {
	var lastErr error
	for i := 0; i < retryAttempts; i++ {
		if i > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryDelay):
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			logger.Warnf(ctx, "Failed to fetch %s (attempt %d/%d): %v", targetURL, i+1, retryAttempts, err)
			continue
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			logger.Warnf(ctx, "Failed to read response body from %s (attempt %d/%d): %v", targetURL, i+1, retryAttempts, readErr)
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			lastErr = fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, targetURL)
			logger.Warnf(ctx, "Unexpected status code from %s (attempt %d/%d): %d", targetURL, i+1, retryAttempts, resp.StatusCode)
			continue
		}

		return body, nil
	}

	return nil, fmt.Errorf("all %d attempts failed for %s: %w", retryAttempts, targetURL, lastErr)
}
