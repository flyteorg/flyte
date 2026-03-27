package auth

import "github.com/flyteorg/flyte/flytestdlib/contextutils"

const (
	// OAuth2 Parameters
	CsrfFormKey                   = "state"
	AuthorizationResponseCodeType = "code"
	DefaultAuthorizationHeader    = "authorization"
	GRPCMetaKeyAccessToken        = "access-token"
	GRPCMetaKeyRefreshToken       = "refresh-token"
	GRPCMetaKeyIDToken            = "id-token"
	GRPCMetaKeyUserInfo           = "user-info"
	BearerScheme                  = "Bearer"
	IDTokenScheme                 = "IDToken"
	// Add the -bin suffix so that the header value is automatically base64 encoded
	UserInfoMDKey = "UserInfo-bin"

	// https://tools.ietf.org/html/rfc8414
	// This should be defined without a leading slash. If there is one, the url library's ResolveReference will make it a root path
	OAuth2MetadataEndpoint = ".well-known/oauth-authorization-server"

	// https://openid.net/specs/openid-connect-discovery-1_0.html
	// This should be defined without a leading slash. If there is one, the url library's ResolveReference will make it a root path
	OIdCMetadataEndpoint = ".well-known/openid-configuration"

	ContextKeyIdentityContext = contextutils.Key("identity_context")
	ScopeAll                  = "all"

	// UserTokenHeader is the HTTP header used to forward the user's raw OIDC
	// token from the /me auth subrequest through nginx to backend services.
	// This enables external authorization servers to receive the token for
	// downstream operations like token exchange.
	UserTokenHeader = "X-User-Token" //nolint:gosec // Not a credential, just a header name
)
