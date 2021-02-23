package storage

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/internal/utils"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
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
		assert.NoError(t, ioutil.WriteFile(filepath.Join("testdata", "config.yaml"), raw, os.ModePerm))
	}

	actual := Config{}
	raw, err := ioutil.ReadFile(filepath.Join("testdata", "config.yaml"))
	assert.NoError(t, err)
	assert.NoError(t, yaml.Unmarshal(raw, &actual))
	assert.True(t, reflect.DeepEqual(expected, actual))
}
