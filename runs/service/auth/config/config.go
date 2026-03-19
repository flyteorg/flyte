package config

import (
	"net/url"

	"github.com/flyteorg/flyte/v2/flytestdlib/config"
)

//go:generate pflags Config --default-var=DefaultConfig

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

	// SecretNameClaimSymmetricKey must be a base64 encoded secret of exactly 32 bytes.
	// #nosec
	SecretNameClaimSymmetricKey SecretName = "claim_symmetric_key"

	// SecretNameTokenSigningRSAKey is the private key used to sign JWT tokens (RS256).
	// #nosec
	SecretNameTokenSigningRSAKey SecretName = "token_rsa_key.pem"

	// SecretNameOldTokenSigningRSAKey is the old private key for key rotation. Only used to
	// validate incoming tokens; new tokens will not be issued with this key.
	// #nosec
	SecretNameOldTokenSigningRSAKey SecretName = "token_rsa_key_old.pem"
)

// AuthorizationServerType defines the type of Authorization Server to use.
type AuthorizationServerType int

const (
	// AuthorizationServerTypeSelf indicates the service acts as its own authorization server.
	AuthorizationServerTypeSelf AuthorizationServerType = iota
	// AuthorizationServerTypeExternal indicates an external authorization server is used.
	AuthorizationServerTypeExternal
)

// SameSite represents the SameSite cookie policy.
type SameSite int

const (
	SameSiteDefaultMode SameSite = iota
	SameSiteLaxMode
	SameSiteStrictMode
	SameSiteNoneMode
)

var (
	DefaultConfig = &Config{
		HTTPAuthorizationHeader: "flyte-authorization",
		GrpcAuthorizationHeader: "flyte-authorization",
		UserAuth: UserAuthConfig{
			RedirectURL:              config.URL{URL: *MustParseURL("/console")},
			CookieHashKeySecretName:  SecretNameCookieHashKey,
			CookieBlockKeySecretName: SecretNameCookieBlockKey,
			OpenID: OpenIDOptions{
				ClientSecretName: SecretNameOIdCClientSecret,
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
			ExternalAuthServer: ExternalAuthorizationServer{
				RetryAttempts: 5,
				RetryDelay:    config.Duration{Duration: 1_000_000_000}, // 1 second
			},
			AuthServerType: AuthorizationServerTypeSelf,
			ThirdParty: ThirdPartyConfigOptions{
				FlyteClientConfig: FlyteClientConfig{
					ClientID:    "flytectl",
					RedirectURI: "http://localhost:53593/callback",
					Scopes:      []string{"all", "offline"},
				},
			},
		},
	}

	cfgSection = config.MustRegisterSection("auth", DefaultConfig)
)

// Config holds the full authentication configuration.
type Config struct {
	// HTTPAuthorizationHeader is the HTTP header name for authorization (for non-standard headers behind Envoy).
	HTTPAuthorizationHeader string `json:"httpAuthorizationHeader"`

	// GrpcAuthorizationHeader is the gRPC metadata key for authorization.
	GrpcAuthorizationHeader string `json:"grpcAuthorizationHeader"`

	// DisableForHTTP disables auth enforcement on HTTP endpoints.
	DisableForHTTP bool `json:"disableForHttp" pflag:",Disables auth enforcement on HTTP Endpoints."`

	// DisableForGrpc disables auth enforcement on gRPC endpoints.
	DisableForGrpc bool `json:"disableForGrpc" pflag:",Disables auth enforcement on Grpc Endpoints."`

	// AuthorizedURIs defines the set of URIs that clients are allowed to visit the service on.
	AuthorizedURIs []config.URL `json:"authorizedUris" pflag:"-,Defines the set of URIs that clients are allowed to visit the service on."`

	// HTTPProxyURL allows accessing external OAuth2 servers through an HTTP proxy.
	HTTPProxyURL config.URL `json:"httpProxyURL" pflag:",OPTIONAL: HTTP Proxy to be used for OAuth requests."`

	// UserAuth settings used to authenticate end users in web-browsers.
	UserAuth UserAuthConfig `json:"userAuth" pflag:",Defines Auth options for users."`

	// AppAuth defines app-level OAuth2 settings.
	AppAuth OAuth2Options `json:"appAuth" pflag:",Defines Auth options for apps. UserAuth must be enabled for AppAuth to work."`

	// SecureCookie sets the Secure flag on auth cookies. Should be true in production (HTTPS).
	SecureCookie bool `json:"secureCookie" pflag:",Set the Secure flag on auth cookies"`

	// TokenEndpointProxyConfig proxies token endpoint calls through admin.
	TokenEndpointProxyConfig TokenEndpointProxyConfig `json:"tokenEndpointProxyConfig" pflag:",Configuration for proxying token endpoint requests."`
}

// OAuth2Options holds OAuth2 authorization server options.
type OAuth2Options struct {
	// AuthServerType determines whether to use a self-hosted or external auth server.
	AuthServerType AuthorizationServerType `json:"authServerType"`

	// SelfAuthServer configures the self-hosted authorization server.
	SelfAuthServer AuthorizationServer `json:"selfAuthServer" pflag:",Authorization Server config to run as a service."`

	// ExternalAuthServer configures the external authorization server.
	ExternalAuthServer ExternalAuthorizationServer `json:"externalAuthServer" pflag:",External Authorization Server config."`

	// ThirdParty configures third-party (public client) settings.
	ThirdParty ThirdPartyConfigOptions `json:"thirdPartyConfig" pflag:",Defines settings to instruct flyte cli tools on what config to use."`
}

// AuthorizationServer configures a self-hosted authorization server.
type AuthorizationServer struct {
	// Issuer is the issuer URL. If empty, the first AuthorizedURI is used.
	Issuer string `json:"issuer" pflag:",Defines the issuer to use when issuing and validating tokens."`

	// AccessTokenLifespan defines the lifespan of issued access tokens.
	AccessTokenLifespan config.Duration `json:"accessTokenLifespan" pflag:",Defines the lifespan of issued access tokens."`

	// RefreshTokenLifespan defines the lifespan of issued refresh tokens.
	RefreshTokenLifespan config.Duration `json:"refreshTokenLifespan" pflag:",Defines the lifespan of issued refresh tokens."`

	// AuthorizationCodeLifespan defines the lifespan of issued authorization codes.
	AuthorizationCodeLifespan config.Duration `json:"authorizationCodeLifespan" pflag:",Defines the lifespan of issued authorization codes."`

	// ClaimSymmetricEncryptionKeySecretName is the secret name for claim encryption.
	ClaimSymmetricEncryptionKeySecretName string `json:"claimSymmetricEncryptionKeySecretName" pflag:",Secret name for claim encryption key."`

	// TokenSigningRSAKeySecretName is the secret name for the RSA signing key.
	TokenSigningRSAKeySecretName string `json:"tokenSigningRSAKeySecretName" pflag:",Secret name for RSA Signing Key."`

	// OldTokenSigningRSAKeySecretName is the secret name for the old RSA signing key (key rotation).
	OldTokenSigningRSAKeySecretName string `json:"oldTokenSigningRSAKeySecretName" pflag:",Secret name for Old RSA Signing Key for key rotation."`
}

// ExternalAuthorizationServer configures an external authorization server.
type ExternalAuthorizationServer struct {
	// BaseURL is the base URL of the external authorization server.
	BaseURL config.URL `json:"baseUrl" pflag:",Base url of the external authorization server."`

	// AllowedAudience is the set of audiences accepted when validating access tokens.
	AllowedAudience []string `json:"allowedAudience" pflag:",A list of allowed audiences."`

	// MetadataEndpointURL overrides the default .well-known/oauth-authorization-server endpoint.
	MetadataEndpointURL config.URL `json:"metadataUrl" pflag:",Custom metadata url if the server doesn't support the standard endpoint."`

	// HTTPProxyURL allows accessing the external auth server through an HTTP proxy.
	HTTPProxyURL config.URL `json:"httpProxyURL" pflag:",HTTP Proxy for external OAuth requests."`

	// RetryAttempts is the number of retry attempts for fetching metadata.
	RetryAttempts int `json:"retryAttempts" pflag:",Number of retry attempts for metadata fetch."`

	// RetryDelay is the delay between retry attempts.
	RetryDelay config.Duration `json:"retryDelay" pflag:",Duration to wait between retries."`
}

// ThirdPartyConfigOptions holds third-party OAuth2 client settings.
type ThirdPartyConfigOptions struct {
	// FlyteClientConfig holds public client configuration.
	FlyteClientConfig FlyteClientConfig `json:"flyteClient"`
}

// IsEmpty returns true if the third-party config has no meaningful values set.
func (o ThirdPartyConfigOptions) IsEmpty() bool {
	return len(o.FlyteClientConfig.ClientID) == 0 &&
		len(o.FlyteClientConfig.RedirectURI) == 0 &&
		len(o.FlyteClientConfig.Scopes) == 0
}

// FlyteClientConfig holds the public client configuration.
type FlyteClientConfig struct {
	// ClientID is the public client ID.
	ClientID string `json:"clientId" pflag:",Public identifier for the app which handles authorization."`

	// RedirectURI is the redirect URI for the client.
	RedirectURI string `json:"redirectUri" pflag:",Callback uri registered with the app which handles authorization."`

	// Scopes are the OAuth2 scopes to request.
	Scopes []string `json:"scopes" pflag:",Recommended scopes for the client to request."`

	// Audience is the intended audience for OAuth2 tokens.
	Audience string `json:"audience" pflag:",Audience to use when initiating OAuth2 authorization requests."`
}

// UserAuthConfig holds user authentication settings (browser-based OAuth2/OIDC flows).
type UserAuthConfig struct {
	// RedirectURL is the default redirect URL after the OAuth2 flow completes.
	RedirectURL config.URL `json:"redirectUrl"`

	// OpenID defines settings for connecting and trusting an OpenID Connect provider.
	OpenID OpenIDOptions `json:"openId" pflag:",OpenID Configuration for User Auth"`

	// HTTPProxyURL allows operators to access external OAuth2 servers using an HTTP Proxy.
	HTTPProxyURL config.URL `json:"httpProxyURL" pflag:",HTTP Proxy for OAuth requests."`

	// CookieHashKeySecretName is the secret name for the cookie hash key.
	CookieHashKeySecretName string `json:"cookieHashKeySecretName" pflag:",Secret name for cookie hash key."`

	// CookieBlockKeySecretName is the secret name for the cookie block key.
	CookieBlockKeySecretName string `json:"cookieBlockKeySecretName" pflag:",Secret name for cookie block key."`

	// CookieSetting configures cookie behavior.
	CookieSetting CookieSettings `json:"cookieSetting" pflag:",Settings for auth cookies."`

	// IDPQueryParameter is used to select a particular IDP for user authentication.
	IDPQueryParameter string `json:"idpQueryParameter" pflag:",IDP query parameter for selecting a particular IDP."`
}

// OpenIDOptions holds OpenID Connect provider configuration.
type OpenIDOptions struct {
	// ClientID is the client ID for this service in the IDP.
	ClientID string `json:"clientId"`

	// ClientSecretName is the secret name containing the OIDC client secret.
	ClientSecretName string `json:"clientSecretName"`

	// BaseURL is the base URL of the OIDC provider.
	BaseURL config.URL `json:"baseUrl"`

	// Scopes to request from the IDP when authenticating.
	Scopes []string `json:"scopes"`
}

// CookieSettings configures cookie behavior.
type CookieSettings struct {
	// SameSitePolicy controls the SameSite attribute on auth cookies.
	SameSitePolicy SameSite `json:"sameSitePolicy" pflag:",SameSite policy for auth cookies."`

	// Domain sets the domain attribute on auth cookies.
	Domain string `json:"domain" pflag:",Domain attribute on auth cookies."`
}

// TokenEndpointProxyConfig configures proxying of token endpoint calls.
type TokenEndpointProxyConfig struct {
	// Enabled enables token endpoint proxying.
	Enabled bool `json:"enabled" pflag:",Enables the token endpoint proxy."`

	// PublicURL is the public URL to use for rewriting the token endpoint.
	PublicURL config.URL `json:"publicUrl" pflag:",Public URL for the token endpoint proxy."`

	// PathPrefix is appended to the public URL when rewriting.
	PathPrefix string `json:"pathPrefix" pflag:",Path prefix for proxying token requests."`
}

// URL is an alias for flytestdlib config.URL, re-exported for convenience.
type URL = config.URL

// Duration is an alias for flytestdlib config.Duration, re-exported for convenience.
type Duration = config.Duration

// GetConfig returns the parsed auth configuration.
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
