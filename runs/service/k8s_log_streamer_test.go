package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/logs/dataplane"
)

func TestParseLogLine_WithTimestamp(t *testing.T) {
	line := "2024-01-15T10:30:00.123456789Z Hello, world!"
	logLine := parseLogLine(line)

	assert.Equal(t, "Hello, world!", logLine.Message)
	assert.NotNil(t, logLine.Timestamp)
	expected := time.Date(2024, 1, 15, 10, 30, 0, 123456789, time.UTC)
	assert.Equal(t, expected, logLine.Timestamp.AsTime())
	assert.Equal(t, dataplane.LogLineOriginator_USER, logLine.Originator)
}

func TestParseLogLine_WithoutTimestamp(t *testing.T) {
	line := "just a plain log message"
	logLine := parseLogLine(line)

	assert.Equal(t, "just a plain log message", logLine.Message)
	assert.Nil(t, logLine.Timestamp)
	assert.Equal(t, dataplane.LogLineOriginator_USER, logLine.Originator)
}

func TestParseLogLine_MalformedTimestamp(t *testing.T) {
	line := "not-a-timestamp some message"
	logLine := parseLogLine(line)

	assert.Equal(t, "not-a-timestamp some message", logLine.Message)
	assert.Nil(t, logLine.Timestamp)
}

func TestParseLogLine_EmptyMessage(t *testing.T) {
	line := "2024-01-15T10:30:00Z "
	logLine := parseLogLine(line)

	assert.Equal(t, "", logLine.Message)
	assert.NotNil(t, logLine.Timestamp)
}

func TestGetPrimaryPodAndContainer_HappyPath(t *testing.T) {
	logCtx := &core.LogContext{
		PrimaryPodName: "my-pod",
		Pods: []*core.PodLogContext{
			{
				PodName:              "my-pod",
				Namespace:            "default",
				PrimaryContainerName: "main",
				Containers: []*core.ContainerContext{
					{ContainerName: "main"},
					{ContainerName: "sidecar"},
				},
			},
		},
	}

	pod, container, err := getPrimaryPodAndContainer(logCtx)
	assert.NoError(t, err)
	assert.Equal(t, "my-pod", pod.GetPodName())
	assert.Equal(t, "default", pod.GetNamespace())
	assert.Equal(t, "main", container.GetContainerName())
}

func TestGetPrimaryPodAndContainer_EmptyPodName(t *testing.T) {
	logCtx := &core.LogContext{
		PrimaryPodName: "",
	}

	_, _, err := getPrimaryPodAndContainer(logCtx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "primary pod name is empty")
}

func TestGetPrimaryPodAndContainer_PodNotFound(t *testing.T) {
	logCtx := &core.LogContext{
		PrimaryPodName: "missing-pod",
		Pods: []*core.PodLogContext{
			{PodName: "other-pod"},
		},
	}

	_, _, err := getPrimaryPodAndContainer(logCtx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found in log context")
}

func TestGetPrimaryPodAndContainer_ContainerNotFound(t *testing.T) {
	logCtx := &core.LogContext{
		PrimaryPodName: "my-pod",
		Pods: []*core.PodLogContext{
			{
				PodName:              "my-pod",
				PrimaryContainerName: "missing-container",
				Containers: []*core.ContainerContext{
					{ContainerName: "other"},
				},
			},
		},
	}

	_, _, err := getPrimaryPodAndContainer(logCtx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "primary container")
}
