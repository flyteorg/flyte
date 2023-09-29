package admin

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"sync"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/flyteorg/flyteidl/clients/go/admin/cache"
	"github.com/flyteorg/flyteidl/clients/go/admin/deviceflow"
	"github.com/flyteorg/flyteidl/clients/go/admin/externalprocess"
	"github.com/flyteorg/flyteidl/clients/go/admin/pkce"
	"github.com/flyteorg/flyteidl/clients/go/admin/tokenorchestrator"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flytestdlib/logger"
)

//go:generate mockery -name TokenSource
type TokenSource interface {
	Token() (*oauth2.Token, error)
}

const (
	audienceKey = "audience"
)

// TokenSourceProvider defines the interface needed to provide a TokenSource that is used to
// create a client with authentication enabled.
type TokenSourceProvider interface {
	GetTokenSource(ctx context.Context) (oauth2.TokenSource, error)
}

func NewTokenSourceProvider(ctx context.Context, cfg *Config, tokenCache cache.TokenCache,
	authClient service.AuthMetadataServiceClient) (TokenSourceProvider, error) {

	var tokenProvider TokenSourceProvider
	var err error
	switch cfg.AuthType {
	case AuthTypeClientSecret:
		tokenURL := cfg.TokenURL
		if len(tokenURL) == 0 {
			metadata, err := authClient.GetOAuth2Metadata(ctx, &service.OAuth2MetadataRequest{})
			if err != nil {
				return nil, fmt.Errorf("failed to fetch auth metadata. Error: %v", err)
			}

			tokenURL = metadata.TokenEndpoint
		}

		scopes := cfg.Scopes
		audienceValue := cfg.Audience

		if len(scopes) == 0 || cfg.UseAudienceFromAdmin {
			publicClientConfig, err := authClient.GetPublicClientConfig(ctx, &service.PublicClientAuthConfigRequest{})
			if err != nil {
				return nil, fmt.Errorf("failed to fetch client metadata. Error: %v", err)
			}
			// Update scopes from publicClientConfig
			if len(scopes) == 0 {
				scopes = publicClientConfig.Scopes
			}
			// Update audience from publicClientConfig
			if cfg.UseAudienceFromAdmin {
				audienceValue = publicClientConfig.Audience
			}
		}

		tokenProvider, err = NewClientCredentialsTokenSourceProvider(ctx, cfg, scopes, tokenURL, tokenCache, audienceValue)
		if err != nil {
			return nil, err
		}
	case AuthTypePkce:
		baseTokenOrchestrator, err := tokenorchestrator.NewBaseTokenOrchestrator(ctx, tokenCache, authClient)
		if err != nil {
			return nil, err
		}

		tokenProvider, err = NewPKCETokenSourceProvider(baseTokenOrchestrator, cfg.PkceConfig)
		if err != nil {
			return nil, err
		}
	case AuthTypeExternalCommand:
		tokenProvider, err = NewExternalTokenSourceProvider(cfg.Command)
		if err != nil {
			return nil, err
		}
	case AuthTypeDeviceFlow:
		baseTokenOrchestrator, err := tokenorchestrator.NewBaseTokenOrchestrator(ctx, tokenCache, authClient)
		if err != nil {
			return nil, err
		}

		tokenProvider, err = NewDeviceFlowTokenSourceProvider(baseTokenOrchestrator, cfg.DeviceFlowConfig)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported type %v", cfg.AuthType)
	}

	return tokenProvider, nil
}

type ExternalTokenSourceProvider struct {
	command []string
}

func NewExternalTokenSourceProvider(command []string) (TokenSourceProvider, error) {
	return &ExternalTokenSourceProvider{command: command}, nil
}

func (e ExternalTokenSourceProvider) GetTokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	output, err := externalprocess.Execute(e.command)
	if err != nil {
		return nil, err
	}

	return oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: strings.Trim(string(output), "\t \n"),
		TokenType:   "bearer",
	}), nil
}

type PKCETokenSourceProvider struct {
	tokenOrchestrator pkce.TokenOrchestrator
}

func NewPKCETokenSourceProvider(baseTokenOrchestrator tokenorchestrator.BaseTokenOrchestrator, pkceCfg pkce.Config) (TokenSourceProvider, error) {
	tokenOrchestrator, err := pkce.NewTokenOrchestrator(baseTokenOrchestrator, pkceCfg)
	if err != nil {
		return nil, err
	}
	return PKCETokenSourceProvider{tokenOrchestrator: tokenOrchestrator}, nil
}

func (p PKCETokenSourceProvider) GetTokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	return GetPKCEAuthTokenSource(ctx, p.tokenOrchestrator)
}

// Returns the token source which would be used for three legged oauth. eg : for admin to authorize access to flytectl
func GetPKCEAuthTokenSource(ctx context.Context, pkceTokenOrchestrator pkce.TokenOrchestrator) (oauth2.TokenSource, error) {
	// explicitly ignore error while fetching token from cache.
	authToken, err := pkceTokenOrchestrator.FetchTokenFromCacheOrRefreshIt(ctx, pkceTokenOrchestrator.Config.BrowserSessionTimeout)
	if err != nil {
		logger.Warnf(ctx, "Failed fetching from cache. Will restart the flow. Error: %v", err)
	}

	if authToken == nil {
		// Fetch using auth flow
		if authToken, err = pkceTokenOrchestrator.FetchTokenFromAuthFlow(ctx); err != nil {
			logger.Errorf(ctx, "Error fetching token using auth flow due to %v", err)
			return nil, err
		}
	}

	return &pkce.SimpleTokenSource{
		CachedToken: authToken,
	}, nil
}

type ClientCredentialsTokenSourceProvider struct {
	ccConfig   clientcredentials.Config
	tokenCache cache.TokenCache
}

func NewClientCredentialsTokenSourceProvider(ctx context.Context, cfg *Config, scopes []string, tokenURL string,
	tokenCache cache.TokenCache, audience string) (TokenSourceProvider, error) {
	var secret string
	if len(cfg.ClientSecretEnvVar) > 0 {
		secret = os.Getenv(cfg.ClientSecretEnvVar)
	} else if len(cfg.ClientSecretLocation) > 0 {
		secretBytes, err := ioutil.ReadFile(cfg.ClientSecretLocation)
		if err != nil {
			logger.Errorf(ctx, "Error reading secret from location %s", cfg.ClientSecretLocation)
			return nil, err
		}
		secret = string(secretBytes)
	}
	endpointParams := url.Values{}
	if len(audience) > 0 {
		endpointParams = url.Values{audienceKey: {audience}}
	}
	secret = strings.TrimSpace(secret)
	if tokenCache == nil {
		tokenCache = &cache.TokenCacheInMemoryProvider{}
	}
	return ClientCredentialsTokenSourceProvider{
		ccConfig: clientcredentials.Config{
			ClientID:       cfg.ClientID,
			ClientSecret:   secret,
			TokenURL:       tokenURL,
			Scopes:         scopes,
			EndpointParams: endpointParams,
		},
		tokenCache: tokenCache}, nil
}

func (p ClientCredentialsTokenSourceProvider) GetTokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	return &customTokenSource{
		ctx:        ctx,
		new:        p.ccConfig.TokenSource(ctx),
		mu:         sync.Mutex{},
		tokenCache: p.tokenCache,
	}, nil
}

type customTokenSource struct {
	ctx        context.Context
	mu         sync.Mutex // guards everything else
	new        oauth2.TokenSource
	tokenCache cache.TokenCache
}

func (s *customTokenSource) Token() (*oauth2.Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if token, err := s.tokenCache.GetToken(); err == nil && token.Valid() {
		return token, nil
	}

	token, err := s.new.Token()
	if err != nil {
		logger.Warnf(s.ctx, "failed to get token: %w", err)
		return nil, fmt.Errorf("failed to get token: %w", err)
	}
	logger.Infof(s.ctx, "retrieved token with expiry %v", token.Expiry)

	err = s.tokenCache.SaveToken(token)
	if err != nil {
		logger.Warnf(s.ctx, "failed to cache token: %w", err)
	}

	return token, nil
}

type DeviceFlowTokenSourceProvider struct {
	tokenOrchestrator deviceflow.TokenOrchestrator
}

func NewDeviceFlowTokenSourceProvider(baseTokenOrchestrator tokenorchestrator.BaseTokenOrchestrator, deviceFlowConfig deviceflow.Config) (TokenSourceProvider, error) {

	tokenOrchestrator, err := deviceflow.NewDeviceFlowTokenOrchestrator(baseTokenOrchestrator, deviceFlowConfig)
	if err != nil {
		return nil, err
	}

	return DeviceFlowTokenSourceProvider{tokenOrchestrator: tokenOrchestrator}, nil
}

func (p DeviceFlowTokenSourceProvider) GetTokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	return GetDeviceFlowAuthTokenSource(ctx, p.tokenOrchestrator)
}

// GetDeviceFlowAuthTokenSource Returns the token source which would be used for device auth flow
func GetDeviceFlowAuthTokenSource(ctx context.Context, deviceFlowOrchestrator deviceflow.TokenOrchestrator) (oauth2.TokenSource, error) {
	// explicitly ignore error while fetching token from cache.
	authToken, err := deviceFlowOrchestrator.FetchTokenFromCacheOrRefreshIt(ctx, deviceFlowOrchestrator.Config.TokenRefreshGracePeriod)
	if err != nil {
		logger.Warnf(ctx, "Failed fetching from cache. Will restart the flow. Error: %v", err)
	}

	if authToken == nil {
		// Fetch using auth flow
		if authToken, err = deviceFlowOrchestrator.FetchTokenFromAuthFlow(ctx); err != nil {
			logger.Errorf(ctx, "Error fetching token using auth flow due to %v", err)
			return nil, err
		}
	}

	return &pkce.SimpleTokenSource{
		CachedToken: authToken,
	}, nil
}
