package namespaceutils

import (
	"strings"
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

// GetNameWithNamespacedPrefix returns a name with the given prefix prepended
func GetNameWithNamespacedPrefix(prefix, name string) string {
	return prefix + name
}
