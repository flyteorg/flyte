/*
 * Copyright (c) 2018 Lyft. All rights reserved.
 */

package k8s

import (
	v1 "k8s.io/api/core/v1"

	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/workqueue"
	"github.com/lyft/flytestdlib/config"
)

//go:generate pflags Config --default-var=defaultConfig

const configSectionKey = "k8s-array"

var (
	defaultConfig = &Config{
		MaxErrorStringLength: 1000,
		MaxArrayJobSize:      5000,
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

	configSection = config.MustRegisterSection(configSectionKey, defaultConfig)
)

// Defines custom config for K8s Array plugin
type Config struct {
	DefaultScheduler     string            `json:"scheduler" pflag:",Decides the scheduler to use when launching array-pods."`
	MaxErrorStringLength int               `json:"maxErrLength" pflag:",Determines the maximum length of the error string returned for the array."`
	MaxArrayJobSize      int64             `json:"maxArrayJobSize" pflag:",Maximum size of array job."`
	NodeSelector         map[string]string `json:"node-selector" pflag:"-,Defines a set of node selector labels to add to the pod."`
	Tolerations          []v1.Toleration   `json:"tolerations"  pflag:"-,Tolerations to be applied for k8s-array pods"`
	OutputAssembler      workqueue.Config
	ErrorAssembler       workqueue.Config
}

func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}
