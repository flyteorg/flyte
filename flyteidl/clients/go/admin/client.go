package admin

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/coreos/go-oidc"
	"github.com/flyteorg/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flytestdlib/logger"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	once            = sync.Once{}
	adminConnection *grpc.ClientConn
)

func NewAdminClient(ctx context.Context, conn *grpc.ClientConn) service.AdminServiceClient {
	logger.Infof(ctx, "Initialized Admin client")
	return service.NewAdminServiceClient(conn)
}

func GetAdditionalAdminClientConfigOptions(cfg Config) []grpc.DialOption {
	opts := make([]grpc.DialOption, 0, 2)
	backoffConfig := grpc.BackoffConfig{
		MaxDelay: cfg.MaxBackoffDelay.Duration,
	}
	opts = append(opts, grpc.WithBackoffConfig(backoffConfig))

	timeoutDialOption := grpc_retry.WithPerRetryTimeout(cfg.PerRetryTimeout.Duration)
	maxRetriesOption := grpc_retry.WithMax(uint(cfg.MaxRetries))

	retryInterceptor := grpc_retry.UnaryClientInterceptor(timeoutDialOption, maxRetriesOption)
	finalUnaryInterceptor := grpc_middleware.ChainUnaryClient(
		grpc_prometheus.UnaryClientInterceptor,
		retryInterceptor,
	)

	// We only make unary calls in this client, no streaming calls.  We can add a streaming interceptor if admin
	// ever has those endpoints
	opts = append(opts, grpc.WithUnaryInterceptor(finalUnaryInterceptor))

	return opts
}

// This function assumes that the authorization server supports the OAuth metadata standard, and uses the oidc
// library to retrieve the token endpoint.
func getTokenEndpointFromAuthServer(ctx context.Context, authorizationServer string) (string, error) {
	if authorizationServer == "" {
		logger.Errorf(ctx, "Attempting to construct provider with empty authorizationServer")
		return "", errors.New("cannot get token URL from empty authorizationServer")
	}

	oidcCtx := oidc.ClientContext(ctx, &http.Client{})
	provider, err := oidc.NewProvider(oidcCtx, authorizationServer)
	if err != nil {
		logger.Errorf(ctx, "Error when constructing new OIDC Provider")
		return "", err
	}
	logger.Infof(ctx, "Constructing Admin client with token endpoint %s", provider.Endpoint().TokenURL)

	return provider.Endpoint().TokenURL, nil
}

// This retrieves a DialOption that contains a source for generating JWTs for authentication with Flyte Admin.
// It will first attempt to retrieve the token endpoint by making a metadata call. If that fails, but the token endpoint
// is set in the config, that will be used instead.
func getAuthenticationDialOption(ctx context.Context, cfg Config) (grpc.DialOption, error) {
	var tokenURL string
	tokenURL, err := getTokenEndpointFromAuthServer(ctx, cfg.AuthorizationServerURL)
	if err != nil || tokenURL == "" {
		logger.Infof(ctx, "No token URL found from configuration Issuer, looking for token endpoint directly")
		if err != nil {
			logger.Errorf(ctx, "Err is %s", err)
		}
		tokenURL = cfg.TokenURL
		if tokenURL == "" {
			return nil, errors.New("no token endpoint could be found")
		}
	}

	secretBytes, err := ioutil.ReadFile(cfg.ClientSecretLocation)
	if err != nil {
		logger.Errorf(ctx, "Error reading secret from location %s", cfg.ClientSecretLocation)
		return nil, err
	}
	secret := strings.TrimSpace(string(secretBytes))

	ccConfig := clientcredentials.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: secret,
		TokenURL:     tokenURL,
		Scopes:       cfg.Scopes,
	}
	tSource := ccConfig.TokenSource(ctx)
	oauthTokenSource := NewCustomHeaderTokenSource(tSource, cfg.AuthorizationHeader)
	return grpc.WithPerRPCCredentials(oauthTokenSource), nil
}

func NewAdminConnection(ctx context.Context, cfg Config) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption

	if cfg.UseInsecureConnection {
		opts = append(opts, grpc.WithInsecure())
	} else {
		// TODO: as of Go 1.11.4, this is not supported on Windows. https://github.com/golang/go/issues/16736
		creds := credentials.NewClientTLSFromCert(nil, "")
		opts = append(opts, grpc.WithTransportCredentials(creds))
		if cfg.UseAuth {
			logger.Infof(ctx, "Instantiating a token source to authenticate against Admin, ID: %s", cfg.ClientID)
			jwtDialOption, err := getAuthenticationDialOption(ctx, cfg)
			if err != nil {
				return nil, err
			}
			opts = append(opts, jwtDialOption)
		}
	}

	opts = append(opts, GetAdditionalAdminClientConfigOptions(cfg)...)
	return grpc.Dial(cfg.Endpoint.String(), opts...)
}

// Create an AdminClient with a shared Admin connection for the process
func InitializeAdminClient(ctx context.Context, cfg Config) service.AdminServiceClient {
	once.Do(func() {
		var err error
		adminConnection, err = NewAdminConnection(ctx, cfg)
		if err != nil {
			logger.Panicf(ctx, "failed to initialize Admin connection. Err: %s", err.Error())
		}
	})

	return NewAdminClient(ctx, adminConnection)
}

func InitializeAdminClientFromConfig(ctx context.Context) (service.AdminServiceClient, error) {
	cfg := GetConfig(ctx)
	if cfg == nil {
		return nil, fmt.Errorf("retrieved Nil config for [%s] key", configSectionKey)
	}
	return InitializeAdminClient(ctx, *cfg), nil
}

func InitializeMockAdminClient() service.AdminServiceClient {
	logger.Infof(context.TODO(), "Initialized Mock Admin client")
	return &mocks.AdminServiceClient{}
}
