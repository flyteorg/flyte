package admin

import (
	"context"
	"fmt"
	"net/http"

	"github.com/flyteorg/flyteidl/clients/go/admin/cache"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flytestdlib/logger"
	"golang.org/x/oauth2"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
)

// MaterializeCredentials will attempt to build a TokenSource given the anonymously available information exposed by the server.
// Once established, it'll invoke PerRPCCredentialsFuture.Store() on perRPCCredentials to populate it with the appropriate values.
func MaterializeCredentials(ctx context.Context, cfg *Config, tokenCache cache.TokenCache, perRPCCredentials *PerRPCCredentialsFuture) error {
	authMetadataClient, err := InitializeAuthMetadataClient(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to initialized Auth Metadata Client. Error: %w", err)
	}

	tokenSourceProvider, err := NewTokenSourceProvider(ctx, cfg, tokenCache, authMetadataClient)
	if err != nil {
		return fmt.Errorf("failed to initialized token source provider. Err: %w", err)
	}

	authorizationMetadataKey := cfg.AuthorizationHeader
	if len(authorizationMetadataKey) == 0 {
		clientMetadata, err := authMetadataClient.GetPublicClientConfig(ctx, &service.PublicClientAuthConfigRequest{})
		if err != nil {
			return fmt.Errorf("failed to fetch client metadata. Error: %v", err)
		}
		authorizationMetadataKey = clientMetadata.AuthorizationMetadataKey
	}

	tokenSource, err := tokenSourceProvider.GetTokenSource(ctx)
	if err != nil {
		return err
	}

	wrappedTokenSource := NewCustomHeaderTokenSource(tokenSource, cfg.UseInsecureConnection, authorizationMetadataKey)
	perRPCCredentials.Store(wrappedTokenSource)
	return nil
}

func shouldAttemptToAuthenticate(errorCode codes.Code) bool {
	return errorCode == codes.Unauthenticated
}

// Set up http client used in oauth2
func setHTTPClientContext(ctx context.Context, cfg *Config) context.Context {
	httpClient := &http.Client{}

	if len(cfg.HTTPProxyURL.String()) > 0 {
		// create a transport that uses the proxy
		transport := &http.Transport{
			Proxy: http.ProxyURL(&cfg.HTTPProxyURL.URL),
		}
		httpClient.Transport = transport
	}

	return context.WithValue(ctx, oauth2.HTTPClient, httpClient)
}

// NewAuthInterceptor creates a new grpc.UnaryClientInterceptor that forwards the grpc call and inspects the error.
// It will first invoke the grpc pipeline (to proceed with the request) with no modifications. It's expected for the grpc
// pipeline to already have a grpc.WithPerRPCCredentials() DialOption. If the perRPCCredentials has already been initialized,
// it'll take care of refreshing when tokens expire... etc.
// If the first invocation succeeds (either due to grpc.PerRPCCredentials setting the right tokens or the server not
// requiring authentication), the interceptor will be no-op.
// If the first invocation fails with an auth error, this interceptor will then attempt to establish a token source once
// more. It'll fail hard if it couldn't do so (i.e. it will no longer attempt to send an unauthenticated request). Once
// a token source has been created, it'll invoke the grpc pipeline again, this time the grpc.PerRPCCredentials should
// be able to find and acquire a valid AccessToken to annotate the request with.
func NewAuthInterceptor(cfg *Config, tokenCache cache.TokenCache, credentialsFuture *PerRPCCredentialsFuture) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = setHTTPClientContext(ctx, cfg)

		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			logger.Debugf(ctx, "Request failed due to [%v]. If it's an unauthenticated error, we will attempt to establish an authenticated context.", err)

			if st, ok := status.FromError(err); ok {
				// If the error we receive from executing the request expects
				if shouldAttemptToAuthenticate(st.Code()) {
					logger.Debugf(ctx, "Request failed due to [%v]. Attempting to establish an authenticated connection and trying again.", st.Code())
					newErr := MaterializeCredentials(ctx, cfg, tokenCache, credentialsFuture)
					if newErr != nil {
						return fmt.Errorf("authentication error! Original Error: %v, Auth Error: %w", err, newErr)
					}

					return invoker(ctx, method, req, reply, cc, opts...)
				}
			}
		}

		return err
	}
}
