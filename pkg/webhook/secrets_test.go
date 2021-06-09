package webhook

import (
	"context"
	"fmt"
	"testing"

	"github.com/flyteorg/flytepropeller/pkg/webhook/config"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flytepropeller/pkg/webhook/mocks"
	"github.com/stretchr/testify/mock"

	corev1 "k8s.io/api/core/v1"
)

func TestSecretsWebhook_Mutate(t *testing.T) {
	t.Run("No injectors", func(t *testing.T) {
		m := SecretsMutator{}
		_, changed, err := m.Mutate(context.Background(), &corev1.Pod{})
		assert.NoError(t, err)
		assert.False(t, changed)
	})

	podWithAnnotations := &corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Annotations: map[string]string{
				"flyte.secrets/s0": "nnsxsorcnv4v623fperca",
			},
		},
	}

	t.Run("First fail", func(t *testing.T) {
		mutator := &mocks.SecretsInjector{}
		mutator.OnInjectMatch(mock.Anything, mock.Anything, mock.Anything).Return(nil, false, fmt.Errorf("failed"))
		mutator.OnType().Return(config.SecretManagerTypeGlobal)

		m := SecretsMutator{
			injectors: []SecretsInjector{mutator},
		}

		_, changed, err := m.Mutate(context.Background(), podWithAnnotations.DeepCopy())
		assert.Error(t, err)
		assert.False(t, changed)
	})

	t.Run("added", func(t *testing.T) {
		mutator := &mocks.SecretsInjector{}
		mutator.OnInjectMatch(mock.Anything, mock.Anything, mock.Anything).Return(&corev1.Pod{}, true, nil)
		mutator.OnType().Return(config.SecretManagerTypeGlobal)

		m := SecretsMutator{
			injectors: []SecretsInjector{mutator},
		}

		_, changed, err := m.Mutate(context.Background(), podWithAnnotations.DeepCopy())
		assert.NoError(t, err)
		assert.True(t, changed)
	})
}
