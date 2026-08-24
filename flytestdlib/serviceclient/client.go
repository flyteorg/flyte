// Package serviceclient builds HTTP clients for service-to-service requests.
package serviceclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const oauthAuthorizationServerMetadataPath = "/.well-known/oauth-authorization-server"

// ServiceConfig describes a downstream service and how to authenticate to it.
type ServiceConfig struct {
	URL  string     `json:"url" pflag:",Service base URL"`
	Auth AuthConfig `json:"auth" pflag:",Service authentication"`
}

// AuthType identifies the authentication mechanism used for service requests.
type AuthType string

const (
	// AuthTypeNone disables authentication.
	AuthTypeNone AuthType = ""
	// AuthTypeOAuth2ClientCredentials uses the OAuth 2.0 client credentials grant.
	AuthTypeOAuth2ClientCredentials AuthType = "OAuth2ClientCredentials"
)

// AuthConfig configures authentication for service requests.
type AuthConfig struct {
	Type             AuthType `json:"type" pflag:",Service authentication type"`
	IssuerURL        string   `json:"issuerUrl" pflag:",OAuth 2.0 authorization server issuer URL"`
	ClientID         string   `json:"clientId" pflag:",OAuth 2.0 client ID"`
	ClientSecret     string   `json:"clientSecret" pflag:",OAuth 2.0 client secret"`
	ClientSecretFile string   `json:"clientSecretFile" pflag:",File containing the OAuth 2.0 client secret"`
	Scopes           []string `json:"scopes" pflag:",OAuth 2.0 scopes to request"`
	Audience         string   `json:"audience" pflag:",OAuth 2.0 token audience"`
}

func (c AuthConfig) hasSettings() bool {
	return c.IssuerURL != "" || c.ClientID != "" || c.ClientSecret != "" ||
		c.ClientSecretFile != "" || len(c.Scopes) > 0 || c.Audience != ""
}

// NewHTTPClient returns an HTTP client configured for the service.
func NewHTTPClient(ctx context.Context, base *http.Client, cfg ServiceConfig) (*http.Client, error) {
	if base == nil {
		base = http.DefaultClient
	}
	if cfg.URL == "" {
		return nil, fmt.Errorf("service client: url is required")
	}
	switch cfg.Auth.Type {
	case AuthTypeNone:
		if cfg.Auth.hasSettings() {
			return nil, fmt.Errorf("service client auth: type is required when auth settings are configured")
		}
		return base, nil
	case AuthTypeOAuth2ClientCredentials:
		return newOAuth2ClientCredentialsClient(ctx, base, cfg.Auth)
	default:
		return nil, fmt.Errorf("service client auth: unsupported type %q", cfg.Auth.Type)
	}
}

func newOAuth2ClientCredentialsClient(ctx context.Context, base *http.Client, authCfg AuthConfig) (*http.Client, error) {
	if authCfg.IssuerURL == "" {
		return nil, fmt.Errorf("service client oauth2: issuerUrl is required")
	}
	if authCfg.ClientID == "" {
		return nil, fmt.Errorf("service client oauth2: clientId is required")
	}
	if authCfg.ClientSecret != "" && authCfg.ClientSecretFile != "" {
		return nil, fmt.Errorf("service client oauth2: clientSecret and clientSecretFile are mutually exclusive")
	}

	clientSecret := authCfg.ClientSecret
	if authCfg.ClientSecretFile != "" {
		raw, err := os.ReadFile(authCfg.ClientSecretFile) // #nosec G304 -- path is explicitly supplied by the service operator
		if err != nil {
			return nil, fmt.Errorf("service client oauth2: read clientSecretFile: %w", err)
		}
		clientSecret = strings.TrimSpace(string(raw))
	}
	if clientSecret == "" {
		return nil, fmt.Errorf("service client oauth2: clientSecret or clientSecretFile is required")
	}
	tokenURL, err := discoverTokenURL(ctx, base, authCfg.IssuerURL)
	if err != nil {
		return nil, err
	}

	endpointParams := url.Values{}
	if authCfg.Audience != "" {
		endpointParams.Set("audience", authCfg.Audience)
	}
	credentials := clientcredentials.Config{
		ClientID:       authCfg.ClientID,
		ClientSecret:   clientSecret,
		TokenURL:       tokenURL,
		Scopes:         authCfg.Scopes,
		EndpointParams: endpointParams,
	}

	baseTransport := base.Transport
	if baseTransport == nil {
		baseTransport = http.DefaultTransport
	}
	// Use the base transport for token requests as well as service requests. This
	// preserves custom TLS/proxy settings without recursively invoking auth.
	tokenHTTPClient := *base
	tokenHTTPClient.Transport = baseTransport
	tokenCtx := context.WithValue(ctx, oauth2.HTTPClient, &tokenHTTPClient)
	tokenSource := oauth2.ReuseTokenSource(nil, credentials.TokenSource(tokenCtx))

	authenticated := *base
	authenticated.Transport = &oauth2.Transport{
		Source: tokenSource,
		Base:   baseTransport,
	}
	return &authenticated, nil
}

func discoverTokenURL(ctx context.Context, httpClient *http.Client, issuerURL string) (string, error) {
	issuer, err := url.Parse(issuerURL)
	if err != nil || issuer.Scheme == "" || issuer.Host == "" {
		return "", fmt.Errorf("service client oauth2: invalid issuerUrl %q", issuerURL)
	}
	issuer.Path = strings.TrimRight(issuer.Path, "/") + oauthAuthorizationServerMetadataPath
	issuer.RawQuery = ""
	issuer.Fragment = ""

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, issuer.String(), nil)
	if err != nil {
		return "", fmt.Errorf("service client oauth2: create metadata request: %w", err)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("service client oauth2: discover token endpoint: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("service client oauth2: metadata endpoint returned status %d", resp.StatusCode)
	}
	var metadata struct {
		TokenEndpoint string `json:"token_endpoint"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return "", fmt.Errorf("service client oauth2: decode authorization server metadata: %w", err)
	}
	tokenEndpoint, err := url.Parse(metadata.TokenEndpoint)
	if err != nil || tokenEndpoint.Scheme == "" || tokenEndpoint.Host == "" {
		return "", fmt.Errorf("service client oauth2: metadata contains invalid token_endpoint %q", metadata.TokenEndpoint)
	}
	return metadata.TokenEndpoint, nil
}
