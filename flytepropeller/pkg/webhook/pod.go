// Package webhook container PodMutator. It's a controller-runtime webhook that intercepts Pod Creation events and
// mutates them. Currently, there is only one registered Mutator, that's the SecretsMutator. It works as follows:
//
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
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/flytepropeller/pkg/secret"
	"github.com/flyteorg/flyte/flytepropeller/pkg/secret/config"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

const (
	secretsWebhookName     = "flyte-pod-webhook.flyte.org" // #nosec G101
	unionWebhookNameDomain = "union.ai"
)

var (
	admissionRegistrationRules = []admissionregistrationv1.RuleWithOperations{
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
	admissionRegistrationVersions = []string{
		"v1",
		"v1beta1",
	}
	admissionRegistrationFailurePolicy = admissionregistrationv1.Fail
	admissionRegistrationSideEffects   = admissionregistrationv1.SideEffectClassNoneOnDryRun
)

// PodCreationWebhookConfig maps one to one to Kubernetes MutatingWebhookConfiguration
// but specifically tagetting Pod creation. Kubernetes MutatingWebhookConfiguration supports
// multiple webhooks. This class is responsible for converting Union specific configuration into a
// Kubernetes MutatingWebhookConfiguration with potentially multiple webhooks.
type PodCreationWebhookConfig struct {
	cfg          *config.Config
	httpHandlers []httpHandler
	caBytes      []byte
}

// Internal struct type to consolidate handling of the HTTP layer.
// Every http handler shares the same decode, encoding logic but has a different Mutator.
type httpHandler struct {
	decoder *admission.Decoder
	mutator PodMutator
	// The unique name of the Mutating Webhook
	mutatingWebhookName string
	// The complete URI Path to register webhook with.
	path string
}

//go:generate mockery --output=./mocks --case=underscore --name=PodMutator

// PodMutator contains the business logic for a unique type of mutation or validation.
type PodMutator interface {
	ID() string
	// Conducts the act of mutating the pod.
	Mutate(ctx context.Context, p *corev1.Pod) (newP *corev1.Pod, changed bool, err *admission.Response)
	// Defines how to select which Pods to apply the webhook to.
	LabelSelector() *metav1.LabelSelector
}

func (h httpHandler) Handle(ctx context.Context, request admission.Request) admission.Response {
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
		marshalled, err := json.Marshal(newObj)
		if err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}

		// Create the patch
		return admission.PatchResponseFromRaw(request.Object.Raw, marshalled)
	}

	return admission.Allowed("No changes")
}

func (pm PodCreationWebhookConfig) Register(ctx context.Context, registerer HTTPHookRegistererIface) error {
	for _, httpHandler := range pm.httpHandlers {
		wh := &admission.Webhook{Handler: httpHandler}
		logger.Infof(ctx, "Registering path [%v]", httpHandler.path)
		registerer.Register(httpHandler.path, wh)
	}
	return nil
}

func getPodMutatePath(subpath string) string {
	pod := flytek8s.BuildIdentityPod()
	return generateMutatePath(pod.GroupVersionKind(), subpath)
}

func generateMutatePath(gvk schema.GroupVersionKind, subpath string) string {
	return "/mutate-" + strings.Replace(gvk.Group, ".", "-", -1) + "-" +
		gvk.Version + "-" + strings.ToLower(gvk.Kind) + "/" + subpath
}

func (pm PodCreationWebhookConfig) CreateMutationWebhookConfiguration(namespace string) (*admissionregistrationv1.MutatingWebhookConfiguration, error) {
	webhooks := make([]admissionregistrationv1.MutatingWebhook, 0, len(pm.httpHandlers))
	for _, httpHandler := range pm.httpHandlers {
		webhooks = append(webhooks, pm.getMutatingWebhook(namespace, httpHandler))
	}

	mutateConfig := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pm.cfg.ServiceName,
			Namespace: namespace,
		},
		Webhooks: webhooks,
	}

	return mutateConfig, nil
}

func NewPodCreationWebhookConfig(ctx context.Context, cfg *config.Config, scheme *runtime.Scheme, scope promutils.Scope) (*PodCreationWebhookConfig, error) {
	caBytes, err := os.ReadFile(filepath.Join(cfg.ExpandCertDir(), "ca.crt"))
	if err != nil {
		// ca.crt is optional. If not provided, API Server will assume the webhook is serving SSL using a certificate
		// issued by a known Cert Authority.
		if os.IsNotExist(err) {
			caBytes = make([]byte, 0)
		} else {
			return nil, err
		}
	}

	secretsMutator, err := secret.NewSecretsMutator(ctx, cfg, scope.NewSubScope("secrets"))
	if err != nil {
		return nil, err
	}

	decoder := admission.NewDecoder(scheme)

	httpHandlers := []httpHandler{
		{
			decoder:             decoder,
			mutator:             secretsMutator,
			mutatingWebhookName: secretsWebhookName,
			path:                getPodMutatePath(secretsMutator.ID()),
		},
	}

	if cfg.ImageBuilderConfig.Enabled {
		imageBuilderMutator := NewImageBuilderMutator(&cfg.ImageBuilderConfig, scope.NewSubScope("image-builder"))
		httpHandlers = append(httpHandlers, httpHandler{
			decoder:             decoder,
			mutator:             imageBuilderMutator,
			mutatingWebhookName: getMutatingWebhookName(imageBuilderMutator.ID()),
			path:                getPodMutatePath(imageBuilderMutator.ID()),
		})
	}

	err = verifyHTTPHandlers(ctx, httpHandlers)
	if err != nil {
		return nil, err
	}

	return &PodCreationWebhookConfig{
		cfg:          cfg,
		httpHandlers: httpHandlers,
		caBytes:      caBytes,
	}, nil
}

func getMutatingWebhookName(id string) string {
	return fmt.Sprintf("%s-webhook.%s", id, unionWebhookNameDomain)
}

func (pm PodCreationWebhookConfig) getMutatingWebhook(namespace string, httpHandler httpHandler) admissionregistrationv1.MutatingWebhook {
	return admissionregistrationv1.MutatingWebhook{
		Name: httpHandler.mutatingWebhookName,
		ClientConfig: admissionregistrationv1.WebhookClientConfig{
			CABundle: pm.caBytes, // CA bundle created earlier
			Service: &admissionregistrationv1.ServiceReference{
				Name:      pm.cfg.ServiceName,
				Namespace: namespace,
				Path:      &httpHandler.path,
				Port:      &pm.cfg.ServicePort,
			},
		},
		Rules:                   admissionRegistrationRules,
		FailurePolicy:           &admissionRegistrationFailurePolicy,
		SideEffects:             &admissionRegistrationSideEffects,
		AdmissionReviewVersions: admissionRegistrationVersions,
		ObjectSelector:          httpHandler.mutator.LabelSelector(),
	}
}

// Verify that there aren't any duplicate webhook names or URI paths.
func verifyHTTPHandlers(ctx context.Context, httpHandlers []httpHandler) error {
	webhookNameOccurrences := make(map[string]bool)
	pathOccurrences := make(map[string]bool)

	duplicateWebhookNames := make([]string, 0)
	duplicatePaths := make([]string, 0)

	for _, handler := range httpHandlers {
		if webhookNameOccurrences[handler.mutatingWebhookName] {
			duplicateWebhookNames = append(duplicateWebhookNames, handler.mutatingWebhookName)
		}
		webhookNameOccurrences[handler.mutatingWebhookName] = true
		if pathOccurrences[handler.path] {
			duplicatePaths = append(duplicatePaths, handler.path)
		}
		pathOccurrences[handler.path] = true
	}

	e := ""
	if len(duplicateWebhookNames) > 0 {
		e += fmt.Sprintf("Duplicate webhook names found: [%v]. ", strings.Join(duplicateWebhookNames, ","))
	}
	if len(duplicatePaths) > 0 {
		e += fmt.Sprintf("Duplicate paths found: [%v]", strings.Join(duplicatePaths, ","))
	}
	if len(e) > 0 {
		logger.Errorf(ctx, "Invalid webhook configuration: %v", e)
		return fmt.Errorf("Invalid webhook configuration: %v", e)
	}
	return nil
}
