package service

import (
	"testing"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/settings"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/stretchr/testify/assert"
)

func runSettings(queue *settings.StringSetting, concurrency *settings.Int64Setting) *settings.Settings {
	return &settings.Settings{
		Run: &settings.RunSettings{DefaultQueue: queue, MaxActionConcurrency: concurrency},
	}
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
			assert.Equal(t, tt.wantRunBaseDir, tt.spec.GetRunBaseDir())
			assert.Equal(t, tt.wantRawDataPrefix, tt.spec.GetRawDataStorage().GetRawDataPrefix())
		})
	}
}
