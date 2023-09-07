package interfaces

import (
	"context"
	"net/http"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"

	"golang.org/x/oauth2"
)

//go:generate mockery -name=CookieHandler -output=mocks/ -case=underscore

type CookieHandler interface {
	SetTokenCookies(ctx context.Context, writer http.ResponseWriter, token *oauth2.Token) error
	RetrieveTokenValues(ctx context.Context, request *http.Request) (idToken, accessToken, refreshToken string, err error)

	SetUserInfoCookie(ctx context.Context, writer http.ResponseWriter, userInfo *service.UserInfoResponse) error
	RetrieveUserInfo(ctx context.Context, request *http.Request) (*service.UserInfoResponse, error)

	// SetAuthCodeCookie stores, in a cookie, the /authorize request url initiated by an app before executing OIdC protocol.
	// This enables the service to recover it after the user completes the login process in an external OIdC provider.
	SetAuthCodeCookie(ctx context.Context, writer http.ResponseWriter, authRequestURL string) error

	// RetrieveAuthCodeRequest retrieves the /authorize request url from stored cookie to complete the OAuth2 app auth
	// flow.
	RetrieveAuthCodeRequest(ctx context.Context, request *http.Request) (authRequestURL string, err error)
	DeleteCookies(ctx context.Context, writer http.ResponseWriter)
}
