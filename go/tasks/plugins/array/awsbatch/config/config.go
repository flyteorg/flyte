package config

import (
	"time"

	"github.com/flyteorg/flyteplugins/go/tasks/aws"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/workqueue"
	"github.com/flyteorg/flytestdlib/config"
)

//go:generate pflags Config --default-var defaultConfig

type Config struct {
	JobStoreConfig     JobStoreConfig        `json:"jobStoreConfig" pflag:",Config for job store"`
	JobDefCacheSize    int                   `json:"defCacheSize" pflag:",Maximum job definition cache size as number of items. Caches are used as an optimization to lessen the load on AWS Services."`
	GetRateLimiter     aws.RateLimiterConfig `json:"getRateLimiter" pflag:",Rate limiter config for batch get API."`
	DefaultRateLimiter aws.RateLimiterConfig `json:"defaultRateLimiter" pflag:",Rate limiter config for all batch APIs except get."`
	MaxArrayJobSize    int64                 `json:"maxArrayJobSize" pflag:",Maximum size of array job."`
	MinRetries         int32                 `json:"minRetries" pflag:",Minimum number of retries"`
	MaxRetries         int32                 `json:"maxRetries" pflag:",Maximum number of retries"`
	DefaultTimeOut     config.Duration       `json:"defaultTimeout" pflag:",Default timeout for the batch job."`
	// Provide additional environment variable pairs that plugin authors will provide to containers
	DefaultEnvVars       map[string]string `json:"defaultEnvVars" pflag:"-,Additional environment variable that should be injected into every resource"`
	MaxErrorStringLength int               `json:"maxErrLength" pflag:",Determines the maximum length of the error string returned for the array."`
	// This can be deprecated. Just having it for backward compatibility
	RoleAnnotationKey string           `json:"roleAnnotationKey" pflag:",Map key to use to lookup role from task annotations."`
	OutputAssembler   workqueue.Config `json:"outputAssembler"`
	ErrorAssembler    workqueue.Config `json:"errorAssembler"`
}

type JobStoreConfig struct {
	CacheSize      int             `json:"jacheSize" pflag:",Maximum informer cache size as number of items. Caches are used as an optimization to lessen the load on AWS Services."`
	Parallelizm    int             `json:"parallelizm"`
	BatchChunkSize int             `json:"batchChunkSize" pflag:",Determines the size of each batch sent to GetJobDetails api."`
	ResyncPeriod   config.Duration `json:"resyncPeriod" pflag:",Defines the duration for syncing job details from AWS Batch."`
}

var (
	defaultConfig = &Config{
		JobStoreConfig: JobStoreConfig{
			CacheSize:      10000,
			Parallelizm:    20,
			BatchChunkSize: 100,
			ResyncPeriod:   config.Duration{Duration: 30 * time.Second},
		},
		JobDefCacheSize: 10000,
		MaxArrayJobSize: 5000,
		GetRateLimiter: aws.RateLimiterConfig{
			Rate:  15,
			Burst: 20,
		},
		DefaultRateLimiter: aws.RateLimiterConfig{
			Rate:  15,
			Burst: 20,
		},
		MinRetries:           1,
		MaxRetries:           10,
		DefaultTimeOut:       config.Duration{Duration: 3 * 24 * time.Hour},
		MaxErrorStringLength: 500,
		OutputAssembler: workqueue.Config{
			IndexCacheMaxItems: 100000,
			MaxRetries:         5,
			Workers:            10,
		},
		ErrorAssembler: workqueue.Config{
			IndexCacheMaxItems: 100000,
			MaxRetries:         5,
			Workers:            10,
		},
	}

	configSection = aws.MustRegisterSubSection("batch", defaultConfig)
)

func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}
