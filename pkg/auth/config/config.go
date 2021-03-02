package config

type OAuthOptions struct {
	// The client ID for Admin in your IDP
	// See https://tools.ietf.org/html/rfc6749#section-2.2 for more information
	ClientID string `json:"clientId"`

	// The client secret used in the exchange of the authorization code for the token.
	// https://tools.ietf.org/html/rfc6749#section-2.3
	ClientSecretFile string `json:"clientSecretFile"`

	// This should be the base url of the authorization server that you are trying to hit. With Okta for instance, it
	// will look something like https://company.okta.com/oauth2/abcdef123456789/
	// TODO: Convert all the URLs in this config to the config.URL type
	BaseURL string `json:"baseUrl"`

	// These two config elements currently need the entire path, including the already specified baseUrl
	// TODO: Refactor to ascertain the paths using discovery (see https://tools.ietf.org/html/rfc8414)
	//       Also refactor to use relative paths when discovery is not available
	AuthorizeURL string `json:"authorizeUrl"`
	TokenURL     string `json:"tokenUrl"`

	// This is the callback URL that will be sent to the IDP authorize endpoint. It is likely that your IDP application
	// needs to have this URL whitelisted before using.
	CallbackURL string `json:"callbackUrl"`
	Claims      Claims `json:"claims"`

	// This is the relative path of the user info endpoint, if there is one, for the given IDP. This will be appended to
	// the base URL of the IDP. This is used to support the /me endpoint that Admin will serve when running with authentication
	// See https://developer.okta.com/docs/reference/api/oidc/#userinfo as an example.
	IdpUserInfoEndpoint string `json:"idpUserInfoEndpoint"`

	// These should point to files that contain base64 encoded secrets.
	// You can run `go test -v github.com/lyft/flyteadmin/pkg/auth -run TestSecureCookieLifecycle` to generate new ones.
	// See https://github.com/gorilla/securecookie#examples for more information
	CookieHashKeyFile  string `json:"cookieHashKeyFile"`
	CookieBlockKeyFile string `json:"cookieBlockKeyFile"`

	// This is where the user will be redirected to at the end of the flow, but you should not use it. Instead,
	// the initial /login handler should be called with a redirect_url parameter, which will get saved to a cookie.
	// This setting will only be used when that cookie is missing.
	// See the login handler code for more comments.
	RedirectURL string `json:"redirectUrl"`

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
	DisableForHTTP bool `json:"disableForHttp"`
	DisableForGrpc bool `json:"disableForGrpc"`

	// Provides a list of scopes to request from the IDP when authenticating. Default value requests claims that should
	// be supported by any OIdC server. Refer to https://openid.net/specs/openid-connect-core-1_0.html#ScopeClaims for
	// a complete list. Other providers might support additional scopes that you can define in a config.
	Scopes []string `json:"scopes"`
}

type Claims struct {
	Audience string `json:"aud"`
	Issuer   string `json:"iss"`
}
