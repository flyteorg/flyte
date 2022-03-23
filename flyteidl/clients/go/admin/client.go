package admin

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"

	"github.com/flyteorg/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyteidl/clients/go/admin/pkce"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flytestdlib/logger"

	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcRetry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	grpcPrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// IDE "Go Generate File". This will create a mocks/AdminServiceClient.go file
//go:generate mockery -dir ../../../gen/pb-go/flyteidl/service -name AdminServiceClient -output ../admin/mocks

// Clientset contains the clients exposed to communicate with various admin services.
type Clientset struct {
	adminServiceClient        service.AdminServiceClient
	authMetadataServiceClient service.AuthMetadataServiceClient
	healthServiceClient       grpc_health_v1.HealthClient
	identityServiceClient     service.IdentityServiceClient
	dataProxyServiceClient    service.DataProxyClient
	authOpt                   grpc.DialOption
}

func (c Clientset) AuthOpt() grpc.DialOption {
	return c.authOpt
}

// AdminClient retrieves the AdminServiceClient
func (c Clientset) AdminClient() service.AdminServiceClient {
	return c.adminServiceClient
}

// AuthMetadataClient retrieves the AuthMetadataServiceClient
func (c Clientset) AuthMetadataClient() service.AuthMetadataServiceClient {
	return c.authMetadataServiceClient
}

// HealthServiceClient retrieves the grpc_health_v1.HealthClient
func (c Clientset) HealthServiceClient() grpc_health_v1.HealthClient {
	return c.healthServiceClient
}

func (c Clientset) IdentityClient() service.IdentityServiceClient {
	return c.identityServiceClient
}

func (c Clientset) DataProxyClient() service.DataProxyClient {
	return c.dataProxyServiceClient
}

func NewAdminClient(ctx context.Context, conn *grpc.ClientConn) service.AdminServiceClient {
	logger.Infof(ctx, "Initialized Admin client")
	return service.NewAdminServiceClient(conn)
}

func GetAdditionalAdminClientConfigOptions(cfg *Config) []grpc.DialOption {
	opts := make([]grpc.DialOption, 0, 2)
	backoffConfig := grpc.BackoffConfig{
		MaxDelay: cfg.MaxBackoffDelay.Duration,
	}

	opts = append(opts, grpc.WithBackoffConfig(backoffConfig))

	timeoutDialOption := grpcRetry.WithPerRetryTimeout(cfg.PerRetryTimeout.Duration)
	maxRetriesOption := grpcRetry.WithMax(uint(cfg.MaxRetries))

	retryInterceptor := grpcRetry.UnaryClientInterceptor(timeoutDialOption, maxRetriesOption)
	finalUnaryInterceptor := grpcMiddleware.ChainUnaryClient(
		grpcPrometheus.UnaryClientInterceptor,
		retryInterceptor,
	)

	// We only make unary calls in this client, no streaming calls.  We can add a streaming interceptor if admin
	// ever has those endpoints
	opts = append(opts, grpc.WithUnaryInterceptor(finalUnaryInterceptor))

	return opts
}

// This retrieves a DialOption that contains a source for generating JWTs for authentication with Flyte Admin. If
// the token endpoint is set in the config, that will be used, otherwise it'll attempt to make a metadata call.
func getAuthenticationDialOption(ctx context.Context, cfg *Config, tokenSourceProvider TokenSourceProvider,
	authClient service.AuthMetadataServiceClient) (grpc.DialOption, error) {
	if tokenSourceProvider == nil {
		return nil, errors.New("can't create authenticated channel without a TokenSourceProvider")
	}

	clientMetadata, err := authClient.GetPublicClientConfig(ctx, &service.PublicClientAuthConfigRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch client metadata. Error: %v", err)
	}

	tSource, err := tokenSourceProvider.GetTokenSource(ctx)
	if err != nil {
		return nil, err
	}

	oauthTokenSource := NewCustomHeaderTokenSource(tSource, cfg.UseInsecureConnection, clientMetadata.AuthorizationMetadataKey)
	return grpc.WithPerRPCCredentials(oauthTokenSource), nil
}

// InitializeAuthMetadataClient creates a new anonymously Auth Metadata Service client.
func InitializeAuthMetadataClient(ctx context.Context, cfg *Config) (client service.AuthMetadataServiceClient, err error) {
	// Create an unauthenticated connection to fetch AuthMetadata
	authMetadataConnection, err := NewAdminConnection(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize admin connection. Error: %w", err)
	}

	return service.NewAuthMetadataServiceClient(authMetadataConnection), nil
}

func NewAdminConnection(ctx context.Context, cfg *Config, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	if opts == nil {
		// Initialize opts list to the potential number of options we will add. Initialization optimizes memory
		// allocation.
		opts = make([]grpc.DialOption, 0, 5)
	}

	if cfg.UseInsecureConnection {
		opts = append(opts, grpc.WithInsecure())
	} else {
		var creds credentials.TransportCredentials
		var caCerts *x509.CertPool
		var err error
		tlsConfig := &tls.Config{} //nolint
		// Use the cacerts passed in from the config parameter
		if len(cfg.CACertFilePath) > 0 {
			caCerts, err = readCACerts(cfg.CACertFilePath)
			if err != nil {
				return nil, err
			}
		}
		if cfg.InsecureSkipVerify {
			logger.Warnf(ctx, "using insecureSkipVerify. Server's certificate chain and host name wont be verified. Caution : shouldn't be used for production usecases")
			tlsConfig.InsecureSkipVerify = true
			creds = credentials.NewTLS(tlsConfig)
		} else {
			creds = credentials.NewClientTLSFromCert(caCerts, "")
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}

	opts = append(opts, GetAdditionalAdminClientConfigOptions(cfg)...)

	return grpc.Dial(cfg.Endpoint.String(), opts...)
}

// InitializeAdminClient creates an AdminClient with a shared Admin connection for the process
// Deprecated: Please use initializeClients instead.
func InitializeAdminClient(ctx context.Context, cfg *Config, opts ...grpc.DialOption) service.AdminServiceClient {
	set, err := initializeClients(ctx, cfg, nil, opts...)
	if err != nil {
		logger.Panicf(ctx, "Failed to initialize client. Error: %v", err)
		return nil
	}

	return set.AdminClient()
}

// initializeClients creates an AdminClient, AuthServiceClient and IdentityServiceClient with a shared Admin connection
// for the process. Note that if called with different cfg/dialoptions, it will not refresh the connection.
func initializeClients(ctx context.Context, cfg *Config, tokenCache pkce.TokenCache, opts ...grpc.DialOption) (*Clientset, error) {
	authMetadataClient, err := InitializeAuthMetadataClient(ctx, cfg)
	if err != nil {
		logger.Panicf(ctx, "failed to initialize Auth Metadata Client. Error: %v", err)
	}

	tokenSourceProvider, err := NewTokenSourceProvider(ctx, cfg, tokenCache, authMetadataClient)
	if err != nil {
		logger.Errorf(ctx, "failed to initialize token source provider. Err: %s", err.Error())
	}

	authOpt, err := getAuthenticationDialOption(ctx, cfg, tokenSourceProvider, authMetadataClient)
	if err != nil {
		logger.Warnf(ctx, "Starting an unauthenticated client because: %v", err)
	}

	if authOpt != nil {
		opts = append(opts, authOpt)
	}

	adminConnection, err := NewAdminConnection(ctx, cfg, opts...)
	if err != nil {
		logger.Panicf(ctx, "failed to initialize Admin connection. Err: %s", err.Error())
	}

	var cs Clientset
	cs.adminServiceClient = NewAdminClient(ctx, adminConnection)
	cs.authMetadataServiceClient = service.NewAuthMetadataServiceClient(adminConnection)
	cs.identityServiceClient = service.NewIdentityServiceClient(adminConnection)
	cs.healthServiceClient = grpc_health_v1.NewHealthClient(adminConnection)
	cs.dataProxyServiceClient = service.NewDataProxyClient(adminConnection)
	if authOpt != nil {
		cs.authOpt = authOpt
	}

	return &cs, nil
}

// Deprecated: Please use NewClientsetBuilder() instead.
func InitializeAdminClientFromConfig(ctx context.Context, tokenCache pkce.TokenCache, opts ...grpc.DialOption) (service.AdminServiceClient, error) {
	clientSet, err := initializeClients(ctx, GetConfig(ctx), tokenCache, opts...)
	if err != nil {
		return nil, err
	}

	return clientSet.AdminClient(), nil
}

func InitializeMockAdminClient() service.AdminServiceClient {
	logger.Infof(context.TODO(), "Initialized Mock Admin client")
	return &mocks.AdminServiceClient{}
}

func InitializeMockClientset() *Clientset {
	logger.Infof(context.TODO(), "Initialized Mock Clientset")
	return &Clientset{
		adminServiceClient:        &mocks.AdminServiceClient{},
		authMetadataServiceClient: &mocks.AuthMetadataServiceClient{},
		identityServiceClient:     &mocks.IdentityServiceClient{},
	}
}
