package secret

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ToImagePullK8sName(t *testing.T) {
	tests := []struct {
		name           string
		components     SecretNameComponents
		expectedOutput string
	}{
		{
			name: "basic case",
			components: SecretNameComponents{
				Org:     "myorg",
				Domain:  "development",
				Project: "myproject",
				Name:    "docker-registry",
			},
			expectedOutput: "img-pull-93ec8baf1e7d7e87",
		},
		{
			name: "empty org",
			components: SecretNameComponents{
				Org:     "",
				Domain:  "development",
				Project: "myproject",
				Name:    "docker-registry",
			},
			expectedOutput: "img-pull-6d48430fc6e4845c",
		},
		{
			name: "all empty",
			components: SecretNameComponents{
				Org:     "",
				Domain:  "",
				Project: "",
				Name:    "",
			},
			expectedOutput: "img-pull-f1ae2a75ed1f9972",
		},
		{
			name: "special characters",
			components: SecretNameComponents{
				Org:     "my-org",
				Domain:  "dev.domain",
				Project: "project_123",
				Name:    "docker@registry",
			},
			expectedOutput: "img-pull-e0a5894d0b16a3c7",
		},
		{
			name: "long values",
			components: SecretNameComponents{
				Org:     "very-long-organization-name-that-exceeds-normal-length",
				Domain:  "development-domain-with-many-characters",
				Project: "project-with-extremely-long-identifier-for-testing",
				Name:    "docker-registry-with-very-long-descriptive-name",
			},
			expectedOutput: "img-pull-b1274c388a3edc4e",
		},
	}

	t.Run("conversion", func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := ToImagePullK8sName(tt.components)
				assert.Equal(t, tt.expectedOutput, result)

				// Verify the format meets Kubernetes naming requirements
				assert.LessOrEqual(t, len(result), 63, "Name should be 63 chars or less")
				assert.Regexp(t, "^[a-z0-9]([-a-z0-9]*[a-z0-9])?$", result, "Name should match Kubernetes DNS naming convention")
			})
		}
	})

	t.Run("should yield consistent results", func(t *testing.T) {
		components := SecretNameComponents{
			Org:     "testorg",
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
				Org:     "myorg",
				Domain:  "development",
				Project: "myproject",
				Name:    "docker-registry",
			},
			expectedLabels: map[string]string{
				secretTypeLabel: "image-pull-secret",
				orgLabel:        "myorg",
				domainLabel:     "development",
				projectLabel:    "myproject",
				secretNameLabel: "docker-registry",
			},
		},
		{
			name: "empty values",
			components: SecretNameComponents{
				Org:     "",
				Domain:  "",
				Project: "",
				Name:    "",
			},
			expectedLabels: map[string]string{
				secretTypeLabel: "image-pull-secret",
				orgLabel:        "none",
				domainLabel:     "none",
				projectLabel:    "none",
				secretNameLabel: "none",
			},
		},
		{
			name: "special characters",
			components: SecretNameComponents{
				Org:     "my@org",
				Domain:  "dev.domain!",
				Project: "project_123&",
				Name:    "docker@registry",
			},
			expectedLabels: map[string]string{
				secretTypeLabel: "image-pull-secret",
				orgLabel:        "my-org",
				domainLabel:     "dev.domainx",
				projectLabel:    "project_123x",
				secretNameLabel: "docker-registry",
			},
		},
		{
			name: "long values",
			components: SecretNameComponents{
				Org:     "very-long-organization-name-that-exceeds-normal-length-and-should-be-truncated",
				Domain:  "development-domain-with-many-characters-that-will-need-to-be-shortened",
				Project: "project-with-extremely-long-identifier-for-testing-purposes-only",
				Name:    "docker-registry-with-very-long-descriptive-name-that-needs-truncation",
			},
			expectedLabels: map[string]string{
				secretTypeLabel: "image-pull-secret",
				orgLabel:        "very-long-organization-name-that-exceeds-normal-length-and-shou",
				domainLabel:     "development-domain-with-many-characters-that-will-need-to-be-sh",
				projectLabel:    "project-with-extremely-long-identifier-for-testing-purposes-onl",
				secretNameLabel: "docker-registry-with-very-long-descriptive-name-that-needs-trun",
			},
		},
		{
			name: "sanitization - starts with invalid character",
			components: SecretNameComponents{
				Org:     "-invalid-start",
				Domain:  "normal",
				Project: "normal",
				Name:    "normal",
			},
			expectedLabels: map[string]string{
				secretTypeLabel: "image-pull-secret",
				orgLabel:        "xinvalid-start",
				domainLabel:     "normal",
				projectLabel:    "normal",
				secretNameLabel: "normal",
			},
		},
		{
			name: "sanitization - ends with invalid character",
			components: SecretNameComponents{
				Org:     "normal",
				Domain:  "normal",
				Project: "invalid-end-",
				Name:    "normal",
			},
			expectedLabels: map[string]string{
				secretTypeLabel: "image-pull-secret",
				orgLabel:        "normal",
				domainLabel:     "normal",
				projectLabel:    "invalid-endx",
				secretNameLabel: "normal",
			},
		},
		{
			name: "sanitization - invalid characters",
			components: SecretNameComponents{
				Org:     "normal",
				Domain:  "normal",
				Project: "normal",
				Name:    "invalid@#$%^&*()value",
			},
			expectedLabels: map[string]string{
				secretTypeLabel: "image-pull-secret",
				orgLabel:        "normal",
				domainLabel:     "normal",
				projectLabel:    "normal",
				secretNameLabel: "invalid---------value",
			},
		},
		{
			name: "sanitization - too long and ends with invalid after truncation",
			components: SecretNameComponents{
				Org:     "normal",
				Domain:  "this-is-a-very-long-label-value-that-exceeds-the-kubernetes-limit-of-sixty-three-",
				Project: "normal",
				Name:    "normal",
			},
			expectedLabels: map[string]string{
				secretTypeLabel: "image-pull-secret",
				orgLabel:        "normal",
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
