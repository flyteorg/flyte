package example

import (
	"context"
	"testing"

	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/config/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	configAccessor := viper.NewAccessor(config.Options{
		StrictMode:  true,
		SearchPaths: []string{"testdata/admin_plugin.yaml"},
	})

	err := configAccessor.UpdateConfig(context.TODO())
	assert.NoError(t, err)

	cfg := GetConfig()
	assert.Len(t, cfg.WebAPI.ResourceQuotas, 1)
}
