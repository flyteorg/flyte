package config

//go:generate pflags Config --default-var=defaultConfig

import (
	"context"
	"net/url"
	"time"

	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/logger"

	pluginsConfig "github.com/flyteorg/flyteplugins/go/tasks/config"
)

const prestoConfigSectionKey = "presto"

func URLMustParse(s string) config.URL {
	r, err := url.Parse(s)
	if err != nil {
		logger.Panicf(context.TODO(), "Bad Presto URL Specified as default, error: %s", err)
	}
	if r != nil {
		return config.URL{URL: *r}
	}
	logger.Panicf(context.TODO(), "Nil Presto URL specified.", err)
	return config.URL{}
}

type RoutingGroupConfig struct {
	Name                             string  `json:"name" pflag:",The name of a given Presto routing group"`
	Limit                            int     `json:"limit" pflag:",Resource quota (in the number of outstanding requests) of the routing group"`
	ProjectScopeQuotaProportionCap   float64 `json:"projectScopeQuotaProportionCap" pflag:",A floating point number between 0 and 1, specifying the maximum proportion of quotas allowed to allocate to a project in the routing group"`
	NamespaceScopeQuotaProportionCap float64 `json:"namespaceScopeQuotaProportionCap" pflag:",A floating point number between 0 and 1, specifying the maximum proportion of quotas allowed to allocate to a namespace in the routing group"`
}

type RefreshCacheConfig struct {
	Name         string          `json:"name" pflag:",The name of the rate limiter"`
	SyncPeriod   config.Duration `json:"syncPeriod" pflag:",The duration to wait before the cache is refreshed again"`
	Workers      int             `json:"workers" pflag:",Number of parallel workers to refresh the cache"`
	LruCacheSize int             `json:"lruCacheSize" pflag:",Size of the cache"`
}

// To execute a single Presto query from a user's point of view, we actually need to send 5 different
// requests to Presto. Together these requests (i.e. queries) take care of retrieving the data, saving
// it to an external table, and performing cleanup.
//
// The Presto plugin currently uses a single allocation token for each set of 5 requests which
// correspond to a single user query. These means that in total, Flyte is able to work on
// 'PrestoConfig.RoutingGroups[routing_group_name].Limit' user queries at a time as configured in the
// configurations for the Presto plugin. This means means that at most, Flyte will be working on this
// number of user Presto queries at a time for each of the configured Presto routing groups.
//
// In addition, these 2 rate limiters control the rate at which requests are sent to Presto for
// read and write requests coming from this client. This includes requests to execute queries (write),
// requests to get the status of a query through the auto-refresh cache (read), and requests to cancel
// queries (write). Together with allocation tokens, these rate limiters will ensure that the rate of
// requests and the number of concurrent requests going to Presto don't overload the cluster.
//
// There is also another important aspect to consider in terms of how the resource manager (where
// allocation tokens get created from) interplays with the rate limiters. From the write side of things
// (e.g. executing a query), if the write rate limiter is low then it will block executing queries until
// the rate falls below the limit. Even if the rate is below the limit, if queries take a long time to
// execute, then you  will be blocked at the resource manager level which only allows a certain number
// of concurrent queries to execute at any given time. Similarly, in more extreme cases, if both the
// resource manager and the write limiter are configured to support a large number of queries but the
// auto refresh cache size is small, then the cache will fill up and items will gets evicted due to the
// cache's LRU nature before the Flyte propeller workers get a chance to update the status of these
// items.
type RateLimiterConfig struct {
	Rate  int64 `json:"rate" pflag:",Allowed rate of calls per second."`
	Burst int   `json:"burst" pflag:",Allowed burst rate of calls per second."`
}

var (
	defaultConfig = Config{
		Environment:         URLMustParse(""),
		DefaultRoutingGroup: "adhoc",
		DefaultUser:         "flyte-default-user",
		UseNamespaceAsUser:  true,
		RoutingGroupConfigs: []RoutingGroupConfig{{Name: "adhoc", Limit: 100}, {Name: "etl", Limit: 25}},
		RefreshCacheConfig: RefreshCacheConfig{
			Name:         "presto",
			SyncPeriod:   config.Duration{Duration: 5 * time.Second},
			Workers:      15,
			LruCacheSize: 10000,
		},
		ReadRateLimiterConfig: RateLimiterConfig{
			Rate:  10,
			Burst: 10,
		},
		WriteRateLimiterConfig: RateLimiterConfig{
			Rate:  5,
			Burst: 10,
		},
	}

	prestoConfigSection = pluginsConfig.MustRegisterSubSection(prestoConfigSectionKey, &defaultConfig)
)

// Presto plugin configs
type Config struct {
	Environment            config.URL           `json:"environment" pflag:",Environment endpoint for Presto to use"`
	DefaultRoutingGroup    string               `json:"defaultRoutingGroup" pflag:",Default Presto routing group"`
	DefaultUser            string               `json:"defaultUser" pflag:",Default Presto user"`
	UseNamespaceAsUser     bool                 `json:"useNamespaceAsUser" pflag:",Use the K8s namespace as the user"`
	RoutingGroupConfigs    []RoutingGroupConfig `json:"routingGroupConfigs" pflag:"-,A list of cluster configs. Each of the configs corresponds to a service cluster"`
	RefreshCacheConfig     RefreshCacheConfig   `json:"refreshCacheConfig" pflag:"Refresh cache config"`
	ReadRateLimiterConfig  RateLimiterConfig    `json:"readRateLimiterConfig" pflag:"Rate limiter config for read requests going to Presto"`
	WriteRateLimiterConfig RateLimiterConfig    `json:"writeRateLimiterConfig" pflag:"Rate limiter config for write requests going to Presto"`
}

// Retrieves the current config value or default.
func GetPrestoConfig() *Config {
	return prestoConfigSection.GetConfig().(*Config)
}

func SetPrestoConfig(cfg *Config) error {
	return prestoConfigSection.SetConfig(cfg)
}
