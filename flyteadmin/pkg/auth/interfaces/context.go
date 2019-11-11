package interfaces

import (
	"net/http"
	"net/url"

	"github.com/coreos/go-oidc"
	"github.com/lyft/flyteadmin/pkg/auth/config"
	"golang.org/x/oauth2"
)

//go:generate mockery -name=AuthenticationContext -case=underscore

// This interface is a convenience wrapper object that holds all the utilities necessary to run Flyte Admin behind authentication
// It is constructed at the root server layer, and passed around to the various auth handlers and utility functions/objects.
type AuthenticationContext interface {
	OAuth2Config() *oauth2.Config
	Claims() config.Claims
	OidcProvider() *oidc.Provider
	CookieManager() CookieHandler
	Options() config.OAuthOptions
	GetUserInfoURL() *url.URL
	GetHTTPClient() *http.Client
}
