// +build manualintegration
// Be sure to add this to your goland build settings Tags in order to get linting/testing
// In order to run these integration tests you will need to run
//
// 	kubectl -n flyte port-forward service/redis-resource-manager 6379:6379

package tests

import (
	"context"
	"github.com/lyft/flyteplugins/go/tasks/v1/resourcemanager"
	"github.com/lyft/flytestdlib/logger"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRedisLookasideBuffer(t *testing.T) {
	_ = logger.SetConfig(&logger.Config{
		IncludeSourceCode: true,
		Level:             6,
	})

	ctx := context.Background()
	redisClient, err := resourcemanager.NewRedisClient(ctx, "localhost:6379", "mypassword")
	assert.NoError(t, err)
	expiry := time.Duration(1) * time.Second // To ensure your local Redis cache stays clean
	buffer := resourcemanager.NewRedisLookasideBuffer(ctx, redisClient, "testPrefix", expiry)
	assert.NoError(t, err)

	err = buffer.ConfirmExecution(ctx, "mykey", "123456")
	assert.NoError(t, err)
	commandId, err := buffer.RetrieveExecution(ctx, "mykey")
	assert.Equal(t, "123456", commandId)
}
