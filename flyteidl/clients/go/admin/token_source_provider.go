package admin

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/flyteorg/flyteidl/clients/go/admin/cache"
	"github.com/flyteorg/flyteidl/clients/go/admin/deviceflow"
	"github.com/flyteorg/flyteidl/clients/go/admin/externalprocess"
	"github.com/flyteorg/flyteidl/clients/go/admin/pkce"
	"github.com/flyteorg/flyteidl/clients/go/admin/tokenorchestrator"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flytestdlib/logger"
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
		if len(scopes) == 0 {
			clientMetadata, err := authClient.GetPublicClientConfig(ctx, &service.PublicClientAuthConfigRequest{})
			if err != nil {
				return nil, fmt.Errorf("failed to fetch client metadata. Error: %v", err)
			}
			scopes = clientMetadata.Scopes
		}

		tokenProvider, err = NewClientCredentialsTokenSourceProvider(ctx, cfg, scopes, tokenURL)
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
	ccConfig           clientcredentials.Config
	TokenRefreshWindow time.Duration
}

func NewClientCredentialsTokenSourceProvider(ctx context.Context, cfg *Config, scopes []string, tokenURL string) (TokenSourceProvider, error) {
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
	secret = strings.TrimSpace(secret)
	return ClientCredentialsTokenSourceProvider{
		ccConfig: clientcredentials.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: secret,
			TokenURL:     tokenURL,
			Scopes:       scopes},
		TokenRefreshWindow: cfg.TokenRefreshWindow.Duration}, nil
}

func (p ClientCredentialsTokenSourceProvider) GetTokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	if p.TokenRefreshWindow > 0 {
		source := p.ccConfig.TokenSource(ctx)
		return &customTokenSource{
			new:                source,
			mu:                 sync.Mutex{},
			t:                  nil,
			tokenRefreshWindow: p.TokenRefreshWindow,
		}, nil
	}
	return p.ccConfig.TokenSource(ctx), nil
}

type customTokenSource struct {
	new                oauth2.TokenSource
	mu                 sync.Mutex // guards everything else
	t                  *oauth2.Token
	refreshTime        time.Time
	failedToRefresh    bool
	tokenRefreshWindow time.Duration
}

func (s *customTokenSource) Token() (*oauth2.Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.t.Valid() {
		if time.Now().After(s.refreshTime) && !s.failedToRefresh {
			t, err := s.new.Token()
			if err != nil {
				s.failedToRefresh = true // don't try to refresh again before expiry
				return s.t, nil
			}
			s.t = t
			s.refreshTime = s.t.Expiry.Add(-getRandomDuration(s.tokenRefreshWindow))
			s.failedToRefresh = false
			return s.t, nil
		}
		return s.t, nil
	}
	t, err := s.new.Token()
	if err != nil {
		return nil, err
	}
	s.t = t
	s.failedToRefresh = false
	s.refreshTime = s.t.Expiry.Add(-getRandomDuration(s.tokenRefreshWindow))
	return t, nil
}

// Get random duration between 0 and maxDuration
func getRandomDuration(maxDuration time.Duration) time.Duration {
	// d is 1.0 to 2.0 times maxDuration
	d := wait.Jitter(maxDuration, 1)
	return d - maxDuration
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
