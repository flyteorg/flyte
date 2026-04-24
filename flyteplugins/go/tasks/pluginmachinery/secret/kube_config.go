package secret

import (
	"context"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

// resolveKubeConfig returns a *rest.Config by trying in-cluster config first,
// then falling back to the default kubeconfig loading rules (KUBECONFIG env
// var, ~/.kube/config, etc.). This lets the webhook run both inside a pod and
// as a local process for development.
func resolveKubeConfig(ctx context.Context) (*rest.Config, error) {
	if cfg, err := rest.InClusterConfig(); err == nil {
		return cfg, nil
	}
	logger.Infof(ctx, "In-cluster config unavailable, falling back to default kubeconfig")
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, &clientcmd.ConfigOverrides{}).ClientConfig()
}
