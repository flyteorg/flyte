package admin

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"reflect"

	"github.com/golang/protobuf/proto"

	grpcRetry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/cache"
	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/artifacts"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flyte/flytestdlib/grpcutils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

// IDE "Go Generate File". This will create a mocks/AdminServiceClient.go file
//go:generate mockery -dir ../../../gen/pb-go/flyteidl/service -name AdminServiceClient -output ../admin/mocks

// Clientset contains the clients exposed to communicate with various admin services.
type Clientset struct {
	adminServiceClient        service.AdminServiceClient
	watchServiceClient        service.WatchServiceClient
	authMetadataServiceClient service.AuthMetadataServiceClient
	healthServiceClient       grpc_health_v1.HealthClient
	identityServiceClient     service.IdentityServiceClient
	dataProxyServiceClient    service.DataProxyServiceClient
	signalServiceClient       service.SignalServiceClient
	artifactServiceClient     artifacts.ArtifactRegistryClient
}

// AdminClient retrieves the AdminServiceClient
func (c Clientset) AdminClient() service.AdminServiceClient {
	return c.adminServiceClient
}

// WatchServiceClient retrieves the WatchServiceClient
func (c Clientset) WatchServiceClient() service.WatchServiceClient {
	return c.watchServiceClient
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

func (c Clientset) DataProxyClient() service.DataProxyServiceClient {
	return c.dataProxyServiceClient
}

func (c Clientset) SignalServiceClient() service.SignalServiceClient {
	return c.signalServiceClient
}

func (c Clientset) ArtifactServiceClient() artifacts.ArtifactRegistryClient {
	return c.artifactServiceClient
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

	// We only make unary calls in this client, no streaming calls.  We can add a streaming interceptor if admin
	// ever has those endpoints
	opts = append(opts, grpc.WithChainUnaryInterceptor(grpcutils.GrpcClientMetrics().UnaryClientInterceptor(), retryInterceptor))

	if cfg.MaxMessageSizeBytes > 0 {
		opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(cfg.MaxMessageSizeBytes)))
	}

	return opts
}

// This retrieves a DialOption that contains a source for generating JWTs for authentication with Flyte Admin. If
// the token endpoint is set in the config, that will be used, otherwise it'll attempt to make a metadata call.
func getAuthenticationDialOption(ctx context.Context, cfg *Config, tokenSourceProvider TokenSourceProvider,
	authClient service.AuthMetadataServiceClient) (grpc.DialOption, error) {
	if tokenSourceProvider == nil {
		return nil, errors.New("can't create authenticated channel without a TokenSourceProvider")
	}

	authorizationMetadataKey := cfg.AuthorizationHeader
	if len(authorizationMetadataKey) == 0 {
		clientMetadata, err := authClient.GetPublicClientConfig(ctx, &service.PublicClientAuthConfigRequest{})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch client metadata. Error: %v", err)
		}
		authorizationMetadataKey = clientMetadata.AuthorizationMetadataKey
	}

	tokenSource, err := tokenSourceProvider.GetTokenSource(ctx)
	if err != nil {
		return nil, err
	}

	wrappedTokenSource := NewCustomHeaderTokenSource(tokenSource, cfg.UseInsecureConnection, authorizationMetadataKey)
	return grpc.WithPerRPCCredentials(wrappedTokenSource), nil
}

// InitializeAuthMetadataClient creates a new anonymously Auth Metadata Service client.
func InitializeAuthMetadataClient(ctx context.Context, cfg *Config, proxyCredentialsFuture *PerRPCCredentialsFuture) (client service.AuthMetadataServiceClient, err error) {
	// Create an unauthenticated connection to fetch AuthMetadata
	authMetadataConnection, err := NewAdminConnection(ctx, cfg, proxyCredentialsFuture)
	if err != nil {
		return nil, fmt.Errorf("failed to initialized admin connection. Error: %w", err)
	}

	return service.NewAuthMetadataServiceClient(authMetadataConnection), nil
}

func NewAdminConnection(ctx context.Context, cfg *Config, proxyCredentialsFuture *PerRPCCredentialsFuture, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	if opts == nil {
		// Initialize opts list to the potential number of options we will add. Initialization optimizes memory
		// allocation.
		opts = make([]grpc.DialOption, 0, 7)
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
			logger.Warnf(ctx, "using insecureSkipVerify. Server's certificate chain and host name won't be verified. Caution : shouldn't be used for production usecases")
			tlsConfig.InsecureSkipVerify = true
			creds = credentials.NewTLS(tlsConfig)
		} else {
			creds = credentials.NewClientTLSFromCert(caCerts, "")
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}

	opts = append(opts, GetAdditionalAdminClientConfigOptions(cfg)...)

	if cfg.ProxyCommand != nil && len(cfg.ProxyCommand) > 0 {
		opts = append(opts, grpc.WithChainUnaryInterceptor(NewProxyAuthInterceptor(cfg, proxyCredentialsFuture)))
		opts = append(opts, grpc.WithPerRPCCredentials(proxyCredentialsFuture))
	}

	return grpc.Dial(cfg.Endpoint.String(), opts...)
}

// InitializeAdminClient creates an AdminClient with a shared Admin connection for the process
// Deprecated: Please use initializeClients instead.
func InitializeAdminClient(ctx context.Context, cfg *Config, opts ...grpc.DialOption) service.AdminServiceClient {
	set, err := initializeClients(ctx, cfg, nil, opts...)
	if err != nil {
		logger.Panicf(ctx, "Failed to initialized client. Error: %v", err)
		return nil
	}

	return set.AdminClient()
}

// replaceOrgFieldIfEmpty recursively replaces "Org" field values in proto messages if they are empty
func replaceOrgFieldIfEmpty(msg proto.Message, org string) {
	visited := make(map[uintptr]bool)
	replaceOrgIfEmptyHelper(reflect.ValueOf(msg).Elem(), visited, org)
}

func replaceOrgIfEmptyHelper(val reflect.Value, visited map[uintptr]bool, org string) {
	if !val.IsValid() {
		return
	}
	// Only mark the actual struct instances as visited
	if val.Kind() == reflect.Ptr && !val.IsNil() {
		addr := val.Pointer()
		if visited[addr] {
			return
		}
		visited[addr] = true
	}

	switch val.Kind() {
	case reflect.Ptr:
		if !val.IsNil() {
			replaceOrgIfEmptyHelper(val.Elem(), visited, org)
		}
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)
			fieldType := val.Type().Field(i)
			if fieldType.Name == "Org" && field.Kind() == reflect.String && field.String() == "" {
				field.SetString(org)
			} else if fieldType.Tag.Get("protobuf_oneof") != "" {
				// Handle oneof fields
				if !field.IsNil() {
					oneofField := field.Elem()
					replaceOrgIfEmptyHelper(oneofField, visited, org)
				}
			} else {
				replaceOrgIfEmptyHelper(field, visited, org)
			}
		}
	case reflect.Slice:
		for i := 0; i < val.Len(); i++ {
			replaceOrgIfEmptyHelper(val.Index(i), visited, org)
		}
	case reflect.Interface:
		if !val.IsNil() {
			replaceOrgIfEmptyHelper(val.Elem(), visited, org)
		}
	}
}

// addOrgUnaryClientInterceptor intercepts unary RPC calls and checks if org field is empty then adds the
func addOrgUnaryClientInterceptor(org string) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if protoMsg, ok := req.(proto.Message); ok {
			replaceOrgFieldIfEmpty(protoMsg, org)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// initializeClients creates an AdminClient, AuthServiceClient and IdentityServiceClient with a shared Admin connection
// for the process. Note that if called with different cfg/dialoptions, it will not refresh the connection.
func initializeClients(ctx context.Context, cfg *Config, tokenCache cache.TokenCache, opts ...grpc.DialOption) (*Clientset, error) {
	credentialsFuture := NewPerRPCCredentialsFuture()
	proxyCredentialsFuture := NewPerRPCCredentialsFuture()

	authInterceptor := NewAuthInterceptor(cfg, tokenCache, credentialsFuture, proxyCredentialsFuture)
	opts = append(opts,
		grpc.WithChainUnaryInterceptor(authInterceptor),
		grpc.WithPerRPCCredentials(credentialsFuture))

	if len(cfg.DefaultOrg) > 0 {
		opts = append(opts, grpc.WithChainUnaryInterceptor(addOrgUnaryClientInterceptor(cfg.DefaultOrg)))
	}
	if cfg.DefaultServiceConfig != "" {
		opts = append(opts, grpc.WithDefaultServiceConfig(cfg.DefaultServiceConfig))
	}

	adminConnection, err := NewAdminConnection(ctx, cfg, proxyCredentialsFuture, opts...)
	if err != nil {
		logger.Panicf(ctx, "failed to initialized Admin connection. Err: %s", err.Error())
	}

	var cs Clientset
	cs.adminServiceClient = NewAdminClient(ctx, adminConnection)
	cs.authMetadataServiceClient = service.NewAuthMetadataServiceClient(adminConnection)
	cs.identityServiceClient = service.NewIdentityServiceClient(adminConnection)
	cs.healthServiceClient = grpc_health_v1.NewHealthClient(adminConnection)
	cs.dataProxyServiceClient = service.NewDataProxyServiceClient(adminConnection)
	cs.signalServiceClient = service.NewSignalServiceClient(adminConnection)
	cs.artifactServiceClient = artifacts.NewArtifactRegistryClient(adminConnection)
	cs.watchServiceClient = service.NewWatchServiceClient(adminConnection)

	return &cs, nil
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
		dataProxyServiceClient:    &mocks.DataProxyServiceClient{},
		healthServiceClient:       grpc_health_v1.NewHealthClient(nil),
	}
}
