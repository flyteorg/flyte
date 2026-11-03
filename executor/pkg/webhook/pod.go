// Package webhook is the executor's mutating admission webhook. It always
// serves the secrets mutator, which injects secret references into pods that
// carry the inject-flyte-secrets label; deployments can add their own pod and
// node mutators (WithPodMutators, WithNodeMutators), each served as its own
// MutatingWebhook with its own path and selector.
package webhook

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret"
	webhookConfig "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
)

// secretsWebhookName is the MutatingWebhook name of the secrets mutator.
const secretsWebhookName = "flyte-pod-webhook.flyte.org"

// Webhook is the set of handlers the webhook server serves.
type Webhook struct {
	cfg            *webhookConfig.Config
	handlers       []ResourceHandler
	secretsMutator *secret.SecretsPodMutator
}

// Option adds to or customizes the webhook built by Setup / NewWebhook.
type Option func(*options)

type options struct {
	podMutators    []PodMutator
	nodeMutators   []NodeMutator
	handlers       []ResourceHandler
	limitNamespace string
}

// WithPodMutators serves additional pod mutators, after the secrets mutator.
func WithPodMutators(mutators ...PodMutator) Option {
	return func(o *options) { o.podMutators = append(o.podMutators, mutators...) }
}

// WithNodeMutators serves node mutators.
func WithNodeMutators(mutators ...NodeMutator) Option {
	return func(o *options) { o.nodeMutators = append(o.nodeMutators, mutators...) }
}

// WithHandlers serves fully custom handlers (e.g. ones built with
// NewPodHandler and a custom name or path).
func WithHandlers(handlers ...ResourceHandler) Option {
	return func(o *options) { o.handlers = append(o.handlers, handlers...) }
}

// WithLimitNamespace scopes the secrets mutator's Secret informer to one
// namespace, for deployments that run every task pod there.
func WithLimitNamespace(namespace string) Option {
	return func(o *options) { o.limitNamespace = namespace }
}

// NewWebhook builds the secrets handler plus any extra handlers.
func NewWebhook(ctx context.Context, cfg *webhookConfig.Config, podNamespace string, scheme *runtime.Scheme,
	scope promutils.Scope, opts ...Option) (*Webhook, error) {

	o := options{}
	for _, opt := range opts {
		opt(&o)
	}

	secretsMutator, err := newSecretsMutator(ctx, cfg, podNamespace, o.limitNamespace, scope.NewSubScope("secrets"))
	if err != nil {
		return nil, fmt.Errorf("failed to create secrets mutator: %w", err)
	}

	decoder := admission.NewDecoder(scheme)
	handlers := []ResourceHandler{
		NewPodHandler(decoder, secretsMutator,
			WithWebhookName(secretsWebhookName),
			// /mutate--v1-pod is where this webhook has always served secrets;
			// /mutate--v1-pod/secrets is the per-mutator path, which charts
			// that render the MutatingWebhookConfiguration themselves use.
			WithPaths(generateMutatePath(podGVK)),
			WithExtraPaths(mutatePath(podGVK, secretsMutator.ID()))),
	}
	for _, m := range o.podMutators {
		handlers = append(handlers, NewPodHandler(decoder, m))
	}
	for _, m := range o.nodeMutators {
		handlers = append(handlers, NewNodeHandler(decoder, m))
	}
	handlers = append(handlers, o.handlers...)

	if err := verifyHandlers(handlers); err != nil {
		return nil, err
	}
	return &Webhook{cfg: cfg, handlers: handlers, secretsMutator: secretsMutator}, nil
}

// verifyHandlers rejects two handlers claiming the same path or name.
func verifyHandlers(handlers []ResourceHandler) error {
	paths := map[string]bool{}
	for _, h := range handlers {
		for _, p := range h.Paths() {
			if paths[p] {
				return fmt.Errorf("invalid webhook configuration: duplicate path %q", p)
			}
			paths[p] = true
		}
	}
	return nil
}

// SecretsMutator returns the mutator that owns the secret caches, so the
// cache invalidation server can clear them.
func (w *Webhook) SecretsMutator() *secret.SecretsPodMutator {
	return w.secretsMutator
}

// Handlers returns every handler the webhook serves.
func (w *Webhook) Handlers() []ResourceHandler {
	return w.handlers
}

// Register serves every handler on the manager's webhook server.
func (w *Webhook) Register(ctx context.Context, mgr manager.Manager) error {
	for _, h := range w.handlers {
		for _, path := range h.Paths() {
			logger.Infof(ctx, "Registering webhook path [%v]", path)
			mgr.GetWebhookServer().Register(path, &admission.Webhook{Handler: h.AdmissionHandler()})
		}
	}
	return nil
}

// CreateMutationWebhookConfiguration builds the configuration that points the
// API server at every handler.
func (w *Webhook) CreateMutationWebhookConfiguration(
	namespace string,
) (*admissionregistrationv1.MutatingWebhookConfiguration, error) {
	caBytes, err := os.ReadFile(filepath.Join(w.cfg.ExpandCertDir(), "ca.crt"))
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		caBytes = []byte{}
	}

	webhooks := make([]admissionregistrationv1.MutatingWebhook, 0, len(w.handlers))
	for _, h := range w.handlers {
		webhooks = append(webhooks, h.MutatingWebhook(namespace, caBytes, w.cfg))
	}
	return &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      w.cfg.ServiceName,
			Namespace: namespace,
		},
		Webhooks: webhooks,
	}, nil
}
