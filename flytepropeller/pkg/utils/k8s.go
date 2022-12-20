package utils

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	flyteScheme "github.com/flyteorg/flytepropeller/pkg/client/clientset/versioned/scheme"
	"github.com/flyteorg/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flytestdlib/logger"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/record"
)

var NotTheOwnerError = errors.Errorf("FlytePropeller is not the owner")

// ResourceNvidiaGPU is the name of the Nvidia GPU resource.
const ResourceNvidiaGPU = "nvidia.com/gpu"

var invalidDNS1123Characters = regexp.MustCompile("[^-a-z0-9]+")

func ToK8sEnvVar(env []*core.KeyValuePair) []v1.EnvVar {
	envVars := make([]v1.EnvVar, 0, len(env))
	for _, kv := range env {
		envVars = append(envVars, v1.EnvVar{Name: kv.Key, Value: kv.Value})
	}
	return envVars
}

// TODO we should modify the container resources to contain a map of enum values?
// Also we should probably create tolerations / taints, but we could do that as a post process
func ToK8sResourceList(resources []*core.Resources_ResourceEntry) (v1.ResourceList, error) {
	k8sResources := make(v1.ResourceList, len(resources))
	for _, r := range resources {
		rVal := r.Value
		v, err := resource.ParseQuantity(rVal)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to parse resource as a valid quantity.")
		}
		switch r.Name {
		case core.Resources_CPU:
			if !v.IsZero() {
				k8sResources[v1.ResourceCPU] = v
			}
		case core.Resources_MEMORY:
			if !v.IsZero() {
				k8sResources[v1.ResourceMemory] = v
			}
		case core.Resources_STORAGE:
			if !v.IsZero() {
				k8sResources[v1.ResourceStorage] = v
			}
		case core.Resources_GPU:
			if !v.IsZero() {
				k8sResources[ResourceNvidiaGPU] = v
			}
		case core.Resources_EPHEMERAL_STORAGE:
			if !v.IsZero() {
				k8sResources[v1.ResourceEphemeralStorage] = v
			}
		}
	}
	return k8sResources, nil
}

func ToK8sResourceRequirements(resources *core.Resources) (*v1.ResourceRequirements, error) {
	res := &v1.ResourceRequirements{}
	if resources == nil {
		return res, nil
	}
	req, err := ToK8sResourceList(resources.Requests)
	if err != nil {
		return res, err
	}
	lim, err := ToK8sResourceList(resources.Limits)
	if err != nil {
		return res, err
	}
	res.Limits = lim
	res.Requests = req
	return res, nil
}

// GetContainer searches the provided pod spec for a container with the specified name
func GetContainer(pod *v1.PodSpec, containerName string) (*v1.Container, error) {
	for i := 0; i < len(pod.Containers); i++ {
		if pod.Containers[i].Name == containerName {
			return &pod.Containers[i], nil
		}
	}

	return nil, fmt.Errorf("container '%s' not found in podtemplate, ", containerName)
}

func GetKubeConfig(_ context.Context, cfg *config.Config) (*kubernetes.Clientset, *restclient.Config, error) {
	var kubecfg *restclient.Config
	var err error
	if cfg.KubeConfigPath != "" {
		kubeConfigPath := os.ExpandEnv(cfg.KubeConfigPath)
		kubecfg, err = clientcmd.BuildConfigFromFlags(cfg.MasterURL, kubeConfigPath)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "Error building kubeconfig")
		}
	} else {
		kubecfg, err = restclient.InClusterConfig()
		if err != nil {
			return nil, nil, errors.Wrapf(err, "Cannot get InCluster kubeconfig")
		}
	}

	kubecfg.QPS = cfg.KubeConfig.QPS
	kubecfg.Burst = cfg.KubeConfig.Burst
	kubecfg.Timeout = cfg.KubeConfig.Timeout.Duration

	kubeClient, err := kubernetes.NewForConfig(kubecfg)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "Error building kubernetes clientset")
	}
	return kubeClient, kubecfg, err
}

func GetWorkflowIDFromOwner(reference *metav1.OwnerReference, namespace string) (v1alpha1.WorkflowID, error) {
	if reference == nil {
		return "", NotTheOwnerError
	}
	if reference.Kind == v1alpha1.FlyteWorkflowKind {
		return namespace + "/" + reference.Name, nil
	}
	return "", NotTheOwnerError
}

func GetProtoTime(t *metav1.Time) *timestamp.Timestamp {
	if t != nil {
		pTime, err := ptypes.TimestampProto(t.Time)
		if err == nil {
			return pTime
		}
	}
	return ptypes.TimestampNow()
}

// SanitizeLabelValue ensures that the label value is a valid DNS-1123 string
func SanitizeLabelValue(name string) string {
	name = strings.ToLower(name)
	name = invalidDNS1123Characters.ReplaceAllString(name, "-")
	if len(name) > validation.DNS1123LabelMaxLength {
		name = name[0:validation.DNS1123LabelMaxLength]
	}
	return strings.Trim(name, "-")
}

func NewK8sEventRecorder(ctx context.Context, kubeclientset kubernetes.Interface, controllerAgentName string, publishK8sEvents bool) (record.EventRecorder, error) {
	// Create event broadcaster
	// Add FlyteWorkflow controller types to the default Kubernetes Scheme so Events can be
	// logged for FlyteWorkflow Controller types.
	err := flyteScheme.AddToScheme(scheme.Scheme)
	if err != nil {
		return nil, err
	}
	logger.Info(ctx, "Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(logger.InfofNoCtx)
	if publishK8sEvents {
		eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	}
	return eventBroadcaster.NewRecorder(scheme.Scheme, v1.EventSource{Component: controllerAgentName}), nil
}
