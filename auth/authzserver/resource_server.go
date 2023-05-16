package authzserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"strings"

	jwtgo "github.com/golang-jwt/jwt/v4"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flytestdlib/config"

	"github.com/coreos/go-oidc"
	authConfig "github.com/flyteorg/flyteadmin/auth/config"
	"github.com/flyteorg/flyteadmin/auth/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
	"golang.org/x/oauth2"
)

// External auth server implementation

// ResourceServer authorizes access requests issued by an external Authorization Server.
type ResourceServer struct {
	signatureVerifier oidc.KeySet
	allowedAudience   []string
}

func (r ResourceServer) ValidateAccessToken(ctx context.Context, expectedAudience, tokenStr string) (interfaces.IdentityContext, error) {
	_, err := r.signatureVerifier.VerifySignature(ctx, tokenStr)
	if err != nil {
		return nil, err
	}

	t, _, err := jwtgo.NewParser().ParseUnverified(tokenStr, jwtgo.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}

	if err = t.Claims.Valid(); err != nil {
		return nil, fmt.Errorf("failed to validate token: %v", err)
	}

	return verifyClaims(sets.NewString(append(r.allowedAudience, expectedAudience)...), t.Claims.(jwtgo.MapClaims))
}

func doRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	client := http.DefaultClient
	if c, ok := ctx.Value(oauth2.HTTPClient).(*http.Client); ok {
		client = c
	}
	return client.Do(req.WithContext(ctx))
}

func unmarshalResp(r *http.Response, body []byte, v interface{}) error {
	err := json.Unmarshal(body, &v)
	if err == nil {
		return nil
	}
	ct := r.Header.Get("Content-Type")
	mediaType, _, parseErr := mime.ParseMediaType(ct)
	if parseErr == nil && mediaType == "application/json" {
		return fmt.Errorf("got Content-Type = application/json, but could not unmarshal as JSON: %v", err)
	}
	return fmt.Errorf("expected Content-Type = application/json, got %q: %v", ct, err)
}

func getJwksForIssuer(ctx context.Context, issuerBaseURL url.URL, cfg authConfig.ExternalAuthorizationServer) (keySet oidc.KeySet, err error) {
	customMetadataURL := cfg.MetadataEndpointURL.URL

	issuerBaseURL.Path = strings.TrimSuffix(issuerBaseURL.Path, "/") + "/"
	var wellKnown *url.URL
	if len(customMetadataURL.String()) > 0 {
		wellKnown = issuerBaseURL.ResolveReference(&customMetadataURL)
	} else {
		wellKnown = issuerBaseURL.ResolveReference(oauth2MetadataEndpoint)
	}

	httpClient := &http.Client{}

	if len(cfg.HTTPProxyURL.String()) > 0 {
		// create a transport that uses the proxy
		transport := &http.Transport{
			Proxy: http.ProxyURL(&cfg.HTTPProxyURL.URL),
		}
		httpClient.Transport = transport
	}

	req, err := http.NewRequest(http.MethodGet, wellKnown.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", resp.Status, body)
	}

	p := &service.OAuth2MetadataResponse{}
	err = unmarshalResp(resp, body, &p)
	if err != nil {
		return nil, fmt.Errorf("failed to decode provider discovery object: %v", err)
	}

	return oidc.NewRemoteKeySet(oidc.ClientContext(ctx, httpClient), p.JwksUri), nil
}

// NewOAuth2ResourceServer initializes a new OAuth2ResourceServer.
func NewOAuth2ResourceServer(ctx context.Context, cfg authConfig.ExternalAuthorizationServer, fallbackBaseURL config.URL) (ResourceServer, error) {
	u := cfg.BaseURL
	if len(u.String()) == 0 {
		u = fallbackBaseURL
	}

	verifier, err := getJwksForIssuer(ctx, u.URL, cfg)
	if err != nil {
		return ResourceServer{}, err
	}

	return ResourceServer{
		signatureVerifier: verifier,
		allowedAudience:   cfg.AllowedAudience,
	}, nil
}
