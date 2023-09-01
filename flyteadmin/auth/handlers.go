package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flytestdlib/errors"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/runtime/protoiface"

	"github.com/flyteorg/flyteadmin/auth/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/common"
	"github.com/flyteorg/flyteadmin/plugins"
)

const (
	RedirectURLParameter = "redirect_url"
	FromHTTPKey          = "from_http"
	FromHTTPVal          = "true"
)

type PreRedirectHookError struct {
	Message string
	Code    int
}

func (e *PreRedirectHookError) Error() string {
	return e.Message
}

// PreRedirectHookFunc Interface used for running custom code before the redirect happens during a successful auth flow.
// This might be useful in cases where the auth flow allows the user to login since the IDP has been configured
// for eg: to allow all users from a particular domain to login
// but you want to restrict access to only a particular set of user ids. eg : users@domain.com are allowed to login but user user1@domain.com, user2@domain.com
// should only be allowed
// PreRedirectHookError is the error interface which allows the user to set correct http status code and Message to be set in case the function returns an error
// without which the current usage in GetCallbackHandler will set this to InternalServerError
type PreRedirectHookFunc func(ctx context.Context, authCtx interfaces.AuthenticationContext, request *http.Request, w http.ResponseWriter) *PreRedirectHookError
type HTTPRequestToMetadataAnnotator func(ctx context.Context, request *http.Request) metadata.MD
type UserInfoForwardResponseHandler func(ctx context.Context, w http.ResponseWriter, m protoiface.MessageV1) error

type AuthenticatedClientMeta struct {
	ClientIds     []string
	TokenIssuedAt time.Time
	ClientIP      string
	Subject       string
}

func RegisterHandlers(ctx context.Context, handler interfaces.HandlerRegisterer, authCtx interfaces.AuthenticationContext, pluginRegistry *plugins.Registry) {
	// Add HTTP handlers for OAuth2 endpoints
	handler.HandleFunc("/login", RefreshTokensIfExists(ctx, authCtx,
		GetLoginHandler(ctx, authCtx)))
	handler.HandleFunc("/callback", GetCallbackHandler(ctx, authCtx, pluginRegistry))

	// The metadata endpoint is an RFC-defined constant, but we need a leading / for the handler to pattern match correctly.
	handler.HandleFunc(fmt.Sprintf("/%s", OIdCMetadataEndpoint), GetOIdCMetadataEndpointRedirectHandler(ctx, authCtx))

	// These endpoints require authentication
	handler.HandleFunc("/logout", GetLogoutEndpointHandler(ctx, authCtx))
}

// Look for access token and refresh token, if both are present and the access token is expired, then attempt to
// refresh. Otherwise do nothing and proceed to the next handler. If successfully refreshed, proceed to the landing page.
func RefreshTokensIfExists(ctx context.Context, authCtx interfaces.AuthenticationContext, authHandler http.HandlerFunc) http.HandlerFunc {

	return func(writer http.ResponseWriter, request *http.Request) {
		ctx = context.WithValue(ctx, oauth2.HTTPClient, authCtx.GetHTTPClient())
		// Since we only do one thing if there are no errors anywhere along the chain, we can save code by just
		// using one variable and checking for errors at the end.
		idToken, accessToken, refreshToken, err := authCtx.CookieManager().RetrieveTokenValues(ctx, request)
		if err != nil {
			logger.Errorf(ctx, "Failed to retrieve tokens from request, redirecting to login handler. Error: %s", err)
			authHandler(writer, request)
			return
		}

		_, err = ParseIDTokenAndValidate(ctx, authCtx.Options().UserAuth.OpenID.ClientID, idToken, authCtx.OidcProvider())
		if err != nil && errors.IsCausedBy(err, ErrTokenExpired) && len(refreshToken) > 0 {
			logger.Debugf(ctx, "Expired id token found, attempting to refresh")
			newToken, err := GetRefreshedToken(ctx, authCtx.OAuth2ClientConfig(GetPublicURL(ctx, request, authCtx.Options())), accessToken, refreshToken)
			if err != nil {
				logger.Infof(ctx, "Failed to refresh tokens. Restarting login flow. Error: %s", err)
				authHandler(writer, request)
				return
			}

			logger.Debugf(ctx, "Tokens are refreshed. Saving new tokens into cookies.")
			err = authCtx.CookieManager().SetTokenCookies(ctx, writer, newToken)
			if err != nil {
				logger.Infof(ctx, "Failed to set token cookies. Restarting login flow. Error: %s", err)
				authHandler(writer, request)
				return
			}

			userInfo, err := QueryUserInfoUsingAccessToken(ctx, request, authCtx, newToken.AccessToken)
			if err != nil {
				logger.Infof(ctx, "Failed to query user info. Restarting login flow. Error: %s", err)
				authHandler(writer, request)
				return
			}

			err = authCtx.CookieManager().SetUserInfoCookie(ctx, writer, userInfo)
			if err != nil {
				logger.Infof(ctx, "Failed to set user info cookie. Restarting login flow. Error: %s", err)
				authHandler(writer, request)
				return
			}
		} else if err != nil {
			logger.Infof(ctx, "Failed to validate tokens. Restarting login flow. Error: %s", err)
			authHandler(writer, request)
			return
		}

		redirectURL := getAuthFlowEndRedirect(ctx, authCtx, request)
		http.Redirect(writer, request, redirectURL, http.StatusTemporaryRedirect)
	}
}

// GetLoginHandler builds an http handler that handles authentication calls. Before redirecting to the authentication
// provider, it saves a cookie that contains the redirect url for after the authentication flow is done.
func GetLoginHandler(ctx context.Context, authCtx interfaces.AuthenticationContext) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		csrfCookie := NewCsrfCookie()
		csrfToken := csrfCookie.Value
		http.SetCookie(writer, &csrfCookie)

		state := HashCsrfState(csrfToken)
		logger.Debugf(ctx, "Setting CSRF state cookie to %s and state to %s\n", csrfToken, state)
		url := authCtx.OAuth2ClientConfig(GetPublicURL(ctx, request, authCtx.Options())).AuthCodeURL(state)
		queryParams := request.URL.Query()
		if flowEndRedirectURL := queryParams.Get(RedirectURLParameter); flowEndRedirectURL != "" {
			redirectCookie := NewRedirectCookie(ctx, flowEndRedirectURL)
			if redirectCookie != nil {
				http.SetCookie(writer, redirectCookie)
			} else {
				logger.Errorf(ctx, "Was not able to create a redirect cookie")
			}
		}
		http.Redirect(writer, request, url, http.StatusTemporaryRedirect)
	}
}

// GetCallbackHandler returns a handler that is called by the OIdC provider with the authorization code to complete
// the user authentication flow.
func GetCallbackHandler(ctx context.Context, authCtx interfaces.AuthenticationContext, pluginRegistry *plugins.Registry) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		logger.Debugf(ctx, "Running callback handler... for RequestURI %v", request.RequestURI)
		authorizationCode := request.FormValue(AuthorizationResponseCodeType)

		ctx = context.WithValue(ctx, oauth2.HTTPClient, authCtx.GetHTTPClient())

		err := VerifyCsrfCookie(ctx, request)
		if err != nil {
			logger.Errorf(ctx, "Invalid CSRF token cookie %s", err)
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

		token, err := authCtx.OAuth2ClientConfig(GetPublicURL(ctx, request, authCtx.Options())).Exchange(ctx, authorizationCode)
		if err != nil {
			logger.Errorf(ctx, "Error when exchanging code %s", err)
			writer.WriteHeader(http.StatusForbidden)
			return
		}

		err = authCtx.CookieManager().SetTokenCookies(ctx, writer, token)
		if err != nil {
			logger.Errorf(ctx, "Error setting encrypted JWT cookie %s", err)
			writer.WriteHeader(http.StatusForbidden)
			return
		}

		userInfo, err := QueryUserInfoUsingAccessToken(ctx, request, authCtx, token.AccessToken)
		if err != nil {
			logger.Errorf(ctx, "Failed to query user info. Error: %v", err)
			writer.WriteHeader(http.StatusForbidden)
			return
		}

		err = authCtx.CookieManager().SetUserInfoCookie(ctx, writer, userInfo)
		if err != nil {
			logger.Errorf(ctx, "Error setting encrypted user info cookie. Error: %v", err)
			writer.WriteHeader(http.StatusForbidden)
			return
		}

		preRedirectHook := plugins.Get[PreRedirectHookFunc](pluginRegistry, plugins.PluginIDPreRedirectHook)
		if preRedirectHook != nil {
			logger.Infof(ctx, "preRedirect hook is set")
			if err := preRedirectHook(ctx, authCtx, request, writer); err != nil {
				logger.Errorf(ctx, "failed the preRedirect hook due %v with status code  %v", err.Message, err.Code)
				if http.StatusText(err.Code) != "" {
					writer.WriteHeader(err.Code)
				} else {
					writer.WriteHeader(http.StatusInternalServerError)
				}
				return
			}
			logger.Info(ctx, "Successfully called the preRedirect hook")
		}
		redirectURL := getAuthFlowEndRedirect(ctx, authCtx, request)
		http.Redirect(writer, request, redirectURL, http.StatusTemporaryRedirect)
	}
}

func AuthenticationLoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Invoke 'handler' to use your gRPC server implementation and get
	// the response.
	identityContext := IdentityContextFromContext(ctx)
	var emailPlaceholder string
	if len(identityContext.UserInfo().GetEmail()) > 0 {
		emailPlaceholder = fmt.Sprintf(" (%s) ", identityContext.UserInfo().GetEmail())
	}
	logger.Debugf(ctx, "gRPC server info in logging interceptor [%s]%smethod [%s]\n", identityContext.UserID(), emailPlaceholder, info.FullMethod)
	return handler(ctx, req)
}

// GetAuthenticationCustomMetadataInterceptor produces a gRPC middleware interceptor intended to be used when running
// authentication with non-default gRPC headers (metadata). Because the default `authorization` header is reserved for
// use by Envoy, clients wishing to pass tokens to Admin will need to use a different string, specified in this
// package's Config object. This interceptor will scan for that arbitrary string, and then rename it to the default
// string, which the downstream auth/auditing interceptors will detect and validate.
func GetAuthenticationCustomMetadataInterceptor(authCtx interfaces.AuthenticationContext) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if authCtx.Options().GrpcAuthorizationHeader != DefaultAuthorizationHeader {
			md, ok := metadata.FromIncomingContext(ctx)
			if ok {
				existingHeader := md.Get(authCtx.Options().GrpcAuthorizationHeader)
				if len(existingHeader) > 0 {
					logger.Debugf(ctx, "Found existing metadata %s", existingHeader[0])
					newAuthorizationMetadata := metadata.Pairs(DefaultAuthorizationHeader, existingHeader[0])
					joinedMetadata := metadata.Join(md, newAuthorizationMetadata)
					newCtx := metadata.NewIncomingContext(ctx, joinedMetadata)
					return handler(newCtx, req)
				}
			} else {
				logger.Debugf(ctx, "Could not extract incoming metadata from context, continuing with original ctx...")
			}
		}
		return handler(ctx, req)
	}
}

func SetContextForIdentity(ctx context.Context, identityContext interfaces.IdentityContext) context.Context {
	email := identityContext.UserInfo().GetEmail()
	newCtx := identityContext.WithContext(ctx)
	if len(email) > 0 {
		newCtx = WithUserEmail(newCtx, email)
	}

	return WithAuditFields(newCtx, identityContext.UserID(), []string{identityContext.AppID()}, identityContext.AuthenticatedAt())
}

// GetAuthenticationInterceptor chooses to enforce or not enforce authentication. It will attempt to get the token
// from the incoming context, validate it, and decide whether or not to let the request through.
func GetAuthenticationInterceptor(authCtx interfaces.AuthenticationContext) func(context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		logger.Debugf(ctx, "Running authentication gRPC interceptor")

		fromHTTP := metautils.ExtractIncoming(ctx).Get(FromHTTPKey)
		isFromHTTP := fromHTTP == FromHTTPVal

		identityContext, err := GRPCGetIdentityFromAccessToken(ctx, authCtx)
		if err == nil {
			return SetContextForIdentity(ctx, identityContext), nil
		}

		logger.Infof(ctx, "Failed to parse Access Token from context. Will attempt to find IDToken. Error: %v", err)

		identityContext, err = GRPCGetIdentityFromIDToken(ctx, authCtx.Options().UserAuth.OpenID.ClientID,
			authCtx.OidcProvider())

		if err == nil {
			return SetContextForIdentity(ctx, identityContext), nil
		}

		// Only enforcement logic is present. The default case is to let things through.
		if (isFromHTTP && !authCtx.Options().DisableForHTTP) ||
			(!isFromHTTP && !authCtx.Options().DisableForGrpc) {
			return ctx, status.Errorf(codes.Unauthenticated, "token parse error %s", err)
		}

		return ctx, nil
	}
}

func WithUserEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, common.PrincipalContextKey, email)
}

func WithAuditFields(ctx context.Context, subject string, clientIds []string, tokenIssuedAt time.Time) context.Context {
	var clientIP string
	peerInfo, ok := peer.FromContext(ctx)
	if ok {
		clientIP = peerInfo.Addr.String()
	}
	return context.WithValue(ctx, common.AuditFieldsContextKey, AuthenticatedClientMeta{
		ClientIds:     clientIds,
		TokenIssuedAt: tokenIssuedAt,
		ClientIP:      clientIP,
		Subject:       subject,
	})
}

// This is effectively middleware for the grpc gateway, it allows us to modify the translation between HTTP request
// and gRPC request. There are two potential sources for bearer tokens, it can come from an authorization header (not
// yet implemented), or encrypted cookies. Note that when deploying behind Envoy, you have the option to look for a
// configurable, non-standard header name. The token is extracted and turned into a metadata object which is then
// attached to the request, from which the token is extracted later for verification.
func GetHTTPRequestCookieToMetadataHandler(authCtx interfaces.AuthenticationContext) HTTPRequestToMetadataAnnotator {
	return func(ctx context.Context, request *http.Request) metadata.MD {
		// TODO: Improve error handling
		idToken, _, _, _ := authCtx.CookieManager().RetrieveTokenValues(ctx, request)
		if len(idToken) == 0 {
			// If no token was found in the cookies, look for an authorization header, starting with a potentially
			// custom header set in the Config object
			if len(authCtx.Options().HTTPAuthorizationHeader) > 0 {
				header := authCtx.Options().HTTPAuthorizationHeader
				// TODO: There may be a potential issue here when running behind a service mesh that uses the default Authorization
				//       header. The grpc-gateway code will automatically translate the 'Authorization' header into the appropriate
				//       metadata object so if two different tokens are presented, one with the default name and one with the
				//       custom name, AuthFromMD will find the wrong one.
				return metadata.MD{
					DefaultAuthorizationHeader: []string{request.Header.Get(header)},
				}
			}

			logger.Infof(ctx, "Could not find access token cookie while requesting %s", request.RequestURI)
			return nil
		}

		// IDtoken is injected into grpc authorization metadata
		meta := metadata.MD{
			DefaultAuthorizationHeader: []string{fmt.Sprintf("%s %s", IDTokenScheme, idToken)},
		}

		userInfo, err := authCtx.CookieManager().RetrieveUserInfo(ctx, request)
		if err != nil {
			logger.Infof(ctx, "Failed to retrieve user info cookie. Ignoring. Error: %v", err)
		}

		raw, err := json.Marshal(userInfo)
		if err != nil {
			logger.Infof(ctx, "Failed to marshal user info. Ignoring. Error: %v", err)
		}

		if len(raw) > 0 {
			meta.Set(UserInfoMDKey, string(raw))
		}

		return meta
	}
}

// Intercepts the incoming HTTP requests and marks it as such so that the downstream code can use it to enforce auth.
// See the enforceHTTP/Grpc options for more information.
func GetHTTPMetadataTaggingHandler() HTTPRequestToMetadataAnnotator {
	return func(ctx context.Context, request *http.Request) metadata.MD {
		return metadata.MD{
			FromHTTPKey: []string{FromHTTPVal},
		}
	}
}

func IdentityContextFromRequest(ctx context.Context, req *http.Request, authCtx interfaces.AuthenticationContext) (
	interfaces.IdentityContext, error) {

	authHeader := DefaultAuthorizationHeader
	if len(authCtx.Options().HTTPAuthorizationHeader) > 0 {
		authHeader = authCtx.Options().HTTPAuthorizationHeader
	}

	headerValue := req.Header.Get(authHeader)
	if len(headerValue) == 0 {
		headerValue = req.Header.Get(DefaultAuthorizationHeader)
	}

	if len(headerValue) > 0 {
		logger.Debugf(ctx, "Found authorization header at [%v] header. Validating.", authHeader)
		if strings.HasPrefix(headerValue, BearerScheme+" ") {
			expectedAudience := GetPublicURL(ctx, req, authCtx.Options()).String()
			return authCtx.OAuth2ResourceServer().ValidateAccessToken(ctx, expectedAudience, strings.TrimPrefix(headerValue, BearerScheme+" "))
		}
	}

	idToken, _, _, err := authCtx.CookieManager().RetrieveTokenValues(ctx, req)
	if err != nil || len(idToken) == 0 {
		return nil, fmt.Errorf("unauthenticated request. IDToken Len [%v], Error: %w", len(idToken), err)
	}

	userInfo, err := authCtx.CookieManager().RetrieveUserInfo(ctx, req)
	if err != nil {
		logger.Infof(ctx, "Failed to retrieve user info cookie from request. Error: %v", err)
		return nil, fmt.Errorf("unauthenticated request. Error: %w", err)
	}

	return IdentityContextFromIDTokenToken(ctx, idToken, authCtx.Options().UserAuth.OpenID.ClientID,
		authCtx.OidcProvider(), userInfo)
}

func QueryUserInfo(ctx context.Context, identityContext interfaces.IdentityContext, request *http.Request,
	authCtx interfaces.AuthenticationContext) (*service.UserInfoResponse, error) {

	if identityContext.IsEmpty() {
		return &service.UserInfoResponse{}, fmt.Errorf("request is unauthenticated, try /login in again")
	}

	if len(identityContext.UserInfo().GetName()) > 0 {
		return identityContext.UserInfo(), nil
	}

	_, accessToken, _, err := authCtx.CookieManager().RetrieveTokenValues(ctx, request)
	if err != nil {
		return &service.UserInfoResponse{}, fmt.Errorf("error decoding identify token, try /login in again")
	}

	return QueryUserInfoUsingAccessToken(ctx, request, authCtx, accessToken)
}

// Extract User info from access token for HTTP request
func QueryUserInfoUsingAccessToken(ctx context.Context, originalRequest *http.Request, authCtx interfaces.AuthenticationContext, accessToken string) (
	*service.UserInfoResponse, error) {

	originalToken := oauth2.Token{
		AccessToken: accessToken,
	}

	tokenSource := authCtx.OAuth2ClientConfig(GetPublicURL(ctx, originalRequest, authCtx.Options())).TokenSource(ctx, &originalToken)

	// TODO: Investigate improving transparency of errors. The errors from this call may be just a local error, or may
	//       be an error from the HTTP request to the IDP. In the latter case, consider passing along the error code/msg.
	userInfo, err := authCtx.OidcProvider().UserInfo(ctx, tokenSource)
	if err != nil {
		logger.Errorf(ctx, "Error getting user info from IDP %s", err)
		return &service.UserInfoResponse{}, fmt.Errorf("error getting user info from IDP")
	}

	resp := &service.UserInfoResponse{}
	err = userInfo.Claims(&resp)
	if err != nil {
		logger.Errorf(ctx, "Error getting user info from IDP %s", err)
		return &service.UserInfoResponse{}, fmt.Errorf("error getting user info from IDP")
	}

	return resp, err
}

// This returns a handler that will redirect (303) to the well-known metadata endpoint for the OAuth2 authorization server
// See https://tools.ietf.org/html/rfc8414 for more information.
func GetOIdCMetadataEndpointRedirectHandler(ctx context.Context, authCtx interfaces.AuthenticationContext) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		metadataURL := authCtx.Options().UserAuth.OpenID.BaseURL.ResolveReference(authCtx.GetOIdCMetadataURL())
		http.Redirect(writer, request, metadataURL.String(), http.StatusSeeOther)
	}
}

func GetLogoutEndpointHandler(ctx context.Context, authCtx interfaces.AuthenticationContext) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		logger.Debugf(ctx, "Deleting auth cookies")
		authCtx.CookieManager().DeleteCookies(ctx, writer)

		// Redirect if one was given
		queryParams := request.URL.Query()
		if redirectURL := queryParams.Get(RedirectURLParameter); redirectURL != "" {
			http.Redirect(writer, request, redirectURL, http.StatusTemporaryRedirect)
		}
	}
}

func GetUserInfoForwardResponseHandler() UserInfoForwardResponseHandler {
	return func(ctx context.Context, w http.ResponseWriter, m protoiface.MessageV1) error {
		info, ok := m.(*service.UserInfoResponse)
		if ok {
			if info.AdditionalClaims != nil {
				for k, v := range info.AdditionalClaims.GetFields() {
					jsonBytes, err := v.MarshalJSON()
					if err != nil {
						logger.Warningf(ctx, "failed to marshal claim [%s] to json: %v", k, err)
						continue
					}
					header := fmt.Sprintf("X-User-Claim-%s", strings.ReplaceAll(k, "_", "-"))
					w.Header().Set(header, string(jsonBytes))
				}
			}
			w.Header().Set("X-User-Subject", info.Subject)
		}
		return nil
	}
}
