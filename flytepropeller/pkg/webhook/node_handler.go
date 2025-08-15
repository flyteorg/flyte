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

	"github.com/flyteorg/flyte/flytepropeller/pkg/secret/config"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

//go:generate mockery --output=./mocks --case=underscore --name=NodeMutator

// NodeMutator contains the business logic for a unique type of node mutation or validation.
type NodeMutator interface {
	ID() string
	// Conducts the act of mutating the node.
	Mutate(ctx context.Context, n *corev1.Node) (newN *corev1.Node, changed bool, err *admission.Response)
	// Defines how to select which Nodes to apply the webhook to.
	LabelSelector() *metav1.LabelSelector
}

// NodeHandler handles webhook mutations for Node resources
type NodeHandler struct {
	decoder             *admission.Decoder
	mutator             NodeMutator
	mutatingWebhookName string
	path                string
}

// NewNodeHandler creates a new node handler for the given mutator
func NewNodeHandler(decoder *admission.Decoder, mutator NodeMutator) *NodeHandler {
	return &NodeHandler{
		decoder:             decoder,
		mutator:             mutator,
		mutatingWebhookName: getMutatingWebhookName(mutator.ID()),
		path:                getNodeMutatePath(mutator.ID()),
	}
}

// GetMutatingWebhook returns the Kubernetes MutatingWebhook configuration for nodes
func (h *NodeHandler) GetMutatingWebhook(namespace string, caBytes []byte, cfg *config.Config) admissionregistrationv1.MutatingWebhook {
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
		Rules:                   getNodeAdmissionRules(),
		TimeoutSeconds:          &cfg.WebhookTimeout,
		FailurePolicy:           &admissionRegistrationFailurePolicy,
		SideEffects:             &admissionRegistrationSideEffects,
		AdmissionReviewVersions: admissionRegistrationVersions,
		ObjectSelector:          h.mutator.LabelSelector(),
	}
}

// GetPath returns the webhook path
func (h *NodeHandler) GetPath() string {
	return h.path
}

// GetAdmissionHandler returns the controller-runtime webhook handler
func (h *NodeHandler) GetAdmissionHandler() admission.Handler {
	return h
}

// Handle processes the admission request for node mutations
func (h *NodeHandler) Handle(ctx context.Context, request admission.Request) admission.Response {
	ctx = contextutils.WithRequestID(ctx, rand.String(10))

	// Get the object in the request
	obj := &corev1.Node{}
	err := h.decoder.Decode(request, obj)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	newObj, changed, admissionError := h.mutator.Mutate(ctx, obj)
	if admissionError != nil {
		return *admissionError
	}

	if changed {
		logger.Infof(ctx, "Mutated Node [%v]", obj.Name)
		marshalled, err := json.Marshal(newObj)
		if err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}

		// Create the patch
		return admission.PatchResponseFromRaw(request.Object.Raw, marshalled)
	}

	return admission.Allowed("No changes")
}

// initializeNodeHandlers creates all node-specific webhook handlers
func initializeNodeHandlers(ctx context.Context, cfg *config.Config, scheme *runtime.Scheme, scope promutils.Scope) ([]ResourceHandler, error) {
	var handlers []ResourceHandler

	// cfg and scope parameters are currently unused but kept for consistency with pod handlers
	// and future extensibility when additional node mutators may need configuration
	_ = cfg
	_ = scope

	decoder := admission.NewDecoder(scheme)

	// Node startup taints mutator (conditional)
	nodeStartupTaintsCfg := GetNodeStartupTaintsConfig()
	if nodeStartupTaintsCfg.Enabled {
		logger.Infof(ctx, "Enabling Node Startup Taint Mutator")
		mutator, err := NewNodeStartupTaintsMutator(nodeStartupTaintsCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create node startup taint mutator: %w", err)
		}
		handlers = append(handlers, NewNodeHandler(decoder, mutator))
	} else {
		logger.Infof(ctx, "Node Startup Taint Mutator is disabled")
	}

	return handlers, nil
}

// getNodeAdmissionRules returns the admission rules for node webhooks
func getNodeAdmissionRules() []admissionregistrationv1.RuleWithOperations {
	return []admissionregistrationv1.RuleWithOperations{
		{
			Operations: []admissionregistrationv1.OperationType{
				admissionregistrationv1.Create,
			},
			Rule: admissionregistrationv1.Rule{
				APIGroups:   []string{""},
				APIVersions: []string{"v1"},
				Resources:   []string{"nodes"},
			},
		},
	}
}

func getNodeMutatePath(subpath string) string {
	node := &corev1.Node{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: corev1.SchemeGroupVersion.String(),
		},
	}
	return generateMutatePath(node.GroupVersionKind(), subpath)
}
