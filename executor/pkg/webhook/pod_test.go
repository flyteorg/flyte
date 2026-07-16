package webhook

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	webhookConfig "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret/config"
)

func TestCreateMutationWebhookConfigurationNamespaceSelector(t *testing.T) {
	t.Run("nil selector matches all namespaces", func(t *testing.T) {
		pm := PodMutator{cfg: &webhookConfig.Config{ServiceName: "flyte-pod-webhook", ServicePort: 443, CertDir: t.TempDir()}}

		mutateConfig, err := pm.CreateMutationWebhookConfiguration("flyte")

		assert.NoError(t, err)
		assert.Nil(t, mutateConfig.Webhooks[0].NamespaceSelector)
	})

	t.Run("selector is propagated", func(t *testing.T) {
		selector := &metav1.LabelSelector{
			MatchLabels: map[string]string{"kubernetes.io/metadata.name": "flyte"},
		}
		pm := PodMutator{cfg: &webhookConfig.Config{ServiceName: "flyte-pod-webhook", ServicePort: 443, CertDir: t.TempDir(), NamespaceSelector: selector}}

		mutateConfig, err := pm.CreateMutationWebhookConfiguration("flyte")

		assert.NoError(t, err)
		assert.Equal(t, selector, mutateConfig.Webhooks[0].NamespaceSelector)
	})
}
