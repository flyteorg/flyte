// Contains types needed to start up a standalone OAuth2 Authorization Server or delegate authentication to an external
// provider. It supports OpenId connect for user authentication.
package auth

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"

	"github.com/coreos/go-oidc"
	"github.com/flyteorg/flyteadmin/auth/config"
	"github.com/flyteorg/flyteadmin/auth/interfaces"
	"github.com/flyteorg/flytestdlib/errors"
	"github.com/flyteorg/flytestdlib/logger"
	"golang.org/x/oauth2"
)

const (
	IdpConnectionTimeout = 10 * time.Second

	ErrauthCtx        errors.ErrorCode = "AUTH_CONTEXT_SETUP_FAILED"
	ErrConfigFileRead errors.ErrorCode = "CONFIG_OPTION_FILE_READ_FAILED"
)

var (
	callbackRelativeURL = config.MustParseURL("/callback")
	rootRelativeURL     = config.MustParseURL("/")
)

// Please see the comment on the corresponding AuthenticationContext for more information.
type Context struct {
	oauth2Client         *oauth2.Config
	cookieManager        interfaces.CookieHandler
	oidcProvider         *oidc.Provider
	options              *config.Config
	oauth2Provider       interfaces.OAuth2Provider
	oauth2ResourceServer interfaces.OAuth2ResourceServer
	authServiceImpl      service.AuthMetadataServiceServer
	identityServiceIml   service.IdentityServiceServer

	userInfoURL       *url.URL
	oauth2MetadataURL *url.URL
	oidcMetadataURL   *url.URL
	httpClient        *http.Client
}

func (c Context) OAuth2Provider() interfaces.OAuth2Provider {
	return c.oauth2Provider
}

func (c Context) OAuth2ClientConfig(requestURL *url.URL) *oauth2.Config {
	if requestURL == nil || strings.HasPrefix(c.oauth2Client.RedirectURL, requestURL.ResolveReference(rootRelativeURL).String()) {
		return c.oauth2Client
	}

	return &oauth2.Config{
		RedirectURL:  requestURL.ResolveReference(callbackRelativeURL).String(),
		ClientID:     c.oauth2Client.ClientID,
		ClientSecret: c.oauth2Client.ClientSecret,
		Scopes:       c.oauth2Client.Scopes,
		Endpoint:     c.oauth2Client.Endpoint,
	}
}

func (c Context) OidcProvider() *oidc.Provider {
	return c.oidcProvider
}

func (c Context) CookieManager() interfaces.CookieHandler {
	return c.cookieManager
}

func (c Context) Options() *config.Config {
	return c.options
}

func (c Context) GetUserInfoURL() *url.URL {
	return c.userInfoURL
}

func (c Context) GetHTTPClient() *http.Client {
	return c.httpClient
}

func (c Context) GetOAuth2MetadataURL() *url.URL {
	return c.oauth2MetadataURL
}

func (c Context) GetOIdCMetadataURL() *url.URL {
	return c.oidcMetadataURL
}

func (c Context) AuthMetadataService() service.AuthMetadataServiceServer {
	return c.authServiceImpl
}

func (c Context) IdentityService() service.IdentityServiceServer {
	return c.identityServiceIml
}

func (c Context) OAuth2ResourceServer() interfaces.OAuth2ResourceServer {
	return c.oauth2ResourceServer
}
func NewAuthenticationContext(ctx context.Context, sm core.SecretManager, oauth2Provider interfaces.OAuth2Provider,
	oauth2ResourceServer interfaces.OAuth2ResourceServer, authMetadataService service.AuthMetadataServiceServer,
	identityService service.IdentityServiceServer, options *config.Config) (Context, error) {

	// Construct the cookie manager object.
	hashKeyBase64, err := sm.Get(ctx, options.UserAuth.CookieHashKeySecretName)
	if err != nil {
		return Context{}, errors.Wrapf(ErrConfigFileRead, err, "Could not read hash key file")
	}

	blockKeyBase64, err := sm.Get(ctx, options.UserAuth.CookieBlockKeySecretName)
	if err != nil {
		return Context{}, errors.Wrapf(ErrConfigFileRead, err, "Could not read hash key file")
	}

	cookieManager, err := NewCookieManager(ctx, hashKeyBase64, blockKeyBase64, options.UserAuth.CookieSetting)
	if err != nil {
		logger.Errorf(ctx, "Error creating cookie manager %s", err)
		return Context{}, errors.Wrapf(ErrauthCtx, err, "Error creating cookie manager")
	}

	// Construct an http client for interacting with the IDP if necessary.
	httpClient := &http.Client{
		Timeout: IdpConnectionTimeout,
	}

	if len(options.UserAuth.HTTPProxyURL.String()) > 0 {
		logger.Infof(ctx, "HTTPProxy URL for OAuth2 is: %s", options.UserAuth.HTTPProxyURL.String())
		httpClient.Transport = &http.Transport{Proxy: http.ProxyURL(&options.UserAuth.HTTPProxyURL.URL)}
	}

	// Construct an oidc Provider, which needs its own http Client.
	oidcCtx := oidc.ClientContext(ctx, httpClient)
	baseURL := options.UserAuth.OpenID.BaseURL.String()
	provider, err := oidc.NewProvider(oidcCtx, baseURL)
	if err != nil {
		return Context{}, errors.Wrapf(ErrauthCtx, err, "Error creating oidc provider w/ issuer [%v]", baseURL)
	}

	// Construct the golang OAuth2 library's own internal configuration object from this package's config
	oauth2Config, err := GetOAuth2ClientConfig(ctx, options.UserAuth.OpenID, provider.Endpoint(), sm)
	if err != nil {
		return Context{}, errors.Wrapf(ErrauthCtx, err, "Error creating OAuth2 library configuration")
	}

	logger.Infof(ctx, "Base IDP URL is %s", options.UserAuth.OpenID.BaseURL)

	oauth2MetadataURL, err := url.Parse(OAuth2MetadataEndpoint)
	if err != nil {
		logger.Errorf(ctx, "Error parsing oauth2 metadata URL %s", err)
		return Context{}, errors.Wrapf(ErrauthCtx, err, "Error parsing metadata URL")
	}

	logger.Infof(ctx, "Metadata endpoint is %s", oauth2MetadataURL)

	oidcMetadataURL, err := url.Parse(OIdCMetadataEndpoint)
	if err != nil {
		logger.Errorf(ctx, "Error parsing oidc metadata URL %s", err)
		return Context{}, errors.Wrapf(ErrauthCtx, err, "Error parsing metadata URL")
	}

	logger.Infof(ctx, "Metadata endpoint is %s", oidcMetadataURL)

	authCtx := Context{
		options:              options,
		oidcMetadataURL:      oidcMetadataURL,
		oauth2MetadataURL:    oauth2MetadataURL,
		oauth2Client:         &oauth2Config,
		oidcProvider:         provider,
		httpClient:           httpClient,
		cookieManager:        cookieManager,
		oauth2Provider:       oauth2Provider,
		oauth2ResourceServer: oauth2ResourceServer,
	}

	authCtx.authServiceImpl = authMetadataService
	authCtx.identityServiceIml = identityService

	return authCtx, nil
}

// This creates a oauth2 library config object, with values from the Flyte Admin config
func GetOAuth2ClientConfig(ctx context.Context, options config.OpenIDOptions, providerEndpoints oauth2.Endpoint, sm core.SecretManager) (cfg oauth2.Config, err error) {
	var secret string
	if len(options.DeprecatedClientSecretFile) > 0 {
		secretBytes, err := ioutil.ReadFile(options.DeprecatedClientSecretFile)
		if err != nil {
			return oauth2.Config{}, err
		}

		secret = string(secretBytes)
	} else {
		secret, err = sm.Get(ctx, options.ClientSecretName)
		if err != nil {
			return oauth2.Config{}, err
		}
	}

	secret = strings.TrimSuffix(secret, "\n")

	return oauth2.Config{
		RedirectURL:  callbackRelativeURL.String(),
		ClientID:     options.ClientID,
		ClientSecret: secret,
		Scopes:       options.Scopes,
		Endpoint:     providerEndpoints,
	}, nil
}
