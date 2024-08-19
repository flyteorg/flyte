package batchscheduler

import (
	"testing"

	"github.com/stretchr/testify/assert"

	schedulerConfig "github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/k8s/batchscheduler/config"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/k8s/batchscheduler/scheduler/kubernetes"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/k8s/batchscheduler/scheduler/yunikorn"
)

func TestCreateSchedulerPlugin(t *testing.T) {
	var tests = []struct {
		input  *schedulerConfig.Config
		expect string
	}{
		{input: &schedulerConfig.Config{Scheduler: kubernetes.DefaultScheduler}, expect: kubernetes.DefaultScheduler},
		{input: &schedulerConfig.Config{Scheduler: yunikorn.Yunikorn}, expect: yunikorn.Yunikorn},
		{input: &schedulerConfig.Config{Scheduler: "Unknown"}, expect: kubernetes.DefaultScheduler},
	}
	for _, tt := range tests {
		t.Run("New scheduler plugin", func(t *testing.T) {
			p := NewSchedulerPlugin(tt.input)
			assert.Equal(t, tt.expect, p.GetSchedulerName())
		})
	}
}
