package runtime

import (
	"context"
	"os"
	"testing"

	"path/filepath"

	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/config/viper"
	"github.com/stretchr/testify/assert"
)

func initTestClusterResourceConfig() error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	configAccessor := viper.NewAccessor(config.Options{
		SearchPaths: []string{filepath.Join(pwd, "testdata/cluster_resource_config.yaml")},
		StrictMode:  false,
	})
	return configAccessor.UpdateConfig(context.TODO())
}

func TestClusterResourceConfig(t *testing.T) {
	err := initTestClusterResourceConfig()
	assert.NoError(t, err)

	configProvider := NewConfigurationProvider()
	clusterResourceConfig := configProvider.ClusterResourceConfiguration()
	assert.Equal(t, "/etc/flyte/clusterresource/templates", clusterResourceConfig.GetTemplatePath())
	assert.Equal(t, "flyte_user", clusterResourceConfig.GetTemplateData()["user"].Value)
	assert.Equal(t, "TEST_SECRET", clusterResourceConfig.GetTemplateData()["secret"].ValueFrom.EnvVar)
	assert.Equal(t, "/etc/flyte/misc.txt", clusterResourceConfig.GetTemplateData()["file"].ValueFrom.FilePath)
}
