package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/flytepropeller/pkg/secret"
	"github.com/flyteorg/flyte/flytepropeller/pkg/secret/config"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

const (
	secretsWebhookName = "flyte-pod-webhook.flyte.org" // #nosec G101
)

//go:generate mockery --output=./mocks --case=underscore --name=PodMutator

// PodMutator contains the business logic for a unique type of mutation or validation.
type PodMutator interface {
	ID() string
	// Conducts the act of mutating the pod.
	Mutate(ctx context.Context, p *corev1.Pod) (newP *corev1.Pod, changed bool, err *admission.Response)
	// Defines how to select which Pods to apply the webhook to.
	LabelSelector() *metav1.LabelSelector
}

// PodHandler handles webhook mutations for Pod resources
type PodHandler struct {
	decoder             *admission.Decoder
	mutator             PodMutator
	mutatingWebhookName string
	path                string
}

// NewPodHandler creates a new pod handler for the given mutator
func NewPodHandler(decoder *admission.Decoder, mutator PodMutator) *PodHandler {
	return &PodHandler{
		decoder:             decoder,
		mutator:             mutator,
		mutatingWebhookName: getMutatingWebhookName(mutator.ID()),
		path:                getPodMutatePath(mutator.ID()),
	}
}

// GetMutatingWebhook returns the Kubernetes MutatingWebhook configuration for pods
func (h *PodHandler) GetMutatingWebhook(namespace string, caBytes []byte, cfg *config.Config) admissionregistrationv1.MutatingWebhook {
	return admissionregistrationv1.MutatingWebhook{
		Name: h.mutatingWebhookName,
		ClientConfig: admissionregistrationv1.WebhookClientConfig{
			CABundle: caBytes,
			Service: &admissionregistrationv1.ServiceReference{
				Name:      cfg.ServiceName,
				Namespace: namespace,
				Path:      &h.path,
				Port:      &cfg.ServicePort,
			},
		},
		Rules:                   getPodAdmissionRules(),
		TimeoutSeconds:          &cfg.WebhookTimeout,
		FailurePolicy:           &admissionRegistrationFailurePolicy,
		SideEffects:             &admissionRegistrationSideEffects,
		AdmissionReviewVersions: admissionRegistrationVersions,
		ObjectSelector:          h.mutator.LabelSelector(),
	}
}

// GetPath returns the webhook path
func (h *PodHandler) GetPath() string {
	return h.path
}

// GetAdmissionHandler returns the controller-runtime webhook handler
func (h *PodHandler) GetAdmissionHandler() admission.Handler {
	return h
}

// GetMutator returns the underlying pod mutator for testing
func (h *PodHandler) GetMutator() PodMutator {
	return h.mutator
}

// Handle processes the admission request for pod mutations
func (h *PodHandler) Handle(ctx context.Context, request admission.Request) admission.Response {
	ctx = contextutils.WithRequestID(ctx, rand.String(10))

	// Get the object in the request
	obj := &corev1.Pod{}
	err := h.decoder.Decode(request, obj)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	newObj, changed, admissionError := h.mutator.Mutate(ctx, obj)
	if admissionError != nil {
		return *admissionError
	}

	if changed {
		logger.Infof(ctx, "Mutated Pod [%v/%v]", obj.Namespace, obj.Name)
		marshalled, err := json.Marshal(newObj)
		if err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}

		// Create the patch
		return admission.PatchResponseFromRaw(request.Object.Raw, marshalled)
	}

	return admission.Allowed("No changes")
}

// initializePodHandlers creates all pod-specific webhook handlers.
// It also returns the SecretsPodMutator so the caller can use it for cache invalidation.
func initializePodHandlers(ctx context.Context, cfg *config.Config, scheme *runtime.Scheme, podNamespace string, scope promutils.Scope) ([]ResourceHandler, *secret.SecretsPodMutator, error) {
	var handlers []ResourceHandler

	decoder := admission.NewDecoder(scheme)

	// Secrets mutator (always enabled). It works as follows:
	//   - The Webhook only works on Pods. If propeller/plugins launch a resource outside K8s (or in a separate k8s
	//     cluster), it's the responsibility of the plugin to correctly pass secret injection information.
	//   - When a k8s-plugin builds a resource, propeller's PluginManager will automatically inject a label `inject-flyte
	//     -secrets: true` and serialize the secret injection information into the annotations.
	//   - If a plugin does not use the K8sPlugin interface, it's its responsibility to pass secret injection information.
	//   - If a k8s plugin creates a CRD that launches other Pods (e.g. Spark/PyTorch... etc.), it's its responsibility to
	//     make sure the labels/annotations set on the CRD by PluginManager are propagated to those launched Pods. This
	//     ensures secret injection happens no matter how many levels of indirections there are.
	//   - The Webhook expects 'inject-flyte-secrets: true' as a label on the Pod. Otherwise, it won't listen/observe that
	//     pod.
	//   - Once it intercepts the admission request, it goes over all registered Mutators and invoke them in the order they
	//     are registered as. If a Mutator fails, and it's marked as `required`, the operation will fail and the admission
	//     will be rejected.
	//   - The SecretsMutator will attempt to look up the requested secret from the process environment. If the secret is
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
	secretsMutator, err := secret.NewSecretsMutator(ctx, cfg, podNamespace, scope.NewSubScope("secrets"))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create secrets mutator: %w", err)
	}
	secretsHandler := NewPodHandler(decoder, secretsMutator)
	secretsHandler.mutatingWebhookName = secretsWebhookName
	handlers = append(handlers, secretsHandler)

	// Image builder mutator (conditional)
	if cfg.ImageBuilderConfig.Enabled {
		imageBuilderMutator := NewImageBuilderMutator(&cfg.ImageBuilderConfig, scope.NewSubScope("image-builder"))
		handlers = append(handlers, NewPodHandler(decoder, imageBuilderMutator))
	}

	// Managed images mutator (conditional)
	managedImagesCfg := GetManagedImagesConfig()
	if managedImagesCfg.Enabled {
		logger.Infof(ctx, "Enabling Managed Image Mutator")
		mutator, err := NewManagedImageMutator(managedImagesCfg)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create managed image mutator: %w", err)
		}
		handlers = append(handlers, NewPodHandler(decoder, mutator))
	} else {
		logger.Infof(ctx, "Managed Image Mutator is disabled")
	}

	return handlers, secretsMutator, nil
}

// getPodAdmissionRules returns the admission rules for pod webhooks
func getPodAdmissionRules() []admissionregistrationv1.RuleWithOperations {
	return []admissionregistrationv1.RuleWithOperations{
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
	}
}

func getPodMutatePath(subpath string) string {
	pod := flytek8s.BuildIdentityPod()
	return generateMutatePath(pod.GroupVersionKind(), subpath)
}
