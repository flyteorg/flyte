package plugin

import (
	"context"
	"fmt"

	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
)

// The webapi plugin machinery resolves task secrets through a SecretManager supplied by the host, at
// both plugin setup and task execution. The v2 executor never wired one up, so a connector task that
// references a secret dereferences a nil SecretManager and panics. This no-op restores the contract.
// It has no secret backend, so Get fails loudly with an error rather than handing a task a blank
// value. Swap in a real SecretManager to resolve secrets.
type noopSecretManager struct{}

var _ pluginsCore.SecretManager = noopSecretManager{}

func (noopSecretManager) Get(_ context.Context, key string) (string, error) {
	return "", fmt.Errorf("secrets are not supported by the executor's no-op secret manager, cannot resolve %q", key)
}

// NewNoopSecretManager returns a SecretManager with no backend that fails every lookup with a clear
// error, so a connector task referencing a secret does not silently receive an empty value.
func NewNoopSecretManager() pluginsCore.SecretManager { return noopSecretManager{} }
