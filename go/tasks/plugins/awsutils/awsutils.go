package awsutils

import "context"

func GetRole(_ context.Context, roleAnnotationKey string, annotations map[string]string) string {
	if len(roleAnnotationKey) > 0 {
		return annotations[roleAnnotationKey]
	}

	return ""
}
