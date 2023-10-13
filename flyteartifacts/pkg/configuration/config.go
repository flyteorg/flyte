package configuration

import (
	"github.com/flyteorg/flyte/flytestdlib/config"
)

const artifactsServer = "artifactsServer"

type ApplicationConfiguration struct {
	MyTestVal string `json:"myTestValue" pflag:",Database configuration"`
}

var defaultApplicationConfiguration = ApplicationConfiguration{}

var ApplicationConfig = config.MustRegisterSection(artifactsServer, &defaultApplicationConfiguration)
