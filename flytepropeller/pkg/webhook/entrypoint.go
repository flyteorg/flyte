package webhook

import (
	"context"
	"encoding/base64"
	"encoding/json"
	errors2 "errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ghodss/yaml"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/config"
	secretConfig "github.com/flyteorg/flyte/flytepropeller/pkg/secret/config"
	"github.com/flyteorg/flyte/flytepropeller/pkg/utils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

const (
	PodNameEnvVar      = "POD_NAME"
	PodNamespaceEnvVar = "POD_NAMESPACE"
)

func Run(ctx context.Context, propellerCfg *config.Config, cfg *secretConfig.Config,
	defaultNamespace string, scope *promutils.Scope, mgr manager.Manager) error {
	err := RunWebhook(ctx, propellerCfg, cfg,
		defaultNamespace, scope, mgr)
	if err != nil {
		return err
	}

	logger.Infof(ctx, "Started propeller webhook")
	<-ctx.Done()

	return nil
}

func RunWebhook(ctx context.Context, propellerCfg *config.Config, cfg *secretConfig.Config,
	defaultNamespace string, scope *promutils.Scope, mgr manager.Manager) error {
	raw, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	fmt.Println(string(raw))

	kubeClient, _, err := utils.GetKubeConfig(ctx, propellerCfg)
	if err != nil {
		return err
	}

	webhookScope := (*scope).NewSubScope("webhook")

	caBytes, err := os.ReadFile(filepath.Join(cfg.ExpandCertDir(), "ca.crt"))
	if err != nil {
		// ca.crt is optional. If not provided, API Server will assume the webhook is serving SSL using a certificate
		// issued by a known Cert Authority.
		if os.IsNotExist(err) {
			caBytes = make([]byte, 0)
		} else {
			return err
		}
	}

	webhookConfig, err := NewWebhookConfig(ctx, cfg, mgr.GetScheme(), defaultNamespace, webhookScope, caBytes)
	if err != nil {
		return err
	}

	mutateConfigCR, err := webhookConfig.CreateMutationWebhookConfiguration(defaultNamespace)
	if err != nil {
		return err
	}

	// It's the responsibility of the platform engineer to create the MutatingWebhookConfiguration otherwise.
	if !cfg.DisableCreateMutatingWebhookConfig {
		// Creates a MutationConfig to instruct ApiServer to call this service whenever a Pod is being created.
		err = createMutationConfig(ctx, kubeClient, mutateConfigCR, defaultNamespace)
		if err != nil {
			return err
		}
	} else {
		logger.Info(ctx, "Skipping create mutation webhook configuration. Enable debug logs to output the raw"+
			" yaml configuration that must exist to use this webhook if you wish to create it manually. Or disable"+
			" disableCreateMutatingWebhookConfig setting.")
		if logger.IsLoggable(ctx, logger.DebugLevel) {
			raw, err := yaml.Marshal(mutateConfigCR)
			if err != nil {
				logger.Warnf(ctx, "Failed to marshal mutateConfigCR to yaml format. Error: %v", err)
			} else {
				logger.Debugf(ctx, "Since disableCreateMutatingWebhookConfig is set to True, I'll not create"+
					" the object. Ensure that a MutatingWebhookConfiguration with the following base64 encoded spec is created:"+
					" [%s]", base64.StdEncoding.EncodeToString(raw))
			}
		}
	}

	err = webhookConfig.Register(ctx, K8sRuntimeHTTPHookRegisterer{mgr: mgr})
	if err != nil {
		logger.Fatalf(ctx, "Failed to register webhook with manager. Error: %v", err)
	}

	// Start cache invalidation server if a secrets mutator is available
	if secretsMutator := webhookConfig.GetSecretsMutator(); secretsMutator != nil && cfg.CacheInvalidationPort > 0 {
		go func() {
			if err := StartCacheInvalidationServer(ctx, cfg.CacheInvalidationPort, secretsMutator); err != nil {
				logger.Errorf(ctx, "Cache invalidation server failed: %v", err)
			}
		}()
	}

	logger.Infof(ctx, "Started propeller webhook")
	<-ctx.Done()

	return nil
}

func createMutationConfig(ctx context.Context, kubeClient *kubernetes.Clientset,
	mutateConfig *admissionregistrationv1.MutatingWebhookConfiguration, podNamespace string) error {

	shouldAddOwnerRef := true
	podName, found := os.LookupEnv(PodNameEnvVar)
	if !found {
		shouldAddOwnerRef = false
	}

	if shouldAddOwnerRef {
		// Lookup the pod to retrieve its UID
		p, err := kubeClient.CoreV1().Pods(podNamespace).Get(ctx, podName, v1.GetOptions{})
		if err != nil {
			logger.Infof(ctx, "Failed to get Pod [%v/%v]. Error: %v", podNamespace, podName, err)
			return fmt.Errorf("failed to get pod. Error: %w", err)
		}

		mutateConfig.OwnerReferences = p.OwnerReferences
	}

	logger.Infof(ctx, "Creating MutatingWebhookConfiguration [%v/%v]", mutateConfig.GetNamespace(), mutateConfig.GetName())

	_, err := kubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Create(ctx, mutateConfig, v1.CreateOptions{})
	var statusErr *errors.StatusError
	if err != nil && errors2.As(err, &statusErr) && statusErr.Status().Reason == v1.StatusReasonAlreadyExists {
		logger.Infof(ctx, "Failed to create MutatingWebhookConfiguration. Will attempt to update. Error: %v", err)
		obj, getErr := kubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(ctx, mutateConfig.Name, v1.GetOptions{})
		if getErr != nil {
			logger.Infof(ctx, "Failed to get MutatingWebhookConfiguration. Error: %v", getErr)
			return err
		}

		obj.Webhooks = mutateConfig.Webhooks
		_, err = kubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Update(ctx, obj, v1.UpdateOptions{})
		if err == nil {
			logger.Infof(ctx, "Successfully updated existing mutating webhook config.")
		}

		return err
	} else if err != nil {
		logger.Infof(ctx, "Failed to create MutatingWebhookConfiguration [%v/%v]. Error: %v", mutateConfig.GetNamespace(), mutateConfig.GetName(), err)
		return fmt.Errorf("failed to create mutatingwebhookconfiguration. Error: %w", err)
	}

	return nil
}
