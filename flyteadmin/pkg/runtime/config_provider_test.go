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

func initConfig(cfg string) error {
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	configAccessor := viper.NewAccessor(config.Options{
		SearchPaths: []string{filepath.Join(pwd, cfg)},
		StrictMode:  false,
	})
	return configAccessor.UpdateConfig(context.TODO())
}

func TestClusterConfig(t *testing.T) {
	err := initConfig("testdata/clusters_config.yaml")
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

func TestGetCloudEventsConfig(t *testing.T) {
	err := initConfig("testdata/event.yaml")
	assert.NoError(t, err)
	configProvider := NewConfigurationProvider()
	cloudEventsConfig := configProvider.ApplicationConfiguration().GetCloudEventsConfig()
	assert.Equal(t, true, cloudEventsConfig.Enable)
	assert.Equal(t, "aws", cloudEventsConfig.Type)
	assert.Equal(t, "us-east-1", cloudEventsConfig.AWSConfig.Region)
	assert.Equal(t, "topic", cloudEventsConfig.EventsPublisherConfig.TopicName)
}

func TestPostgresConfig(t *testing.T) {
	err := initConfig("testdata/postgres_config.yaml")
	assert.NoError(t, err)

	configProvider := NewConfigurationProvider()
	dbConfig := configProvider.ApplicationConfiguration().GetDbConfig()
	assert.Equal(t, 5432, dbConfig.Postgres.Port)
	assert.Equal(t, "postgres", dbConfig.Postgres.Host)
	assert.Equal(t, "postgres", dbConfig.Postgres.User)
	assert.Equal(t, "postgres", dbConfig.Postgres.DbName)
	assert.Equal(t, "sslmode=disable", dbConfig.Postgres.ExtraOptions)
}

func TestSqliteConfig(t *testing.T) {
	err := initConfig("testdata/sqlite_config.yaml")
	assert.NoError(t, err)

	configProvider := NewConfigurationProvider()
	dbConfig := configProvider.ApplicationConfiguration().GetDbConfig()
	assert.Equal(t, "admin.db", dbConfig.SQLite.File)
}
