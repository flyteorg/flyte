package webhook

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
)

func TestMutateContainer(t *testing.T) {
	ctx := context.Background()
	imgConfig := ManagedImageConfig{
		Regex:              "my-image:.*",
		ReplacementPattern: "my-image:latest",
	}
	mutator, err := NewManagedImageMutator(&ManagedImagesConfig{Images: []ManagedImageConfig{imgConfig}})
	require.NoError(t, err)

	container := v1.Container{
		Name:  "test-container",
		Image: "my-image:v1.0.0",
	}

	newContainer, changed, responseErr := mutator.MutateContainer(ctx, imgConfig, container)
	assert.Nil(t, responseErr)
	assert.True(t, changed)
	assert.Equal(t, "my-image:latest", newContainer.Image)
}

func TestMutateContainer_NoMatch(t *testing.T) {
	ctx := context.Background()
	imgConfig := ManagedImageConfig{
		Regex:              "other-image:.*",
		ReplacementPattern: "other-image:latest",
	}
	mutator, err := NewManagedImageMutator(&ManagedImagesConfig{Images: []ManagedImageConfig{imgConfig}})
	require.NoError(t, err)

	container := v1.Container{
		Name:  "test-container",
		Image: "my-image:v1.0.0",
	}

	newContainer, changed, responseErr := mutator.MutateContainer(ctx, imgConfig, container)
	assert.Nil(t, responseErr)
	assert.False(t, changed)
	assert.Equal(t, "my-image:v1.0.0", newContainer.Image)
}

func TestMutateContainers(t *testing.T) {
	ctx := context.Background()
	imgConfig := ManagedImageConfig{
		Regex:              "my-image:.*",
		ReplacementPattern: "my-image:latest",
	}
	mutator, err := NewManagedImageMutator(&ManagedImagesConfig{Images: []ManagedImageConfig{imgConfig}})
	require.NoError(t, err)

	containers := []v1.Container{
		{Name: "container1", Image: "my-image:v1.0.0"},
		{Name: "container2", Image: "other-image:v1.0.0"},
	}

	newContainers, changed, responseErr := mutator.MutateContainers(ctx, imgConfig, containers)
	assert.Nil(t, responseErr)
	assert.True(t, changed)
	assert.Equal(t, "my-image:latest", newContainers[0].Image)
	assert.Equal(t, "other-image:v1.0.0", newContainers[1].Image)
}

func TestMutateImage(t *testing.T) {
	ctx := context.Background()
	imgConfig := ManagedImageConfig{
		Regex:              "my-image:.*",
		ReplacementPattern: "my-image:latest",
	}
	mutator, err := NewManagedImageMutator(&ManagedImagesConfig{Images: []ManagedImageConfig{imgConfig}})
	require.NoError(t, err)

	newImage := mutator.MutateImage(ctx, imgConfig, "my-image:v1.0.0")
	assert.Equal(t, "my-image:latest", newImage)
}

func TestMutateContainer_WithNodeSelectorAndTolerations(t *testing.T) {
	ctx := context.Background()
	imgConfig := ManagedImageConfig{
		Regex:              "my-image:.*",
		ReplacementPattern: "my-image:latest",
		NodeSelector: map[string]string{
			"key1": "value1",
		},
		Tolerations: []v1.Toleration{
			{
				Key:      "key2",
				Operator: v1.TolerationOpExists,
				Effect:   v1.TaintEffectNoSchedule,
			},
		},
	}
	mutator, err := NewManagedImageMutator(&ManagedImagesConfig{Images: []ManagedImageConfig{imgConfig}})
	require.NoError(t, err)

	pod := &v1.Pod{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{Name: "test-container", Image: "my-image:v1.0.0"},
			},
		},
	}

	newPod, changed, responseErr := mutator.Mutate(ctx, pod)
	assert.Nil(t, responseErr)
	assert.True(t, changed)
	assert.Equal(t, "my-image:latest", newPod.Spec.Containers[0].Image)
	assert.Equal(t, map[string]string{"key1": "value1"}, newPod.Spec.NodeSelector)
	assert.Len(t, newPod.Spec.Tolerations, 1)
	assert.Equal(t, "key2", newPod.Spec.Tolerations[0].Key)
	assert.Equal(t, v1.TolerationOpExists, newPod.Spec.Tolerations[0].Operator)
	assert.Equal(t, v1.TaintEffectNoSchedule, newPod.Spec.Tolerations[0].Effect)
}

func TestMutateContainer_NoNodeSelectorOrTolerations(t *testing.T) {
	ctx := context.Background()
	imgConfig := ManagedImageConfig{
		Regex:              "my-image:.*",
		ReplacementPattern: "my-image:latest",
	}
	mutator, err := NewManagedImageMutator(&ManagedImagesConfig{Images: []ManagedImageConfig{imgConfig}})
	require.NoError(t, err)

	pod := &v1.Pod{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{Name: "test-container", Image: "my-image:v1.0.0"},
			},
		},
	}

	newPod, changed, responseErr := mutator.Mutate(ctx, pod)
	assert.Nil(t, responseErr)
	assert.True(t, changed)
	assert.Equal(t, "my-image:latest", newPod.Spec.Containers[0].Image)
	assert.Empty(t, newPod.Spec.NodeSelector)
	assert.Empty(t, newPod.Spec.Tolerations)
}
