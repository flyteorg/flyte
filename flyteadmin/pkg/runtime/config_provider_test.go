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
	cluster := clusterConfig.GetCurrentCluster()
	assert.Equal(t, "testcluster2", cluster.Name)
	assert.Equal(t, "testcluster2_endpoint", cluster.Endpoint)
	assert.Equal(t, "/path/to/testcluster2/cert", cluster.Auth.CertPath)
	assert.Equal(t, "/path/to/testcluster2/token", cluster.Auth.TokenPath)
	assert.Equal(t, "file_path", cluster.Auth.Type)
}
