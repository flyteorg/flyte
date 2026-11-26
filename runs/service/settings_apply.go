package service

import (
	"maps"
	"slices"

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

	// Projected onto tasks that set no service account of their own, at action creation.
	// Only the service account is filled; other security context fields are kept.
	if serviceAccount := resolved.GetSecurity().GetServiceAccount(); spec.GetSecurityContext().GetRunAs().GetK8SServiceAccount() == "" &&
		serviceAccount.GetState() == settings.SettingState_SETTING_STATE_VALUE {
		if spec.SecurityContext == nil {
			spec.SecurityContext = &core.SecurityContext{}
		}
		if spec.SecurityContext.RunAs == nil {
			spec.SecurityContext.RunAs = &core.Identity{}
		}
		spec.SecurityContext.RunAs.K8SServiceAccount = serviceAccount.GetStringValue()
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

	// Map settings merge per key: settings add only the keys the request does not set.
	if labels := mapSettingValues(resolved.GetLabels()); len(labels) > 0 {
		if spec.Labels == nil {
			spec.Labels = &task.Labels{}
		}
		spec.Labels.Values = addMissing(spec.Labels.Values, labels)
	}
	if annotations := mapSettingValues(resolved.GetAnnotations()); len(annotations) > 0 {
		if spec.Annotations == nil {
			spec.Annotations = &task.Annotations{}
		}
		spec.Annotations.Values = addMissing(spec.Annotations.Values, annotations)
	}
	applyEnvSettings(spec, mapSettingValues(resolved.GetEnvironmentVariables()))

	if defaults := newTaskResourceDefaults(resolved.GetTaskResource()); defaults != nil {
		if spec.DefaultSettings == nil {
			spec.DefaultSettings = &task.DefaultSettings{}
		}
		spec.DefaultSettings.TaskResourceDefaults = defaults
	}
}

// mapSettingValues returns the entries of a map setting in VALUE state, else nil.
func mapSettingValues(s *settings.StringMapSetting) map[string]string {
	if s.GetState() != settings.SettingState_SETTING_STATE_VALUE {
		return nil
	}
	return s.GetMapValue().GetEntries()
}

// addMissing copies the entries of src whose key dst lacks, allocating dst if needed.
func addMissing(dst, src map[string]string) map[string]string {
	if dst == nil {
		dst = make(map[string]string, len(src))
	}
	for k, v := range src {
		if _, ok := dst[k]; !ok {
			dst[k] = v
		}
	}
	return dst
}

// applyEnvSettings appends settings env vars whose key the request does not set. The
// request's entries keep their order; settings entries follow in key order so the
// stored spec is deterministic.
func applyEnvSettings(spec *task.RunSpec, envs map[string]string) {
	if len(envs) == 0 {
		return
	}
	specEnvs := spec.GetEnvs().GetValues()
	present := make(map[string]struct{}, len(specEnvs))
	for _, kv := range specEnvs {
		present[kv.GetKey()] = struct{}{}
	}
	for _, k := range slices.Sorted(maps.Keys(envs)) {
		if _, ok := present[k]; ok {
			continue
		}
		if spec.Envs == nil {
			spec.Envs = &task.Envs{}
		}
		spec.Envs.Values = append(spec.Envs.Values, &core.KeyValuePair{Key: k, Value: envs[k]})
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
