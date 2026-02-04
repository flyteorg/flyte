package pod

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestIsTerminal(t *testing.T) {
	p := plugin{}
	ctx := context.Background()

	tests := []struct {
		name           string
		phase          v1.PodPhase
		expectedResult bool
	}{
		{"Succeeded", v1.PodSucceeded, true},
		{"Failed", v1.PodFailed, true},
		{"Running", v1.PodRunning, false},
		{"Pending", v1.PodPending, false},
		{"Unknown", v1.PodUnknown, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pod := &v1.Pod{
				Status: v1.PodStatus{
					Phase: tt.phase,
				},
			}
			result, err := p.IsTerminal(ctx, pod)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestIsTerminal_WrongResourceType(t *testing.T) {
	p := plugin{}
	ctx := context.Background()

	var wrongResource client.Object = &v1.ConfigMap{}
	result, err := p.IsTerminal(ctx, wrongResource)
	assert.Error(t, err)
	assert.False(t, result)
	assert.Contains(t, err.Error(), "unexpected resource type")
}

func TestGetCompletionTime(t *testing.T) {
	p := plugin{}

	now := time.Now().Truncate(time.Second)
	earlier := now.Add(-1 * time.Hour)
	evenEarlier := now.Add(-2 * time.Hour)

	tests := []struct {
		name         string
		pod          *v1.Pod
		expectedTime time.Time
	}{
		{
			name: "uses container termination time",
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.NewTime(evenEarlier),
				},
				Status: v1.PodStatus{
					ContainerStatuses: []v1.ContainerStatus{
						{
							State: v1.ContainerState{
								Terminated: &v1.ContainerStateTerminated{
									FinishedAt: metav1.NewTime(now),
								},
							},
						},
					},
				},
			},
			expectedTime: now,
		},
		{
			name: "uses latest of multiple container termination times",
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.NewTime(evenEarlier),
				},
				Status: v1.PodStatus{
					ContainerStatuses: []v1.ContainerStatus{
						{
							State: v1.ContainerState{
								Terminated: &v1.ContainerStateTerminated{
									FinishedAt: metav1.NewTime(earlier),
								},
							},
						},
						{
							State: v1.ContainerState{
								Terminated: &v1.ContainerStateTerminated{
									FinishedAt: metav1.NewTime(now),
								},
							},
						},
					},
				},
			},
			expectedTime: now,
		},
		{
			name: "uses init container termination time",
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.NewTime(evenEarlier),
				},
				Status: v1.PodStatus{
					InitContainerStatuses: []v1.ContainerStatus{
						{
							State: v1.ContainerState{
								Terminated: &v1.ContainerStateTerminated{
									FinishedAt: metav1.NewTime(now),
								},
							},
						},
					},
				},
			},
			expectedTime: now,
		},
		{
			name: "uses LastTerminationState",
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.NewTime(evenEarlier),
				},
				Status: v1.PodStatus{
					ContainerStatuses: []v1.ContainerStatus{
						{
							LastTerminationState: v1.ContainerState{
								Terminated: &v1.ContainerStateTerminated{
									FinishedAt: metav1.NewTime(now),
								},
							},
						},
					},
				},
			},
			expectedTime: now,
		},
		{
			name: "falls back to DeletionTimestamp",
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.NewTime(evenEarlier),
					DeletionTimestamp: &metav1.Time{Time: now},
				},
				Status: v1.PodStatus{
					StartTime: &metav1.Time{Time: earlier},
				},
			},
			expectedTime: now,
		},
		{
			name: "falls back to StartTime",
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.NewTime(evenEarlier),
				},
				Status: v1.PodStatus{
					StartTime: &metav1.Time{Time: now},
				},
			},
			expectedTime: now,
		},
		{
			name: "falls back to CreationTimestamp",
			pod: &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.NewTime(now),
				},
				Status: v1.PodStatus{},
			},
			expectedTime: now,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := p.GetCompletionTime(tt.pod)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedTime.Unix(), result.Unix())
		})
	}
}

func TestGetCompletionTime_WrongResourceType(t *testing.T) {
	p := plugin{}

	var wrongResource client.Object = &v1.ConfigMap{}
	result, err := p.GetCompletionTime(wrongResource)
	assert.Error(t, err)
	assert.True(t, result.IsZero())
	assert.Contains(t, err.Error(), "unexpected resource type")
}
