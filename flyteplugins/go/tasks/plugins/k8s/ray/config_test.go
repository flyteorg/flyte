package ray

import (
	"testing"

	"gotest.tools/assert"

	pluginmachinery "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/k8s"
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

func TestLoadDefaultServiceAccountConfig(t *testing.T) {
	rayConfig := GetConfig()
	assert.Assert(t, rayConfig != nil)

	t.Run("serviceAccount", func(t *testing.T) {
		config := GetConfig()
		assert.Equal(t, config.ServiceAccount, "default")
	})
}
