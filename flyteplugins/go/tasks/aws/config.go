/*
 * Copyright (c) 2018 Lyft. All rights reserved.
 */

package aws

import (
	"time"

	"github.com/lyft/flytestdlib/config"

	pluginsConfig "github.com/lyft/flyteplugins/go/tasks/config"
)

//go:generate pflags Config --default-var defaultConfig

const ConfigSectionKey = "aws"

var (
	defaultConfig = &Config{
		Region:               "us-east-1",
		MaxErrorStringLength: 150,
		Retries:              3,
		CatalogCacheTimeout:  config.Duration{Duration: time.Second * 5},
	}

	configSection = pluginsConfig.MustRegisterSubSection(ConfigSectionKey, defaultConfig)
)

// Config section for AWS Package
type Config struct {
	Region               string          `json:"region" pflag:",AWS Region to connect to."`
	AccountID            string          `json:"accountId" pflag:",AWS Account Identifier."`
	Retries              int             `json:"retries" pflag:",Number of retries."`
	MaxErrorStringLength int             `json:"maxErrorLength" pflag:",Maximum size of error messages."`
	CatalogCacheTimeout  config.Duration `json:"catalog-timeout" pflag:"\"5s\",Timeout duration for checking catalog for all batch tasks"`
}

type RateLimiterConfig struct {
	Rate  int64 `json:"rate" pflag:",Allowed rate of calls per second."`
	Burst int   `json:"burst" pflag:",Allowed burst rate of calls."`
}

// Gets loaded config for AWS
func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}

func MustRegisterSubSection(key config.SectionKey, cfg config.Config) config.Section {
	return configSection.MustRegisterSection(key, cfg)
}
