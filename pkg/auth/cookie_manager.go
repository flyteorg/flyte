package auth

import (
	"context"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/lyft/flytestdlib/errors"
	"github.com/lyft/flytestdlib/logger"
	"golang.org/x/oauth2"
)

type CookieManager struct {
	hashKey  []byte
	blockKey []byte
}

const (
	ErrB64Decoding errors.ErrorCode = "BINARY_DECODING_FAILED"
	// #nosec
	ErrTokenNil errors.ErrorCode = "EMPTY_OAUTH_TOKEN"
	// #nosec
	ErrNoIDToken errors.ErrorCode = "NO_ID_TOKEN_IN_RESPONSE"
)

func NewCookieManager(ctx context.Context, hashKeyEncoded, blockKeyEncoded string) (CookieManager, error) {
	logger.Infof(ctx, "Instantiating cookie manager")

	hashKey, err := base64.RawStdEncoding.DecodeString(hashKeyEncoded)
	if err != nil {
		return CookieManager{}, errors.Wrapf(ErrB64Decoding, err, "Error decoding hash key bytes")
	}

	blockKey, err := base64.RawStdEncoding.DecodeString(blockKeyEncoded)
	if err != nil {
		return CookieManager{}, errors.Wrapf(ErrB64Decoding, err, "Error decoding block key bytes")
	}

	return CookieManager{
		hashKey:  hashKey,
		blockKey: blockKey,
	}, nil
}

// TODO: Separate refresh token from access token, remove named returns, and use stdlib errors.
// RetrieveTokenValues retrieves id, access and refresh tokens from cookies if they exist. The existence of a refresh token
// in a cookie is optional and hence failure to find or read that cookie is tolerated. An error is returned in case of failure
// to retrieve and read either the id or the access tokens.
func (c CookieManager) RetrieveTokenValues(ctx context.Context, request *http.Request) (idToken, accessToken,
	refreshToken string, err error) {

	idToken, err = retrieveSecureCookie(ctx, request, idTokenCookieName, c.hashKey, c.blockKey)
	if err != nil {
		return "", "", "", err
	}

	accessToken, err = retrieveSecureCookie(ctx, request, accessTokenCookieName, c.hashKey, c.blockKey)
	if err != nil {
		return "", "", "", err
	}

	refreshToken, err = retrieveSecureCookie(ctx, request, refreshTokenCookieName, c.hashKey, c.blockKey)
	if err != nil {
		// Refresh tokens are optional. Depending on the auth url (IdP specific) we might or might not receive a refresh
		// token. In case we do not, we will just have to redirect to IdP whenever access/id tokens expire.
		logger.Infof(ctx, "Refresh token doesn't exist or failed to read it. Ignoring this error. Error: %v", err)
		err = nil
	}

	return
}

func (c CookieManager) SetTokenCookies(ctx context.Context, writer http.ResponseWriter, token *oauth2.Token) error {
	if token == nil {
		logger.Errorf(ctx, "Attempting to set cookies with nil token")
		return errors.Errorf(ErrTokenNil, "Attempting to set cookies with nil token")
	}

	atCookie, err := NewSecureCookie(accessTokenCookieName, token.AccessToken, c.hashKey, c.blockKey)
	if err != nil {
		logger.Errorf(ctx, "Error generating encrypted accesstoken cookie %s", err)
		return err
	}

	http.SetCookie(writer, &atCookie)

	if idTokenRaw, converted := token.Extra(idTokenExtra).(string); converted {
		idCookie, err := NewSecureCookie(idTokenCookieName, idTokenRaw, c.hashKey, c.blockKey)
		if err != nil {
			logger.Errorf(ctx, "Error generating encrypted id token cookie %s", err)
			return err
		}

		http.SetCookie(writer, &idCookie)
	} else {
		logger.Errorf(ctx, "Response does not contain an id_token.")
		return errors.Errorf(ErrNoIDToken, "Response does not contain an id_token.")
	}

	// Set the refresh cookie if there is a refresh token
	if token.RefreshToken != "" {
		refreshCookie, err := NewSecureCookie(refreshTokenCookieName, token.RefreshToken, c.hashKey, c.blockKey)
		if err != nil {
			logger.Errorf(ctx, "Error generating encrypted refresh token cookie %s", err)
			return err
		}
		http.SetCookie(writer, &refreshCookie)
	}

	return nil
}

func getLogoutAccessCookie() *http.Cookie {
	return &http.Cookie{
		Name:     accessTokenCookieName,
		Value:    "",
		MaxAge:   0,
		HttpOnly: true,
		Expires:  time.Now().Add(-1 * time.Hour),
	}
}

func getLogoutRefreshCookie() *http.Cookie {
	return &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    "",
		MaxAge:   0,
		HttpOnly: true,
		Expires:  time.Now().Add(-1 * time.Hour),
	}
}

func (c CookieManager) DeleteCookies(ctx context.Context, writer http.ResponseWriter) {
	http.SetCookie(writer, getLogoutAccessCookie())
	http.SetCookie(writer, getLogoutRefreshCookie())
}
