package connectionmanager

import (
	"context"
	"fmt"
	"os"

	flyteidl "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

// FileEnvConnectionManager allows retrieving secrets mounted to this process through Env Vars or Files.
type FileEnvConnectionManager struct{}

// Get retrieves a secret from the environment of the running process or from a file.
func (f FileEnvConnectionManager) Get(ctx context.Context, key string) (*flyteidl.Connection, error) {
	cfg := GetConfig()
	connection, ok := cfg.Connection[key]
	if !ok {
		return nil, fmt.Errorf("connection not found [%s]", key)
	}
	secret := make(map[string]string)
	for k, v := range connection.Secrets {
		// TODO: Read the secret from a local file
		v, ok := os.LookupEnv(v)
		if !ok {
			return nil, fmt.Errorf("secret not found in env [%s]", v)
		}
		secret[k] = v
	}
	config := make(map[string]string)
	for k, v := range connection.Configs {
		config[k] = v
	}

	return &flyteidl.Connection{Secrets: secret, Config: config}, nil
}

func NewFileEnvConnectionManager() FileEnvConnectionManager {
	return FileEnvConnectionManager{}
}
