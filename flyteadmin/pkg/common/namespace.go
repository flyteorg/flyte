package common

import "fmt"

type NamespaceMapping int

const namespaceFormat = "%s-%s"

const (
	ProjectDomain NamespaceMapping = iota
	Domain        NamespaceMapping = iota
)

// GetNamespaceName returns kubernetes namespace name
func GetNamespaceName(mapping NamespaceMapping, project, domain string) string {
	switch mapping {
	case Domain:
		return domain
	case ProjectDomain:
		fallthrough
	default:
		return fmt.Sprintf(namespaceFormat, project, domain)
	}
}
