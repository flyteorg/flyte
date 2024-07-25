package batchscheduler

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
			p := NewSchedulerPlugin(tt.input)
			assert.Equal(t, tt.expect, p.GetSchedulerName())
		})
	}
}