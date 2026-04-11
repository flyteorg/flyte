package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/metadata"
)

func TestNewOAuthTokenFromRaw(t *testing.T) {
	tok := NewOAuthTokenFromRaw("access", "refresh", "id-token")
	require.NotNil(t, tok)
	assert.Equal(t, "access", tok.AccessToken)
	assert.Equal(t, "refresh", tok.RefreshToken)
	assert.Equal(t, "id-token", tok.Extra(idTokenExtra))
}

func TestExtractTokensFromOauthToken(t *testing.T) {
	src := NewOAuthTokenFromRaw("a", "r", "i")
	id, access, refresh, err := ExtractTokensFromOauthToken(src)
	require.NoError(t, err)
	assert.Equal(t, "i", id)
	assert.Equal(t, "a", access)
	assert.Equal(t, "r", refresh)
}

func TestExtractTokensFromOauthToken_Nil(t *testing.T) {
	_, _, _, err := ExtractTokensFromOauthToken(nil)
	assert.Error(t, err)
}

func TestExtractTokensFromOauthToken_MissingIDToken(t *testing.T) {
	// bare oauth2.Token without id_token extra should fail
	tok := &oauth2.Token{AccessToken: "a", RefreshToken: "r"}
	_, _, _, err := ExtractTokensFromOauthToken(tok)
	assert.Error(t, err)
}

func ctxWithMD(pairs ...string) context.Context {
	md := metadata.Pairs(pairs...)
	return metadata.NewIncomingContext(context.Background(), md)
}

func TestBearerTokenFromMD(t *testing.T) {
	ctx := ctxWithMD(DefaultAuthorizationHeader, "Bearer my-token")
	tok, err := bearerTokenFromMD(ctx)
	require.NoError(t, err)
	assert.Equal(t, "my-token", tok)
}

func TestBearerTokenFromMD_NoMetadata(t *testing.T) {
	_, err := bearerTokenFromMD(context.Background())
	assert.Error(t, err)
}

func TestBearerTokenFromMD_MissingHeader(t *testing.T) {
	_, err := bearerTokenFromMD(ctxWithMD("other", "v"))
	assert.Error(t, err)
}

func TestBearerTokenFromMD_WrongScheme(t *testing.T) {
	_, err := bearerTokenFromMD(ctxWithMD(DefaultAuthorizationHeader, "IDToken abc"))
	assert.Error(t, err)
}

func TestBearerTokenFromMD_Blank(t *testing.T) {
	_, err := bearerTokenFromMD(ctxWithMD(DefaultAuthorizationHeader, "Bearer "))
	assert.Error(t, err)
}

func TestIDTokenFromMD(t *testing.T) {
	ctx := ctxWithMD(DefaultAuthorizationHeader, "IDToken my-id-token")
	tok, err := idTokenFromMD(ctx)
	require.NoError(t, err)
	assert.Equal(t, "my-id-token", tok)
}

func TestIDTokenFromMD_WrongScheme(t *testing.T) {
	_, err := idTokenFromMD(ctxWithMD(DefaultAuthorizationHeader, "Bearer abc"))
	assert.Error(t, err)
}

func TestIDTokenFromMD_Blank(t *testing.T) {
	_, err := idTokenFromMD(ctxWithMD(DefaultAuthorizationHeader, "IDToken "))
	assert.Error(t, err)
}

func TestIDTokenFromMD_NoMetadata(t *testing.T) {
	_, err := idTokenFromMD(context.Background())
	assert.Error(t, err)
}
