package config

import (
	"fmt"

	"github.com/lyft/flytestdlib/config"
)

const SectionKey = "application"

//go:generate pflags Config
//TODO add instructions on how to generate certs

type Config struct {
	HTTPPort   int    `json:"httpPort" pflag:",On which http port to serve admin"`
	GrpcPort   int    `json:"grpcPort" pflag:",On which grpc port to serve admin"`
	KubeConfig string `json:"kube-config" pflag:",Path to kubernetes client config file."`
	Master     string `json:"master" pflag:",The address of the Kubernetes API server."`
	Secure     bool   `json:"secure" pflag:",Whether to run admin in secure mode or not"`
}

var applicationConfig = config.MustRegisterSection(SectionKey, &Config{})

func GetConfig() *Config {
	return applicationConfig.GetConfig().(*Config)
}

func SetConfig(c *Config) {
	if err := applicationConfig.SetConfig(c); err != nil {
		panic(err)
	}
}

func (c Config) GetHostAddress() string {
	return fmt.Sprintf(":%d", c.HTTPPort)
}

func (c Config) GetGrpcHostAddress() string {
	return fmt.Sprintf(":%d", c.GrpcPort)
}

func init() {
	SetConfig(&Config{})
}
