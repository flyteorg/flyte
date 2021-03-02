package auth

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
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
	oauth2            *oauth2.Config
	claims            config.Claims
	cookieManager     interfaces.CookieHandler
	oidcProvider      *oidc.Provider
	options           config.OAuthOptions
	userInfoURL       *url.URL
	baseURL           *url.URL
	oauth2MetadataURL *url.URL
	oidcMetadataURL   *url.URL
	httpClient        *http.Client
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

func (c Context) GetBaseURL() *url.URL {
	return c.baseURL
}

func (c Context) GetOAuth2MetadataURL() *url.URL {
	return c.oauth2MetadataURL
}

func (c Context) GetOIdCMetadataURL() *url.URL {
	return c.oidcMetadataURL
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
		return Context{}, errors.Wrapf(ErrAuthContext, err, "Error creating oidc provider w/ issuer [%v]", options.Claims.Issuer)
	}

	result.oidcProvider = provider

	// TODO: Convert all the URLs in this config to the config.URL type
	// Then we will not have to do any of the parsing in this code here, and the error handling will be taken care for
	// us by the flytestdlib config parser.
	// Construct base URL object
	base, err := url.Parse(options.BaseURL)
	if err != nil {
		logger.Errorf(ctx, "Error parsing base URL %s", err)
		return Context{}, errors.Wrapf(ErrAuthContext, err, "Error parsing IDP base URL")
	}

	logger.Infof(ctx, "Base IDP URL is %s", base)
	result.baseURL = base

	result.oauth2MetadataURL, err = url.Parse(OAuth2MetadataEndpoint)
	if err != nil {
		logger.Errorf(ctx, "Error parsing oauth2 metadata URL %s", err)
		return Context{}, errors.Wrapf(ErrAuthContext, err, "Error parsing metadata URL")
	}

	logger.Infof(ctx, "Metadata endpoint is %s", result.oauth2MetadataURL)

	result.oidcMetadataURL, err = url.Parse(OIdCMetadataEndpoint)
	if err != nil {
		logger.Errorf(ctx, "Error parsing oidc metadata URL %s", err)
		return Context{}, errors.Wrapf(ErrAuthContext, err, "Error parsing metadata URL")
	}

	logger.Infof(ctx, "Metadata endpoint is %s", result.oidcMetadataURL)

	// Construct the URL object for the user info endpoint if applicable
	if options.IdpUserInfoEndpoint != "" {
		parsedURL, err := url.Parse(options.IdpUserInfoEndpoint)
		if err != nil {
			logger.Errorf(ctx, "Error parsing total IDP user info path %s as URL %s", options.IdpUserInfoEndpoint, err)
			return Context{}, errors.Wrapf(ErrAuthContext, err,
				"Error parsing IDP user info path as URL while constructing IDP user info endpoint")
		}
		finalURL := result.baseURL.ResolveReference(parsedURL)
		logger.Infof(ctx, "User info URL for IDP is %s", finalURL.String())
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
		Scopes:       options.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  options.AuthorizeURL,
			TokenURL: options.TokenURL,
		},
	}, nil
}
