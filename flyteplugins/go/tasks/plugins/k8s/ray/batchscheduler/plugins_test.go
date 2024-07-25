package batchscheduler

import (
	"testing"
)

func TestCreateSchedulerPlugin(t *testing.T) {
	var tests = []struct{
		input *BatchSchedulerConfig
		expect string
	}{
		{input: &BatchSchedulerConfig{Scheduler: DefaultScheduler}, expect: DefaultScheduler},
		{input: &BatchSchedulerConfig{Scheduler: Yunikorn}, expect: Yunikorn},
		{input: &BatchSchedulerConfig{Scheduler:"Unknown"}, expect: DefaultScheduler},
	}
	for _, tt := range tests {
		t.Run("New scheduler plugin", func(t *testing.T) {
			if got := NewSchedulerPlugin(tt.input); got.GetSchedulerName() != tt.expect {
				t.Errorf("got %s, expect %s", got, tt.expect)
			}
		})
	}
}