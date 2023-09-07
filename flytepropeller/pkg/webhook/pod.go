// The PodMutator is a controller-runtime webhook that intercepts Pod Creation events and mutates them. Currently, there
// is only one registered Mutator, that's the SecretsMutator. It works as follows:
//
//   - The Webhook only works on Pods. If propeller/plugins launch a resource outside of K8s (or in a separate k8s
//     cluster), it's the responsibility of the plugin to correctly pass secret injection information.
//   - When a k8s-plugin builds a resource, propeller's PluginManager will automatically inject a label `inject-flyte
//     -secrets: true` and serialize the secret injection information into the annotations.
//   - If a plugin does not use the K8sPlugin interface, it's its responsibility to pass secret injection information.
//   - If a k8s plugin creates a CRD that launches other Pods (e.g. Spark/PyTorch... etc.), it's its responsibility to
//     make sure the labels/annotations set on the CRD by PluginManager are propagated to those launched Pods. This
//     ensures secret injection happens no matter how many levels of indirections there are.
//   - The Webhook expects 'inject-flyte-secrets: true' as a label on the Pod. Otherwise it won't listen/observe that pod.
//   - Once it intercepts the admission request, it goes over all registered Mutators and invoke them in the order they
//     are registered as. If a Mutator fails and it's marked as `required`, the operation will fail and the admission
//     will be rejected.
//   - The SecretsMutator will attempt to lookup the requested secret from the process environment. If the secret is
//     already mounted, it'll inject it as plain-text into the Pod Spec (Less secure).
//   - If it's not found in the environment it'll, instead, fallback to the enabled Secrets Injector (K8s, Confidant,
//     Vault... etc.).
//   - Each SecretsInjector will mutate the Pod differently depending on how its backend secrets system injects the secrets
//     for example:
//   - For K8s secrets, it'll either add EnvFromSecret or VolumeMountSource (depending on the MountRequirement
//     stated in the flyteIdl.Secret object) into the Pod. There is no validation that the secret exist and is available
//     to the Pod at this point. If the secret is not accessible, the Pod will fail with ContainerCreationConfigError and
//     will be retried.
//   - For Vault secrets, it'll inject the right annotations to trigger Vault's own sidecar/webhook to mount the secret.
package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/flyteorg/flytepropeller/pkg/webhook/config"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils/secrets"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/flyteorg/flytestdlib/promutils"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	corev1 "k8s.io/api/core/v1"
)

const webhookName = "flyte-pod-webhook.flyte.org"

// PodMutator implements controller-runtime WebHook interface.
type PodMutator struct {
	decoder  *admission.Decoder
	cfg      *config.Config
	Mutators []MutatorConfig
}

type MutatorConfig struct {
	Mutator  Mutator
	Required bool
}

type Mutator interface {
	ID() string
	Mutate(ctx context.Context, p *corev1.Pod) (newP *corev1.Pod, changed bool, err error)
}

func (pm *PodMutator) InjectClient(_ client.Client) error {
	return nil
}

// InjectDecoder injects the decoder into a mutatingHandler.
func (pm *PodMutator) InjectDecoder(d *admission.Decoder) error {
	pm.decoder = d
	return nil
}

func (pm *PodMutator) Handle(ctx context.Context, request admission.Request) admission.Response {
	// Get the object in the request
	obj := &corev1.Pod{}
	err := pm.decoder.Decode(request, obj)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	newObj, changed, err := pm.Mutate(ctx, obj)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if changed {
		marshalled, err := json.Marshal(newObj)
		if err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}

		// Create the patch
		return admission.PatchResponseFromRaw(request.Object.Raw, marshalled)
	}

	return admission.Allowed("No changes")
}

func (pm PodMutator) Mutate(ctx context.Context, p *corev1.Pod) (newP *corev1.Pod, changed bool, err error) {
	newP = p
	for _, m := range pm.Mutators {
		tempP := newP
		tempChanged := false
		tempP, tempChanged, err = m.Mutator.Mutate(ctx, tempP)
		if err != nil {
			if m.Required {
				err = fmt.Errorf("failed to mutate using [%v]. Since it's a required mutator, failing early. Error: %v", m.Mutator.ID(), err)
				logger.Info(ctx, err)
				return p, false, err
			}

			logger.Infof(ctx, "Failed to mutate using [%v]. Since it's not a required mutator, skipping. Error: %v", m.Mutator.ID(), err)
			continue
		}

		newP = tempP
		if tempChanged {
			changed = true
		}
	}

	return newP, changed, nil
}

func (pm *PodMutator) Register(ctx context.Context, mgr manager.Manager) error {
	wh := &admission.Webhook{
		Handler: pm,
	}

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
	caBytes, err := ioutil.ReadFile(filepath.Join(pm.cfg.CertDir, "ca.crt"))
	if err != nil {
		// ca.crt is optional. If not provided, API Server will assume the webhook is serving SSL using a certificate
		// issued by a known Cert Authority.
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
				Name: webhookName,
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					CABundle: caBytes, // CA bundle created earlier
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
				FailurePolicy: &fail,
				SideEffects:   &sideEffects,
				AdmissionReviewVersions: []string{
					"v1",
					"v1beta1",
				},
				ObjectSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						secrets.PodLabel: secrets.PodLabelValue,
					},
				},
			}},
	}

	return mutateConfig, nil
}

func NewPodMutator(cfg *config.Config, scope promutils.Scope) *PodMutator {
	return &PodMutator{
		cfg: cfg,
		Mutators: []MutatorConfig{
			{
				Mutator: NewSecretsMutator(cfg, scope.NewSubScope("secrets")),
			},
		},
	}
}
