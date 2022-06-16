package authzserver

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/flyteorg/flyteadmin/auth"
	"github.com/ory/fosite"

	"github.com/flyteorg/flyteadmin/auth/interfaces"
	"github.com/flyteorg/flytestdlib/logger"
)

const (
	requestedScopePrefix = "f."
	accessTokenScope     = "access_token"
	refreshTokenScope    = "offline"
)

func getAuthEndpoint(authCtx interfaces.AuthenticationContext) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		authEndpoint(authCtx, writer, request)
	}
}

func getAuthCallbackEndpoint(authCtx interfaces.AuthenticationContext) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		authCallbackEndpoint(authCtx, writer, request)
	}
}

// authCallbackEndpoint is the endpoint that gets called after the user-auth flow finishes. It retrieves the original
// /authorize request and issues an auth_code in response.
func authCallbackEndpoint(authCtx interfaces.AuthenticationContext, rw http.ResponseWriter, req *http.Request) {
	issuer := GetIssuer(req.Context(), req, authCtx.Options())

	// This context will be passed to all methods.
	ctx := req.Context()
	oauth2Provider := authCtx.OAuth2Provider()

	// Get the user's identity
	identityContext, err := auth.IdentityContextFromRequest(ctx, req, authCtx)
	if err != nil {
		logger.Infof(ctx, "Failed to acquire user identity from request: %+v", err)
		oauth2Provider.WriteAuthorizeError(rw, fosite.NewAuthorizeRequest(), err)
		return
	}

	// Get latest user's info either from identity or by making a UserInfo() call to the original
	userInfo, err := auth.QueryUserInfo(ctx, identityContext, req, authCtx)
	if err != nil {
		err = fmt.Errorf("failed to query user info. Error: %w", err)
		http.Error(rw, err.Error(), http.StatusUnauthorized)
		return
	}

	// Rehydrate the original auth code request
	arURL, err := authCtx.CookieManager().RetrieveAuthCodeRequest(ctx, req)
	if err != nil {
		logger.Infof(ctx, "Error occurred in NewAuthorizeRequest: %+v", err)
		oauth2Provider.WriteAuthorizeError(rw, fosite.NewAuthorizeRequest(), err)
		return
	}

	arReq, err := http.NewRequest(http.MethodGet, arURL, nil)
	if err != nil {
		logger.Infof(ctx, "Error occurred in NewAuthorizeRequest: %+v", err)
		oauth2Provider.WriteAuthorizeError(rw, fosite.NewAuthorizeRequest(), err)
		return
	}

	ar, err := oauth2Provider.NewAuthorizeRequest(ctx, arReq)
	if err != nil {
		logger.Infof(ctx, "Error occurred in NewAuthorizeRequest: %+v", err)
		oauth2Provider.WriteAuthorizeError(rw, ar, err)
		return
	}

	// TODO: Ideally this is where we show users a consent form.

	// let's see what scopes the user gave consent to
	for _, scope := range req.PostForm["scopes"] {
		ar.GrantScope(scope)
	}

	// Now that the user is authorized, we set up a session:
	mySessionData := oauth2Provider.NewJWTSessionToken(identityContext.UserID(), ar.GetClient().GetID(), issuer, issuer, userInfo)
	mySessionData.JWTClaims.ExpiresAt = time.Now().Add(authCtx.Options().AppAuth.SelfAuthServer.AccessTokenLifespan.Duration)
	mySessionData.SetExpiresAt(fosite.AuthorizeCode, time.Now().Add(authCtx.Options().AppAuth.SelfAuthServer.AuthorizationCodeLifespan.Duration))
	mySessionData.SetExpiresAt(fosite.AccessToken, time.Now().Add(authCtx.Options().AppAuth.SelfAuthServer.AccessTokenLifespan.Duration))
	mySessionData.SetExpiresAt(fosite.RefreshToken, time.Now().Add(authCtx.Options().AppAuth.SelfAuthServer.RefreshTokenLifespan.Duration))

	// Now we need to get a response. This is the place where the AuthorizeEndpointHandlers kick in and start processing the request.
	// NewAuthorizeResponse is capable of running multiple response type handlers.
	response, err := oauth2Provider.NewAuthorizeResponse(ctx, ar, mySessionData)
	if err != nil {
		log.Printf("Error occurred in NewAuthorizeResponse: %+v", err)
		oauth2Provider.WriteAuthorizeError(rw, ar, err)
		return
	}

	// Last but not least, send the response!
	oauth2Provider.WriteAuthorizeResponse(rw, ar, response)
}

// Get the /authorize endpoint handler that is supposed to be invoked in the browser for the user to log in and consent.
func authEndpoint(authCtx interfaces.AuthenticationContext, rw http.ResponseWriter, req *http.Request) {
	// This context will be passed to all methods.
	ctx := req.Context()

	oauth2Provider := authCtx.OAuth2Provider()

	// Let's create an AuthorizeRequest object!
	// It will analyze the request and extract important information like scopes, response type and others.
	ar, err := oauth2Provider.NewAuthorizeRequest(ctx, req)
	if err != nil {
		logger.Infof(ctx, "Error occurred in NewAuthorizeRequest: %+v", err)
		oauth2Provider.WriteAuthorizeError(rw, ar, err)
		return
	}

	err = authCtx.CookieManager().SetAuthCodeCookie(ctx, rw, req.URL.String())
	if err != nil {
		logger.Infof(ctx, "Error occurred in NewAuthorizeRequest: %+v", err)
		oauth2Provider.WriteAuthorizeError(rw, ar, err)
		return
	}

	redirectURL := fmt.Sprintf("/login?redirect_url=%v", authorizeCallbackRelativeURL.String())
	http.Redirect(rw, req, redirectURL, http.StatusTemporaryRedirect)
}
