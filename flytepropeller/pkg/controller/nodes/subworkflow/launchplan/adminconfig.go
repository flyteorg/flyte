package launchplan

import (
	"time"

	ctrlConfig "github.com/flyteorg/flyte/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flyte/flytestdlib/config"
)

//go:generate pflags AdminConfig --default-var defaultAdminConfig

var (
	defaultAdminConfig = &AdminConfig{
		TPS:          100,
		Burst:        10,
		MaxCacheSize: 10000,
		Workers:      10,
		CacheResyncDuration: config.Duration{
			Duration: 30 * time.Second,
		},
	}

	adminConfigSection = ctrlConfig.MustRegisterSubSection("admin-launcher", defaultAdminConfig)
)

// AdminConfig provides a "admin-launcher" section in core Flytepropeller configuration and can be used to configure
// the rate at which Flytepropeller can query for status of workflows in flyteadmin or create new executions
type AdminConfig struct {
	// TPS indicates the maximum transactions per second to flyte admin from this client.
	// If it's zero, the created client will use DefaultTPS: 5
	TPS int64 `json:"tps" pflag:",The maximum number of transactions per second to flyte admin from this client."`

	// Maximum burst for throttle.
	// If it's zero, the created client will use DefaultBurst: 10.
	Burst int `json:"burst" pflag:",Maximum burst for throttle"`

	MaxCacheSize int `json:"cacheSize" pflag:",Maximum cache in terms of number of items stored."`

	Workers int `json:"workers" pflag:",Number of parallel workers to work on the queue."`

	// CacheResyncDuration defines the interval that the admin launcher should refresh the launchplan cache.
	CacheResyncDuration config.Duration `json:"cache-resync-duration" pflag:",Frequency of re-syncing launchplans within the auto refresh cache."`
}

func GetAdminConfig() *AdminConfig {
	return adminConfigSection.GetConfig().(*AdminConfig)
}
