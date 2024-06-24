package secret

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	secretUtils "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/utils/secrets"
	"github.com/flyteorg/flyte/flytepropeller/pkg/secret/config"
	"github.com/flyteorg/flyte/flytepropeller/pkg/secret/mocks"
)

func TestSecretsWebhook_Mutate(t *testing.T) {
	t.Run("No injectors", func(t *testing.T) {
		m := SecretsPodMutator{}
		_, changed, err := m.Mutate(context.Background(), &corev1.Pod{})
		assert.Nil(t, err)
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

		m := SecretsPodMutator{
			enabledSecretManagerTypes: []config.SecretManagerType{config.SecretManagerTypeGlobal},
			injectors: map[config.SecretManagerType]SecretsInjector{
				config.SecretManagerTypeGlobal: mutator,
			},
		}

		_, changed, err := m.Mutate(context.Background(), podWithAnnotations.DeepCopy())
		assert.NotNil(t, err)
		assert.False(t, changed)
	})

	t.Run("added", func(t *testing.T) {
		mutator := &mocks.SecretsInjector{}
		mutator.OnInjectMatch(mock.Anything, mock.Anything, mock.Anything).Return(&corev1.Pod{}, true, nil)
		mutator.OnType().Return(config.SecretManagerTypeGlobal)

		m := SecretsPodMutator{
			enabledSecretManagerTypes: []config.SecretManagerType{config.SecretManagerTypeGlobal},
			injectors: map[config.SecretManagerType]SecretsInjector{
				config.SecretManagerTypeGlobal: mutator,
			},
		}

		_, changed, err := m.Mutate(context.Background(), podWithAnnotations.DeepCopy())
		assert.Nil(t, err)
		assert.True(t, changed)
	})
}

func TestSecrets_LabelSelector(t *testing.T) {
	m := SecretsPodMutator{}
	expected := metav1.LabelSelector{
		MatchLabels: map[string]string{
			secretUtils.PodLabel: secretUtils.PodLabelValue,
		},
	}
	assert.Equal(t, expected, *m.LabelSelector())
}
