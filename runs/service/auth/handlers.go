package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	authpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/auth"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/runs/service/auth/config"
)

// PreRedirectHookError is returned by PreRedirectHookFunc to signal an error with an HTTP status code.
type PreRedirectHookError struct {
	Message string
	Code    int
}

func (e *PreRedirectHookError) Error() string {
	return e.Message
}

// PreRedirectHookFunc is called before the redirect at the end of a successful auth callback flow.
type PreRedirectHookFunc func(ctx context.Context, request *http.Request, w http.ResponseWriter) *PreRedirectHookError

// LogoutHookFunc is called during logout to perform additional cleanup.
type LogoutHookFunc func(ctx context.Context, request *http.Request, w http.ResponseWriter) error

// AuthHandlerConfig holds dependencies needed by the HTTP auth handlers.
type AuthHandlerConfig struct {
	CookieManager     CookieManager
	OAuth2Config      *oauth2.Config
	OIDCProvider      *oidc.Provider
	ResourceServer    OAuth2ResourceServer
	AuthConfig        config.Config
	HTTPClient        *http.Client
	PreRedirectHook   PreRedirectHookFunc
	LogoutHook        LogoutHookFunc
}

// RegisterHandlers registers the standard OAuth2/OIDC HTTP handlers on the given mux.
func RegisterHandlers(ctx context.Context, mux *http.ServeMux, h *AuthHandlerConfig) {
	mux.HandleFunc("/login", RefreshTokensIfNeededHandler(ctx, h,
		GetLoginHandler(ctx, h)))
	mux.HandleFunc("/callback", GetCallbackHandler(ctx, h))
	mux.HandleFunc(fmt.Sprintf("/%s", OIdCMetadataEndpoint), GetOIdCMetadataEndpointRedirectHandler(ctx, h))
	mux.HandleFunc("/logout", GetLogoutEndpointHandler(ctx, h))
}

// RefreshTokensIfNeededHandler wraps a handler to attempt token refresh before redirecting.
func RefreshTokensIfNeededHandler(ctx context.Context, h *AuthHandlerConfig, authHandler http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		newToken, userInfo, refreshed, err := RefreshTokensIfNeeded(ctx, h, request)
		if err != nil {
			logger.Infof(ctx, "Failed to refresh tokens. Restarting login flow. Error: %s", err)
			authHandler(writer, request)
			return
		}

		if refreshed {
			logger.Debugf(ctx, "Tokens are refreshed. Saving new tokens into cookies.")
			if err = h.CookieManager.SetTokenCookies(ctx, writer, newToken); err != nil {
				logger.Infof(ctx, "Failed to write tokens to response. Restarting login flow. Error: %s", err)
				authHandler(writer, request)
				return
			}

			if err = h.CookieManager.SetUserInfoCookie(ctx, writer, userInfo); err != nil {
				logger.Infof(ctx, "Failed to write user info to response. Restarting login flow. Error: %s", err)
				authHandler(writer, request)
				return
			}
		}

		redirectURL := GetAuthFlowEndRedirect(ctx, h.AuthConfig.UserAuth.RedirectURL.String(), h.AuthConfig.AuthorizedURIs, request)
		http.Redirect(writer, request, redirectURL, http.StatusTemporaryRedirect)
	}
}

// RefreshTokensIfNeeded checks if tokens need refreshing and returns refreshed tokens if so.
func RefreshTokensIfNeeded(ctx context.Context, h *AuthHandlerConfig, request *http.Request) (
	token *oauth2.Token, userInfo *authpb.UserInfoResponse, refreshed bool, err error) {

	ctx = context.WithValue(ctx, oauth2.HTTPClient, h.HTTPClient)

	idToken, accessToken, refreshToken, err := h.CookieManager.RetrieveTokenValues(ctx, request)
	if err != nil {
		return nil, nil, false, fmt.Errorf("failed to retrieve tokens from request: %w", err)
	}

	_, err = ParseIDTokenAndValidate(ctx, h.AuthConfig.UserAuth.OpenID.ClientID, idToken, h.OIDCProvider)
	if err != nil {
		if strings.Contains(err.Error(), "token is expired") && len(refreshToken) > 0 {
			logger.Debugf(ctx, "Expired id token found, attempting to refresh")
			newToken, refreshErr := GetRefreshedToken(ctx, h.OAuth2Config, accessToken, refreshToken)
			if refreshErr != nil {
				return nil, nil, false, fmt.Errorf("failed to refresh tokens: %w", refreshErr)
			}

			userInfo, queryErr := QueryUserInfoUsingAccessToken(ctx, request, h, newToken.AccessToken)
			if queryErr != nil {
				return nil, nil, false, fmt.Errorf("failed to query user info: %w", queryErr)
			}

			return newToken, userInfo, true, nil
		}
		return nil, nil, false, fmt.Errorf("failed to validate tokens: %w", err)
	}

	return NewOAuthTokenFromRaw(accessToken, refreshToken, idToken), nil, false, nil
}

// GetLoginHandler returns an HTTP handler that starts the OAuth2 login flow.
func GetLoginHandler(ctx context.Context, h *AuthHandlerConfig) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		csrfCookie := NewCsrfCookie()
		csrfToken := csrfCookie.Value
		http.SetCookie(writer, &csrfCookie)

		state := HashCsrfState(csrfToken)
		logger.Debugf(ctx, "Setting CSRF state cookie to %s and state to %s\n", csrfToken, state)
		urlString := h.OAuth2Config.AuthCodeURL(state)
		queryParams := request.URL.Query()
		if !GetRedirectURLAllowed(ctx, queryParams.Get(RedirectURLParameter), h.AuthConfig.AuthorizedURIs) {
			logger.Infof(ctx, "unauthorized redirect URI")
			writer.WriteHeader(http.StatusForbidden)
			return
		}
		if flowEndRedirectURL := queryParams.Get(RedirectURLParameter); flowEndRedirectURL != "" {
			redirectCookie := NewRedirectCookie(ctx, flowEndRedirectURL)
			if redirectCookie != nil {
				http.SetCookie(writer, redirectCookie)
			}
		}

		http.Redirect(writer, request, urlString, http.StatusTemporaryRedirect)
	}
}

// GetCallbackHandler returns an HTTP handler that completes the OAuth2 authorization code flow.
func GetCallbackHandler(ctx context.Context, h *AuthHandlerConfig) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		logger.Debugf(ctx, "Running callback handler... for RequestURI %v", request.RequestURI)
		authorizationCode := request.FormValue(AuthorizationResponseCodeType)

		ctx = context.WithValue(ctx, oauth2.HTTPClient, h.HTTPClient)

		if err := VerifyCsrfCookie(ctx, request); err != nil {
			logger.Errorf(ctx, "Invalid CSRF token cookie %s", err)
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}

		token, err := h.OAuth2Config.Exchange(ctx, authorizationCode)
		if err != nil {
			logger.Errorf(ctx, "Error when exchanging code %s", err)
			writer.WriteHeader(http.StatusForbidden)
			return
		}

		if err = h.CookieManager.SetTokenCookies(ctx, writer, token); err != nil {
			logger.Errorf(ctx, "Error setting encrypted JWT cookie %s", err)
			writer.WriteHeader(http.StatusForbidden)
			return
		}

		userInfo, err := QueryUserInfoUsingAccessToken(ctx, request, h, token.AccessToken)
		if err != nil {
			logger.Errorf(ctx, "Failed to query user info. Error: %v", err)
			writer.WriteHeader(http.StatusForbidden)
			return
		}

		if err = h.CookieManager.SetUserInfoCookie(ctx, writer, userInfo); err != nil {
			logger.Errorf(ctx, "Error setting encrypted user info cookie. Error: %v", err)
			writer.WriteHeader(http.StatusForbidden)
			return
		}

		if h.PreRedirectHook != nil {
			if hookErr := h.PreRedirectHook(ctx, request, writer); hookErr != nil {
				logger.Errorf(ctx, "failed the preRedirect hook due %v with status code %v", hookErr.Message, hookErr.Code)
				if http.StatusText(hookErr.Code) != "" {
					writer.WriteHeader(hookErr.Code)
				} else {
					writer.WriteHeader(http.StatusInternalServerError)
				}
				return
			}
		}

		redirectURL := GetAuthFlowEndRedirect(ctx, h.AuthConfig.UserAuth.RedirectURL.String(), h.AuthConfig.AuthorizedURIs, request)
		http.Redirect(writer, request, redirectURL, http.StatusTemporaryRedirect)
	}
}

// GetOIdCMetadataEndpointRedirectHandler returns a handler that redirects to the OIDC metadata endpoint.
func GetOIdCMetadataEndpointRedirectHandler(_ context.Context, h *AuthHandlerConfig) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		oidcMetadataURL := h.AuthConfig.UserAuth.OpenID.BaseURL.JoinPath("/").JoinPath(OIdCMetadataEndpoint)
		http.Redirect(writer, request, oidcMetadataURL.String(), http.StatusSeeOther)
	}
}

// GetLogoutEndpointHandler returns a handler that clears auth cookies and optionally redirects.
func GetLogoutEndpointHandler(ctx context.Context, h *AuthHandlerConfig) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		if h.LogoutHook != nil {
			if err := h.LogoutHook(ctx, request, writer); err != nil {
				logger.Errorf(ctx, "logout hook failed: %v", err)
				writer.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		logger.Debugf(ctx, "deleting auth cookies")
		h.CookieManager.DeleteCookies(ctx, writer)

		queryParams := request.URL.Query()
		if redirectURL := queryParams.Get(RedirectURLParameter); redirectURL != "" {
			if !GetRedirectURLAllowed(ctx, redirectURL, h.AuthConfig.AuthorizedURIs) {
				logger.Warnf(ctx, "Rejecting unauthorized redirect_url in logout: %s", redirectURL)
				redirectURL = h.AuthConfig.UserAuth.RedirectURL.String()
			}
			http.Redirect(writer, request, redirectURL, http.StatusTemporaryRedirect)
		}
	}
}

// QueryUserInfoUsingAccessToken fetches user info from the OIDC provider using an access token.
func QueryUserInfoUsingAccessToken(ctx context.Context, originalRequest *http.Request, h *AuthHandlerConfig, accessToken string) (
	*authpb.UserInfoResponse, error) {

	originalToken := oauth2.Token{
		AccessToken: accessToken,
	}

	tokenSource := h.OAuth2Config.TokenSource(ctx, &originalToken)

	userInfo, err := h.OIDCProvider.UserInfo(ctx, tokenSource)
	if err != nil {
		logger.Errorf(ctx, "Error getting user info from IDP %s", err)
		return &authpb.UserInfoResponse{}, fmt.Errorf("error getting user info from IDP")
	}

	resp := &authpb.UserInfoResponse{}
	if err = userInfo.Claims(resp); err != nil {
		logger.Errorf(ctx, "Error getting user info from IDP %s", err)
		return &authpb.UserInfoResponse{}, fmt.Errorf("error getting user info from IDP")
	}

	return resp, nil
}

// IdentityContextFromRequest extracts identity from an HTTP request (header or cookies).
func IdentityContextFromRequest(ctx context.Context, req *http.Request, h *AuthHandlerConfig) (
	*IdentityContext, error) {

	authHeader := DefaultAuthorizationHeader
	if len(h.AuthConfig.HTTPAuthorizationHeader) > 0 {
		authHeader = h.AuthConfig.HTTPAuthorizationHeader
	}

	headerValue := req.Header.Get(authHeader)
	if len(headerValue) == 0 {
		headerValue = req.Header.Get(DefaultAuthorizationHeader)
	}

	if len(headerValue) > 0 {
		if strings.HasPrefix(headerValue, BearerScheme+" ") {
			expectedAudience := GetPublicURL(ctx, req, h.AuthConfig).String()
			return h.ResourceServer.ValidateAccessToken(ctx, expectedAudience, strings.TrimPrefix(headerValue, BearerScheme+" "))
		}
	}

	idToken, _, _, err := h.CookieManager.RetrieveTokenValues(ctx, req)
	if err != nil || len(idToken) == 0 {
		return nil, fmt.Errorf("unauthenticated request. IDToken Len [%v], Error: %w", len(idToken), err)
	}

	userInfo, err := h.CookieManager.RetrieveUserInfo(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("unauthenticated request: %w", err)
	}

	return IdentityContextFromIDToken(ctx, idToken, h.AuthConfig.UserAuth.OpenID.ClientID, h.OIDCProvider, userInfo)
}
