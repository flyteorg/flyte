package configuration

import (
	"github.com/flyteorg/flyte/flytestdlib/config"
)

const artifactsServer = "artifactsServer"

type ApplicationConfiguration struct {
	MyTestValue string `json:"myTestValue" pflag:",Test config value"`
}

var ApplicationConfig = config.MustRegisterSection(artifactsServer, &ApplicationConfiguration{})
