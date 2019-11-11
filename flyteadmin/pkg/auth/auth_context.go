package auth

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/lyft/flyteadmin/pkg/auth/config"
	"github.com/lyft/flyteadmin/pkg/auth/interfaces"
	"github.com/lyft/flytestdlib/errors"
	"github.com/lyft/flytestdlib/logger"
	"golang.org/x/oauth2"
)

const (
	IdpConnectionTimeout = 10 * time.Second
)

// Please see the comment on the corresponding AuthenticationContext for more information.
type Context struct {
	oauth2        *oauth2.Config
	claims        config.Claims
	cookieManager interfaces.CookieHandler
	oidcProvider  *oidc.Provider
	options       config.OAuthOptions
	userInfoURL   *url.URL
	httpClient    *http.Client
}

func (c Context) OAuth2Config() *oauth2.Config {
	return c.oauth2
}

func (c Context) Claims() config.Claims {
	return c.claims
}

func (c Context) OidcProvider() *oidc.Provider {
	return c.oidcProvider
}

func (c Context) CookieManager() interfaces.CookieHandler {
	return c.cookieManager
}

func (c Context) Options() config.OAuthOptions {
	return c.options
}

func (c Context) GetUserInfoURL() *url.URL {
	return c.userInfoURL
}

func (c Context) GetHTTPClient() *http.Client {
	return c.httpClient
}

const (
	ErrAuthContext    errors.ErrorCode = "AUTH_CONTEXT_SETUP_FAILED"
	ErrConfigFileRead errors.ErrorCode = "CONFIG_OPTION_FILE_READ_FAILED"
)

func NewAuthenticationContext(ctx context.Context, options config.OAuthOptions) (Context, error) {
	result := Context{
		claims:  options.Claims,
		options: options,
	}

	// Construct the golang OAuth2 library's own internal configuration object from this package's config
	oauth2Config, err := GetOauth2Config(options)
	if err != nil {
		return Context{}, errors.Wrapf(ErrAuthContext, err, "Error creating OAuth2 library configuration")
	}
	result.oauth2 = &oauth2Config

	// Construct the cookie manager object.
	hashKeyBytes, err := ioutil.ReadFile(options.CookieHashKeyFile)
	if err != nil {
		return Context{}, errors.Wrapf(ErrConfigFileRead, err, "Could not read hash key file")
	}
	blockKeyBytes, err := ioutil.ReadFile(options.CookieBlockKeyFile)
	if err != nil {
		return Context{}, errors.Wrapf(ErrConfigFileRead, err, "Could not read block key file")
	}
	cookieManager, err := NewCookieManager(ctx, string(hashKeyBytes), string(blockKeyBytes))
	if err != nil {
		logger.Errorf(ctx, "Error creating cookie manager %s", err)
		return Context{}, errors.Wrapf(ErrAuthContext, err, "Error creating cookie manager")
	}
	result.cookieManager = cookieManager

	// Construct an oidc Provider, which needs its own http Client.
	oidcCtx := oidc.ClientContext(ctx, &http.Client{})
	provider, err := oidc.NewProvider(oidcCtx, options.Claims.Issuer)
	if err != nil {
		return Context{}, errors.Wrapf(ErrAuthContext, err, "Error creating oidc provider")
	}
	result.oidcProvider = provider

	// Construct the URL object for the user info endpoint if applicable
	if options.IdpUserInfoEndpoint != "" {
		base, err := url.Parse(options.BaseURL)
		if err != nil {
			logger.Errorf(ctx, "Error parsing base URL %s", err)
			return Context{}, errors.Wrapf(ErrAuthContext, err,
				"Error parsing base URL while constructing IDP user info endpoint")
		}
		joinedPath := path.Join(base.EscapedPath(), options.IdpUserInfoEndpoint)

		parsedURL, err := url.Parse(joinedPath)
		if err != nil {
			logger.Errorf(ctx, "Error parsing total IDP user info path %s as URL %s", joinedPath, err)
			return Context{}, errors.Wrapf(ErrAuthContext, err,
				"Error parsing IDP user info path as URL while constructing IDP user info endpoint")
		}
		finalURL := base.ResolveReference(parsedURL)
		logger.Infof(ctx, "The /userinfo URL for the IDP is %s", finalURL.String())
		result.userInfoURL = finalURL
	}

	// Construct an http client for interacting with the IDP if necessary.
	result.httpClient = &http.Client{
		Timeout: IdpConnectionTimeout,
	}

	return result, nil
}

// This creates a oauth2 library config object, with values from the Flyte Admin config
func GetOauth2Config(options config.OAuthOptions) (oauth2.Config, error) {
	secretBytes, err := ioutil.ReadFile(options.ClientSecretFile)
	if err != nil {
		return oauth2.Config{}, err
	}
	secret := strings.TrimSuffix(string(secretBytes), "\n")
	return oauth2.Config{
		RedirectURL:  options.CallbackURL,
		ClientID:     options.ClientID,
		ClientSecret: secret,
		// Offline access needs to be specified in order to return a refresh token in the exchange.
		// TODO: Second parameter is IDP specific - move to config. Also handle case where a refresh token is not allowed
		Scopes: []string{OidcScope, OfflineAccessType, ProfileScope},
		Endpoint: oauth2.Endpoint{
			AuthURL:  options.AuthorizeURL,
			TokenURL: options.TokenURL,
		},
	}, nil
}
