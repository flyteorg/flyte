package config

import (
	"fmt"

	"github.com/flyteorg/flytestdlib/config"
)

const SectionKey = "application"

//go:generate pflags Config

type Config struct {
	GrpcPort                 int  `json:"grpcPort" pflag:",On which grpc port to serve Catalog"`
	GrpcServerReflection     bool `json:"grpcServerReflection" pflag:",Enable GRPC Server Reflection"`
	HTTPPort                 int  `json:"httpPort" pflag:",On which http port to serve Catalog"`
	Secure                   bool `json:"secure" pflag:",Whether to run Catalog in secure mode or not"`
	ReadHeaderTimeoutSeconds int  `json:"readHeaderTimeoutSeconds" pflag:",The amount of time allowed to read request headers."`
}

var defaultConfig = &Config{
	GrpcPort:             8081,
	HTTPPort:             8080,
	GrpcServerReflection: true,
	// Set the HTTP timeout to avoid security vulnerabilities with expired, inactive connections:
	// https://deepsource.io/directory/analyzers/go/issues/GO-S2114
	// just shy of requestTimeoutUpperBound
	ReadHeaderTimeoutSeconds: 32,
}
var applicationConfig = config.MustRegisterSection(SectionKey, defaultConfig)

func GetConfig() *Config {
	return applicationConfig.GetConfig().(*Config)
}

func SetConfig(c *Config) {
	if err := applicationConfig.SetConfig(c); err != nil {
		panic(err)
	}
}

func (c Config) GetGrpcHostAddress() string {
	return fmt.Sprintf(":%d", c.GrpcPort)
}

func (c Config) GetHTTPHostAddress() string {
	return fmt.Sprintf(":%d", c.HTTPPort)
}

func init() {
	SetConfig(&Config{})
}
