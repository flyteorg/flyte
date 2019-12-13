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
func (c CookieManager) RetrieveTokenValues(ctx context.Context, request *http.Request) (accessToken string,
	refreshToken string, err error) {

	jwtCookie, err := request.Cookie(accessTokenCookieName)
	if err != nil || jwtCookie == nil {
		logger.Errorf(ctx, "Could not detect existing access token cookie")
		return
	}
	logger.Debugf(ctx, "Existing JWT cookie found")
	accessToken, err = ReadSecureCookie(ctx, *jwtCookie, c.hashKey, c.blockKey)
	if err != nil {
		logger.Errorf(ctx, "Error reading existing secure JWT cookie %s", err)
		return
	}

	refreshCookie, err := request.Cookie(refreshTokenCookieName)
	if err != nil || refreshCookie == nil {
		logger.Debugf(ctx, "Could not detect existing access token cookie")
		return
	}
	logger.Debugf(ctx, "Existing refresh cookie found")
	refreshToken, err = ReadSecureCookie(ctx, *refreshCookie, c.hashKey, c.blockKey)
	if err != nil {
		logger.Errorf(ctx, "Error reading existing secure refresh cookie %s", err)
		return
	}
	return
}

func (c CookieManager) SetTokenCookies(ctx context.Context, writer http.ResponseWriter, token *oauth2.Token) error {
	if token == nil {
		logger.Errorf(ctx, "Attempting to set cookies with nil token")
		return errors.Errorf(ErrTokenNil, "Attempting to set cookies with nil token")
	}

	jwtCookie, err := NewSecureCookie(accessTokenCookieName, token.AccessToken, c.hashKey, c.blockKey)
	if err != nil {
		logger.Errorf(ctx, "Error generating encrypted JWT cookie %s", err)
		return err
	}
	http.SetCookie(writer, &jwtCookie)

	// Set the refresh cookie if there is a refresh token
	if token.RefreshToken != "" {
		refreshCookie, err := NewSecureCookie(refreshTokenCookieName, token.RefreshToken, c.hashKey, c.blockKey)
		if err != nil {
			logger.Errorf(ctx, "Error generating encrypted refresh cookie %s", err)
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
