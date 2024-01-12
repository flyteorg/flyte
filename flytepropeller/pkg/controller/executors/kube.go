package executors

import (
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate mockery -name Client -case=underscore

// Client is a friendlier controller-runtime client that gets passed to executors
type Client interface {
	// GetClient returns a client configured with the Config
	GetClient() client.Client

	// GetCache returns a cache.Cache
	GetCache() cache.Cache
}
