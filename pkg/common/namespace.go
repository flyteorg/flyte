package common

import "fmt"

type NamespaceMapping int

const namespaceFormat = "%s-%s"

const (
	NamespaceMappingProjectDomain NamespaceMapping = iota
	NamespaceMappingDomain        NamespaceMapping = iota
	NamespaceMappingProject       NamespaceMapping = iota
)

// GetNamespaceName returns kubernetes namespace name
func GetNamespaceName(mapping NamespaceMapping, project, domain string) string {
	switch mapping {
	case NamespaceMappingDomain:
		return domain
	case NamespaceMappingProject:
		return project
	case NamespaceMappingProjectDomain:
		fallthrough
	default:
		return fmt.Sprintf(namespaceFormat, project, domain)
	}
}
