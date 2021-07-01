package admin

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

	"github.com/flyteorg/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyteidl/clients/go/admin/pkce"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flytestdlib/logger"

	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcRetry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	grpcPrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	once            = sync.Once{}
	adminConnection *grpc.ClientConn

	// A new connection just for auth metadata service since it will be used to retrieve auth
	// related information that's needed to initialize the Clientset.
	onceAuthMetadata       = sync.Once{}
	authMetadataConnection *grpc.ClientConn
)

// Clientset contains the clients exposed to communicate with various admin services.
type Clientset struct {
	adminServiceClient        service.AdminServiceClient
	authMetadataServiceClient service.AuthMetadataServiceClient
	identityServiceClient     service.IdentityServiceClient
}

// AdminClient retrieves the AdminServiceClient
func (c Clientset) AdminClient() service.AdminServiceClient {
	return c.adminServiceClient
}

// AuthMetadataClient retrieves the AuthMetadataServiceClient
func (c Clientset) AuthMetadataClient() service.AuthMetadataServiceClient {
	return c.authMetadataServiceClient
}

func (c Clientset) IdentityClient() service.IdentityServiceClient {
	return c.identityServiceClient
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
func getAuthenticationDialOption(ctx context.Context, cfg *Config, tokenCache pkce.TokenCache,
	authClient service.AuthMetadataServiceClient) (grpc.DialOption, error) {

	tokenURL := cfg.TokenURL
	if len(tokenURL) == 0 {
		metadata, err := authClient.GetOAuth2Metadata(ctx, &service.OAuth2MetadataRequest{})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch auth metadata. Error: %v", err)
		}

		tokenURL = metadata.TokenEndpoint
	}

	clientMetadata, err := authClient.GetPublicClientConfig(ctx, &service.PublicClientAuthConfigRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch client metadata. Error: %v", err)
	}

	var tSource oauth2.TokenSource
	if cfg.AuthType == AuthTypeClientSecret {
		tSource, err = getClientCredentialsTokenSource(ctx, cfg, clientMetadata, tokenURL)
		if err != nil {
			return nil, err
		}
	} else if cfg.AuthType == AuthTypePkce {
		tokenOrchestrator, err := pkce.NewTokenOrchestrator(ctx, cfg.PkceConfig, tokenCache, authClient)
		if err != nil {
			return nil, err
		}

		tSource, err = getPkceAuthTokenSource(ctx, tokenOrchestrator)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("unsupported type %v", cfg.AuthType)
	}

	oauthTokenSource := NewCustomHeaderTokenSource(tSource, cfg.UseInsecureConnection, clientMetadata.AuthorizationMetadataKey)
	return grpc.WithPerRPCCredentials(oauthTokenSource), nil
}

// Returns the client credentials token source to be used eg by flytepropeller to communicate with admin/ by CI
func getClientCredentialsTokenSource(ctx context.Context, cfg *Config,
	clientMetadata *service.PublicClientAuthConfigResponse, tokenURL string) (oauth2.TokenSource, error) {

	secretBytes, err := ioutil.ReadFile(cfg.ClientSecretLocation)
	if err != nil {
		logger.Errorf(ctx, "Error reading secret from location %s", cfg.ClientSecretLocation)
		return nil, err
	}

	secret := strings.TrimSpace(string(secretBytes))
	scopes := cfg.Scopes
	if len(scopes) == 0 {
		scopes = clientMetadata.Scopes
	}

	ccConfig := clientcredentials.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: secret,
		TokenURL:     tokenURL,
		Scopes:       scopes,
	}

	return ccConfig.TokenSource(ctx), nil
}

// Returns the token source which would be used for three legged oauth. eg : for admin to authorize access to flytectl
func getPkceAuthTokenSource(ctx context.Context, tokenOrchestrator pkce.TokenOrchestrator) (oauth2.TokenSource, error) {
	// explicitly ignore error while fetching token from cache.
	authToken, err := tokenOrchestrator.FetchTokenFromCacheOrRefreshIt(ctx)
	if err != nil {
		logger.Warnf(ctx, "Failed fetching from cache. Will restart the flow. Error: %v", err)
	}

	if authToken == nil {
		// Fetch using auth flow
		if authToken, err = tokenOrchestrator.FetchTokenFromAuthFlow(ctx); err != nil {
			logger.Errorf(ctx, "Error fetching token using auth flow due to %v", err)
			return nil, err
		}
	}

	return &pkce.SimpleTokenSource{
		CachedToken: authToken,
	}, nil
}

// InitializeAuthMetadataClient creates a new anonymously Auth Metadata Service client.
func InitializeAuthMetadataClient(ctx context.Context, cfg *Config) (client service.AuthMetadataServiceClient, err error) {
	onceAuthMetadata.Do(func() {
		authMetadataConnection, err = NewAdminConnection(ctx, cfg)
	})

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
		// TODO: as of Go 1.11.4, this is not supported on Windows. https://github.com/golang/go/issues/16736
		var creds credentials.TransportCredentials
		if cfg.InsecureSkipVerify {
			logger.Warnf(ctx, "using insecureSkipVerify. Server's certificate chain and host name wont be verified. Caution : shouldn't be used for production usecases")
			tlsConfig := &tls.Config{
				InsecureSkipVerify: true, //nolint

			}
			creds = credentials.NewTLS(tlsConfig)
		} else {
			creds = credentials.NewClientTLSFromCert(nil, "")
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}

	opts = append(opts, GetAdditionalAdminClientConfigOptions(cfg)...)

	return grpc.Dial(cfg.Endpoint.String(), opts...)
}

// Create an AdminClient with a shared Admin connection for the process
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
	once.Do(func() {
		authMetadataClient, err := InitializeAuthMetadataClient(ctx, cfg)
		if err != nil {
			logger.Panicf(ctx, "failed to initialize Auth Metadata Client. Error: %v", err)
		}

		// If auth is enabled, this call will return the required information to use to authenticate, otherwise,
		// start the client without authentication.
		opt, err := getAuthenticationDialOption(ctx, cfg, tokenCache, authMetadataClient)
		if err != nil {
			logger.Warnf(ctx, "Starting an unauthenticated client because: %v", err)
		}

		if opt != nil {
			opts = append(opts, opt)
		}

		adminConnection, err = NewAdminConnection(ctx, cfg, opts...)
		if err != nil {
			logger.Panicf(ctx, "failed to initialize Admin connection. Err: %s", err.Error())
		}
	})

	var cs Clientset
	cs.adminServiceClient = NewAdminClient(ctx, adminConnection)
	cs.authMetadataServiceClient = service.NewAuthMetadataServiceClient(adminConnection)
	cs.identityServiceClient = service.NewIdentityServiceClient(adminConnection)
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
