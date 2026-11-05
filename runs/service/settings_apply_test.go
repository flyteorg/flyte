package service

import (
	"testing"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/settings"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func runSettings(queue *settings.StringSetting, concurrency *settings.Int64Setting) *settings.Settings {
	return &settings.Settings{
		Run: &settings.RunSettings{DefaultQueue: queue, MaxActionConcurrency: concurrency},
	}
}

func podTemplateSettings(name *settings.StringSetting) *settings.Settings {
	return &settings.Settings{PodTemplateName: name}
}

func TestApplyRunSettings(t *testing.T) {
	tests := []struct {
		name            string
		spec            *task.RunSpec
		resolved        *settings.Settings
		wantQueue       string
		wantConcurrency uint32
		wantPodTemplate string
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			applyRunSettings(tt.spec, tt.resolved)
			assert.Equal(t, tt.wantQueue, tt.spec.GetQueue())
			assert.Equal(t, tt.wantConcurrency, tt.spec.GetMaxActionConcurrency())
			assert.Equal(t, tt.wantPodTemplate, tt.spec.GetDefaultSettings().GetPodTemplateName())
		})
	}
}

func mapSetting(state settings.SettingState, entries map[string]string) *settings.StringMapSetting {
	return &settings.StringMapSetting{State: state, MapValue: &settings.StringMap{Entries: entries}}
}

func envs(pairs ...string) *task.Envs {
	values := make([]*core.KeyValuePair, 0, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		values = append(values, &core.KeyValuePair{Key: pairs[i], Value: pairs[i+1]})
	}
	return &task.Envs{Values: values}
}

func TestApplyRunSettingsMaps(t *testing.T) {
	tests := []struct {
		name            string
		spec            *task.RunSpec
		resolved        *settings.Settings
		wantLabels      map[string]string
		wantAnnotations map[string]string
		wantEnvs        *task.Envs
	}{
		{
			name: "settings only fill an empty spec",
			spec: &task.RunSpec{},
			resolved: &settings.Settings{
				Labels:               mapSetting(stateValue, map[string]string{"team": "ml"}),
				Annotations:          mapSetting(stateValue, map[string]string{"owner": "alice"}),
				EnvironmentVariables: mapSetting(stateValue, map[string]string{"B": "2", "A": "1"}),
			},
			wantLabels:      map[string]string{"team": "ml"},
			wantAnnotations: map[string]string{"owner": "alice"},
			// Settings entries are appended in key order so the stored spec is deterministic.
			wantEnvs: envs("A", "1", "B", "2"),
		},
		{
			name: "request only is left untouched",
			spec: &task.RunSpec{
				Labels:      &task.Labels{Values: map[string]string{"team": "data"}},
				Annotations: &task.Annotations{Values: map[string]string{"owner": "bob"}},
				Envs:        envs("A", "req"),
			},
			resolved:        &settings.Settings{},
			wantLabels:      map[string]string{"team": "data"},
			wantAnnotations: map[string]string{"owner": "bob"},
			wantEnvs:        envs("A", "req"),
		},
		{
			name: "request wins on a conflicting key and other keys merge in",
			spec: &task.RunSpec{
				Labels:      &task.Labels{Values: map[string]string{"team": "data"}},
				Annotations: &task.Annotations{Values: map[string]string{"owner": "bob"}},
				Envs:        envs("Z", "req-z", "A", "req-a"),
			},
			resolved: &settings.Settings{
				Labels:               mapSetting(stateValue, map[string]string{"team": "ml", "env": "prod"}),
				Annotations:          mapSetting(stateValue, map[string]string{"owner": "alice", "cost": "42"}),
				EnvironmentVariables: mapSetting(stateValue, map[string]string{"A": "set-a", "C": "set-c", "B": "set-b"}),
			},
			wantLabels:      map[string]string{"team": "data", "env": "prod"},
			wantAnnotations: map[string]string{"owner": "bob", "cost": "42"},
			// The request's entries keep their order; only missing keys are appended.
			wantEnvs: envs("Z", "req-z", "A", "req-a", "B", "set-b", "C", "set-c"),
		},
		{
			name: "UNSET and INHERIT contribute nothing",
			spec: &task.RunSpec{},
			resolved: &settings.Settings{
				Labels:               mapSetting(stateUnset, map[string]string{"team": "ml"}),
				Annotations:          mapSetting(stateInherit, map[string]string{"owner": "alice"}),
				EnvironmentVariables: mapSetting(stateUnset, map[string]string{"A": "1"}),
			},
		},
		{
			name: "an empty VALUE map allocates nothing",
			spec: &task.RunSpec{},
			resolved: &settings.Settings{
				Labels:               mapSetting(stateValue, nil),
				Annotations:          mapSetting(stateValue, map[string]string{}),
				EnvironmentVariables: mapSetting(stateValue, nil),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			applyRunSettings(tt.spec, tt.resolved)
			assert.Equal(t, tt.wantLabels, tt.spec.GetLabels().GetValues())
			assert.Equal(t, tt.wantAnnotations, tt.spec.GetAnnotations().GetValues())
			assert.True(t, proto.Equal(tt.wantEnvs, tt.spec.GetEnvs()),
				"envs: want %v, got %v", tt.wantEnvs, tt.spec.GetEnvs())
		})
	}
}
