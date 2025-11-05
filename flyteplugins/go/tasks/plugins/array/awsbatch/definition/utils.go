package definition

import "strings"

// Replaces not allowed characters to ensure the resulting name conforms to:
// https://docs.aws.amazon.com/batch/latest/APIReference/API_RegisterJobDefinition.html. Specifically: The name of the
// job definition to register. Up to 128 letters (uppercase and lowercase), numbers, hyphens, and underscores are
// allowed.
func GetJobDefinitionSafeName(jobName string) string {
	sb := strings.Builder{}
	for _, c := range jobName {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' {
			sb.WriteRune(c)
		} else {
			sb.WriteRune('-')
		}
	}

	if sb.Len() < 128 {
		return sb.String()
	}

	return sb.String()[:128]
}
