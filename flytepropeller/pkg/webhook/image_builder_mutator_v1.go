package webhook

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/flyteorg/flyte/flytepropeller/pkg/secret/config"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

const (
	// Convention established with image builder that all orgs images are places under orgs namespace
	orgsNamespace = "orgs"
	// Convention where cr.union.ai/unionai/image-name map to <project + region hostname>/union/image-name
	unionaiOrgPlaceholder = "unionai"
	unionPathReplacement  = "union"
	// Version 1 URI format part that is expected in URI
	version1URIPart = "v1"
	// Currently excluding prefix to support versionless URIs
	v1CloudTaskPart   = "cloud/task"               // Build-image container task
	v1OrgsPart        = "orgs/%s/"                 // Any of the org's images, orgs/<namespace/org name>
	v1UnionPublicPart = unionPathReplacement + "/" // Any publicly accessible image

	ImageBuilderV1ID = "image-builder"

	promVersionLabelKey       = "version"
	promTypeLabelKey          = "type"
	promPublicTypeLabelValue  = "public"
	promOrgTypeLabelValue     = "org"
	promUnversionedLabelValue = "none"
	promV1LabelValue          = "v1"
)

type versionedCounters struct {
	Unversioned prometheus.CounterVec
	V1          prometheus.CounterVec
}

type replacementMetrics struct {
	PublicImages versionedCounters
	OrgImages    versionedCounters
}

type metrics struct {
	Scope                         promutils.Scope
	RoundTime                     promutils.StopWatch
	Attempts                      prometheus.Counter
	Failures                      prometheus.Counter
	V1ContainerValidationAttempts prometheus.Counter
	V1ContainerValidationFailures prometheus.Counter
	Replaced                      replacementMetrics
}

type ImageBuilderMutatorV1 struct {
	hostnameReplacement config.HostnameReplacement
	metrics             metrics
	labelSelector       metav1.LabelSelector

	// Prefixes used as replacement sources
	initialV1UnionPublicPrefixes          []string // For Org based images with v1 URI format
	initialUnversionedUnionPublicPrefixes []string // For union public images
	initialV1OrgPrefixes                  []string // For Org based images with v1 URI format
	initialUnversionedOrgPrefixes         []string // For Org based images with legacy unversioned format

	// Prefixes to be used for replacing hostnames
	targetPublicPrefix string
	targetOrgPrefix    string

	// Acceptable prefixes is a list of known prefixes
	// used for post replacement validation.
	acceptablePrefixes []string

	excludedContainers sets.Set[string]
}

func (i ImageBuilderMutatorV1) ID() string {
	return ImageBuilderV1ID
}

func (i *ImageBuilderMutatorV1) Mutate(ctx context.Context, pod *corev1.Pod) (newP *corev1.Pod, podChanged bool, errResponse *admission.Response) {
	t := i.metrics.RoundTime.Start()
	defer t.Stop()
	hr := i.hostnameReplacement

	logger.Debugf(ctx, "Replacing hostname [%v] with [%v] for Pod [%v/%v]", hr.Existing, hr.Replacement, pod.Namespace, pod.Name)
	i.metrics.Attempts.Inc()
	newContainers, changed, err := i.replaceHostnames(ctx, pod.Name, pod.Namespace, pod.Spec.Containers, hr)
	if err != nil { // Failed to replace or validate
		logger.Warnf(ctx, "Failed to replace container image names for pod [%v/%v] due to error: %w", pod.Namespace, pod.Name, err)
		i.metrics.Failures.Inc()
		admissionResponse := admission.Errored(http.StatusForbidden, err)
		return nil, false, &admissionResponse
	}
	pod.Spec.Containers = *newContainers

	newInitContainers, initContainersChanged, err := i.replaceHostnames(ctx, pod.Name, pod.Namespace, pod.Spec.InitContainers, hr)
	if err != nil { // Failed to replace or validate
		logger.Warnf(ctx, "Failed to replace init container image names for pod [%v/%v] due to error: %w", pod.Namespace, pod.Name, err)
		i.metrics.Failures.Inc()
		admissionResponse := admission.Errored(http.StatusForbidden, err)
		return nil, false, &admissionResponse
	}
	pod.Spec.InitContainers = *newInitContainers
	logger.Debugf(ctx, "Finished replacing hostname [%v] with [%v] for relevant Pod [%v/%v] containers",
		hr.Existing, hr.Replacement, pod.Namespace, pod.Name)
	return pod, changed || initContainersChanged, nil
}

func (i *ImageBuilderMutatorV1) LabelSelector() *metav1.LabelSelector {
	return &i.labelSelector
}

func newMetrics(scope promutils.Scope) metrics {
	replacementCounterBase := *scope.MustNewCounterVec("replacements", "Number of image replacements", promVersionLabelKey, promTypeLabelKey)
	return metrics{
		Scope:     scope,
		RoundTime: scope.MustNewStopWatch("round_time", "Time taken to complete a round of image builder mutator", time.Millisecond),
		Attempts:  scope.MustNewCounter("attempts", "Number of Image Builder webhook mutation attempts"),
		Failures:  scope.MustNewCounter("failures", "Number of Image Builder webhook mutation failures"),
		Replaced: replacementMetrics{
			PublicImages: versionedCounters{
				Unversioned: *replacementCounterBase.MustCurryWith(prometheus.Labels{promVersionLabelKey: promUnversionedLabelValue, promTypeLabelKey: promPublicTypeLabelValue}),
				V1:          *replacementCounterBase.MustCurryWith(prometheus.Labels{promVersionLabelKey: promV1LabelValue, promTypeLabelKey: promPublicTypeLabelValue}),
			},
			OrgImages: versionedCounters{
				Unversioned: *replacementCounterBase.MustCurryWith(prometheus.Labels{promVersionLabelKey: promUnversionedLabelValue, promTypeLabelKey: promOrgTypeLabelValue}),
				V1:          *replacementCounterBase.MustCurryWith(prometheus.Labels{promVersionLabelKey: promV1LabelValue, promTypeLabelKey: promOrgTypeLabelValue}),
			},
		},
		V1ContainerValidationAttempts: scope.MustNewCounter("v1_container_validation_attempts", "Number of attempts to validate image URI format for a specific container"),
		V1ContainerValidationFailures: scope.MustNewCounter("v1_container_validation_failures", "Number of failures to validate image URI format for a specific container"),
	}
}

func replaceByPrefix(c *corev1.Container, prefix string, replacementPrefix string) bool {
	originalImage := c.Image
	if strings.HasPrefix(originalImage, prefix) {
		c.Image = replacementPrefix + strings.TrimPrefix(originalImage, prefix)
		return true
	}
	return false
}

// Replaces hostnames in the container image names
// Adheres to version 1 URI format, assumes versionless URIs are version 1 and valid.
func (i ImageBuilderMutatorV1) replaceV1Hostname(ctx context.Context, c *corev1.Container) bool {
	originalImage := c.Image
	for _, prefix := range i.initialV1UnionPublicPrefixes {
		containerChanged := replaceByPrefix(c, prefix, i.targetPublicPrefix)
		if containerChanged {
			i.metrics.Replaced.PublicImages.V1.WithLabelValues().Inc()
			return true
		}
	}

	for _, prefix := range i.initialUnversionedUnionPublicPrefixes {
		containerChanged := replaceByPrefix(c, prefix, i.targetPublicPrefix)
		if containerChanged {
			i.metrics.Replaced.PublicImages.Unversioned.WithLabelValues().Inc()
			logger.Warnf(ctx, "Container named %s still using unversioned, public image %s", c.Name, originalImage)
			return true
		}
	}

	// If the image is not a publicly accessible image, assume it is an org based image
	for _, prefix := range i.initialV1OrgPrefixes {
		containerChanged := replaceByPrefix(c, prefix, i.targetOrgPrefix)
		if containerChanged {
			i.metrics.Replaced.OrgImages.V1.WithLabelValues().Inc()
			return true
		}
	}

	for _, prefix := range i.initialUnversionedOrgPrefixes {
		containerChanged := replaceByPrefix(c, prefix, i.targetOrgPrefix)
		if containerChanged {
			i.metrics.Replaced.OrgImages.Unversioned.WithLabelValues().Inc()
			logger.Warnf(ctx, "Container named %s still using unversioned, org specific image %s", c.Name, originalImage)
			return true
		}
	}

	return false // Container image name did not match any of the expected prefixes
}

// Validates hostname replacement to validate against v1 URI format.
// Validates non versioned URI formats as version 1 for backwards compatibility
// Returns true if the image is authorized, false otherwise
func (i ImageBuilderMutatorV1) verifyV1Prefix(ctx context.Context, containerImage string, targetHostname string, podNamespace string) bool {
	// Note, theoretically, Prefix Tree is a better data structure to use here, but the number of prefixes is small
	// Consider using if the number of prefixes grows significantly.
	orgPart := fmt.Sprintf(v1OrgsPart, podNamespace)
	validPrefixes := []string{
		// Support version-less URIs for V1. This is for backward compatibility with pre-versioned URI formats.
		fmt.Sprintf("%s/%s", targetHostname, orgPart),
		fmt.Sprintf("%s/%s/%s", targetHostname, version1URIPart, orgPart),
	}
	validPrefixes = append(validPrefixes, i.acceptablePrefixes...)

	i.metrics.V1ContainerValidationAttempts.Inc()
	// Must match one of the valid prefixes
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(containerImage, prefix) {
			return true
		}
	}
	logger.Warnf(ctx, "Container Image name %s forbidden", containerImage)
	i.metrics.V1ContainerValidationFailures.Inc()
	return false
}

func (i ImageBuilderMutatorV1) replaceHostnames(ctx context.Context, podName string, podNameSpace string, containers []corev1.Container, hr config.HostnameReplacement) (newContainers *[]corev1.Container, anyContainerChanged bool, err error) {
	anyContainerChanged = false
	res := make([]corev1.Container, 0, len(containers))

	for ic := range containers {
		container := containers[ic]
		label := fmt.Sprintf("Pod [%v/%v] container [%v]", podNameSpace, podName, container.Name)
		if i.excludedContainers.Has(container.Name) {
			logger.Debugf(ctx, "%v - Skipping container with whitelist container name [%v] from hostname replacement", label, container.Name)
			res = append(res, container)
			continue
		}
		logger.Debugf(ctx, "%v - Replacing hostname [%v] with [%v]", label, hr.Existing, hr.Replacement)
		originalImage := container.Image

		containerChanged := i.replaceV1Hostname(ctx, &container)
		if !i.hostnameReplacement.DisableVerification {
			canUse := i.verifyV1Prefix(ctx, container.Image, hr.Replacement, podNameSpace)
			if !canUse {
				logger.Warnf(ctx, "%v - Original image name [%v] resulted in unauthorized image name [%v]", label, originalImage, container.Image)
				return nil, false, fmt.Errorf("access to %s is not authorized", originalImage)
			}
		}

		if containerChanged {
			logger.Debugf(ctx, "%v - Replaced hostname [%v] with [%v]", label, hr.Existing, hr.Replacement)
		} else {
			logger.Debugf(ctx, "%v - Not replacing hostname in [%v] with [%v]", label, hr.Replacement, hr.Existing)
		}
		anyContainerChanged = anyContainerChanged || containerChanged
		res = append(res, container)
	}
	return &res, anyContainerChanged, nil
}

func NewImageBuilderMutator(cfg *config.ImageBuilderConfig, scope promutils.Scope) *ImageBuilderMutatorV1 {
	parts := []string{v1CloudTaskPart, v1UnionPublicPart}
	validPrefixes := make([]string, len(parts)*2)
	for ip, part := range parts {
		// Support pre-versioned URI formats for backward compatibility
		validPrefixes[ip] = fmt.Sprintf("%s/%s", cfg.HostnameReplacement.Replacement, part)
		// Support version 1 URI formats
		validPrefixes[ip+len(parts)] = fmt.Sprintf("%s/%s/%s", cfg.HostnameReplacement.Replacement, version1URIPart, part)
	}

	validPrefixes = append(validPrefixes, cfg.ExcludedImagePrefixes...)

	return &ImageBuilderMutatorV1{
		hostnameReplacement: cfg.HostnameReplacement,
		metrics:             newMetrics(scope),
		labelSelector:       cfg.LabelSelector,
		initialV1UnionPublicPrefixes: []string{
			// Versioned URI formats first
			fmt.Sprintf("%s/%s/%s/", cfg.HostnameReplacement.Existing, version1URIPart, unionaiOrgPlaceholder),
		},
		initialUnversionedUnionPublicPrefixes: []string{
			// Legacy non-versioned URI formats
			fmt.Sprintf("%s/%s/", cfg.HostnameReplacement.Existing, unionaiOrgPlaceholder),
		},
		initialV1OrgPrefixes: []string{
			// Scenario: Both Unionai SDK and Build-image task updated to include version.
			// Does not monkey patch replace existing hostname.
			fmt.Sprintf("%s/%s", cfg.HostnameReplacement.Existing, version1URIPart),

			// Scenario: Unionai SDK not updated and Build-image task updated to include version.
			// Build image task updated with versioned URI format
			fmt.Sprintf("%s/orgs/%s", cfg.HostnameReplacement.Replacement, version1URIPart),
		},
		initialUnversionedOrgPrefixes: []string{
			cfg.HostnameReplacement.Existing,
		},
		targetPublicPrefix: fmt.Sprintf("%s/%s/", cfg.HostnameReplacement.Replacement, unionPathReplacement),
		targetOrgPrefix:    fmt.Sprintf("%s/%s", cfg.HostnameReplacement.Replacement, orgsNamespace),
		acceptablePrefixes: validPrefixes,
		excludedContainers: sets.New[string](cfg.ExcludedContainerNames...),
	}
}
