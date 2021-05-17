package authzserver

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lestrrat-go/jwx/jwk"

	"github.com/coreos/go-oidc"
	"github.com/flyteorg/flyteadmin/auth/config"
	authConfig "github.com/flyteorg/flyteadmin/auth/config"
	stdlibConfig "github.com/flyteorg/flytestdlib/config"
)

func newMockResourceServer(t testing.TB) ResourceServer {
	ctx := context.Background()
	dummy := ""
	serverURL := &dummy
	hf := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/oauth-authorization-server" {
			w.Header().Set("Content-Type", "application/json")
			_, err := io.WriteString(w, strings.ReplaceAll(`{
				"issuer": "https://dev-14186422.okta.com",
				"authorization_endpoint": "https://example.com/auth",
				"token_endpoint": "https://example.com/token",
				"jwks_uri": "URL/keys",
				"id_token_signing_alg_values_supported": ["RS256"]
			}`, "URL", *serverURL))

			if !assert.NoError(t, err) {
				t.FailNow()
			}

			return
		} else if r.URL.Path == "/keys" {
			keys := jwk.NewSet()
			raw, err := json.Marshal(keys)
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_, err = io.WriteString(w, string(raw))

			if !assert.NoError(t, err) {
				t.FailNow()
			}
		}

		http.NotFound(w, r)
	}

	s := httptest.NewServer(http.HandlerFunc(hf))
	defer s.Close()

	*serverURL = s.URL

	http.DefaultClient = s.Client()

	r, err := NewOAuth2ResourceServer(ctx, authConfig.ExternalAuthorizationServer{
		BaseURL: stdlibConfig.URL{URL: *config.MustParseURL(s.URL)},
	}, stdlibConfig.URL{})
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	return r
}

func TestNewOAuth2ResourceServer(t *testing.T) {
	newMockResourceServer(t)
}

func TestResourceServer_ValidateAccessToken(t *testing.T) {
	r := newMockResourceServer(t)
	_, err := r.ValidateAccessToken(context.Background(), "myserver", sampleIDToken)
	assert.Error(t, err)
}

func Test_doRequest(t *testing.T) {
	type args struct {
		ctx context.Context
		req *http.Request
	}
	tests := []struct {
		name    string
		args    args
		want    *http.Response
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := doRequest(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("doRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("doRequest() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getJwksForIssuer(t *testing.T) {
	type args struct {
		ctx           context.Context
		issuerBaseURL url.URL
		customMetaURL url.URL
	}
	tests := []struct {
		name    string
		args    args
		want    oidc.KeySet
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getJwksForIssuer(tt.args.ctx, tt.args.issuerBaseURL, tt.args.customMetaURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("getJwksForIssuer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getJwksForIssuer() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_unmarshalResp(t *testing.T) {
	type args struct {
		r    *http.Response
		body []byte
		v    interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := unmarshalResp(tt.args.r, tt.args.body, tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("unmarshalResp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
