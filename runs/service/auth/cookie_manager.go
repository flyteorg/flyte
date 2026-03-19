package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/oauth2"

	authpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/auth"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/runs/service/auth/config"
)

// CookieManager manages encrypted cookie operations for auth tokens.
type CookieManager struct {
	hashKey        []byte
	blockKey       []byte
	domain         string
	sameSitePolicy config.SameSite
}

func NewCookieManager(ctx context.Context, hashKeyEncoded, blockKeyEncoded string, cookieSettings config.CookieSettings) (CookieManager, error) {
	logger.Infof(ctx, "Instantiating cookie manager")

	hashKey, err := base64.RawStdEncoding.DecodeString(hashKeyEncoded)
	if err != nil {
		return CookieManager{}, fmt.Errorf("error decoding hash key bytes: %w", err)
	}

	blockKey, err := base64.RawStdEncoding.DecodeString(blockKeyEncoded)
	if err != nil {
		return CookieManager{}, fmt.Errorf("error decoding block key bytes: %w", err)
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

// RetrieveTokenValues retrieves id, access and refresh tokens from cookies if they exist.
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
		// Refresh tokens are optional.
		logger.Infof(ctx, "Refresh token doesn't exist or failed to read it. Ignoring this error. Error: %v", err)
		err = nil
	}

	return
}

func (c CookieManager) SetUserInfoCookie(ctx context.Context, writer http.ResponseWriter, userInfo *authpb.UserInfoResponse) error {
	raw, err := json.Marshal(userInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal user info to store in a cookie: %w", err)
	}

	return c.SetUserInfoCookieRaw(ctx, writer, string(raw))
}

func (c CookieManager) SetUserInfoCookieRaw(ctx context.Context, writer http.ResponseWriter, userInfoStr string) error {
	userInfoCookie, err := NewSecureCookie(userInfoCookieName, userInfoStr, c.hashKey, c.blockKey, c.domain, c.getHTTPSameSitePolicy())
	if err != nil {
		logger.Errorf(ctx, "Error generating encrypted user info cookie %s", err)
		return err
	}

	http.SetCookie(writer, &userInfoCookie)
	return nil
}

func (c CookieManager) RetrieveUserInfo(ctx context.Context, request *http.Request) (*authpb.UserInfoResponse, error) {
	userInfoCookie, err := retrieveSecureCookie(ctx, request, userInfoCookieName, c.hashKey, c.blockKey)
	if err != nil {
		return nil, err
	}

	res := authpb.UserInfoResponse{}
	err = json.Unmarshal([]byte(userInfoCookie), &res)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user info cookie: %w", err)
	}

	return &res, nil
}

func (c CookieManager) RetrieveAuthCodeRequest(ctx context.Context, request *http.Request) (string, error) {
	return retrieveSecureCookie(ctx, request, authCodeCookieName, c.hashKey, c.blockKey)
}

func (c CookieManager) SetAuthCodeCookie(ctx context.Context, writer http.ResponseWriter, authRequestURL string) error {
	authCodeCookie, err := NewSecureCookie(authCodeCookieName, authRequestURL, c.hashKey, c.blockKey, c.domain, c.getHTTPSameSitePolicy())
	if err != nil {
		logger.Errorf(ctx, "Error generating encrypted auth code cookie %s", err)
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
	idToken, accessToken, refreshToken, err := ExtractTokensFromOauthToken(token)
	if err != nil {
		logger.Errorf(ctx, "Unable to read all token values from oauth token: %s", err)
		return fmt.Errorf("unable to read all token values from oauth token: %w", err)
	}

	idCookie, err := NewSecureCookie(idTokenCookieName, idToken, c.hashKey, c.blockKey, c.domain, c.getHTTPSameSitePolicy())
	if err != nil {
		logger.Errorf(ctx, "Error generating encrypted id token cookie %s", err)
		return err
	}

	http.SetCookie(writer, &idCookie)

	err = c.StoreAccessToken(ctx, accessToken, writer)
	if err != nil {
		logger.Errorf(ctx, "Error storing access token %s", err)
		return err
	}

	// Set the refresh cookie if there is a refresh token
	if len(refreshToken) > 0 {
		refreshCookie, err := NewSecureCookie(refreshTokenCookieName, token.RefreshToken, c.hashKey, c.blockKey, c.domain, c.getHTTPSameSitePolicy())
		if err != nil {
			logger.Errorf(ctx, "Error generating encrypted refresh token cookie %s", err)
			return err
		}
		http.SetCookie(writer, &refreshCookie)
	}

	return nil
}

func (c CookieManager) getLogoutCookie(name string) *http.Cookie {
	return &http.Cookie{
		Name:     name,
		Value:    "",
		Domain:   c.domain,
		MaxAge:   0,
		HttpOnly: true,
		Secure:   config.GetConfig().SecureCookie,
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
	switch c.sameSitePolicy {
	case config.SameSiteDefaultMode:
		return http.SameSiteDefaultMode
	case config.SameSiteLaxMode:
		return http.SameSiteLaxMode
	case config.SameSiteStrictMode:
		return http.SameSiteStrictMode
	case config.SameSiteNoneMode:
		return http.SameSiteNoneMode
	default:
		return http.SameSiteDefaultMode
	}
}
