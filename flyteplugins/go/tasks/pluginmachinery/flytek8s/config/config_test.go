package config

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testNvidiaClassLower = "nvidia_gpu"

func saveAndRestoreK8sConfig(t *testing.T) {
	t.Helper()
	prev := K8sPluginConfigSection.GetConfig().(*K8sPluginConfig)
	t.Cleanup(func() {
		require.NoError(t, SetK8sPluginConfig(prev))
	})
}

func TestGetK8sPluginConfig(t *testing.T) {
	assert.Equal(t, defaultCPURequest, GetK8sPluginConfig().DefaultCPURequest)
	assert.Equal(t, defaultMemoryRequest, GetK8sPluginConfig().DefaultMemoryRequest)
}

func TestGetK8sPluginConfigNormalizationAndUpdates(t *testing.T) {
	saveAndRestoreK8sConfig(t)

	t.Run("upper-cases device classes and merges accelerator devices over defaults", func(t *testing.T) {
		require.NoError(t, SetK8sPluginConfig(&K8sPluginConfig{
			GpuDeviceNodeLabel: "first-label",
			AcceleratorDeviceClasses: map[string]AcceleratorDeviceClassConfig{
				testNvidiaClassLower: {ResourceName: ResourceNvidiaGPU},
				"Google_TPU":         {ResourceName: "google.com/tpu"},
			},
			AcceleratorDevices: map[string]string{
				"t4":       "custom-t4",
				"custom_x": "custom-device-x",
			},
		}))

		cfg := GetK8sPluginConfig()
		assert.Equal(t, "first-label", cfg.GpuDeviceNodeLabel)

		// Device class keys are upper-cased.
		assert.Contains(t, cfg.AcceleratorDeviceClasses, "NVIDIA_GPU")
		assert.Contains(t, cfg.AcceleratorDeviceClasses, "GOOGLE_TPU")
		assert.NotContains(t, cfg.AcceleratorDeviceClasses, testNvidiaClassLower)
		assert.NotContains(t, cfg.AcceleratorDeviceClasses, "Google_TPU")
		assert.Equal(t, ResourceNvidiaGPU, cfg.AcceleratorDeviceClasses["NVIDIA_GPU"].ResourceName)

		// Configured accelerator devices are upper-cased and override defaults while preserving other defaults.
		assert.Equal(t, "custom-t4", cfg.AcceleratorDevices["T4"])
		assert.Equal(t, "custom-device-x", cfg.AcceleratorDevices["CUSTOM_X"])
		assert.NotContains(t, cfg.AcceleratorDevices, "t4")
		assert.NotContains(t, cfg.AcceleratorDevices, "custom_x")
		assert.Equal(t, "nvidia-tesla-a100", cfg.AcceleratorDevices["A100"])
		assert.Equal(t, "nvidia-l40s", cfg.AcceleratorDevices["L40S"])
	})

	t.Run("picks up updates after SetK8sPluginConfig", func(t *testing.T) {
		require.NoError(t, SetK8sPluginConfig(&K8sPluginConfig{
			GpuDeviceNodeLabel: "updated-label",
			AcceleratorDeviceClasses: map[string]AcceleratorDeviceClassConfig{
				"amd_gpu": {ResourceName: "amd.com/gpu"},
			},
			AcceleratorDevices: map[string]string{
				"a100": "custom-a100",
			},
		}))

		cfg := GetK8sPluginConfig()
		assert.Equal(t, "updated-label", cfg.GpuDeviceNodeLabel)
		assert.Contains(t, cfg.AcceleratorDeviceClasses, "AMD_GPU")
		assert.NotContains(t, cfg.AcceleratorDeviceClasses, "NVIDIA_GPU")
		assert.Equal(t, "custom-a100", cfg.AcceleratorDevices["A100"])
		// Prior test override for T4 is gone and default is back.
		assert.Equal(t, "nvidia-tesla-t4", cfg.AcceleratorDevices["T4"])
	})

	t.Run("does not mutate stored config so identical SetConfig is a no-op", func(t *testing.T) {
		makeRaw := func() *K8sPluginConfig {
			return &K8sPluginConfig{
				GpuDeviceNodeLabel: "stable-label",
				AcceleratorDeviceClasses: map[string]AcceleratorDeviceClassConfig{
					testNvidiaClassLower: {ResourceName: ResourceNvidiaGPU},
				},
				AcceleratorDevices: map[string]string{
					"t4": "custom-t4",
				},
			}
		}

		require.NoError(t, SetK8sPluginConfig(makeRaw()))
		_ = K8sPluginConfigSection.GetConfigChangedAndClear()

		cfg := GetK8sPluginConfig()
		assert.Equal(t, "custom-t4", cfg.AcceleratorDevices["T4"])
		assert.Contains(t, cfg.AcceleratorDeviceClasses, "NVIDIA_GPU")

		// Re-applying an identical raw config should compare equal and not mark the section dirty.
		require.NoError(t, SetK8sPluginConfig(makeRaw()))
		assert.False(t, K8sPluginConfigSection.GetConfigChangedAndClear())
	})
}

func TestGetK8sPluginConfigConcurrent(t *testing.T) {
	saveAndRestoreK8sConfig(t)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	const (
		numReaders       = 16
		writerIterations = 200
	)

	var wg sync.WaitGroup
	for range numReaders {
		wg.Go(func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
				}
				cfg := GetK8sPluginConfig()
				if _, ok := cfg.AcceleratorDevices["A100"]; !ok {
					t.Errorf("missing default accelerator device A100")
					return
				}
				for k := range cfg.AcceleratorDeviceClasses {
					if k != strings.ToUpper(k) {
						t.Errorf("unexpected unnormalized device class key %q", k)
						return
					}
				}
			}
		})
	}

	for i := range writerIterations {
		var next *K8sPluginConfig
		if i%2 == 0 {
			next = &K8sPluginConfig{
				GpuDeviceNodeLabel: "label-even",
				AcceleratorDeviceClasses: map[string]AcceleratorDeviceClassConfig{
					testNvidiaClassLower: {ResourceName: ResourceNvidiaGPU},
				},
				AcceleratorDevices: map[string]string{
					"t4": "even-t4",
				},
			}
		} else {
			next = &K8sPluginConfig{
				GpuDeviceNodeLabel: "label-odd",
				AcceleratorDeviceClasses: map[string]AcceleratorDeviceClassConfig{
					"amd_gpu": {ResourceName: "amd.com/gpu"},
				},
				AcceleratorDevices: map[string]string{
					"t4": "odd-t4",
				},
			}
		}
		require.NoError(t, SetK8sPluginConfig(next))
	}

	cancel()
	wg.Wait()
}
