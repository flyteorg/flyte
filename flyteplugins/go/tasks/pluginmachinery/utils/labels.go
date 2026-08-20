package utils

import (
	"regexp"

	"k8s.io/apimachinery/pkg/util/validation"
)

var labelInvalidCharsRegex = regexp.MustCompile(`[^a-zA-Z0-9._-]`)
var labelInvalidLeadingRegex = regexp.MustCompile(`^[^a-zA-Z0-9]`)
var labelInvalidTrailingRegex = regexp.MustCompile(`[^a-zA-Z0-9]$`)

// SanitizeLabelValue coerces an arbitrary string into a valid Kubernetes label value:
// at most 63 characters, made up of alphanumerics, dots, dashes and underscores, and
// beginning and ending with an alphanumeric character. An empty input stays empty so
// callers can decide whether to stamp the label at all.
func SanitizeLabelValue(value string) string {
	if value == "" {
		return ""
	}

	sanitized := labelInvalidCharsRegex.ReplaceAllString(value, "-")
	if len(sanitized) > validation.LabelValueMaxLength {
		sanitized = sanitized[:validation.LabelValueMaxLength]
	}
	if labelInvalidLeadingRegex.MatchString(sanitized) {
		sanitized = "x" + sanitized[1:]
	}
	if labelInvalidTrailingRegex.MatchString(sanitized) {
		sanitized = sanitized[:len(sanitized)-1] + "x"
	}
	return sanitized
}
