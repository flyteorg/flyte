package pkce

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flytestdlib/logger"

	"github.com/pkg/browser"
	"golang.org/x/oauth2"
)

// TokenOrchestrator implements the main logic to initiate Pkce flow to issue access token and refresh token as well as
// refreshing the access token if a refresh token is present.
type TokenOrchestrator struct {
	cfg          Config
	clientConfig *oauth2.Config
	tokenCache   TokenCache
}

// RefreshToken attempts to refresh the access token if a refresh token is provided.
func (f TokenOrchestrator) RefreshToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error) {
	ts := f.clientConfig.TokenSource(ctx, token)
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
			if err = f.tokenCache.SaveToken(refreshedToken); err != nil {
				logger.Errorf(ctx, "unable to save the new token due to %v", err)
				return nil, err
			}
		}
	}

	return refreshedToken, nil
}

// FetchTokenFromCacheOrRefreshIt fetches the token from cache and refreshes it if it'll expire within the
// Config.TokenRefreshGracePeriod period.
func (f TokenOrchestrator) FetchTokenFromCacheOrRefreshIt(ctx context.Context) (token *oauth2.Token, err error) {
	token, err = f.tokenCache.GetToken()
	if err != nil {
		return nil, err
	}

	if !token.Valid() {
		return nil, fmt.Errorf("token from cache is invalid")
	}

	// If token doesn't need to be refreshed, return it.
	if token.Expiry.Add(f.cfg.TokenRefreshGracePeriod.Duration).Before(time.Now()) {
		return token, nil
	}

	token, err = f.RefreshToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token using cached token. Error: %w", err)
	}

	if !token.Valid() {
		return nil, fmt.Errorf("refreshed token is invalid")
	}

	err = f.tokenCache.SaveToken(token)
	if err != nil {
		return nil, fmt.Errorf("failed to save token in the token cache. Error: %w", err)
	}

	return token, nil
}

// FetchTokenFromAuthFlow starts a webserver to listen to redirect callback from the authorization server at the end
// of the flow. It then launches the browser to authenticate the user.
func (f TokenOrchestrator) FetchTokenFromAuthFlow(ctx context.Context) (*oauth2.Token, error) {
	var err error
	var redirectURL *url.URL
	if redirectURL, err = url.Parse(f.clientConfig.RedirectURL); err != nil {
		return nil, err
	}

	// pkceCodeVerifier stores the generated random value which the client will on-send to the auth server with the received
	// authorization code. This way the oauth server can verify that the base64URLEncoded(sha265(codeVerifier)) matches
	// the stored code challenge, which was initially sent through with the code+PKCE authorization request to ensure
	// that this is the original user-agent who requested the access token.
	pkceCodeVerifier := generateCodeVerifier(64)

	// pkceCodeChallenge stores the base64(sha256(codeVerifier)) which is sent from the
	// client to the auth server as required for PKCE.
	pkceCodeChallenge := generateCodeChallenge(pkceCodeVerifier)

	stateString := state(32)
	nonces := state(32)

	tokenChannel := make(chan *oauth2.Token, 1)
	errorChannel := make(chan error, 1)

	// Replace S256 with one from cient config and provide a support to generate code challenge using the passed
	// in method.
	urlToOpen := f.clientConfig.AuthCodeURL(stateString) + "&nonce=" + nonces + "&code_challenge=" +
		pkceCodeChallenge + "&code_challenge_method=S256"

	serveMux := http.NewServeMux()
	server := &http.Server{Addr: redirectURL.Host, Handler: serveMux}
	// Register the call back handler
	serveMux.HandleFunc(redirectURL.Path, getAuthServerCallbackHandler(f.clientConfig, pkceCodeVerifier,
		tokenChannel, errorChannel, stateString)) // the oauth2 callback endpoint
	defer server.Close()

	go func() {
		if err = server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(ctx, "Couldn't start the callback http server on host %v due to %v", redirectURL.Host,
				err)
		}
	}()

	logger.Infof(ctx, "Opening the browser at "+urlToOpen)
	if err = browser.OpenURL(urlToOpen); err != nil {
		return nil, err
	}

	ctx, cancelNow := context.WithTimeout(ctx, f.cfg.BrowserSessionTimeout.Duration)
	defer cancelNow()

	var token *oauth2.Token
	select {
	case err = <-errorChannel:
		return nil, err
	case <-ctx.Done():
		return nil, fmt.Errorf("context was canceled during auth flow")
	case token = <-tokenChannel:
		if err = f.tokenCache.SaveToken(token); err != nil {
			logger.Errorf(ctx, "unable to save the new token due to. Will ignore the error and use the issued token. Error: %v", err)
		}

		return token, nil
	}
}

// NewTokenOrchestrator creates a new TokenOrchestrator that implements the main logic to initiate Pkce flow to issue
// access token and refresh token as well as refreshing the access token if a refresh token is present.
func NewTokenOrchestrator(ctx context.Context, cfg Config, tokenCache TokenCache, authMetadataClient service.AuthMetadataServiceClient) (TokenOrchestrator, error) {
	clientConf, err := BuildClientConfig(ctx, authMetadataClient)
	if err != nil {
		return TokenOrchestrator{}, err
	}

	return TokenOrchestrator{
		cfg:          cfg,
		clientConfig: clientConf,
		tokenCache:   tokenCache,
	}, nil
}
