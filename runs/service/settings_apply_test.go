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

func TestApplyRunSettings(t *testing.T) {
	tests := []struct {
		name            string
		spec            *task.RunSpec
		resolved        *settings.Settings
		wantQueue       string
		wantConcurrency uint32
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			applyRunSettings(tt.spec, tt.resolved)
			assert.Equal(t, tt.wantQueue, tt.spec.GetQueue())
			assert.Equal(t, tt.wantConcurrency, tt.spec.GetMaxActionConcurrency())
		})
	}
}
