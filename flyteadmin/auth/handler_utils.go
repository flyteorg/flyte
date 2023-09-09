package auth

import (
	"context"
	"net/http"
	"net/url"

	"github.com/flyteorg/flyteadmin/auth/config"

	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
)

const (
	metadataXForwardedHost = "x-forwarded-host"
	metadataAuthority      = ":authority"
)

//func NewMockOIdCProvider() (*oidc.Provider, error) {
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	var issuer string
//	hf := func(w http.ResponseWriter, r *http.Request) {
//		if r.URL.Path == "/.well-known/openid-configuration" {
//			w.Header().Set("Content-Type", "application/json")
//			io.WriteString(w, strings.ReplaceAll(`{
//				"issuer": "ISSUER",
//				"authorization_endpoint": "https://example.com/auth",
//				"token_endpoint": "https://example.com/token",
//				"jwks_uri": "ISSUER/keys",
//				"id_token_signing_alg_values_supported": ["RS256"]
//			}`, "ISSUER", issuer))
//			return
//		} else if r.URL.Path == "/keys" {
//			w.Header().Set("Content-Type", "application/json")
//			io.WriteString(w, `{"keys":[{"kty":"RSA","alg":"RS256","kid":"Z6dmZ_TXhduw-jUBZ6uEEzvnh-jhNO0YhemB7qa_LOc","use":"sig","e":"AQAB","n":"jyMcudBiz7XqeDIvxfMlmG4fvAUU7cl3R4iSIv_ahHanCcVRvqcXOsIknwn7i4rOUjP6MlH45uIYsaj6MuLYgoaIbC-Z823Tu4asoC-rGbpZgf-bMcJLxtZVBNsSagr_M0n8xA1oogHRF1LGRiD93wNr2b9OkKVbWnyNdASk5_xui024nVzakm2-RAEyaC048nHfnjVBvwo4BdJVDgBEK03fbkBCyuaZyE1ZQF545MTbD4keCv58prSCmbDRJgRk48FzaFnQeYTho-pUxXxM9pvhMykeI62WZ7diDfIc9isOpv6ALFOHgKy7Ihhve6pLIylLRTnn2qhHFkGPtU3djQ"}]}`)
//			return
//		}
//
//		http.NotFound(w, r)
//		return
//
//	}
//
//	s := httptest.NewServer(http.HandlerFunc(hf))
//	defer s.Close()
//
//	issuer = s.URL
//	return oidc.NewProvider(ctx, issuer)
//}

// URLFromRequest attempts to reconstruct the url from the request object. Or nil if not possible
func URLFromRequest(req *http.Request) *url.URL {
	if req == nil {
		return nil
	}

	// from browser req.RequestURI is "/login" and u.scheme is ""
	// from unit test req.RequestURI is "" and u is nil
	// That means that this function, URLFromRequest(req) returns https://localhost:8088 even though there's no SSL,
	// when the request is made from http://localhost:8088 in the web browser.
	// Given how this function is used however, it's okay - we're only picking which option to use from the list of
	// authorized URIs.
	u, _ := url.ParseRequestURI(req.RequestURI)
	if u != nil && u.IsAbs() {
		return u
	}

	if len(req.Host) == 0 {
		return nil
	}

	scheme := "https://"
	if req.URL != nil && len(req.URL.Scheme) > 0 {
		scheme = req.URL.Scheme + "://"
	}

	u, _ = url.Parse(scheme + req.Host)
	return u
}

// URLFromContext attempts to retrieve the original url from context. gRPC gateway sets metadata in context that refers
// to the original host. Or nil if metadata isn't set.
func URLFromContext(ctx context.Context) *url.URL {
	md := metautils.ExtractIncoming(ctx)
	forwardedHost := md.Get(metadataXForwardedHost)
	if len(forwardedHost) == 0 {
		forwardedHost = md.Get(metadataAuthority)
	}

	if len(forwardedHost) == 0 {
		return nil
	}

	u, _ := url.Parse("https://" + forwardedHost)
	return u
}

// FirstURL gets the first non-nil url from a list of given urls.
func FirstURL(urls ...*url.URL) *url.URL {
	for _, u := range urls {
		if u != nil {
			return u
		}
	}

	return nil
}

// GetPublicURL attempts to retrieve the public url of the service. If httpPublicUri is set in the config, it takes
// precedence. If the request is not nil and has a host set, it comes second and lastly it attempts to retrieve the url
// from context if set (e.g. by gRPC gateway).
func GetPublicURL(ctx context.Context, req *http.Request, cfg *config.Config) *url.URL {
	u := FirstURL(URLFromRequest(req), URLFromContext(ctx))
	var hostMatching *url.URL
	var hostAndPortMatching *url.URL

	for i, authorized := range cfg.AuthorizedURIs {
		if u == nil {
			return &authorized.URL
		}

		if u.Hostname() == authorized.Hostname() {
			hostMatching = &cfg.AuthorizedURIs[i].URL
			if u.Port() == authorized.Port() {
				hostAndPortMatching = &cfg.AuthorizedURIs[i].URL
			}

			if u.Scheme == authorized.Scheme {
				return &cfg.AuthorizedURIs[i].URL
			}
		}
	}

	if hostAndPortMatching != nil {
		return hostAndPortMatching
	}

	if hostMatching != nil {
		return hostMatching
	}

	if len(cfg.AuthorizedURIs) > 0 {
		return &cfg.AuthorizedURIs[0].URL
	}

	return u
}
