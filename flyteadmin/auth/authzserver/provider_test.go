package authzserver

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"testing"
	"time"

	jwtgo "github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteadmin/auth"
	"github.com/flyteorg/flyte/flyteadmin/auth/config"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
)

func newMockProvider(t testing.TB) (Provider, auth.SecretsSet) {
	secrets, err := auth.NewSecrets()
	assert.NoError(t, err)

	ctx := context.Background()
	sm := &mocks.SecretManager{}
	sm.OnGet(ctx, config.SecretNameClaimSymmetricKey).Return(base64.RawStdEncoding.EncodeToString(secrets.TokenHashKey), nil)
	sm.OnGet(ctx, config.SecretNameCookieBlockKey).Return(base64.RawStdEncoding.EncodeToString(secrets.CookieBlockKey), nil)
	sm.OnGet(ctx, config.SecretNameCookieHashKey).Return(base64.RawStdEncoding.EncodeToString(secrets.CookieHashKey), nil)

	privBytes := x509.MarshalPKCS1PrivateKey(secrets.TokenSigningRSAPrivateKey)
	var buf bytes.Buffer
	assert.NoError(t, pem.Encode(&buf, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes}))
	sm.OnGet(ctx, config.SecretNameTokenSigningRSAKey).Return(buf.String(), nil)
	sm.OnGet(ctx, config.SecretNameOldTokenSigningRSAKey).Return(buf.String(), nil)

	p, err := NewProvider(ctx, config.DefaultConfig.AppAuth.SelfAuthServer, sm)
	assert.NoError(t, err)
	return p, secrets
}

func TestNewProvider(t *testing.T) {
	newMockProvider(t)
}

func newInvalidMockProvider(ctx context.Context, t *testing.T, secrets auth.SecretsSet, sm *mocks.SecretManager, invalidFunc func() *mocks.SecretManager_Get, errorContains string) {

	sm.OnGet(ctx, config.SecretNameClaimSymmetricKey).Return(base64.RawStdEncoding.EncodeToString(secrets.TokenHashKey), nil)
	sm.OnGet(ctx, config.SecretNameCookieBlockKey).Return(base64.RawStdEncoding.EncodeToString(secrets.CookieBlockKey), nil)
	sm.OnGet(ctx, config.SecretNameCookieHashKey).Return(base64.RawStdEncoding.EncodeToString(secrets.CookieHashKey), nil)

	privBytes := x509.MarshalPKCS1PrivateKey(secrets.TokenSigningRSAPrivateKey)
	var buf bytes.Buffer
	assert.NoError(t, pem.Encode(&buf, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes}))
	sm.OnGet(ctx, config.SecretNameTokenSigningRSAKey).Return(buf.String(), nil)
	sm.OnGet(ctx, config.SecretNameOldTokenSigningRSAKey).Return(buf.String(), nil)

	invalidFunc()
	p, err := NewProvider(ctx, config.DefaultConfig.AppAuth.SelfAuthServer, sm)
	assert.Error(t, err)
	assert.ErrorContains(t, err, errorContains)
	assert.Equal(t, Provider{}, p)
}

func TestNewInvalidProviderSecretTokenHashBad(t *testing.T) {
	secrets, err := auth.NewSecrets()
	assert.NoError(t, err)

	ctx := context.Background()
	sm := &mocks.SecretManager{}

	invalidFunc := func() *mocks.SecretManager_Get {
		sm.OnGet(ctx, config.SecretNameClaimSymmetricKey).Unset()
		return sm.OnGet(ctx, config.SecretNameClaimSymmetricKey).Return("", fmt.Errorf("test error"))
	}
	newInvalidMockProvider(ctx, t, secrets, sm, invalidFunc, "failed to read secretTokenHash file. Error: test error")
}

func TestNewInvalidProviderSecretTokenHashEmpty(t *testing.T) {
	secrets, err := auth.NewSecrets()
	assert.NoError(t, err)

	ctx := context.Background()
	sm := &mocks.SecretManager{}

	invalidFunc := func() *mocks.SecretManager_Get {
		sm.OnGet(ctx, config.SecretNameClaimSymmetricKey).Unset()
		return sm.OnGet(ctx, config.SecretNameClaimSymmetricKey).Return("", nil)
	}
	newInvalidMockProvider(ctx, t, secrets, sm, invalidFunc, "failed to read secretTokenHash. Error: empty value")
}

func TestNewInvalidProviderTokenSigningRSAKeyBad(t *testing.T) {
	secrets, err := auth.NewSecrets()
	assert.NoError(t, err)

	ctx := context.Background()
	sm := &mocks.SecretManager{}

	invalidFunc := func() *mocks.SecretManager_Get {
		sm.OnGet(ctx, config.SecretNameTokenSigningRSAKey).Unset()
		return sm.OnGet(ctx, config.SecretNameTokenSigningRSAKey).Return("", fmt.Errorf("test error"))
	}
	newInvalidMockProvider(ctx, t, secrets, sm, invalidFunc, "failed to read token signing RSA Key. Error: test error")
}

func TestNewInvalidProviderTokenSigningRSAKeyEmpty(t *testing.T) {
	secrets, err := auth.NewSecrets()
	assert.NoError(t, err)

	ctx := context.Background()
	sm := &mocks.SecretManager{}

	invalidFunc := func() *mocks.SecretManager_Get {
		sm.OnGet(ctx, config.SecretNameTokenSigningRSAKey).Unset()
		return sm.OnGet(ctx, config.SecretNameTokenSigningRSAKey).Return("", nil)
	}
	newInvalidMockProvider(ctx, t, secrets, sm, invalidFunc, "failed to read token signing RSA Key. Error: empty value")
}

func TestNewInvalidProviderTokenSigningRSAKeyNoPEMData(t *testing.T) {
	secrets, err := auth.NewSecrets()
	assert.NoError(t, err)

	ctx := context.Background()
	sm := &mocks.SecretManager{}

	invalidFunc := func() *mocks.SecretManager_Get {
		sm.OnGet(ctx, config.SecretNameTokenSigningRSAKey).Unset()
		return sm.OnGet(ctx, config.SecretNameTokenSigningRSAKey).Return("this is no PEM data", nil)
	}
	newInvalidMockProvider(ctx, t, secrets, sm, invalidFunc, "failed to decode token signing RSA Key. Error: no PEM data found")
}

func TestNewInvalidProviderOldTokenSigningRSAKeyEmpty(t *testing.T) {
	secrets, err := auth.NewSecrets()
	assert.NoError(t, err)

	ctx := context.Background()
	sm := &mocks.SecretManager{}

	invalidFunc := func() *mocks.SecretManager_Get {
		sm.OnGet(ctx, config.SecretNameOldTokenSigningRSAKey).Unset()
		return sm.OnGet(ctx, config.SecretNameOldTokenSigningRSAKey).Return("", nil)
	}
	newInvalidMockProvider(ctx, t, secrets, sm, invalidFunc, "failed to read PKCS1PrivateKey. Error: empty value")
}

func TestNewInvalidProviderOldTokenSigningRSAKeyNoPEMData(t *testing.T) {
	secrets, err := auth.NewSecrets()
	assert.NoError(t, err)

	ctx := context.Background()
	sm := &mocks.SecretManager{}

	invalidFunc := func() *mocks.SecretManager_Get {
		sm.OnGet(ctx, config.SecretNameOldTokenSigningRSAKey).Unset()
		return sm.OnGet(ctx, config.SecretNameOldTokenSigningRSAKey).Return("this is no PEM data", nil)
	}
	newInvalidMockProvider(ctx, t, secrets, sm, invalidFunc, "failed to decode PKCS1PrivateKey. Error: no PEM data found")
}

func TestProvider_KeySet(t *testing.T) {
	p, _ := newMockProvider(t)
	assert.Equal(t, 2, p.KeySet().Len())
}

func TestProvider_NewJWTSessionToken(t *testing.T) {
	p, _ := newMockProvider(t)
	s := p.NewJWTSessionToken("userID", "appID", "my-issuer", "my-audience", &service.UserInfoResponse{
		Email: "foo@localhost",
	})

	k, found := p.KeySet().Get(0)
	assert.True(t, found)
	assert.NotEmpty(t, k.KeyID())
	assert.Equal(t, k.KeyID(), s.JWTHeader.Extra[KeyIDClaim])
}

func TestProvider_PublicKeys(t *testing.T) {
	p, _ := newMockProvider(t)
	assert.Len(t, p.PublicKeys(), 2)
}

type CustomClaimsExample struct {
	*jwtgo.StandardClaims
	ClientID string   `json:"client_id"`
	Scopes   []string `json:"scp"`
	UserID   string   `json:"user_id"`
}

func TestProvider_findPublicKeyForTokenOrFirst(t *testing.T) {
	ctx := context.Background()
	secrets, err := auth.NewSecrets()
	assert.NoError(t, err)

	secrets2, err := auth.NewSecrets()
	assert.NoError(t, err)

	keySet, err := newJSONWebKeySet([]rsa.PublicKey{secrets.TokenSigningRSAPrivateKey.PublicKey, secrets2.TokenSigningRSAPrivateKey.PublicKey})
	assert.NoError(t, err)

	// set our claims
	secondKey, found := keySet.Get(1)
	assert.True(t, found)

	t.Run("KeyID Exists", func(t *testing.T) {
		// create a signer for rsa 256
		tok := jwtgo.New(jwtgo.GetSigningMethod("RS256"))

		tok.Header[KeyIDClaim] = secondKey.KeyID()
		tok.Claims = &CustomClaimsExample{
			StandardClaims: &jwtgo.StandardClaims{},
		}

		// Create token string
		_, err = tok.SignedString(secrets2.TokenSigningRSAPrivateKey)
		assert.NoError(t, err)

		k, err := findPublicKeyForTokenOrFirst(ctx, tok, keySet)
		assert.NoError(t, err)
		assert.Equal(t, secrets2.TokenSigningRSAPrivateKey.PublicKey, *k)
	})

	t.Run("Unknown KeyID, Default to first key", func(t *testing.T) {
		// create a signer for rsa 256
		tok := jwtgo.New(jwtgo.GetSigningMethod("RS256"))

		tok.Header[KeyIDClaim] = "not found"
		tok.Claims = &CustomClaimsExample{
			StandardClaims: &jwtgo.StandardClaims{
				ExpiresAt: time.Now().Add(time.Minute * 1).Unix(),
			},
		}

		// Create token string
		_, err = tok.SignedString(secrets2.TokenSigningRSAPrivateKey)
		assert.NoError(t, err)

		k, err := findPublicKeyForTokenOrFirst(ctx, tok, keySet)
		assert.NoError(t, err)
		assert.Equal(t, secrets.TokenSigningRSAPrivateKey.PublicKey, *k)
	})

	t.Run("No KeyID Claim, Default to first key", func(t *testing.T) {
		// create a signer for rsa 256
		tok := jwtgo.New(jwtgo.GetSigningMethod("RS256"))

		tok.Claims = &CustomClaimsExample{
			StandardClaims: &jwtgo.StandardClaims{
				ExpiresAt: time.Now().Add(time.Minute * 1).Unix(),
			},
		}

		// Create token string
		_, err = tok.SignedString(secrets2.TokenSigningRSAPrivateKey)
		assert.NoError(t, err)

		k, err := findPublicKeyForTokenOrFirst(ctx, tok, keySet)
		assert.NoError(t, err)
		assert.Equal(t, secrets.TokenSigningRSAPrivateKey.PublicKey, *k)
	})
}

func TestProvider_ValidateAccessToken(t *testing.T) {
	p, _ := newMockProvider(t)
	ctx := context.Background()

	t.Run("Invalid JWT", func(t *testing.T) {
		_, err := p.ValidateAccessToken(ctx, "myserver", "abc")
		assert.Error(t, err)
	})

	t.Run("Invalid Signature", func(t *testing.T) {
		_, err := p.ValidateAccessToken(ctx, "myserver", "sampleIDToken")
		assert.Error(t, err)
	})

	t.Run("Valid", func(t *testing.T) {
		ctx := context.Background()
		secrets, err := auth.NewSecrets()
		assert.NoError(t, err)

		sm := &mocks.SecretManager{}
		sm.OnGet(ctx, config.SecretNameClaimSymmetricKey).Return(base64.RawStdEncoding.EncodeToString(secrets.TokenHashKey), nil)
		sm.OnGet(ctx, config.SecretNameCookieBlockKey).Return(base64.RawStdEncoding.EncodeToString(secrets.CookieBlockKey), nil)
		sm.OnGet(ctx, config.SecretNameCookieHashKey).Return(base64.RawStdEncoding.EncodeToString(secrets.CookieHashKey), nil)

		privBytes := x509.MarshalPKCS1PrivateKey(secrets.TokenSigningRSAPrivateKey)
		var buf bytes.Buffer
		assert.NoError(t, pem.Encode(&buf, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes}))
		sm.OnGet(ctx, config.SecretNameTokenSigningRSAKey).Return(buf.String(), nil)
		sm.OnGet(ctx, config.SecretNameOldTokenSigningRSAKey).Return(buf.String(), nil)

		p, err := NewProvider(ctx, config.DefaultConfig.AppAuth.SelfAuthServer, sm)
		assert.NoError(t, err)

		// create a signer for rsa 256
		tok := jwtgo.New(jwtgo.GetSigningMethod("RS256"))

		keySet, err := newJSONWebKeySet([]rsa.PublicKey{secrets.TokenSigningRSAPrivateKey.PublicKey})
		assert.NoError(t, err)

		// set our claims
		k, found := keySet.Get(0)
		assert.True(t, found)

		tok.Header[KeyIDClaim] = k.KeyID()
		tok.Claims = &CustomClaimsExample{
			StandardClaims: &jwtgo.StandardClaims{
				Audience:  "https://myserver",
				ExpiresAt: time.Now().Add(time.Minute * 1).Unix(),
			},
			ClientID: "client-1",
			UserID:   "1234",
			Scopes:   []string{"all"},
		}

		// Create token string
		str, err := tok.SignedString(secrets.TokenSigningRSAPrivateKey)
		assert.NoError(t, err)

		identity, err := p.ValidateAccessToken(ctx, "https://myserver", str)
		assert.NoError(t, err)
		assert.False(t, identity.IsEmpty())
	})
}
