package config

import (
	"net/url"
	"time"

	"github.com/ory/fosite"

	"github.com/flyteorg/flytestdlib/config"
)

//go:generate pflags Config --default-var=DefaultConfig
//go:generate enumer --type=AuthorizationServerType --trimprefix=AuthorizationServerType -json

type SecretName = string

const (
	// SecretNameOIdCClientSecret defines the default OIdC client secret name to use.
	// #nosec
	SecretNameOIdCClientSecret SecretName = "oidc_client_secret"

	// SecretNameCookieHashKey defines the default cookie hash key secret name to use.
	// #nosec
	SecretNameCookieHashKey SecretName = "cookie_hash_key"

	// SecretNameCookieBlockKey defines the default cookie block key secret name to use.
	// #nosec
	SecretNameCookieBlockKey SecretName = "cookie_block_key"

	// SecretNameClaimSymmetricKey must be a base64 encoded secret of exactly 32 bytes
	// #nosec
	SecretNameClaimSymmetricKey SecretName = "claim_symmetric_key"

	// SecretNameTokenSigningRSAKey is the privateKey used to sign JWT tokens. The default strategy uses RS256 (RSA Signature with SHA-256)
	// #nosec
	SecretNameTokenSigningRSAKey SecretName = "token_rsa_key.pem"

	// SecretNameOldTokenSigningRSAKey is the privateKey used to sign old JWT tokens. The default strategy uses RS256 (RSA Signature with SHA-256)
	// This is used to support key rotation. When present, it'll only be used to validate incoming tokens. New tokens
	// will not be issued using this key.
	// #nosec
	SecretNameOldTokenSigningRSAKey SecretName = "token_rsa_key_old.pem"
)

// AuthorizationServerType defines the type of Authorization Server to use.
type AuthorizationServerType int

const (
	// AuthorizationServerTypeSelf determines that FlyteAdmin should act as the authorization server to serve
	// OAuth2 token requests
	AuthorizationServerTypeSelf AuthorizationServerType = iota

	// AuthorizationServerTypeExternal determines that FlyteAdmin should rely on an external authorization server (e.g.
	// Okta) to serve OAuth2 token requests
	AuthorizationServerTypeExternal
)

var (
	DefaultConfig = &Config{
		// Please see the comments in this struct's definition for more information
		HTTPAuthorizationHeader: "flyte-authorization",
		GrpcAuthorizationHeader: "flyte-authorization",
		UserAuth: UserAuthConfig{
			RedirectURL:              config.URL{URL: *MustParseURL("/console")},
			CookieHashKeySecretName:  SecretNameCookieHashKey,
			CookieBlockKeySecretName: SecretNameCookieBlockKey,
			OpenID: OpenIDOptions{
				ClientSecretName: SecretNameOIdCClientSecret,
				// Default claims that should be supported by any OIdC server. Refer to https://openid.net/specs/openid-connect-core-1_0.html#ScopeClaims
				// for a complete list.
				Scopes: []string{
					"openid",
					"profile",
				},
			},
			CookieSetting: CookieSettings{
				Domain:         "",
				SameSitePolicy: SameSiteDefaultMode,
			},
		},
		AppAuth: OAuth2Options{
			AuthServerType: AuthorizationServerTypeSelf,
			ThirdParty: ThirdPartyConfigOptions{
				FlyteClientConfig: FlyteClientConfig{
					ClientID:    "flytectl",
					RedirectURI: "http://localhost:53593/callback",
					Scopes:      []string{"all", "offline"},
				},
			},
			SelfAuthServer: AuthorizationServer{
				AccessTokenLifespan:                   config.Duration{Duration: 30 * time.Minute},
				RefreshTokenLifespan:                  config.Duration{Duration: 60 * time.Minute},
				AuthorizationCodeLifespan:             config.Duration{Duration: 5 * time.Minute},
				ClaimSymmetricEncryptionKeySecretName: SecretNameClaimSymmetricKey,
				TokenSigningRSAKeySecretName:          SecretNameTokenSigningRSAKey,
				OldTokenSigningRSAKeySecretName:       SecretNameOldTokenSigningRSAKey,
				StaticClients: map[string]*fosite.DefaultClient{
					"flyte-cli": {
						ID:            "flyte-cli",
						RedirectURIs:  []string{"http://localhost:53593/callback", "http://localhost:12345/callback"},
						ResponseTypes: []string{"code", "token"},
						GrantTypes:    []string{"refresh_token", "authorization_code"},
						Scopes:        []string{"all", "offline", "access_token"},
						Public:        true,
					},
					"flytectl": {
						ID:            "flytectl",
						RedirectURIs:  []string{"http://localhost:53593/callback", "http://localhost:12345/callback"},
						ResponseTypes: []string{"code", "token"},
						GrantTypes:    []string{"refresh_token", "authorization_code"},
						Scopes:        []string{"all", "offline", "access_token"},
						Public:        true,
					},
					"flytepropeller": {
						ID:            "flytepropeller",
						Secret:        []byte(`$2a$06$d6PQn2QAFU3cL5V8MDkeuuk63xubqUxNxjtfPw.Fc9MgV6vpmyOIy`), // Change this.
						RedirectURIs:  []string{"http://localhost:3846/callback"},
						ResponseTypes: []string{"token"},
						GrantTypes:    []string{"refresh_token", "client_credentials"},
						Scopes:        []string{"all", "offline", "access_token"},
					},
				},
			},
		},
	}

	cfgSection = config.MustRegisterSection("auth", DefaultConfig)
)

type Config struct {
	// These settings are for non-SSL authentication modes, where Envoy is handling SSL termination
	// This is not yet used, but this is the HTTP variant of the setting below.
	HTTPAuthorizationHeader string `json:"httpAuthorizationHeader"`

	// In order to support deployments of this Admin service where Envoy is terminating SSL connections, the metadata
	// header name cannot be "authorization", which is the standard metadata name. Envoy has special handling for that
	// name. Instead, there is a gRPC interceptor, GetAuthenticationCustomMetadataInterceptor, that will translate
	// incoming metadata headers with this config setting's name, into that standard header
	GrpcAuthorizationHeader string `json:"grpcAuthorizationHeader"`

	// To help ease migration, it was helpful to be able to only selectively enforce authentication.  The
	// dimension that made the most sense to cut by at time of writing is HTTP vs gRPC as the web UI mainly used HTTP
	// and the backend used mostly gRPC.  Cutting by individual endpoints is another option but it possibly falls more
	// into the realm of authorization rather than authentication.
	DisableForHTTP bool `json:"disableForHttp" pflag:",Disables auth enforcement on HTTP Endpoints."`
	DisableForGrpc bool `json:"disableForGrpc" pflag:",Disables auth enforcement on Grpc Endpoints."`

	// AuthorizedURIs is optional and defines the set of URIs that clients are allowed to visit the service on. If set,
	// the system will attempt to match the incoming host to the first authorized URIs and use that (including the scheme)
	// when generating metadata endpoints and when validating audience and issuer claims. If no matching authorizedUri
	// is found, it'll default to the first one. If not provided, the urls will be deduced based on the request url and
	// the `secure` setting.
	AuthorizedURIs []config.URL `json:"authorizedUris" pflag:"-,Optional: Defines the set of URIs that clients are allowed to visit the service on. If set, the system will attempt to match the incoming host to the first authorized URIs and use that (including the scheme) when generating metadata endpoints and when validating audience and issuer claims. If not provided, the urls will be deduced based on the request url and the 'secure' setting."`

	// HTTPProxyURL allows operators to access external OAuth2 servers using an external HTTP Proxy
	HTTPProxyURL config.URL `json:"httpProxyURL" pflag:",OPTIONAL: HTTP Proxy to be used for OAuth requests."`

	// UserAuth settings used to authenticate end users in web-browsers.
	UserAuth UserAuthConfig `json:"userAuth" pflag:",Defines Auth options for users."`

	// AppAuth settings used to authenticate and control/limit access scopes for apps.
	AppAuth OAuth2Options `json:"appAuth" pflag:",Defines Auth options for apps. UserAuth must be enabled for AppAuth to work."`
}

type AuthorizationServer struct {
	// Defines the issuer to use when issuing and validating tokens. The default value is https://<requestUri.HostAndPort>/
	Issuer string `json:"issuer" pflag:",Defines the issuer to use when issuing and validating tokens. The default value is https://<requestUri.HostAndPort>/"`

	// Defines the lifespan of issued access tokens.
	AccessTokenLifespan config.Duration `json:"accessTokenLifespan" pflag:",Defines the lifespan of issued access tokens."`

	// Defines the lifespan of issued access tokens.
	RefreshTokenLifespan config.Duration `json:"refreshTokenLifespan" pflag:",Defines the lifespan of issued access tokens."`

	// Defines the lifespan of issued access tokens.
	AuthorizationCodeLifespan config.Duration `json:"authorizationCodeLifespan" pflag:",Defines the lifespan of issued access tokens."`

	// Secret names, defaults are set in DefaultConfig variable above but are possible to override through configs.
	ClaimSymmetricEncryptionKeySecretName string `json:"claimSymmetricEncryptionKeySecretName" pflag:",OPTIONAL: Secret name to use to encrypt claims in authcode token."`
	TokenSigningRSAKeySecretName          string `json:"tokenSigningRSAKeySecretName" pflag:",OPTIONAL: Secret name to use to retrieve RSA Signing Key."`
	OldTokenSigningRSAKeySecretName       string `json:"oldTokenSigningRSAKeySecretName" pflag:",OPTIONAL: Secret name to use to retrieve Old RSA Signing Key. This can be useful during key rotation to continue to accept older tokens."`

	// A list of clients to grant access to.
	StaticClients map[string]*fosite.DefaultClient `json:"staticClients" pflag:"-,Defines statically defined list of clients to allow."`
}

type ExternalAuthorizationServer struct {
	// BaseURL should be the base url of the authorization server that you are trying to hit. With Okta for instance, it will look something like https://company.okta.com/oauth2/abcdef123456789/
	// If not provided, the OpenID.BaseURL will be assumed instead.
	BaseURL             config.URL `json:"baseUrl" pflag:",This should be the base url of the authorization server that you are trying to hit. With Okta for instance, it will look something like https://company.okta.com/oauth2/abcdef123456789/"`
	AllowedAudience     []string   `json:"allowedAudience" pflag:",Optional: A list of allowed audiences. If not provided, the audience is expected to be the public Uri of the service."`
	MetadataEndpointURL config.URL `json:"metadataUrl" pflag:",Optional: If the server doesn't support /.well-known/oauth-authorization-server, you can set a custom metadata url here.'"`
	// HTTPProxyURL allows operators to access external OAuth2 servers using an external HTTP Proxy
	HTTPProxyURL config.URL `json:"httpProxyURL" pflag:",OPTIONAL: HTTP Proxy to be used for OAuth requests."`
}

// OAuth2Options defines settings for app auth.
type OAuth2Options struct {
	// AuthServerType defines the type of AuthServer to connect to.
	AuthServerType AuthorizationServerType `json:"authServerType" pflag:"-,Determines authorization server type to use. Additional config should be provided for the chosen AuthorizationServer"`

	// SelfAuthServer defines settings for running authorization server locally.
	SelfAuthServer AuthorizationServer `json:"selfAuthServer" pflag:",Authorization Server config to run as a service. Use this when using an IdP that does not offer a custom OAuth2 Authorization Server."`

	// ExternalAuthServer defines settings for connecting with an external authorization server.
	ExternalAuthServer ExternalAuthorizationServer `json:"externalAuthServer" pflag:",External Authorization Server config."`

	// ThirdParty defines settings for the public client flyte-clients (e.g. flytectl, flytecli) should use.
	ThirdParty ThirdPartyConfigOptions `json:"thirdPartyConfig" pflag:",Defines settings to instruct flyte cli tools (and optionally others) on what config to use to setup their client."`
}

type UserAuthConfig struct {
	// This is where the user will be redirected to at the end of the flow, but you should not use it. Instead,
	// the initial /login handler should be called with a redirect_url parameter, which will get saved to a cookie.
	// This setting will only be used when that cookie is missing.
	// See the login handler code for more comments.
	RedirectURL config.URL `json:"redirectUrl"`

	// OpenID defines settings for connecting and trusting an OpenIDConnect provider.
	OpenID OpenIDOptions `json:"openId" pflag:",OpenID Configuration for User Auth"`
	// Possibly add basicAuth & SAML/p support.

	// HTTPProxyURL allows operators to access external OAuth2 servers using an external HTTP Proxy
	HTTPProxyURL config.URL `json:"httpProxyURL" pflag:",OPTIONAL: HTTP Proxy to be used for OAuth requests."`

	// Secret names, defaults are set in DefaultConfig variable above but are possible to override through configs.
	CookieHashKeySecretName  string         `json:"cookieHashKeySecretName" pflag:",OPTIONAL: Secret name to use for cookie hash key."`
	CookieBlockKeySecretName string         `json:"cookieBlockKeySecretName" pflag:",OPTIONAL: Secret name to use for cookie block key."`
	CookieSetting            CookieSettings `json:"cookieSetting" pflag:", settings used by cookies created for user auth"`
}

//go:generate enumer --type=SameSite --trimprefix=SameSite -json
type SameSite int

const (
	SameSiteDefaultMode SameSite = iota
	SameSiteLaxMode
	SameSiteStrictMode
	SameSiteNoneMode
)

type CookieSettings struct {
	SameSitePolicy SameSite `json:"sameSitePolicy" pflag:",OPTIONAL: Allows you to declare if your cookie should be restricted to a first-party or same-site context.Wrapper around http.SameSite."`
	Domain         string   `json:"domain" pflag:",OPTIONAL: Allows you to set the domain attribute on the auth cookies."`
}

type OpenIDOptions struct {
	// The client ID for Admin in your IDP
	// See https://tools.ietf.org/html/rfc6749#section-2.2 for more information
	ClientID string `json:"clientId"`

	// The client secret used in the exchange of the authorization code for the token.
	// https://tools.ietf.org/html/rfc6749#section-2.3
	ClientSecretName string `json:"clientSecretName"`

	// Deprecated: Please use ClientSecretName instead.
	DeprecatedClientSecretFile string `json:"clientSecretFile"`

	// This should be the base url of the authorization server that you are trying to hit. With Okta for instance, it
	// will look something like https://company.okta.com/oauth2/abcdef123456789/
	BaseURL config.URL `json:"baseUrl"`

	// Provides a list of scopes to request from the IDP when authenticating. Default value requests claims that should
	// be supported by any OIdC server. Refer to https://openid.net/specs/openid-connect-core-1_0.html#ScopeClaims for
	// a complete list. Other providers might support additional scopes that you can define in a config.
	Scopes []string `json:"scopes"`
}

func GetConfig() *Config {
	return cfgSection.GetConfig().(*Config)
}

// MustParseURL panics if the provided url fails parsing. Should only be used in package initialization or tests.
func MustParseURL(rawURL string) *url.URL {
	res, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}

	return res
}
