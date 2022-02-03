package k8s

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
)

func TestApplyNodeSelectorLabels(t *testing.T) {
	ctx := context.Background()
	cfg := &Config{
		NodeSelector: map[string]string{
			"disktype": "ssd",
		},
	}
	pod := &corev1.Pod{}

	pod = applyNodeSelectorLabels(ctx, cfg, pod)

	assert.Equal(t, pod.Spec.NodeSelector, cfg.NodeSelector)
}

func TestApplyPodTolerations(t *testing.T) {
	ctx := context.Background()
	cfg := &Config{
		Tolerations: []v1.Toleration{{
			Key:      "reserved",
			Operator: "equal",
			Value:    "value",
			Effect:   "NoSchedule",
		}},
	}
	pod := &corev1.Pod{}

	pod = applyPodTolerations(ctx, cfg, pod)

	assert.Equal(t, pod.Spec.Tolerations, cfg.Tolerations)
}

func TestFormatSubTaskName(t *testing.T) {
	ctx := context.Background()
	parentName := "foo"

	tests := []struct {
		index        int
		retryAttempt uint64
		want         string
	}{
		{0, 0, fmt.Sprintf("%v-%v", parentName, 0)},
		{1, 0, fmt.Sprintf("%v-%v", parentName, 1)},
		{0, 1, fmt.Sprintf("%v-%v-%v", parentName, 0, 1)},
		{1, 1, fmt.Sprintf("%v-%v-%v", parentName, 1, 1)},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("format-subtask-name-%v", i), func(t *testing.T) {
			assert.Equal(t, tt.want, formatSubTaskName(ctx, parentName, tt.index, tt.retryAttempt))
		})
	}
}
