package connectionmanager

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	coreMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
)

func TestConnectionManager(t *testing.T) {
	ctx := context.Background()
	fakeSecretManager := &coreMocks.SecretManager{}
	fakeSecretManager.OnGet(ctx, mock.Anything).Return("fake-token", nil)

	connectionManager := NewFileEnvConnectionManager()
	connectionManager.secretManager = fakeSecretManager

	cfg := defaultConfig
	cfg.Connection = map[string]Connection{
		"openai": {
			Secrets: map[string]string{
				"openai_api_key": "api_key",
			},
			Configs: map[string]string{
				"openai_organization": "flyteorg",
			},
		},
	}
	err := SetConfig(cfg)
	assert.Nil(t, err)

	connection, err := connectionManager.Get(ctx, "openai")
	assert.Nil(t, err)
	assert.Equal(t, "fake-token", connection.Secrets["openai_api_key"])
	assert.Equal(t, "flyteorg", connection.Configs["openai_organization"])
}
