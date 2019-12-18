package logs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

const podName = "PodName"

func TestGetLogsForContainerInPod_NoPlugins(t *testing.T) {
	assert.NoError(t, SetLogConfig(&LogConfig{}))
	l, err := GetLogsForContainerInPod(context.TODO(), nil, 0, " Suffix")
	assert.NoError(t, err)
	assert.Nil(t, l)
}

func TestGetLogsForContainerInPod_NoLogs(t *testing.T) {
	assert.NoError(t, SetLogConfig(&LogConfig{
		IsCloudwatchEnabled: true,
		CloudwatchRegion:    "us-east-1",
		CloudwatchLogGroup:  "/kubernetes/flyte-production",
	}))
	p, err := GetLogsForContainerInPod(context.TODO(), nil, 0, " Suffix")
	assert.NoError(t, err)
	assert.Nil(t, p)
}

func TestGetLogsForContainerInPod_BadIndex(t *testing.T) {
	assert.NoError(t, SetLogConfig(&LogConfig{
		IsCloudwatchEnabled: true,
		CloudwatchRegion:    "us-east-1",
		CloudwatchLogGroup:  "/kubernetes/flyte-production",
	}))

	pod := &v1.Pod{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "ContainerName",
				},
			},
		},
		Status: v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{
				{
					ContainerID: "ContainerID",
				},
			},
		},
	}
	pod.Name = podName

	p, err := GetLogsForContainerInPod(context.TODO(), pod, 1, " Suffix")
	assert.NoError(t, err)
	assert.Nil(t, p)
}

func TestGetLogsForContainerInPod_MissingStatus(t *testing.T) {
	assert.NoError(t, SetLogConfig(&LogConfig{
		IsCloudwatchEnabled: true,
		CloudwatchRegion:    "us-east-1",
		CloudwatchLogGroup:  "/kubernetes/flyte-production",
	}))

	pod := &v1.Pod{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "ContainerName",
				},
			},
		},
		Status: v1.PodStatus{},
	}
	pod.Name = podName

	p, err := GetLogsForContainerInPod(context.TODO(), pod, 1, " Suffix")
	assert.NoError(t, err)
	assert.Nil(t, p)
}

func TestGetLogsForContainerInPod_Cloudwatch(t *testing.T) {
	assert.NoError(t, SetLogConfig(&LogConfig{IsCloudwatchEnabled: true,
		CloudwatchRegion:   "us-east-1",
		CloudwatchLogGroup: "/kubernetes/flyte-production",
	}))

	pod := &v1.Pod{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "ContainerName",
				},
			},
		},
		Status: v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{
				{
					ContainerID: "ContainerID",
				},
			},
		},
	}
	pod.Name = podName

	logs, err := GetLogsForContainerInPod(context.TODO(), pod, 0, " Suffix")
	assert.Nil(t, err)
	assert.Len(t, logs, 1)
}

func TestGetLogsForContainerInPod_K8s(t *testing.T) {
	assert.NoError(t, SetLogConfig(&LogConfig{
		IsKubernetesEnabled: true,
		KubernetesURL:       "k8s.com",
	}))

	pod := &v1.Pod{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "ContainerName",
				},
			},
		},
		Status: v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{
				{
					ContainerID: "ContainerID",
				},
			},
		},
	}
	pod.Name = podName

	logs, err := GetLogsForContainerInPod(context.TODO(), pod, 0, " Suffix")
	assert.Nil(t, err)
	assert.Len(t, logs, 1)
}

func TestGetLogsForContainerInPod_All(t *testing.T) {
	assert.NoError(t, SetLogConfig(&LogConfig{
		IsKubernetesEnabled: true,
		KubernetesURL:       "k8s.com",
		IsCloudwatchEnabled: true,
		CloudwatchRegion:    "us-east-1",
		CloudwatchLogGroup:  "/kubernetes/flyte-production",
	}))

	pod := &v1.Pod{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "ContainerName",
				},
			},
		},
		Status: v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{
				{
					ContainerID: "ContainerID",
				},
			},
		},
	}
	pod.Name = podName

	logs, err := GetLogsForContainerInPod(context.TODO(), pod, 0, " Suffix")
	assert.Nil(t, err)
	assert.Len(t, logs, 2)
}

func TestGetLogsForContainerInPod_Stackdriver(t *testing.T) {

	assert.NoError(t, SetLogConfig(&LogConfig{
		IsStackDriverEnabled:       true,
		GCPProjectName:             "myGCPProject",
		StackdriverLogResourceName: "aws_ec2_instance",
	}))

	pod := &v1.Pod{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "ContainerName",
				},
			},
		},
		Status: v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{
				{
					ContainerID: "ContainerID",
				},
			},
		},
	}
	pod.Name = podName

	logs, err := GetLogsForContainerInPod(context.TODO(), pod, 0, " Suffix")
	assert.Nil(t, err)
	assert.Len(t, logs, 1)
}
