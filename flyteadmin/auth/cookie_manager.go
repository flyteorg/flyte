package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"

	"github.com/flyteorg/flyte/flyteadmin/auth/config"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flyte/flytestdlib/errors"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

type CookieManager struct {
	hashKey        []byte
	blockKey       []byte
	domain         string
	sameSitePolicy config.SameSite
}

const (
	ErrB64Decoding errors.ErrorCode = "BINARY_DECODING_FAILED"
	// #nosec
	ErrTokenNil errors.ErrorCode = "EMPTY_OAUTH_TOKEN"
	// #nosec
	ErrNoIDToken errors.ErrorCode = "NO_ID_TOKEN_IN_RESPONSE"
)

func NewCookieManager(ctx context.Context, hashKeyEncoded, blockKeyEncoded string, cookieSettings config.CookieSettings) (CookieManager, error) {
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
		hashKey:        hashKey,
		blockKey:       blockKey,
		domain:         cookieSettings.Domain,
		sameSitePolicy: cookieSettings.SameSitePolicy,
	}, nil
}

func (c CookieManager) RetrieveAccessToken(ctx context.Context, request *http.Request) (string, error) {
	// If there is an old access token, we will retrieve it
	oldAccessToken, err := retrieveSecureCookie(ctx, request, accessTokenCookieName, c.hashKey, c.blockKey)
	if err == nil && oldAccessToken != "" {
		return oldAccessToken, nil
	}
	// If there is no old access token, we will retrieve the new split access token
	accessTokenFirstHalf, err := retrieveSecureCookie(ctx, request, accessTokenCookieNameSplitFirst, c.hashKey, c.blockKey)
	if err != nil {
		return "", err
	}
	accessTokenSecondHalf, err := retrieveSecureCookie(ctx, request, accessTokenCookieNameSplitSecond, c.hashKey, c.blockKey)
	if err != nil {
		return "", err
	}
	return accessTokenFirstHalf + accessTokenSecondHalf, nil
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

	accessToken, err = c.RetrieveAccessToken(ctx, request)
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

func (c CookieManager) SetUserInfoCookie(ctx context.Context, writer http.ResponseWriter, userInfo *service.UserInfoResponse) error {
	raw, err := json.Marshal(userInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal user info to store in a cookie. Error: %w", err)
	}

	userInfoCookie, err := NewSecureCookie(userInfoCookieName, string(raw), c.hashKey, c.blockKey, c.domain, c.getHTTPSameSitePolicy())
	if err != nil {
		logger.Errorf(ctx, "Error generating encrypted user info cookie %s", err)
		return err
	}

	http.SetCookie(writer, &userInfoCookie)

	return nil

}

func (c CookieManager) RetrieveUserInfo(ctx context.Context, request *http.Request) (*service.UserInfoResponse, error) {
	userInfoCookie, err := retrieveSecureCookie(ctx, request, userInfoCookieName, c.hashKey, c.blockKey)
	if err != nil {
		return nil, err
	}

	res := service.UserInfoResponse{}
	err = json.Unmarshal([]byte(userInfoCookie), &res)
	if err != nil {
		return nil, fmt.Errorf("failed to unmasharl user info cookie. Error: %w", err)
	}

	return &res, nil

}

func (c CookieManager) RetrieveAuthCodeRequest(ctx context.Context, request *http.Request) (authRequestURL string, err error) {
	authCodeCookie, err := retrieveSecureCookie(ctx, request, authCodeCookieName, c.hashKey, c.blockKey)
	if err != nil {
		return "", err
	}

	return authCodeCookie, nil
}

func (c CookieManager) SetAuthCodeCookie(ctx context.Context, writer http.ResponseWriter, authRequestURL string) error {
	authCodeCookie, err := NewSecureCookie(authCodeCookieName, authRequestURL, c.hashKey, c.blockKey, c.domain, c.getHTTPSameSitePolicy())
	if err != nil {
		logger.Errorf(ctx, "Error generating encrypted accesstoken cookie %s", err)
		return err
	}

	http.SetCookie(writer, &authCodeCookie)

	return nil
}

func (c CookieManager) StoreAccessToken(ctx context.Context, accessToken string, writer http.ResponseWriter) error {
	midpoint := len(accessToken) / 2
	firstHalf := accessToken[:midpoint]
	secondHalf := accessToken[midpoint:]
	atCookieFirst, err := NewSecureCookie(accessTokenCookieNameSplitFirst, firstHalf, c.hashKey, c.blockKey, c.domain, c.getHTTPSameSitePolicy())
	if err != nil {
		logger.Errorf(ctx, "Error generating encrypted accesstoken cookie first half %s", err)
		return err
	}
	http.SetCookie(writer, &atCookieFirst)
	atCookieSecond, err := NewSecureCookie(accessTokenCookieNameSplitSecond, secondHalf, c.hashKey, c.blockKey, c.domain, c.getHTTPSameSitePolicy())
	if err != nil {
		logger.Errorf(ctx, "Error generating encrypted accesstoken cookie second half %s", err)
		return err
	}
	http.SetCookie(writer, &atCookieSecond)
	return nil
}

func (c CookieManager) SetTokenCookies(ctx context.Context, writer http.ResponseWriter, token *oauth2.Token) error {
	if token == nil {
		logger.Errorf(ctx, "Attempting to set cookies with nil token")
		return errors.Errorf(ErrTokenNil, "Attempting to set cookies with nil token")
	}

	err := c.StoreAccessToken(ctx, token.AccessToken, writer)

	if err != nil {
		logger.Errorf(ctx, "Error storing access token %s", err)
		return err
	}

	if idTokenRaw, converted := token.Extra(idTokenExtra).(string); converted {
		idCookie, err := NewSecureCookie(idTokenCookieName, idTokenRaw, c.hashKey, c.blockKey, c.domain, c.getHTTPSameSitePolicy())
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
		refreshCookie, err := NewSecureCookie(refreshTokenCookieName, token.RefreshToken, c.hashKey, c.blockKey, c.domain, c.getHTTPSameSitePolicy())
		if err != nil {
			logger.Errorf(ctx, "Error generating encrypted refresh token cookie %s", err)
			return err
		}
		http.SetCookie(writer, &refreshCookie)
	}

	return nil
}

func (c *CookieManager) getLogoutCookie(name string) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    "",
		Domain:   c.domain,
		MaxAge:   0,
		HttpOnly: true,
		Expires:  time.Now().Add(-1 * time.Hour),
	}
}

func (c CookieManager) DeleteCookies(_ context.Context, writer http.ResponseWriter) {
	http.SetCookie(writer, c.getLogoutCookie(accessTokenCookieName))
	http.SetCookie(writer, c.getLogoutCookie(accessTokenCookieNameSplitFirst))
	http.SetCookie(writer, c.getLogoutCookie(accessTokenCookieNameSplitSecond))
	http.SetCookie(writer, c.getLogoutCookie(refreshTokenCookieName))
	http.SetCookie(writer, c.getLogoutCookie(idTokenCookieName))
}

func (c CookieManager) getHTTPSameSitePolicy() http.SameSite {
	httpSameSite := http.SameSiteDefaultMode
	switch c.sameSitePolicy {
	case config.SameSiteDefaultMode:
		httpSameSite = http.SameSiteDefaultMode
	case config.SameSiteLaxMode:
		httpSameSite = http.SameSiteLaxMode
	case config.SameSiteStrictMode:
		httpSameSite = http.SameSiteStrictMode
	case config.SameSiteNoneMode:
		httpSameSite = http.SameSiteNoneMode
	}
	return httpSameSite
}
