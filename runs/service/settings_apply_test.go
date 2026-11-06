package service

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/settings"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
)

func runSettings(queue *settings.StringSetting, concurrency *settings.Int64Setting) *settings.Settings {
	return &settings.Settings{
		Run: &settings.RunSettings{DefaultQueue: queue, MaxActionConcurrency: concurrency},
	}
}

func podTemplateSettings(name *settings.StringSetting) *settings.Settings {
	return &settings.Settings{PodTemplateName: name}
}

func runBaseDirSettings(baseDir *settings.StringSetting) *settings.Settings {
	return &settings.Settings{Run: &settings.RunSettings{RunBaseDir: baseDir}}
}

func rawDataSettings(rawDataPath *settings.StringSetting) *settings.Settings {
	return &settings.Settings{Storage: &settings.StorageSettings{RawDataPath: rawDataPath}}
}

func TestApplyRunSettings(t *testing.T) {
	tests := []struct {
		name              string
		spec              *task.RunSpec
		resolved          *settings.Settings
		wantQueue         string
		wantConcurrency   uint32
		wantPodTemplate   string
		wantRunBaseDir    string
		wantRawDataPrefix string
	}{
		{
			name:      "empty queue takes the settings value",
			spec:      &task.RunSpec{},
			resolved:  runSettings(&settings.StringSetting{State: stateValue, StringValue: "fast-queue"}, nil),
			wantQueue: "fast-queue",
		},
		{
			name:      "an explicit queue wins over settings",
			spec:      &task.RunSpec{Queue: "user-queue"},
			resolved:  runSettings(&settings.StringSetting{State: stateValue, StringValue: "fast-queue"}, nil),
			wantQueue: "user-queue",
		},
		{
			name:      "a queue in INHERIT contributes nothing",
			spec:      &task.RunSpec{},
			resolved:  runSettings(&settings.StringSetting{State: stateInherit, StringValue: "fast-queue"}, nil),
			wantQueue: "",
		},
		{
			name:            "zero concurrency takes the settings value",
			spec:            &task.RunSpec{},
			resolved:        runSettings(nil, &settings.Int64Setting{State: stateValue, IntValue: 5}),
			wantConcurrency: 5,
		},
		{
			name:            "an explicit concurrency wins over settings",
			spec:            &task.RunSpec{MaxActionConcurrency: 3},
			resolved:        runSettings(nil, &settings.Int64Setting{State: stateValue, IntValue: 5}),
			wantConcurrency: 3,
		},
		{
			name:            "concurrency in UNSET contributes nothing",
			spec:            &task.RunSpec{},
			resolved:        runSettings(nil, &settings.Int64Setting{State: stateUnset, IntValue: 5}),
			wantConcurrency: 0,
		},
		{
			name:     "no settings at all",
			spec:     &task.RunSpec{},
			resolved: &settings.Settings{},
		},
		{
			name:     "nil spec does not panic",
			spec:     nil,
			resolved: runSettings(&settings.StringSetting{State: stateValue, StringValue: "fast-queue"}, nil),
		},
		{
			name:            "empty pod template takes the settings value",
			spec:            &task.RunSpec{},
			resolved:        podTemplateSettings(&settings.StringSetting{State: stateValue, StringValue: "gpu-template"}),
			wantPodTemplate: "gpu-template",
		},
		{
			name:            "an explicit pod template wins over settings",
			spec:            &task.RunSpec{DefaultSettings: &task.DefaultSettings{PodTemplateName: "user-template"}},
			resolved:        podTemplateSettings(&settings.StringSetting{State: stateValue, StringValue: "gpu-template"}),
			wantPodTemplate: "user-template",
		},
		{
			name:            "a pod template in INHERIT contributes nothing",
			spec:            &task.RunSpec{},
			resolved:        podTemplateSettings(&settings.StringSetting{State: stateInherit, StringValue: "gpu-template"}),
			wantPodTemplate: "",
		},
		{
			name:           "empty run base dir takes the settings value",
			spec:           &task.RunSpec{},
			resolved:       runBaseDirSettings(&settings.StringSetting{State: stateValue, StringValue: "s3://settings-base"}),
			wantRunBaseDir: "s3://settings-base",
		},
		{
			name:           "an explicit run base dir wins over settings",
			spec:           &task.RunSpec{RunBaseDir: "s3://user-base"},
			resolved:       runBaseDirSettings(&settings.StringSetting{State: stateValue, StringValue: "s3://settings-base"}),
			wantRunBaseDir: "s3://user-base",
		},
		{
			name:           "a run base dir in INHERIT contributes nothing",
			spec:           &task.RunSpec{},
			resolved:       runBaseDirSettings(&settings.StringSetting{State: stateInherit, StringValue: "s3://settings-base"}),
			wantRunBaseDir: "",
		},
		{
			name:              "empty raw data prefix takes the settings value",
			spec:              &task.RunSpec{},
			resolved:          rawDataSettings(&settings.StringSetting{State: stateValue, StringValue: "s3://settings-raw"}),
			wantRawDataPrefix: "s3://settings-raw",
		},
		{
			name:              "an explicit raw data prefix wins over settings",
			spec:              &task.RunSpec{RawDataStorage: &task.RawDataStorage{RawDataPrefix: "s3://user-raw"}},
			resolved:          rawDataSettings(&settings.StringSetting{State: stateValue, StringValue: "s3://settings-raw"}),
			wantRawDataPrefix: "s3://user-raw",
		},
		{
			name:              "a raw data path in UNSET contributes nothing",
			spec:              &task.RunSpec{},
			resolved:          rawDataSettings(&settings.StringSetting{State: stateUnset, StringValue: "s3://settings-raw"}),
			wantRawDataPrefix: "",
		},
		{
			// Present but empty still counts as no request value. Guarding on a nil
			// RawDataStorage instead would skip this row and leave the prefix empty.
			name:              "a raw data storage message with no prefix takes the settings value",
			spec:              &task.RunSpec{RawDataStorage: &task.RawDataStorage{}},
			resolved:          rawDataSettings(&settings.StringSetting{State: stateValue, StringValue: "s3://settings-raw"}),
			wantRawDataPrefix: "s3://settings-raw",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			applyRunSettings(tt.spec, tt.resolved)
			assert.Equal(t, tt.wantQueue, tt.spec.GetQueue())
			assert.Equal(t, tt.wantConcurrency, tt.spec.GetMaxActionConcurrency())
			assert.Equal(t, tt.wantPodTemplate, tt.spec.GetDefaultSettings().GetPodTemplateName())
			assert.Equal(t, tt.wantRunBaseDir, tt.spec.GetRunBaseDir())
			assert.Equal(t, tt.wantRawDataPrefix, tt.spec.GetRawDataStorage().GetRawDataPrefix())
		})
	}
}

func mapSetting(state settings.SettingState, entries map[string]string) *settings.StringMapSetting {
	return &settings.StringMapSetting{State: state, MapValue: &settings.StringMap{Entries: entries}}
}

func mapSettings(state settings.SettingState, entries map[string]string) *settings.Settings {
	return &settings.Settings{
		Labels:               mapSetting(state, entries),
		Annotations:          mapSetting(state, entries),
		EnvironmentVariables: mapSetting(state, entries),
	}
}

func envs(kvs ...string) *task.Envs {
	e := &task.Envs{}
	for i := 0; i < len(kvs); i += 2 {
		e.Values = append(e.Values, &core.KeyValuePair{Key: kvs[i], Value: kvs[i+1]})
	}
	return e
}

func TestApplyRunSettings_Maps(t *testing.T) {
	tests := []struct {
		name     string
		spec     *task.RunSpec
		resolved *settings.Settings
		wantMap  map[string]string // expected labels and annotations
		wantEnvs []*core.KeyValuePair
	}{
		{
			name:     "settings only",
			spec:     &task.RunSpec{},
			resolved: mapSettings(stateValue, map[string]string{"b": "s-b", "a": "s-a"}),
			wantMap:  map[string]string{"a": "s-a", "b": "s-b"},
			wantEnvs: envs("a", "s-a", "b", "s-b").Values,
		},
		{
			name: "request only",
			spec: &task.RunSpec{
				Labels:      &task.Labels{Values: map[string]string{"a": "r-a"}},
				Annotations: &task.Annotations{Values: map[string]string{"a": "r-a"}},
				Envs:        envs("a", "r-a"),
			},
			resolved: &settings.Settings{},
			wantMap:  map[string]string{"a": "r-a"},
			wantEnvs: envs("a", "r-a").Values,
		},
		{
			name: "the request wins on a conflicting key and keeps its env order",
			spec: &task.RunSpec{
				Labels:      &task.Labels{Values: map[string]string{"b": "r-b"}},
				Annotations: &task.Annotations{Values: map[string]string{"b": "r-b"}},
				Envs:        envs("z", "r-z", "b", "r-b"),
			},
			resolved: mapSettings(stateValue, map[string]string{"c": "s-c", "b": "s-b", "a": "s-a"}),
			wantMap:  map[string]string{"a": "s-a", "b": "r-b", "c": "s-c"},
			wantEnvs: envs("z", "r-z", "b", "r-b", "a", "s-a", "c", "s-c").Values,
		},
		{
			name:     "maps in UNSET contribute nothing",
			spec:     &task.RunSpec{},
			resolved: mapSettings(stateUnset, map[string]string{"a": "s-a"}),
		},
		{
			name:     "maps in INHERIT contribute nothing",
			spec:     &task.RunSpec{},
			resolved: mapSettings(stateInherit, map[string]string{"a": "s-a"}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			applyRunSettings(tt.spec, tt.resolved)
			assert.Equal(t, tt.wantMap, tt.spec.GetLabels().GetValues())
			assert.Equal(t, tt.wantMap, tt.spec.GetAnnotations().GetValues())
			assert.True(t, slices.EqualFunc(tt.wantEnvs, tt.spec.GetEnvs().GetValues(), func(a, b *core.KeyValuePair) bool {
				return proto.Equal(a, b)
			}), "envs: got %v", tt.spec.GetEnvs().GetValues())
		})
	}
}
