package webhook

import (
	"context"
	"fmt"
	"strings"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/flyteorg/flyte/flytepropeller/pkg/secret"
	"github.com/flyteorg/flyte/flytepropeller/pkg/secret/config"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

// WebhookConfig is a unified configuration manager for all webhook mutators.
// It handles the registration and configuration of pod, node, and potentially other
// Kubernetes resource mutators in a single place.
type WebhookConfig struct {
	cfg            *config.Config
	handlers       []ResourceHandler
	caBytes        []byte
	podNamespace   string
	secretsMutator *secret.SecretsPodMutator
}

// ResourceHandler represents a generic handler for any Kubernetes resource type
type ResourceHandler interface {
	// GetMutatingWebhook returns the Kubernetes MutatingWebhook configuration
	GetMutatingWebhook(namespace string, caBytes []byte, cfg *config.Config) admissionregistrationv1.MutatingWebhook
	// GetPath returns the webhook path for registration
	GetPath() string
	// GetAdmissionHandler returns the controller-runtime webhook handler
	GetAdmissionHandler() admission.Handler
}

// verifyHandlers ensures there are no duplicate webhook names or paths
func verifyHandlers(ctx context.Context, handlers []ResourceHandler) error {
	pathOccurrences := make(map[string]bool)
	duplicatePaths := make([]string, 0)

	for _, handler := range handlers {
		path := handler.GetPath()

		if pathOccurrences[path] {
			duplicatePaths = append(duplicatePaths, path)
		}
		pathOccurrences[path] = true
	}

	if len(duplicatePaths) > 0 {
		e := fmt.Sprintf("Duplicate paths found: [%v]", strings.Join(duplicatePaths, ","))
		logger.Errorf(ctx, "Invalid webhook configuration: %v", e)
		return fmt.Errorf("Invalid webhook configuration: %v", e)
	}
	return nil
}

// NewWebhookConfig creates a new unified webhook configuration that manages
// all types of mutators (pod, node, etc.)
func NewWebhookConfig(ctx context.Context, cfg *config.Config, scheme *runtime.Scheme, podNamespace string, scope promutils.Scope, caBytes []byte) (*WebhookConfig, error) {
	var handlers []ResourceHandler

	// Initialize pod handlers
	podHandlers, secretsMutator, err := initializePodHandlers(ctx, cfg, scheme, podNamespace, scope)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize pod handlers: %w", err)
	}
	handlers = append(handlers, podHandlers...)

	// Initialize node handlers
	nodeHandlers, err := initializeNodeHandlers(ctx, cfg, scheme, scope)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize node handlers: %w", err)
	}
	handlers = append(handlers, nodeHandlers...)

	// Verify no conflicts
	if err := verifyHandlers(ctx, handlers); err != nil {
		return nil, err
	}

	return &WebhookConfig{
		cfg:            cfg,
		handlers:       handlers,
		caBytes:        caBytes,
		podNamespace:   podNamespace,
		secretsMutator: secretsMutator,
	}, nil
}

// Register registers all webhook handlers with the provided registerer
func (wc *WebhookConfig) Register(ctx context.Context, registerer HTTPHookRegistererIface) error {
	for _, handler := range wc.handlers {
		path := handler.GetPath()
		admissionHandler := handler.GetAdmissionHandler()
		wh := &admission.Webhook{Handler: admissionHandler}

		logger.Infof(ctx, "Registering webhook path [%v]", path)
		registerer.Register(path, wh)
	}
	return nil
}

// CreateMutationWebhookConfiguration creates a Kubernetes MutatingWebhookConfiguration
// that includes all registered handlers
func (wc *WebhookConfig) CreateMutationWebhookConfiguration(namespace string) (*admissionregistrationv1.MutatingWebhookConfiguration, error) {
	webhooks := make([]admissionregistrationv1.MutatingWebhook, 0, len(wc.handlers))

	for _, handler := range wc.handlers {
		webhook := handler.GetMutatingWebhook(namespace, wc.caBytes, wc.cfg)
		webhooks = append(webhooks, webhook)
	}

	return &admissionregistrationv1.MutatingWebhookConfiguration{
		TypeMeta: metav1.TypeMeta{
			APIVersion: admissionregistrationv1.SchemeGroupVersion.String(),
			Kind:       "MutatingWebhookConfiguration",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      wc.cfg.ServiceName,
			Namespace: namespace,
		},
		Webhooks: webhooks,
	}, nil
}

// GetHandlers returns all registered handlers for testing purposes
func (wc *WebhookConfig) GetHandlers() []ResourceHandler {
	return wc.handlers
}

// GetSecretsMutator returns the secrets pod mutator for cache invalidation
func (wc *WebhookConfig) GetSecretsMutator() *secret.SecretsPodMutator {
	return wc.secretsMutator
}
