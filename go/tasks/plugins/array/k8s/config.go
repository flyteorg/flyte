/*
 * Copyright (c) 2018 Lyft. All rights reserved.
 */

package k8s

import (
	"fmt"
	"io/ioutil"

	"github.com/flyteorg/flyteplugins/go/tasks/logs"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	restclient "k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	pluginsConfig "github.com/flyteorg/flyteplugins/go/tasks/config"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/workqueue"
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
		LogConfig: LogConfig{
			Config: logs.DefaultConfig,
		},
	}

	configSection = pluginsConfig.MustRegisterSubSection(configSectionKey, defaultConfig)
)

type ResourceConfig struct {
	PrimaryLabel string `json:"primaryLabel" pflag:",PrimaryLabel of a given service cluster"`
	Limit        int    `json:"limit" pflag:",Resource quota (in the number of outstanding requests) for the cluster"`
}

type ClusterConfig struct {
	Name     string `json:"name" pflag:",Friendly name of the remote cluster"`
	Endpoint string `json:"endpoint" pflag:", Remote K8s cluster endpoint"`
	Auth     Auth   `json:"auth" pflag:"-, Auth setting for the cluster"`
	Enabled  bool   `json:"enabled" pflag:", Boolean flag to enable or disable"`
}

type Auth struct {
	Type      string `json:"type" pflag:", Authentication type"`
	TokenPath string `json:"tokenPath" pflag:", Token path"`
	CertPath  string `json:"certPath" pflag:", Certificate path"`
}

func (auth Auth) GetCA() ([]byte, error) {
	cert, err := ioutil.ReadFile(auth.CertPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read k8s CA cert from configured path")
	}
	return cert, nil
}

func (auth Auth) GetToken() (string, error) {
	token, err := ioutil.ReadFile(auth.TokenPath)
	if err != nil {
		return "", errors.Wrap(err, "failed to read k8s bearer token from configured path")
	}
	return string(token), nil
}

// RemoteClusterConfig reads secret values from paths specified in the config to initialize a Kubernetes rest client Config.
// TODO: Move logic to flytestdlib
func RemoteClusterConfig(host string, auth Auth) (*restclient.Config, error) {
	tokenString, err := auth.GetToken()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to get auth token: %+v", err))
	}

	caCert, err := auth.GetCA()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to get auth CA: %+v", err))
	}

	tlsClientConfig := restclient.TLSClientConfig{}
	tlsClientConfig.CAData = caCert
	return &restclient.Config{
		Host:            host,
		TLSClientConfig: tlsClientConfig,
		BearerToken:     tokenString,
	}, nil
}

func GetK8sClient(config ClusterConfig) (client.Client, error) {
	kubeConf, err := RemoteClusterConfig(config.Endpoint, config.Auth)
	if err != nil {
		return nil, err
	}
	return client.New(kubeConf, client.Options{})
}

// Config defines custom config for K8s Array plugin
type Config struct {
	DefaultScheduler     string            `json:"scheduler" pflag:",Decides the scheduler to use when launching array-pods."`
	MaxErrorStringLength int               `json:"maxErrorLength" pflag:",Determines the maximum length of the error string returned for the array."`
	MaxArrayJobSize      int64             `json:"maxArrayJobSize" pflag:",Maximum size of array job."`
	ResourceConfig       ResourceConfig    `json:"resourceConfig" pflag:"-,ResourceConfiguration to limit number of resources used by k8s-array."`
	RemoteClusterConfig  ClusterConfig     `json:"remoteClusterConfig" pflag:"-,Configuration of remote K8s cluster for array jobs"`
	NodeSelector         map[string]string `json:"node-selector" pflag:"-,Defines a set of node selector labels to add to the pod."`
	Tolerations          []v1.Toleration   `json:"tolerations"  pflag:"-,Tolerations to be applied for k8s-array pods"`
	NamespaceTemplate    string            `json:"namespaceTemplate"  pflag:"-,Namespace pattern to spawn array-jobs in. Defaults to parent namespace if not set"`
	OutputAssembler      workqueue.Config
	ErrorAssembler       workqueue.Config
	LogConfig            LogConfig `json:"logs" pflag:",Config for log links for k8s array jobs."`
}

type LogConfig struct {
	Config logs.LogConfig `json:"config" pflag:",Defines the log config for k8s logs."`
}

func GetConfig() *Config {
	return configSection.GetConfig().(*Config)
}

func IsResourceConfigSet(resourceConfig ResourceConfig) bool {
	emptyResouceConfig := ResourceConfig{}
	return resourceConfig != emptyResouceConfig
}
