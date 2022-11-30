package tasks_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/flyteorg/flyteplugins/go/tasks/logs"
	flyteK8sConfig "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	sagemakerConfig "github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/sagemaker/config"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/k8s/spark"
	"github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/config/viper"
)

func TestLoadConfig(t *testing.T) {
	configAccessor := viper.NewAccessor(config.Options{
		StrictMode:  true,
		SearchPaths: []string{"testdata/config.yaml"},
	})

	err := configAccessor.UpdateConfig(context.TODO())
	assert.NoError(t, err)
	t.Run("k8s-config-test", func(t *testing.T) {

		k8sConfig := flyteK8sConfig.GetK8sPluginConfig()
		assert.True(t, k8sConfig.InjectFinalizer)
		assert.Equal(t, map[string]string{
			"annotationKey1": "annotationValue1",
			"annotationKey2": "annotationValue2",
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
		expectedCPU := resource.MustParse("1000m")
		assert.True(t, expectedCPU.Equal(k8sConfig.DefaultCPURequest))
		expectedMemory := resource.MustParse("1024Mi")
		assert.True(t, expectedMemory.Equal(k8sConfig.DefaultMemoryRequest))
		assert.Equal(t, map[string]string{"x/interruptible": "true"}, k8sConfig.InterruptibleNodeSelector)
		assert.Equal(t, "x/flyte", k8sConfig.InterruptibleTolerations[0].Key)
		assert.Equal(t, "interruptible", k8sConfig.InterruptibleTolerations[0].Value)
		assert.NotNil(t, k8sConfig.DefaultPodSecurityContext)
		assert.NotNil(t, k8sConfig.DefaultPodSecurityContext.FSGroup)
		assert.Equal(t, *k8sConfig.DefaultPodSecurityContext.FSGroup, int64(2000))
		assert.NotNil(t, k8sConfig.DefaultPodSecurityContext.RunAsGroup)
		assert.Equal(t, *k8sConfig.DefaultPodSecurityContext.RunAsGroup, int64(3000))
		assert.NotNil(t, k8sConfig.DefaultPodSecurityContext.RunAsUser)
		assert.Equal(t, *k8sConfig.DefaultPodSecurityContext.RunAsUser, int64(1000))
		assert.NotNil(t, k8sConfig.DefaultSecurityContext)
		assert.NotNil(t, k8sConfig.DefaultSecurityContext.AllowPrivilegeEscalation)
		assert.False(t, *k8sConfig.DefaultSecurityContext.AllowPrivilegeEscalation)
		assert.NotNil(t, k8sConfig.EnableHostNetworkingPod)
		assert.True(t, *k8sConfig.EnableHostNetworkingPod)
		assert.NotNil(t, k8sConfig.DefaultPodDNSConfig)
		assert.NotNil(t, k8sConfig.DefaultPodDNSConfig.Options)
		assert.NotNil(t, k8sConfig.DefaultPodDNSConfig.Options[0].Name)
		assert.Equal(t, "ndots", k8sConfig.DefaultPodDNSConfig.Options[0].Name)
		assert.NotNil(t, k8sConfig.DefaultPodDNSConfig.Options[0].Value)
		assert.Equal(t, "1", *k8sConfig.DefaultPodDNSConfig.Options[0].Value)
		assert.NotNil(t, k8sConfig.DefaultPodDNSConfig.Options[1].Name)
		assert.Equal(t, "single-request-reopen", k8sConfig.DefaultPodDNSConfig.Options[1].Name)
		assert.Nil(t, k8sConfig.DefaultPodDNSConfig.Options[1].Value)
		assert.NotNil(t, k8sConfig.DefaultPodDNSConfig.Options[2].Name)
		assert.Equal(t, "timeout", k8sConfig.DefaultPodDNSConfig.Options[2].Name)
		assert.NotNil(t, k8sConfig.DefaultPodDNSConfig.Options[2].Value)
		assert.Equal(t, "1", *k8sConfig.DefaultPodDNSConfig.Options[2].Value)
		assert.NotNil(t, k8sConfig.DefaultPodDNSConfig.Options[3].Name)
		assert.Equal(t, "attempts", k8sConfig.DefaultPodDNSConfig.Options[3].Name)
		assert.NotNil(t, k8sConfig.DefaultPodDNSConfig.Options[3].Value)
		assert.Equal(t, "3", *k8sConfig.DefaultPodDNSConfig.Options[3].Value)
		assert.NotNil(t, k8sConfig.DefaultPodDNSConfig.Nameservers)
		assert.Equal(t, []string{"8.8.8.8", "8.8.4.4"}, k8sConfig.DefaultPodDNSConfig.Nameservers)
		assert.NotNil(t, k8sConfig.DefaultPodDNSConfig.Searches)
		assert.Equal(t, []string{"ns1.svc.cluster-domain.example", "my.dns.search.suffix"}, k8sConfig.DefaultPodDNSConfig.Searches)
	})

	t.Run("logs-config-test", func(t *testing.T) {
		assert.NotNil(t, logs.GetLogConfig())
		assert.True(t, logs.GetLogConfig().IsKubernetesEnabled)
	})

	t.Run("spark-config-test", func(t *testing.T) {
		assert.NotNil(t, spark.GetSparkConfig())
		assert.NotNil(t, spark.GetSparkConfig().DefaultSparkConfig)
		assert.Equal(t, 2, len(spark.GetSparkConfig().Features))
		assert.Equal(t, "feature1", spark.GetSparkConfig().Features[0].Name)
		assert.Equal(t, "feature2", spark.GetSparkConfig().Features[1].Name)
		assert.Equal(t, 2, len(spark.GetSparkConfig().Features[0].SparkConfig))
		assert.Equal(t, 2, len(spark.GetSparkConfig().Features[1].SparkConfig))

	})

	t.Run("sagemaker-config-test", func(t *testing.T) {
		assert.NotNil(t, sagemakerConfig.GetSagemakerConfig())
	})
}

func TestLoadIncorrectConfig(t *testing.T) {
	t.Run("logs-config-test-accept-bad-config", func(t *testing.T) {
		configAccessor := viper.NewAccessor(config.Options{
			StrictMode:  false,
			SearchPaths: []string{"testdata/incorrect-config.yaml"},
		})

		err := configAccessor.UpdateConfig(context.TODO())
		assert.NoError(t, err)
		assert.NotNil(t, logs.GetLogConfig())
		assert.True(t, logs.GetLogConfig().IsKubernetesEnabled)
	})

	t.Run("logs-config-test-failfast", func(t *testing.T) {
		configAccessor := viper.NewAccessor(config.Options{
			StrictMode:  true,
			SearchPaths: []string{"testdata/incorrect-config.yaml"},
		})

		err := configAccessor.UpdateConfig(context.TODO())
		assert.Error(t, err)
	})
}
