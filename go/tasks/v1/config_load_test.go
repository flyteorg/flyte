package v1_test

import (
	"context"
	"testing"

	"github.com/lyft/flytestdlib/config"
	"github.com/lyft/flytestdlib/config/viper"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"

	pluginsConfig "github.com/lyft/flyteplugins/go/tasks/v1/config"
	flyteK8sConfig "github.com/lyft/flyteplugins/go/tasks/v1/flytek8s/config"
	"github.com/lyft/flyteplugins/go/tasks/v1/k8splugins"
	"github.com/lyft/flyteplugins/go/tasks/v1/logs"
	quboleConfig "github.com/lyft/flyteplugins/go/tasks/v1/qubole/config"
)

func TestLoadConfig(t *testing.T) {
	configAccessor := viper.NewAccessor(config.Options{
		StrictMode:  true,
		SearchPaths: []string{"testdata/config.yaml"},
	})

	err := configAccessor.UpdateConfig(context.TODO())
	assert.NoError(t, err)

	t.Run("root-config-test", func(t *testing.T) {
		assert.Equal(t, 1, len(pluginsConfig.GetConfig().EnabledPlugins))
	})

	t.Run("k8s-config-test", func(t *testing.T) {

		k8sConfig := flyteK8sConfig.GetK8sPluginConfig()
		assert.True(t, k8sConfig.InjectFinalizer)
		assert.Equal(t, map[string]string{
			"annotationKey1": "annotationValue1",
			"annotationKey2": "annotationValue2",
			"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
		}, k8sConfig.DefaultAnnotations)
		assert.Equal(t, map[string]string{
			"label1": "labelValue1",
			"label2": "labelValue2",
		}, k8sConfig.DefaultLabels)
		assert.Equal(t, map[string]string{
			"AWS_METADATA_SERVICE_NUM_ATTEMPTS": "20",
			"AWS_METADATA_SERVICE_TIMEOUT":      "5",
			"FLYTE_AWS_ACCESS_KEY_ID":           "minio",
			"FLYTE_AWS_ENDPOINT":                "http://minio.flyte:9000",
			"FLYTE_AWS_SECRET_ACCESS_KEY":       "miniostorage",
		}, k8sConfig.DefaultEnvVars)
		assert.NotNil(t, k8sConfig.ResourceTolerations)
		assert.Contains(t, k8sConfig.ResourceTolerations, v1.ResourceName("nvidia.com/gpu"))
		assert.Contains(t, k8sConfig.ResourceTolerations, v1.ResourceStorage)
		tolGPU := v1.Toleration{
			Key:      "flyte/gpu",
			Value:    "dedicated",
			Operator: v1.TolerationOpEqual,
			Effect:   v1.TaintEffectNoSchedule,
		}

		tolStorage := v1.Toleration{
			Key:      "storage",
			Value:    "special",
			Operator: v1.TolerationOpEqual,
			Effect:   v1.TaintEffectPreferNoSchedule,
		}

		assert.Equal(t, []v1.Toleration{tolGPU}, k8sConfig.ResourceTolerations[v1.ResourceName("nvidia.com/gpu")])
		assert.Equal(t, []v1.Toleration{tolStorage}, k8sConfig.ResourceTolerations[v1.ResourceStorage])
		assert.Equal(t, "1000m", k8sConfig.DefaultCpuRequest)
		assert.Equal(t, "1024Mi", k8sConfig.DefaultMemoryRequest)
	})

	t.Run("logs-config-test", func(t *testing.T) {
		assert.NotNil(t, logs.GetLogConfig())
		assert.True(t, logs.GetLogConfig().IsKubernetesEnabled)
	})

	t.Run("spark-config-test", func(t *testing.T) {
		assert.NotNil(t, k8splugins.GetSparkConfig())
		assert.NotNil(t, k8splugins.GetSparkConfig().DefaultSparkConfig)
	})

	t.Run("qubole-config-test", func(t *testing.T) {
		assert.NotNil(t, quboleConfig.GetQuboleConfig())
		assert.Equal(t, "redis-resource-manager.flyte:6379", quboleConfig.GetQuboleConfig().RedisHostPath)
	})

	t.Run("waitable-config-test", func(t *testing.T) {
		assert.NotNil(t, k8splugins.GetWaitableConfig())
		assert.Equal(t, "http://localhost:30081/console", k8splugins.GetWaitableConfig().ConsoleURI.String())
	})
}
