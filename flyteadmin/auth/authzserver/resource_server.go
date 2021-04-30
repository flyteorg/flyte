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

	"github.com/flyteorg/flytestdlib/config"

	"github.com/coreos/go-oidc"
	"github.com/flyteorg/flyteadmin/auth"

	authConfig "github.com/flyteorg/flyteadmin/auth/config"
	"github.com/flyteorg/flyteadmin/auth/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
	"golang.org/x/oauth2"
)

// ResourceServer authorizes access requests issued by an external Authorization Server.
type ResourceServer struct {
	signatureVerifier oidc.KeySet
}

func (r ResourceServer) ValidateAccessToken(ctx context.Context, expectedAudience, tokenStr string) (interfaces.IdentityContext, error) {
	raw, err := r.signatureVerifier.VerifySignature(ctx, tokenStr)
	if err != nil {
		return nil, err
	}

	claimsRaw := map[string]interface{}{}
	if err = json.Unmarshal(raw, &claimsRaw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user info claim into UserInfo type. Error: %w", err)
	}

	return verifyClaims(expectedAudience, claimsRaw)
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

func getJwksForIssuer(ctx context.Context, issuerBaseURL url.URL) (oidc.KeySet, error) {
	u, err := url.Parse(auth.OAuth2MetadataEndpoint)
	if err != nil {
		return nil, err
	}

	issuerBaseURL.Path = strings.TrimSuffix(issuerBaseURL.Path, "/") + "/"
	wellKnown := issuerBaseURL.ResolveReference(u)
	req, err := http.NewRequest(http.MethodGet, wellKnown.String(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := doRequest(ctx, req)
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

	return oidc.NewRemoteKeySet(ctx, p.JwksUri), nil
}

// NewOAuth2ResourceServer initializes a new OAuth2ResourceServer.
func NewOAuth2ResourceServer(ctx context.Context, cfg authConfig.ExternalAuthorizationServer, fallbackBaseURL config.URL) (ResourceServer, error) {
	u := cfg.BaseURL
	if len(u.String()) == 0 {
		u = fallbackBaseURL
	}

	verifier, err := getJwksForIssuer(ctx, u.URL)
	if err != nil {
		return ResourceServer{}, err
	}

	return ResourceServer{
		signatureVerifier: verifier,
	}, nil
}
