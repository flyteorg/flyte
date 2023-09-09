package authzserver

import (
	"crypto/rsa"

	"github.com/ory/fosite/handler/oauth2"

	"github.com/ory/fosite"

	"github.com/flyteorg/flyteadmin/auth/interfaces"

	"github.com/ory/fosite/compose"
	"github.com/ory/fosite/token/jwt"
)

// RegisterHandlers registers http endpoints for handling OAuth2 flow (/authorize,
func RegisterHandlers(handler interfaces.HandlerRegisterer, authCtx interfaces.AuthenticationContext) {
	// If using flyte self auth server, OAuth2Provider != nil
	if authCtx.OAuth2Provider() != nil {
		// Set up oauthserver endpoints. You could also use gorilla/mux or any other router.
		handler.HandleFunc(authorizeRelativeURL.String(), getAuthEndpoint(authCtx))
		handler.HandleFunc(authorizeCallbackRelativeURL.String(), getAuthCallbackEndpoint(authCtx))
		handler.HandleFunc(tokenRelativeURL.String(), getTokenEndpointHandler(authCtx))
		handler.HandleFunc(jsonWebKeysURL.String(), GetJSONWebKeysEndpoint(authCtx))
	}
}

// composeOAuth2Provider builds a fosite.OAuth2Provider that uses JWT for issuing access tokens and uses the provided
// codeProvider to issue AuthCode and RefreshTokens.
func composeOAuth2Provider(codeProvider oauth2.CoreStrategy, config *compose.Config, storage fosite.Storage,
	key *rsa.PrivateKey) fosite.OAuth2Provider {

	commonStrategy := &compose.CommonStrategy{
		CoreStrategy:               codeProvider,
		OpenIDConnectTokenStrategy: compose.NewOpenIDConnectStrategy(config, key),
		JWTStrategy: &jwt.RS256JWTStrategy{
			PrivateKey: key,
		},
	}

	return compose.Compose(
		config,
		storage,
		commonStrategy,
		nil,

		compose.OAuth2AuthorizeExplicitFactory,
		compose.OAuth2ClientCredentialsGrantFactory,
		compose.OAuth2RefreshTokenGrantFactory,

		compose.OAuth2StatelessJWTIntrospectionFactory,
		//compose.OAuth2TokenRevocationFactory,

		compose.OAuth2PKCEFactory,
	)
}
