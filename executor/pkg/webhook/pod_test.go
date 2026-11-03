package webhook

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	webhookConfig "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
)

func newTestWebhook(t *testing.T, cfg *webhookConfig.Config, opts ...Option) *Webhook {
	t.Helper()
	cfg.CertDir = t.TempDir()
	if cfg.ServiceName == "" {
		cfg.ServiceName = "flyte-pod-webhook"
	}
	if cfg.ServicePort == 0 {
		cfg.ServicePort = 443
	}
	w, err := NewWebhook(context.Background(), cfg, "flyte", clientgoscheme.Scheme, promutils.NewTestScope(), opts...)
	require.NoError(t, err)
	return w
}

func TestCreateMutationWebhookConfigurationNamespaceSelector(t *testing.T) {
	t.Run("nil selector matches all namespaces", func(t *testing.T) {
		w := newTestWebhook(t, &webhookConfig.Config{})

		mutateConfig, err := w.CreateMutationWebhookConfiguration("flyte")

		assert.NoError(t, err)
		assert.Nil(t, mutateConfig.Webhooks[0].NamespaceSelector)
	})

	t.Run("selector is propagated", func(t *testing.T) {
		selector := &metav1.LabelSelector{
			MatchLabels: map[string]string{"kubernetes.io/metadata.name": "flyte"},
		}
		w := newTestWebhook(t, &webhookConfig.Config{NamespaceSelector: selector})

		mutateConfig, err := w.CreateMutationWebhookConfiguration("flyte")

		assert.NoError(t, err)
		assert.Equal(t, selector, mutateConfig.Webhooks[0].NamespaceSelector)
	})
}

func TestSecretsHandlerPaths(t *testing.T) {
	w := newTestWebhook(t, &webhookConfig.Config{})

	require.Len(t, w.Handlers(), 1)
	// The legacy path stays first so generated configurations are unchanged.
	assert.Equal(t, []string{"/mutate--v1-pod", "/mutate--v1-pod/secrets"}, w.Handlers()[0].Paths())

	mutateConfig, err := w.CreateMutationWebhookConfiguration("flyte")
	require.NoError(t, err)
	require.Len(t, mutateConfig.Webhooks, 1)
	assert.Equal(t, secretsWebhookName, mutateConfig.Webhooks[0].Name)
	assert.Equal(t, "/mutate--v1-pod", *mutateConfig.Webhooks[0].ClientConfig.Service.Path)
}

func TestWebhookTimeout(t *testing.T) {
	w := newTestWebhook(t, &webhookConfig.Config{WebhookTimeout: 7})

	mutateConfig, err := w.CreateMutationWebhookConfiguration("flyte")

	require.NoError(t, err)
	require.NotNil(t, mutateConfig.Webhooks[0].TimeoutSeconds)
	assert.Equal(t, int32(7), *mutateConfig.Webhooks[0].TimeoutSeconds)
}

type labelPod struct{ id string }

func (m labelPod) ID() string { return m.id }

func (m labelPod) Mutate(_ context.Context, p *corev1.Pod) (*corev1.Pod, bool, *admission.Response) {
	p = p.DeepCopy()
	if p.Labels == nil {
		p.Labels = map[string]string{}
	}
	p.Labels["mutated-by"] = m.id
	return p, true, nil
}

func (m labelPod) LabelSelector() *metav1.LabelSelector {
	return &metav1.LabelSelector{MatchLabels: map[string]string{"select": m.id}}
}

type noopNode struct{}

func (noopNode) ID() string { return "taints" }
func (noopNode) Mutate(_ context.Context, n *corev1.Node) (*corev1.Node, bool, *admission.Response) {
	return n, false, nil
}
func (noopNode) LabelSelector() *metav1.LabelSelector { return nil }

func TestExtraMutators(t *testing.T) {
	w := newTestWebhook(t, &webhookConfig.Config{},
		WithPodMutators(labelPod{id: "selinux"}),
		WithNodeMutators(noopNode{}))

	require.Len(t, w.Handlers(), 3)
	assert.Equal(t, []string{"/mutate--v1-pod/selinux"}, w.Handlers()[1].Paths())
	assert.Equal(t, []string{"/mutate--v1-node/taints"}, w.Handlers()[2].Paths())

	mutateConfig, err := w.CreateMutationWebhookConfiguration("flyte")
	require.NoError(t, err)
	require.Len(t, mutateConfig.Webhooks, 3)

	pod := mutateConfig.Webhooks[1]
	assert.Equal(t, "selinux.flyte.org", pod.Name)
	assert.Equal(t, map[string]string{"select": "selinux"}, pod.ObjectSelector.MatchLabels)
	assert.Equal(t, []string{"pods"}, pod.Rules[0].Resources)

	node := mutateConfig.Webhooks[2]
	assert.Equal(t, []string{"nodes"}, node.Rules[0].Resources)
	assert.Equal(t, []string{""}, node.Rules[0].APIGroups)
	assert.Nil(t, node.NamespaceSelector)
}

func TestCustomHandlerNameAndDuplicatePaths(t *testing.T) {
	decoder := admission.NewDecoder(clientgoscheme.Scheme)
	custom := NewPodHandler(decoder, labelPod{id: "managed-image"}, WithWebhookName("managed-image-webhook.example.com"))
	w := newTestWebhook(t, &webhookConfig.Config{}, WithHandlers(custom))

	mutateConfig, err := w.CreateMutationWebhookConfiguration("flyte")
	require.NoError(t, err)
	assert.Equal(t, "managed-image-webhook.example.com", mutateConfig.Webhooks[1].Name)

	cfg := &webhookConfig.Config{CertDir: t.TempDir()}
	_, err = NewWebhook(context.Background(), cfg, "flyte", clientgoscheme.Scheme, promutils.NewTestScope(),
		WithPodMutators(labelPod{id: "a"}, labelPod{id: "a"}))
	assert.ErrorContains(t, err, "duplicate path")
}

func TestHandlerPatchesPod(t *testing.T) {
	decoder := admission.NewDecoder(clientgoscheme.Scheme)
	h := NewPodHandler(decoder, labelPod{id: "x"})

	raw, err := json.Marshal(&corev1.Pod{
		TypeMeta:   metav1.TypeMeta{Kind: "Pod", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "p", Namespace: "n"},
	})
	require.NoError(t, err)

	req := admission.Request{AdmissionRequest: admissionv1.AdmissionRequest{
		Object: runtime.RawExtension{Raw: raw},
	}}
	resp := h.AdmissionHandler().Handle(context.Background(), req)

	assert.True(t, resp.Allowed)
	require.Len(t, resp.Patches, 1)
	assert.Equal(t, "/metadata/labels", resp.Patches[0].Path)
}
