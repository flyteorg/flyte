package webhook

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	promtestutil "github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyte/flytepropeller/pkg/secret/config"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

const (
	org                 = "test-org"
	otherOrg            = "test-other-org"
	existingHostname    = "test.original.hostname"
	replacementHostname = "test.replacement.hostname"
)

var (
	differentImage      = fmt.Sprintf("%s/orgs/%s/other-image", replacementHostname, org)
	unusedLabelSelector = metav1.LabelSelector{}
	invalidImageNames   = []string{
		// Orgs that share similar prefixes and suffixes
		fmt.Sprintf("%s/%s-suffix/image", existingHostname, org),
		fmt.Sprintf("%s/prefix-%s/image", existingHostname, org),
		fmt.Sprintf("%s/%s/image", existingHostname, otherOrg),
	}
	replacedInvalidImageNames = []string{
		fmt.Sprintf("%s/orgs/%s-suffix/image", replacementHostname, org),
		fmt.Sprintf("%s/orgs/prefix-%s/image", replacementHostname, org),
		fmt.Sprintf("%s/orgs/%s/image", replacementHostname, otherOrg),
	}
	emptyLabels = prometheus.Labels{}
)

func defaultTestImageBuilderMutator() *ImageBuilderMutatorV1 {
	return NewImageBuilderMutator(&config.ImageBuilderConfig{
		HostnameReplacement: config.HostnameReplacement{
			Existing:    existingHostname,
			Replacement: replacementHostname,
		},
		LabelSelector: unusedLabelSelector,
	}, promutils.NewTestScope())
}

func extractReplacedMetric(c prometheus.CounterVec) int {
	return int(promtestutil.ToFloat64(c.With(emptyLabels)))
}

func retrieveReplacedMetrics(m *ImageBuilderMutatorV1) (unversionedPublicImages int, v1PublicImages int, unversionedOrgImages int, v1OrgImages int) {
	unversionedPublicImagesReplaced := extractReplacedMetric(m.metrics.Replaced.PublicImages.Unversioned)
	v1PublicImagesReplaced := extractReplacedMetric(m.metrics.Replaced.PublicImages.V1)
	unversionedOrgImagesReplaced := extractReplacedMetric(m.metrics.Replaced.OrgImages.Unversioned)
	v1OrgImagesReplaced := extractReplacedMetric(m.metrics.Replaced.OrgImages.V1)
	return unversionedPublicImagesReplaced, v1PublicImagesReplaced, unversionedOrgImagesReplaced, v1OrgImagesReplaced
}

func TestImageBuilderWebhook_Mutate(t *testing.T) {

	t.Run("Valid hostnames", func(t *testing.T) {

		validImageNames := []string{
			// URL Paths without version. Backwards compatible with older
			// versions of unionai SDK
			// Build-image container task
			fmt.Sprintf("%s/cloud/task", replacementHostname),
			// Any image in users org
			fmt.Sprintf("%s/orgs/%s/image", replacementHostname, org),
			fmt.Sprintf("%s/orgs/%s/other-image", replacementHostname, org),
			// Any image in union publicly accessible namespace
			fmt.Sprintf("%s/union/image", replacementHostname),
			fmt.Sprintf("%s/union/other-image", replacementHostname),

			// Version 1 URI paths
			fmt.Sprintf("%s/%s/cloud/task", replacementHostname, version1URIPart),
			// Any image in users org
			fmt.Sprintf("%s/%s/orgs/%s/image", replacementHostname, version1URIPart, org),
			fmt.Sprintf("%s/%s/orgs/%s/other-image", replacementHostname, version1URIPart, org),
			// Any image in union publicly accessible namespace
			fmt.Sprintf("%s/%s/union/image", replacementHostname, version1URIPart),
			fmt.Sprintf("%s/%s/union/other-image", replacementHostname, version1URIPart),
		}
		for _, validImageName := range validImageNames {
			m := defaultTestImageBuilderMutator()
			pod := corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: org,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "testcontainer1",
							Image: validImageName,
						},
					},
				},
			}
			returnedPod, changed, err := m.Mutate(context.Background(), &pod)
			assert.Nil(t, err)
			assert.False(t, changed, fmt.Sprintf("Expected no change for %s", validImageName))
			assert.Equal(t, &pod, returnedPod, fmt.Sprintf("Expected no Pod differences for %s", validImageName))
			assert.Equal(t, validImageName, returnedPod.Spec.Containers[0].Image, fmt.Sprintf("Expected image name to be left at %s", validImageName))
			assert.Equal(t, 1, int(promtestutil.ToFloat64(m.metrics.Attempts)))
			assert.Equal(t, 0, int(promtestutil.ToFloat64(m.metrics.Failures)))
			assert.Equal(t, 1, int(promtestutil.ToFloat64(m.metrics.V1ContainerValidationAttempts)))
			assert.Equal(t, 0, int(promtestutil.ToFloat64(m.metrics.V1ContainerValidationFailures)))
			unversionedPublicImagesReplaced, v1PublicImagesReplaced, unversionedOrgImagesReplaced, v1OrgImagesReplaced := retrieveReplacedMetrics(m)
			assert.Equal(t, 0, unversionedPublicImagesReplaced)
			assert.Equal(t, 0, v1PublicImagesReplaced)
			assert.Equal(t, 0, unversionedOrgImagesReplaced)
			assert.Equal(t, 0, v1OrgImagesReplaced)
		}
	})

	t.Run("Replaces hostname for pre-versioned union public images", func(t *testing.T) {
		m := defaultTestImageBuilderMutator()
		existingImage := fmt.Sprintf("%s/%s/image", existingHostname, unionaiOrgPlaceholder)
		expectedImage := fmt.Sprintf("%s/%s/image", replacementHostname, unionPathReplacement)
		otherImage := differentImage
		pod := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: org,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "testcontainer1",
						Image: existingImage,
					},
					{
						Name:  "testcontainer2",
						Image: otherImage,
					},
				},
			},
		}
		returnedPod, changed, err := m.Mutate(context.Background(), &pod)
		assert.Nil(t, err)
		assert.True(t, changed)
		assert.Equal(t, &pod, returnedPod)
		assert.Equal(t, expectedImage, returnedPod.Spec.Containers[0].Image)
		assert.Equal(t, otherImage, returnedPod.Spec.Containers[1].Image)
		assert.Equal(t, 1, int(promtestutil.ToFloat64(m.metrics.Attempts)))
		assert.Equal(t, 0, int(promtestutil.ToFloat64(m.metrics.Failures)))
		assert.Equal(t, 2, int(promtestutil.ToFloat64(m.metrics.V1ContainerValidationAttempts)))
		assert.Equal(t, 0, int(promtestutil.ToFloat64(m.metrics.V1ContainerValidationFailures)))
		unversionedPublicImagesReplaced, v1PublicImagesReplaced, unversionedOrgImagesReplaced, v1OrgImagesReplaced := retrieveReplacedMetrics(m)
		assert.Equal(t, 1, unversionedPublicImagesReplaced)
		assert.Equal(t, 0, v1PublicImagesReplaced)
		assert.Equal(t, 0, unversionedOrgImagesReplaced)
		assert.Equal(t, 0, v1OrgImagesReplaced)
	})

	t.Run("Replaces hostname for versioned union public images", func(t *testing.T) {
		m := defaultTestImageBuilderMutator()
		existingImage := fmt.Sprintf("%s/%s/%s/image", existingHostname, version1URIPart, unionaiOrgPlaceholder)
		expectedImage := fmt.Sprintf("%s/%s/image", replacementHostname, unionPathReplacement)
		otherImage := differentImage
		pod := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: org,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "testcontainer1",
						Image: existingImage,
					},
					{
						Name:  "testcontainer2",
						Image: otherImage,
					},
				},
			},
		}
		returnedPod, changed, err := m.Mutate(context.Background(), &pod)
		assert.Nil(t, err)
		assert.True(t, changed)
		assert.Equal(t, &pod, returnedPod)
		assert.Equal(t, expectedImage, returnedPod.Spec.Containers[0].Image)
		assert.Equal(t, otherImage, returnedPod.Spec.Containers[1].Image)
		assert.Equal(t, 1, int(promtestutil.ToFloat64(m.metrics.Attempts)))
		assert.Equal(t, 0, int(promtestutil.ToFloat64(m.metrics.Failures)))
		assert.Equal(t, 2, int(promtestutil.ToFloat64(m.metrics.V1ContainerValidationAttempts)))
		assert.Equal(t, 0, int(promtestutil.ToFloat64(m.metrics.V1ContainerValidationFailures)))
		unversionedPublicImagesReplaced, v1PublicImagesReplaced, unversionedOrgImagesReplaced, v1OrgImagesReplaced := retrieveReplacedMetrics(m)
		assert.Equal(t, 0, unversionedPublicImagesReplaced)
		assert.Equal(t, 1, v1PublicImagesReplaced)
		assert.Equal(t, 0, unversionedOrgImagesReplaced)
		assert.Equal(t, 0, v1OrgImagesReplaced)
	})

	t.Run("Preserves hostname org images with older unionai SDK and build-image task versions", func(t *testing.T) {
		m := defaultTestImageBuilderMutator()
		// In this scenario, unionai SDK is assumed to do the replacement and no version number gets applied
		existingImage := fmt.Sprintf("%s/orgs/%s/image", replacementHostname, org)
		expectedImage := existingImage
		otherImage := differentImage
		pod := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: org,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "testcontainer1",
						Image: existingImage,
					},
					{
						Name:  "testcontainer2",
						Image: otherImage,
					},
				},
			},
		}
		returnedPod, changed, err := m.Mutate(context.Background(), &pod)
		assert.Nil(t, err)
		assert.False(t, changed)
		assert.Equal(t, &pod, returnedPod)
		assert.Equal(t, expectedImage, returnedPod.Spec.Containers[0].Image)
		assert.Equal(t, otherImage, returnedPod.Spec.Containers[1].Image)
		assert.Equal(t, 1, int(promtestutil.ToFloat64(m.metrics.Attempts)))
		assert.Equal(t, 0, int(promtestutil.ToFloat64(m.metrics.Failures)))
		assert.Equal(t, 2, int(promtestutil.ToFloat64(m.metrics.V1ContainerValidationAttempts)))
		assert.Equal(t, 0, int(promtestutil.ToFloat64(m.metrics.V1ContainerValidationFailures)))
		unversionedPublicImagesReplaced, v1PublicImagesReplaced, unversionedOrgImagesReplaced, v1OrgImagesReplaced := retrieveReplacedMetrics(m)
		assert.Equal(t, 0, unversionedPublicImagesReplaced)
		assert.Equal(t, 0, v1PublicImagesReplaced)
		assert.Equal(t, 0, unversionedOrgImagesReplaced)
		assert.Equal(t, 0, v1OrgImagesReplaced)
	})

	t.Run("Replaces hostname org images with newer unionai SDK and older build-image task version", func(t *testing.T) {
		m := defaultTestImageBuilderMutator()
		existingImage := fmt.Sprintf("%s/%s/image", existingHostname, org)
		expectedImage := fmt.Sprintf("%s/orgs/%s/image", replacementHostname, org)
		otherImage := differentImage
		pod := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: org,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "testcontainer1",
						Image: existingImage,
					},
					{
						Name:  "testcontainer2",
						Image: otherImage,
					},
				},
			},
		}
		returnedPod, changed, err := m.Mutate(context.Background(), &pod)
		assert.Nil(t, err)
		assert.True(t, changed)
		assert.Equal(t, &pod, returnedPod)
		assert.Equal(t, expectedImage, returnedPod.Spec.Containers[0].Image)
		assert.Equal(t, otherImage, returnedPod.Spec.Containers[1].Image)
		assert.Equal(t, 1, int(promtestutil.ToFloat64(m.metrics.Attempts)))
		assert.Equal(t, 0, int(promtestutil.ToFloat64(m.metrics.Failures)))
		assert.Equal(t, 2, int(promtestutil.ToFloat64(m.metrics.V1ContainerValidationAttempts)))
		assert.Equal(t, 0, int(promtestutil.ToFloat64(m.metrics.V1ContainerValidationFailures)))
		unversionedPublicImagesReplaced, v1PublicImagesReplaced, unversionedOrgImagesReplaced, v1OrgImagesReplaced := retrieveReplacedMetrics(m)
		assert.Equal(t, 0, unversionedPublicImagesReplaced)
		assert.Equal(t, 0, v1PublicImagesReplaced)
		assert.Equal(t, 1, unversionedOrgImagesReplaced)
		assert.Equal(t, 0, v1OrgImagesReplaced)
	})

	t.Run("Replaces hostname org images with older unionai SDK and newer build-image task version", func(t *testing.T) {
		m := defaultTestImageBuilderMutator()
		// In this scenario, unionai SDK does hostname replacing and i
		existingImage := fmt.Sprintf("%s/orgs/v1/%s/image", replacementHostname, org)
		expectedImage := fmt.Sprintf("%s/orgs/%s/image", replacementHostname, org)
		otherImage := differentImage
		pod := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: org,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "testcontainer1",
						Image: existingImage,
					},
					{
						Name:  "testcontainer2",
						Image: otherImage,
					},
				},
			},
		}
		returnedPod, changed, err := m.Mutate(context.Background(), &pod)
		assert.Nil(t, err)
		assert.True(t, changed)
		assert.Equal(t, &pod, returnedPod)
		assert.Equal(t, expectedImage, returnedPod.Spec.Containers[0].Image)
		assert.Equal(t, otherImage, returnedPod.Spec.Containers[1].Image)
		assert.Equal(t, 1, int(promtestutil.ToFloat64(m.metrics.Attempts)))
		assert.Equal(t, 0, int(promtestutil.ToFloat64(m.metrics.Failures)))
		assert.Equal(t, 2, int(promtestutil.ToFloat64(m.metrics.V1ContainerValidationAttempts)))
		assert.Equal(t, 0, int(promtestutil.ToFloat64(m.metrics.V1ContainerValidationFailures)))
		unversionedPublicImagesReplaced, v1PublicImagesReplaced, unversionedOrgImagesReplaced, v1OrgImagesReplaced := retrieveReplacedMetrics(m)
		assert.Equal(t, 0, unversionedPublicImagesReplaced)
		assert.Equal(t, 0, v1PublicImagesReplaced)
		assert.Equal(t, 0, unversionedOrgImagesReplaced)
		assert.Equal(t, 1, v1OrgImagesReplaced)
	})

	t.Run("Replaces hostname org images with newer unionai SDK and older build-image task version", func(t *testing.T) {
		m := defaultTestImageBuilderMutator()
		// In this scenario, unionai SDK does not do any hostname replacement.
		existingImage := fmt.Sprintf("%s/%s/image", existingHostname, org)
		expectedImage := fmt.Sprintf("%s/orgs/%s/image", replacementHostname, org)
		otherImage := differentImage
		pod := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: org,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "testcontainer1",
						Image: existingImage,
					},
					{
						Name:  "testcontainer2",
						Image: otherImage,
					},
				},
			},
		}
		returnedPod, changed, err := m.Mutate(context.Background(), &pod)
		assert.Nil(t, err)
		assert.True(t, changed)
		assert.Equal(t, &pod, returnedPod)
		assert.Equal(t, expectedImage, returnedPod.Spec.Containers[0].Image)
		assert.Equal(t, otherImage, returnedPod.Spec.Containers[1].Image)
		assert.Equal(t, 1, int(promtestutil.ToFloat64(m.metrics.Attempts)))
		assert.Equal(t, 0, int(promtestutil.ToFloat64(m.metrics.Failures)))
		assert.Equal(t, 2, int(promtestutil.ToFloat64(m.metrics.V1ContainerValidationAttempts)))
		assert.Equal(t, 0, int(promtestutil.ToFloat64(m.metrics.V1ContainerValidationFailures)))
		unversionedPublicImagesReplaced, v1PublicImagesReplaced, unversionedOrgImagesReplaced, v1OrgImagesReplaced := retrieveReplacedMetrics(m)
		assert.Equal(t, 0, unversionedPublicImagesReplaced)
		assert.Equal(t, 0, v1PublicImagesReplaced)
		assert.Equal(t, 1, unversionedOrgImagesReplaced)
		assert.Equal(t, 0, v1OrgImagesReplaced)
	})

	t.Run("Replaces hostname org images with newer unionai SDK and build-image task version", func(t *testing.T) {
		m := defaultTestImageBuilderMutator()
		existingImage := fmt.Sprintf("%s/%s/%s/image", existingHostname, version1URIPart, org)
		expectedImage := fmt.Sprintf("%s/orgs/%s/image", replacementHostname, org)
		otherImage := differentImage
		pod := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: org,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "testcontainer1",
						Image: existingImage,
					},
					{
						Name:  "testcontainer2",
						Image: otherImage,
					},
				},
			},
		}
		returnedPod, changed, err := m.Mutate(context.Background(), &pod)
		assert.Nil(t, err)
		assert.True(t, changed)
		assert.Equal(t, &pod, returnedPod)
		assert.Equal(t, expectedImage, returnedPod.Spec.Containers[0].Image)
		assert.Equal(t, otherImage, returnedPod.Spec.Containers[1].Image)
		assert.Equal(t, 1, int(promtestutil.ToFloat64(m.metrics.Attempts)))
		assert.Equal(t, 0, int(promtestutil.ToFloat64(m.metrics.Failures)))
		assert.Equal(t, 2, int(promtestutil.ToFloat64(m.metrics.V1ContainerValidationAttempts)))
		assert.Equal(t, 0, int(promtestutil.ToFloat64(m.metrics.V1ContainerValidationFailures)))
		unversionedPublicImagesReplaced, v1PublicImagesReplaced, unversionedOrgImagesReplaced, v1OrgImagesReplaced := retrieveReplacedMetrics(m)
		assert.Equal(t, 0, unversionedPublicImagesReplaced)
		assert.Equal(t, 0, v1PublicImagesReplaced)
		assert.Equal(t, 0, unversionedOrgImagesReplaced)
		assert.Equal(t, 1, v1OrgImagesReplaced)
	})

	t.Run("Replaces multiple hostname match occurrence", func(t *testing.T) {
		m := defaultTestImageBuilderMutator()
		originalImages := make([]string, 3)
		expectedImages := make([]string, 3)
		for i := 0; i < 3; i++ {
			originalImages[i] = fmt.Sprintf("%s/%s/image-%d", existingHostname, org, i)
			expectedImages[i] = fmt.Sprintf("%s/orgs/%s/image-%d", replacementHostname, org, i)
		}

		otherImage := differentImage
		pod := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: org,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "testcontainer1",
						Image: originalImages[0],
					},
					{
						Name:  "testcontainer2",
						Image: originalImages[1],
					},
					{
						Name:  "testcontainer3",
						Image: originalImages[2],
					},
					{
						Name:  "testunchangedcontainer",
						Image: otherImage,
					},
				},
			},
		}
		returnedPod, changed, err := m.Mutate(context.Background(), &pod)
		assert.Nil(t, err)
		assert.True(t, changed)
		assert.Equal(t, &pod, returnedPod)
		for i := 0; i < 3; i++ {
			assert.Equal(t, expectedImages[i], returnedPod.Spec.Containers[i].Image)
		}
		assert.Equal(t, otherImage, returnedPod.Spec.Containers[3].Image)
		assert.Equal(t, 1, int(promtestutil.ToFloat64(m.metrics.Attempts)))
		assert.Equal(t, 0, int(promtestutil.ToFloat64(m.metrics.Failures)))
		assert.Equal(t, 4, int(promtestutil.ToFloat64(m.metrics.V1ContainerValidationAttempts)))
		assert.Equal(t, 0, int(promtestutil.ToFloat64(m.metrics.V1ContainerValidationFailures)))
		unversionedPublicImagesReplaced, v1PublicImagesReplaced, unversionedOrgImagesReplaced, v1OrgImagesReplaced := retrieveReplacedMetrics(m)
		assert.Equal(t, 0, unversionedPublicImagesReplaced)
		assert.Equal(t, 0, v1PublicImagesReplaced)
		assert.Equal(t, len(expectedImages), unversionedOrgImagesReplaced)
		assert.Equal(t, 0, v1OrgImagesReplaced)
	})

	t.Run("Rejects different organization", func(t *testing.T) {
		otherOrg := "other-org"
		m := defaultTestImageBuilderMutator()
		originalImage := fmt.Sprintf("%s/%s/image", otherOrg, existingHostname)
		pod := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: org,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "testcontainer1",
						Image: originalImage,
					},
				},
			},
		}
		returnedPod, changed, err := m.Mutate(context.Background(), &pod)
		assert.NotNil(t, err)
		assert.Equal(t, http.StatusForbidden, int(err.Result.Code))
		assert.Equal(t, fmt.Sprintf("access to %s is not authorized", originalImage), err.Result.Message)
		assert.False(t, changed)
		assert.Nil(t, returnedPod)
		assert.Equal(t, 1, int(promtestutil.ToFloat64(m.metrics.Attempts)))
		assert.Equal(t, 1, int(promtestutil.ToFloat64(m.metrics.Failures)))
		assert.Equal(t, 1, int(promtestutil.ToFloat64(m.metrics.V1ContainerValidationAttempts)))
		assert.Equal(t, 1, int(promtestutil.ToFloat64(m.metrics.V1ContainerValidationFailures)))
		unversionedPublicImagesReplaced, v1PublicImagesReplaced, unversionedOrgImagesReplaced, v1OrgImagesReplaced := retrieveReplacedMetrics(m)
		assert.Equal(t, 0, unversionedPublicImagesReplaced)
		assert.Equal(t, 0, v1PublicImagesReplaced)
		assert.Equal(t, 0, unversionedOrgImagesReplaced)
		assert.Equal(t, 0, v1OrgImagesReplaced)
	})

	t.Run("Rejects unauthorized paths", func(t *testing.T) {
		for _, invalidImageName := range invalidImageNames {
			m := defaultTestImageBuilderMutator()
			pod := corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: org,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "testcontainer1",
							Image: invalidImageName,
						},
					},
				},
			}
			returnedPod, changed, err := m.Mutate(context.Background(), &pod)
			assert.NotNil(t, err, fmt.Sprintf("Expected error for %s", invalidImageName))
			assert.Equal(t, http.StatusForbidden, int(err.Result.Code), fmt.Sprintf("Expected forbidden error code for %s", invalidImageName))
			assert.Equal(t, fmt.Sprintf("access to %s is not authorized", invalidImageName), err.Result.Message)
			assert.False(t, changed, fmt.Sprintf("Expected no change for %s", invalidImageName))
			assert.Nil(t, returnedPod, fmt.Sprintf("Expected nil returned pod pointer for %s", invalidImageName))
			assert.Equal(t, 1, int(promtestutil.ToFloat64(m.metrics.Attempts)))
			assert.Equal(t, 1, int(promtestutil.ToFloat64(m.metrics.Failures)))
			assert.Equal(t, 1, int(promtestutil.ToFloat64(m.metrics.V1ContainerValidationAttempts)))
			assert.Equal(t, 1, int(promtestutil.ToFloat64(m.metrics.V1ContainerValidationFailures)))
			unversionedPublicImagesReplaced, v1PublicImagesReplaced, unversionedOrgImagesReplaced, v1OrgImagesReplaced := retrieveReplacedMetrics(m)
			assert.Equal(t, 0, unversionedPublicImagesReplaced)
			assert.Equal(t, 0, v1PublicImagesReplaced)
			assert.Equal(t, 1, unversionedOrgImagesReplaced)
			assert.Equal(t, 0, v1OrgImagesReplaced)
		}
	})

	t.Run("Skips verification for invalid URI paths", func(t *testing.T) {
		for i, unVerifiedImageName := range invalidImageNames {
			m := NewImageBuilderMutator(&config.ImageBuilderConfig{
				HostnameReplacement: config.HostnameReplacement{
					Existing:            existingHostname,
					Replacement:         replacementHostname,
					DisableVerification: true,
				},
				LabelSelector: unusedLabelSelector,
			}, promutils.NewTestScope())
			pod := corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: org,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "testcontainer1",
							Image: unVerifiedImageName,
						},
					},
				},
			}
			returnedPod, changed, err := m.Mutate(context.Background(), &pod)
			assert.Nil(t, err, fmt.Sprintf("Expected no error for %s", unVerifiedImageName))
			assert.Equal(t, &pod, returnedPod, fmt.Sprintf("Expected no Pod address change for %s", unVerifiedImageName))
			assert.Equal(t, replacedInvalidImageNames[i], returnedPod.Spec.Containers[0].Image)
			assert.True(t, changed, fmt.Sprintf("Expected image still changed for %s", unVerifiedImageName))
			assert.Equal(t, 1, int(promtestutil.ToFloat64(m.metrics.Attempts)), fmt.Sprintf("Expected 1 attempt for %s", unVerifiedImageName))
			assert.Equal(t, 0, int(promtestutil.ToFloat64(m.metrics.Failures)), fmt.Sprintf("Expected 0 failures for %s", unVerifiedImageName))
			assert.Equal(t, 0, int(promtestutil.ToFloat64(m.metrics.V1ContainerValidationAttempts)), fmt.Sprintf("Expected 1 attempt for %s", unVerifiedImageName))
			assert.Equal(t, 0, int(promtestutil.ToFloat64(m.metrics.V1ContainerValidationFailures)), fmt.Sprintf("Expected 0 failures for %s", unVerifiedImageName))
			unversionedPublicImagesReplaced, v1PublicImagesReplaced, unversionedOrgImagesReplaced, v1OrgImagesReplaced := retrieveReplacedMetrics(m)
			assert.Equal(t, 0, unversionedPublicImagesReplaced)
			assert.Equal(t, 0, v1PublicImagesReplaced)
			assert.Equal(t, 1, unversionedOrgImagesReplaced)
			assert.Equal(t, 0, v1OrgImagesReplaced)
		}
	})

	t.Run("Skips verification for different host", func(t *testing.T) {
		// Pod uses a hostname different from hostnameReplacement.Existing
		otherHostname := "test.other.hostname"
		otherHostImageNames := []string{
			fmt.Sprintf("%s/cloud/task", otherHostname),
			fmt.Sprintf("%s/orgs/%s/image", otherHostname, org),
			fmt.Sprintf("%s/union/image", otherHostname),
			fmt.Sprintf("%s/%s/cloud/task", otherHostname, version1URIPart),
			fmt.Sprintf("%s/%s/orgs/%s/image", otherHostname, version1URIPart, org),
			fmt.Sprintf("%s/%s/union/image", otherHostname, version1URIPart),
		}
		for _, unVerifiedImageName := range otherHostImageNames {
			m := NewImageBuilderMutator(&config.ImageBuilderConfig{
				HostnameReplacement: config.HostnameReplacement{
					Existing:            existingHostname,
					Replacement:         replacementHostname,
					DisableVerification: true,
				},
				LabelSelector: unusedLabelSelector,
			}, promutils.NewTestScope())
			pod := corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: org,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "testcontainer1",
							Image: unVerifiedImageName,
						},
					},
				},
			}
			returnedPod, changed, err := m.Mutate(context.Background(), &pod)
			assert.Nil(t, err, fmt.Sprintf("Expected no error for %s", unVerifiedImageName))
			assert.Equal(t, &pod, returnedPod, fmt.Sprintf("Expected no Pod address change for %s", unVerifiedImageName))
			assert.Equal(t, unVerifiedImageName, returnedPod.Spec.Containers[0].Image)
			assert.False(t, changed, fmt.Sprintf("Expected no change for %s", unVerifiedImageName))
			assert.Equal(t, 1, int(promtestutil.ToFloat64(m.metrics.Attempts)), fmt.Sprintf("Expected 1 attempt for %s", unVerifiedImageName))
			assert.Equal(t, 0, int(promtestutil.ToFloat64(m.metrics.Failures)), fmt.Sprintf("Expected 0 failures for %s", unVerifiedImageName))
			assert.Equal(t, 0, int(promtestutil.ToFloat64(m.metrics.V1ContainerValidationAttempts)), fmt.Sprintf("Expected 1 attempt for %s", unVerifiedImageName))
			assert.Equal(t, 0, int(promtestutil.ToFloat64(m.metrics.V1ContainerValidationFailures)), fmt.Sprintf("Expected 0 failures for %s", unVerifiedImageName))
			unversionedPublicImagesReplaced, v1PublicImagesReplaced, unversionedOrgImagesReplaced, v1OrgImagesReplaced := retrieveReplacedMetrics(m)
			assert.Equal(t, 0, unversionedPublicImagesReplaced)
			assert.Equal(t, 0, v1PublicImagesReplaced)
			assert.Equal(t, 0, unversionedOrgImagesReplaced)
			assert.Equal(t, 0, v1OrgImagesReplaced)
		}
	})
}
