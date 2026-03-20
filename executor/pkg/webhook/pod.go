// Package webhook contains PodMutator. It's a controller-runtime webhook that intercepts Pod Creation events and
// mutates them. The SecretsMutator injects secret references into pods that have the inject-flyte-secrets label.
package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret"
	webhookConfig "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret/config"
	secretUtils "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/utils/secrets"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
)

const webhookName = "flyte-pod-webhook.flyte.org"

// PodMutator implements controller-runtime WebHook interface.
type PodMutator struct {
	decoder        admission.Decoder
	cfg            *webhookConfig.Config
	secretsMutator *secret.SecretsPodMutator
}

func (pm PodMutator) Handle(ctx context.Context, request admission.Request) admission.Response {
	obj := &corev1.Pod{}
	err := pm.decoder.Decode(request, obj)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	newObj, changed, admissionErr := pm.secretsMutator.Mutate(ctx, obj)
	if admissionErr != nil {
		return *admissionErr
	}

	if changed {
		marshalled, err := json.Marshal(newObj)
		if err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}
		return admission.PatchResponseFromRaw(request.Object.Raw, marshalled)
	}

	return admission.Allowed("No changes")
}

func (pm PodMutator) Register(ctx context.Context, mgr manager.Manager) error {
	wh := &admission.Webhook{Handler: pm}
	mutatePath := getPodMutatePath()
	logger.Infof(ctx, "Registering path [%v]", mutatePath)
	mgr.GetWebhookServer().Register(mutatePath, wh)
	return nil
}

func (pm PodMutator) GetMutatePath() string {
	return getPodMutatePath()
}

func getPodMutatePath() string {
	pod := flytek8s.BuildIdentityPod()
	return generateMutatePath(pod.GroupVersionKind())
}

func generateMutatePath(gvk schema.GroupVersionKind) string {
	return "/mutate-" + strings.Replace(gvk.Group, ".", "-", -1) + "-" +
		gvk.Version + "-" + strings.ToLower(gvk.Kind)
}

func (pm PodMutator) CreateMutationWebhookConfiguration(namespace string) (*admissionregistrationv1.MutatingWebhookConfiguration, error) {
	caBytes, err := os.ReadFile(filepath.Join(pm.cfg.ExpandCertDir(), "ca.crt"))
	if err != nil {
		if os.IsNotExist(err) {
			caBytes = make([]byte, 0)
		} else {
			return nil, err
		}
	}

	path := pm.GetMutatePath()
	fail := admissionregistrationv1.Fail
	sideEffects := admissionregistrationv1.SideEffectClassNoneOnDryRun

	mutateConfig := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pm.cfg.ServiceName,
			Namespace: namespace,
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{
			{
				Name:         webhookName,
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					CABundle: caBytes,
					Service: &admissionregistrationv1.ServiceReference{
						Name:      pm.cfg.ServiceName,
						Namespace: namespace,
						Path:      &path,
						Port:      &pm.cfg.ServicePort,
					},
				},
				Rules: []admissionregistrationv1.RuleWithOperations{
					{
						Operations: []admissionregistrationv1.OperationType{
							admissionregistrationv1.Create,
						},
						Rule: admissionregistrationv1.Rule{
							APIGroups:   []string{"*"},
							APIVersions: []string{"v1"},
							Resources:   []string{"pods"},
						},
					},
				},
				FailurePolicy:           &fail,
				SideEffects:             &sideEffects,
				AdmissionReviewVersions: []string{"v1", "v1beta1"},
				ObjectSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						secretUtils.PodLabel: secretUtils.PodLabelValue,
					},
				},
			},
		},
	}

	return mutateConfig, nil
}

func NewPodMutator(ctx context.Context, cfg *webhookConfig.Config, podNamespace string, scheme *runtime.Scheme, scope promutils.Scope) (*PodMutator, error) {
	secretsMutator, err := secret.NewSecretsMutator(ctx, cfg, podNamespace, scope.NewSubScope("secrets"))
	if err != nil {
		return nil, fmt.Errorf("failed to create secrets mutator: %w", err)
	}

	return &PodMutator{
		decoder:        *admission.NewDecoder(scheme),
		cfg:            cfg,
		secretsMutator: secretsMutator,
	}, nil
}
