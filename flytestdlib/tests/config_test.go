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

	"github.com/flyteorg/flyte/v2/flytestdlib/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/config/viper"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
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
			Stow: storage.StowConfig{
				Kind: "s3",
				Config: map[string]string{
					"endpoint":  "http://minio:9000",
					"authType":  "accesskey",
					"accessKey": "minio",
					"secretKey": "miniostorage",
					"region":    "us-east-1",
				},
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
