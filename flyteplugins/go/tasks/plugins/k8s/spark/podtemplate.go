package spark

import (
	"context"
	"sync"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

const sparkApplicationCRDName = "sparkapplications.sparkoperator.k8s.io"

var podTemplateCapability struct {
	once      sync.Once
	supported bool
}

// podTemplateSupported reports whether the cluster's SparkApplication CRD schema accepts the
// driver `template` field. The check runs once per process; a restart is required to pick up a
// CRD upgrade. Any failure to determine the answer (missing RBAC, API error) reports false,
// which keeps the plugin on the legacy fields only — safe against every operator version.
func podTemplateSupported(ctx context.Context) bool {
	podTemplateCapability.once.Do(func() {
		podTemplateCapability.supported = detectPodTemplateSupport(ctx)
		logger.Infof(ctx, "SparkApplication CRD pod template support: %v", podTemplateCapability.supported)
	})
	return podTemplateCapability.supported
}

func detectPodTemplateSupport(ctx context.Context) bool {
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		restConfig, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(), &clientcmd.ConfigOverrides{}).ClientConfig()
		if err != nil {
			logger.Warnf(ctx, "Cannot load kube config to inspect the SparkApplication CRD, assuming no pod template support: %v", err)
			return false
		}
	}

	clientset, err := apiextensionsclientset.NewForConfig(restConfig)
	if err != nil {
		logger.Warnf(ctx, "Cannot build apiextensions client to inspect the SparkApplication CRD, assuming no pod template support: %v", err)
		return false
	}

	getCtx, cancel := context.WithTimeout(ctx, GetSparkConfig().PodTemplateDetectionTimeout.Duration)
	defer cancel()
	crd, err := clientset.ApiextensionsV1().CustomResourceDefinitions().Get(getCtx, sparkApplicationCRDName, metav1.GetOptions{})
	if err != nil {
		logger.Warnf(ctx, "Cannot get the SparkApplication CRD (needs get on customresourcedefinitions), assuming no pod template support: %v", err)
		return false
	}

	return crdHasDriverTemplate(crd)
}

func crdHasDriverTemplate(crd *apiextensionsv1.CustomResourceDefinition) bool {
	for _, version := range crd.Spec.Versions {
		if version.Name != "v1beta2" || !version.Served {
			continue
		}
		if version.Schema == nil || version.Schema.OpenAPIV3Schema == nil {
			return false
		}
		spec, ok := version.Schema.OpenAPIV3Schema.Properties["spec"]
		if !ok {
			return false
		}
		driver, ok := spec.Properties["driver"]
		if !ok {
			return false
		}
		_, ok = driver.Properties["template"]
		return ok
	}
	return false
}
