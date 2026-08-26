package service

import (
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/settings"
)

// scopeLevelAt maps a position in the level chain to the scope it represents.
// Callers build the chain broadest first: instance, domain, project.
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

// mergeStringSettings resolves one string leaf across the level chain. The last
// level whose state is not INHERIT wins, so UNSET at a child level blocks a
// value set higher up. Returns nil when no level had an opinion.
func mergeStringSettings(levels []*settings.StringSetting) *settings.StringSetting {
	var winner *settings.StringSetting
	var level settings.ScopeLevel

	for i, s := range levels {
		if s.GetState() == settings.SettingState_SETTING_STATE_INHERIT {
			continue
		}
		winner = s
		level = scopeLevelAt(i)
	}
	if winner == nil {
		return nil
	}

	out := proto.Clone(winner).(*settings.StringSetting)
	out.ScopeLevel = level

	return out
}

// mergeInt64Settings resolves one int64 leaf across the level chain, following the
// same rule as mergeStringSettings.
func mergeInt64Settings(levels []*settings.Int64Setting) *settings.Int64Setting {
	var winner *settings.Int64Setting
	var level settings.ScopeLevel

	for i, s := range levels {
		if s.GetState() == settings.SettingState_SETTING_STATE_INHERIT {
			continue
		}
		winner = s
		level = scopeLevelAt(i)
	}
	if winner == nil {
		return nil
	}

	out := proto.Clone(winner).(*settings.Int64Setting)
	out.ScopeLevel = level

	return out
}

// mergeBoolSettings resolves one bool leaf across the level chain, following the
// same rule as mergeStringSettings.
func mergeBoolSettings(levels []*settings.BoolSetting) *settings.BoolSetting {
	var winner *settings.BoolSetting
	var level settings.ScopeLevel

	for i, s := range levels {
		if s.GetState() == settings.SettingState_SETTING_STATE_INHERIT {
			continue
		}
		winner = s
		level = scopeLevelAt(i)
	}
	if winner == nil {
		return nil
	}

	out := proto.Clone(winner).(*settings.BoolSetting)
	out.ScopeLevel = level

	return out
}

// mergeQuantitySettings resolves one quantity leaf across the level chain, following
// the same rule as mergeStringSettings.
func mergeQuantitySettings(levels []*settings.QuantitySetting) *settings.QuantitySetting {
	var winner *settings.QuantitySetting
	var level settings.ScopeLevel

	for i, s := range levels {
		if s.GetState() == settings.SettingState_SETTING_STATE_INHERIT {
			continue
		}
		winner = s
		level = scopeLevelAt(i)
	}
	if winner == nil {
		return nil
	}

	out := proto.Clone(winner).(*settings.QuantitySetting)
	out.ScopeLevel = level

	return out
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
