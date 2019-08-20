// +build manualintegration
// Be sure to add this to your goland build settings Tags in order to get linting/testing
// In order to run these integration tests you will need to run
//
// 	kubectl -n flyte port-forward service/redis-resource-manager 6379:6379

package tests

import (
	"context"
	"fmt"
	"github.com/lyft/flyteplugins/go/tasks/v1/resourcemanager"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRedisResourceManager(t *testing.T) {
	_ = logger.SetConfig(&logger.Config{
		IncludeSourceCode: true,
		Level:             6,
	})

	ctx := context.Background()
	scope := promutils.NewScope("test")
	redisClient, err := resourcemanager.NewRedisClient(ctx, "localhost:6379", "mypassword")
	assert.NoError(t, err)
	manager, err := resourcemanager.NewRedisResourceManager(ctx, redisClient, scope)
	assert.NoError(t, err)

	status, err := manager.AllocateResource(ctx, "default", "my-token-1")
	assert.Equal(t, resourcemanager.AllocationStatusGranted, status)

	fmt.Println(status)
}
