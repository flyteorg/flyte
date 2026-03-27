package webhook

import (
	"context"
	errors2 "errors"
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	webhookConfig "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
)

const (
	PodNameEnvVar      = "POD_NAME"
	PodNamespaceEnvVar = "POD_NAMESPACE"
)

// Setup initializes the webhook: generates certs, registers MutatingWebhookConfiguration, and registers the HTTP handler.
// It is called before mgr.Start() so that the webhook server is ready to receive requests.
func Setup(ctx context.Context, kubeClient kubernetes.Interface, cfg *webhookConfig.Config,
	defaultNamespace string, scope promutils.Scope, mgr manager.Manager) error {

	if err := InitCerts(ctx, kubeClient, cfg); err != nil {
		return fmt.Errorf("webhook: failed to initialize certs: %w", err)
	}

	podMutator, err := NewPodMutator(ctx, cfg, defaultNamespace, mgr.GetScheme(), scope)
	if err != nil {
		return fmt.Errorf("webhook: failed to create pod mutator: %w", err)
	}

	if err := createMutationConfig(ctx, kubeClient, podMutator, defaultNamespace); err != nil {
		return fmt.Errorf("webhook: failed to create MutatingWebhookConfiguration: %w", err)
	}

	if err := podMutator.Register(ctx, mgr); err != nil {
		return fmt.Errorf("webhook: failed to register handler: %w", err)
	}

	logger.Infof(ctx, "Webhook setup complete")
	return nil
}

func createMutationConfig(ctx context.Context, kubeClient kubernetes.Interface, webhookObj *PodMutator, defaultNamespace string) error {
	shouldAddOwnerRef := true
	podName, found := os.LookupEnv(PodNameEnvVar)
	if !found {
		shouldAddOwnerRef = false
	}

	podNamespace, found := os.LookupEnv(PodNamespaceEnvVar)
	if !found {
		shouldAddOwnerRef = false
		podNamespace = defaultNamespace
	}

	mutateConfig, err := webhookObj.CreateMutationWebhookConfiguration(podNamespace)
	if err != nil {
		return err
	}

	if shouldAddOwnerRef {
		p, err := kubeClient.CoreV1().Pods(podNamespace).Get(ctx, podName, v1.GetOptions{})
		if err != nil {
			logger.Infof(ctx, "Failed to get Pod [%v/%v]. Error: %v", podNamespace, podName, err)
			return fmt.Errorf("failed to get pod: %w", err)
		}
		mutateConfig.OwnerReferences = p.OwnerReferences
	}

	logger.Infof(ctx, "Creating MutatingWebhookConfiguration [%v/%v]", mutateConfig.GetNamespace(), mutateConfig.GetName())

	_, err = kubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Create(ctx, mutateConfig, v1.CreateOptions{})
	var statusErr *errors.StatusError
	if err != nil && errors2.As(err, &statusErr) && statusErr.Status().Reason == v1.StatusReasonAlreadyExists {
		logger.Infof(ctx, "MutatingWebhookConfiguration already exists. Attempting update.")
		obj, getErr := kubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(ctx, mutateConfig.Name, v1.GetOptions{})
		if getErr != nil {
			return err
		}
		obj.Webhooks = mutateConfig.Webhooks
		_, err = kubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Update(ctx, obj, v1.UpdateOptions{})
		if err == nil {
			logger.Infof(ctx, "Successfully updated MutatingWebhookConfiguration.")
		}
		return err
	} else if err != nil {
		return fmt.Errorf("failed to create MutatingWebhookConfiguration: %w", err)
	}

	return nil
}
