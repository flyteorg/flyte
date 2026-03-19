package authzserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"

	"connectrpc.com/connect"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/auth"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/auth/authconnect"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/runs/service/auth/config"
	authpkg "github.com/flyteorg/flyte/v2/runs/service/auth"
)

const (
	oauth2MetadataEndpoint = ".well-known/oauth-authorization-server"
)

var (
	tokenRelativeURL     = mustParseURLPath("/oauth2/token")
	authorizeRelativeURL = mustParseURLPath("/oauth2/authorize")
	jsonWebKeysURL       = mustParseURLPath("/oauth2/jwks")
	oauth2MetadataRelURL = mustParseURLPath(oauth2MetadataEndpoint)

	supportedGrantTypes = []string{"client_credentials", "refresh_token", "authorization_code"}
)

func mustParseURLPath(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return u
}

type authMetadataService struct {
	authconnect.UnimplementedAuthMetadataServiceHandler
	cfg config.Config
}

func NewAuthMetadataService(cfg config.Config) authconnect.AuthMetadataServiceHandler {
	return &authMetadataService{cfg: cfg}
}

func (s *authMetadataService) GetOAuth2Metadata(
	ctx context.Context,
	_ *connect.Request[auth.GetOAuth2MetadataRequest],
) (*connect.Response[auth.GetOAuth2MetadataResponse], error) {
	switch s.cfg.AppAuth.AuthServerType {
	case config.AuthorizationServerTypeSelf:
		return s.getOAuth2MetadataSelf(ctx)
	default:
		return s.getOAuth2MetadataExternal(ctx)
	}
}

func (s *authMetadataService) getOAuth2MetadataSelf(ctx context.Context) (*connect.Response[auth.GetOAuth2MetadataResponse], error) {
	publicURL := authpkg.GetPublicURL(ctx, nil, s.cfg)

	resp := &auth.GetOAuth2MetadataResponse{
		Issuer:                            authpkg.GetIssuer(ctx, nil, s.cfg),
		AuthorizationEndpoint:             publicURL.ResolveReference(authorizeRelativeURL).String(),
		TokenEndpoint:                     publicURL.ResolveReference(tokenRelativeURL).String(),
		JwksUri:                           publicURL.ResolveReference(jsonWebKeysURL).String(),
		CodeChallengeMethodsSupported:     []string{"S256"},
		ResponseTypesSupported:            []string{"code", "token", "code token"},
		GrantTypesSupported:               supportedGrantTypes,
		ScopesSupported:                   []string{authpkg.ScopeAll},
		TokenEndpointAuthMethodsSupported: []string{"client_secret_basic"},
	}

	return connect.NewResponse(resp), nil
}

func (s *authMetadataService) getOAuth2MetadataExternal(ctx context.Context) (*connect.Response[auth.GetOAuth2MetadataResponse], error) {
	baseURL := s.cfg.AppAuth.ExternalAuthServer.BaseURL
	if baseURL.String() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("external auth server base URL is not configured"))
	}

	// issuer urls, conventionally, do not end with a '/', however, metadata urls are usually relative of those.
	// This adds a '/' to ensure ResolveReference behaves intuitively.
	baseURL.Path = strings.TrimSuffix(baseURL.Path, "/") + "/"

	var externalMetadataURL *url.URL
	if len(s.cfg.AppAuth.ExternalAuthServer.MetadataEndpointURL.String()) > 0 {
		externalMetadataURL = baseURL.ResolveReference(&s.cfg.AppAuth.ExternalAuthServer.MetadataEndpointURL.URL)
	} else {
		externalMetadataURL = baseURL.ResolveReference(oauth2MetadataRelURL)
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

	response, err := sendAndRetryHTTPRequest(ctx, httpClient, externalMetadataURL.String(), retryAttempts, retryDelay)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, fmt.Errorf("failed to fetch OAuth2 metadata: %w", err))
	}

	raw, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to read OAuth2 metadata response: %w", err))
	}

	resp := &auth.GetOAuth2MetadataResponse{}
	if err := unmarshalResp(response, raw, resp); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to unmarshal OAuth2 metadata: %w", err))
	}

	tokenProxyConfig := s.cfg.TokenEndpointProxyConfig
	if tokenProxyConfig.Enabled {
		tokenEndpoint, parseErr := url.Parse(resp.TokenEndpoint)
		if parseErr != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to parse token endpoint [%v], err: %v", resp.TokenEndpoint, parseErr))
		}
		if len(tokenProxyConfig.PublicURL.Host) == 0 {
			publicURL := authpkg.GetPublicURL(ctx, nil, s.cfg)
			tokenProxyConfig.PublicURL = config.URL{URL: *publicURL}
		}
		tokenEndpoint.Host = tokenProxyConfig.PublicURL.Host
		tokenEndpoint.Path = tokenProxyConfig.PathPrefix + tokenEndpoint.Path
		tokenEndpoint.RawPath = tokenProxyConfig.PathPrefix + tokenEndpoint.RawPath
		resp.TokenEndpoint = tokenEndpoint.String()
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

// unmarshalResp unmarshals a JSON response body, providing a detailed error if the Content-Type is unexpected.
func unmarshalResp(r *http.Response, body []byte, v interface{}) error {
	err := json.Unmarshal(body, &v)
	if err == nil {
		return nil
	}
	ct := r.Header.Get("Content-Type")
	mediaType, _, parseErr := mime.ParseMediaType(ct)
	if parseErr == nil && mediaType == "application/json" {
		return fmt.Errorf("got Content-Type = application/json, but could not unmarshal as JSON: %v", err)
	}
	return fmt.Errorf("expected Content-Type = application/json, got %q: %v", ct, err)
}

// sendAndRetryHTTPRequest fetches the given URL with retry logic.
func sendAndRetryHTTPRequest(ctx context.Context, client *http.Client, targetURL string, retryAttempts int, retryDelay time.Duration) (*http.Response, error) {
	var lastErr error
	var lastResp *http.Response
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

		if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
			return resp, nil
		}

		lastErr = fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, targetURL)
		lastResp = resp
		logger.Warnf(ctx, "Unexpected status code from %s (attempt %d/%d): %d", targetURL, i+1, retryAttempts, resp.StatusCode)
	}

	if lastResp != nil && lastResp.StatusCode != http.StatusOK {
		return lastResp, fmt.Errorf("failed to get oauth metadata with status code %v: %w", lastResp.StatusCode, lastErr)
	}

	return nil, fmt.Errorf("all %d attempts failed for %s: %w", retryAttempts, targetURL, lastErr)
}
