package config

import (
	"testing"

	"gotest.tools/assert"
)

func TestGetK8sPluginConfig(t *testing.T) {
	assert.Equal(t, GetK8sPluginConfig().DefaultCPURequest, defaultCPURequest)
	assert.Equal(t, GetK8sPluginConfig().DefaultMemoryRequest, defaultMemoryRequest)
}

func TestAcceleratorDeviceAliases(t *testing.T) {
	// A canonical name: the spelling the SDK writes today (the default label,
	// upper-cased) and the short name that shares the label.
	assert.DeepEqual(t, AcceleratorDeviceAliases("NVIDIA-T4"), []string{"NVIDIA-TESLA-T4", "T4"})
	assert.DeepEqual(t, AcceleratorDeviceAliases("AWS-TRN1"), []string{"AWS-NEURON-TRN1", "TRN1"})
	// H100: the short-name row points elsewhere, so only the SDK spelling remains.
	assert.DeepEqual(t, AcceleratorDeviceAliases("nvidia-h100"), []string{"NVIDIA-TESLA-H100"})
	// The label equals the canonical name: only the short name is an alias.
	assert.DeepEqual(t, AcceleratorDeviceAliases("NVIDIA-H200"), []string{"H200"})
	// A legacy spelling gets the canonical one back.
	assert.DeepEqual(t, AcceleratorDeviceAliases("T4"), []string{"NVIDIA-T4", "NVIDIA-TESLA-T4"})
	assert.Assert(t, AcceleratorDeviceAliases("CUSTOM-DEVICE") == nil)

	// The table an override replaces is not the one the aliases come from.
	assert.NilError(t, SetK8sPluginConfig(&K8sPluginConfig{AcceleratorDevices: map[string]string{"X": "y"}}))
	assert.DeepEqual(t, AcceleratorDeviceAliases("NVIDIA-T4"), []string{"NVIDIA-TESLA-T4", "T4"})
}
