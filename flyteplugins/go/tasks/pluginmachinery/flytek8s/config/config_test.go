package config

import (
	"testing"

	"gotest.tools/assert"
)

func TestGetK8sPluginConfig(t *testing.T) {
	assert.Equal(t, GetK8sPluginConfig().DefaultCPURequest, defaultCPURequest)
	assert.Equal(t, GetK8sPluginConfig().DefaultMemoryRequest, defaultMemoryRequest)
}
