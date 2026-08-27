package service

import (
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/settings"
)

// scopeLevelAt maps a position in the level chain to the scope it represents.
// Callers build the chain broadest first: org, domain, project.
func scopeLevelAt(i int) settings.ScopeLevel {
	switch i {
	case 0:
		return settings.ScopeLevel_SCOPE_LEVEL_ORG
	case 1:
		return settings.ScopeLevel_SCOPE_LEVEL_DOMAIN
	default:
		return settings.ScopeLevel_SCOPE_LEVEL_PROJECT
	}
}

// scalarSetting is the shared shape of the scalar leaf wrappers: a proto message
// carrying a SettingState.
type scalarSetting interface {
	proto.Message
	GetState() settings.SettingState
}

// mergeScalar resolves one scalar leaf across the level chain: the last level whose
// state is not INHERIT wins, so an UNSET child blocks a value set above it, and nil
// comes back when nothing did. setLevel stamps the winning scope onto the copy,
// because a generic constraint can require methods but not fields.
func mergeScalar[T scalarSetting](levels []T, setLevel func(T, settings.ScopeLevel)) T {
	var out T
	won := -1

	for i, s := range levels {
		if s.GetState() == settings.SettingState_SETTING_STATE_INHERIT {
			continue
		}
		won = i
	}

	if won < 0 {
		return out
	}

	out = proto.Clone(levels[won]).(T)
	setLevel(out, scopeLevelAt(won))
	return out
}

// mergeStringSettings resolves one string leaf across the level chain. The last
// level whose state is not INHERIT wins, so UNSET at a child level blocks a
// value set higher up. Returns nil when no level had an opinion.
func mergeStringSettings(levels []*settings.StringSetting) *settings.StringSetting {
	return mergeScalar(levels, func(s *settings.StringSetting, level settings.ScopeLevel) {
		s.ScopeLevel = level
	})
}

// mergeInt64Settings resolves one int64 leaf across the level chain, following the
// same rule as mergeStringSettings.
func mergeInt64Settings(levels []*settings.Int64Setting) *settings.Int64Setting {
	return mergeScalar(levels, func(s *settings.Int64Setting, level settings.ScopeLevel) {
		s.ScopeLevel = level
	})
}

// mergeBoolSettings resolves one bool leaf across the level chain, following the
// same rule as mergeStringSettings.
func mergeBoolSettings(levels []*settings.BoolSetting) *settings.BoolSetting {
	return mergeScalar(levels, func(s *settings.BoolSetting, level settings.ScopeLevel) {
		s.ScopeLevel = level
	})
}

// mergeQuantitySettings resolves one quantity leaf across the level chain, following
// the same rule as mergeStringSettings.
func mergeQuantitySettings(levels []*settings.QuantitySetting) *settings.QuantitySetting {
	return mergeScalar(levels, func(s *settings.QuantitySetting, level settings.ScopeLevel) {
		s.ScopeLevel = level
	})
}

// mergeStringMapSettings resolves one string-map leaf across the level chain.
// Unlike scalars, every level contributes: entries accumulate parent first with
// child entries overwriting on key conflict, and a level in state UNSET clears
// everything accumulated above it. ScopeLevel records the most specific level
// that contributed, since entries may come from several levels.
func mergeStringMapSettings(levels []*settings.StringMapSetting) *settings.StringMapSetting {
	entries := map[string]string{}
	var state settings.SettingState
	var level settings.ScopeLevel

	for i, s := range levels {
		switch s.GetState() {
		case settings.SettingState_SETTING_STATE_INHERIT:
			continue
		case settings.SettingState_SETTING_STATE_UNSET:
			entries = map[string]string{}
		default:
			for k, v := range s.GetMapValue().GetEntries() {
				entries[k] = v
			}
		}
		state = s.GetState()
		level = scopeLevelAt(i)
	}

	// state is still its zero value when no level contributed.
	if state == settings.SettingState_SETTING_STATE_INHERIT {
		return nil
	}
	if state == settings.SettingState_SETTING_STATE_UNSET {
		return &settings.StringMapSetting{
			State:      settings.SettingState_SETTING_STATE_UNSET,
			ScopeLevel: level,
		}
	}
	return &settings.StringMapSetting{
		State:      settings.SettingState_SETTING_STATE_VALUE,
		MapValue:   &settings.StringMap{Entries: entries},
		ScopeLevel: level,
	}
}

// mergeTaskResourceDefaults resolves one resource bound (min or max) across the level
// chain. It holds no merge rules of its own: it regroups the four dimensions by level
// and delegates each to mergeQuantitySettings. Returns nil when no dimension resolved.
func mergeTaskResourceDefaults(levels []*settings.TaskResourceDefaults) *settings.TaskResourceDefaults {
	cpu := make([]*settings.QuantitySetting, len(levels))
	gpu := make([]*settings.QuantitySetting, len(levels))
	memory := make([]*settings.QuantitySetting, len(levels))
	storage := make([]*settings.QuantitySetting, len(levels))

	for i, l := range levels {
		cpu[i] = l.GetCpu()
		gpu[i] = l.GetGpu()
		memory[i] = l.GetMemory()
		storage[i] = l.GetStorage()
	}

	out := &settings.TaskResourceDefaults{
		Cpu:     mergeQuantitySettings(cpu),
		Gpu:     mergeQuantitySettings(gpu),
		Memory:  mergeQuantitySettings(memory),
		Storage: mergeQuantitySettings(storage),
	}

	if out.Cpu == nil && out.Gpu == nil && out.Memory == nil && out.Storage == nil {
		return nil
	}
	return out
}

// mergeTaskResourceSettings resolves the task_resource group across the level chain.
// It holds no merge rules of its own: it regroups min, max, and mirror_limits_request
// by level and delegates each. Returns nil when no level set anything under
// task_resource, so the merged output carries no empty group.
func mergeTaskResourceSettings(levels []*settings.TaskResourceSettings) *settings.TaskResourceSettings {
	minLevels := make([]*settings.TaskResourceDefaults, len(levels))
	maxLevels := make([]*settings.TaskResourceDefaults, len(levels))
	mirror := make([]*settings.BoolSetting, len(levels))

	for i, l := range levels {
		minLevels[i] = l.GetMin()
		maxLevels[i] = l.GetMax()
		mirror[i] = l.GetMirrorLimitsRequest()
	}

	out := &settings.TaskResourceSettings{
		Min:                 mergeTaskResourceDefaults(minLevels),
		Max:                 mergeTaskResourceDefaults(maxLevels),
		MirrorLimitsRequest: mergeBoolSettings(mirror),
	}

	if out.Min == nil && out.Max == nil && out.MirrorLimitsRequest == nil {
		return nil
	}
	return out
}

// mergeRunSettings resolves the run group across the level chain. Like the other
// group mergers it holds no rules of its own, only regrouping and delegation.
// Returns nil when no level set anything under run.
func mergeRunSettings(levels []*settings.RunSettings) *settings.RunSettings {
	defaultQueue := make([]*settings.StringSetting, len(levels))
	maxActionConcurrency := make([]*settings.Int64Setting, len(levels))
	runBaseDir := make([]*settings.StringSetting, len(levels))

	for i, l := range levels {
		defaultQueue[i] = l.GetDefaultQueue()
		maxActionConcurrency[i] = l.GetMaxActionConcurrency()
		runBaseDir[i] = l.GetRunBaseDir()
	}

	out := &settings.RunSettings{
		DefaultQueue:         mergeStringSettings(defaultQueue),
		MaxActionConcurrency: mergeInt64Settings(maxActionConcurrency),
		RunBaseDir:           mergeStringSettings(runBaseDir),
	}

	if out.DefaultQueue == nil && out.MaxActionConcurrency == nil && out.RunBaseDir == nil {
		return nil
	}
	return out
}

// mergeSecuritySettings resolves the security group across the level chain. Like the
// other group mergers it holds no rules of its own. Returns nil when nothing resolved.
func mergeSecuritySettings(levels []*settings.SecuritySettings) *settings.SecuritySettings {
	serviceAccount := make([]*settings.StringSetting, len(levels))

	for i, l := range levels {
		serviceAccount[i] = l.GetServiceAccount()
	}

	out := &settings.SecuritySettings{
		ServiceAccount: mergeStringSettings(serviceAccount),
	}

	if out.ServiceAccount == nil {
		return nil
	}
	return out
}

// mergeStorageSettings resolves the storage group across the level chain. Like the
// other group mergers it holds no rules of its own. Returns nil when nothing resolved.
func mergeStorageSettings(levels []*settings.StorageSettings) *settings.StorageSettings {
	rawDataPath := make([]*settings.StringSetting, len(levels))

	for i, l := range levels {
		rawDataPath[i] = l.GetRawDataPath()
	}

	out := &settings.StorageSettings{
		RawDataPath: mergeStringSettings(rawDataPath),
	}

	if out.RawDataPath == nil {
		return nil
	}
	return out
}

// mergeAppSettings resolves the app group across the level chain. Like the other
// group mergers it holds no rules of its own. Returns nil when nothing resolved.
func mergeAppSettings(levels []*settings.AppSettings) *settings.AppSettings {
	disallowAnonymous := make([]*settings.BoolSetting, len(levels))

	for i, l := range levels {
		disallowAnonymous[i] = l.GetDisallowAnonymous()
	}

	out := &settings.AppSettings{
		DisallowAnonymous: mergeBoolSettings(disallowAnonymous),
	}

	if out.DisallowAnonymous == nil {
		return nil
	}
	return out
}

// mergeSettings resolves a full settings document across the level chain. Callers
// pass one entry per level, broadest first, using nil for levels with no stored
// record.
func mergeSettings(levels []*settings.Settings) *settings.Settings {
	run := make([]*settings.RunSettings, len(levels))
	security := make([]*settings.SecuritySettings, len(levels))
	storage := make([]*settings.StorageSettings, len(levels))
	taskResource := make([]*settings.TaskResourceSettings, len(levels))
	labels := make([]*settings.StringMapSetting, len(levels))
	annotations := make([]*settings.StringMapSetting, len(levels))
	envVars := make([]*settings.StringMapSetting, len(levels))
	app := make([]*settings.AppSettings, len(levels))
	podTemplateName := make([]*settings.StringSetting, len(levels))

	for i, l := range levels {
		run[i] = l.GetRun()
		security[i] = l.GetSecurity()
		storage[i] = l.GetStorage()
		taskResource[i] = l.GetTaskResource()
		labels[i] = l.GetLabels()
		annotations[i] = l.GetAnnotations()
		envVars[i] = l.GetEnvironmentVariables()
		app[i] = l.GetApp()
		podTemplateName[i] = l.GetPodTemplateName()
	}

	return &settings.Settings{
		Run:                  mergeRunSettings(run),
		Security:             mergeSecuritySettings(security),
		Storage:              mergeStorageSettings(storage),
		TaskResource:         mergeTaskResourceSettings(taskResource),
		Labels:               mergeStringMapSettings(labels),
		Annotations:          mergeStringMapSettings(annotations),
		EnvironmentVariables: mergeStringMapSettings(envVars),
		App:                  mergeAppSettings(app),
		PodTemplateName:      mergeStringSettings(podTemplateName),
	}
}
