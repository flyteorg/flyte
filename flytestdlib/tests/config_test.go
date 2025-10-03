package tests

import (
	"context"
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/config/viper"
	"github.com/flyteorg/flyte/flytestdlib/internal/utils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

// Make sure existing config file(s) parse correctly before overriding them with this flag!
var update = flag.Bool("update", false, "Updates testdata")

func TestStorageAndLoggerConfig(t *testing.T) {
	type CompositeConfig struct {
		Storage storage.Config `json:"storage"`
		Logger  logger.Config  `json:"logger"`
	}

	expected := CompositeConfig{
		Storage: storage.Config{
			Type: "s3",
			Connection: storage.ConnectionConfig{
				Endpoint:   config.URL{URL: utils.MustParseURL("http://minio:9000")},
				AuthType:   "accesskey",
				AccessKey:  "minio",
				SecretKey:  "miniostorage",
				Region:     "us-east-1",
				DisableSSL: true,
			},
		},
		Logger: logger.Config{
			Level: logger.DebugLevel,
		},
	}

	configPath := filepath.Join("testdata", "combined.yaml")
	if *update {
		t.Log("Updating golden files.")
		raw, err := yaml.Marshal(expected)
		assert.NoError(t, err)
		assert.NoError(t, os.WriteFile(configPath, raw, os.ModePerm)) // #nosec G306
	}

	actual := CompositeConfig{}
	/* #nosec */
	raw, err := os.ReadFile(configPath)
	assert.NoError(t, err)
	assert.NoError(t, yaml.Unmarshal(raw, &actual))
	assert.True(t, reflect.DeepEqual(expected, actual))
}

func TestParseExistingConfig(t *testing.T) {
	accessor := viper.NewAccessor(config.Options{
		SearchPaths: []string{filepath.Join("testdata", "combined.yaml")},
	})

	assert.NoError(t, accessor.UpdateConfig(context.TODO()))

	assert.NotNil(t, storage.ConfigSection)

	if _, ok := storage.ConfigSection.GetConfig().(*storage.Config); ok {
		assert.True(t, ok)
	} else {
		assert.FailNow(t, "Retrieved section is not of type storage.")
	}
}
