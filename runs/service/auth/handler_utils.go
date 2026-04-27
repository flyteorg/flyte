package auth

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"google.golang.org/grpc/metadata"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/runs/service/auth/config"
)

const (
	metadataXForwardedHost = "x-forwarded-host"
	metadataAuthority      = ":authority"
)

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
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}

	forwardedHost := getMetadataValue(md, metadataXForwardedHost)
	if len(forwardedHost) == 0 {
		forwardedHost = getMetadataValue(md, metadataAuthority)
	}

	if len(forwardedHost) == 0 {
		return nil
	}

	u, _ := url.Parse("https://" + forwardedHost)
	return u
}

// getMetadataValue retrieves the first value for a given key from gRPC metadata.
func getMetadataValue(md metadata.MD, key string) string {
	vals := md.Get(key)
	if len(vals) == 0 {
		return ""
	}
	return vals[0]
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

// wildcardMatch checks if hostname matches a wildcard pattern (only one level deep)
// Supports patterns like "*.union.ai" matching "tenant1.union.ai"
func wildcardMatch(hostname, pattern string) bool {
	if strings.HasPrefix(pattern, "*.") {
		urlParts := strings.SplitN(hostname, ".", 2)
		if len(urlParts) < 2 {
			return false
		}
		return urlParts[1] == pattern[2:]
	}
	return hostname == pattern
}

// buildURL constructs a URL using the authorized template but with the matched hostname
func buildURL(authorizedURL *url.URL, matchedHostname string) *url.URL {
	result := *authorizedURL // Copy the URL to avoid modifying the original
	if authorizedURL.Port() != "" {
		result.Host = matchedHostname + ":" + authorizedURL.Port()
	} else {
		result.Host = matchedHostname
	}
	return &result
}

// GetPublicURL attempts to retrieve the public url of the service. If httpPublicUri is set in the config, it takes
// precedence. If the request is not nil and has a host set, it comes second and lastly it attempts to retrieve the url
// from context if set (e.g. by gRPC gateway).
func GetPublicURL(ctx context.Context, req *http.Request, cfg config.Config) *url.URL {
	u := FirstURL(URLFromRequest(req), URLFromContext(ctx))
	var hostMatching *url.URL
	var hostAndPortMatching *url.URL
	var matchedHostname string

	for i, authorized := range cfg.AuthorizedURIs {
		if u == nil {
			return &authorized.URL
		}

		if wildcardMatch(u.Hostname(), authorized.Hostname()) {
			matchedHostname = u.Hostname()
			hostMatching = &cfg.AuthorizedURIs[i].URL
			if u.Port() == authorized.Port() {
				hostAndPortMatching = &cfg.AuthorizedURIs[i].URL
			}

			if u.Scheme == authorized.Scheme {
				return buildURL(&cfg.AuthorizedURIs[i].URL, matchedHostname)
			}
		}
	}

	if hostAndPortMatching != nil {
		return buildURL(hostAndPortMatching, matchedHostname)
	}

	if hostMatching != nil {
		return buildURL(hostMatching, matchedHostname)
	}

	if len(cfg.AuthorizedURIs) > 0 {
		return &cfg.AuthorizedURIs[0].URL
	}

	return u
}

// GetIssuer returns the issuer from SelfAuthServer config, or falls back to public URL.
func GetIssuer(ctx context.Context, req *http.Request, cfg config.Config) string {
	if configIssuer := cfg.AppAuth.SelfAuthServer.Issuer; len(configIssuer) > 0 {
		return configIssuer
	}

	return GetPublicURL(ctx, req, cfg).String()
}

// isAuthorizedRedirectURL checks if a redirect URL matches an authorized URL pattern.
func isAuthorizedRedirectURL(u *url.URL, authorizedURL *url.URL) bool {
	if u == nil || authorizedURL == nil {
		return false
	}
	if u.Scheme != authorizedURL.Scheme || u.Port() != authorizedURL.Port() {
		return false
	}
	return wildcardMatch(u.Hostname(), authorizedURL.Hostname())
}

// GetRedirectURLAllowed checks whether a redirect URL is in the list of authorized URIs.
func GetRedirectURLAllowed(ctx context.Context, urlRedirectParam string, authorizedURIs []config.URL) bool {
	if len(urlRedirectParam) == 0 {
		logger.Debugf(ctx, "not validating whether empty redirect url is authorized")
		return true
	}
	redirectURL, err := url.Parse(urlRedirectParam)
	if err != nil {
		logger.Debugf(ctx, "failed to parse user-supplied redirect url: %s with err: %v", urlRedirectParam, err)
		return false
	}
	if redirectURL.Host == "" {
		logger.Debugf(ctx, "not validating whether relative redirect url is authorized")
		return true
	}
	logger.Debugf(ctx, "validating whether redirect url: %s is authorized", redirectURL)
	for i := range authorizedURIs {
		if isAuthorizedRedirectURL(redirectURL, &authorizedURIs[i].URL) {
			logger.Debugf(ctx, "authorizing redirect url: %s against authorized uri: %s", redirectURL.String(), authorizedURIs[i].String())
			return true
		}
	}
	logger.Debugf(ctx, "not authorizing redirect url: %s", redirectURL.String())
	return false
}
