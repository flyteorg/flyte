package namespaceutils

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation"
)

const orgTemplate = "{{ org }}"
const projectTemplate = "{{ project }}"
const domainTemplate = "{{ domain }}"

// GetNamespaceName returns kubernetes namespace name according to user defined template from config
func GetNamespaceName(template string, org, project, domain string) string {
	return strings.NewReplacer(
		orgTemplate, org,
		projectTemplate, project,
		domainTemplate, domain,
	).Replace(template)
}

// MappedNamespace returns the namespace a project-domain runs in: template rendered for org,
// project and domain, or defaultNamespace when template is empty. A rendered name that is
// not a valid namespace name (a DNS-1123 label) is an error.
func MappedNamespace(template, defaultNamespace, org, project, domain string) (string, error) {
	if template == "" {
		return defaultNamespace, nil
	}
	namespace := GetNamespaceName(template, org, project, domain)
	if errs := validation.IsDNS1123Label(namespace); len(errs) > 0 {
		return "", fmt.Errorf("namespace %q for project %q and domain %q is not a valid Kubernetes namespace name: %s",
			namespace, project, domain, strings.Join(errs, "; "))
	}
	return namespace, nil
}

// GetNameWithNamespacedPrefix returns a name with the given prefix prepended
func GetNameWithNamespacedPrefix(prefix, name string) string {
	return prefix + name
}
