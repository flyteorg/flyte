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

func initTestConfig() error {
	var searchPaths []string
	for _, goPath := range strings.Split(build.Default.GOPATH, string(os.PathListSeparator)) {
		searchPaths = append(searchPaths, filepath.Join(goPath, "src/github.com/lyft/flyteadmin/pkg/runtime/testdata/clusters_config.yaml"))
	}

	configAccessor := viper.NewAccessor(config.Options{
		SearchPaths: searchPaths,
		StrictMode:  false,
	})
	return configAccessor.UpdateConfig(context.TODO())
}

func TestClusterConfig(t *testing.T) {
	err := initTestConfig()
	assert.NoError(t, err)

	configProvider := NewConfigurationProvider()
	clusterConfig := configProvider.ClusterConfiguration()
	clusters := clusterConfig.GetClusterConfigs()
	assert.Equal(t, 2, len(clusters))

	assert.Equal(t, "testcluster", clusters[0].Name)
	assert.Equal(t, "testcluster_endpoint", clusters[0].Endpoint)
	assert.Equal(t, "/path/to/testcluster/cert", clusters[0].Auth.CertPath)
	assert.Equal(t, "/path/to/testcluster/token", clusters[0].Auth.TokenPath)
	assert.Equal(t, "file_path", clusters[0].Auth.Type)
	assert.False(t, clusters[0].Enabled)

	assert.Equal(t, "testcluster2", clusters[1].Name)
	assert.Equal(t, "testcluster2_endpoint", clusters[1].Endpoint)
	assert.Equal(t, "/path/to/testcluster2/cert", clusters[1].Auth.CertPath)
	assert.Equal(t, "/path/to/testcluster2/token", clusters[1].Auth.TokenPath)
	assert.True(t, clusters[1].Enabled)

	assert.Equal(t, "file_path", clusters[1].Auth.Type)
}
