/*
 * Copyright (c) 2018 Lyft. All rights reserved.
 */

package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/flyteorg/flytestdlib/config"

	pluginsConfig "github.com/flyteorg/flyteplugins/go/tasks/config"
)

//go:generate pflags Config --default-var defaultConfig

const ConfigSectionKey = "aws"

var (
	defaultConfig = &Config{
		Region:  "us-east-2",
		Retries: 3,
	}

	configSection = pluginsConfig.MustRegisterSubSection(ConfigSectionKey, defaultConfig)
)

// Config section for AWS Package
type Config struct {
	Region    string            `json:"region" pflag:",AWS Region to connect to."`
	AccountID string            `json:"accountId" pflag:",AWS Account Identifier."`
	Retries   int               `json:"retries" pflag:",Number of retries."`
	LogLevel  aws.ClientLogMode `json:"logLevel" pflag:"-,Defines the Sdk Log Level."`
}

type RateLimiterConfig struct {
	Rate  int64 `json:"rate" pflag:",Allowed rate of calls per second."`
	Burst int   `json:"burst" pflag:",Allowed burst rate of calls."`
}

// Gets loaded config for AWS
func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}

func (cfg Config) GetSdkConfig() (aws.Config, error) {
	sdkConfig, err := awsConfig.LoadDefaultConfig(context.TODO(),
		awsConfig.WithRegion(cfg.Region),
		awsConfig.WithRetryer(func() aws.Retryer {
			return retry.NewStandard(func(options *retry.StandardOptions) {
				options.MaxAttempts = cfg.Retries
			})
		}),
		awsConfig.WithClientLogMode(cfg.LogLevel))
	if err != nil {
		return aws.Config{}, err
	}

	return sdkConfig, nil
}

func MustRegisterSubSection(key config.SectionKey, cfg config.Config) config.Section {
	return configSection.MustRegisterSection(key, cfg)
}
