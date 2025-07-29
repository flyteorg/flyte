package webhook

import (
	"fmt"
	"strings"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	unionWebhookNameDomain = "union.ai"
)

var (
	admissionRegistrationVersions = []string{
		"v1",
		"v1beta1",
	}
	admissionRegistrationFailurePolicy = admissionregistrationv1.Fail
	admissionRegistrationSideEffects   = admissionregistrationv1.SideEffectClassNoneOnDryRun
)

// generateMutatePath creates a webhook path for the given GVK and subpath
func generateMutatePath(gvk schema.GroupVersionKind, subpath string) string {
	return "/mutate-" + strings.Replace(gvk.Group, ".", "-", -1) + "-" +
		gvk.Version + "-" + strings.ToLower(gvk.Kind) + "/" + subpath
}

// getMutatingWebhookName generates a webhook name for the given ID
func getMutatingWebhookName(id string) string {
	return fmt.Sprintf("%s-webhook.%s", id, unionWebhookNameDomain)
}
