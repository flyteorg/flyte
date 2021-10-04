package utils

import (
	"regexp"
	"strings"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/encoding"
	"k8s.io/apimachinery/pkg/util/validation"
)

var dns1123InvalidRegex = regexp.MustCompile("[^-.a-z0-9]")
var camelCaseRegex = regexp.MustCompile("([a-z0-9])([A-Z])")

const maxUniqueIDLength = 20

// ConvertToDNS1123SubdomainCompatibleString converts a string that doesn't conform to the definition of a subdomain in DNS (RFC 1123) to a string that conforms. It doesn't do well on labels (separated by dots) starting or ending with hyphens.
func ConvertToDNS1123SubdomainCompatibleString(name string) string {
	if errs := validation.IsDNS1123Subdomain(name); len(errs) == 0 {
		return name
	}
	name = ConvertCamelCaseToKebabCase(name) // best effort to preserve readability for Java class name
	name = strings.ToLower(name)
	name = dns1123InvalidRegex.ReplaceAllString(name, "")
	name = strings.Trim(name, ".-")
	if len(name) > validation.DNS1123SubdomainMaxLength {
		fixedLengthID, err := encoding.FixedLengthUniqueID(name, maxUniqueIDLength)
		if err == nil {
			name = name[:validation.DNS1123SubdomainMaxLength-maxUniqueIDLength-1] + "-" + fixedLengthID
		} else {
			name = name[:validation.DNS1123SubdomainMaxLength]
		}
	}
	return name
}

// ConvertCamelCaseToKebabCase rewrites a string written in camel case (e.g. PenPineappleApplePen) in kebab case (pen-pineapple-apple-pen)
func ConvertCamelCaseToKebabCase(name string) string {
	return strings.ToLower(camelCaseRegex.ReplaceAllString(name, "${1}-${2}"))
}
