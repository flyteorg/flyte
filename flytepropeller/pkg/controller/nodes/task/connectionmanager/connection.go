package connectionmanager

import (
	"context"
	flyteidl "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

// Env Var Lookup based on Prefix + SecretGroup + _ + SecretKey
const envVarLookupFormatter = "%s%s_%s"

// FileEnvSecretManager allows retrieving secrets mounted to this process through Env Vars or Files.
type FileEnvConnectionManager struct {
	secretPath string
	envPrefix  string
}

func (f FileEnvConnectionManager) Get(ctx context.Context, key string) (*flyteidl.Connection, error) {
	// TODO: Read connection from config map
	return &flyteidl.Connection{}, nil
}

func NewFileEnvConnectionManager(cfg *Config) FileEnvConnectionManager {
	return FileEnvConnectionManager{
		secretPath: cfg.SecretFilePrefix,
		envPrefix:  cfg.EnvironmentPrefix,
	}
}
