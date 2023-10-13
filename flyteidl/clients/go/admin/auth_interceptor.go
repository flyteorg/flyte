package admin

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/cache"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"golang.org/x/oauth2"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
)

const ProxyAuthorizationHeader = "proxy-authorization"

// MaterializeCredentials will attempt to build a TokenSource given the anonymously available information exposed by the server.
// Once established, it'll invoke PerRPCCredentialsFuture.Store() on perRPCCredentials to populate it with the appropriate values.
func MaterializeCredentials(ctx context.Context, cfg *Config, tokenCache cache.TokenCache, perRPCCredentials *PerRPCCredentialsFuture, proxyCredentialsFuture *PerRPCCredentialsFuture) error {
	authMetadataClient, err := InitializeAuthMetadataClient(ctx, cfg, proxyCredentialsFuture)
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

func GetProxyTokenSource(ctx context.Context, cfg *Config) (oauth2.TokenSource, error) {
	tokenSourceProvider, err := NewExternalTokenSourceProvider(cfg.ProxyCommand)
	if err != nil {
		return nil, fmt.Errorf("failed to initialized proxy authorization token source provider. Err: %w", err)
	}
	proxyTokenSource, err := tokenSourceProvider.GetTokenSource(ctx)
	if err != nil {
		return nil, err
	}
	return proxyTokenSource, nil
}

func MaterializeProxyAuthCredentials(ctx context.Context, cfg *Config, proxyCredentialsFuture *PerRPCCredentialsFuture) error {
	proxyTokenSource, err := GetProxyTokenSource(ctx, cfg)
	if err != nil {
		return err
	}

	wrappedTokenSource := NewCustomHeaderTokenSource(proxyTokenSource, cfg.UseInsecureConnection, ProxyAuthorizationHeader)
	proxyCredentialsFuture.Store(wrappedTokenSource)

	return nil
}

func shouldAttemptToAuthenticate(errorCode codes.Code) bool {
	return errorCode == codes.Unauthenticated
}

type proxyAuthTransport struct {
	transport              http.RoundTripper
	proxyCredentialsFuture *PerRPCCredentialsFuture
}

func (c *proxyAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// check if the proxy credentials future is initialized
	if !c.proxyCredentialsFuture.IsInitialized() {
		return nil, errors.New("proxy credentials future is not initialized")
	}

	metadata, err := c.proxyCredentialsFuture.GetRequestMetadata(context.Background(), "")
	if err != nil {
		return nil, err
	}
	token := metadata[ProxyAuthorizationHeader]
	req.Header.Add(ProxyAuthorizationHeader, token)
	return c.transport.RoundTrip(req)
}

// Set up http client used in oauth2
func setHTTPClientContext(ctx context.Context, cfg *Config, proxyCredentialsFuture *PerRPCCredentialsFuture) context.Context {
	httpClient := &http.Client{}
	transport := &http.Transport{}

	if len(cfg.HTTPProxyURL.String()) > 0 {
		// create a transport that uses the proxy
		transport.Proxy = http.ProxyURL(&cfg.HTTPProxyURL.URL)
	}

	if cfg.ProxyCommand != nil && len(cfg.ProxyCommand) > 0 {
		httpClient.Transport = &proxyAuthTransport{
			transport:              transport,
			proxyCredentialsFuture: proxyCredentialsFuture,
		}
	} else {
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
func NewAuthInterceptor(cfg *Config, tokenCache cache.TokenCache, credentialsFuture *PerRPCCredentialsFuture, proxyCredentialsFuture *PerRPCCredentialsFuture) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = setHTTPClientContext(ctx, cfg, proxyCredentialsFuture)

		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			logger.Debugf(ctx, "Request failed due to [%v]. If it's an unauthenticated error, we will attempt to establish an authenticated context.", err)

			if st, ok := status.FromError(err); ok {
				// If the error we receive from executing the request expects
				if shouldAttemptToAuthenticate(st.Code()) {
					logger.Debugf(ctx, "Request failed due to [%v]. Attempting to establish an authenticated connection and trying again.", st.Code())
					newErr := MaterializeCredentials(ctx, cfg, tokenCache, credentialsFuture, proxyCredentialsFuture)
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

func NewProxyAuthInterceptor(cfg *Config, proxyCredentialsFuture *PerRPCCredentialsFuture) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			newErr := MaterializeProxyAuthCredentials(ctx, cfg, proxyCredentialsFuture)
			if newErr != nil {
				return fmt.Errorf("proxy authorization error! Original Error: %v, Proxy Auth Error: %w", err, newErr)
			}
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		return err
	}
}
