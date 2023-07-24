package ray

import (
	"testing"

	pluginmachinery "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/k8s"
	"gotest.tools/assert"
)

func TestLoadConfig(t *testing.T) {
	rayConfig := GetConfig()
	assert.Assert(t, rayConfig != nil)

	t.Run("remote cluster", func(t *testing.T) {
		config := GetConfig()
		remoteConfig := pluginmachinery.ClusterConfig{
			Enabled:  false,
			Endpoint: "",
			Auth: pluginmachinery.Auth{
				TokenPath:  "",
				CaCertPath: "",
			},
		}
		assert.DeepEqual(t, config.RemoteClusterConfig, remoteConfig)
	})
}
