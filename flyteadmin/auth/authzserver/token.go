package authzserver

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/ory/fosite"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flyteadmin/auth/interfaces"
)

var (
	supportedGrantTypes = []string{"client_credentials", "refresh_token", "authorization_code"}
)

func getTokenEndpointHandler(authCtx interfaces.AuthenticationContext) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		tokenEndpoint(authCtx, writer, request)
	}
}

func tokenEndpoint(authCtx interfaces.AuthenticationContext, rw http.ResponseWriter, req *http.Request) {
	// This context will be passed to all methods.
	ctx := req.Context()

	oauth2Provider := authCtx.OAuth2Provider()

	// Create an empty session object which will be passed to the request handlers
	emptySession := oauth2Provider.NewJWTSessionToken("", "", "", "", nil)

	// This will create an access request object and iterate through the registered TokenEndpointHandlers to validate the request.
	accessRequest, err := oauth2Provider.NewAccessRequest(ctx, req, emptySession)
	if err != nil {
		logger.Infof(ctx, "Error occurred in NewAccessRequest: %+v", err)
		oauth2Provider.WriteAccessError(rw, accessRequest, err)
		return
	}

	fositeAccessRequest, casted := accessRequest.(*fosite.AccessRequest)
	if !casted {
		logger.Errorf(ctx, "Invalid type. Expected *fosite.AccessRequest. Found: %v", reflect.TypeOf(accessRequest))
		oauth2Provider.WriteAccessError(rw, accessRequest, fosite.ErrInvalidRequest)
		return
	}

	// If this is a client_credentials grant, grant all requested scopes
	// NewAccessRequest validated that all requested scopes the client is allowed to perform
	// based on configured scope matching strategy.
	// If this is authorization_code, we should have consented the user for the requested scopes, so grant those too
	if fositeAccessRequest.GetGrantTypes().HasOneOf(supportedGrantTypes...) {
		requestedScopes := fositeAccessRequest.GetRequestedScopes()
		fositeAccessRequest.GrantedScope = fosite.Arguments{}
		for _, scope := range requestedScopes {
			fositeAccessRequest.GrantScope(strings.TrimPrefix(scope, requestedScopePrefix))
		}

		aud := GetIssuer(ctx, req, authCtx.Options())
		fositeAccessRequest.GrantAudience(aud)
	} else {
		logger.Infof(ctx, "Unsupported grant types [%+v]", fositeAccessRequest.GetGrantTypes())
		oauth2Provider.WriteAccessError(rw, fositeAccessRequest, fosite.ErrUnsupportedGrantType)
		return
	}

	// Next we create a response for the access request. Again, we iterate through the TokenEndpointHandlers
	// and aggregate the result in response.
	response, err := oauth2Provider.NewAccessResponse(ctx, fositeAccessRequest)
	if err != nil {
		logger.Infof(ctx, "Error occurred in NewAccessResponse: %+v", err)
		oauth2Provider.WriteAccessError(rw, fositeAccessRequest, err)
		return
	}

	// All done, send the response.
	oauth2Provider.WriteAccessResponse(rw, fositeAccessRequest, response)

	// The client now has a valid access token
}
