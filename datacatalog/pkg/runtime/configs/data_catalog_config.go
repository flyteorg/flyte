package configs

import (
	"github.com/flyteorg/flytestdlib/config"
)

//go:generate pflags DataCatalogConfig --default-var=defaultConfig

var defaultConfig = &DataCatalogConfig{}

// DataCatalogConfig is the base configuration to start datacatalog
type DataCatalogConfig struct {
	StoragePrefix                  string          `json:"storage-prefix" pflag:",StoragePrefix specifies the prefix where DataCatalog stores offloaded ArtifactData in CloudStorage. If not specified, the data will be stored in the base container directly."`
	MetricsScope                   string          `json:"metrics-scope" pflag:",Scope that the metrics will record under."`
	ProfilerPort                   int             `json:"profiler-port" pflag:",Port that the profiling service is listening on."`
	HeartbeatGracePeriodMultiplier int             `json:"heartbeat-grace-period-multiplier" pflag:",Number of heartbeats before a reservation expires without an extension."`
	MaxReservationHeartbeat        config.Duration `json:"max-reservation-heartbeat" pflag:",The maximum available reservation extension heartbeat interval."`
}
