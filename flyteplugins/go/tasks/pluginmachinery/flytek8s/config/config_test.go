package config

import (
	"testing"
	"time"

	"gotest.tools/assert"
)

func TestGetK8sPluginConfig(t *testing.T) {
	assert.Equal(t, GetK8sPluginConfig().DefaultCPURequest, defaultCPURequest)
	assert.Equal(t, GetK8sPluginConfig().DefaultMemoryRequest, defaultMemoryRequest)
	assert.Equal(t, GetK8sPluginConfig().QuotaWaitTimeout.Duration, time.Hour)
}
