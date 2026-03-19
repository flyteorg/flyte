package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"

	"github.com/gorilla/securecookie"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/runs/service/auth/config"
)

const (
	// #nosec
	accessTokenCookieName = "flyte_at"
	// #nosec
	accessTokenCookieNameSplitFirst = "flyte_at_1"
	// #nosec
	accessTokenCookieNameSplitSecond = "flyte_at_2"
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
		return http.Cookie{}, fmt.Errorf("error creating secure cookie: %w", err)
	}

	return http.Cookie{
		Name:     cookieName,
		Value:    encoded,
		Domain:   domain,
		SameSite: sameSiteMode,
		HttpOnly: true,
		Secure:   config.GetConfig().SecureCookie,
		Path:     "/",
	}, nil
}

func retrieveSecureCookie(ctx context.Context, request *http.Request, cookieName string, hashKey, blockKey []byte) (string, error) {
	cookie, err := request.Cookie(cookieName)
	if err != nil {
		logger.Infof(ctx, "Could not detect existing cookie [%v]. Error: %v", cookieName, err)
		return "", fmt.Errorf("failure to retrieve cookie [%v]: %w", cookieName, err)
	}

	if cookie == nil {
		logger.Infof(ctx, "Retrieved empty cookie [%v].", cookieName)
		return "", fmt.Errorf("retrieved empty cookie [%v]", cookieName)
	}

	logger.Debugf(ctx, "Existing [%v] cookie found", cookieName)
	token, err := ReadSecureCookie(ctx, *cookie, hashKey, blockKey)
	if err != nil {
		logger.Errorf(ctx, "Error reading existing secure cookie [%v]. Error: %s", cookieName, err)
		return "", fmt.Errorf("error reading existing secure cookie [%v]: %w", cookieName, err)
	}

	if len(token) == 0 {
		logger.Errorf(ctx, "Read empty token from secure cookie [%v].", cookieName)
		return "", fmt.Errorf("read empty token from secure cookie [%v]", cookieName)
	}

	return token, nil
}

func ReadSecureCookie(ctx context.Context, cookie http.Cookie, hashKey, blockKey []byte) (string, error) {
	s := securecookie.New(hashKey, blockKey)
	var value string
	if err := s.Decode(cookie.Name, cookie.Value, &value); err == nil {
		return value, nil
	}
	logger.Errorf(ctx, "Error reading secure cookie %s", cookie.Name)
	return "", fmt.Errorf("error reading secure cookie %s", cookie.Name)
}

func NewCsrfToken(seed int64) string {
	r := rand.New(rand.NewSource(seed)) //nolint:gosec
	csrfToken := [10]rune{}
	for i := 0; i < len(csrfToken); i++ {
		csrfToken[i] = AllowedChars[r.Intn(len(AllowedChars))]
	}
	return string(csrfToken[:])
}

func NewCsrfCookie() http.Cookie {
	csrfStateToken := NewCsrfToken(rand.Int63()) //nolint:gosec
	return http.Cookie{
		Name:     csrfStateCookieName,
		Value:    csrfStateToken,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   config.GetConfig().SecureCookie,
	}
}

func VerifyCsrfCookie(_ context.Context, request *http.Request) error {
	csrfState := request.FormValue(CsrfFormKey)
	if csrfState == "" {
		return fmt.Errorf("empty state in callback, %s", request.Form)
	}
	csrfCookie, err := request.Cookie(csrfStateCookieName)
	if csrfCookie == nil || err != nil {
		return fmt.Errorf("could not find csrf cookie: %v", err)
	}
	if HashCsrfState(csrfCookie.Value) != csrfState {
		return fmt.Errorf("CSRF token does not match state %s, %s vs %s", csrfCookie.Value,
			HashCsrfState(csrfCookie.Value), csrfState)
	}
	return nil
}

// NewRedirectCookie creates a cookie to keep track of where to send the user after
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
		Secure:   config.GetConfig().SecureCookie,
	}
}

// GetAuthFlowEndRedirect returns the redirect URI according to data in request.
// At the end of the OAuth flow, the server needs to send the user somewhere. This should have been stored as a cookie
// during the initial /login call. If that cookie is missing from the request, it will default to the one configured.
func GetAuthFlowEndRedirect(ctx context.Context, defaultRedirect string, authorizedURIs []config.URL, request *http.Request) string {
	queryParams := request.URL.Query()
	if redirectURL := queryParams.Get(RedirectURLParameter); len(redirectURL) > 0 {
		if GetRedirectURLAllowed(ctx, redirectURL, authorizedURIs) {
			return redirectURL
		}
		logger.Warnf(ctx, "Rejecting unauthorized redirect_url from query parameter: %s", redirectURL)
		return defaultRedirect
	}

	cookie, err := request.Cookie(redirectURLCookieName)
	if err != nil {
		logger.Debugf(ctx, "Could not detect end-of-flow redirect url cookie")
		return defaultRedirect
	}

	if GetRedirectURLAllowed(ctx, cookie.Value, authorizedURIs) {
		return cookie.Value
	}
	logger.Warnf(ctx, "Rejecting unauthorized redirect_url from cookie: %s", cookie.Value)
	return defaultRedirect
}
