package interfaces

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"
)

type CookieHandler interface {
	RetrieveTokenValues(ctx context.Context, request *http.Request) (accessToken string, refreshToken string, err error)
	SetTokenCookies(ctx context.Context, writer http.ResponseWriter, token *oauth2.Token) error
	DeleteCookies(ctx context.Context, writer http.ResponseWriter)
}
