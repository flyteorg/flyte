package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	grpcauth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/lyft/flyteadmin/pkg/auth/interfaces"
	"github.com/lyft/flytestdlib/contextutils"
	"github.com/lyft/flytestdlib/errors"
	"github.com/lyft/flytestdlib/logger"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	RedirectURLParameter                   = "redirect_url"
	bearerTokenContextKey contextutils.Key = "bearer"
	PrincipalContextKey   contextutils.Key = "principal"
)

type HTTPRequestToMetadataAnnotator func(ctx context.Context, request *http.Request) metadata.MD

// Look for access token and refresh token, if both are present and the access token is expired, then attempt to
// refresh. Otherwise do nothing and proceed to the next handler. If successfully refreshed, proceed to the landing page.
func RefreshTokensIfExists(ctx context.Context, authContext interfaces.AuthenticationContext, handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		// Since we only do one thing if there are no errors anywhere along the chain, we can save code by just
		// using one variable and checking for errors at the end.
		var err error
		accessToken, refreshToken, err := authContext.CookieManager().RetrieveTokenValues(ctx, request)

		if err == nil && accessToken != "" && refreshToken != "" {
			_, e := ParseAndValidate(ctx, authContext.Claims(), accessToken, authContext.OidcProvider())
			err = e
			if err != nil && errors.IsCausedBy(err, ErrTokenExpired) {
				logger.Debugf(ctx, "Expired access token found, attempting to refresh")
				newToken, e := GetRefreshedToken(ctx, authContext.OAuth2Config(), accessToken, refreshToken)
				err = e
				if err == nil {
					logger.Debugf(ctx, "Access token refreshed. Saving new tokens into cookies.")
					err = authContext.CookieManager().SetTokenCookies(ctx, writer, newToken)
				}
			}
		}

		if err != nil {
			logger.Errorf(ctx, "Non-expiration error in refresh token handler %s, redirecting to login handler", err)
			handlerFunc(writer, request)
			return
		}
		redirectURL := getAuthFlowEndRedirect(ctx, authContext, request)
		http.Redirect(writer, request, redirectURL, http.StatusTemporaryRedirect)
	}
}

func GetLoginHandler(ctx context.Context, authContext interfaces.AuthenticationContext) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		csrfCookie := NewCsrfCookie()
		csrfToken := csrfCookie.Value
		http.SetCookie(writer, &csrfCookie)

		state := HashCsrfState(csrfToken)
		logger.Debugf(ctx, "Setting CSRF state cookie to %s and state to %s\n", csrfToken, state)
		url := authContext.OAuth2Config().AuthCodeURL(state)
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

func GetCallbackHandler(ctx context.Context, authContext interfaces.AuthenticationContext) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		logger.Debugf(ctx, "Running callback handler...")
		authorizationCode := request.FormValue(AuthorizationResponseCodeType)

		err := VerifyCsrfCookie(ctx, request)
		if err != nil {
			logger.Errorf(ctx, "Invalid CSRF token cookie %s", err)
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

		// TODO: The second parameter is IDP specific but seems to be convention, make configurable anyways.
		// The second parameter is necessary to get the initial refresh token
		offlineAccessParam := oauth2.SetAuthURLParam(RefreshToken, OfflineAccessType)

		token, err := authContext.OAuth2Config().Exchange(ctx, authorizationCode, offlineAccessParam)
		if err != nil {
			logger.Errorf(ctx, "Error when exchanging code %s", err)
			writer.WriteHeader(http.StatusForbidden)
			return
		}

		err = authContext.CookieManager().SetTokenCookies(ctx, writer, token)
		if err != nil {
			logger.Errorf(ctx, "Error setting encrypted JWT cookie %s", err)
			writer.WriteHeader(http.StatusForbidden)
			return
		}

		redirectURL := getAuthFlowEndRedirect(ctx, authContext, request)
		http.Redirect(writer, request, redirectURL, http.StatusTemporaryRedirect)
	}
}

func AuthenticationLoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Invoke 'handler' to use your gRPC server implementation and get
	// the response.
	logger.Debugf(ctx, "gRPC server info in logging interceptor email %s method %s\n", ctx.Value(PrincipalContextKey), info.FullMethod)
	return handler(ctx, req)
}

// This function produces a gRPC middleware interceptor intended to be used when running authentication with non-default
// gRPC headers (metadata). Because the default `authorization` header is reserved for use by Envoy, clients wishing
// to pass tokens to Admin will need to use a different string, specified in this package's Config object. This interceptor
// will scan for that arbitrary string, and then rename it to the default string, which the downstream auth/auditing
// interceptors will detect and validate.
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

// This function will only look for a token from the request metadata, verify it, and extract the user email if valid.
// Unless there is an error, it will not return an unauthorized status. That is up to subsequent functions to decide,
// based on configuration.  We don't want to require authentication for all endpoints.
func GetAuthenticationInterceptor(authContext interfaces.AuthenticationContext) func(context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		logger.Debugf(ctx, "Running authentication gRPC interceptor")
		tokenStr, err := grpcauth.AuthFromMD(ctx, BearerScheme)
		if err != nil {
			logger.Debugf(ctx, "Could not retrieve bearer token from metadata %v", err)
			return ctx, nil
		}

		// Currently auth is optional...
		if tokenStr == "" {
			logger.Debugf(ctx, "Bearer token is empty, skipping parsing")
			return ctx, nil
		}

		// ...however, if there _is_ a bearer token, but there are additional errors downstream, then we return an
		// authentication error.
		token, err := ParseAndValidate(ctx, authContext.Claims(), tokenStr, authContext.OidcProvider())
		if err != nil {
			return ctx, status.Errorf(codes.Unauthenticated, "could not parse token string into object: %s %s", tokenStr, err)
		}
		if token == nil {
			return ctx, status.Errorf(codes.Unauthenticated, "Token was nil after parsing %s", tokenStr)
		} else if token.Subject == "" {
			return ctx, status.Errorf(codes.Unauthenticated, "no email or empty email found")
		} else {
			newCtx := WithUserEmail(context.WithValue(ctx, bearerTokenContextKey, tokenStr), token.Subject)
			return newCtx, nil
		}
	}
}

func WithUserEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, PrincipalContextKey, email)
}

// This is effectively middleware for the grpc gateway, it allows us to modify the translation between HTTP request
// and gRPC request. There are two potential sources for bearer tokens, it can come from an authorization header (not
// yet implemented), or encrypted cookies. Note that when deploying behind Envoy, you have the option to look for a
// configurable, non-standard header name. The token is extracted and turned into a metadata object which is then
// attached to the request, from which the token is extracted later for verification.
func GetHTTPRequestCookieToMetadataHandler(authContext interfaces.AuthenticationContext) HTTPRequestToMetadataAnnotator {
	return func(ctx context.Context, request *http.Request) metadata.MD {
		// TODO: Add read from Authorization header first, using the custom header if necessary.
		// TODO: Improve error handling
		accessToken, _, _ := authContext.CookieManager().RetrieveTokenValues(ctx, request)
		if accessToken == "" {
			logger.Infof(ctx, "Could not find access token cookie while requesting %s", request.RequestURI)
			return nil
		}
		return metadata.MD{
			DefaultAuthorizationHeader: []string{fmt.Sprintf("%s %s", BearerScheme, accessToken)},
		}
	}
}

// TODO: Add this to the Admin service IDL in Flyte IDL so that this can be exposed from gRPC as well.
// This returns a handler that will retrieve user info, from the OAuth2 authorization server.
// See the OpenID Connect spec at https://openid.net/specs/openid-connect-core-1_0.html#UserInfoResponse for more information.
func GetMeEndpointHandler(ctx context.Context, authCtx interfaces.AuthenticationContext) http.HandlerFunc {
	idpUserInfoEndpoint := authCtx.GetUserInfoURL().String()
	return func(writer http.ResponseWriter, request *http.Request) {
		access, _, err := authCtx.CookieManager().RetrieveTokenValues(ctx, request)
		if err != nil {
			http.Error(writer, "Error decoding identify token, try /login in again", http.StatusUnauthorized)
			return
		}
		// TODO: Investigate improving transparency of errors. The errors from this call may be just a local error, or may
		//       be an error from the HTTP request to the IDP. In the latter case, consider passing along the error code/msg.
		userInfo, err := postToIdp(ctx, authCtx.GetHTTPClient(), idpUserInfoEndpoint, access)
		if err != nil {
			logger.Errorf(ctx, "Error getting user info from IDP %s", err)
			http.Error(writer, "Error getting user info from IDP", http.StatusFailedDependency)
			return
		}

		bytes, err := json.Marshal(userInfo)
		if err != nil {
			logger.Errorf(ctx, "Error marshaling response into JSON %s", err)
			http.Error(writer, "Error marshaling response into JSON", http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		size, err := writer.Write(bytes)
		if err != nil {
			logger.Errorf(ctx, "Wrote user info response size %d, err %s", size, err)
		}
	}
}

// This returns a handler that will redirect (303) to the well-known metadata endpoint for the OAuth2 authorization server
// See https://tools.ietf.org/html/rfc8414 for more information.
func GetMetadataEndpointRedirectHandler(ctx context.Context, authCtx interfaces.AuthenticationContext) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		metadataURL := authCtx.GetBaseURL().ResolveReference(authCtx.GetMetadataURL())
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

// These are here for CORS handling. Actual serving of the OPTIONS request will be done by the gorilla/handlers package
type CorsHandlerDecorator func(http.Handler) http.Handler
