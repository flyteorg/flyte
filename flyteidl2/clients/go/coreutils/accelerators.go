package coreutils

import (
	"fmt"
	"strings"
	"sync"

	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

// acceleratorIndex is the AcceleratorModel enum read once from its descriptor:
// each value's accelerator_name option, and the reverse lookup by name. The
// enum is the one list of devices; nothing here is spelled twice.
type acceleratorIndex struct {
	names  map[core.AcceleratorModel]string
	models map[string]core.AcceleratorModel
	order  []core.AcceleratorModel // enum order, UNSPECIFIED excluded
}

var (
	acceleratorsOnce sync.Once
	accelerators     acceleratorIndex
)

func loadAccelerators() acceleratorIndex {
	acceleratorsOnce.Do(func() {
		idx := acceleratorIndex{
			names:  map[core.AcceleratorModel]string{},
			models: map[string]core.AcceleratorModel{},
		}
		values := core.AcceleratorModel(0).Descriptor().Values()
		for i := 0; i < values.Len(); i++ {
			v := values.Get(i)
			m := core.AcceleratorModel(v.Number())
			if m == core.AcceleratorModel_ACCELERATOR_MODEL_UNSPECIFIED {
				continue
			}
			name, _ := proto.GetExtension(v.Options(), core.E_AcceleratorName).(string)
			if name == "" {
				continue
			}
			idx.names[m] = name
			idx.models[name] = m
			idx.order = append(idx.order, m)
		}
		accelerators = idx
	})
	return accelerators
}

// AcceleratorModelName returns the accelerator_name of m: the string a task
// writes into GPUAccelerator.device for that model. Empty for UNSPECIFIED and
// for a number the enum does not define.
func AcceleratorModelName(m core.AcceleratorModel) string {
	return loadAccelerators().names[m]
}

// AcceleratorModelForDevice returns the model whose accelerator_name is
// device, matched exactly. false when no model has that name.
func AcceleratorModelForDevice(device string) (core.AcceleratorModel, bool) {
	m, ok := loadAccelerators().models[device]
	return m, ok
}

// AcceleratorModelClass returns the device class of m's block: the enum is
// numbered in blocks of 100 per GPUAccelerator.DeviceClass.
func AcceleratorModelClass(m core.AcceleratorModel) (core.GPUAccelerator_DeviceClass, bool) {
	switch int32(m) / 100 {
	case 0:
		return core.GPUAccelerator_NVIDIA_GPU, m != core.AcceleratorModel_ACCELERATOR_MODEL_UNSPECIFIED
	case 1:
		return core.GPUAccelerator_GOOGLE_TPU, true
	case 2:
		return core.GPUAccelerator_AMAZON_NEURON, true
	case 3:
		return core.GPUAccelerator_AMD_GPU, true
	case 4:
		return core.GPUAccelerator_HABANA_GAUDI, true
	default:
		return core.GPUAccelerator_NVIDIA_GPU, false
	}
}

// AcceleratorDeviceNames lists every accelerator_name in enum order, or only
// those of class when classes are given: what a user-facing message offers
// after rejecting an unknown device.
func AcceleratorDeviceNames(classes ...core.GPUAccelerator_DeviceClass) []string {
	idx := loadAccelerators()
	out := make([]string, 0, len(idx.order))
	for _, m := range idx.order {
		if len(classes) > 0 {
			c, _ := AcceleratorModelClass(m)
			found := false
			for _, want := range classes {
				if c == want {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		out = append(out, idx.names[m])
	}
	return out
}

// ValidateAccelerator reports whether acc names a device the AcceleratorModel
// list knows, of the class acc declares. The partitioning is not checked: an
// unset partition_size_value reads as unpartitioned everywhere the message is
// consumed. Used wherever an accelerator is accepted from a user rather than
// a task template — the task_resource.default_accelerator setting today.
func ValidateAccelerator(acc *core.GPUAccelerator) error {
	if acc == nil {
		return fmt.Errorf("no accelerator")
	}
	device := acc.GetDevice()
	if device == "" {
		return fmt.Errorf("device is required; one of %s", strings.Join(AcceleratorDeviceNames(acc.GetDeviceClass()), ", "))
	}
	m, ok := AcceleratorModelForDevice(device)
	if !ok {
		return fmt.Errorf("unknown device %q for class %s; one of %s",
			device, acc.GetDeviceClass(), strings.Join(AcceleratorDeviceNames(acc.GetDeviceClass()), ", "))
	}
	if class, _ := AcceleratorModelClass(m); class != acc.GetDeviceClass() {
		return fmt.Errorf("device %q is a %s, not a %s", device, class, acc.GetDeviceClass())
	}
	return nil
}
