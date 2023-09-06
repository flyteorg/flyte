package authzserver

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/stretchr/testify/assert"

	"github.com/lestrrat-go/jwx/jwk"

	"github.com/coreos/go-oidc"
	"github.com/flyteorg/flyteadmin/auth/config"
	authConfig "github.com/flyteorg/flyteadmin/auth/config"
	stdlibConfig "github.com/flyteorg/flytestdlib/config"
)

func newMockResourceServer(t testing.TB, publicKey rsa.PublicKey) (resourceServer ResourceServer, closer func()) {
	ctx := context.Background()
	dummy := ""
	serverURL := &dummy
	hf := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/oauth-authorization-server" {
			w.Header().Set("Content-Type", "application/json")
			_, err := io.WriteString(w, strings.ReplaceAll(`{
				"issuer": "https://whatever.okta.com",
				"authorization_endpoint": "https://example.com/auth",
				"token_endpoint": "https://example.com/token",
				"jwks_uri": "{URL}/keys",
				"id_token_signing_alg_values_supported": ["RS256"]
			}`, "{URL}", *serverURL))

			if !assert.NoError(t, err) {
				t.FailNow()
			}

			return
		} else if r.URL.Path == "/keys" {
			keys := jwk.NewSet()
			key := jwk.NewRSAPublicKey()
			err := key.FromRaw(&publicKey)
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}

			keys.Add(key)
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

			return
		}

		http.NotFound(w, r)
	}

	s := httptest.NewServer(http.HandlerFunc(hf))
	*serverURL = s.URL

	http.DefaultClient = s.Client()

	r, err := NewOAuth2ResourceServer(ctx, authConfig.ExternalAuthorizationServer{
		BaseURL:         stdlibConfig.URL{URL: *config.MustParseURL(s.URL)},
		AllowedAudience: []string{"https://localhost"},
	}, stdlibConfig.URL{})
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	return r, func() {
		s.Close()
	}
}

func TestResourceServer_ValidateAccessToken(t *testing.T) {
	sampleRSAKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	r, closer := newMockResourceServer(t, sampleRSAKey.PublicKey)
	defer closer()

	t.Run("No signature", func(t *testing.T) {
		sampleIDToken, err := jwt.NewWithClaims(jwt.SigningMethodRS512, jwt.RegisteredClaims{
			Audience:  r.allowedAudience,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "localhost",
			Subject:   "someone",
		}).SignedString(sampleRSAKey)
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		parts := strings.Split(sampleIDToken, ".")
		sampleIDToken = strings.Join(parts[:len(parts)-1], ".") + "."

		_, err = r.ValidateAccessToken(context.Background(), "myserver", sampleIDToken)
		if !assert.Error(t, err) {
			t.FailNow()
		}

		assert.Contains(t, err.Error(), "failed to verify id token signature")
	})

	t.Run("Invalid signature", func(t *testing.T) {
		sampleRSAKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		sampleIDToken, err := jwt.NewWithClaims(jwt.SigningMethodRS512, jwt.RegisteredClaims{
			Audience:  r.allowedAudience,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "localhost",
			Subject:   "someone",
		}).SignedString(sampleRSAKey)
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		_, err = r.ValidateAccessToken(context.Background(), "myserver", sampleIDToken)
		if !assert.Error(t, err) {
			t.FailNow()
		}

		assert.Contains(t, err.Error(), "failed to verify id token signature")
	})

	t.Run("Invalid audience", func(t *testing.T) {
		sampleIDToken, err := jwt.NewWithClaims(jwt.SigningMethodRS512, jwt.RegisteredClaims{
			Audience:  []string{"https://hello world"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "localhost",
			Subject:   "someone",
		}).SignedString(sampleRSAKey)
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		_, err = r.ValidateAccessToken(context.Background(), "myserver", sampleIDToken)
		if !assert.Error(t, err) {
			t.FailNow()
		}

		assert.Contains(t, err.Error(), "invalid audience")
	})

	t.Run("Expired token", func(t *testing.T) {
		sampleIDToken, err := jwt.NewWithClaims(jwt.SigningMethodRS512, jwt.StandardClaims{
			Audience:  r.allowedAudience[0],
			ExpiresAt: time.Now().Add(-time.Hour).Unix(),
			IssuedAt:  time.Now().Add(-2 * time.Hour).Unix(),
			Issuer:    "localhost",
			Subject:   "someone",
		}).SignedString(sampleRSAKey)
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		_, err = r.ValidateAccessToken(context.Background(), "myserver", sampleIDToken)
		if !assert.Error(t, err) {
			t.FailNow()
		}

		assert.Contains(t, err.Error(), "failed to validate token: Token is expired")
	})
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
		cfg           authConfig.ExternalAuthorizationServer
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
			got, err := getJwksForIssuer(tt.args.ctx, tt.args.issuerBaseURL, tt.args.cfg)
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
