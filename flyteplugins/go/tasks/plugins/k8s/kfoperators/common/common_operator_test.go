package common

import (
	"testing"
	"time"

	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	commonOp "github.com/kubeflow/tf-operator/pkg/apis/common/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestExtractCurrentCondition(t *testing.T) {
	jobCreated := commonOp.JobCondition{
		Type:   commonOp.JobCreated,
		Status: corev1.ConditionTrue,
	}
	jobRunningActive := commonOp.JobCondition{
		Type:   commonOp.JobRunning,
		Status: corev1.ConditionFalse,
	}
	jobConditions := []commonOp.JobCondition{
		jobCreated,
		jobRunningActive,
	}
	currentCondition, err := ExtractCurrentCondition(jobConditions)
	assert.NoError(t, err)
	assert.Equal(t, currentCondition, jobCreated)
}

func TestGetPhaseInfo(t *testing.T) {
	jobCreated := commonOp.JobCondition{
		Type: commonOp.JobCreated,
	}
	taskPhase, err := GetPhaseInfo(jobCreated, time.Now(), pluginsCore.TaskInfo{})
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseQueued, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	jobSucceeded := commonOp.JobCondition{
		Type: commonOp.JobSucceeded,
	}
	taskPhase, err = GetPhaseInfo(jobSucceeded, time.Now(), pluginsCore.TaskInfo{})
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseSuccess, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	jobFailed := commonOp.JobCondition{
		Type: commonOp.JobFailed,
	}
	taskPhase, err = GetPhaseInfo(jobFailed, time.Now(), pluginsCore.TaskInfo{})
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRetryableFailure, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	jobRestarting := commonOp.JobCondition{
		Type: commonOp.JobRestarting,
	}
	taskPhase, err = GetPhaseInfo(jobRestarting, time.Now(), pluginsCore.TaskInfo{})
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)
}
