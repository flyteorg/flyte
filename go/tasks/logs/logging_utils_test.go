package logs

import (
	"context"
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/go-test/deep"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

func TestGetLogsForContainerInPod_LegacyTemplate(t *testing.T) {
	t.Run("All Templates available", func(t *testing.T) {
		assertTestSucceeded(t, &LogConfig{
			IsKubernetesEnabled:   true,
			KubernetesTemplateURI: "https://k8s-my-log-server/{{ .namespace }}/{{ .podName }}/{{ .containerName }}/{{ .containerId }}",

			IsCloudwatchEnabled:   true,
			CloudwatchTemplateURI: "https://cw-my-log-server/{{ .namespace }}/{{ .podName }}/{{ .containerName }}/{{ .containerId }}",

			IsStackDriverEnabled:   true,
			StackDriverTemplateURI: "https://sd-my-log-server/{{ .namespace }}/{{ .podName }}/{{ .containerName }}/{{ .containerId }}",
		}, []*core.TaskLog{
			{
				Uri:           "https://k8s-my-log-server/my-namespace/my-pod/ContainerName/ContainerID",
				MessageFormat: core.TaskLog_JSON,
				Name:          "Kubernetes Logs my-Suffix",
			},
			{
				Uri:           "https://cw-my-log-server/my-namespace/my-pod/ContainerName/ContainerID",
				MessageFormat: core.TaskLog_JSON,
				Name:          "Cloudwatch Logs my-Suffix",
			},
			{
				Uri:           "https://sd-my-log-server/my-namespace/my-pod/ContainerName/ContainerID",
				MessageFormat: core.TaskLog_JSON,
				Name:          "Stackdriver Logs my-Suffix",
			},
		})
	})

	t.Run("StackDriver", func(t *testing.T) {
		assertTestSucceeded(t, &LogConfig{
			IsStackDriverEnabled:   true,
			StackDriverTemplateURI: "https://sd-my-log-server/{{ .namespace }}/{{ .podName }}/{{ .containerName }}/{{ .containerId }}",
		}, []*core.TaskLog{
			{
				Uri:           "https://sd-my-log-server/my-namespace/my-pod/ContainerName/ContainerID",
				MessageFormat: core.TaskLog_JSON,
				Name:          "Stackdriver Logs my-Suffix",
			},
		})
	})
}

func assertTestSucceeded(tb testing.TB, config *LogConfig, expectedTaskLogs []*core.TaskLog) {
	assert.NoError(tb, SetLogConfig(config))

	pod := &v1.Pod{
		ObjectMeta: v12.ObjectMeta{
			Namespace: "my-namespace",
			Name:      "my-pod",
		},
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

	logs, err := GetLogsForContainerInPod(context.TODO(), pod, 0, " my-Suffix")
	assert.Nil(tb, err)
	assert.Len(tb, logs, len(expectedTaskLogs))
	if diff := deep.Equal(logs, expectedTaskLogs); len(diff) > 0 {
		assert.FailNowf(tb, "Not Equal.", "Diff: %v", diff)
	}
}

func TestGetLogsForContainerInPod_Templates(t *testing.T) {
	assertTestSucceeded(t, &LogConfig{
		Templates: []TemplateLogPluginConfig{
			{
				DisplayName: "StackDriver",
				TemplateURIs: []string{
					"https://my-log-server/{{ .namespace }}/{{ .podName }}/{{ .containerName }}/{{ .containerId }}",
				},
				MessageFormat: core.TaskLog_JSON,
			},
		},
	}, []*core.TaskLog{
		{
			Uri:           "https://my-log-server/my-namespace/my-pod/ContainerName/ContainerID",
			MessageFormat: core.TaskLog_JSON,
			Name:          "StackDriver my-Suffix",
		},
	})
}
