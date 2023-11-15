package admin

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/cache/mocks"
	adminMocks "github.com/flyteorg/flyte/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

// authMetadataServer is a fake AuthMetadataServer that takes in an AuthMetadataServer implementation (usually one
// initialized through mockery) and starts a local server that uses it to respond to grpc requests.
type authMetadataServer struct {
	s           *httptest.Server
	t           testing.TB
	port        int
	grpcServer  *grpc.Server
	netListener net.Listener
	impl        service.AuthMetadataServiceServer
	lck         *sync.RWMutex
}

func (s authMetadataServer) GetOAuth2Metadata(ctx context.Context, in *service.OAuth2MetadataRequest) (*service.OAuth2MetadataResponse, error) {
	return s.impl.GetOAuth2Metadata(ctx, in)
}

func (s authMetadataServer) GetPublicClientConfig(ctx context.Context, in *service.PublicClientAuthConfigRequest) (*service.PublicClientAuthConfigResponse, error) {
	return s.impl.GetPublicClientConfig(ctx, in)
}

func (s authMetadataServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var issuer string
	switch r.URL.Path {
	case "/.well-known/oauth-authorization-server":
		w.Header().Set("Content-Type", "application/json")
		_, err := io.WriteString(w, strings.ReplaceAll(`{
				"issuer": "https://dev-14186422.okta.com",
				"authorization_endpoint": "https://example.com/auth",
				"token_endpoint": "https://example.com/token",
				"jwks_uri": "https://example.com/keys",
				"id_token_signing_alg_values_supported": ["RS256"]
			}`, "ISSUER", issuer))
		if !assert.NoError(s.t, err) {
			s.t.FailNow()
		}

		return
	}

	http.NotFound(w, r)
}

func (s *authMetadataServer) Start(_ context.Context) error {
	s.lck.Lock()
	defer s.lck.Unlock()

	/***** Set up the server serving channelz service. *****/
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen on port [%v]: %w", s.port, err)
	}

	grpcS := grpc.NewServer()
	service.RegisterAuthMetadataServiceServer(grpcS, s)
	go func() {
		_ = grpcS.Serve(lis)
		//assert.NoError(s.t, err)
	}()

	s.grpcServer = grpcS
	s.netListener = lis

	s.s = httptest.NewServer(s)

	return nil
}

func (s *authMetadataServer) Close() {
	s.lck.RLock()
	defer s.lck.RUnlock()

	s.grpcServer.Stop()
	s.s.Close()
}

func newAuthMetadataServer(t testing.TB, port int, impl service.AuthMetadataServiceServer) *authMetadataServer {
	return &authMetadataServer{
		port: port,
		t:    t,
		impl: impl,
		lck:  &sync.RWMutex{},
	}
}

func Test_newAuthInterceptor(t *testing.T) {
	t.Run("Other Error", func(t *testing.T) {
		f := NewPerRPCCredentialsFuture()
		p := NewPerRPCCredentialsFuture()
		interceptor := NewAuthInterceptor(&Config{}, &mocks.TokenCache{}, f, p)
		otherError := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return status.New(codes.Canceled, "").Err()
		}

		assert.Error(t, interceptor(context.Background(), "POST", nil, nil, nil, otherError))
	})

	t.Run("Unauthenticated first time, succeed the second time", func(t *testing.T) {
		assert.NoError(t, logger.SetConfig(&logger.Config{
			Level: logger.DebugLevel,
		}))

		port := rand.IntnRange(10000, 60000)
		m := &adminMocks.AuthMetadataServiceServer{}
		m.OnGetOAuth2MetadataMatch(mock.Anything, mock.Anything).Return(&service.OAuth2MetadataResponse{
			AuthorizationEndpoint: fmt.Sprintf("http://localhost:%d/oauth2/authorize", port),
			TokenEndpoint:         fmt.Sprintf("http://localhost:%d/oauth2/token", port),
			JwksUri:               fmt.Sprintf("http://localhost:%d/oauth2/jwks", port),
		}, nil)
		m.OnGetPublicClientConfigMatch(mock.Anything, mock.Anything).Return(&service.PublicClientAuthConfigResponse{
			Scopes: []string{"all"},
		}, nil)
		s := newAuthMetadataServer(t, port, m)
		ctx := context.Background()
		assert.NoError(t, s.Start(ctx))
		defer s.Close()

		u, err := url.Parse(fmt.Sprintf("dns:///localhost:%d", port))
		assert.NoError(t, err)

		f := NewPerRPCCredentialsFuture()
		p := NewPerRPCCredentialsFuture()
		interceptor := NewAuthInterceptor(&Config{
			Endpoint:              config.URL{URL: *u},
			UseInsecureConnection: true,
			AuthType:              AuthTypeClientSecret,
		}, &mocks.TokenCache{}, f, p)
		unauthenticated := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return status.New(codes.Unauthenticated, "").Err()
		}

		err = interceptor(ctx, "POST", nil, nil, nil, unauthenticated)
		assert.Error(t, err)
		assert.Truef(t, f.IsInitialized(), "PerRPCCredentialFuture should be initialized")
		assert.False(t, f.Get().RequireTransportSecurity(), "Insecure should be true leading to RequireTLS false")
	})

	t.Run("Already authenticated", func(t *testing.T) {
		assert.NoError(t, logger.SetConfig(&logger.Config{
			Level: logger.DebugLevel,
		}))

		port := rand.IntnRange(10000, 60000)
		m := &adminMocks.AuthMetadataServiceServer{}
		s := newAuthMetadataServer(t, port, m)
		ctx := context.Background()
		assert.NoError(t, s.Start(ctx))
		defer s.Close()

		u, err := url.Parse(fmt.Sprintf("dns:///localhost:%d", port))
		assert.NoError(t, err)

		f := NewPerRPCCredentialsFuture()
		p := NewPerRPCCredentialsFuture()

		interceptor := NewAuthInterceptor(&Config{
			Endpoint:              config.URL{URL: *u},
			UseInsecureConnection: true,
			AuthType:              AuthTypeClientSecret,
		}, &mocks.TokenCache{}, f, p)
		authenticated := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return nil
		}

		err = interceptor(ctx, "POST", nil, nil, nil, authenticated)
		assert.NoError(t, err)
		assert.Falsef(t, f.IsInitialized(), "PerRPCCredentialFuture should not need to be initialized")
	})

	t.Run("Other error, doesn't authenticate", func(t *testing.T) {
		assert.NoError(t, logger.SetConfig(&logger.Config{
			Level: logger.DebugLevel,
		}))

		port := rand.IntnRange(10000, 60000)
		m := &adminMocks.AuthMetadataServiceServer{}
		m.OnGetOAuth2MetadataMatch(mock.Anything, mock.Anything).Return(&service.OAuth2MetadataResponse{
			AuthorizationEndpoint: fmt.Sprintf("http://localhost:%d/oauth2/authorize", port),
			TokenEndpoint:         fmt.Sprintf("http://localhost:%d/oauth2/token", port),
			JwksUri:               fmt.Sprintf("http://localhost:%d/oauth2/jwks", port),
		}, nil)
		m.OnGetPublicClientConfigMatch(mock.Anything, mock.Anything).Return(&service.PublicClientAuthConfigResponse{
			Scopes: []string{"all"},
		}, nil)

		s := newAuthMetadataServer(t, port, m)
		ctx := context.Background()
		assert.NoError(t, s.Start(ctx))
		defer s.Close()

		u, err := url.Parse(fmt.Sprintf("dns:///localhost:%d", port))
		assert.NoError(t, err)

		f := NewPerRPCCredentialsFuture()
		p := NewPerRPCCredentialsFuture()

		interceptor := NewAuthInterceptor(&Config{
			Endpoint:              config.URL{URL: *u},
			UseInsecureConnection: true,
			AuthType:              AuthTypeClientSecret,
		}, &mocks.TokenCache{}, f, p)
		unauthenticated := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
			return status.New(codes.Aborted, "").Err()
		}

		err = interceptor(ctx, "POST", nil, nil, nil, unauthenticated)
		assert.Error(t, err)
		assert.Falsef(t, f.IsInitialized(), "PerRPCCredentialFuture should not be initialized")
	})
}

func TestMaterializeCredentials(t *testing.T) {
	port := rand.IntnRange(10000, 60000)
	t.Run("No oauth2 metadata endpoint or Public client config lookup", func(t *testing.T) {
		m := &adminMocks.AuthMetadataServiceServer{}
		m.OnGetOAuth2MetadataMatch(mock.Anything, mock.Anything).Return(nil, errors.New("unexpected call to get oauth2 metadata"))
		m.OnGetPublicClientConfigMatch(mock.Anything, mock.Anything).Return(nil, errors.New("unexpected call to get public client config"))
		s := newAuthMetadataServer(t, port, m)
		ctx := context.Background()
		assert.NoError(t, s.Start(ctx))
		defer s.Close()

		u, err := url.Parse(fmt.Sprintf("dns:///localhost:%d", port))
		assert.NoError(t, err)

		f := NewPerRPCCredentialsFuture()
		p := NewPerRPCCredentialsFuture()

		err = MaterializeCredentials(ctx, &Config{
			Endpoint:              config.URL{URL: *u},
			UseInsecureConnection: true,
			AuthType:              AuthTypeClientSecret,
			TokenURL:              fmt.Sprintf("http://localhost:%d/api/v1/token", port),
			Scopes:                []string{"all"},
			Audience:              "http://localhost:30081",
			AuthorizationHeader:   "authorization",
		}, &mocks.TokenCache{}, f, p)
		assert.NoError(t, err)
	})
	t.Run("Failed to fetch client metadata", func(t *testing.T) {
		m := &adminMocks.AuthMetadataServiceServer{}
		m.OnGetOAuth2MetadataMatch(mock.Anything, mock.Anything).Return(nil, errors.New("unexpected call to get oauth2 metadata"))
		failedPublicClientConfigLookup := errors.New("expected err")
		m.OnGetPublicClientConfigMatch(mock.Anything, mock.Anything).Return(nil, failedPublicClientConfigLookup)
		s := newAuthMetadataServer(t, port, m)
		ctx := context.Background()
		assert.NoError(t, s.Start(ctx))
		defer s.Close()

		u, err := url.Parse(fmt.Sprintf("dns:///localhost:%d", port))
		assert.NoError(t, err)

		f := NewPerRPCCredentialsFuture()
		p := NewPerRPCCredentialsFuture()

		err = MaterializeCredentials(ctx, &Config{
			Endpoint:              config.URL{URL: *u},
			UseInsecureConnection: true,
			AuthType:              AuthTypeClientSecret,
			TokenURL:              fmt.Sprintf("http://localhost:%d/api/v1/token", port),
			Scopes:                []string{"all"},
		}, &mocks.TokenCache{}, f, p)
		assert.EqualError(t, err, "failed to fetch client metadata. Error: rpc error: code = Unknown desc = expected err")
	})
}

func TestNewProxyAuthInterceptor(t *testing.T) {
	cfg := &Config{
		ProxyCommand: []string{"echo", "test-token"},
	}

	p := NewPerRPCCredentialsFuture()

	interceptor := NewProxyAuthInterceptor(cfg, p)

	ctx := context.Background()
	method := "/test.method"
	req := "request"
	reply := "reply"
	cc := new(grpc.ClientConn)

	errorInvoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		return errors.New("test error")
	}

	// Call should return an error and trigger the interceptor to materialize proxy auth credentials
	err := interceptor(ctx, method, req, reply, cc, errorInvoker)
	assert.Error(t, err)

	// Check if proxyCredentialsFuture contains a proxy auth header token
	creds, err := p.Get().GetRequestMetadata(ctx, "")
	assert.True(t, p.IsInitialized())
	assert.NoError(t, err)
	assert.Equal(t, "Bearer test-token", creds[ProxyAuthorizationHeader])
}

type testRoundTripper struct {
	RoundTripFunc func(req *http.Request) (*http.Response, error)
}

func (t *testRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.RoundTripFunc(req)
}

func TestSetHTTPClientContext(t *testing.T) {
	ctx := context.Background()

	t.Run("no proxy command and no proxy url", func(t *testing.T) {
		cfg := &Config{}

		newCtx := setHTTPClientContext(ctx, cfg, nil)

		httpClient, ok := newCtx.Value(oauth2.HTTPClient).(*http.Client)
		assert.True(t, ok)

		transport, ok := httpClient.Transport.(*http.Transport)
		assert.True(t, ok)
		assert.Nil(t, transport.Proxy)
	})

	t.Run("proxy url", func(t *testing.T) {
		cfg := &Config{
			HTTPProxyURL: config.
				URL{URL: url.URL{
				Scheme: "http",
				Host:   "localhost:8080",
			}},
		}
		newCtx := setHTTPClientContext(ctx, cfg, nil)

		httpClient, ok := newCtx.Value(oauth2.HTTPClient).(*http.Client)
		assert.True(t, ok)

		transport, ok := httpClient.Transport.(*http.Transport)
		assert.True(t, ok)
		assert.NotNil(t, transport.Proxy)
	})

	t.Run("proxy command adds proxy-authorization header", func(t *testing.T) {
		cfg := &Config{
			ProxyCommand: []string{"echo", "test-token-http-client"},
		}

		p := NewPerRPCCredentialsFuture()
		err := MaterializeProxyAuthCredentials(ctx, cfg, p)
		assert.NoError(t, err)

		newCtx := setHTTPClientContext(ctx, cfg, p)

		httpClient, ok := newCtx.Value(oauth2.HTTPClient).(*http.Client)
		assert.True(t, ok)

		pat, ok := httpClient.Transport.(*proxyAuthTransport)
		assert.True(t, ok)

		testRoundTripper := &testRoundTripper{
			RoundTripFunc: func(req *http.Request) (*http.Response, error) {
				// Check if the ProxyAuthorizationHeader is correctly set
				assert.Equal(t, "Bearer test-token-http-client", req.Header.Get(ProxyAuthorizationHeader))
				return &http.Response{StatusCode: http.StatusOK}, nil
			},
		}
		pat.transport = testRoundTripper

		req, _ := http.NewRequest("GET", "http://example.com", nil)
		_, err = httpClient.Do(req)
		assert.NoError(t, err)
	})
}
