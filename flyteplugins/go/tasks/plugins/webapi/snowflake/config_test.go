package snowflake

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetAndSetConfig(t *testing.T) {
	cfg := defaultConfig
	cfg.DefaultWarehouse = "test-warehouse"
	cfg.WebAPI.Caching.Workers = 1
	cfg.WebAPI.Caching.ResyncInterval.Duration = 5 * time.Second
	err := SetConfig(&cfg)
	assert.NoError(t, err)
	assert.Equal(t, &cfg, GetConfig())
}
