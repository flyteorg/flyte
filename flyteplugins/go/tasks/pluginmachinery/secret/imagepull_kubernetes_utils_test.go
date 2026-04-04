package secret

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ToImagePullK8sName(t *testing.T) {
	tests := []struct {
		name       string
		components SecretNameComponents
	}{
		{
			name: "basic case",
			components: SecretNameComponents{
				Domain:  "development",
				Project: "myproject",
				Name:    "docker-registry",
			},
		},
		{
			name: "all empty",
			components: SecretNameComponents{
				Domain:  "",
				Project: "",
				Name:    "",
			},
		},
		{
			name: "special characters",
			components: SecretNameComponents{
				Domain:  "dev.domain",
				Project: "project_123",
				Name:    "docker@registry",
			},
		},
		{
			name: "long values",
			components: SecretNameComponents{
				Domain:  "development-domain-with-many-characters",
				Project: "project-with-extremely-long-identifier-for-testing",
				Name:    "docker-registry-with-very-long-descriptive-name",
			},
		},
	}

	t.Run("conversion", func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := ToImagePullK8sName(tt.components)

				// Verify the format meets Kubernetes naming requirements
				assert.LessOrEqual(t, len(result), 63, "Name should be 63 chars or less")
				assert.Regexp(t, "^[a-z0-9]([-a-z0-9]*[a-z0-9])?$", result, "Name should match Kubernetes DNS naming convention")
			})
		}
	})

	t.Run("should yield consistent results", func(t *testing.T) {
		components := SecretNameComponents{
			Domain:  "testdomain",
			Project: "testproject",
			Name:    "testsecret",
		}

		firstResult := ToImagePullK8sName(components)

		// Call multiple times to ensure consistency
		for i := 0; i < 10; i++ {
			result := ToImagePullK8sName(components)
			assert.Equal(t, firstResult, result, "Hash result should be consistent across calls")
		}
	})
}

func Test_ToImagePullK8sLabels(t *testing.T) {
	tests := []struct {
		name           string
		components     SecretNameComponents
		expectedLabels map[string]string
	}{
		{
			name: "basic case",
			components: SecretNameComponents{
				Domain:  "development",
				Project: "myproject",
				Name:    "docker-registry",
			},
			expectedLabels: map[string]string{
				secretTypeLabel: "image-pull-secret",
				domainLabel:     "development",
				projectLabel:    "myproject",
				secretNameLabel: "docker-registry",
			},
		},
		{
			name: "empty values",
			components: SecretNameComponents{
				Domain:  "",
				Project: "",
				Name:    "",
			},
			expectedLabels: map[string]string{
				secretTypeLabel: "image-pull-secret",
				domainLabel:     "none",
				projectLabel:    "none",
				secretNameLabel: "none",
			},
		},
		{
			name: "special characters",
			components: SecretNameComponents{
				Domain:  "normal",
				Project: "normal",
				Name:    "invalid@#$%^&*()value",
			},
			expectedLabels: map[string]string{
				secretTypeLabel: "image-pull-secret",
				domainLabel:     "normal",
				projectLabel:    "normal",
				secretNameLabel: "invalid---------value",
			},
		},
		{
			name: "long values",
			components: SecretNameComponents{
				Domain:  "development-domain-with-many-characters-that-will-need-to-be-shortened",
				Project: "project-with-extremely-long-identifier-for-testing-purposes-only",
				Name:    "docker-registry-with-very-long-descriptive-name-that-needs-truncation",
			},
			expectedLabels: map[string]string{
				secretTypeLabel: "image-pull-secret",
				domainLabel:     "development-domain-with-many-characters-that-will-need-to-be-sh",
				projectLabel:    "project-with-extremely-long-identifier-for-testing-purposes-onl",
				secretNameLabel: "docker-registry-with-very-long-descriptive-name-that-needs-trun",
			},
		},
		{
			name: "sanitization - starts with invalid character",
			components: SecretNameComponents{
				Domain:  "normal",
				Project: "normal",
				Name:    "-invalid-start",
			},
			expectedLabels: map[string]string{
				secretTypeLabel: "image-pull-secret",
				domainLabel:     "normal",
				projectLabel:    "normal",
				secretNameLabel: "xinvalid-start",
			},
		},
		{
			name: "sanitization - ends with invalid character",
			components: SecretNameComponents{
				Domain:  "normal",
				Project: "invalid-end-",
				Name:    "normal",
			},
			expectedLabels: map[string]string{
				secretTypeLabel: "image-pull-secret",
				domainLabel:     "normal",
				projectLabel:    "invalid-endx",
				secretNameLabel: "normal",
			},
		},
		{
			name: "sanitization - too long and ends with invalid after truncation",
			components: SecretNameComponents{
				Domain:  "this-is-a-very-long-label-value-that-exceeds-the-kubernetes-limit-of-sixty-three-",
				Project: "normal",
				Name:    "normal",
			},
			expectedLabels: map[string]string{
				secretTypeLabel: "image-pull-secret",
				domainLabel:     "this-is-a-very-long-label-value-that-exceeds-the-kubernetes-lim",
				projectLabel:    "normal",
				secretNameLabel: "normal",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			labels := ToImagePullK8sLabels(tt.components)
			assert.Equal(t, tt.expectedLabels, labels)

			// Verify all label values meet Kubernetes requirements
			for k, value := range labels {
				assert.LessOrEqual(t, len(value), 63, "Label value should be 63 chars or less: %s", k)
				assert.Regexp(t, "^[a-zA-Z0-9]([-_.a-zA-Z0-9]*[a-zA-Z0-9])?$", value,
					"Label value should match Kubernetes label value convention: %s", k)
			}
		})
	}
}
