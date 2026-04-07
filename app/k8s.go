package app

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flyte/v2/flytestdlib/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/k8s"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

// K8sConfig holds Kubernetes connection options used by InitKubernetesClient.
type K8sConfig struct {
	KubeConfig string // path to kubeconfig (empty → in-cluster → default)
	Namespace  string // namespace to ensure exists
	QPS        int    // API server QPS limit
	Burst      int    // API server burst limit
	Timeout    string // API server request timeout (e.g. "30s")
}

// InitKubernetesClient creates a controller-runtime WithWatch client.
// It tries in-cluster config first, then falls back to the default kubeconfig.
// If namespace is non-empty and Scheme is provided, it ensures the namespace exists.
func InitKubernetesClient(ctx context.Context, cfg K8sConfig, scheme *runtime.Scheme) (client.WithWatch, *rest.Config, error) {
	restConfig, err := buildRESTConfig(ctx, cfg.KubeConfig)
	if err != nil {
		return nil, nil, err
	}

	if cfg.QPS > 0 {
		restConfig.QPS = float32(cfg.QPS)
	}
	if cfg.Burst > 0 {
		restConfig.Burst = cfg.Burst
	}
	if cfg.Timeout != "" {
		d, err := time.ParseDuration(cfg.Timeout)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid timeout %q: %w", cfg.Timeout, err)
		}
		restConfig.Timeout = d
	}

	opts := client.Options{}
	if scheme != nil {
		opts.Scheme = scheme
	}

	k8sClient, err := client.NewWithWatch(restConfig, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	logger.Infof(ctx, "Kubernetes client initialized (QPS=%.0f, Burst=%d)", restConfig.QPS, restConfig.Burst)

	if cfg.Namespace != "" {
		if err := k8s.EnsureNamespaceExists(ctx, k8sClient, cfg.Namespace); err != nil {
			return nil, nil, fmt.Errorf("failed to ensure namespace exists: %w", err)
		}
	}

	return k8sClient, restConfig, nil
}

func buildRESTConfig(ctx context.Context, kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		kubeconfig = config.NormalizePath(kubeconfig)
		logger.Infof(ctx, "Using kubeconfig from: %s", kubeconfig)
		cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build k8s config from %s: %w", kubeconfig, err)
		}
		return cfg, nil
	}

	// Try in-cluster first
	logger.Infof(ctx, "Attempting in-cluster Kubernetes config")
	cfg, err := rest.InClusterConfig()
	if err == nil {
		return cfg, nil
	}

	// Fall back to default kubeconfig
	logger.Infof(ctx, "In-cluster config unavailable, falling back to default kubeconfig")
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{})
	cfg, err = kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get Kubernetes config: %w", err)
	}

	logger.Infof(ctx, "Using default kubeconfig (~/.kube/config)")
	return cfg, nil
}
