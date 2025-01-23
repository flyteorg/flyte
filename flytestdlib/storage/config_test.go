package storage

import (
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/internal/utils"
)

// Make sure existing config file(s) parse correctly before overriding them with this flag!
var update = flag.Bool("update", false, "Updates testdata")

func TestMarshal(t *testing.T) {
	expected := Config{
		Type: "s3",
		Connection: ConnectionConfig{
			Endpoint:   config.URL{URL: utils.MustParseURL("http://minio:9000")},
			AuthType:   "accesskey",
			AccessKey:  "minio",
			SecretKey:  "miniostorage",
			Region:     "us-east-1",
			DisableSSL: true,
		},
	}

	if *update {
		t.Log("Updating config file.")
		raw, err := yaml.Marshal(expected)
		assert.NoError(t, err)
		assert.NoError(t, os.WriteFile(filepath.Join("testdata", "config.yaml"), raw, os.ModePerm)) // #nosec G306
	}

	actual := Config{}
	raw, err := os.ReadFile(filepath.Join("testdata", "config.yaml"))
	assert.NoError(t, err)
	assert.NoError(t, yaml.Unmarshal(raw, &actual))
	assert.True(t, reflect.DeepEqual(expected, actual))
}
