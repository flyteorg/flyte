package webapi

import (
	"context"
	"testing"
	"time"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/webapi"
	"github.com/flyteorg/flytestdlib/config"
)

func Test_validateConfig(t *testing.T) {
	t.Run("In range", func(t *testing.T) {
		cfg := webapi.PluginConfig{
			ReadRateLimiter: webapi.RateLimiterConfig{
				QPS:   10,
				Burst: 100,
			},
			WriteRateLimiter: webapi.RateLimiterConfig{
				QPS:   10,
				Burst: 100,
			},
			Caching: webapi.CachingConfig{
				Size:           10,
				ResyncInterval: config.Duration{Duration: 10 * time.Second},
				Workers:        10,
			},
		}

		assert.NoError(t, validateConfig(cfg))
	})

	t.Run("Below min", func(t *testing.T) {
		cfg := webapi.PluginConfig{
			ReadRateLimiter: webapi.RateLimiterConfig{
				QPS:   0,
				Burst: 0,
			},
			WriteRateLimiter: webapi.RateLimiterConfig{
				QPS:   0,
				Burst: 0,
			},
			Caching: webapi.CachingConfig{
				Size:           0,
				ResyncInterval: config.Duration{Duration: 0 * time.Second},
				Workers:        0,
			},
		}

		err := validateConfig(cfg)
		assert.Error(t, err)
		assert.Equal(t, "\ncache size is expected to be between 10 and 500000. Provided value is 0\nworkers count is expected to be between 1 and 100. Provided value is 0\nresync interval is expected to be between 5 and 3600. Provided value is 0\nread burst is expected to be between 5 and 10000. Provided value is 0\nread qps is expected to be between 1 and 100000. Provided value is 0\nwrite burst is expected to be between 5 and 10000. Provided value is 0\nwrite qps is expected to be between 1 and 100000. Provided value is 0", err.Error())
	})

	t.Run("Above max", func(t *testing.T) {
		cfg := webapi.PluginConfig{
			ReadRateLimiter: webapi.RateLimiterConfig{
				QPS:   1000,
				Burst: 1000000,
			},
			WriteRateLimiter: webapi.RateLimiterConfig{
				QPS:   1000,
				Burst: 1000000,
			},
			Caching: webapi.CachingConfig{
				Size:           1000000000,
				ResyncInterval: config.Duration{Duration: 10000 * time.Hour},
				Workers:        1000000000,
			},
		}

		err := validateConfig(cfg)
		assert.Error(t, err)
		assert.Equal(t, "\ncache size is expected to be between 10 and 500000. Provided value is 1000000000\nworkers count is expected to be between 1 and 100. Provided value is 1000000000\nresync interval is expected to be between 5 and 3600. Provided value is 3.6e+07\nread burst is expected to be between 5 and 10000. Provided value is 1000000\nwrite burst is expected to be between 5 and 10000. Provided value is 1000000", err.Error())
	})
}

func TestCreateRemotePlugin(t *testing.T) {
	CreateRemotePlugin(webapi.PluginEntry{
		ID:                 "MyTestPlugin",
		SupportedTaskTypes: []core.TaskType{"test-task"},
		PluginLoader: func(ctx context.Context, iCtx webapi.PluginSetupContext) (webapi.AsyncPlugin, error) {
			return newPluginWithProperties(webapi.PluginConfig{
				Caching: webapi.CachingConfig{
					Size: 10,
				},
			}), nil
		},
		IsDefault:           false,
		DefaultForTaskTypes: []core.TaskType{"test-task"},
	})
}
