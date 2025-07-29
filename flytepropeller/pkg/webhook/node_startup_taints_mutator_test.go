package webhook

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNodeStartupTaintsMutator_ID(t *testing.T) {
	mutator := NodeStartupTaintsMutator{}
	assert.Equal(t, NodeStartupTaintsMutatorID, mutator.ID())
}

func TestNodeStartupTaintsMutator_Mutate_Disabled(t *testing.T) {
	ctx := context.Background()
	cfg := &NodeStartupTaintsConfig{
		Enabled: false,
	}
	mutator, err := NewNodeStartupTaintsMutator(cfg)
	require.NoError(t, err)

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
			Labels: map[string]string{
				"flyte.org/node-role": "worker",
			},
		},
	}

	newNode, changed, responseErr := mutator.Mutate(ctx, node)
	assert.Nil(t, responseErr)
	assert.False(t, changed)
	assert.Equal(t, node, newNode)
}

func TestNodeStartupTaintsMutator_Mutate_NoMatchingLabel(t *testing.T) {
	ctx := context.Background()
	cfg := &NodeStartupTaintsConfig{
		Enabled: true,
		StartupTaints: []corev1.Taint{
			{
				Key:    "startup-taint.cluster-autoscaler.kubernetes.io/unionai-node-not-ready",
				Value:  "true",
				Effect: corev1.TaintEffectNoSchedule,
			},
		},
		LabelSelector: metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "flyte.org/node-role",
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{"worker"},
				},
			},
		},
	}
	mutator, err := NewNodeStartupTaintsMutator(cfg)
	require.NoError(t, err)

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
			Labels: map[string]string{
				"flyte.org/node-role": "master", // Different value
			},
		},
	}

	newNode, changed, responseErr := mutator.Mutate(ctx, node)
	assert.Nil(t, responseErr)
	assert.False(t, changed)
	assert.Equal(t, node, newNode)
}

func TestNodeStartupTaintsMutator_Mutate_NoLabels(t *testing.T) {
	ctx := context.Background()
	cfg := &NodeStartupTaintsConfig{
		Enabled: true,
		StartupTaints: []corev1.Taint{
			{
				Key:    "startup-taint.cluster-autoscaler.kubernetes.io/unionai-node-not-ready",
				Value:  "true",
				Effect: corev1.TaintEffectNoSchedule,
			},
		},
		LabelSelector: metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "flyte.org/node-role",
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{"worker"},
				},
			},
		},
	}
	mutator, err := NewNodeStartupTaintsMutator(cfg)
	require.NoError(t, err)

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
			// No labels
		},
	}

	newNode, changed, responseErr := mutator.Mutate(ctx, node)
	assert.Nil(t, responseErr)
	assert.False(t, changed)
	assert.Equal(t, node, newNode)
}

func TestNodeStartupTaintsMutator_Mutate_AddsTaint(t *testing.T) {
	ctx := context.Background()
	cfg := &NodeStartupTaintsConfig{
		Enabled: true,
		StartupTaints: []corev1.Taint{
			{
				Key:    "startup-taint.cluster-autoscaler.kubernetes.io/unionai-node-not-ready",
				Value:  "true",
				Effect: corev1.TaintEffectNoSchedule,
			},
		},
		LabelSelector: metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "flyte.org/node-role",
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{"worker"},
				},
			},
		},
	}
	mutator, err := NewNodeStartupTaintsMutator(cfg)
	require.NoError(t, err)

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
			Labels: map[string]string{
				"flyte.org/node-role": "worker",
			},
		},
		Spec: corev1.NodeSpec{
			Taints: []corev1.Taint{
				{
					Key:    "existing-taint",
					Value:  "existing-value",
					Effect: corev1.TaintEffectNoExecute,
				},
			},
		},
	}

	newNode, changed, responseErr := mutator.Mutate(ctx, node)
	assert.Nil(t, responseErr)
	assert.True(t, changed)

	// Should have both the existing taint and the new startup taint
	assert.Len(t, newNode.Spec.Taints, 2)

	// Check that existing taint is preserved
	assert.Contains(t, newNode.Spec.Taints, corev1.Taint{
		Key:    "existing-taint",
		Value:  "existing-value",
		Effect: corev1.TaintEffectNoExecute,
	})

	// Check that new taint is added
	assert.Contains(t, newNode.Spec.Taints, corev1.Taint{
		Key:    "startup-taint.cluster-autoscaler.kubernetes.io/unionai-node-not-ready",
		Value:  "true",
		Effect: corev1.TaintEffectNoSchedule,
	})
}

func TestNodeStartupTaintsMutator_Mutate_TaintAlreadyExists(t *testing.T) {
	ctx := context.Background()
	cfg := &NodeStartupTaintsConfig{
		Enabled: true,
		StartupTaints: []corev1.Taint{
			{
				Key:    "startup-taint.cluster-autoscaler.kubernetes.io/unionai-node-not-ready",
				Value:  "true",
				Effect: corev1.TaintEffectNoSchedule,
			},
		},
		LabelSelector: metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "flyte.org/node-role",
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{"worker"},
				},
			},
		},
	}
	mutator, err := NewNodeStartupTaintsMutator(cfg)
	require.NoError(t, err)

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
			Labels: map[string]string{
				"flyte.org/node-role": "worker",
			},
		},
		Spec: corev1.NodeSpec{
			Taints: []corev1.Taint{
				{
					Key:    "startup-taint.cluster-autoscaler.kubernetes.io/unionai-node-not-ready",
					Value:  "true",
					Effect: corev1.TaintEffectNoSchedule,
				},
			},
		},
	}

	newNode, changed, responseErr := mutator.Mutate(ctx, node)
	assert.Nil(t, responseErr)
	assert.False(t, changed)
	assert.Equal(t, node, newNode)
	assert.Len(t, newNode.Spec.Taints, 1) // Should still have only the original taint
}

func TestNodeStartupTaintsMutator_Mutate_WithCustomTaintConfig(t *testing.T) {
	ctx := context.Background()
	cfg := &NodeStartupTaintsConfig{
		Enabled: true,
		StartupTaints: []corev1.Taint{
			{
				Key:    "custom.example.com/not-ready",
				Value:  "custom-value",
				Effect: corev1.TaintEffectNoExecute,
			},
		},
		LabelSelector: metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "custom.example.com/managed",
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{"enabled"},
				},
			},
		},
	}
	mutator, err := NewNodeStartupTaintsMutator(cfg)
	require.NoError(t, err)

	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
			Labels: map[string]string{
				"custom.example.com/managed": "enabled",
			},
		},
		Spec: corev1.NodeSpec{
			Taints: []corev1.Taint{},
		},
	}

	newNode, changed, responseErr := mutator.Mutate(ctx, node)
	assert.Nil(t, responseErr)
	assert.True(t, changed)

	assert.Len(t, newNode.Spec.Taints, 1)
	assert.Equal(t, corev1.Taint{
		Key:    "custom.example.com/not-ready",
		Value:  "custom-value",
		Effect: corev1.TaintEffectNoExecute,
	}, newNode.Spec.Taints[0])
}

func TestNodeStartupTaintsMutator_LabelSelector(t *testing.T) {
	cfg := &NodeStartupTaintsConfig{
		LabelSelector: metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "flyte.org/node-role",
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{"worker"},
				},
				{
					Key:      "environment",
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{"production"},
				},
			},
		},
	}
	mutator, err := NewNodeStartupTaintsMutator(cfg)
	require.NoError(t, err)

	labelSelector := mutator.LabelSelector()
	assert.Equal(t, &cfg.LabelSelector, labelSelector)
}

func TestGetNodeStartupTaintsConfig(t *testing.T) {
	cfg := GetNodeStartupTaintsConfig()
	assert.NotNil(t, cfg)
	assert.Equal(t, false, cfg.Enabled) // Default should be disabled
	assert.Equal(t, "startup-taint.cluster-autoscaler.kubernetes.io/unionai-node-not-ready", cfg.StartupTaints[0].Key)
	assert.Equal(t, "true", cfg.StartupTaints[0].Value)
	assert.Equal(t, corev1.TaintEffectNoSchedule, cfg.StartupTaints[0].Effect)
}
