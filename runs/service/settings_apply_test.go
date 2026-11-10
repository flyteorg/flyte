package service

import (
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

func TestApplyRunSettingsTaskResourceDefaults(t *testing.T) {
	tests := []struct {
		name     string
		spec     *task.RunSpec
		resolved *settings.Settings
		want     *task.TaskResourceDefaults
	}{
		{
			name:     "Get nil when setting is empty",
			spec:     &task.RunSpec{},
			resolved: &settings.Settings{},
			want:     nil,
		},
		{
			name: "set requests and max values from setting",
			spec: &task.RunSpec{},
			resolved: &settings.Settings{TaskResource: &settings.TaskResourceSettings{
				Min: &settings.TaskResourceDefaults{Memory: quantity(stateValue, "256Mi"), Storage: quantity(stateValue, "1Gi")},
				Max: &settings.TaskResourceDefaults{Cpu: quantity(stateValue, "1"), Gpu: quantity(stateValue, "2")},
			}},
			want: &task.TaskResourceDefaults{
				Requests: []*core.Resources_ResourceEntry{
					{Name: core.Resources_MEMORY, Value: "256Mi"},
					{Name: core.Resources_EPHEMERAL_STORAGE, Value: "1Gi"},
				},
				Max: []*core.Resources_ResourceEntry{
					{Name: core.Resources_CPU, Value: "1"},
					{Name: core.Resources_GPU, Value: "2"},
				},
			},
		},
		{
			name: "set request with max value when min value is greater than max value",
			spec: &task.RunSpec{},
			resolved: &settings.Settings{TaskResource: &settings.TaskResourceSettings{
				Min: &settings.TaskResourceDefaults{Cpu: quantity(stateValue, "2000m")},
				Max: &settings.TaskResourceDefaults{Cpu: quantity(stateValue, "1")},
			}},
			want: &task.TaskResourceDefaults{
				Requests: []*core.Resources_ResourceEntry{{Name: core.Resources_CPU, Value: "1"}},
				Max:      []*core.Resources_ResourceEntry{{Name: core.Resources_CPU, Value: "1"}},
			},
		},
		{
			name: "INHERIT and UNSET dimensions is ignored",
			spec: &task.RunSpec{},
			resolved: &settings.Settings{TaskResource: &settings.TaskResourceSettings{
				Min: &settings.TaskResourceDefaults{Cpu: &settings.QuantitySetting{State: stateInherit, QuantityValue: "1"}},
				Max: &settings.TaskResourceDefaults{Cpu: &settings.QuantitySetting{State: stateUnset, QuantityValue: "2"}},
			}},
			want: nil,
		},
		{
			name: "default accelerator is carried over",
			spec: &task.RunSpec{},
			resolved: &settings.Settings{TaskResource: &settings.TaskResourceSettings{
				DefaultAccelerator: &settings.AcceleratorSetting{State: stateValue, AcceleratorValue: &core.GPUAccelerator{Device: "nvidia-a100-80gb"}},
			}},
			want: &task.TaskResourceDefaults{DefaultAccelerator: &core.GPUAccelerator{Device: "nvidia-a100-80gb"}},
		},
		{
			name: "a caller-supplied value is overwritten",
			spec: &task.RunSpec{DefaultSettings: &task.DefaultSettings{
				PodTemplateName: "user-template",
				TaskResourceDefaults: &task.TaskResourceDefaults{
					Max: []*core.Resources_ResourceEntry{{Name: core.Resources_CPU, Value: "64"}},
				},
			}},
			resolved: &settings.Settings{TaskResource: &settings.TaskResourceSettings{
				Max: &settings.TaskResourceDefaults{Cpu: quantity(stateValue, "1")},
			}},
			want: &task.TaskResourceDefaults{
				Max: []*core.Resources_ResourceEntry{{Name: core.Resources_CPU, Value: "1"}},
			},
		},
		{
			name: "a caller-supplied value is left as is when no scope sets task_resource",
			spec: &task.RunSpec{DefaultSettings: &task.DefaultSettings{
				TaskResourceDefaults: &task.TaskResourceDefaults{
					Max: []*core.Resources_ResourceEntry{{Name: core.Resources_CPU, Value: "64"}},
				},
			}},
			resolved: &settings.Settings{},
			want: &task.TaskResourceDefaults{
				Max: []*core.Resources_ResourceEntry{{Name: core.Resources_CPU, Value: "64"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			applyRunSettings(tt.spec, tt.resolved)
			got := tt.spec.GetDefaultSettings().GetTaskResourceDefaults()
			assert.True(t, proto.Equal(tt.want, got), "got %v, want %v", got, tt.want)
		})
	}
}
