package admin

import (
	"context"

	"google.golang.org/grpc"

	"github.com/flyteorg/flyteidl/clients/go/admin/cache"
)

// ClientsetBuilder is used to build the clientset. This allows custom token cache implementations to be plugged in.
type ClientsetBuilder struct {
	config     *Config
	tokenCache cache.TokenCache
	opts       []grpc.DialOption
}

// ClientSetBuilder is constructor function to be used by the clients in interacting with the builder
func ClientSetBuilder() *ClientsetBuilder {
	return &ClientsetBuilder{}
}

// WithConfig provides the admin config to be used for constructing the clientset
func (cb *ClientsetBuilder) WithConfig(config *Config) *ClientsetBuilder {
	cb.config = config
	return cb
}

// WithTokenCache allows pluggable token cache implemetations. eg; flytectl uses keyring as tokenCache
func (cb *ClientsetBuilder) WithTokenCache(tokenCache cache.TokenCache) *ClientsetBuilder {
	cb.tokenCache = tokenCache
	return cb
}

func (cb *ClientsetBuilder) WithDialOptions(opts ...grpc.DialOption) *ClientsetBuilder {
	cb.opts = opts
	return cb
}

// Build the clientset using the current state of the ClientsetBuilder
func (cb *ClientsetBuilder) Build(ctx context.Context) (*Clientset, error) {
	if cb.tokenCache == nil {
		cb.tokenCache = &cache.TokenCacheInMemoryProvider{}
	}

	if cb.config == nil {
		cb.config = GetConfig(ctx)
	}

	return initializeClients(ctx, cb.config, cb.tokenCache, cb.opts...)
}

func NewClientsetBuilder() *ClientsetBuilder {
	return &ClientsetBuilder{}
}
