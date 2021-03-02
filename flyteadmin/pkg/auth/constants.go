package auth

// OAuth2 Parameters
const CsrfFormKey = "state"
const AuthorizationResponseCodeType = "code"
const RefreshToken = "refresh_token"
const DefaultAuthorizationHeader = "authorization"
const BearerScheme = "Bearer"

// https://tools.ietf.org/html/rfc8414
// This should be defined without a leading slash. If there is one, the url library's ResolveReference will make it a root path
const OAuth2MetadataEndpoint = ".well-known/oauth-authorization-server"

// https://openid.net/specs/openid-connect-discovery-1_0.html
// This should be defined without a leading slash. If there is one, the url library's ResolveReference will make it a root path
const OIdCMetadataEndpoint = ".well-known/openid-configuration"
