package admin

import (
	"context"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/flyteorg/flytestdlib/config"
	"github.com/stretchr/testify/assert"
)

func TestInitializeAndGetAdminClient(t *testing.T) {

	ctx := context.TODO()
	t.Run("legal", func(t *testing.T) {
		u, err := url.Parse("http://localhost:8089")
		assert.NoError(t, err)
		assert.NotNil(t, InitializeAdminClient(ctx, Config{
			Endpoint: config.URL{URL: *u},
		}))
	})

	t.Run("illegal", func(t *testing.T) {
		adminConnection = nil
		once = sync.Once{}
		assert.NotNil(t, InitializeAdminClient(ctx, Config{}))
	})
}

func TestInitializeMockAdminClient(t *testing.T) {
	c := InitializeMockAdminClient()
	assert.NotNil(t, c)
}

func TestGetAdditionalAdminClientConfigOptions(t *testing.T) {
	u, _ := url.Parse("localhost:8089")
	adminServiceConfig := Config{
		Endpoint:              config.URL{URL: *u},
		UseInsecureConnection: true,
		PerRetryTimeout:       config.Duration{Duration: 1 * time.Second},
		MaxRetries:            1,
	}
	opts := GetAdditionalAdminClientConfigOptions(adminServiceConfig)
	assert.Equal(t, 2, len(opts))
}
