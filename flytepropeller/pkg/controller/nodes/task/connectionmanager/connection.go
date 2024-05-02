package connectionmanager

import (
	"context"
	"fmt"

	flyteidl "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	pluginCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/task/secretmanager"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

type connectionManager string

const (
	noop    connectionManager = "noop"
	fileEnv connectionManager = "fileEnv"
)

type NoopConnectionManager struct{}

// Get returns an empty connection.
func (n NoopConnectionManager) Get(ctx context.Context, key string) (flyteidl.Connection, error) {
	logger.Debugf(ctx, "NoopConnectionManager Get called with key [%s]", key)
	return flyteidl.Connection{}, nil
}

// FileEnvConnectionManager allows retrieving secrets mounted to this process through Env Vars or Files.
type FileEnvConnectionManager struct {
	secretManager pluginCore.SecretManager
}

// Get retrieves a secret from the environment of the running process or from a file.
func (f FileEnvConnectionManager) Get(ctx context.Context, key string) (flyteidl.Connection, error) {
	cfg := GetConfig()
	connection, ok := cfg.Connection[key]
	if !ok {
		return flyteidl.Connection{}, fmt.Errorf("connection not found: [%s]", key)
	}
	secret := make(map[string]string)
	for k, v := range connection.Secrets {
		val, err := f.secretManager.Get(ctx, v)
		if err != nil {
			logger.Errorf(ctx, "failed to get secret [%s] for connection [%s] with error: %v", v, k, err)
			return flyteidl.Connection{}, err
		}
		secret[k] = val
	}

	return flyteidl.Connection{Secrets: secret, Configs: connection.Configs}, nil
}

func NewConnectionManager(config *Config) ConnectionManager {
	switch config.Type {
	case fileEnv:
		return FileEnvConnectionManager{
			secretManager: secretmanager.NewFileEnvSecretManager(secretmanager.GetConfig()),
		}
	case noop:
		fallthrough
	default:
		return NoopConnectionManager{}
	}
}
