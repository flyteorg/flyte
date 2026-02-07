package secret

import (
	"crypto/sha256"
	"fmt"
	"regexp"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/utils"
)

const (
	imagePullPrefix = "img-pull"
)

const (
	orgLabel        = "union.ai/org"
	projectLabel    = "union.ai/project"
	domainLabel     = "union.ai/domain"
	secretNameLabel = "union.ai/secret-name"
	secretTypeLabel = "union.ai/secret-type"
)

var (
	ImagePullLabels = map[string]string{
		secretTypeLabel: "image-pull-secret",
	}
)

// ToImagePullK8sName generates a Kubernetes secret name based on the provided secret name components.
// The name includes a consistent hash of the components to ensure uniqueness, be Kubernetes compliant, and avoid collisions.
func ToImagePullK8sName(components SecretNameComponents) string {
	// Create a deterministic string representation of the components
	componentStr := fmt.Sprintf("%s:%s:%s:%s",
		components.Org,
		components.Domain,
		components.Project,
		components.Name)

	// Create a hash of the components
	h := sha256.New()
	h.Write([]byte(componentStr))
	hash := fmt.Sprintf("%x", h.Sum(nil))
	return fmt.Sprintf("%s-%s", imagePullPrefix, hash[:16])
}

// ToImagePullK8sLabels generates a map of labels that can be used to identify the image pull Kubernetes secret.
// These labels are intended to supplement the hashed secret name and provide additional metadata.
func ToImagePullK8sLabels(components SecretNameComponents) map[string]string {
	return utils.UnionMaps(ImagePullLabels, map[string]string{
		// TODO Create a function that cleans the values of the labels to be Kubernetes compliant
		orgLabel:        sanitizeLabelValue(components.Org),
		projectLabel:    sanitizeLabelValue(components.Project),
		domainLabel:     sanitizeLabelValue(components.Domain),
		secretNameLabel: sanitizeLabelValue(components.Name),
	})
}

// sanitizeLabelValue ensures that a label value conforms to Kubernetes label value requirements:
// - must be 63 characters or less
// - must begin and end with an alphanumeric character
// - may contain dots, dashes, and underscores
func sanitizeLabelValue(value string) string {
	// If empty, return a default value
	if value == "" {
		return "none"
	}

	// Replace invalid characters with dashes
	re := regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	sanitized := re.ReplaceAllString(value, "-")

	// Ensure it starts with an alphanumeric character
	re = regexp.MustCompile(`^[^a-zA-Z0-9]`)
	if re.MatchString(sanitized) {
		sanitized = "x" + sanitized[1:]
	}

	// Ensure it ends with an alphanumeric character
	re = regexp.MustCompile(`[^a-zA-Z0-9]$`)
	if re.MatchString(sanitized) {
		sanitized = sanitized[:len(sanitized)-1] + "x"
	}

	// Truncate to 63 characters if needed
	if len(sanitized) > 63 {
		sanitized = sanitized[:63]

		// After truncation, ensure it still ends with an alphanumeric character
		re = regexp.MustCompile(`[^a-zA-Z0-9]$`)
		if re.MatchString(sanitized) {
			sanitized = sanitized[:len(sanitized)-1] + "x"
		}
	}

	return sanitized
}
