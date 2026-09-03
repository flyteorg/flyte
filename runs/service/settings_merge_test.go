package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/settings"
)

const (
	levelOrg     = settings.ScopeLevel_SCOPE_LEVEL_ORG
	levelDomain  = settings.ScopeLevel_SCOPE_LEVEL_DOMAIN
	levelProject = settings.ScopeLevel_SCOPE_LEVEL_PROJECT
)

func str(state settings.SettingState, value string) *settings.StringSetting {
	return &settings.StringSetting{State: state, StringValue: value}
}

func strMap(state settings.SettingState, entries map[string]string) *settings.StringMapSetting {
	m := &settings.StringMapSetting{State: state}
	if entries != nil {
		m.MapValue = &settings.StringMap{Entries: entries}
	}
	return m
}

// wantMap builds an expected merged map leaf: state VALUE plus a scope level.
func wantMap(level settings.ScopeLevel, entries map[string]string) *settings.StringMapSetting {
	return &settings.StringMapSetting{
		State:      stateValue,
		MapValue:   &settings.StringMap{Entries: entries},
		ScopeLevel: level,
	}
}

func TestMergeStringSettings(t *testing.T) {
	tests := []struct {
		name   string
		levels []*settings.StringSetting
		want   *settings.StringSetting
	}{
		{
			name:   "nothing set anywhere",
			levels: []*settings.StringSetting{nil, nil, nil},
			want:   nil,
		},
		{
			name:   "all inherit",
			levels: []*settings.StringSetting{str(stateInherit, ""), str(stateInherit, ""), str(stateInherit, "")},
			want:   nil,
		},
		{
			name:   "only org sets it",
			levels: []*settings.StringSetting{str(stateValue, "default"), nil, nil},
			want:   &settings.StringSetting{State: stateValue, StringValue: "default", ScopeLevel: levelOrg},
		},
		{
			name:   "project overrides org",
			levels: []*settings.StringSetting{str(stateValue, "default"), nil, str(stateValue, "gpu-pool")},
			want:   &settings.StringSetting{State: stateValue, StringValue: "gpu-pool", ScopeLevel: levelProject},
		},
		{
			name:   "domain wins when project inherits",
			levels: []*settings.StringSetting{str(stateValue, "default"), str(stateValue, "fast"), str(stateInherit, "")},
			want:   &settings.StringSetting{State: stateValue, StringValue: "fast", ScopeLevel: levelDomain},
		},
		{
			name:   "unset at project blocks the org value",
			levels: []*settings.StringSetting{str(stateValue, "default-runner"), nil, str(stateUnset, "")},
			want:   &settings.StringSetting{State: stateUnset, ScopeLevel: levelProject},
		},
		{
			name:   "unset at domain survives an inheriting project",
			levels: []*settings.StringSetting{str(stateValue, "default-runner"), str(stateUnset, ""), str(stateInherit, "")},
			want:   &settings.StringSetting{State: stateUnset, ScopeLevel: levelDomain},
		},
		{
			name:   "project overrides an unset domain",
			levels: []*settings.StringSetting{str(stateValue, "default"), str(stateUnset, ""), str(stateValue, "mine")},
			want:   &settings.StringSetting{State: stateValue, StringValue: "mine", ScopeLevel: levelProject},
		},
		{
			name:   "shorter chain for an org-scoped request",
			levels: []*settings.StringSetting{str(stateValue, "default")},
			want:   &settings.StringSetting{State: stateValue, StringValue: "default", ScopeLevel: levelOrg},
		},
		{
			name:   "unset at org with nothing below",
			levels: []*settings.StringSetting{str(stateUnset, ""), nil, nil},
			want:   &settings.StringSetting{State: stateUnset, ScopeLevel: levelOrg},
		},
		{
			name:   "domain is the only level that sets it",
			levels: []*settings.StringSetting{nil, str(stateValue, "fast"), nil},
			want:   &settings.StringSetting{State: stateValue, StringValue: "fast", ScopeLevel: levelDomain},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeStringSettings(tt.levels)
			assert.Truef(t, proto.Equal(tt.want, got), "want %v, got %v", tt.want, got)
		})
	}
}

func TestMergeStringMapSettings(t *testing.T) {
	tests := []struct {
		name   string
		levels []*settings.StringMapSetting
		want   *settings.StringMapSetting
	}{
		{
			name:   "nothing set anywhere",
			levels: []*settings.StringMapSetting{nil, nil, nil},
			want:   nil,
		},
		{
			name:   "only org contributes",
			levels: []*settings.StringMapSetting{strMap(stateValue, map[string]string{"A": "1"}), nil, nil},
			want:   wantMap(levelOrg, map[string]string{"A": "1"}),
		},
		{
			name: "levels accumulate when keys do not clash",
			levels: []*settings.StringMapSetting{
				strMap(stateValue, map[string]string{"A": "1"}),
				strMap(stateValue, map[string]string{"B": "2"}),
				nil,
			},
			want: wantMap(levelDomain, map[string]string{"A": "1", "B": "2"}),
		},
		{
			name: "child overrides parent on key conflict",
			levels: []*settings.StringMapSetting{
				strMap(stateValue, map[string]string{"LOG_LEVEL": "info", "REGION": "us-east-1"}),
				strMap(stateValue, map[string]string{"LOG_LEVEL": "debug"}),
				nil,
			},
			want: wantMap(levelDomain, map[string]string{"LOG_LEVEL": "debug", "REGION": "us-east-1"}),
		},
		{
			name: "unset clears everything accumulated above it",
			levels: []*settings.StringMapSetting{
				strMap(stateValue, map[string]string{"A": "1", "B": "2"}),
				strMap(stateUnset, nil),
				nil,
			},
			want: &settings.StringMapSetting{State: stateUnset, ScopeLevel: levelDomain},
		},
		{
			name: "a level below unset refills an empty bucket",
			levels: []*settings.StringMapSetting{
				strMap(stateValue, map[string]string{"A": "1", "B": "2"}),
				strMap(stateUnset, nil),
				strMap(stateValue, map[string]string{"C": "3"}),
			},
			want: wantMap(levelProject, map[string]string{"C": "3"}),
		},
		{
			name: "value with no entries contributes nothing but claims the level",
			levels: []*settings.StringMapSetting{
				strMap(stateValue, map[string]string{"A": "1"}),
				strMap(stateValue, nil),
				nil,
			},
			want: wantMap(levelDomain, map[string]string{"A": "1"}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeStringMapSettings(tt.levels)
			assert.Truef(t, proto.Equal(tt.want, got), "want %v, got %v", tt.want, got)
		})
	}
}

func TestMergeSettings(t *testing.T) {
	org := &settings.Settings{
		Run:                  &settings.RunSettings{DefaultQueue: str(stateValue, "default")},
		EnvironmentVariables: strMap(stateValue, map[string]string{"LOG_LEVEL": "info"}),
		TaskResource: &settings.TaskResourceSettings{
			Max: &settings.TaskResourceDefaults{Cpu: quantity(stateValue, "16")},
		},
	}
	domain := &settings.Settings{
		Run:      &settings.RunSettings{DefaultQueue: str(stateValue, "fast-queue")},
		Security: &settings.SecuritySettings{ServiceAccount: str(stateValue, "runner")},
	}
	project := &settings.Settings{
		EnvironmentVariables: strMap(stateValue, map[string]string{"TEAM": "ml"}),
		TaskResource: &settings.TaskResourceSettings{
			Max: &settings.TaskResourceDefaults{Cpu: quantity(stateValue, "32")},
		},
		PodTemplateName: str(stateValue, "gpu-template"),
	}

	got := mergeSettings([]*settings.Settings{org, domain, project})

	t.Run("scalars resolve from the most specific level that set them", func(t *testing.T) {
		assert.Equal(t, "fast-queue", got.GetRun().GetDefaultQueue().GetStringValue())
		assert.Equal(t, levelDomain, got.GetRun().GetDefaultQueue().GetScopeLevel())
		assert.Equal(t, "runner", got.GetSecurity().GetServiceAccount().GetStringValue())
		assert.Equal(t, "gpu-template", got.GetPodTemplateName().GetStringValue())
	})

	t.Run("nested quantities resolve independently", func(t *testing.T) {
		assert.Equal(t, "32", got.GetTaskResource().GetMax().GetCpu().GetQuantityValue())
		assert.Equal(t, levelProject, got.GetTaskResource().GetMax().GetCpu().GetScopeLevel())
		assert.Nil(t, got.GetTaskResource().GetMax().GetMemory())
		assert.Nil(t, got.GetTaskResource().GetMin())
	})

	t.Run("maps accumulate across levels", func(t *testing.T) {
		assert.Equal(t, map[string]string{"LOG_LEVEL": "info", "TEAM": "ml"},
			got.GetEnvironmentVariables().GetMapValue().GetEntries())
	})

	t.Run("groups nobody configured are absent, not empty", func(t *testing.T) {
		assert.Nil(t, got.GetStorage())
		assert.Nil(t, got.GetApp())
		assert.Nil(t, got.GetLabels())
		assert.Nil(t, got.GetAnnotations())
		assert.Nil(t, got.GetRun().GetMaxActionConcurrency())
	})
}

func TestMergeSettings_NothingStored(t *testing.T) {
	got := mergeSettings([]*settings.Settings{nil, nil, nil})

	require.NotNil(t, got)
	assert.Nil(t, got.GetRun())
	assert.Nil(t, got.GetTaskResource())
	assert.Nil(t, got.GetEnvironmentVariables())
}

// TestMergeSettingsHandlesEveryField fails when a field is added to the Settings
// proto without being handled in mergeSettings. Adding a field there is easy to
// forget, and a forgotten field silently never resolves.
func TestMergeSettingsHandlesEveryField(t *testing.T) {
	handled := map[string]bool{
		"run":                   true,
		"security":              true,
		"storage":               true,
		"task_resource":         true,
		"labels":                true,
		"annotations":           true,
		"environment_variables": true,
		"app":                   true,
		"pod_template_name":     true,
	}

	fields := (&settings.Settings{}).ProtoReflect().Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		name := string(fields.Get(i).Name())
		assert.Truef(t, handled[name],
			"Settings field %q is not handled by mergeSettings; add it there, then add it to this list", name)
	}
}
