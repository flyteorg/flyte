package webhook

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	webhookConfig "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

// PodMutator is one kind of Pod mutation served by the webhook. Each mutator
// gets its own MutatingWebhook entry, path and object selector, so a mutator
// only sees the pods it selects.
type PodMutator interface {
	// ID names the mutator; it is part of the default webhook name and path.
	ID() string
	// Mutate returns the mutated pod and whether it changed, or an admission
	// response that rejects the request.
	Mutate(ctx context.Context, p *corev1.Pod) (newP *corev1.Pod, changed bool, err *admission.Response)
	// LabelSelector selects the pods this mutator applies to.
	LabelSelector() *metav1.LabelSelector
}

// NodeMutator is one kind of Node mutation served by the webhook.
type NodeMutator interface {
	ID() string
	Mutate(ctx context.Context, n *corev1.Node) (newN *corev1.Node, changed bool, err *admission.Response)
	LabelSelector() *metav1.LabelSelector
}

// ResourceHandler serves one MutatingWebhook: its admission handler, the
// paths it is served at and its entry in the MutatingWebhookConfiguration.
type ResourceHandler interface {
	// Paths are the URL paths the handler is served at. The first is the one
	// written into a generated MutatingWebhookConfiguration; the others keep
	// externally managed configurations that use another path working.
	Paths() []string
	// MutatingWebhook is the handler's entry in the MutatingWebhookConfiguration.
	MutatingWebhook(namespace string, caBytes []byte, cfg *webhookConfig.Config) admissionregistrationv1.MutatingWebhook
	// AdmissionHandler handles the admission request.
	AdmissionHandler() admission.Handler
}

// HandlerOption customizes a handler built by NewPodHandler or NewNodeHandler.
type HandlerOption func(*handlerOptions)

type handlerOptions struct {
	name       string
	paths      []string
	extraPaths []string
}

// WithWebhookName overrides the MutatingWebhook name (default <id>.flyte.org).
func WithWebhookName(name string) HandlerOption {
	return func(o *handlerOptions) { o.name = name }
}

// WithPaths overrides the paths the handler is served at; the first one goes
// into the generated MutatingWebhookConfiguration.
func WithPaths(paths ...string) HandlerOption {
	return func(o *handlerOptions) { o.paths = paths }
}

// WithExtraPaths serves the handler at additional paths, after the default.
func WithExtraPaths(paths ...string) HandlerOption {
	return func(o *handlerOptions) { o.extraPaths = append(o.extraPaths, paths...) }
}

type mutationHandler[T client.Object] struct {
	decoder  admission.Decoder
	id       string
	name     string
	paths    []string
	rules    []admissionregistrationv1.RuleWithOperations
	selector *metav1.LabelSelector
	// namespaced handlers honour cfg.NamespaceSelector; cluster-scoped
	// objects (nodes) have no namespace to select on.
	namespaced bool
	newObject  func() T
	mutate     func(ctx context.Context, obj T) (T, bool, *admission.Response)
}

// NewPodHandler serves a PodMutator at /mutate--v1-pod/<id>.
func NewPodHandler(decoder admission.Decoder, mutator PodMutator, opts ...HandlerOption) ResourceHandler {
	return newMutationHandler(decoder, mutator.ID(), mutator.LabelSelector(), true,
		mutatePath(podGVK, mutator.ID()), admissionRules("*", "pods"),
		func() *corev1.Pod { return &corev1.Pod{} }, mutator.Mutate, opts)
}

// NewNodeHandler serves a NodeMutator at /mutate--v1-node/<id>.
func NewNodeHandler(decoder admission.Decoder, mutator NodeMutator, opts ...HandlerOption) ResourceHandler {
	return newMutationHandler(decoder, mutator.ID(), mutator.LabelSelector(), false,
		mutatePath(nodeGVK, mutator.ID()), admissionRules("", "nodes"),
		func() *corev1.Node { return &corev1.Node{} }, mutator.Mutate, opts)
}

func newMutationHandler[T client.Object](
	decoder admission.Decoder, id string, selector *metav1.LabelSelector, namespaced bool,
	defaultPath string, rules []admissionregistrationv1.RuleWithOperations, newObject func() T,
	mutate func(context.Context, T) (T, bool, *admission.Response), opts []HandlerOption,
) *mutationHandler[T] {

	o := handlerOptions{name: id + ".flyte.org", paths: []string{defaultPath}}
	for _, opt := range opts {
		opt(&o)
	}
	return &mutationHandler[T]{
		decoder:    decoder,
		id:         id,
		name:       o.name,
		paths:      append(append([]string{}, o.paths...), o.extraPaths...),
		rules:      rules,
		selector:   selector,
		namespaced: namespaced,
		newObject:  newObject,
		mutate:     mutate,
	}
}

func (h *mutationHandler[T]) Paths() []string { return h.paths }

func (h *mutationHandler[T]) AdmissionHandler() admission.Handler { return h }

func (h *mutationHandler[T]) MutatingWebhook(
	namespace string, caBytes []byte, cfg *webhookConfig.Config,
) admissionregistrationv1.MutatingWebhook {
	path := h.paths[0]
	fail := admissionregistrationv1.Fail
	sideEffects := admissionregistrationv1.SideEffectClassNoneOnDryRun
	wh := admissionregistrationv1.MutatingWebhook{
		Name: h.name,
		ClientConfig: admissionregistrationv1.WebhookClientConfig{
			CABundle: caBytes,
			Service: &admissionregistrationv1.ServiceReference{
				Name:      cfg.ServiceName,
				Namespace: namespace,
				Path:      &path,
				Port:      &cfg.ServicePort,
			},
		},
		Rules:                   h.rules,
		FailurePolicy:           &fail,
		SideEffects:             &sideEffects,
		AdmissionReviewVersions: []string{"v1", "v1beta1"},
		ObjectSelector:          h.selector,
	}
	if cfg.WebhookTimeout > 0 {
		timeout := cfg.WebhookTimeout
		wh.TimeoutSeconds = &timeout
	}
	if h.namespaced {
		wh.NamespaceSelector = cfg.NamespaceSelector
	}
	return wh
}

func (h *mutationHandler[T]) Handle(ctx context.Context, request admission.Request) admission.Response {
	ctx = contextutils.WithRequestID(ctx, rand.String(10))
	obj := h.newObject()
	if err := h.decoder.Decode(request, obj); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	newObj, changed, admissionErr := h.mutate(ctx, obj)
	if admissionErr != nil {
		return *admissionErr
	}
	if !changed {
		return admission.Allowed("No changes")
	}

	logger.Infof(ctx, "[%s] mutated [%v/%v]", h.id, obj.GetNamespace(), obj.GetName())
	marshalled, err := json.Marshal(newObj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(request.Object.Raw, marshalled)
}

var (
	podGVK  = schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"}
	nodeGVK = schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Node"}
)

func admissionRules(group, resource string) []admissionregistrationv1.RuleWithOperations {
	return []admissionregistrationv1.RuleWithOperations{{
		Operations: []admissionregistrationv1.OperationType{admissionregistrationv1.Create},
		Rule: admissionregistrationv1.Rule{
			APIGroups:   []string{group},
			APIVersions: []string{"v1"},
			Resources:   []string{resource},
		},
	}}
}

// generateMutatePath is /mutate-<group with dashes>-<version>-<kind>, e.g.
// /mutate--v1-pod for core pods.
func generateMutatePath(gvk schema.GroupVersionKind) string {
	return "/mutate-" + strings.ReplaceAll(gvk.Group, ".", "-") + "-" +
		gvk.Version + "-" + strings.ToLower(gvk.Kind)
}

func mutatePath(gvk schema.GroupVersionKind, id string) string {
	return generateMutatePath(gvk) + "/" + id
}
