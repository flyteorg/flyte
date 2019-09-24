package runtime

import (
	"context"
	"go/build"
	"os"
	"strings"
	"testing"

	"path/filepath"

	"github.com/lyft/flytestdlib/config"
	"github.com/lyft/flytestdlib/config/viper"
	"github.com/stretchr/testify/assert"
)

func initTestClusterResourceConfig() error {
	var searchPaths []string
	for _, goPath := range strings.Split(build.Default.GOPATH, string(os.PathListSeparator)) {
		searchPaths = append(searchPaths, filepath.Join(goPath, "src/github.com/lyft/flyteadmin/pkg/runtime/testdata/cluster_resource_config.yaml"))
	}

	configAccessor := viper.NewAccessor(config.Options{
		SearchPaths: searchPaths,
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
