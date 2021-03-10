package controller

import (
	"context"
	"testing"
	"time"

	config2 "github.com/flyteorg/flytepropeller/pkg/controller/config"

	"github.com/flyteorg/flytestdlib/config"
	"github.com/stretchr/testify/assert"
)

func TestNewWorkQueue(t *testing.T) {
	ctx := context.TODO()

	t.Run("emptyConfig", func(t *testing.T) {
		cfg := config2.WorkqueueConfig{}
		w, err := NewWorkQueue(ctx, cfg, "q_test1")
		assert.NoError(t, err)
		assert.NotNil(t, w)
	})

	t.Run("simpleConfig", func(t *testing.T) {
		cfg := config2.WorkqueueConfig{
			Type: config2.WorkqueueTypeDefault,
		}
		w, err := NewWorkQueue(ctx, cfg, "q_test2")
		assert.NoError(t, err)
		assert.NotNil(t, w)
	})

	t.Run("bucket", func(t *testing.T) {
		cfg := config2.WorkqueueConfig{
			Type:     config2.WorkqueueTypeBucketRateLimiter,
			Capacity: 5,
			Rate:     1,
		}
		w, err := NewWorkQueue(ctx, cfg, "q_test3")
		assert.NoError(t, err)
		assert.NotNil(t, w)
	})

	t.Run("expfailure", func(t *testing.T) {
		cfg := config2.WorkqueueConfig{
			Type:      config2.WorkqueueTypeExponentialFailureRateLimiter,
			MaxDelay:  config.Duration{Duration: time.Second * 10},
			BaseDelay: config.Duration{Duration: time.Second * 1},
		}
		w, err := NewWorkQueue(ctx, cfg, "q_test4")
		assert.NoError(t, err)
		assert.NotNil(t, w)
	})
}
