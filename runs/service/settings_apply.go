package service

import (
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/settings"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
)

// applyRunSettings fills fields the caller left unset with values resolved from
// settings. An explicit value in the request always wins, so this only ever fills
// gaps, and a setting that is INHERIT or UNSET contributes nothing.
func applyRunSettings(spec *task.RunSpec, resolved *settings.Settings) {
	if spec == nil {
		return
	}

	if spec.GetQueue() == "" && resolved.GetRun().GetDefaultQueue().GetState() ==
		settings.SettingState_SETTING_STATE_VALUE {
		spec.Queue = resolved.GetRun().GetDefaultQueue().GetStringValue()
	}

	// The proto defines 0 as unset for this field, so there is no explicit request for
	// "unlimited" to override. The settings validator bounds the value to 0 or
	// [2, MaxUint32], so narrowing to uint32 cannot overflow.
	if concurrency := resolved.GetRun().GetMaxActionConcurrency(); spec.GetMaxActionConcurrency() == 0 &&
		concurrency.GetState() == settings.SettingState_SETTING_STATE_VALUE {
		spec.MaxActionConcurrency = uint32(concurrency.GetIntValue())
	}

	// Projected onto tasks that name no template of their own, at action creation.
	if podTemplate := resolved.GetPodTemplateName(); spec.GetDefaultSettings().GetPodTemplateName() == "" &&
		podTemplate.GetState() == settings.SettingState_SETTING_STATE_VALUE {
		if spec.DefaultSettings == nil {
			spec.DefaultSettings = &task.DefaultSettings{}
		}
		spec.DefaultSettings.PodTemplateName = podTemplate.GetStringValue()
	}

	// Base for run metadata; also raw outputs when raw_data_prefix is empty.
	if baseDir := resolved.GetRun().GetRunBaseDir(); spec.GetRunBaseDir() == "" &&
		baseDir.GetState() == settings.SettingState_SETTING_STATE_VALUE {
		spec.RunBaseDir = baseDir.GetStringValue()
	}

	// RawDataStorage carries only this prefix, so replacing the message loses nothing.
	if rawDataPath := resolved.GetStorage().GetRawDataPath(); spec.GetRawDataStorage().GetRawDataPrefix() == "" &&
		rawDataPath.GetState() == settings.SettingState_SETTING_STATE_VALUE {
		spec.RawDataStorage = &task.RawDataStorage{RawDataPrefix: rawDataPath.GetStringValue()}
	}

	if defaults := newTaskResourceDefaults(resolved.GetTaskResource()); defaults != nil {
		if spec.DefaultSettings == nil {
			spec.DefaultSettings = &task.DefaultSettings{}
		}
		spec.DefaultSettings.TaskResourceDefaults = defaults
	}
}

func newTaskResourceDefaults(tr *settings.TaskResourceSettings) *task.TaskResourceDefaults {
	out := &task.TaskResourceDefaults{}

	if accelerator := tr.GetDefaultAccelerator(); accelerator.GetState() == settings.SettingState_SETTING_STATE_VALUE {
		out.DefaultAccelerator = accelerator.GetAcceleratorValue()
	}

	minRes, maxRes := tr.GetMin(), tr.GetMax()
	dimensions := []struct {
		name     core.Resources_ResourceName
		min, max *settings.QuantitySetting
	}{
		{core.Resources_CPU, minRes.GetCpu(), maxRes.GetCpu()},
		{core.Resources_GPU, minRes.GetGpu(), maxRes.GetGpu()},
		{core.Resources_MEMORY, minRes.GetMemory(), maxRes.GetMemory()},
		{core.Resources_EPHEMERAL_STORAGE, minRes.GetStorage(), maxRes.GetStorage()},
	}

	for _, d := range dimensions {
		maxValue := d.max.GetQuantityValue()
		hasMax := d.max.GetState() == settings.SettingState_SETTING_STATE_VALUE && maxValue != ""
		if hasMax {
			out.Max = append(out.Max, &core.Resources_ResourceEntry{Name: d.name, Value: maxValue})
		}
		if minValue := d.min.GetQuantityValue(); d.min.GetState() == settings.SettingState_SETTING_STATE_VALUE && minValue != "" {
			if hasMax {
				minQ, minErr := resource.ParseQuantity(minValue)
				maxQ, maxErr := resource.ParseQuantity(maxValue)
				if minErr == nil && maxErr == nil && minQ.Cmp(maxQ) > 0 {
					minValue = maxValue
				}
			}
			out.Requests = append(out.Requests, &core.Resources_ResourceEntry{Name: d.name, Value: minValue})
		}
	}

	if len(out.GetRequests()) == 0 && len(out.GetMax()) == 0 && out.GetDefaultAccelerator() == nil {
		return nil
	}
	return out
}
