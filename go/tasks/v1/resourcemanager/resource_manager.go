package resourcemanager

import (
	"context"
	"github.com/go-redis/redis"

	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/promutils"
)

//go:generate mockery -name ResourceManager -case=underscore

type AllocationStatus string

const (
	// This is the enum returned when there's an error
	AllocationUndefined                    AllocationStatus = "ResourceGranted"

	// Go for it
	AllocationStatusGranted                AllocationStatus = "ResourceGranted"

	// This means that no resources are available globally.  This is the only rejection message we use right now.
	AllocationStatusExhausted              AllocationStatus = "ResourceExhausted"

	// We're not currently using this - but this would indicate that things globally are okay, but that your
	// own namespace is too busy
	AllocationStatusNamespaceQuotaExceeded AllocationStatus = "NamespaceQuotaExceeded"
)

// Resource Manager manages a single resource type, and each allocation is of size one
type ResourceManager interface {
	AllocateResource(ctx context.Context, namespace string, allocationToken string) (AllocationStatus, error)
	ReleaseResource(ctx context.Context, namespace string, allocationToken string) error
}

type NoopResourceManager struct {
}

func (NoopResourceManager) AllocateResource(ctx context.Context, namespace string, allocationToken string) (
	AllocationStatus, error) {

	return AllocationStatusGranted, nil
}

func (NoopResourceManager) ReleaseResource(ctx context.Context, namespace string, allocationToken string) error {
	return nil
}

// Gets or creates a resource manager to the given resource name. This function is thread-safe and calling it with the
// same resource name will return the same instance of resource manager every time.
func GetOrCreateResourceManagerFor(ctx context.Context, resourceName string) (ResourceManager, error) {
	return NoopResourceManager{}, nil
}

func GetResourceManagerByType(ctx context.Context, managerType string, scope promutils.Scope, redisClient *redis.Client) (
	ResourceManager, error) {

	switch managerType {
	case "noop":
		logger.Infof(ctx, "Using the NOOP resource manager")
		return NoopResourceManager{}, nil
	case "redis":
		logger.Infof(ctx, "Using Redis based resource manager")
		return NewRedisResourceManager(ctx, redisClient, scope.NewSubScope("resourcemanager:redis"))
	}
	logger.Infof(ctx, "Using the NOOP resource manager by default")
	return NoopResourceManager{}, nil
}
