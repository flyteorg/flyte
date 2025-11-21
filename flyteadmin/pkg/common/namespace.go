package common

import (
	"github.com/flyteorg/flyte/flytestdlib/namespaceutils"
)

// GetNamespaceName returns kubernetes namespace name according to user defined template from config
// Deprecated: Use namespaceutils.GetNamespaceName instead
func GetNamespaceName(template string, org, project, domain string) string {
	return namespaceutils.GetNamespaceName(template, org, project, domain)
}
