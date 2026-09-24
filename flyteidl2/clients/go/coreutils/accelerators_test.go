package coreutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

// Every model carries a name, every name maps back to its model, and the
// names are the strings the SDK writes into GPUAccelerator.device.
func TestAcceleratorModelNames_RoundTrip(t *testing.T) {
	assert.Equal(t, "nvidia-tesla-t4", AcceleratorModelName(core.AcceleratorModel_NVIDIA_T4))
	assert.Equal(t, "nvidia-a100-80gb", AcceleratorModelName(core.AcceleratorModel_NVIDIA_A100_80GB))
	assert.Equal(t, "tpu-v5p-slice", AcceleratorModelName(core.AcceleratorModel_GOOGLE_TPU_V5P))
	assert.Empty(t, AcceleratorModelName(core.AcceleratorModel_ACCELERATOR_MODEL_UNSPECIFIED))
	assert.Empty(t, AcceleratorModelName(core.AcceleratorModel(99999)))

	values := core.AcceleratorModel(0).Descriptor().Values()
	seen := map[string]core.AcceleratorModel{}
	for i := 0; i < values.Len(); i++ {
		m := core.AcceleratorModel(values.Get(i).Number())
		if m == core.AcceleratorModel_ACCELERATOR_MODEL_UNSPECIFIED {
			continue
		}
		name := AcceleratorModelName(m)
		require.NotEmptyf(t, name, "%s has no accelerator_name", m)
		_, dup := seen[name]
		require.Falsef(t, dup, "accelerator_name %q is used twice (%s, %s)", name, seen[name], m)
		seen[name] = m
		back, ok := AcceleratorModelForDevice(name)
		require.True(t, ok)
		assert.Equal(t, m, back)
		_, known := AcceleratorModelClass(m)
		assert.Truef(t, known, "%s (%d) is outside every device-class block", m, m)
	}
	assert.Len(t, AcceleratorDeviceNames(), len(seen))
}

func TestAcceleratorModelClass(t *testing.T) {
	for m, want := range map[core.AcceleratorModel]core.GPUAccelerator_DeviceClass{
		core.AcceleratorModel_NVIDIA_T4:      core.GPUAccelerator_NVIDIA_GPU,
		core.AcceleratorModel_GOOGLE_TPU_V5E: core.GPUAccelerator_GOOGLE_TPU,
		core.AcceleratorModel_AMAZON_INF2:    core.GPUAccelerator_AMAZON_NEURON,
		core.AcceleratorModel_AMD_MI300X:     core.GPUAccelerator_AMD_GPU,
		core.AcceleratorModel_HABANA_GAUDI1:  core.GPUAccelerator_HABANA_GAUDI,
	} {
		got, ok := AcceleratorModelClass(m)
		assert.True(t, ok, m)
		assert.Equal(t, want, got, m)
	}
	_, ok := AcceleratorModelClass(core.AcceleratorModel_ACCELERATOR_MODEL_UNSPECIFIED)
	assert.False(t, ok)
	_, ok = AcceleratorModelClass(core.AcceleratorModel(700))
	assert.False(t, ok, "a block no class owns")

	assert.Equal(t, []string{"tpu-v5-lite-podslice", "tpu-v5p-slice", "tpu-v6e-slice"},
		AcceleratorDeviceNames(core.GPUAccelerator_GOOGLE_TPU))
	assert.Contains(t, AcceleratorDeviceNames(core.GPUAccelerator_NVIDIA_GPU), "nvidia-tesla-a100")
	assert.NotContains(t, AcceleratorDeviceNames(core.GPUAccelerator_NVIDIA_GPU), "tpu-v5p-slice")
}

func TestValidateAccelerator(t *testing.T) {
	assert.NoError(t, ValidateAccelerator(&core.GPUAccelerator{Device: "nvidia-tesla-t4"}), "NVIDIA_GPU is the default class")
	assert.NoError(t, ValidateAccelerator(&core.GPUAccelerator{
		DeviceClass: core.GPUAccelerator_GOOGLE_TPU, Device: "tpu-v5p-slice",
		PartitionSizeValue: &core.GPUAccelerator_PartitionSize{PartitionSize: "2x2"},
	}))
	assert.NoError(t, ValidateAccelerator(&core.GPUAccelerator{
		Device: "nvidia-tesla-a100", PartitionSizeValue: &core.GPUAccelerator_Unpartitioned{Unpartitioned: true},
	}))

	err := ValidateAccelerator(nil)
	require.Error(t, err)

	err = ValidateAccelerator(&core.GPUAccelerator{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "device is required")
	assert.Contains(t, err.Error(), "nvidia-tesla-t4", "the message offers the class's devices")

	err = ValidateAccelerator(&core.GPUAccelerator{Device: "T4"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown device "T4"`, "SDK short names are not the wire spelling")
	assert.Contains(t, err.Error(), "nvidia-tesla-t4")
	assert.NotContains(t, err.Error(), "tpu-v5p-slice", "only the requested class is offered")

	err = ValidateAccelerator(&core.GPUAccelerator{Device: "tpu-v5p-slice"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "is a GOOGLE_TPU, not a NVIDIA_GPU")
}
