package flytek8s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	pluginserrors "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/errors"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	stdErrors "github.com/flyteorg/flyte/v2/flytestdlib/errors"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

var appleSiliconAccelerator = &core.GPUAccelerator{Device: "AppleSilicon", DeviceClass: core.GPUAccelerator_APPLE_SILICON}

// setK8sConfig swaps in cfg for the rest of the test and puts the previous config back afterwards:
// other tests in this package rely on the defaults (e.g. AcceleratorDeviceClasses) being in place.
func setK8sConfig(t *testing.T, cfg *config.K8sPluginConfig) {
	t.Helper()
	previous := *config.GetK8sPluginConfig()
	t.Cleanup(func() { require.NoError(t, config.SetK8sPluginConfig(&previous)) })
	require.NoError(t, config.SetK8sPluginConfig(cfg))
}

func envMap(c v1.Container) map[string]string {
	m := map[string]string{}
	for _, e := range c.Env {
		m[e.Name] = e.Value
	}
	return m
}

func TestMintAppleSiliconToken(t *testing.T) {
	// Same vector as flyte-sdk's tests/flyte/apple_silicon/test_apple_silicon_api.py: both sides must agree on the format.
	assert.Equal(t,
		"as1.abc-123.7b94a54d7edb63bba60a50cb2818a3d36fcd12ed3a85b68650c57e010e03cc50",
		MintAppleSiliconToken("s3cret", "abc-123"))
}

func TestIsAppleSilicon(t *testing.T) {
	assert.True(t, IsAppleSilicon(appleSiliconAccelerator))
	assert.False(t, IsAppleSilicon(&core.GPUAccelerator{Device: "nvidia-tesla-a100"}))
	assert.False(t, IsAppleSilicon(nil))
}

func TestValidateAppleSiliconBridge(t *testing.T) {
	t.Run("cluster without a bridge", func(t *testing.T) {
		setK8sConfig(t, &config.K8sPluginConfig{})
		err := ValidateAppleSiliconBridge()
		require.Error(t, err)
		assert.True(t, stdErrors.IsCausedBy(err, pluginserrors.AppleSiliconNotAvailable))
		assert.Contains(t, err.Error(), "never cluster-schedulable")
	})

	t.Run("devbox on a host that is not Apple Silicon", func(t *testing.T) {
		setK8sConfig(t, &config.K8sPluginConfig{
			AppleSiliconBridge: config.AppleSiliconBridgeConfig{HostPlatform: "linux/amd64"},
		})
		err := ValidateAppleSiliconBridge()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "requires an Apple Silicon devbox; this host is linux/amd64")
	})

	t.Run("Apple Silicon devbox started without the bridge", func(t *testing.T) {
		setK8sConfig(t, &config.K8sPluginConfig{
			AppleSiliconBridge: config.AppleSiliconBridgeConfig{HostPlatform: "darwin/arm64"},
		})
		err := ValidateAppleSiliconBridge()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "flyte start devbox --apple-silicon")
	})

	t.Run("unrendered chart placeholders are not a bridge", func(t *testing.T) {
		setK8sConfig(t, &config.K8sPluginConfig{
			AppleSiliconBridge: config.AppleSiliconBridgeConfig{
				Endpoint:     "%{APPLE_SILICON_BRIDGE_ENDPOINT}%",
				TokenSecret:  "%{APPLE_SILICON_BRIDGE_TOKEN_SECRET}%",
				HostPlatform: "%{DEVBOX_HOST_PLATFORM}%",
			},
		})
		err := ValidateAppleSiliconBridge()
		require.Error(t, err)
		assert.True(t, stdErrors.IsCausedBy(err, pluginserrors.AppleSiliconNotAvailable))
		assert.NotContains(t, err.Error(), "%{")
	})

	t.Run("devbox with a bridge", func(t *testing.T) {
		setK8sConfig(t, &config.K8sPluginConfig{
			AppleSiliconBridge: config.AppleSiliconBridgeConfig{
				Endpoint: "http://192.168.65.254:30090", TokenSecret: "s3cret", HostPlatform: "darwin/arm64",
			},
		})
		assert.NoError(t, ValidateAppleSiliconBridge())
	})
}

func TestApplyAppleSiliconBridge(t *testing.T) {
	setK8sConfig(t, &config.K8sPluginConfig{
		GpuResourceName: ResourceNvidiaGPU,
		AppleSiliconBridge: config.AppleSiliconBridgeConfig{
			Endpoint: "http://192.168.65.254:30090/", TokenSecret: "s3cret",
		},
	})
	gpu := resource.MustParse("1")
	podSpec := &v1.PodSpec{
		Containers: []v1.Container{
			{
				Name: "primary",
				Env:  []v1.EnvVar{{Name: OpenAIBaseURLEnvVar, Value: "http://user-choice/v1"}},
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{resourceGPU: gpu, v1.ResourceCPU: resource.MustParse("1")},
					Limits:   v1.ResourceList{ResourceNvidiaGPU: gpu},
				},
			},
			{Name: "sidecar", Resources: v1.ResourceRequirements{Limits: v1.ResourceList{ResourceNvidiaGPU: gpu}}},
		},
	}

	require.NoError(t, ApplyAppleSiliconBridge(podSpec, "primary", "abc-123"))

	for _, c := range podSpec.Containers {
		assert.NotContains(t, c.Resources.Requests, resourceGPU, c.Name)
		assert.NotContains(t, c.Resources.Limits, ResourceNvidiaGPU, c.Name)
	}
	assert.Contains(t, podSpec.Containers[0].Resources.Requests, v1.ResourceCPU)

	env := envMap(podSpec.Containers[0])
	token := MintAppleSiliconToken("s3cret", "abc-123")
	assert.Equal(t, "http://192.168.65.254:30090", env[AppleSiliconEndpointEnvVar])
	assert.Equal(t, "abc-123", env[AppleSiliconWorkloadEnvVar])
	assert.Equal(t, token, env[AppleSiliconTokenEnvVar])
	// The user pointed OPENAI_BASE_URL elsewhere: the bridge token must not be sent there as their API key.
	assert.Equal(t, "http://user-choice/v1", env[OpenAIBaseURLEnvVar])
	assert.NotContains(t, env, OpenAIAPIKeyEnvVar)
	assert.Empty(t, podSpec.Containers[1].Env, "only the primary container talks to the bridge")

	t.Run("the OpenAI pair goes in together when neither is set", func(t *testing.T) {
		spec := &v1.PodSpec{Containers: []v1.Container{{Name: "primary"}}}
		require.NoError(t, ApplyAppleSiliconBridge(spec, "primary", "abc-123"))
		env := envMap(spec.Containers[0])
		assert.Equal(t, "http://192.168.65.254:30090/v1", env[OpenAIBaseURLEnvVar])
		assert.Equal(t, token, env[OpenAIAPIKeyEnvVar])
	})

	t.Run("a user's own OpenAI key is left alone, and no bridge URL is added next to it", func(t *testing.T) {
		spec := &v1.PodSpec{Containers: []v1.Container{{Name: "primary", Env: []v1.EnvVar{{Name: OpenAIAPIKeyEnvVar, Value: "sk-user"}}}}}
		require.NoError(t, ApplyAppleSiliconBridge(spec, "primary", "abc-123"))
		env := envMap(spec.Containers[0])
		assert.Equal(t, "sk-user", env[OpenAIAPIKeyEnvVar])
		assert.NotContains(t, env, OpenAIBaseURLEnvVar)
	})

	assert.Error(t, ApplyAppleSiliconBridge(&v1.PodSpec{Containers: []v1.Container{{Name: "x"}}}, "primary", "w"))
}

func TestSanitizeGPUResourceRequirementsAppleSilicon(t *testing.T) {
	setK8sConfig(t, &config.K8sPluginConfig{GpuResourceName: ResourceNvidiaGPU})
	resources := v1.ResourceRequirements{
		Requests: v1.ResourceList{resourceGPU: resource.MustParse("1")},
		Limits:   v1.ResourceList{resourceGPU: resource.MustParse("1")},
	}
	SanitizeGPUResourceRequirements(&resources, appleSiliconAccelerator)
	assert.Empty(t, resources.Requests)
	assert.Empty(t, resources.Limits)
}

func TestToK8sPodSpecAppleSilicon(t *testing.T) {
	taskTemplate := dummyTaskTemplate()
	taskTemplate.ExtendedResources = &core.ExtendedResources{GpuAccelerator: appleSiliconAccelerator}
	limits := &v1.ResourceRequirements{Limits: v1.ResourceList{ResourceNvidiaGPU: resource.MustParse("1")}}

	t.Run("bridge configured", func(t *testing.T) {
		setK8sConfig(t, &config.K8sPluginConfig{
			GpuResourceName:    ResourceNvidiaGPU,
			GpuDeviceNodeLabel: "gpu-node-label",
			AppleSiliconBridge: config.AppleSiliconBridgeConfig{
				Endpoint: "http://192.168.65.254:30090", TokenSecret: "s3cret",
			},
		})
		p, _, _, err := ToK8sPodSpec(context.TODO(), dummyExecContext(taskTemplate, limits, nil, "", nil))
		require.NoError(t, err)
		// No device plugin backs Apple Silicon: no GPU resource, no GPU node affinity or tolerations.
		assert.NotContains(t, p.Containers[0].Resources.Limits, ResourceNvidiaGPU)
		assert.NotContains(t, p.Containers[0].Resources.Requests, ResourceNvidiaGPU)
		if p.Affinity != nil {
			assert.Nil(t, p.Affinity.NodeAffinity)
		}
		assert.Empty(t, p.Tolerations)
		env := envMap(p.Containers[0])
		assert.Equal(t, "http://192.168.65.254:30090/v1", env[OpenAIBaseURLEnvVar])
		assert.Equal(t, MintAppleSiliconToken("s3cret", "some-acceptable-name"), env[AppleSiliconTokenEnvVar])
	})

	t.Run("no bridge fails with APPLE_SILICON_NOT_AVAILABLE", func(t *testing.T) {
		setK8sConfig(t, &config.K8sPluginConfig{GpuResourceName: ResourceNvidiaGPU})
		_, _, _, err := ToK8sPodSpec(context.TODO(), dummyExecContext(taskTemplate, limits, nil, "", nil))
		require.Error(t, err)
		assert.True(t, stdErrors.IsCausedBy(err, pluginserrors.AppleSiliconNotAvailable))
	})
}
