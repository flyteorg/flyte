package admin

import (
	"context"
	"sync"

	"google.golang.org/grpc"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
)

// cachedMetadataClient wraps an AuthMetadataServiceClient and caches the
// GetPublicClientConfig response. The server returns static configuration
// that does not change at runtime, so a single successful fetch is enough.
// Errors are not cached: a failed call (for example a transient RPC error or
// a cancelled context) is retried on the next call.
type cachedMetadataClient struct {
	service.AuthMetadataServiceClient
	mu     sync.Mutex
	cached *service.PublicClientAuthConfigResponse
}

// NewCachedMetadataClient returns an AuthMetadataServiceClient that calls
// GetPublicClientConfig until it succeeds once and returns the cached
// response thereafter.
func NewCachedMetadataClient(inner service.AuthMetadataServiceClient) service.AuthMetadataServiceClient {
	return &cachedMetadataClient{AuthMetadataServiceClient: inner}
}

func (c *cachedMetadataClient) GetPublicClientConfig(
	ctx context.Context,
	in *service.PublicClientAuthConfigRequest,
	opts ...grpc.CallOption,
) (*service.PublicClientAuthConfigResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cached != nil {
		return c.cached, nil
	}
	resp, err := c.AuthMetadataServiceClient.GetPublicClientConfig(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	c.cached = resp
	return resp, nil
}
