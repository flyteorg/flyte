package common

import (
	"strings"
)

const projectTemplate = "{{ project }}"
const domainTemplate = "{{ domain }}"

const replaceAllInstancesOfString = -1

// GetNamespaceName returns kubernetes namespace name according to user defined template from config
func GetNamespaceName(template string, project, domain string) string {
	var namespace = template
	namespace = strings.Replace(namespace, projectTemplate, project, replaceAllInstancesOfString)
	namespace = strings.Replace(namespace, domainTemplate, domain, replaceAllInstancesOfString)

	return namespace
}
