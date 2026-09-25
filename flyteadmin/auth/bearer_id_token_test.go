package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	jwtgo "github.com/golang-jwt/jwt/v4"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyte/flyteadmin/auth/config"
	"github.com/flyteorg/flyte/flyteadmin/auth/interfaces/mocks"
	stdconfig "github.com/flyteorg/flyte/flytestdlib/config"
)

const (
	testIDTokenClientID = "flyteadmin"
	testIDTokenSubject  = "user-subject"
	testIDTokenKeyID    = "test-key"
)

// fakeOIDCProvider serves an OpenID Connect discovery document and a JWKS for a generated RSA key, and mints ID tokens
// signed with it.
type fakeOIDCProvider struct {
	server *httptest.Server
	key    *rsa.PrivateKey
}

func newFakeOIDCProvider(t *testing.T) *fakeOIDCProvider {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.NoError(t, err)

	p := &fakeOIDCProvider{key: key}
	p.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"issuer":                                p.server.URL,
				"authorization_endpoint":                p.server.URL + "/auth",
				"token_endpoint":                        p.server.URL + "/token",
				"jwks_uri":                              p.server.URL + "/keys",
				"id_token_signing_alg_values_supported": []string{"RS256"},
			})
		case "/keys":
			pub, err := jwk.New(&key.PublicKey)
			assert.NoError(t, err)
			assert.NoError(t, pub.Set(jwk.KeyIDKey, testIDTokenKeyID))
			assert.NoError(t, pub.Set(jwk.AlgorithmKey, "RS256"))
			set := jwk.NewSet()
			set.Add(pub)
			_ = json.NewEncoder(w).Encode(set)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	return p
}

func (p *fakeOIDCProvider) provider(t *testing.T) *oidc.Provider {
	provider, err := oidc.NewProvider(oidc.ClientContext(context.Background(), p.server.Client()), p.server.URL)
	assert.NoError(t, err)
	return provider
}

func (p *fakeOIDCProvider) idToken(t *testing.T, audience string) string {
	token := jwtgo.NewWithClaims(jwtgo.SigningMethodRS256, jwtgo.MapClaims{
		"iss":   p.server.URL,
		"aud":   audience,
		"sub":   testIDTokenSubject,
		"email": "user@example.com",
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(time.Hour).Unix(),
	})
	token.Header["kid"] = testIDTokenKeyID
	signed, err := token.SignedString(p.key)
	assert.NoError(t, err)
	return signed
}

func newBearerIDTokenAuthContext(t *testing.T, provider *oidc.Provider) *mocks.AuthenticationContext {
	resourceServer := &mocks.OAuth2ResourceServer{}
	resourceServer.EXPECT().ValidateAccessToken(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("not an access token issued by this server"))

	authCtx := &mocks.AuthenticationContext{}
	authCtx.EXPECT().Options().Return(&config.Config{
		AuthorizedURIs: []stdconfig.URL{{URL: url.URL{Scheme: "https", Host: "flyte.example.com"}}},
		UserAuth:       config.UserAuthConfig{OpenID: config.OpenIDOptions{ClientID: testIDTokenClientID}},
	})
	authCtx.EXPECT().OAuth2ResourceServer().Return(resourceServer)
	authCtx.EXPECT().OidcProvider().Return(provider)
	return authCtx
}

func TestGetAuthenticationInterceptor_BearerIDToken(t *testing.T) {
	idp := newFakeOIDCProvider(t)
	defer idp.server.Close()
	provider := idp.provider(t)

	t.Run("id token sent with the Bearer scheme is accepted", func(t *testing.T) {
		authCtx := newBearerIDTokenAuthContext(t, provider)
		ctx := metadata.NewIncomingContext(context.Background(),
			metadata.Pairs(DefaultAuthorizationHeader, BearerScheme+" "+idp.idToken(t, testIDTokenClientID)))

		newCtx, err := GetAuthenticationInterceptor(authCtx)(ctx)
		assert.NoError(t, err)
		assert.Equal(t, testIDTokenSubject, IdentityContextFromContext(newCtx).UserID())
		assert.True(t, IdentityContextFromContext(newCtx).Scopes().Has(ScopeAll))
	})

	t.Run("id token sent with the IDToken scheme still works", func(t *testing.T) {
		authCtx := newBearerIDTokenAuthContext(t, provider)
		ctx := metadata.NewIncomingContext(context.Background(),
			metadata.Pairs(DefaultAuthorizationHeader, IDTokenScheme+" "+idp.idToken(t, testIDTokenClientID)))

		newCtx, err := GetAuthenticationInterceptor(authCtx)(ctx)
		assert.NoError(t, err)
		assert.Equal(t, testIDTokenSubject, IdentityContextFromContext(newCtx).UserID())
	})

	t.Run("bearer id token for another audience is rejected", func(t *testing.T) {
		authCtx := newBearerIDTokenAuthContext(t, provider)
		ctx := metadata.NewIncomingContext(context.Background(),
			metadata.Pairs(DefaultAuthorizationHeader, BearerScheme+" "+idp.idToken(t, "some-other-client")))

		_, err := GetAuthenticationInterceptor(authCtx)(ctx)
		assert.Error(t, err)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("garbage bearer token is rejected", func(t *testing.T) {
		authCtx := newBearerIDTokenAuthContext(t, provider)
		ctx := metadata.NewIncomingContext(context.Background(),
			metadata.Pairs(DefaultAuthorizationHeader, BearerScheme+" not.a.jwt"))

		_, err := GetAuthenticationInterceptor(authCtx)(ctx)
		assert.Error(t, err)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("no provider configured falls through to the usual rejection", func(t *testing.T) {
		authCtx := newBearerIDTokenAuthContext(t, nil)
		ctx := metadata.NewIncomingContext(context.Background(),
			metadata.Pairs(DefaultAuthorizationHeader, BearerScheme+" "+idp.idToken(t, testIDTokenClientID)))

		_, err := GetAuthenticationInterceptor(authCtx)(ctx)
		assert.Error(t, err)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})
}

func TestIdentityContextFromRequest_BearerIDToken(t *testing.T) {
	idp := newFakeOIDCProvider(t)
	defer idp.server.Close()
	provider := idp.provider(t)
	ctx := context.Background()

	t.Run("id token sent with the Bearer scheme is accepted", func(t *testing.T) {
		authCtx := newBearerIDTokenAuthContext(t, provider)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
		req.Header.Set(DefaultAuthorizationHeader, BearerScheme+" "+idp.idToken(t, testIDTokenClientID))

		identityCtx, err := IdentityContextFromRequest(ctx, req, authCtx)
		assert.NoError(t, err)
		assert.Equal(t, testIDTokenSubject, identityCtx.UserID())
	})

	t.Run("bearer id token for another audience is rejected", func(t *testing.T) {
		authCtx := newBearerIDTokenAuthContext(t, provider)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
		req.Header.Set(DefaultAuthorizationHeader, BearerScheme+" "+idp.idToken(t, "some-other-client"))

		identityCtx, err := IdentityContextFromRequest(ctx, req, authCtx)
		assert.Error(t, err)
		assert.Nil(t, identityCtx)
	})
}
