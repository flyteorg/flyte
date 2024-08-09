package configs

import (
	"time"

	"github.com/flyteorg/flyte/flytestdlib/config"
)

type DataStoreType = string

const (
	Mem      DataStoreType = "mem"
	Postgres DataStoreType = "postgres"
)

//go:generate pflags CacheServiceConfig --default-var=defaultConfig

var defaultConfig = &CacheServiceConfig{
	StoragePrefix:                  "cached_outputs",
	MetricsScope:                   "flyte",
	ProfilerPort:                   10254,
	HeartbeatGracePeriodMultiplier: 3,
	MaxReservationHeartbeat:        config.Duration{Duration: time.Second * 10},
	OutputDataStoreType:            Postgres,
	ReservationDataStoreType:       Postgres,
	MaxInlineSizeBytes:             0,
	AwsRegion:                      "us-west-2",
	RedisAddress:                   "localhost:6379",
	RedisUsername:                  "",
	RedisPassword:                  "",
}

// CacheServiceConfig is the base configuration to start cacheservice
type CacheServiceConfig struct {
	StoragePrefix                  string          `json:"storage-prefix" pflag:",StoragePrefix specifies the prefix where CacheService stores offloaded output in CloudStorage. If not ..."`
	MetricsScope                   string          `json:"metrics-scope" pflag:",Scope that the metrics will record under."`
	ProfilerPort                   int             `json:"profiler-port" pflag:",Port that the profiling service is listening on."`
	HeartbeatGracePeriodMultiplier int             `json:"heartbeat-grace-period-multiplier" pflag:",Number of heartbeats before a reservation expires without an extension."`
	MaxReservationHeartbeat        config.Duration `json:"max-reservation-heartbeat" pflag:",The maximum available reservation extension heartbeat interval."`
	OutputDataStoreType            DataStoreType   `json:"data-store-type" pflag:",Cache storage implementation to use"`
	ReservationDataStoreType       DataStoreType   `json:"reservation-data-store-type" pflag:",Reservation storage implementation to use"`
	MaxInlineSizeBytes             int64           `json:"maxInlineSizeBytes" pflag:",The maximum size that an output literal will be stored in line. Default 0 means everything will be offloaded to blob storage."`
	AwsRegion                      string          `json:"aws-region" pflag:",Region to connect to."`
	RedisAddress                   string          `json:"redis-address" pflag:",Address of the Redis server."`
	RedisUsername                  string          `json:"redis-username" pflag:",Username for the Redis server."`
	RedisPassword                  string          `json:"redis-password" pflag:",Password for the Redis server."`
}
