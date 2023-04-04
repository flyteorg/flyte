package tokenorchestrator

import (
	"context"
	"fmt"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"

	"golang.org/x/oauth2"

	"github.com/flyteorg/flyteidl/clients/go/admin/cache"
	"github.com/flyteorg/flyteidl/clients/go/admin/oauth"
	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/logger"
)

// BaseTokenOrchestrator implements the main logic to initiate device authorization flow
type BaseTokenOrchestrator struct {
	ClientConfig *oauth.Config
	TokenCache   cache.TokenCache
}

// RefreshToken attempts to refresh the access token if a refresh token is provided.
func (t BaseTokenOrchestrator) RefreshToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error) {
	ts := t.ClientConfig.TokenSource(ctx, token)
	var refreshedToken *oauth2.Token
	var err error
	refreshedToken, err = ts.Token()
	if err != nil {
		logger.Warnf(ctx, "failed to refresh the token due to %v and will be doing re-auth", err)
		return nil, err
	}

	if refreshedToken != nil {
		logger.Debugf(ctx, "got a response from the refresh grant for old expiry %v with new expiry %v",
			token.Expiry, refreshedToken.Expiry)
		if refreshedToken.AccessToken != token.AccessToken {
			if err = t.TokenCache.SaveToken(refreshedToken); err != nil {
				logger.Errorf(ctx, "unable to save the new token due to %v", err)
				return nil, err
			}
		}
	}

	return refreshedToken, nil
}

// FetchTokenFromCacheOrRefreshIt fetches the token from cache and refreshes it if it'll expire within the
// Config.TokenRefreshGracePeriod period.
func (t BaseTokenOrchestrator) FetchTokenFromCacheOrRefreshIt(ctx context.Context, tokenRefreshGracePeriod config.Duration) (token *oauth2.Token, err error) {
	token, err = t.TokenCache.GetToken()
	if err != nil {
		return nil, err
	}

	if !token.Valid() {
		return nil, fmt.Errorf("token from cache is invalid")
	}

	// If token doesn't need to be refreshed, return it.
	if time.Now().Before(token.Expiry.Add(-tokenRefreshGracePeriod.Duration)) {
		logger.Infof(ctx, "found the token in the cache")
		return token, nil
	}
	token.Expiry = token.Expiry.Add(-tokenRefreshGracePeriod.Duration)

	token, err = t.RefreshToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token using cached token. Error: %w", err)
	}

	if !token.Valid() {
		return nil, fmt.Errorf("refreshed token is invalid")
	}

	err = t.TokenCache.SaveToken(token)
	if err != nil {
		return nil, fmt.Errorf("failed to save token in the token cache. Error: %w", err)
	}

	return token, nil
}

func NewBaseTokenOrchestrator(ctx context.Context, tokenCache cache.TokenCache, authClient service.AuthMetadataServiceClient) (BaseTokenOrchestrator, error) {
	clientConfig, err := oauth.BuildConfigFromMetadataService(ctx, authClient)
	if err != nil {
		return BaseTokenOrchestrator{}, err
	}
	return BaseTokenOrchestrator{
		ClientConfig: clientConfig,
		TokenCache:   tokenCache,
	}, nil
}
