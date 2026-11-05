package service

import (
	"maps"
	"slices"

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

	// Map settings merge per key: the request's entry wins on a conflict, and the
	// settings only add keys the request does not set.
	if entries := valueEntries(resolved.GetLabels()); len(entries) > 0 {
		if spec.Labels == nil {
			spec.Labels = &task.Labels{}
		}
		spec.Labels.Values = fillMissing(spec.Labels.Values, entries)
	}

	if entries := valueEntries(resolved.GetAnnotations()); len(entries) > 0 {
		if spec.Annotations == nil {
			spec.Annotations = &task.Annotations{}
		}
		spec.Annotations.Values = fillMissing(spec.Annotations.Values, entries)
	}

	// Envs is a list, so the request's entries keep their order and settings entries
	// are appended in key order, keeping the stored spec deterministic.
	if entries := valueEntries(resolved.GetEnvironmentVariables()); len(entries) > 0 {
		if spec.Envs == nil {
			spec.Envs = &task.Envs{}
		}
		present := make(map[string]struct{}, len(spec.Envs.Values))
		for _, kv := range spec.Envs.Values {
			present[kv.GetKey()] = struct{}{}
		}
		for _, k := range slices.Sorted(maps.Keys(entries)) {
			if _, ok := present[k]; !ok {
				spec.Envs.Values = append(spec.Envs.Values, &core.KeyValuePair{Key: k, Value: entries[k]})
			}
		}
	}
}

// valueEntries returns the entries of a map setting, or nil unless it is in VALUE
// state. Resolution has already merged the scopes, so UNSET and INHERIT add nothing.
func valueEntries(s *settings.StringMapSetting) map[string]string {
	if s.GetState() != settings.SettingState_SETTING_STATE_VALUE {
		return nil
	}
	return s.GetMapValue().GetEntries()
}

// fillMissing copies src entries into dst for keys dst does not already have,
// allocating dst if needed.
func fillMissing(dst, src map[string]string) map[string]string {
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
