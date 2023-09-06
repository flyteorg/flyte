package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/flyteorg/flytestdlib/errors"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/gorilla/securecookie"

	"github.com/flyteorg/flyteadmin/auth/interfaces"
)

const (
	// #nosec
	accessTokenCookieName = "flyte_at"
	// #nosec
	idTokenCookieName = "flyte_idt"
	// #nosec
	refreshTokenCookieName = "flyte_rt"
	// #nosec
	csrfStateCookieName = "flyte_csrf_state"
	// #nosec
	redirectURLCookieName = "flyte_redirect_location"

	// #nosec
	idTokenExtra = "id_token"

	// #nosec
	authCodeCookieName = "flyte_auth_code"

	// #nosec
	userInfoCookieName = "flyte_user_info"
)

const (
	ErrSecureCookie errors.ErrorCode = "SECURE_COOKIE_ERROR"
	// #nosec
	ErrInvalidCsrfToken errors.ErrorCode = "CSRF_TOKEN_VALIDATION_FAILED"
)

var AllowedChars = []rune("abcdefghijklmnopqrstuvwxyz1234567890")

func HashCsrfState(csrf string) string {
	shaBytes := sha256.Sum256([]byte(csrf))
	hash := hex.EncodeToString(shaBytes[:])
	return hash
}

func NewSecureCookie(cookieName, value string, hashKey, blockKey []byte, domain string, sameSiteMode http.SameSite) (http.Cookie, error) {
	s := securecookie.New(hashKey, blockKey)
	encoded, err := s.Encode(cookieName, value)
	if err != nil {
		return http.Cookie{}, errors.Wrapf(ErrSecureCookie, err, "Error creating secure cookie")
	}

	return http.Cookie{
		Name:     cookieName,
		Value:    encoded,
		Domain:   domain,
		SameSite: sameSiteMode,
	}, nil
}

func retrieveSecureCookie(ctx context.Context, request *http.Request, cookieName string, hashKey, blockKey []byte) (string, error) {
	cookie, err := request.Cookie(cookieName)
	if err != nil {
		logger.Infof(ctx, "Could not detect existing cookie [%v]. Error: %v", cookieName, err)
		return "", errors.Wrapf(ErrTokenNil, err, "Failure to retrieve cookie [%v]", cookieName)
	}

	if cookie == nil {
		logger.Infof(ctx, "Retrieved empty cookie [%v].", cookieName)
		return "", errors.Errorf(ErrTokenNil, "Retrieved empty cookie [%v]", cookieName)
	}

	logger.Debugf(ctx, "Existing [%v] cookie found", cookieName)
	token, err := ReadSecureCookie(ctx, *cookie, hashKey, blockKey)
	if err != nil {
		logger.Errorf(ctx, "Error reading existing secure cookie [%v]. Error: %s", cookieName, err)
		return "", errors.Errorf(ErrTokenNil, "Error reading existing secure cookie [%v]. Error: %s", cookieName, err)
	}

	if len(token) == 0 {
		logger.Errorf(ctx, "Read empty token from secure cookie [%v].", cookieName)
		return "", errors.Errorf(ErrTokenNil, "Read empty token from secure cookie [%v].", cookieName)
	}

	return token, nil
}

func ReadSecureCookie(ctx context.Context, cookie http.Cookie, hashKey, blockKey []byte) (string, error) {
	var s = securecookie.New(hashKey, blockKey)
	var value string
	var err error
	if err = s.Decode(cookie.Name, cookie.Value, &value); err == nil {
		return value, nil
	}
	logger.Errorf(ctx, "Error reading secure cookie %s %s", cookie.Name, err)
	return "", errors.Wrapf(ErrSecureCookie, err, "Error reading secure cookie %s", cookie.Name)
}

func NewCsrfToken(seed int64) string {
	rand.Seed(seed)
	csrfToken := [10]rune{}
	for i := 0; i < len(csrfToken); i++ {
		/* #nosec */
		csrfToken[i] = AllowedChars[rand.Intn(len(AllowedChars))]
	}
	return string(csrfToken[:])
}

func NewCsrfCookie() http.Cookie {
	csrfStateToken := NewCsrfToken(time.Now().UnixNano())
	return http.Cookie{
		Name:     csrfStateCookieName,
		Value:    csrfStateToken,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
	}
}

func VerifyCsrfCookie(ctx context.Context, request *http.Request) error {
	csrfState := request.FormValue(CsrfFormKey)
	if csrfState == "" {
		return errors.Errorf(ErrInvalidCsrfToken, "Empty state in callback, %s", request.Form)
	}
	csrfCookie, err := request.Cookie(csrfStateCookieName)
	if csrfCookie == nil || err != nil {
		return errors.Errorf(ErrInvalidCsrfToken, "Could not find csrf cookie %s", err)
	}
	if HashCsrfState(csrfCookie.Value) != csrfState {
		return errors.Errorf(ErrInvalidCsrfToken, "CSRF token does not match state %s, %s vs %s", csrfCookie.Value,
			HashCsrfState(csrfCookie.Value), csrfState)
	}
	return nil
}

// This function takes in a string and returns a cookie that's used to keep track of where to send the user after
// the OAuth2 login flow is complete.
func NewRedirectCookie(ctx context.Context, redirectURL string) *http.Cookie {
	urlObj, err := url.Parse(redirectURL)
	if err != nil || urlObj == nil {
		logger.Errorf(ctx, "Error creating redirect cookie %s %s", urlObj, err)
		return nil
	}

	if urlObj.EscapedPath() == "" {
		logger.Errorf(ctx, "Error parsing URL, redirect %s resolved to empty string", redirectURL)
		return nil
	}

	return &http.Cookie{
		Name:     redirectURLCookieName,
		Value:    urlObj.String(),
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
	}
}

// At the end of the OAuth flow, the server needs to send the user somewhere. This should have been stored as a cookie
// during the initial /login call. If that cookie is missing from the request, it will default to the one configured
// in this package's Config object.
func getAuthFlowEndRedirect(ctx context.Context, authCtx interfaces.AuthenticationContext, request *http.Request) string {
	queryParams := request.URL.Query()
	// Use the redirect URL specified in the request if one is available.
	if redirectURL := queryParams.Get(RedirectURLParameter); len(redirectURL) > 0 {
		return redirectURL
	}

	cookie, err := request.Cookie(redirectURLCookieName)
	if err != nil {
		logger.Debugf(ctx, "Could not detect end-of-flow redirect url cookie")
		return authCtx.Options().UserAuth.RedirectURL.String()
	}

	return cookie.Value
}
