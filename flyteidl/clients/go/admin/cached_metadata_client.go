package admin

import (
	"context"
	"sync"

	"google.golang.org/grpc"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
)

// cachedMetadataClient wraps an AuthMetadataServiceClient and caches the
// GetPublicClientConfig response. The server returns static configuration
// that does not change at runtime, so a single fetch is sufficient.
type cachedMetadataClient struct {
	service.AuthMetadataServiceClient
	once   sync.Once
	cached *service.PublicClientAuthConfigResponse
	err    error
}

// NewCachedMetadataClient returns an AuthMetadataServiceClient that calls
// GetPublicClientConfig at most once and returns the cached result thereafter.
func NewCachedMetadataClient(inner service.AuthMetadataServiceClient) service.AuthMetadataServiceClient {
	return &cachedMetadataClient{AuthMetadataServiceClient: inner}
}

func (c *cachedMetadataClient) GetPublicClientConfig(
	ctx context.Context,
	in *service.PublicClientAuthConfigRequest,
	opts ...grpc.CallOption,
) (*service.PublicClientAuthConfigResponse, error) {
	c.once.Do(func() {
		c.cached, c.err = c.AuthMetadataServiceClient.GetPublicClientConfig(ctx, in, opts...)
	})
	return c.cached, c.err
}
