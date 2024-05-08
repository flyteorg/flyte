package authzserver

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"time"

	jwtgo "github.com/golang-jwt/jwt/v4"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
	fositeOAuth2 "github.com/ory/fosite/handler/oauth2"
	"github.com/ory/fosite/storage"
	"github.com/ory/fosite/token/jwt"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyte/flyteadmin/auth"
	"github.com/flyteorg/flyte/flyteadmin/auth/config"
	"github.com/flyteorg/flyte/flyteadmin/auth/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

const (
	ClientIDClaim = "client_id"
	UserIDClaim   = "user_info"
	ScopeClaim    = "scp"
	KeyIDClaim    = "key_id"
)

// Provider implements OAuth2 Authorization Server.
type Provider struct {
	fosite.OAuth2Provider
	cfg       config.AuthorizationServer
	publicKey []rsa.PublicKey
	keySet    jwk.Set
}

func (p Provider) PublicKeys() []rsa.PublicKey {
	return p.publicKey
}

func (p Provider) KeySet() jwk.Set {
	return p.keySet
}

// NewJWTSessionToken is a helper function for creating a new session.
func (p Provider) NewJWTSessionToken(subject, appID, issuer, audience string, userInfoClaims *service.UserInfoResponse) *fositeOAuth2.JWTSession {
	key, found := p.keySet.Get(0)
	keyID := ""
	if found {
		keyID = key.KeyID()
	}

	return &fositeOAuth2.JWTSession{
		JWTClaims: &jwt.JWTClaims{
			Audience:  []string{audience},
			Issuer:    issuer,
			Subject:   subject,
			ExpiresAt: time.Now().Add(p.cfg.AccessTokenLifespan.Duration),
			IssuedAt:  time.Now(),
			Extra: map[string]interface{}{
				ClientIDClaim: appID,
				UserIDClaim:   userInfoClaims,
			},
		},
		JWTHeader: &jwt.Headers{
			Extra: map[string]interface{}{
				KeyIDClaim: keyID,
			},
		},
	}
}

func findPublicKeyForTokenOrFirst(ctx context.Context, t *jwtgo.Token, publicKeys jwk.Set) (*rsa.PublicKey, error) {
	if _, ok := t.Method.(*jwtgo.SigningMethodRSA); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
	}

	if publicKeys.Len() == 0 {
		return nil, fmt.Errorf("no keys exist to match")
	}

	publicKey := &rsa.PublicKey{}
	k, _ := publicKeys.Get(0)
	if err := k.Raw(publicKey); err != nil {
		return nil, err
	}

	if keyID, found := t.Header[KeyIDClaim]; !found {
		return publicKey, nil
	} else if key, found := publicKeys.LookupKeyID(keyID.(string)); !found {
		return publicKey, nil
	} else if err := key.Raw(publicKey); err != nil {
		logger.Errorf(ctx, "Failed to load public key from key [%v]. Will default to the first key. Error: %v", keyID)
		return publicKey, nil
	}

	return publicKey, nil
}

func (p Provider) ValidateAccessToken(ctx context.Context, expectedAudience, tokenStr string) (interfaces.IdentityContext, error) {
	// Parse and validate the token.
	parsedToken, err := jwtgo.Parse(tokenStr, func(t *jwtgo.Token) (interface{}, error) {
		return findPublicKeyForTokenOrFirst(ctx, t, p.KeySet())
	})

	if err != nil {
		return nil, err
	}

	if !parsedToken.Valid {
		return nil, fmt.Errorf("parsed token is invalid")
	}

	claimsRaw := parsedToken.Claims.(jwtgo.MapClaims)
	return verifyClaims(sets.NewString(expectedAudience), claimsRaw)
}

// NewProvider creates a new OAuth2 Provider that is able to do OAuth 2-legged and 3-legged flows. It'll lookup
// config.SecretNameClaimSymmetricKey and config.SecretNameTokenSigningRSAKey secrets from the secret manager to use to
// sign and generate hashes for tokens. The RSA Private key is expected to be in PEM format with the public key embedded.
// Use auth.GetInitSecretsCommand() to generate new valid secrets that will be accepted by this provider.
// The config.SecretNameClaimSymmetricKey must be a 32-bytes long key in Base64Encoding.
func NewProvider(ctx context.Context, cfg config.AuthorizationServer, sm core.SecretManager) (Provider, error) {
	// fosite requires four parameters for the server to get up and running:
	// 1. config - for any enforcement you may desire, you can do this using `compose.Config`. You like PKCE, enforce it!
	// 2. store - no auth service is generally useful unless it can remember clients and users.
	//    fosite is incredibly composable, and the store parameter enables you to build and BYODb (Bring Your Own Database)
	// 3. secret - required for code, access and refresh token generation.
	// 4. privateKey - required for id/jwt token generation.

	composeConfig := &compose.Config{
		AccessTokenLifespan:   cfg.AccessTokenLifespan.Duration,
		RefreshTokenLifespan:  cfg.RefreshTokenLifespan.Duration,
		AuthorizeCodeLifespan: cfg.AuthorizationCodeLifespan.Duration,
		RefreshTokenScopes:    []string{refreshTokenScope},
	}

	// This secret is used to encryptString/decrypt challenge code to maintain a stateless authcode token.
	tokenHashBase64, err := sm.Get(ctx, cfg.ClaimSymmetricEncryptionKeySecretName)
	if err != nil {
		return Provider{}, fmt.Errorf("failed to read secretTokenHash file. Error: %w", err)
	}
	if tokenHashBase64 == "" {
		return Provider{}, fmt.Errorf("failed to read secretTokenHash. Error: empty value")
	}

	secret, err := base64.RawStdEncoding.DecodeString(tokenHashBase64)
	if err != nil {
		return Provider{}, fmt.Errorf("failed to decode token hash using base64 encoding. Error: %w", err)
	}

	// privateKey is used to sign JWT tokens. The default strategy uses RS256 (RSA Signature with SHA-256)
	privateKeyPEM, err := sm.Get(ctx, cfg.TokenSigningRSAKeySecretName)
	if err != nil {
		return Provider{}, fmt.Errorf("failed to read token signing RSA Key. Error: %w", err)
	}
	if privateKeyPEM == "" {
		return Provider{}, fmt.Errorf("failed to read token signing RSA Key. Error: empty value")
	}

	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return Provider{}, fmt.Errorf("failed to decode token signing RSA Key. Error: no PEM data found")
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return Provider{}, fmt.Errorf("failed to parse PKCS1PrivateKey. Error: %w", err)
	}

	// Build an in-memory store with static clients defined in Config. This gives us the potential to move the clients
	// storage into DB and allow registration of new clients to users.
	store := &StatelessTokenStore{
		MemoryStore: &storage.MemoryStore{
			IDSessions:             make(map[string]fosite.Requester),
			Clients:                toClientIface(cfg.StaticClients),
			AuthorizeCodes:         map[string]storage.StoreAuthorizeCode{},
			AccessTokens:           map[string]fosite.Requester{},
			RefreshTokens:          map[string]storage.StoreRefreshToken{},
			PKCES:                  map[string]fosite.Requester{},
			AccessTokenRequestIDs:  map[string]string{},
			RefreshTokenRequestIDs: map[string]string{},
			IssuerPublicKeys:       map[string]storage.IssuerPublicKeys{},
		},
	}

	sec := [auth.SymmetricKeyLength]byte{}
	copy(sec[:], secret)
	codeProvider := NewStatelessCodeProvider(cfg, sec, compose.NewOAuth2JWTStrategy(privateKey, nil))

	// Build a fosite instance with all OAuth2 and OpenID Connect handlers enabled, plugging in our configurations as specified above.
	oauth2Provider := composeOAuth2Provider(codeProvider, composeConfig, store, privateKey)
	store.JWTStrategy = &jwt.RS256JWTStrategy{
		PrivateKey: privateKey,
	}
	store.encryptor = codeProvider

	publicKeys := []rsa.PublicKey{privateKey.PublicKey}

	// Try to load old key to validate tokens using it to support key rotation.
	privateKeyPEM, err = sm.Get(ctx, cfg.OldTokenSigningRSAKeySecretName)
	if err == nil {
		if privateKeyPEM == "" {
			return Provider{}, fmt.Errorf("failed to read PKCS1PrivateKey. Error: empty value")
		}
		block, _ = pem.Decode([]byte(privateKeyPEM))
		if block == nil {
			return Provider{}, fmt.Errorf("failed to decode PKCS1PrivateKey. Error: no PEM data found")
		}
		oldPrivateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return Provider{}, fmt.Errorf("failed to parse PKCS1PrivateKey. Error: %w", err)
		}

		publicKeys = append(publicKeys, oldPrivateKey.PublicKey)
	}

	keysSet, err := newJSONWebKeySet(publicKeys)
	if err != nil {
		return Provider{}, err
	}

	return Provider{
		OAuth2Provider: oauth2Provider,
		publicKey:      publicKeys,
		keySet:         keysSet,
	}, nil
}
