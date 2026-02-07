package secret

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret/config"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret/mocks"
)

func TestSecretsWebhook_Mutate(t *testing.T) {
	t.Run("No injectors", func(t *testing.T) {
		m := SecretsPodMutator{}
		_, changed, err := m.Mutate(context.Background(), &corev1.Pod{})
		assert.Nil(t, err)
		assert.False(t, changed)
	})

	namespace := "test-namespace"
	podWithAnnotations := &corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Namespace: namespace,
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
		ctx := context.Background()

		m := SecretsPodMutator{
			enabledSecretManagerTypes: []config.SecretManagerType{config.SecretManagerTypeGlobal},
			injectors: map[config.SecretManagerType]SecretsInjector{
				config.SecretManagerTypeGlobal: mutator,
			},
		}

		_, changed, err := m.Mutate(ctx, podWithAnnotations.DeepCopy())
		assert.Nil(t, err)
		assert.True(t, changed)
	})
}
