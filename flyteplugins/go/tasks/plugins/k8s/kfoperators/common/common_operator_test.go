package common

import (
	"fmt"
	"testing"
	"time"

	"github.com/flyteorg/flyteplugins/go/tasks/logs"

	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	commonOp "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestExtractMPICurrentCondition(t *testing.T) {
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
	currentCondition, err := ExtractMPICurrentCondition(jobConditions)
	assert.NoError(t, err)
	assert.Equal(t, currentCondition, jobCreated)

	jobConditions = nil
	currentCondition, err = ExtractMPICurrentCondition(jobConditions)
	assert.Error(t, err)
	assert.Equal(t, currentCondition, commonOp.JobCondition{})
	assert.Equal(t, err, fmt.Errorf("found no current condition. Conditions: %+v", jobConditions))
}

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

	jobConditions = nil
	currentCondition, err = ExtractCurrentCondition(jobConditions)
	assert.Error(t, err)
	assert.Equal(t, currentCondition, commonOp.JobCondition{})
	assert.Equal(t, err, fmt.Errorf("found no current condition. Conditions: %+v", jobConditions))
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

	jobRestarting = commonOp.JobCondition{
		Type: commonOp.JobRunning,
	}
	taskPhase, err = GetPhaseInfo(jobRestarting, time.Now(), pluginsCore.TaskInfo{})
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)
}

func TestGetMPIPhaseInfo(t *testing.T) {
	jobCreated := commonOp.JobCondition{
		Type: commonOp.JobCreated,
	}
	taskPhase, err := GetMPIPhaseInfo(jobCreated, time.Now(), pluginsCore.TaskInfo{})
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseQueued, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	jobSucceeded := commonOp.JobCondition{
		Type: commonOp.JobSucceeded,
	}
	taskPhase, err = GetMPIPhaseInfo(jobSucceeded, time.Now(), pluginsCore.TaskInfo{})
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseSuccess, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	jobFailed := commonOp.JobCondition{
		Type: commonOp.JobFailed,
	}
	taskPhase, err = GetMPIPhaseInfo(jobFailed, time.Now(), pluginsCore.TaskInfo{})
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRetryableFailure, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	jobRestarting := commonOp.JobCondition{
		Type: commonOp.JobRestarting,
	}
	taskPhase, err = GetMPIPhaseInfo(jobRestarting, time.Now(), pluginsCore.TaskInfo{})
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)

	jobRestarting = commonOp.JobCondition{
		Type: commonOp.JobRunning,
	}
	taskPhase, err = GetMPIPhaseInfo(jobRestarting, time.Now(), pluginsCore.TaskInfo{})
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, taskPhase.Phase())
	assert.NotNil(t, taskPhase.Info())
	assert.Nil(t, err)
}

func TestGetLogs(t *testing.T) {
	assert.NoError(t, logs.SetLogConfig(&logs.LogConfig{
		IsKubernetesEnabled: true,
		KubernetesURL:       "k8s.com",
	}))

	workers := int32(1)
	launcher := int32(1)

	jobLogs, err := GetLogs(MPITaskType, "test", "mpi-namespace", workers, launcher, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(jobLogs))
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-worker-0/pod?namespace=mpi-namespace", "mpi-namespace", "test"), jobLogs[0].Uri)

	jobLogs, err = GetLogs(PytorchTaskType, "test", "pytorch-namespace", workers, launcher, 0)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(jobLogs))
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-master-0/pod?namespace=pytorch-namespace", "pytorch-namespace", "test"), jobLogs[0].Uri)
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-worker-0/pod?namespace=pytorch-namespace", "pytorch-namespace", "test"), jobLogs[1].Uri)

	jobLogs, err = GetLogs(TensorflowTaskType, "test", "tensorflow-namespace", workers, launcher, 1)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(jobLogs))
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-worker-0/pod?namespace=tensorflow-namespace", "tensorflow-namespace", "test"), jobLogs[0].Uri)
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-psReplica-0/pod?namespace=tensorflow-namespace", "tensorflow-namespace", "test"), jobLogs[1].Uri)
	assert.Equal(t, fmt.Sprintf("k8s.com/#!/log/%s/%s-chiefReplica-0/pod?namespace=tensorflow-namespace", "tensorflow-namespace", "test"), jobLogs[2].Uri)

}
